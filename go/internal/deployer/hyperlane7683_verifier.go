package deployer

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Hyperlane7683Verifier verifies Hyperlane7683 contracts on forked networks
type Hyperlane7683Verifier struct {
	client *ethclient.Client
}

// NewHyperlane7683Verifier creates a new verifier
func NewHyperlane7683Verifier(rpcURL string) (*Hyperlane7683Verifier, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", rpcURL, err)
	}

	return &Hyperlane7683Verifier{
		client: client,
	}, nil
}

// VerifyContract verifies that a Hyperlane7683 contract exists and is accessible
func (h *Hyperlane7683Verifier) VerifyContract(contractAddress common.Address) error {
	log.Printf("🔍 Verifying Hyperlane7683 contract at %s", contractAddress.Hex())

	// Check if the contract exists by calling a simple view function
	// For now, we'll just check if the address has code
	ctx := context.Background()
	code, err := h.client.CodeAt(ctx, contractAddress, nil)
	if err != nil {
		return fmt.Errorf("failed to get code at %s: %w", contractAddress.Hex(), err)
	}

	if len(code) == 0 {
		return fmt.Errorf("no code found at address %s", contractAddress.Hex())
	}

	log.Printf("   ✅ Contract code found (size: %d bytes)", len(code))
	return nil
}

// Close closes the underlying client connection
func (h *Hyperlane7683Verifier) Close() {
	if h.client != nil {
		h.client.Close()
	}
}
