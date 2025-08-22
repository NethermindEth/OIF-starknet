package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NetworkConfig struct {
	Name       string
	RPCURL     string
	PrivateKey string
}

func main() {
	// Get RPC URLs from environment variables with defaults
	networks := []NetworkConfig{
		{
			Name:       "Sepolia",
			RPCURL:     getEnvWithDefault("SEPOLIA_RPC_URL", "http://localhost:8545"),
			PrivateKey: os.Getenv("PRIVATE_KEY"),
		},
		{
			Name:       "Optimism Sepolia",
			RPCURL:     getEnvWithDefault("OPTIMISM_RPC_URL", "http://localhost:8546"),
			PrivateKey: os.Getenv("PRIVATE_KEY"),
		},
		{
			Name:       "Arbitrum Sepolia",
			RPCURL:     getEnvWithDefault("ARBITRUM_RPC_URL", "http://localhost:8547"),
			PrivateKey: os.Getenv("PRIVATE_KEY"),
		},
		{
			Name:       "Base Sepolia",
			RPCURL:     getEnvWithDefault("BASE_RPC_URL", "http://localhost:8548"),
			PrivateKey: os.Getenv("PRIVATE_KEY"),
		},
	}

	// Get Hyperlane address from environment
	hyperlaneAddress := getEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")

	for _, network := range networks {
		fmt.Printf("🔍 Verifying Hyperlane7683 on %s...\n", network.Name)
		fmt.Printf("   RPC URL: %s\n", network.RPCURL)
		fmt.Printf("   Contract Address: %s\n", hyperlaneAddress)

		client, err := ethclient.Dial(network.RPCURL)
		if err != nil {
			fmt.Printf("   ❌ Failed to connect: %v\n", err)
			continue
		}

		// Check if contract exists
		code, err := client.CodeAt(context.Background(), common.HexToAddress(hyperlaneAddress), nil)
		if err != nil {
			fmt.Printf("   ❌ Failed to get contract code: %v\n", err)
			continue
		}

		if len(code) == 0 {
			fmt.Printf("   ❌ Contract not deployed or no code at address\n")
		} else {
			fmt.Printf("   ✅ Contract deployed")
		}

		client.Close()
	}
}

// getEnvWithDefault gets an environment variable with a default fallback
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
