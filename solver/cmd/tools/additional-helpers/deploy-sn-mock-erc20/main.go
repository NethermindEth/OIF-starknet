package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
)

// Token deployment configuration
const (
	// Default class hash file path (local go/state/deployment)
	DeclarationFilePath = "state/deployment/starknet-mock-erc20-declaration.json"
	// File permission constants
	deploymentDirPerms  = 0700
	deploymentFilePerms = 0600
)

// DeclarationInfo represents the structure of the declaration file
type DeclarationInfo struct {
	ClassHash       string `json:"classHash"`
	DeclarationTime string `json:"declarationTime"`
	NetworkName     string `json:"networkName"`
	TransactionHash string `json:"transactionHash"`
}

// TokenInfo represents a deployed token
type TokenInfo struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Address   string `json:"address"`
	ClassHash string `json:"classHash"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	fmt.Println("🚀 Deploying MockERC20 tokens to Starknet...")

	// Load environment variables
	networkName := "Starknet"

	// Get network configuration
	networkConfig, err := config.GetNetworkConfig(networkName)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to get network config for %s: %s", networkName, err))
	}

	// Load Starknet account details from .env
	deployerAddress := os.Getenv("STARKNET_DEPLOYER_ADDRESS")
	deployerPrivateKey := os.Getenv("STARKNET_DEPLOYER_PRIVATE_KEY")
	deployerPublicKey := os.Getenv("STARKNET_DEPLOYER_PUBLIC_KEY")

	if deployerAddress == "" || deployerPrivateKey == "" || deployerPublicKey == "" {
		fmt.Println("❌ Missing required environment variables:")
		fmt.Println("   STARKNET_DEPLOYER_ADDRESS: Your Starknet account address")
		fmt.Println("   STARKNET_DEPLOYER_PRIVATE_KEY: Your private key")
		fmt.Println("   STARKNET_DEPLOYER_PUBLIC_KEY: Your public key")
		os.Exit(1)
	}

	fmt.Printf("📋 Network: %s\n", networkName)
	fmt.Printf("📋 RPC URL: %s\n", networkConfig.RPCURL)
	fmt.Printf("📋 Chain ID: %d\n", networkConfig.ChainID)
	fmt.Printf("📋 Deployer: %s\n", deployerAddress)

	// Initialize connection to RPC provider
	client, err := rpc.NewProvider(networkConfig.RPCURL)
	if err != nil {
		panic(fmt.Sprintf("❌ Error connecting to RPC provider: %s", err))
	}

	// Convert account address to felt
	accountAddressFelt, err := utils.HexToFelt(deployerAddress)
	if err != nil {
		panic(fmt.Sprintf("❌ Invalid account address: %s", err))
	}

	// Initialize the account memkeyStore
	ks := account.NewMemKeystore()
	privKeyBI, ok := new(big.Int).SetString(deployerPrivateKey, 0)
	if !ok {
		panic("❌ Failed to convert private key to big.Int")
	}
	ks.Put(deployerPublicKey, privKeyBI)

	fmt.Println("✅ Connected to Starknet RPC")

	// Initialize the account (Cairo v2)
	accnt, err := account.NewAccount(client, accountAddressFelt, deployerPublicKey, ks, account.CairoV2)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to initialize account: %s", err))
	}

	// Get class hash from declaration file or environment variable
	classHash, err := getClassHash()
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to get class hash: %s", err))
	}

	// Convert class hash to felt
	classHashFelt, err := utils.HexToFelt(classHash)
	if err != nil {
		panic(fmt.Sprintf("❌ Invalid class hash: %s", err))
	}

	// Deploy DogCoin (destination chain token)
	fmt.Println("\n🪙 Deploying DogCoin...")
	dogCoinAddress, err := deployMockERC20(accnt, classHashFelt, "DogCoin", "DOG")
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to deploy DogCoin: %s", err))
	}
	fmt.Printf("✅ DogCoin deployed at: %s\n", dogCoinAddress)

	// Note: .env file updates removed - addresses should be set manually after live deployment

	// Save deployment info
	tokens := []TokenInfo{
		{Name: "DogCoin", Symbol: "DOG", Address: dogCoinAddress, ClassHash: classHash},
	}
	saveDeploymentInfo(tokens, networkName)

	fmt.Printf("\n🎯 MockERC20 tokens deployed successfully!\n")
	fmt.Printf("   • DogCoin: %s\n", dogCoinAddress)
	fmt.Printf("   • Ready for funding and approval setup!\n")
}

// deployMockERC20 deploys a single mock ERC20 token
func deployMockERC20(accnt *account.Account, classHashFelt *felt.Felt, tokenName, tokenSymbol string) (string, error) {
	fmt.Printf("   📝 Deploying %s (%s)...\n", tokenName, tokenSymbol)

	// MockERC20 constructor takes: name, symbol
	// Convert name and symbol to felt arrays (Cairo strings)
	nameFelt, err := utils.StringToByteArrFelt(tokenName)
	if err != nil {
		return "", fmt.Errorf("failed to convert name to felt: %w", err)
	}

	symbolFelt, err := utils.StringToByteArrFelt(tokenSymbol)
	if err != nil {
		return "", fmt.Errorf("failed to convert symbol to felt: %w", err)
	}

	// Build constructor calldata: [name_bytes..., symbol_bytes...]
	constructorCalldata := make([]*felt.Felt, 0, len(nameFelt)+len(symbolFelt))
	constructorCalldata = append(constructorCalldata, nameFelt...)
	constructorCalldata = append(constructorCalldata, symbolFelt...)

	fmt.Printf("   📋 Constructor calldata: name='%s', symbol='%s'\n", tokenName, tokenSymbol)

	fmt.Printf("   📤 Sending deployment transaction...\n")

	// Deploy the contract with UDC using the modern approach
	resp, salt, err := accnt.DeployContractWithUDC(context.Background(), classHashFelt, constructorCalldata, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to deploy contract: %w", err)
	}

	// Extract transaction hash from response
	txHash := resp.Hash
	fmt.Printf("   ⏳ Transaction sent! Hash: %s\n", txHash.String())
	fmt.Printf("   ⏳ Waiting for transaction confirmation...\n")

	// Wait for transaction receipt
	txReceipt, err := accnt.WaitForTransactionReceipt(context.Background(), txHash, time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to wait for transaction receipt: %w", err)
	}

	fmt.Printf("   ✅ Deployment completed!\n")
	fmt.Printf("   📋 Transaction Hash: %s\n", txHash.String())
	fmt.Printf("   📋 Execution Status: %s\n", txReceipt.ExecutionStatus)
	fmt.Printf("   📋 Finality Status: %s\n", txReceipt.FinalityStatus)

	// Compute the deployed contract address
	deployedAddress := utils.PrecomputeAddressForUDC(classHashFelt, salt, constructorCalldata, utils.UDCCairoV0, accnt.Address)
	fmt.Printf("   🏗️  Contract deployed at: %s\n", deployedAddress.String())

	return deployedAddress.String(), nil
}

// getClassHash retrieves the class hash from declaration file or environment variable
func getClassHash() (string, error) {
	// First try to get from environment variable
	if envClassHash := os.Getenv("MOCK_ERC20_CLASS_HASH"); envClassHash != "" {
		fmt.Printf("📋 Using class hash from environment variable: %s\n", envClassHash)
		return envClassHash, nil
	}

	// Try to read from declaration file in deployment directory, with fallback to legacy path
	deploymentDir := filepath.Clean(filepath.Join("state", "deployment"))
	declarationFile := filepath.Join(deploymentDir, "starknet-mock-erc20-declaration.json")

	// Check if declaration file exists
	if _, err := os.Stat(declarationFile); os.IsNotExist(err) {
		// Try the default declaration file
		if _, err := os.Stat(DeclarationFilePath); err == nil {
			declarationFile = DeclarationFilePath
		} else {
			legacy := "mock_erc20_declaration_starknet.json"
			if _, err := os.Stat(legacy); err == nil {
				declarationFile = legacy
			} else {
				return "", fmt.Errorf("no declaration file found and MOCK_ERC20_CLASS_HASH not set. Expected one of: %s, %s, %s", declarationFile, DeclarationFilePath, legacy)
			}
		}
	}

	// Read and parse declaration file
	data, err := os.ReadFile(declarationFile)
	if err != nil {
		return "", fmt.Errorf("failed to read declaration file %s: %w", declarationFile, err)
	}

	var declaration DeclarationInfo
	if err := json.Unmarshal(data, &declaration); err != nil {
		return "", fmt.Errorf("failed to parse declaration file %s: %w", declarationFile, err)
	}

	if declaration.ClassHash == "" {
		return "", fmt.Errorf("class hash not found in declaration file %s", declarationFile)
	}

	fmt.Printf("📋 Using class hash from declaration file %s: %s\n", declarationFile, declaration.ClassHash)
	return declaration.ClassHash, nil
}

// saveDeploymentInfo saves deployment information to a file
func saveDeploymentInfo(tokens []TokenInfo, networkName string) {
	deploymentInfo := map[string]interface{}{
		"networkName":    networkName,
		"deploymentTime": time.Now().Format(time.RFC3339),
		"tokens":         tokens,
	}

	data, err := json.MarshalIndent(deploymentInfo, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Failed to marshal deployment info: %s\n", err)
		return
	}

	// Ensure deployment directory exists
	deploymentDir := filepath.Clean(filepath.Join("state", "deployment"))
	if err := os.MkdirAll(deploymentDir, deploymentDirPerms); err != nil {
		fmt.Printf("⚠️  Failed to create deployment directory: %s\n", err)
		return
	}

	filename := filepath.Join(deploymentDir, fmt.Sprintf("%s-mock-erc20-deployment.json", sanitizeNetworkName(networkName)))
	if err := os.WriteFile(filename, data, deploymentFilePerms); err != nil {
		fmt.Printf("⚠️  Failed to save deployment info: %s\n", err)
		return
	}

	fmt.Printf("💾 Deployment info saved to %s\n", filename)
}

// sanitizeNetworkName converts a human network name to a safe slug
func sanitizeNetworkName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
