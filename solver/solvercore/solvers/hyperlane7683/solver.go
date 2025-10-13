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
	"time"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/logutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"

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

	// Allow/block lists for controlling which orders to process
	allowBlockLists types.AllowBlockLists

	// Metadata for this solver
	metadata types.Hyperlane7683Metadata
}

func NewHyperlane7683Solver(
	getEVMClient func(chainID uint64) (*ethclient.Client, error),
	getStarknetClient func() (*rpc.Provider, error),
	getEVMSigner func(chainID uint64) (*bind.TransactOpts, error),
	getStarknetSigner func() (*account.Account, error),
	allowBlockLists types.AllowBlockLists,
) *Hyperlane7683Solver {
	metadata := types.Hyperlane7683Metadata{
		BaseMetadata:  types.BaseMetadata{ProtocolName: "Hyperlane7683"},
		IntentSources: []types.IntentSource{},
		CustomRules:   types.CustomRules{Rules: []types.RuleConfig{}},
	}

	return &Hyperlane7683Solver{
		getEVMClient:      getEVMClient,
		getStarknetClient: getStarknetClient,
		getEVMSigner:      getEVMSigner,
		getStarknetSigner: getStarknetSigner,
		evmHandlers:       make(map[uint64]ChainHandler),
		evmHandlersMux:    sync.RWMutex{},
		hyperlaneStarknet: nil, // Will be created when needed
		allowBlockLists:   allowBlockLists,
		metadata:          metadata,
	}
}

func (f *Hyperlane7683Solver) ProcessIntent(ctx context.Context, args *types.ParsedArgs) (bool, error) {
	// Log the cross-chain operation
	logutil.LogOrderProcessing(args, "Processing Order")

	// Check allow/block lists first
	if !f.isAllowedIntent(args) {
		logutil.LogOperationComplete(args, "Order processing", false)
		return false, fmt.Errorf("order blocked by allow/block lists")
	}

	// Run validation rules before processing
	rulesEngine := NewRulesEngine()
	if result := rulesEngine.EvaluateAll(ctx, args); !result.Passed {
		logutil.LogOperationComplete(args, "Order validation", false)
		return false, fmt.Errorf("order validation failed: %s", result.Reason)
	}

	// Fill method handles its own status checks efficiently (skip if already filled)
	action, err := f.Fill(ctx, args)
	if err != nil {
		logutil.LogOperationComplete(args, "Fill execution", false)
		return false, fmt.Errorf("fill execution failed: %w", err)
	}

	// Check if order is already complete (filled + settled)
	if action == OrderActionComplete {
		fmt.Printf("✅ Order already complete (filled + settled), nothing to do\n")
		return true, nil
	}

	// If fill returned OrderActionSettle, we need to settle the order
	if action == OrderActionSettle {
		// Add a small delay to ensure fill transaction is processed before settling
		time.Sleep(2 * time.Second)

		// Settle the order
		if err := f.SettleOrder(ctx, args); err != nil {
			logutil.LogOperationComplete(args, "Order settlement", false)
			return false, fmt.Errorf("order settlement failed: %w", err)
		}
	}

	// Only return true when settle completes successfully
	logutil.LogOperationComplete(args, "Order processing", true)
	return true, nil
}

func (f *Hyperlane7683Solver) Fill(ctx context.Context, args *types.ParsedArgs) (OrderAction, error) {
	logutil.LogOrderProcessing(args, "Filling Order")

	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return OrderActionError, fmt.Errorf("no fill instructions found")
	}

	// Process all fill instructions (supports both single and multiple instructions)
	for i, instruction := range args.ResolvedOrder.FillInstructions {
		logutil.LogWithNetworkTagf("", "Processing fill instruction %d/%d for chain %s",
			i+1, len(args.ResolvedOrder.FillInstructions), instruction.DestinationChainID.String())

		action, err := f.executeChainOperation(ctx, args, instruction.DestinationChainID, "fill", func(handler ChainHandler) (OrderAction, error) {
			return handler.Fill(ctx, args)
		})
		if err != nil {
			return OrderActionError, fmt.Errorf("fill instruction %d failed: %w", i+1, err)
		}

		// If any instruction fails or returns an error, return immediately
		if action == OrderActionError {
			return OrderActionError, fmt.Errorf("fill instruction %d returned error", i+1)
		}

		// If this instruction needs settlement, return that action
		if action == OrderActionSettle {
			logutil.LogWithNetworkTagf("", "Fill instruction %d completed, needs settlement", i+1)
			return OrderActionSettle, nil
		}

		// If this instruction completed successfully, continue to next
		// (In most cases there will only be one instruction, but this supports multiple)
		if action == OrderActionComplete {
			logutil.LogWithNetworkTagf("", "Fill instruction %d completed successfully", i+1)
		}
	}

	// All instructions processed successfully
	return OrderActionComplete, nil
}

