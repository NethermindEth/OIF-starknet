package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/listener"
	contracts "github.com/NethermindEth/oif-starknet/go/internal/solvers/hyperlane7683"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SolverConfig defines configuration for a solver
type SolverConfig struct {
	Enabled bool                   `json:"enabled"`
	Options map[string]interface{} `json:"options"`
}

// SolverRegistry maps solver names to their configurations
type SolverRegistry map[string]SolverConfig

// SolverManager manages multiple protocol solvers
// Following the TypeScript SolverManager pattern
type SolverManager struct {
	client          *ethclient.Client
	activeShutdowns []func()
	solverRegistry  SolverRegistry
}

// NewSolverManager creates a new solver manager
func NewSolverManager(client *ethclient.Client) *SolverManager {
	// Default solver registry - could be loaded from config file
	registry := SolverRegistry{
		"hyperlane7683": {
			Enabled: true,
			Options: map[string]interface{}{},
		},
	}

	return &SolverManager{
		client:          client,
		activeShutdowns: make([]func(), 0),
		solverRegistry:  registry,
	}
}

// InitializeSolvers starts all enabled solvers
func (sm *SolverManager) InitializeSolvers(ctx context.Context) error {
	fmt.Printf("🚀 Initializing solvers...\n")

	for solverName, config := range sm.solverRegistry {
		if !config.Enabled {
			fmt.Printf("   ⏭️  Solver %s is disabled, skipping...\n", solverName)
			continue
		}

		fmt.Printf("🔧 Initializing solver: %s\n", solverName)

		if err := sm.initializeSolver(ctx, solverName, config); err != nil {
			return fmt.Errorf("failed to initialize solver %s: %w", solverName, err)
		}

		fmt.Printf("✅ Solver %s initialized successfully\n", solverName)
	}

	fmt.Printf("✅ All solvers initialized\n")
	return nil
}

// initializeSolver starts a specific solver
func (sm *SolverManager) initializeSolver(ctx context.Context, name string, config SolverConfig) error {
	switch name {
	case "hyperlane7683":
		return sm.initializeHyperlane7683(ctx)
	default:
		return fmt.Errorf("unknown solver: %s", name)
	}
}

// initializeHyperlane7683 starts the Hyperlane 7683 solver
func (sm *SolverManager) initializeHyperlane7683(ctx context.Context) error {
	// Create filler
	hyperlane7683Filler := contracts.NewHyperlane7683Filler(sm.client)
	hyperlane7683Filler.AddDefaultRules()

	// Event handler that processes intents
	eventHandler := func(args types.ParsedArgs, originChainName string, blockNumber uint64) (bool, error) {
		return hyperlane7683Filler.ProcessIntent(ctx, args, originChainName, blockNumber)
	}

	// Start listeners for each intent source
	for _, source := range []string{"Base", "Optimism", "Arbitrum", "Ethereum", "Starknet"} {
		networkConfig, exists := config.Networks[source]
		if !exists {
			fmt.Printf("   ⚠️  Network %s not found in config, skipping...\n", source)
			continue
		}

		var shutdown listener.ShutdownFunc

		// Create appropriate listener based on chain type
		if source == "Starknet" {
			hyperlaneAddr, err := getStarknetHyperlaneAddress(networkConfig)
			if err != nil {
				return fmt.Errorf("failed to get Starknet Hyperlane address: %w", err)
			}

			// Create Starknet listener config
			listenerConfig := listener.NewListenerConfig(
				hyperlaneAddr,
				source,
				big.NewInt(int64(networkConfig.SolverStartBlock)), // start from configured block
				networkConfig.PollInterval,                        // poll interval from config
				uint64(networkConfig.ConfirmationBlocks),          // confirmation blocks from config
				networkConfig.MaxBlockRange,                       // max block range from config
			)

			starknetListener, err := contracts.NewStarknetListener(listenerConfig, networkConfig.RPCURL)
			if err != nil {
				return fmt.Errorf("failed to create Starknet listener: %w", err)
			}
			shutdown, err = starknetListener.Start(ctx, eventHandler)
			if err != nil {
				return fmt.Errorf("failed to start Starknet listener for %s: %w", source, err)
			}
		} else {
			// Create EVM listener config
			listenerConfig := listener.NewListenerConfig(
				networkConfig.HyperlaneAddress.Hex(),
				source,
				big.NewInt(int64(networkConfig.SolverStartBlock)), // start from configured block
				networkConfig.PollInterval,                        // poll interval from config
				uint64(networkConfig.ConfirmationBlocks),          // confirmation blocks from config
				networkConfig.MaxBlockRange,                       // max block range from config
			)

			evmListener, err := contracts.NewEVMListener(listenerConfig, networkConfig.RPCURL)
			if err != nil {
				return fmt.Errorf("failed to create EVM listener: %w", err)
			}
			shutdown, err = evmListener.Start(ctx, eventHandler)
			if err != nil {
				return fmt.Errorf("failed to start EVM listener for %s: %w", source, err)
			}
		}

		sm.activeShutdowns = append(sm.activeShutdowns, shutdown)
		//fmt.Printf("     📡 Started listener for %s\n", source)
	}

	return nil
}

