package hyperlane7683

import (
	"context"
	"fmt"
	"sync"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/listener"
)

type multiNetworkListener struct {
	state     *deployer.DeploymentState
	listeners map[string]listener.BaseListener
	stopChan  chan struct{}
	mu        sync.RWMutex
}

func NewMultiNetworkListener(state *deployer.DeploymentState) listener.BaseListener {
	return &multiNetworkListener{state: state, listeners: make(map[string]listener.BaseListener), stopChan: make(chan struct{})}
}

func (m *multiNetworkListener) Start(ctx context.Context, handler listener.EventHandler) (listener.ShutdownFunc, error) {
	fmt.Printf("Starting multi-network event listener...\n")
	for networkName, networkState := range m.state.Networks {
		if err := m.createNetworkListener(networkName, networkState, handler, ctx); err != nil {
			fmt.Printf("❌ Failed to create listener for %s: %v\n", networkName, err)
			continue
		}
	}
	fmt.Printf("Multi-network listener started with %d networks\n", len(m.listeners))
	return func() { close(m.stopChan) }, nil
}

func (m *multiNetworkListener) createNetworkListener(networkName string, networkState deployer.NetworkState, handler listener.EventHandler, ctx context.Context) error {
	rpcURL := m.getRPCURLForNetwork(networkName)

	// Get network-specific listener configuration
	pollInterval, confirmationBlocks, maxBlockRange, err := config.GetListenerConfig(networkName)
	if err != nil {
		return fmt.Errorf("failed to get listener config for %s: %v", networkName, err)
	}

	// Create appropriate listener based on network type
	var l listener.BaseListener

	if networkName == "Starknet Sepolia" {
		fmt.Printf("🟣 Creating Starknet listener for %s\n", networkName)

		// Use the proper configuration helper with network-specific values
		cfg := listener.NewListenerConfig(
			networkState.HyperlaneAddress,
			networkName,
			nil, // InitialBlock not needed - we use deployment state
			pollInterval,
			confirmationBlocks,
			maxBlockRange,
		)
		fmt.Printf("   ➤ %s Hyperlane address: %s\n", networkName, networkState.HyperlaneAddress)

		l, err = NewStarknetListener(cfg, rpcURL)
		if err != nil {
			return fmt.Errorf("failed to create Starknet listener for %s: %v", networkName, err)
		}
	} else {
		// EVM networks - fallback to static config Hyperlane address if state missing
		addrToUse := networkState.HyperlaneAddress
		if addrToUse == "" {
			if staticAddr, e := config.GetHyperlaneAddress(networkName); e == nil {
				addrToUse = staticAddr.Hex()
				fmt.Printf("ℹ️  Using static Hyperlane address for %s: %s (state was empty)\n", networkName, addrToUse)
			} else {
				fmt.Printf("⚠️  No Hyperlane address available for %s in state or static config\n", networkName)
			}
		}
		// EVM networks - use the proper configuration helper with network-specific values
		cfg := listener.NewListenerConfig(
			addrToUse,
			networkName,
			nil, // InitialBlock not needed - we use deployment state
			pollInterval,
			confirmationBlocks,
			maxBlockRange,
		)
		fmt.Printf("   ➤ %s Hyperlane address: %s\n", networkName, addrToUse)

		l, err = NewEVMListener(cfg, rpcURL)
		if err != nil {
			return fmt.Errorf("failed to create EVM listener for %s: %v", networkName, err)
		}
	}

	if _, err = l.Start(ctx, handler); err != nil {
		return fmt.Errorf("failed to start listener for %s: %v", networkName, err)
	}

	m.mu.Lock()
	m.listeners[networkName] = l
	m.mu.Unlock()

	fmt.Printf("✅ Started listener for %s on %s\n", networkName, rpcURL)
	return nil
}

func (m *multiNetworkListener) getRPCURLForNetwork(networkName string) string {
	rpcURL, err := config.GetRPCURL(networkName)
	if err != nil {
		fmt.Printf("⚠️  Failed to get RPC URL for network %s, using default: %v\n", networkName, err)
		return config.GetDefaultRPCURL()
	}
	return rpcURL
}

func (m *multiNetworkListener) Stop() error {
	fmt.Printf("Stopping multi-network event listener...\n")
	close(m.stopChan)
	m.mu.RLock()
	defer m.mu.RUnlock()
	for networkName, l := range m.listeners {
		if err := l.Stop(); err != nil {
			fmt.Printf("❌ Failed to stop listener for %s: %v\n", networkName, err)
		}
	}
	fmt.Printf("Multi-network event listener stopped\n")
	return nil
}

func (m *multiNetworkListener) GetLastProcessedBlock() uint64 {
	var highest uint64
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, l := range m.listeners {
		if block := l.GetLastProcessedBlock(); block > highest {
			highest = block
		}
	}
	return highest
}
