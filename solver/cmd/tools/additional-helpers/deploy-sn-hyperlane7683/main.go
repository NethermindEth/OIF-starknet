package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
)

const (
	// File permissions
	deploymentDirPerms = 0700
	deploymentFilePerms = 0600
)

// DeclarationInfo represents the structure of the declaration file
type DeclarationInfo struct {
	ClassHash       string `json:"classHash"`
	DeclarationTime string `json:"declarationTime"`
	NetworkName     string `json:"networkName"`
	TransactionHash string `json:"transactionHash"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	fmt.Println("🚀 Deploying Hyperlane7683 contract to Starknet...")

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

	// Load constructor parameters from environment variables
	permit2Addr := os.Getenv("STARKNET_PERMIT2_ADDRESS")
	mailboxAddr := os.Getenv("STARKNET_MAILBOX_ADDRESS")
	hookAddr := os.Getenv("STARKNET_HOOK_ADDRESS")
	ismAddr := os.Getenv("STARKNET_ISM_ADDRESS")

	if deployerAddress == "" || deployerPrivateKey == "" || deployerPublicKey == "" {
		fmt.Println("❌ Missing required environment variables:")
		fmt.Println("   STARKNET_DEPLOYER_ADDRESS: Your Starknet account address")
		fmt.Println("   STARKNET_DEPLOYER_PRIVATE_KEY: Your private key")
		fmt.Println("   STARKNET_DEPLOYER_PUBLIC_KEY: Your public key")
		fmt.Println("")
		fmt.Println("Optional constructor parameters (will use 0x0 if not provided):")
		fmt.Println("   STARKNET_PERMIT2_ADDRESS: Permit2 contract address")
		fmt.Println("   STARKNET_MAILBOX_ADDRESS: Mailbox contract address")
		fmt.Println("   STARKNET_HOOK_ADDRESS: Hook contract address")
		fmt.Println("   STARKNET_ISM_ADDRESS: ISM contract address")
		os.Exit(1)
	}

	if permit2Addr == "" || mailboxAddr == "" || hookAddr == "" || ismAddr == "" {
		fmt.Println("❌ Missing required environment variables:")
		fmt.Println("   STARKNET_PERMIT2_ADDRESS: Permit2 contract address")
		fmt.Println("   STARKNET_MAILBOX_ADDRESS: Mailbox contract address")
		fmt.Println("   STARKNET_HOOK_ADDRESS: Hook contract address")
		fmt.Println("   STARKNET_ISM_ADDRESS: ISM contract address")
		os.Exit(1)
	}

	// Get class hash from declaration file or environment variable
	classHash, err := getClassHash()
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to get class hash: %s", err))
	}

	fmt.Printf("📋 Network: %s\n", networkName)
	fmt.Printf("📋 RPC URL: %s\n", networkConfig.RPCURL)
	fmt.Printf("📋 Chain ID: %d\n", networkConfig.ChainID)
	fmt.Printf("📋 Deployer: %s\n", deployerAddress)
	fmt.Printf("📋 Contract Class Hash: %s\n", classHash)

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

	// Initialize the account (Cairo v1)
	accnt, err := account.NewAccount(client, accountAddressFelt, deployerPublicKey, ks, account.CairoV2)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to initialize account: %s", err))
	}

	// Convert class hash to felt
	classHashFelt, err := utils.HexToFelt(classHash)
	if err != nil {
		panic(fmt.Sprintf("❌ Invalid class hash: %s", err))
	}

	// Build constructor calldata
	constructorCalldata := buildConstructorCalldata(permit2Addr, mailboxAddr, deployerAddress, hookAddr, ismAddr)

	fmt.Println("📤 Sending deployment transaction...")

	// Deploy the contract with UDC using the modern approach
	resp, salt, err := accnt.DeployContractWithUDC(context.Background(), classHashFelt, constructorCalldata, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to deploy contract: %s", err))
	}

	// Extract transaction hash from response
	txHash := resp.Hash
	fmt.Printf("⏳ Transaction sent! Hash: %s\n", txHash.String())
	fmt.Println("⏳ Waiting for transaction confirmation...")

	// Wait for transaction receipt
	txReceipt, err := accnt.WaitForTransactionReceipt(context.Background(), txHash, time.Second)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to get transaction receipt: %s", err))
	}

	fmt.Printf("✅ Deployment completed!\n")
	fmt.Printf("   Transaction Hash: %s\n", txHash.String())
	fmt.Printf("   Execution Status: %s\n", txReceipt.ExecutionStatus)
	fmt.Printf("   Finality Status: %s\n", txReceipt.FinalityStatus)

	// Compute the deployed contract address
	deployedAddress := utils.PrecomputeAddressForUDC(classHashFelt, salt, constructorCalldata, utils.UDCCairoV0, accnt.Address)
	fmt.Printf("🏗️  Contract deployed at: %s\n", deployedAddress)

	// Note: Contract addresses are now managed via .env file, not deployment state

	// Note: .env file updates removed - addresses should be set manually after live deployment

	// Save deployment info
	saveDeploymentInfo(classHash, deployedAddress.String(), txHash.String(), salt.String())
}

// getClassHash retrieves the class hash from declaration file or environment variable
func getClassHash() (string, error) {
	// First try to get from environment variable
	if envClassHash := os.Getenv("HYPERLANE7683_CLASS_HASH"); envClassHash != "" {
		fmt.Printf("📋 Using class hash from environment variable: %s\n", envClassHash)
		return envClassHash, nil
	}

	// Try to read from declaration file in deployment directory
	deploymentDir := filepath.Clean(filepath.Join("state", "deployment"))
	declarationFile := filepath.Join(deploymentDir, "starknet-hyperlane7683-declaration.json")

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
func saveDeploymentInfo(classHash, deployedAddress, txHash, salt string) {
	deploymentInfo := map[string]string{
		"classHash":       classHash,
		"deployedAddress": deployedAddress,
		"transactionHash": txHash,
		"salt":            salt,
		"deploymentTime":  time.Now().Format(time.RFC3339),
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

	filename := filepath.Join(deploymentDir, "starknet-hyperlane7683-deployment.json")
	if err := os.WriteFile(filename, data, deploymentFilePerms); err != nil {
		fmt.Printf("⚠️  Failed to save deployment info: %s\n", err)
		return
	}

	fmt.Printf("💾 Deployment info saved to %s\n", filename)
}

// buildConstructorCalldata builds the constructor calldata for Hyperlane7683
func buildConstructorCalldata(permit2Addr, mailboxAddr, ownerAddr, hookAddr, ismAddr string) []*felt.Felt {
	// Convert addresses to felt, using 0x0 if not provided
	toFelt := func(hexAddr string) *felt.Felt {
		if hexAddr == "" {
			zero := felt.Zero
			return &zero
		}
		f, err := utils.HexToFelt(hexAddr)
		if err != nil {
			panic(fmt.Sprintf("❌ Invalid address %s: %s", hexAddr, err))
		}
		return f
	}

	// Constructor parameters in order: permit2, mailbox, owner, hook, ism
	permit2 := toFelt(permit2Addr)
	mailbox := toFelt(mailboxAddr)
	owner := toFelt(ownerAddr)
	hook := toFelt(hookAddr)
	ism := toFelt(ismAddr)

	return []*felt.Felt{permit2, mailbox, owner, hook, ism}
}
