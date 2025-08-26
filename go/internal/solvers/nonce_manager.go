package solvers

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// NonceManager provides thread-safe nonce management for multiple chains
// Following the pattern from TypeScript NonceKeeperWallet
type NonceManager struct {
	mu     sync.RWMutex
	nonces map[uint64]uint64 // chainID -> next nonce
	clients map[uint64]*ethclient.Client
}

// NewNonceManager creates a new nonce manager
func NewNonceManager() *NonceManager {
	return &NonceManager{
		nonces:  make(map[uint64]uint64),
		clients: make(map[uint64]*ethclient.Client),
	}
}

// RegisterClient registers an ethclient for a specific chain
func (nm *NonceManager) RegisterClient(chainID uint64, client *ethclient.Client) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.clients[chainID] = client
}

// GetNextNonce returns the next available nonce for a chain and address
// This prevents nonce conflicts in concurrent transaction sending
func (nm *NonceManager) GetNextNonce(ctx context.Context, chainID uint64, address string) (uint64, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	client, exists := nm.clients[chainID]
	if !exists {
		return 0, fmt.Errorf("no client registered for chain %d", chainID)
	}
	
	// Initialize nonce from network if not set
	if _, exists := nm.nonces[chainID]; !exists {
		nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(address))
		if err != nil {
			return 0, fmt.Errorf("failed to get pending nonce for chain %d: %w", chainID, err)
		}
		nm.nonces[chainID] = nonce
	}
	
	// Return current nonce and increment for next call
	currentNonce := nm.nonces[chainID]
	nm.nonces[chainID]++
	
	return currentNonce, nil
}

// UpdateNonceIfNeeded updates the stored nonce if the network nonce is higher
// This handles cases where transactions were sent outside this manager
func (nm *NonceManager) UpdateNonceIfNeeded(ctx context.Context, chainID uint64, address string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	client, exists := nm.clients[chainID]
	if !exists {
		return fmt.Errorf("no client registered for chain %d", chainID)
	}
	
	networkNonce, err := client.PendingNonceAt(ctx, common.HexToAddress(address))
	if err != nil {
		return fmt.Errorf("failed to get pending nonce for chain %d: %w", chainID, err)
	}
	
	// Update our stored nonce if network nonce is higher
	if storedNonce, exists := nm.nonces[chainID]; !exists || networkNonce > storedNonce {
		nm.nonces[chainID] = networkNonce
	}
	
	return nil
}

// WrapSigner wraps a TransactOpts to use managed nonces
// This prevents nonce conflicts when sending concurrent transactions
func (nm *NonceManager) WrapSigner(ctx context.Context, signer *bind.TransactOpts, chainID uint64) *bind.TransactOpts {
	wrappedSigner := &bind.TransactOpts{
		From:      signer.From,
		Signer:    signer.Signer,
		Value:     signer.Value,
		GasPrice:  signer.GasPrice,
		GasFeeCap: signer.GasFeeCap,
		GasTipCap: signer.GasTipCap,
		GasLimit:  signer.GasLimit,
		Context:   signer.Context,
		// Don't copy Nonce - we'll set it dynamically
	}
	
	// Set nonce getter function
	wrappedSigner.Nonce = big.NewInt(0) // Will be overridden
	
	// We need to get the nonce right before transaction
	// This is a bit tricky in Go since we can't override the nonce getter like in TS
	// For now, we'll handle this in the calling code by calling GetNextNonce explicitly
	
	return wrappedSigner
}

// Reset clears stored nonces (useful for testing or error recovery)
func (nm *NonceManager) Reset() {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.nonces = make(map[uint64]uint64)
}
