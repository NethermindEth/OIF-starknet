package hyperlane7683

import (
	"context"
	"fmt"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/base"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/logutil"
)

// BlockNumberProvider defines the interface for getting the current block number
// This allows both EVM and Starknet listeners to use the same block processing logic
type BlockNumberProvider interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

// BaseListener provides common functionality for both EVM and Starknet listeners
type BaseListener struct {
	config             base.ListenerConfig
	lastProcessedBlock uint64
	blockProvider      BlockNumberProvider
	networkType        string // "EVM" or "Starknet" for logging
}

// NewBaseListener creates a new base listener with common functionality
func NewBaseListener(config base.ListenerConfig, blockProvider BlockNumberProvider, networkType string) *BaseListener {
	return &BaseListener{
		config:             config,
		lastProcessedBlock: 0,
		blockProvider:      blockProvider,
		networkType:        networkType,
	}
}

// ResolveSolverStartBlock resolves the actual start block based on solver start block configuration
// - Positive number: start at that specific block
// - Zero: start at current block (live)
// - Negative number: start N blocks before current block
func ResolveSolverStartBlock(ctx context.Context, solverStartBlock int64, blockProvider BlockNumberProvider) (uint64, error) {
	if solverStartBlock > 0 {
		// Positive number - use as-is
		return uint64(solverStartBlock), nil
	}

	// Zero or negative number - need current block
	currentBlock, err := blockProvider.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current block number: %v", err)
	}

	if solverStartBlock == 0 {
		// Zero - start at current block (live)
		return currentBlock, nil
	}

	// Negative number - start N blocks before current block
	// Calculate start block: current - abs(solverStartBlock)
	startBlock := currentBlock - uint64(-solverStartBlock)

	// Ensure we don't go below block 0
	if startBlock > currentBlock {
		startBlock = 0
	}

	return startBlock, nil
}

// ProcessCurrentBlockRangeCommon processes the current block range using the common algorithm
// This eliminates duplication between EVM and Starknet listeners
func ProcessCurrentBlockRangeCommon(
	ctx context.Context,
	handler base.EventHandler,
	blockProvider BlockNumberProvider,
	listenerConfig *base.ListenerConfig,
	lastProcessedBlock *uint64,
	networkType string,
	processBlockRange func(context.Context, uint64, uint64, base.EventHandler) (uint64, error),
) error {
	currentBlock, err := blockProvider.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}

	// Apply confirmations window if configured
	safeBlock := currentBlock
	if listenerConfig.ConfirmationBlocks > 0 && currentBlock > listenerConfig.ConfirmationBlocks {
		safeBlock = currentBlock - listenerConfig.ConfirmationBlocks
	}

	fromBlock := *lastProcessedBlock + 1
	toBlock := safeBlock

	// Check if we have any new blocks to process
	if fromBlock > toBlock {
		// No new blocks to process, we're up to date
		return nil
	}

	// Respect MaxBlockRange by chunking large ranges
	chunkSize := listenerConfig.MaxBlockRange
	newLast := *lastProcessedBlock

	for start := fromBlock; start <= toBlock; start += chunkSize {
		end := start + chunkSize - 1
		if end > toBlock {
			end = toBlock
		}

		logutil.LogWithNetworkTagf(listenerConfig.ChainName, "🧭 %s range: from=%d to=%d (current=%d, conf=%d)\n",
			networkType, start, end, currentBlock, listenerConfig.ConfirmationBlocks)

		chunkLast, err := processBlockRange(ctx, start, end, handler)
		if err != nil {
			return fmt.Errorf("failed to process blocks %d-%d: %v", start, end, err)
		}

		newLast = chunkLast
		if err := config.UpdateLastIndexedBlock(listenerConfig.ChainName, newLast); err != nil {
			fmt.Printf("⚠️  Failed to persist LastIndexedBlock for %s: %v\n", listenerConfig.ChainName, err)
		}
	}

	// Block processing complete
	*lastProcessedBlock = newLast
	return nil
}

// GetLastProcessedBlock returns the last processed block number
func (bl *BaseListener) GetLastProcessedBlock() uint64 {
	return bl.lastProcessedBlock
}

// SetLastProcessedBlock sets the last processed block number
func (bl *BaseListener) SetLastProcessedBlock(block uint64) {
	bl.lastProcessedBlock = block
}

// GetConfig returns the listener configuration
func (bl *BaseListener) GetConfig() base.ListenerConfig {
	return bl.config
}

