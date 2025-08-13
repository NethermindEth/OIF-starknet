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
	Name    string
	Listener listener.BaseListener
	Filler   filler.BaseFiller
	Config   *listener.ListenerConfig
}

// SolverManager manages multiple solvers
type SolverManager struct {
	config     *config.Config
	logger     *logrus.Logger
	solvers    map[string]*SolverModule
	shutdownWg sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewSolverManager creates a new solver manager
func NewSolverManager(cfg *config.Config, logger *logrus.Logger) *SolverManager {
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
	sm.logger.Info("Initializing solvers...")
	
	for solverName, solverConfig := range sm.config.Solvers {
		if !solverConfig.Enabled {
			sm.logger.Infof("Solver %s is disabled, skipping...", solverName)
			continue
		}
		
		if err := sm.initializeSolver(solverName); err != nil {
			sm.logger.Errorf("Failed to initialize solver %s: %v", solverName, err)
			return fmt.Errorf("failed to initialize solver %s: %w", solverName, err)
		}
	}
	
	sm.logger.Info("All solvers initialized successfully")
	return nil
}

// initializeSolver initializes a single solver
func (sm *SolverManager) initializeSolver(name string) error {
	sm.logger.Infof("Initializing solver: %s", name)
	
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
	
	sm.logger.Infof("Solver %s initialized and started", name)
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
				Address:         "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3", // Testnet address
				ChainName:       "Base Sepolia",
				InitialBlock:    nil, // Start from current block
				PollInterval:    1000, // 1 second
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
	multiListener := listener.NewMultiNetworkListener(state, sm.logger)
	
	hyperlaneFiller := filler.NewHyperlane7683Filler(allowBlockLists, metadata)
	
	// Add default rules
	hyperlaneFiller.AddDefaultRules()
	
	return &SolverModule{
		Name:     "hyperlane7683",
		Listener: multiListener,
		Filler:   hyperlaneFiller,
		Config:   &listener.ListenerConfig{
			ContractAddress:    "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
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
	// Map chain names to RPC URLs for our local testnet setup
	switch chainName {
	case "Base Sepolia":
		return "http://localhost:8548"
	case "Sepolia":
		return "http://localhost:8545"
	case "Optimism Sepolia":
		return "http://localhost:8546"
	case "Arbitrum Sepolia":
		return "http://localhost:8547"
	default:
		// Default to Base Sepolia for now
		return "http://localhost:8548"
	}
}

// startSolver starts a solver and begins listening for events
func (sm *SolverManager) startSolver(solver *SolverModule) error {
	if solver.Listener == nil {
		return fmt.Errorf("solver %s has no listener configured", solver.Name)
	}
	
	// Create event handler
	handler := func(args types.ParsedArgs, originChainName string, blockNumber uint64) error {
		sm.logger.WithFields(logrus.Fields{
			"solver":        solver.Name,
			"orderID":       args.OrderID,
			"originChain":   originChainName,
			"blockNumber":   blockNumber,
		}).Info("Processing intent")
		
		// Process the intent through the filler
		if err := solver.Filler.ProcessIntent(sm.ctx, args, originChainName, blockNumber); err != nil {
			sm.logger.WithFields(logrus.Fields{
				"solver":      solver.Name,
				"orderID":     args.OrderID,
				"error":       err,
			}).Error("Failed to process intent")
			return err
		}
		
		// Update the deployment state with the last indexed block
		if err := deployer.UpdateLastIndexedBlock(originChainName, blockNumber); err != nil {
			sm.logger.Warnf("Failed to update last indexed block for %s: %v", originChainName, err)
		}
		
		sm.logger.WithFields(logrus.Fields{
			"solver":    solver.Name,
			"orderID":   args.OrderID,
		}).Info("Intent processed successfully")
		
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
	sm.logger.Info("Shutting down solvers...")
	
	// Cancel context to stop all goroutines
	sm.cancel()
	
	// Wait for all solvers to shut down
	sm.shutdownWg.Wait()
	
	sm.logger.Info("All solvers shut down successfully")
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
