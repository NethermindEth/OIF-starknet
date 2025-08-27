package types

import (
	"math/big"
)

// ParsedArgs represents the parsed arguments from an Open event
type ParsedArgs struct {
	OrderID       string                  `json:"orderId"`
	SenderAddress string                  `json:"senderAddress"`
	Recipients    []Recipient             `json:"recipients"`
	ResolvedOrder ResolvedCrossChainOrder `json:"resolvedOrder"`
}

// Recipient represents a destination recipient
type Recipient struct {
	DestinationChainName string `json:"destinationChainName"`
	RecipientAddress     string `json:"recipientAddress"`
}

// Output represents tokens that must be received for a valid order fulfillment
type Output struct {
	Token     string   `json:"token"`     // Token address as string (preserves original format)
	Amount    *big.Int `json:"amount"`    // Amount of tokens
	Recipient string   `json:"recipient"` // Recipient address as string (preserves original format)
	ChainID   *big.Int `json:"chainId"`   // Destination chain ID
}

// FillInstruction represents instructions to parameterize each leg of the fill
type FillInstruction struct {
	DestinationChainID *big.Int `json:"destinationChainId"` // Chain to fill on
	DestinationSettler string   `json:"destinationSettler"` // Contract address as string (preserves original format)
	OriginData         []byte   `json:"originData"`         // Data needed by destinationSettler
}

// ResolvedCrossChainOrder contains the order details
type ResolvedCrossChainOrder struct {
	User             string    `json:"user"`             // User initiating the transfer
	OriginChainID    *big.Int          `json:"originChainId"`    // Origin chain ID
	OpenDeadline     uint32            `json:"openDeadline"`     // Timestamp by which order must be opened
	FillDeadline     uint32            `json:"fillDeadline"`     // Timestamp by which order must be filled
	OrderID          [32]byte          `json:"orderId"`          // Unique order identifier
	MaxSpent         []Output          `json:"maxSpent"`         // Max outputs filler will send
	MinReceived      []Output          `json:"minReceived"`      // Min outputs filler must receive
	FillInstructions []FillInstruction `json:"fillInstructions"` // Instructions for each leg
}

// IntentData contains the data needed to fill an intent
type IntentData struct {
	FillInstructions []FillInstruction `json:"fillInstructions"`
	MaxSpent         []Output          `json:"maxSpent"`
}

// OrderStatus represents the status of an order
type OrderStatus uint8

const (
	OrderStatusUnknown OrderStatus = iota
	OrderStatusFilled
	OrderStatusCancelled
	OrderStatusExpired
)

// BaseMetadata contains common metadata for all solvers
type BaseMetadata struct {
	ProtocolName string `json:"protocolName"`
}

// Hyperlane7683Metadata extends base metadata with Hyperlane-specific config
type Hyperlane7683Metadata struct {
	BaseMetadata
	IntentSources []IntentSource `json:"intentSources"`
	CustomRules   CustomRules    `json:"customRules"`
}

// IntentSource represents a contract to monitor for events
type IntentSource struct {
	Address            string   `json:"address"`
	ChainName          string   `json:"chainName"`
	InitialBlock       *big.Int `json:"initialBlock,omitempty"`
	PollInterval       int      `json:"pollInterval,omitempty"`
	ConfirmationBlocks int      `json:"confirmationBlocks,omitempty"`
	ProcessedIDs       []string `json:"processedIds,omitempty"`
}

// CustomRules contains configuration for custom rules
type CustomRules struct {
	Rules []RuleConfig `json:"rules"`
}

// RuleConfig represents a single rule configuration
type RuleConfig struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// AllowBlockListItem represents a single allow/block list item
type AllowBlockListItem struct {
	SenderAddress     string `json:"senderAddress"`
	DestinationDomain string `json:"destinationDomain"`
	RecipientAddress  string `json:"recipientAddress"`
}

// AllowBlockLists contains allow and block lists
type AllowBlockLists struct {
	AllowList []AllowBlockListItem `json:"allowList"`
	BlockList []AllowBlockListItem `json:"blockList"`
}

// Result represents a generic result with success/error handling
type Result[T any] struct {
	Data    T      `json:"data,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// NewSuccessResult creates a successful result
func NewSuccessResult[T any](data T) Result[T] {
	return Result[T]{
		Data:    data,
		Success: true,
	}
}

// NewErrorResult creates an error result
func NewErrorResult[T any](err string) Result[T] {
	return Result[T]{
		Success: false,
		Error:   err,
	}
}
