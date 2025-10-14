package solvercore

import (
	"context"
	"fmt"
	"math/big"

	"strings"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/base"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	contracts "github.com/NethermindEth/oif-starknet/solver/solvercore/solvers/hyperlane7683"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Module: Solver Manager for Hyperlane7683 Protocol
// - Manages multiple protocol solvers (EVM and Starknet)
// - Provides centralized client and signer management
// - Coordinates solver initialization and lifecycle

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
	evmClients      map[uint64]*ethclient.Client
	starknetClient  *rpc.Provider
	activeShutdowns []func()
	solverRegistry  SolverRegistry
	allowBlockLists types.AllowBlockLists
}

// NewSolverManager creates a new solver manager
func NewSolverManager(cfg *config.Config) *SolverManager {
	// Default solver registry - could be loaded from config file
	registry := SolverRegistry{
		"hyperlane7683": {
			Enabled: true,
			Options: map[string]interface{}{},
		},
	}

	return &SolverManager{
		evmClients:      make(map[uint64]*ethclient.Client),
		starknetClient:  nil, // Will be initialized later
		activeShutdowns: make([]func(), 0),
		solverRegistry:  registry,
		allowBlockLists: types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{},
			BlockList: []types.AllowBlockListItem{},
		},
	}
}

// SetAllowBlockLists configures the allow/block lists for the solver manager
// This allows runtime configuration of which orders to process
func (sm *SolverManager) SetAllowBlockLists(allowBlockLists types.AllowBlockLists) {
	sm.allowBlockLists = allowBlockLists
}

// GetAllowBlockLists returns the current allow/block lists configuration
func (sm *SolverManager) GetAllowBlockLists() types.AllowBlockLists {
	return sm.allowBlockLists
}

// InitializeSolvers starts all enabled solvers
func (sm *SolverManager) InitializeSolvers(ctx context.Context) error {
	fmt.Printf("🚀 Initializing solvers...\n")

	// Initialize EVM clients for all EVM networks
	if err := sm.initializeEVMClients(); err != nil {
		return fmt.Errorf("failed to initialize EVM clients: %w", err)
	}

	// Initialize Starknet client
	if err := sm.initializeStarknetClients(); err != nil {
		return fmt.Errorf("failed to initialize Starknet client: %w", err)
	}

	// Initialize individual solvers
	for solverName, config := range sm.solverRegistry {
		if !config.Enabled {
			fmt.Printf("   ⏭️  Solver %s is disabled, skipping...\n", solverName)
			continue
		}

		if err := sm.initializeSolver(ctx, solverName); err != nil {
			return fmt.Errorf("failed to initialize solver %s: %w", solverName, err)
		}
	}

	fmt.Printf("✅ All solvers initialized successfully\n")
	return nil
}

// initializeSolver starts a specific solver
func (sm *SolverManager) initializeSolver(ctx context.Context, name string) error {
	switch name {
	case "hyperlane7683":
		return sm.initializeHyperlane7683(ctx)
	default:
		return fmt.Errorf("unknown solver: %s", name)
	}
}

// initializeEVMClients initializes EVM RPC connections for all EVM networks
func (sm *SolverManager) initializeEVMClients() error {
	fmt.Printf("🔗 Initializing EVM clients...\n")

	evmCount := 0
	for networkName, networkConfig := range config.Networks {
		// Check if this is NOT a Starknet network (i.e., it's an EVM network)
		if strings.Contains(strings.ToLower(networkName), "starknet") {
			continue
		}
		
		fmt.Printf("   🔗 Initializing EVM client for %s (Chain ID: %d)\n", networkName, networkConfig.ChainID)

		client, err := ethclient.Dial(networkConfig.RPCURL)
		if err != nil {
			return fmt.Errorf("failed to create EVM client for %s: %w", networkName, err)
		}

		sm.evmClients[networkConfig.ChainID] = client
		fmt.Printf("   ✅ EVM client initialized for %s\n", networkName)
		evmCount++
	}

	fmt.Printf("✅ All EVM clients initialized (%d networks)\n", evmCount)
	return nil
}

// initializeStarknetClients initializes Starknet RPC connection for the first Starknet network found
func (sm *SolverManager) initializeStarknetClients() error {
	fmt.Printf("🔗 Initializing Starknet client...\n")

	for networkName, networkConfig := range config.Networks {
		// Check if this is a Starknet network
		if !strings.Contains(strings.ToLower(networkName), "starknet") {
			continue
		}
		
		fmt.Printf("   🔗 Initializing Starknet client for %s (Chain ID: %d)\n", networkName, networkConfig.ChainID)

		provider, err := rpc.NewProvider(networkConfig.RPCURL)
		if err != nil {
			return fmt.Errorf("failed to create Starknet provider for %s: %w", networkName, err)
		}

		sm.starknetClient = provider
		fmt.Printf("✅ Starknet client initialized successfully\n")
		return nil // Only need one Starknet client
	}

	fmt.Printf("⚠️  No Starknet networks found in config\n")
	return nil
}

