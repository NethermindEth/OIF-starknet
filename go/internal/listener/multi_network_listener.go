package listener

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/sirupsen/logrus"
)

// MultiNetworkListener listens to events from multiple networks simultaneously
type MultiNetworkListener struct {
	state     *deployer.DeploymentState
	logger    interface{}
	listeners map[string]BaseListener
	stopChan  chan struct{}
	mu        sync.RWMutex
}

// NewMultiNetworkListener creates a new multi-network listener
func NewMultiNetworkListener(state *deployer.DeploymentState, logger interface{}) *MultiNetworkListener {
	return &MultiNetworkListener{
		state:     state,
		logger:    logger,
		listeners: make(map[string]BaseListener),
		stopChan:  make(chan struct{}),
	}
}

// Start begins listening for events on all networks
func (m *MultiNetworkListener) Start(ctx context.Context, handler EventHandler) (ShutdownFunc, error) {
	fmt.Printf("Starting multi-network event listener...\n")
	
	// Create listeners for each network
	for networkName, networkState := range m.state.Networks {
		if err := m.createNetworkListener(networkName, networkState, handler, ctx); err != nil {
			fmt.Printf("❌ Failed to create listener for %s: %v\n", networkName, err)
			continue
		}
	}
	
	fmt.Printf("Multi-network listener started with %d networks\n", len(m.listeners))
	
	// Return shutdown function
	return func() {
		close(m.stopChan)
	}, nil
}

// createNetworkListener creates a listener for a specific network
func (m *MultiNetworkListener) createNetworkListener(networkName string, networkState deployer.NetworkState, handler EventHandler, ctx context.Context) error {
	// Get RPC URL for the network
	rpcURL := m.getRPCURLForNetwork(networkName)
	
	// Create listener config with the correct initial block from deployment state
	config := &ListenerConfig{
		ContractAddress:    networkState.HyperlaneAddress,
		ChainName:          networkName,
		InitialBlock:       big.NewInt(int64(networkState.LastIndexedBlock)), // Use the last indexed block from deployment state
		PollInterval:       1000, // 1 second
		ConfirmationBlocks: 2,
		MaxBlockRange:      500,
	}
	
	// Create EVM listener
	listener, err := NewEVMListener(config, rpcURL, m.logger.(*logrus.Logger))
	if err != nil {
		return fmt.Errorf("failed to create EVM listener for %s: %v", networkName, err)
	}
	
	// Start the listener with the proper context
	_, err = listener.Start(ctx, handler)
	if err != nil {
		return fmt.Errorf("failed to start listener for %s: %v", networkName, err)
	}
	
	// Store the listener and shutdown function
	m.mu.Lock()
	m.listeners[networkName] = listener
	m.mu.Unlock()
	
	fmt.Printf("✅ Started listener for %s on %s\n", networkName, rpcURL)
	
	return nil
}

// getRPCURLForNetwork returns the RPC URL for a given network
func (m *MultiNetworkListener) getRPCURLForNetwork(networkName string) string {
	rpcURL, err := config.GetRPCURL(networkName)
	if err != nil {
		fmt.Printf("⚠️  Failed to get RPC URL for network %s, using default: %v\n", networkName, err)
		return config.GetDefaultRPCURL()
	}
	return rpcURL
}

// Stop gracefully stops all network listeners
func (m *MultiNetworkListener) Stop() error {
	fmt.Printf("Stopping multi-network event listener...\n")
	
	close(m.stopChan)
	
	// Stop all individual listeners
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for networkName, listener := range m.listeners {
		if err := listener.Stop(); err != nil {
			fmt.Printf("❌ Failed to stop listener for %s: %v\n", networkName, err)
		}
	}
	
	fmt.Printf("Multi-network event listener stopped\n")
	return nil
}

// GetLastProcessedBlock returns the last processed block across all networks
func (l *MultiNetworkListener) GetLastProcessedBlock() uint64 {
	// For multi-network, return the highest block number
	var highestBlock uint64
	
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	for _, listener := range l.listeners {
		if block := listener.GetLastProcessedBlock(); block > highestBlock {
			highestBlock = block
		}
	}
	
	return highestBlock
}
