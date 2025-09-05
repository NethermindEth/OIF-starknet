package hyperlane7683

import (
	"context"

	"github.com/NethermindEth/oif-starknet/go/solvercore/types"
)

// OrderAction represents what action should be taken after Fill
type OrderAction int

const (
	OrderActionSettle   OrderAction = iota // Order needs settlement
	OrderActionComplete                    // Order is 100% complete (filled + settled)
	OrderActionError                       // Error occurred during fill
)

// ChainHandler defines the interface that all chain-specific handlers must implement.
// This allows easy extension to new blockchains (Cosmos, Solana, etc.) by simply
// implementing this interface and registering the handler in the solver.
//
// Usage for adding new chains:
//  1. Create hyperlane_newchain.go implementing ChainHandler
//  2. Create listener_newchain.go for event listening
//  3. Add chain detection logic to solver.go
//  4. Register in solver manager - that's it!
type ChainHandler interface {
	// Fill executes a fill operation on the chain
	// Returns OrderAction indicating next step (settle, complete, or error)
	Fill(ctx context.Context, args types.ParsedArgs) (OrderAction, error)

	// Settle executes settlement on the chain after successful fill
	// Should handle gas payments, status checks, and final settlement
	Settle(ctx context.Context, args types.ParsedArgs) error

	// GetOrderStatus returns the current status of an order on the chain
	// Common statuses: "UNKNOWN", "FILLED", "SETTLED"
	GetOrderStatus(ctx context.Context, args types.ParsedArgs) (string, error)
}

// ChainHandlerFactory creates chain handlers for specific networks
// This allows the solver to create handlers on-demand for different chains
type ChainHandlerFactory interface {
	// CreateHandler creates a new chain handler for the given chain configuration
	CreateHandler(chainID uint64, rpcURL string) (ChainHandler, error)

	// SupportsChain returns true if this factory can handle the given chain ID
	SupportsChain(chainID uint64) bool

	// GetChainType returns a human-readable name for this chain type (e.g., "EVM", "Starknet")
	GetChainType() string
}