// AddSolver dynamically adds a new solver to the registry
func (sm *SolverManager) AddSolver(name string, config SolverConfig) {
	sm.solverRegistry[name] = config
}

// EnableSolver enables a solver
func (sm *SolverManager) EnableSolver(name string) error {
	if config, exists := sm.solverRegistry[name]; exists {
		config.Enabled = true
		sm.solverRegistry[name] = config
		return nil
	}
	return fmt.Errorf("solver %s not found", name)
}

// DisableSolver disables a solver
func (sm *SolverManager) DisableSolver(name string) error {
	if config, exists := sm.solverRegistry[name]; exists {
		config.Enabled = false
		sm.solverRegistry[name] = config
		return nil
	}
	return fmt.Errorf("solver %s not found", name)
}

// Shutdown stops all active solvers
func (sm *SolverManager) Shutdown() {
	fmt.Printf("🛑 Shutting down solvers...\n")

	for i, shutdown := range sm.activeShutdowns {
		fmt.Printf("   📡 Stopping listener %d\n", i+1)
		shutdown()
	}

	sm.activeShutdowns = make([]func(), 0)
	fmt.Printf("✅ All solvers shut down\n")
}

// GetSolverStatus returns the status of all solvers
func (sm *SolverManager) GetSolverStatus() map[string]bool {
	status := make(map[string]bool)
	for name, config := range sm.solverRegistry {
		status[name] = config.Enabled
	}
	return status
}

// getStarknetHyperlaneAddress gets the correct Starknet Hyperlane address based on FORKING mode
func getStarknetHyperlaneAddress(networkConfig config.NetworkConfig) (string, error) {
	forkingStr := strings.ToLower(os.Getenv("FORKING"))
	// Check FORKING environment variable (default: true for local forks)
	if forkingStr == "" {
		forkingStr = "true"
	}
	isForking, _ := strconv.ParseBool(forkingStr)

	if isForking {
		// Local forks: Use deployment state (fresh deployments)
		if deploymentAddr := getStarknetHyperlaneFromDeploymentState(); deploymentAddr != "" {
			fmt.Printf("🔄 FORKING=true: Using Starknet Hyperlane address from deployment state: %s\n", deploymentAddr)
			return deploymentAddr, nil
		} else {
			return "", fmt.Errorf("FORKING=true but no Starknet Hyperlane address found in deployment state")
		}
	} else {
		// Live networks: Use .env address (manually configured)
		envAddr := networkConfig.HyperlaneAddress.Hex()
		if envAddr != "0x0000000000000000000000000000000000000000" && envAddr != "" {
			fmt.Printf("   🔄 FORKING=false: Using Starknet Hyperlane address from .env: %s\n", envAddr)
			return envAddr, nil
		} else {
			return "", fmt.Errorf("FORKING=false but no STARKNET_HYPERLANE_ADDRESS set in .env")
		}
	}
}

// getStarknetHyperlaneFromDeploymentState loads Starknet Hyperlane address from deployment state
func getStarknetHyperlaneFromDeploymentState() string {
	paths := []string{"state/network_state/deployment-state.json", "../state/network_state/deployment-state.json", "../../state/network_state/deployment-state.json"}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var deploymentState struct {
			Networks map[string]struct {
				ChainID          uint64 `json:"chainId"`
				HyperlaneAddress string `json:"hyperlaneAddress"`
				OrcaCoinAddress  string `json:"orcaCoinAddress"`
				DogCoinAddress   string `json:"dogCoinAddress"`
			} `json:"networks"`
		}
		if err := json.Unmarshal(data, &deploymentState); err != nil {
			continue
		}
		if stark, ok := deploymentState.Networks["Starknet"]; ok && stark.HyperlaneAddress != "" {
			return stark.HyperlaneAddress
		}
		if starkLegacy, ok := deploymentState.Networks["Starknet"]; ok && starkLegacy.HyperlaneAddress != "" {
			return starkLegacy.HyperlaneAddress
		}
	}
	return ""
}