func (f *Hyperlane7683Solver) SettleOrder(ctx context.Context, args *types.ParsedArgs) error {
	logutil.LogOrderProcessing(args, "Settling Order")

	// Settlement happens on the destination chain - same as fill
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found for settlement")
	}

	// Process all settlement instructions (supports both single and multiple instructions)
	for i, instruction := range args.ResolvedOrder.FillInstructions {
		logutil.LogWithNetworkTagf("", "Processing settlement instruction %d/%d for chain %s",
			i+1, len(args.ResolvedOrder.FillInstructions), instruction.DestinationChainID.String())

		_, err := f.executeChainOperation(ctx, args, instruction.DestinationChainID, "settle", func(handler ChainHandler) (OrderAction, error) {
			err := handler.Settle(ctx, args)
			return OrderActionComplete, err // Return OrderActionComplete for successful settlement
		})
		if err != nil {
			return fmt.Errorf("settlement instruction %d failed: %w", i+1, err)
		}

		logutil.LogWithNetworkTagf("", "Settlement instruction %d completed successfully", i+1)
	}

	logutil.LogOperationComplete(args, "Settlement", true)
	return nil
}

// executeChainOperation is a common helper that handles chain detection, handler retrieval, and operation execution
// This eliminates duplication between Fill, Settle, and other chain operations
func (f *Hyperlane7683Solver) executeChainOperation(
	_ context.Context,
	_ *types.ParsedArgs,
	chainID *big.Int,
	operation string,
	operationFunc func(ChainHandler) (OrderAction, error),
) (OrderAction, error) {
	var handler ChainHandler
	var err error
	var chainType string

	// Chain detection and handler retrieval
	switch {
	case f.isStarknetChain(chainID):
		chainType = "Starknet"
		handler, err = f.getStarknetHandler(chainID)
		if err != nil {
			return OrderActionError, fmt.Errorf("failed to get %s handler for chain %s: %w", chainType, chainID.String(), err)
		}

	case f.isEVMChain(chainID):
		chainType = "EVM"
		handler, err = f.getEVMHandler(chainID)
		if err != nil {
			return OrderActionError, fmt.Errorf("failed to get %s handler for chain %s: %w", chainType, chainID.String(), err)
		}

	default:
		return OrderActionError, fmt.Errorf("unsupported destination chain: %s", chainID.String())
	}

	// Execute the operation
	action, err := operationFunc(handler)
	if err != nil {
		return OrderActionError, fmt.Errorf("%s %s failed for chain %s: %w", chainType, operation, chainID.String(), err)
	}

	return action, nil
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

	handler := NewHyperlaneEVM(client, signer, chainIDUint)
	f.evmHandlers[chainIDUint] = handler
	// logutil.LogWithNetworkTag(networkName, "   🔧 Created new EVM handler\n")
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

	f.hyperlaneStarknet = NewHyperlaneStarknet(chainConfig.RPCURL, chainConfig.ChainID)
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

// isAllowedIntent checks if an intent is allowed based on allow/block lists
func (f *Hyperlane7683Solver) isAllowedIntent(args *types.ParsedArgs) bool {
	// Check block list first
	for _, blockItem := range f.allowBlockLists.BlockList {
		if f.matchesAllowBlockItem(blockItem, args) {
			return false
		}
	}

	// If no allow list is specified, allow everything
	if len(f.allowBlockLists.AllowList) == 0 {
		return true
	}

	// Check allow list
	for _, allowItem := range f.allowBlockLists.AllowList {
		if f.matchesAllowBlockItem(allowItem, args) {
			return true
		}
	}

	return false
}

// matchesAllowBlockItem checks if args match an allow/block list item
func (f *Hyperlane7683Solver) matchesAllowBlockItem(item types.AllowBlockListItem, args *types.ParsedArgs) bool {
	// Check sender address
	if item.SenderAddress != "*" && item.SenderAddress != args.SenderAddress {
		return false
	}

	// Check recipients
	for _, recipient := range args.Recipients {
		// Check destination domain
		if item.DestinationDomain != "*" && item.DestinationDomain != recipient.DestinationChainName {
			continue
		}

		// Check recipient address
		if item.RecipientAddress != "*" && item.RecipientAddress != recipient.RecipientAddress {
			continue
		}

		// If we get here, this recipient matches
		return true
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

