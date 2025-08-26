package hyperlane7683

// This file demonstrates the proposed renaming of ParsedArgs to more descriptive names
// as discussed in the refactoring plan

import (
	"math/big"

	"github.com/NethermindEth/oif-starknet/go/internal/types"
)

// 🎯 RENAMING OPTIONS for ParsedArgs:

// Option 1: CrossChainOrderEvent - emphasizes it's an event from cross-chain order
type CrossChainOrderEvent struct {
	OrderID       string                          `json:"orderId"`
	SenderAddress string                          `json:"senderAddress"`
	Recipients    []types.Recipient               `json:"recipients"`
	ResolvedOrder types.ResolvedCrossChainOrder   `json:"resolvedOrder"`
	
	// Additional context that could be useful
	OriginChainID   *big.Int `json:"originChainId,omitempty"`   // Where this event originated
	OriginBlockHash string   `json:"originBlockHash,omitempty"` // Block hash where event was emitted
	OriginTxHash    string   `json:"originTxHash,omitempty"`    // Transaction hash
	EventIndex      uint64   `json:"eventIndex,omitempty"`      // Event index within transaction
}

// Option 2: OpenOrderArgs - emphasizes it's arguments from an Open event
type OpenOrderArgs struct {
	OrderID       string                          `json:"orderId"`
	SenderAddress string                          `json:"senderAddress"`
	Recipients    []types.Recipient               `json:"recipients"`
	ResolvedOrder types.ResolvedCrossChainOrder   `json:"resolvedOrder"`
}

// Option 3: HyperlaneOpenEvent - very specific to Hyperlane Open events
type HyperlaneOpenEvent struct {
	OrderID       string                          `json:"orderId"`
	SenderAddress string                          `json:"senderAddress"`
	Recipients    []types.Recipient               `json:"recipients"`
	ResolvedOrder types.ResolvedCrossChainOrder   `json:"resolvedOrder"`
	
	// Hyperlane-specific fields
	HyperlaneDomain uint32 `json:"hyperlaneDomain"` // Hyperlane domain where event originated
	MessageID       string `json:"messageId"`       // Hyperlane message ID if applicable
}

// 🎯 RECOMMENDED: CrossChainOrderEvent
// 
// Reasons:
// 1. ✅ More descriptive than ParsedArgs
// 2. ✅ Indicates it's an event (not just args)
// 3. ✅ Emphasizes cross-chain nature
// 4. ✅ Generic enough for other protocols beyond Hyperlane
// 5. ✅ Clear that it represents a parsed blockchain event
//
// Usage would change from:
//   func ProcessIntent(ctx context.Context, args types.ParsedArgs, originChain string, blockNumber uint64) (bool, error)
// To:
//   func ProcessIntent(ctx context.Context, event types.CrossChainOrderEvent, originChain string, blockNumber uint64) (bool, error)

// 🎯 ADDITIONAL BENEFIT: Enhanced Event Context
//
// CrossChainOrderEvent could include additional blockchain context that ParsedArgs lacks:

type EnhancedCrossChainOrderEvent struct {
	// Core order data (from current ParsedArgs)
	OrderID       string                          `json:"orderId"`
	SenderAddress string                          `json:"senderAddress"`
	Recipients    []types.Recipient               `json:"recipients"`
	ResolvedOrder types.ResolvedCrossChainOrder   `json:"resolvedOrder"`
	
	// Enhanced blockchain context
	Origin struct {
		ChainID     *big.Int `json:"chainId"`
		ChainName   string   `json:"chainName"`
		BlockNumber uint64   `json:"blockNumber"`
		BlockHash   string   `json:"blockHash"`
		TxHash      string   `json:"txHash"`
		EventIndex  uint64   `json:"eventIndex"`
		Timestamp   uint64   `json:"timestamp"`
	} `json:"origin"`
	
	// Processing metadata
	ProcessedAt      uint64 `json:"processedAt"`      // When this event was processed
	ProcessingStatus string `json:"processingStatus"` // "pending", "processing", "filled", "settled"
	RetryCount       uint32 `json:"retryCount"`       // Number of processing attempts
}

// This enhanced version provides:
// 1. 🔍 Full traceability of where the event came from
// 2. 📊 Processing metadata for debugging and monitoring  
// 3. 🔄 Retry tracking for failed intents
// 4. ⏰ Timestamps for performance analysis
