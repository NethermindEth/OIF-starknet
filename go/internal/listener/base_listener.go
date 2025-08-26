package listener

import (
	"context"
	"math/big"

	"github.com/NethermindEth/oif-starknet/go/internal/types"
)

// EventHandler is a function that processes parsed event arguments
// Returns (settled, error) where settled=true means the order was fully settled
type EventHandler func(args types.ParsedArgs, originChainName string, blockNumber uint64) (bool, error)

// ShutdownFunc is a function that stops the listener
type ShutdownFunc func()

// BaseListener defines the interface for event listeners
type BaseListener interface {
	// Start begins listening for events
	Start(ctx context.Context, handler EventHandler) (ShutdownFunc, error)

	// Stop gracefully stops the listener
	Stop() error

	// GetLastProcessedBlock returns the last processed block number
	GetLastProcessedBlock() uint64
}

// ListenerConfig contains configuration for a listener
type ListenerConfig struct {
	ContractAddress    string
	ChainName          string
	InitialBlock       *big.Int
	PollInterval       int // milliseconds
	ConfirmationBlocks uint64
	MaxBlockRange      uint64
}

// NewListenerConfig creates a new listener configuration
func NewListenerConfig(
	contractAddress string,
	chainName string,
	initialBlock *big.Int,
	pollInterval int,
	confirmationBlocks uint64,
	maxBlockRange uint64,
) *ListenerConfig {
	if pollInterval == 0 {
		pollInterval = 3000 // default 3 seconds
	}
	// IMPORTANT: Respect explicit 0 confirmations (disable confirmations)
	// Do not override confirmationBlocks when it's 0.
	if maxBlockRange == 0 {
		maxBlockRange = 500 // default 500 blocks
	}

	return &ListenerConfig{
		ContractAddress:    contractAddress,
		ChainName:          chainName,
		InitialBlock:       initialBlock,
		PollInterval:       pollInterval,
		ConfirmationBlocks: confirmationBlocks,
		MaxBlockRange:      maxBlockRange,
	}
}
