package hyperlane7683

import (
	"context"
	"fmt"

	"github.com/NethermindEth/oif-starknet/go/solvercore/base"
	"github.com/NethermindEth/oif-starknet/go/solvercore/config"
	"github.com/NethermindEth/oif-starknet/go/solvercore/logutil"
)

// BlockNumberProvider defines the interface for getting the current block number
// This allows both EVM and Starknet listeners to use the same block processing logic
type BlockNumberProvider interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

// BaseListener provides common functionality for both EVM and Starknet listeners
type BaseListener struct {
	config            base.ListenerConfig
	lastProcessedBlock uint64
	blockProvider     BlockNumberProvider
	networkType       string // "EVM" or "Starknet" for logging
}

// NewBaseListener creates a new base listener with common functionality
func NewBaseListener(config base.ListenerConfig, blockProvider BlockNumberProvider, networkType string) *BaseListener {
	return &BaseListener{
		config:        config,
		blockProvider: blockProvider,
		networkType:   networkType,
	}
}

// ProcessCurrentBlockRangeCommon processes the current block range using the common algorithm
// This eliminates duplication between EVM and Starknet listeners
func ProcessCurrentBlockRangeCommon(ctx context.Context, handler base.EventHandler, blockProvider BlockNumberProvider, listenerConfig *base.ListenerConfig, lastProcessedBlock *uint64, networkType string, processBlockRange func(context.Context, uint64, uint64, base.EventHandler) (uint64, error)) error {
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

		logutil.LogWithNetworkTag(listenerConfig.ChainName, "🧭 %s range: from=%d to=%d (current=%d, conf=%d)\n", 
			networkType, start, end, currentBlock, listenerConfig.ConfirmationBlocks)

		chunkLast, err := processBlockRange(ctx, start, end, handler)
		if err != nil {
			return fmt.Errorf("failed to process blocks %d-%d: %v", start, end, err)
		}

		newLast = chunkLast
		if err := config.UpdateLastIndexedBlock(listenerConfig.ChainName, newLast); err != nil {
			fmt.Printf("⚠️  Failed to persist LastIndexedBlock for %s: %v\n", listenerConfig.ChainName, err)
		} else {
			//fmt.Printf("💾 Persisted LastIndexedBlock=%d for %s\n", newLast, listenerConfig.ChainName)
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
