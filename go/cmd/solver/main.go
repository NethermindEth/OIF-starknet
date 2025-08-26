package main

import (
	"context"
	"os"
	"os/signal"
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

	// Setup logging
	logger := setupLogger(cfg)
	logger.Info("🙍 Intent Solver 📝")

	// Create ethereum client - use any EVM chain from config for the base client
	// The solver manager will create specific clients for each chain as needed
	var ethClient *ethclient.Client
	for chainName, network := range config.Networks {
		if chainName != "Starknet Sepolia" { // Use any EVM chain
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
