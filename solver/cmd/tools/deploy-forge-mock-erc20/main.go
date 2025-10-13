package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

// NetworkInfo contains deployment information for each network
type NetworkInfo struct {
	Name    string
	ChainID string
	EnvVar  string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Define networks for deployment
	networks := []NetworkInfo{
		{
			Name:    "Ethereum Sepolia",
			ChainID: "11155111",
			EnvVar:  "ETHEREUM_DOG_COIN_ADDRESS",
		},
		{
			Name:    "Optimism Sepolia",
			ChainID: "11155420",
			EnvVar:  "OPTIMISM_DOG_COIN_ADDRESS",
		},
		{
			Name:    "Arbitrum Sepolia",
			ChainID: "421614",
			EnvVar:  "ARBITRUM_DOG_COIN_ADDRESS",
		},
		{
			Name:    "Base Sepolia",
			ChainID: "84532",
			EnvVar:  "BASE_DOG_COIN_ADDRESS",
		},
	}

	// Check if specific network was requested
	var targetNetworks []NetworkInfo
	if len(os.Args) > 1 {
		networkName := strings.ToLower(os.Args[1])
		found := false
		for _, network := range networks {
			if strings.Contains(strings.ToLower(network.Name), networkName) {
				targetNetworks = []NetworkInfo{network}
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("Invalid network name: %s. Available: ethereum, optimism, arbitrum, base", os.Args[1])
		}
	} else {
		targetNetworks = networks
	}

	fmt.Printf("🚀 Deploying MockERC20 with Forge to %d network(s)...\n", len(targetNetworks))
	fmt.Printf("   These will have matching compiler settings for verification!\n\n")

	successCount := 0
	deployedAddresses := make([]string, 0, len(targetNetworks))

	for _, network := range targetNetworks {
		fmt.Printf("📡 Deploying to %s (Chain ID: %s)...\n", network.Name, network.ChainID)

		address, err := deployWithForge(network.ChainID)
		if err != nil {
			fmt.Printf("   ❌ Failed to deploy: %v\n\n", err)
			continue
		}

		fmt.Printf("   ✅ Deployed at: %s\n", address)
		fmt.Printf("   🔗 Explorer: %s\n", getExplorerURL(network.ChainID, address))
		fmt.Printf("   📝 Update .env: %s=%s\n\n", network.EnvVar, address)

		deployedAddresses = append(deployedAddresses, fmt.Sprintf("%s=%s", network.EnvVar, address))
		successCount++
	}

	fmt.Printf("🎯 Deployment Summary:\n")
	fmt.Printf("   ✅ Successful: %d/%d\n", successCount, len(targetNetworks))

	if len(deployedAddresses) > 0 {
		fmt.Printf("\n📝 Update your .env file with these new addresses:\n")
		for _, addr := range deployedAddresses {
			fmt.Printf("   %s\n", addr)
		}
		fmt.Printf("\n🔍 Then run: make verify-mock-erc20\n")
		fmt.Printf("   These contracts will verify successfully!\n")
	}
}

func deployWithForge(chainID string) (string, error) {
	// Change to solidity directory
	solidityDir := "../solidity"

	// Get RPC URL based on chain ID
	rpcURL := getRPCURL(chainID)
	if rpcURL == "" {
		return "", fmt.Errorf("no RPC URL configured for chain ID %s", chainID)
	}

	// Run forge script with broadcast and verify
	cmd := exec.Command("forge", "script",
		"script/DeployMockERC20.s.sol:DeployMockERC20",
		"--rpc-url", rpcURL,
		"--chain", chainID,
		"--broadcast",
		"--verify",
		"--slow",
		"-v",
	)

	cmd.Dir = solidityDir

	// Capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return "", fmt.Errorf("forge deployment failed: %v\nOutput: %s", err, outputStr)
	}

	// Extract deployed address from output
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "MockERC20 deployed at:") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("could not extract deployed address from output")
}

func getRPCURL(chainID string) string {
	switch chainID {
	case "11155111": // Sepolia
		return os.Getenv("ETHEREUM_RPC_URL")
	case "11155420": // OP Sepolia
		return os.Getenv("OPTIMISM_RPC_URL")
	case "421614": // Arbitrum Sepolia
		return os.Getenv("ARBITRUM_RPC_URL")
	case "84532": // Base Sepolia
		return os.Getenv("BASE_RPC_URL")
	default:
		return ""
	}
}

func getExplorerURL(chainID, address string) string {
	switch chainID {
	case "11155111": // Sepolia
		return fmt.Sprintf("https://sepolia.etherscan.io/address/%s", address)
	case "11155420": // OP Sepolia
		return fmt.Sprintf("https://sepolia-optimism.etherscan.io/address/%s", address)
	case "421614": // Arbitrum Sepolia
		return fmt.Sprintf("https://sepolia.arbiscan.io/address/%s", address)
	case "84532": // Base Sepolia
		return fmt.Sprintf("https://sepolia.basescan.org/address/%s", address)
	default:
		return fmt.Sprintf("Unknown chain ID: %s", chainID)
	}
}
