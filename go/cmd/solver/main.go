package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/NethermindEth/oif-starknet/go/internal"
	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

// Custom formatter that outputs only the message
type cleanFormatter struct{}

func (f *cleanFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return append([]byte(entry.Message), '\n'), nil
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	// Setup logging
	logger := setupLogger(cfg)
	logger.Info("🙍 Intent Solver 📝")
	
	// Debug: Show what start blocks are being used
	logger.Info("📊 Solver Start Blocks from .env:")
	for networkName, networkConfig := range config.Networks {
		startBlock := networkConfig.SolverStartBlock
		logger.Infof("   %s: block %d", networkName, startBlock)
		
		// Warn about potentially old blocks that will cause expensive backfilling
		if startBlock > 0 {
			isOldBlock := false
			switch networkName {
			case "Ethereum":
				isOldBlock = startBlock < 20000000 // Warn if older than ~2024
			case "Optimism": 
				isOldBlock = startBlock < 28000000 // Recent OP Sepolia blocks
			case "Arbitrum":
				isOldBlock = startBlock < 140000000 // Recent Arb Sepolia blocks
			case "Base":
				isOldBlock = startBlock < 27000000 // Recent Base Sepolia blocks  
			case "Starknet":
				isOldBlock = startBlock < 1600000 // Recent Starknet blocks
			}
			
			if isOldBlock {
				logger.Warnf("   ⚠️  %s start block %d seems old - may cause expensive backfilling!", networkName, startBlock)
				logger.Warnf("   💡 Consider setting %s_SOLVER_START_BLOCK=0 in .env to start from latest", strings.ToUpper(networkName))
			}
		}
	}



	// Create ethereum client - use any EVM chain from config for the base client
	// The solver manager will create specific clients for each chain as needed
	var ethClient *ethclient.Client
	for chainName, network := range config.Networks {
		if chainName != "Starknet" { // Use any EVM chain
			ethClient, err = ethclient.Dial(network.RPCURL)
			if err != nil {
				logger.Fatalf("Failed to connect to %s at %s: %v", chainName, network.RPCURL, err)
			}
			logger.Infof("📡 Connected to %s", chainName)
			break
		}
	}
	
	if ethClient == nil {
		logger.Fatalf("❌ No EVM chains found in config")
	}

	// Create solver manager
	solverManager := internal.NewSolverManager(ethClient)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create context for initialization
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize solvers
	if err := solverManager.InitializeSolvers(ctx); err != nil {
		logger.Fatalf("❌ Failed to initialize solvers: %v", err)
	}

	// Wait for shutdown signal
	<-sigChan
	logger.Info("🔄 Received shutdown signal, shutting down...")

	// Cancel context and shutdown gracefully
	cancel()
	solverManager.Shutdown()
	logger.Info("✅ Solver shutdown complete")
}

// setupLogger configures the logger based on configuration
func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Warnf("Invalid log level %s, using info: %v", cfg.LogLevel, err)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.LogFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		// Custom formatter that outputs only the message text
		logger.SetFormatter(&cleanFormatter{})
	}

	return logger
}
