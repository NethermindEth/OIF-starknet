package hyperlane7683

// Module: Solver orchestrator for Hyperlane7683
// - Applies core and custom rules to ParsedArgs
// - Routes to chain-specific handlers (EVM/Starknet) for fill and settle
// - Provides simple chain detection and client/signer acquisition

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/types"

	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Hyperlane7683Solver struct {
	// Centralized client and signer management functions from SolverManager
	getEVMClient      func(chainID uint64) (*ethclient.Client, error)
	getStarknetClient func() (*rpc.Provider, error)
	getEVMSigner      func(chainID uint64) (*bind.TransactOpts, error)
	getStarknetSigner func() (*account.Account, error)
	
	// Chain handlers implementing ChainHandler interface - now per-chain
	evmHandlers       map[uint64]ChainHandler // Map of chainID -> handler
	evmHandlersMux    sync.RWMutex            // Protects evmHandlers map
	hyperlaneStarknet ChainHandler
	
	// Metadata for this solver
	metadata          types.Hyperlane7683Metadata
}

func NewHyperlane7683Solver(
	getEVMClient func(chainID uint64) (*ethclient.Client, error),
	getStarknetClient func() (*rpc.Provider, error),
	getEVMSigner func(chainID uint64) (*bind.TransactOpts, error),
	getStarknetSigner func() (*account.Account, error),
) *Hyperlane7683Solver {
	metadata := types.Hyperlane7683Metadata{
		BaseMetadata:  types.BaseMetadata{ProtocolName: "Hyperlane7683"},
		IntentSources: []types.IntentSource{},
		CustomRules:   types.CustomRules{},
	}

	return &Hyperlane7683Solver{
		getEVMClient:      getEVMClient,
		getStarknetClient: getStarknetClient,
		getEVMSigner:      getEVMSigner,
		getStarknetSigner: getStarknetSigner,
		evmHandlers:       make(map[uint64]ChainHandler),
		metadata:          metadata,
	}
}

func (f *Hyperlane7683Solver) ProcessIntent(ctx context.Context, args types.ParsedArgs) (bool, error) {
	fmt.Printf("🔵 Processing Intent: %s-%s\n", f.metadata.ProtocolName, args.OrderID)

	// Fill method handles its own status checks efficiently (skip if already filled)
	action, err := f.Fill(ctx, args)
	if err != nil {
		return false, fmt.Errorf("fill execution failed: %w", err)
	}

	// Check if order is already complete (filled + settled)
	if action == OrderActionComplete {
		fmt.Printf("✅ Order already complete (filled + settled), nothing to do\n")
		return true, nil
	}

	// Always settle (regardless of whether we filled or skipped)
	if err := f.SettleOrder(ctx, args); err != nil {
		return false, fmt.Errorf("order settlement failed: %w", err)
	}

	// Only return true when settle completes successfully
	fmt.Printf("✅ Order processing completed successfully (fill + settle)\n")
	return true, nil
}

