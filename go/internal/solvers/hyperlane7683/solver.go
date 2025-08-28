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

	"github.com/NethermindEth/oif-starknet/go/internal/base"
	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/types"

	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

// OrderAction represents what action should be taken after Fill
type OrderAction int

const (
	OrderActionSettle   OrderAction = iota // Order needs settlement
	OrderActionComplete                    // Order is 100% complete (filled + settled)
	OrderActionError                       // Error occurred during fill
)

type Hyperlane7683Solver struct {
	*base.BaseSolverImpl
	// Centralized client and signer management functions from SolverManager
	getEVMClient      func(chainID uint64) (*ethclient.Client, error)
	getStarknetClient func() (*rpc.Provider, error)
	getEVMSigner      func(chainID uint64) (*bind.TransactOpts, error)
	getStarknetSigner func() (*account.Account, error)
	hyperlaneEVM      *HyperlaneEVM
	hyperlaneStarknet *HyperlaneStarknet
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

	allowBlockLists := types.AllowBlockLists{AllowList: []types.AllowBlockListItem{}, BlockList: []types.AllowBlockListItem{}}

	return &Hyperlane7683Solver{
		BaseSolverImpl:    base.NewBaseSolver(allowBlockLists, metadata),
		getEVMClient:      getEVMClient,
		getStarknetClient: getStarknetClient,
		getEVMSigner:      getEVMSigner,
		getStarknetSigner: getStarknetSigner,
		metadata:          metadata,
	}
}

func (f *Hyperlane7683Solver) ProcessIntent(ctx context.Context, args types.ParsedArgs) (bool, error) {
	fmt.Printf("🔵 Processing Intent: %s-%s\n", f.metadata.ProtocolName, args.OrderID)

	// Always process the intent - rules only check balance/profitability, not fill status
	intent, err := f.PrepareIntent(ctx, args)
	if err != nil {
		return false, err
	}
	if !intent.Success {
		// Rules rejected the order (insufficient balance, etc.) - don't advance block
		fmt.Printf("⏭️  Intent rejected by rules: %s\n", intent.Error)
		return false, nil
	}

	// Fill method handles its own status checks efficiently (skip if already filled)
	action, err := f.Fill(ctx, args, intent.Data)
	if err != nil {
		return false, fmt.Errorf("fill execution failed: %w", err)
	}

	// Check if order is already complete (filled + settled)
	if action == OrderActionComplete {
		fmt.Printf("✅ Order already complete (filled + settled), nothing to do\n")
		return true, nil
	}

	// Always settle (regardless of whether we filled or skipped)
	if err := f.SettleOrder(ctx, args, intent.Data); err != nil {
		return false, fmt.Errorf("order settlement failed: %w", err)
	}

	// Only return true when settle completes successfully
	fmt.Printf("✅ Order processing completed successfully (fill + settle)\n")
	return true, nil
}

