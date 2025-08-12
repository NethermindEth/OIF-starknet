package main

import (
	"log"
	"os"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Warning: .env file not found: %v", err)
	}

	// Load configuration
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Printf("🔍 Verifying pre-deployed Hyperlane7683 contracts on forked networks...")

	// Create fork verification configurations
	forkConfigs := []deployer.ForkVerificationConfig{
		{
			RPCURL:     "http://localhost:8545",
			PrivateKey: os.Getenv("PRIVATE_KEY"),
			ChainName:  "ethereum",
			ChainID:    11155111, // Sepolia fork
		},
		{
			RPCURL:     "http://localhost:8546",
			PrivateKey: os.Getenv("PRIVATE_KEY"),
			ChainName:  "optimism",
			ChainID:    11155420, // Optimism Sepolia fork
		},
		{
			RPCURL:     "http://localhost:8547",
			PrivateKey: os.Getenv("PRIVATE_KEY"),
			ChainName:  "arbitrum",
			ChainID:    421614, // Arbitrum sepolia fork
		},
		{
			RPCURL:     "http://localhost:8548",
			PrivateKey: os.Getenv("PRIVATE_KEY"),
			ChainName:  "base",
			ChainID:    84532, // Base mainnet fork
		},
	}

	// Create fork verification manager
	forkManager := deployer.NewForkVerificationManager(forkConfigs)

	// Verify pre-deployed contracts
	if err := forkManager.VerifyPreDeployedContracts(); err != nil {
		log.Fatalf("❌ Failed to verify contracts: %v", err)
	}

	// Get contract addresses
	addresses := forkManager.GetContractAddresses()

	log.Printf("📋 Contract addresses:")
	for chainName, address := range addresses {
		log.Printf("   %s: %s", chainName, address.Hex())
	}

	log.Printf("🎉 All pre-deployed contracts verified successfully!")
	log.Printf("💡 These contracts are ready to use for intent solving!")
}