func (f *Hyperlane7683Solver) Fill(ctx context.Context, args types.ParsedArgs) (OrderAction, error) {
	fmt.Printf("🔵 Filling Intent: %s-%s\n", f.metadata.ProtocolName, args.OrderID)

	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return OrderActionError, fmt.Errorf("no fill instructions found")
	}

	for _, instruction := range args.ResolvedOrder.FillInstructions {
		switch {
		case f.isStarknetChain(instruction.DestinationChainID):
			// Get or create Starknet chain handler
			handler, err := f.getStarknetHandler(instruction.DestinationChainID)
			if err != nil {
				return OrderActionError, fmt.Errorf("failed to get Starknet handler for chain %s: %w", instruction.DestinationChainID.String(), err)
			}

			action, err := handler.Fill(ctx, args)
			if err != nil {
				return OrderActionError, fmt.Errorf("starknet fill failed for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			return action, nil

		case f.isEVMChain(instruction.DestinationChainID):
			// Get or create EVM chain handler
			handler, err := f.getEVMHandler(instruction.DestinationChainID)
			if err != nil {
				return OrderActionError, fmt.Errorf("failed to get EVM handler for chain %s: %w", instruction.DestinationChainID.String(), err)
			}

			action, err := handler.Fill(ctx, args)
			if err != nil {
				return OrderActionError, fmt.Errorf("EVM fill failed for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			return action, nil

		default:
			return OrderActionError, fmt.Errorf("unsupported destination chain: %s", instruction.DestinationChainID.String())
		}
	}

	// This should never happen since we return early in each case
	return OrderActionError, fmt.Errorf("no valid chain found for fill instructions")
}

func (f *Hyperlane7683Solver) SettleOrder(ctx context.Context, args types.ParsedArgs) error {
	fmt.Printf("🔵 Settling Order: %s on destination chain\n", args.OrderID)

	// Settlement happens on the destination chain - same as fill
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found for settlement")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	// Simple chain router for settlement
	switch {
	case f.isStarknetChain(instruction.DestinationChainID):
		// Get or create Starknet chain handler
		handler, err := f.getStarknetHandler(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get Starknet handler for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

		if err := handler.Settle(ctx, args); err != nil {
			return fmt.Errorf("starknet settlement failed for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

	case f.isEVMChain(instruction.DestinationChainID):
		// Get or create EVM chain handler  
		handler, err := f.getEVMHandler(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get EVM handler for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

		if err := handler.Settle(ctx, args); err != nil {
			return fmt.Errorf("EVM settlement failed for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

	default:
		return fmt.Errorf("unsupported destination chain: %s", instruction.DestinationChainID.String())
	}

	fmt.Printf("✅ Settlement successful for order %s\n", args.OrderID)
	return nil
}

// getEVMHandler gets or creates an EVM chain handler for the given chain ID
func (f *Hyperlane7683Solver) getEVMHandler(chainID *big.Int) (ChainHandler, error) {
	chainIDUint := chainID.Uint64()
	
	// Check if handler already exists for this specific chain (read lock)
	f.evmHandlersMux.RLock()
	if handler, exists := f.evmHandlers[chainIDUint]; exists {
		f.evmHandlersMux.RUnlock()
		return handler, nil
	}
	f.evmHandlersMux.RUnlock()

	// Create new EVM handler for this specific chain (write lock)
	f.evmHandlersMux.Lock()
	defer f.evmHandlersMux.Unlock()
	
	// Double-check in case another goroutine created it while we were waiting
	if handler, exists := f.evmHandlers[chainIDUint]; exists {
		return handler, nil
	}

	client, err := f.getEVMClient(chainIDUint)
	if err != nil {
		return nil, fmt.Errorf("failed to get EVM client for chain %d: %w", chainIDUint, err)
	}
	
	signer, err := f.getEVMSigner(chainIDUint)
	if err != nil {
		return nil, fmt.Errorf("failed to get EVM signer for chain %d: %w", chainIDUint, err)
	}

	handler := NewHyperlaneEVM(client, signer)
	f.evmHandlers[chainIDUint] = handler
	fmt.Printf("   🔧 Created new EVM handler for chain %d\n", chainIDUint)
	return handler, nil
}

// getStarknetHandler gets or creates a Starknet chain handler for the given chain ID  
func (f *Hyperlane7683Solver) getStarknetHandler(chainID *big.Int) (ChainHandler, error) {
	// Reuse existing handler if available
	if f.hyperlaneStarknet != nil {
		return f.hyperlaneStarknet, nil
	}

	// Create new Starknet handler
	chainConfig, err := f.getNetworkConfigByChainID(chainID)
	if err != nil {
		return nil, fmt.Errorf("starknet network not found for chain ID %s: %w", chainID.String(), err)
	}

	f.hyperlaneStarknet = NewHyperlaneStarknet(chainConfig.RPCURL)
	return f.hyperlaneStarknet, nil
}

// AddDefaultRules adds standard validation rules to the solver
func (f *Hyperlane7683Solver) AddDefaultRules() {
	// Default rules can be added here if needed in the future
	// For now, validation happens within the chain handlers themselves
}

// Simple chain identification helpers - works with any Starknet/EVM network names
func (f *Hyperlane7683Solver) isStarknetChain(chainID *big.Int) bool {
	// Ensure config is initialized to prevent segfault
	config.InitializeNetworks()

	// Find any network with "Starknet" in the name that matches this chain ID
	for networkName, network := range config.Networks {
		if network.ChainID == chainID.Uint64() {
			// Check if network name contains "Starknet" (case insensitive)
			return strings.Contains(strings.ToLower(networkName), "starknet")
		}
	}
	return false
}

func (f *Hyperlane7683Solver) isEVMChain(chainID *big.Int) bool {
	// Ensure config is initialized to prevent segfault
	config.InitializeNetworks()

	// Find any network that matches this chain ID and is NOT a Starknet chain
	for networkName, network := range config.Networks {
		if network.ChainID == chainID.Uint64() {
			// If it's not Starknet, it's EVM
			return !strings.Contains(strings.ToLower(networkName), "starknet")
		}
	}
	return false
}

// getNetworkConfigByChainID finds the network config for a given chain ID
func (f *Hyperlane7683Solver) getNetworkConfigByChainID(chainID *big.Int) (config.NetworkConfig, error) {
	// Ensure config is initialized to prevent segfault
	config.InitializeNetworks()

	chainIDUint := chainID.Uint64()
	for _, network := range config.Networks {
		if network.ChainID == chainIDUint {
			return network, nil
		}
	}
	return config.NetworkConfig{}, fmt.Errorf("network config not found for chain ID %d", chainIDUint)
}