func (f *Hyperlane7683Solver) Fill(ctx context.Context, args types.ParsedArgs, data types.IntentData) (OrderAction, error) {
	fmt.Printf("🔵 Filling Intent: %s-%s\n", f.metadata.ProtocolName, args.OrderID)

	for i, instruction := range data.FillInstructions {
		fmt.Printf("📦 Instruction %d: Chain %s, Settler %s\n", i+1, instruction.DestinationChainID.String(), instruction.DestinationSettler)

		// Simple chain router - clean and extensible
		switch {
		case f.isStarknetChain(instruction.DestinationChainID):
			// Get Starknet RPC URL from config by finding the network with matching chain ID
			chainConfig, err := f.getNetworkConfigByChainID(instruction.DestinationChainID)
			if err != nil {
				return OrderActionError, fmt.Errorf("starknet network not found for chain ID %s: %w", instruction.DestinationChainID.String(), err)
			}

			// Reuse existing instance or create new one
			if f.hyperlaneStarknet == nil {
				f.hyperlaneStarknet = NewHyperlaneStarknet(chainConfig.RPCURL)
			}

			action, err := f.hyperlaneStarknet.Fill(ctx, args)
			if err != nil {
				return OrderActionError, fmt.Errorf("starknet fill failed for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			return action, nil

		case f.isEVMChain(instruction.DestinationChainID):
			// Get EVM client and signer for this chain
			client, err := f.getClientForChain(instruction.DestinationChainID)
			if err != nil {
				return OrderActionError, fmt.Errorf("failed to get client for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			signer, err := f.getSignerForChain(instruction.DestinationChainID)
			if err != nil {
				return OrderActionError, fmt.Errorf("failed to get signer for chain %s: %w", instruction.DestinationChainID.String(), err)
			}

			// Reuse existing instance or create new one
			if f.hyperlaneEVM == nil || f.hyperlaneEVM.client != client {
				f.hyperlaneEVM = NewHyperlaneEVM(client, signer)
			}

			action, err := f.hyperlaneEVM.Fill(ctx, args)
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

func (f *Hyperlane7683Solver) SettleOrder(ctx context.Context, args types.ParsedArgs, data types.IntentData) error {
	fmt.Printf("🔵 Settling Order: %s on destination chain\n", args.OrderID)

	// Settlement happens on the destination chain - same as fill
	if len(data.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found for settlement")
	}

	instruction := data.FillInstructions[0]

	// Simple chain router for settlement
	switch {
	case f.isStarknetChain(instruction.DestinationChainID):
		// Get Starknet RPC URL from config by finding the network with matching chain ID
		chainConfig, err := f.getNetworkConfigByChainID(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("starknet network not found for chain ID %s: %w", instruction.DestinationChainID.String(), err)
		}

		// Reuse existing instance or create new one
		if f.hyperlaneStarknet == nil {
			f.hyperlaneStarknet = NewHyperlaneStarknet(chainConfig.RPCURL)
		}

		if err := f.hyperlaneStarknet.Settle(ctx, args); err != nil {
			return fmt.Errorf("starknet settlement failed for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

	case f.isEVMChain(instruction.DestinationChainID):
		// Get EVM client and signer for this chain
		client, err := f.getClientForChain(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get client for chain %s: %w", instruction.DestinationChainID.String(), err)
		}
		signer, err := f.getSignerForChain(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get signer for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

		// Reuse existing instance or create new one
		if f.hyperlaneEVM == nil || f.hyperlaneEVM.client != client {
			f.hyperlaneEVM = NewHyperlaneEVM(client, signer)
		}

		if err := f.hyperlaneEVM.Settle(ctx, args); err != nil {
			return fmt.Errorf("EVM settlement failed for chain %s: %w", instruction.DestinationChainID.String(), err)
		}

	default:
		return fmt.Errorf("unsupported destination chain: %s", instruction.DestinationChainID.String())
	}

	fmt.Printf("✅ Settlement successful for order %s\n", args.OrderID)
	return nil
}

func (f *Hyperlane7683Solver) AddDefaultRules() {
	f.AddRule(f.enoughBalanceOnDestination) // Pre-validate solver has enough tokens
	f.AddRule(f.filterByTokenAndAmount)     // Validate profitability and limits
}

func (f *Hyperlane7683Solver) getClientForChain(chainID *big.Int) (*ethclient.Client, error) {
	chainIDUint := chainID.Uint64()
	// Use the centralized client management from SolverManager
	return f.getEVMClient(chainIDUint)
}

func (f *Hyperlane7683Solver) getSignerForChain(chainID *big.Int) (*bind.TransactOpts, error) {
	chainIDUint := chainID.Uint64()
	// Use centralized signer management from SolverManager
	return f.getEVMSigner(chainIDUint)
}

// Simple chain identification helpers - works with any Starknet/EVM network names
func (f *Hyperlane7683Solver) isStarknetChain(chainID *big.Int) bool {
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
	chainIDUint := chainID.Uint64()
	for _, network := range config.Networks {
		if network.ChainID == chainIDUint {
			return network, nil
		}
	}
	return config.NetworkConfig{}, fmt.Errorf("network config not found for chain ID %d", chainIDUint)
}