// GetStarknetClient returns the Starknet client
func (sm *SolverManager) GetStarknetClient() (*rpc.Provider, error) {
	if sm.starknetClient == nil {
		return nil, fmt.Errorf("starknet client not initialized")
	}
	return sm.starknetClient, nil
}

// GetEVMClient returns an EVM client for the given chain ID
func (sm *SolverManager) GetEVMClient(chainID uint64) (*ethclient.Client, error) {
	if client, exists := sm.evmClients[chainID]; exists {
		return client, nil
	}
	return nil, fmt.Errorf("EVM client not found for chain ID %d", chainID)
}

// GetEVMSigner returns an EVM signer for the given chain ID
func (sm *SolverManager) GetEVMSigner(chainID uint64) (*bind.TransactOpts, error) {
	// For now, create a new signer for each chain
	// In the future, this could be cached per chain

	// Use conditional environment variable based on IS_DEVNET
	solverPrivateKey := envutil.GetConditionalAccountEnv("SOLVER_PRIVATE_KEY")
	if solverPrivateKey == "" {
		return nil, fmt.Errorf("SOLVER_PRIVATE_KEY environment variable not set")
	}

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(solverPrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse solver private key: %w", err)
	}

	from := crypto.PubkeyToAddress(pk.PublicKey)
	signer, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(int64(chainID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create signer with chain ID %d: %w", chainID, err)
	}
	signer.From = from
	return signer, nil
}

// GetStarknetSigner returns the Starknet signer
func (sm *SolverManager) GetStarknetSigner() (*account.Account, error) {
	// For now, create a new signer each time
	// In the future, this could be cached
	if sm.starknetClient == nil {
		return nil, fmt.Errorf("starknet client not initialized")
	}

	// Use conditional environment variables based on IS_DEVNET
	pub := envutil.GetStarknetSolverPublicKey()
	addrHex := envutil.GetStarknetSolverAddress()
	priv := envutil.GetStarknetSolverPrivateKey()

	if pub == "" || addrHex == "" || priv == "" {
		return nil, fmt.Errorf("missing STARKNET_SOLVER_* env vars for Starknet signer")
	}

	addrF, err := utils.HexToFelt(addrHex)
	if err != nil {
		return nil, fmt.Errorf("invalid STARKNET_SOLVER_ADDRESS: %w", err)
	}

	ks := account.NewMemKeystore()
	privBI, ok := new(big.Int).SetString(priv, 0)
	if !ok {
		return nil, fmt.Errorf("failed to parse STARKNET_SOLVER_PRIVATE_KEY")
	}
	ks.Put(pub, privBI)

	acct, err := account.NewAccount(sm.starknetClient, addrF, pub, ks, account.CairoV2)
	if err != nil {
		return nil, fmt.Errorf("failed to create Starknet account: %w", err)
	}

	return acct, nil
}

// initializeHyperlane7683 starts the Hyperlane 7683 solver
func (sm *SolverManager) initializeHyperlane7683(ctx context.Context) error {
	fmt.Printf("   🔧 Setting up Hyperlane7683 solver components...\n")

	// Create solver with client and signer getter functions
	hyperlane7683Solver := contracts.NewHyperlane7683Solver(
		sm.GetEVMClient,      // EVM client getter
		sm.GetStarknetClient, // Starknet client getter
		sm.GetEVMSigner,      // EVM signer getter
		sm.GetStarknetSigner, // Starknet signer getter
		sm.allowBlockLists,   // Allow/block lists
	)
	hyperlane7683Solver.AddDefaultRules()

	// Event handler that processes intents
	eventHandler := func(args types.ParsedArgs, originChainName string, blockNumber uint64) (bool, error) {
		return hyperlane7683Solver.ProcessIntent(ctx, &args)
	}

	// Start listeners for each intent source
	fmt.Printf("   📡 Starting network listeners...\n")
	listenerCount := 0

	for _, source := range []string{"Base", "Optimism", "Arbitrum", "Ethereum", "Starknet"} {
		networkConfig, exists := config.Networks[source]
		if !exists {
			fmt.Printf("     ⚠️  Network %s not found in config, skipping...\n", source)
			continue
		}

		var shutdown base.ShutdownFunc

		// Create appropriate listener based on chain type
		if source == "Starknet" {
			hyperlaneAddr, err := getStarknetHyperlaneAddress(&networkConfig)
			if err != nil {
				return fmt.Errorf("failed to get Starknet Hyperlane address: %w", err)
			}

			// Create Starknet listener config with original solver start block
			// The listener will handle negative value resolution
			listenerConfig := base.NewListenerConfig(
				hyperlaneAddr,
				source,
				big.NewInt(networkConfig.SolverStartBlock), // pass original value (can be negative)
				networkConfig.PollInterval,                 // poll interval from config
				networkConfig.ConfirmationBlocks,   // confirmation blocks from config
				networkConfig.MaxBlockRange,                // max block range from config
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
			// Create EVM listener config with original solver start block
			// The listener will handle negative value resolution
			listenerConfig := base.NewListenerConfig(
				networkConfig.HyperlaneAddress.Hex(),
				source,
				big.NewInt(networkConfig.SolverStartBlock), // pass original value (can be negative)
				networkConfig.PollInterval,                 // poll interval from config
				networkConfig.ConfirmationBlocks,   // confirmation blocks from config
				networkConfig.MaxBlockRange,                // max block range from config
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
		listenerCount++
		fmt.Printf("     ✅ Started listener for %s\n", source)
	}

	fmt.Printf("   📡 All network listeners started (%d networks)\n", listenerCount)
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

// Start initializes and runs all solvers
func (sm *SolverManager) Start(ctx context.Context) error {
	// Initialize all solvers
	if err := sm.InitializeSolvers(ctx); err != nil {
		return fmt.Errorf("failed to initialize solvers: %w", err)
	}

	// Wait for context cancellation (shutdown signal)
	<-ctx.Done()

	// Graceful shutdown
	sm.Shutdown()
	return nil
}

// Shutdown stops all active solvers
func (sm *SolverManager) Shutdown() {
	fmt.Printf("🛑 Shutting down solvers...\n")

	listenerCount := len(sm.activeShutdowns)
	for i, shutdown := range sm.activeShutdowns {
		fmt.Printf("   📡 Stopping listener %d/%d\n", i+1, listenerCount)
		shutdown()
	}

	sm.activeShutdowns = make([]func(), 0)
	fmt.Printf("✅ All solvers shut down successfully (%d listeners stopped)\n", listenerCount)
}

// GetSolverStatus returns the status of all solvers
func (sm *SolverManager) GetSolverStatus() map[string]bool {
	status := make(map[string]bool)
	for name, config := range sm.solverRegistry {
		status[name] = config.Enabled
	}
	return status
}


// getStarknetHyperlaneAddress gets the Starknet Hyperlane address from environment
func getStarknetHyperlaneAddress(_ *config.NetworkConfig) (string, error) {
	envAddr := envutil.GetEnvWithDefault("STARKNET_HYPERLANE_ADDRESS", "")
	if envAddr != "" {
		fmt.Printf("   🔄 Using Starknet Hyperlane address from .env: %s\n", envAddr)
		return envAddr, nil
	} else {
		return "", fmt.Errorf("no STARKNET_HYPERLANE_ADDRESS set in .env")
	}
}

//// getStarknetHyperlaneFromDeploymentState loads Starknet Hyperlane address from deployment state
// func getStarknetHyperlaneFromDeploymentState() string {
//	paths := []string{
//		"state/network_state/deployment-state.json",
//		"../state/network_state/deployment-state.json",
//		"../../state/network_state/deployment-state.json",
//	}
//	for _, path := range paths {
//		data, err := os.ReadFile(path)
//		if err != nil {
//			continue
//		}
//		var deploymentState struct {
//			Networks map[string]struct {
//				ChainID          uint64 `json:"chainId"`
//				HyperlaneAddress string `json:"hyperlaneAddress"`
//				DogCoinAddress   string `json:"dogCoinAddress"`
//			} `json:"networks"`
//		}
//		if err := json.Unmarshal(data, &deploymentState); err != nil {
//			continue
//		}
//		if stark, ok := deploymentState.Networks["Starknet"]; ok && stark.HyperlaneAddress != "" {
//			return stark.HyperlaneAddress
//		}
//		if starkLegacy, ok := deploymentState.Networks["Starknet"]; ok && starkLegacy.HyperlaneAddress != "" {
//			return starkLegacy.HyperlaneAddress
//		}
//	}
//	return ""
//}