// CatchUpHistoricalBlocks processes historical blocks using the common logic
func (bl *BaseListener) CatchUpHistoricalBlocks(
	ctx context.Context,
	handler base.EventHandler,
	processBlockRange func(context.Context, uint64, uint64, base.EventHandler) (uint64, error),
) error {
	p := logutil.Prefix(bl.config.ChainName)
	fmt.Printf("%s🔄 Catching up on historical blocks...\n", p)

	currentBlock, err := bl.blockProvider.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("%sfailed to get current block number: %v", p, err)
	}

	// Apply confirmations during backfill as well
	safeBlock := currentBlock
	if bl.config.ConfirmationBlocks > 0 && currentBlock > bl.config.ConfirmationBlocks {
		safeBlock = currentBlock - bl.config.ConfirmationBlocks
	}

	// Start from the last processed block + 1 (which should be the solver start block)
	fromBlock := bl.lastProcessedBlock + 1
	toBlock := safeBlock
	if fromBlock >= toBlock {
		fmt.Printf("%s✅ Already up to date, no historical blocks to process\n", p)
		return nil
	}

	chunkSize := bl.config.MaxBlockRange
	for start := fromBlock; start < toBlock; start += chunkSize {
		end := start + chunkSize
		if end > toBlock {
			end = toBlock
		}

		newLast, err := processBlockRange(ctx, start, end, handler)
		if err != nil {
			return fmt.Errorf("%sfailed to process historical blocks %d-%d: %v", p, start, end, err)
		}

		bl.lastProcessedBlock = newLast
		if err := config.UpdateLastIndexedBlock(bl.config.ChainName, newLast); err != nil {
			fmt.Printf("%s⚠️  Failed to persist LastIndexedBlock: %v\n", p, err)
		}
	}

	fmt.Printf("%s✅ Historical block processing complete\n", p)
	return nil
}

// CommonListenerConfig holds common configuration for both EVM and Starknet listeners
type CommonListenerConfig struct {
	ListenerConfig *base.ListenerConfig
	LastProcessedBlock uint64
}

// ResolveCommonListenerConfig resolves common listener configuration
func ResolveCommonListenerConfig(
	ctx context.Context,
	listenerConfig *base.ListenerConfig,
	blockProvider BlockNumberProvider,
) (*CommonListenerConfig, error) {
	configStartBlock := listenerConfig.InitialBlock.Int64()
	var resolvedStartBlock uint64

	if configStartBlock > 0 {
		// Positive number - use as-is
		resolvedStartBlock = uint64(configStartBlock)
	} else {
		// Zero or negative number - need current block
		currentBlock, err := blockProvider.BlockNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current block number: %v", err)
		}

		if configStartBlock == 0 {
			// Zero - start at current block (live)
			resolvedStartBlock = currentBlock
			fmt.Printf("%s📚 Start block was 0, using current block %d\n",
				logutil.Prefix(listenerConfig.ChainName), currentBlock)
		} else {
			// Negative number - start N blocks before current block
			// Calculate start block: current - abs(configStartBlock)
			resolvedStartBlock = currentBlock - uint64(-configStartBlock)

			// Ensure we don't go below block 0
			if resolvedStartBlock > currentBlock {
				resolvedStartBlock = 0
			}

			fmt.Printf("%s📚 Start block was %d, using current block %d - %d = %d\n",
				logutil.Prefix(listenerConfig.ChainName), configStartBlock, currentBlock, -configStartBlock, resolvedStartBlock)
		}
	}

	state, err := config.GetSolverState()
	if err != nil {
		return nil, fmt.Errorf("failed to get solver state: %w", err)
	}

	var lastProcessedBlock uint64
	if networkState, exists := state.Networks[listenerConfig.ChainName]; exists {
		deploymentStateBlock := networkState.LastIndexedBlock
		if deploymentStateBlock > resolvedStartBlock {
			lastProcessedBlock = deploymentStateBlock
			fmt.Printf("%s📚 Using deployment state block %d (higher than config start block %d)\n",
				logutil.Prefix(listenerConfig.ChainName), deploymentStateBlock, resolvedStartBlock)
		} else {
			lastProcessedBlock = resolvedStartBlock
			fmt.Printf("%s📚 Using config start block %d (deployment state block %d is lower)\n",
				logutil.Prefix(listenerConfig.ChainName), resolvedStartBlock, deploymentStateBlock)
		}
	} else {
		return nil, fmt.Errorf("network %s not found in solver state", listenerConfig.ChainName)
	}

	return &CommonListenerConfig{
		ListenerConfig:     listenerConfig,
		LastProcessedBlock: lastProcessedBlock,
	}, nil
}
