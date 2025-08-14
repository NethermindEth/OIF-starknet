package internal

import (
	"context"
	"fmt"
	"sync"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/filler"
	"github.com/NethermindEth/oif-starknet/go/internal/listener"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/sirupsen/logrus"
)

// SolverModule represents a complete solver with listener and filler
type SolverModule struct {
	Name     string
	Listener listener.BaseListener
	Filler   filler.BaseFiller
	Config   *listener.ListenerConfig
}

// SolverManager manages multiple solvers
type SolverManager struct {
	config     *config.Config
	logger     interface{}
	solvers    map[string]*SolverModule
	shutdownWg sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewSolverManager creates a new solver manager
func NewSolverManager(cfg *config.Config, logger interface{}) *SolverManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &SolverManager{
		config:  cfg,
		logger:  logger,
		solvers: make(map[string]*SolverModule),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// InitializeSolvers initializes all enabled solvers
func (sm *SolverManager) InitializeSolvers() error {
	fmt.Printf("⚙️  Initializing solvers...\n")

	for solverName, solverConfig := range sm.config.Solvers {
		if !solverConfig.Enabled {
			fmt.Printf("⏭️  Solver %s is disabled, skipping...\n", solverName)
			continue
		}

		if err := sm.initializeSolver(solverName); err != nil {
			fmt.Printf("❌ Failed to initialize solver %s: %v\n", solverName, err)
			return fmt.Errorf("failed to initialize solver %s: %w", solverName, err)
		}
	}

	fmt.Printf("✅ All solvers initialized successfully\n")
	return nil
}

// initializeSolver initializes a single solver
func (sm *SolverManager) initializeSolver(name string) error {
	fmt.Printf("⚙️  Initializing solver: %s...\n", name)

	// Create solver module based on name
	solver, err := sm.createSolverModule(name)
	if err != nil {
		return fmt.Errorf("failed to create solver module: %w", err)
	}

	// Store the solver
	sm.solvers[name] = solver

	// Start the solver
	if err := sm.startSolver(solver); err != nil {
		return fmt.Errorf("failed to start solver: %w", err)
	}

	fmt.Printf("✅ Solver %s initialized and started\n", name)
	return nil
}

// createSolverModule creates a solver module based on the name
func (sm *SolverManager) createSolverModule(name string) (*SolverModule, error) {
	switch name {
	case "hyperlane7683":
		return sm.createHyperlane7683Solver()
	default:
		return nil, fmt.Errorf("unknown solver type: %s", name)
	}
}

// createHyperlane7683Solver creates the Hyperlane7683 solver
func (sm *SolverManager) createHyperlane7683Solver() (*SolverModule, error) {
	// Create metadata for Hyperlane7683
	metadata := types.Hyperlane7683Metadata{
		BaseMetadata: types.BaseMetadata{
			ProtocolName: "Hyperlane7683",
		},
		IntentSources: []types.IntentSource{
			{
				Address:            config.Networks["Base Sepolia"].HyperlaneAddress.Hex(),
				ChainName:          "Base Sepolia",
				InitialBlock:       nil,  // Start from current block
				PollInterval:       1000, // 1 second
				ConfirmationBlocks: 2,
			},
		},
		CustomRules: types.CustomRules{
			Rules: []types.RuleConfig{
				{
					Name: "filterByTokenAndAmount",
				},
				{
					Name: "intentNotFilled",
				},
			},
		},
	}

	// Create allow/block lists (empty for now - allow everything)
	allowBlockLists := types.AllowBlockLists{
		AllowList: []types.AllowBlockListItem{},
		BlockList: []types.AllowBlockListItem{},
	}

	// Get deployment state to create listeners for all networks
	state, err := deployer.GetDeploymentState()
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment state: %w", err)
	}

	// Create a multi-network listener that listens to all networks
	multiListener := listener.NewMultiNetworkListener(state, sm.logger.(*logrus.Logger))

	hyperlaneFiller := filler.NewHyperlane7683Filler(allowBlockLists, metadata)

	// Add default rules
	hyperlaneFiller.AddDefaultRules()

	return &SolverModule{
		Name:     "hyperlane7683",
		Listener: multiListener,
		Filler:   hyperlaneFiller,
		Config: &listener.ListenerConfig{
			ContractAddress:    config.Networks["Base Sepolia"].HyperlaneAddress.Hex(),
			ChainName:          "Multi-Network",
			InitialBlock:       nil,
			PollInterval:       1000,
			ConfirmationBlocks: 2,
			MaxBlockRange:      500,
		},
	}, nil
}

// getRPCURLForChain returns the RPC URL for a given chain name
func (sm *SolverManager) getRPCURLForChain(chainName string) string {
	rpcURL, err := config.GetRPCURL(chainName)
	if err != nil {
		fmt.Printf("⚠️  Failed to get RPC URL for chain %s, using default: %v\n", chainName, err)
		return config.GetDefaultRPCURL()
	}
	return rpcURL
}

// startSolver starts a solver and begins listening for events
func (sm *SolverManager) startSolver(solver *SolverModule) error {
	if solver.Listener == nil {
		return fmt.Errorf("solver %s has no listener configured", solver.Name)
	}

	// Create event handler
	handler := func(args types.ParsedArgs, originChainName string, blockNumber uint64) error {
		fmt.Printf("🔵 Processing intent: solver=%s, orderID=%s, chain=%s, block=%d\n", 
			solver.Name, args.OrderID, originChainName, blockNumber)

		// Process the intent through the filler
		if err := solver.Filler.ProcessIntent(sm.ctx, args, originChainName, blockNumber); err != nil {
					fmt.Printf("❌ Failed to process intent: solver=%s, orderID=%s, error=%v\n", 
			solver.Name, args.OrderID, err)
			return err
		}

		// TODO: Don't update LastIndexedBlock here - we only update after fully processing all events in a block
		// This ensures we don't skip blocks with unprocessed events
		// LastIndexedBlock will be updated by the listener when MarkBlockFullyProcessed is called
		// 
		// The flow should be:
		// 1. Event detected → ProcessIntent called → Event processed/filled
		// 2. After ALL events in a block are processed → MarkBlockFullyProcessed called
		// 3. Only then is LastIndexedBlock updated

		fmt.Printf("✅ Intent processed successfully: solver=%s, orderID=%s\n", 
			solver.Name, args.OrderID)

		return nil
	}

	// Start the listener
	shutdownFunc, err := solver.Listener.Start(sm.ctx, handler)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	// Store shutdown function for cleanup
	sm.shutdownWg.Add(1)
	go func() {
		defer sm.shutdownWg.Done()
		<-sm.ctx.Done()
		shutdownFunc()
	}()

	return nil
}

// Shutdown gracefully shuts down all solvers
func (sm *SolverManager) Shutdown() {
	fmt.Printf("🔄 Shutting down solvers...\n")

	// Cancel context to stop all goroutines
	sm.cancel()

	// Wait for all solvers to shut down
	sm.shutdownWg.Wait()

	fmt.Printf("✅ All solvers shut down successfully\n")
}

// GetSolver returns a solver by name
func (sm *SolverManager) GetSolver(name string) (*SolverModule, bool) {
	solver, exists := sm.solvers[name]
	return solver, exists
}

// GetSolvers returns all solvers
func (sm *SolverManager) GetSolvers() map[string]*SolverModule {
	return sm.solvers
}

// MarkBlockFullyProcessed marks a block as fully processed across all solvers
// This should be called after all events in a block have been processed/filled
func (sm *SolverManager) MarkBlockFullyProcessed(chainName string, blockNumber uint64) error {
	fmt.Printf("🔵 Marking block as fully processed: chain=%s, block=%d\n", 
		chainName, blockNumber)
	
	// Update the deployment state with the last indexed block
	if err := deployer.UpdateLastIndexedBlock(chainName, blockNumber); err != nil {
		return fmt.Errorf("failed to update last indexed block for %s: %w", chainName, err)
	}
	
	fmt.Printf("✅ Block marked as fully processed and LastIndexedBlock updated: chain=%s, block=%d\n", 
		chainName, blockNumber)
	
	return nil
}
