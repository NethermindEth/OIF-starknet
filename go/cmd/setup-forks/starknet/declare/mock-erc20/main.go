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

	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/contracts"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
)

const (
	sierraContractFilePath = "../cairo/target/dev/oif_starknet_MockERC20.contract_class.json"
	casmContractFilePath   = "../cairo/target/dev/oif_starknet_MockERC20.compiled_contract_class.json"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	fmt.Println("📋 Declaring MockERC20 contract on Starknet...")

	// Load environment variables
	networkName := os.Getenv("NETWORK_NAME")
	if networkName == "" {
		networkName = "Starknet Sepolia" // Default to Starknet Sepolia
	}

	// Get network configuration
	networkConfig, err := config.GetNetworkConfig(networkName)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to get network config for %s: %s", networkName, err))
	}

	// Load Starknet account details from .env
	accountAddress := os.Getenv("STARKNET_DEPLOYER_ADDRESS")
	accountPrivateKey := os.Getenv("STARKNET_DEPLOYER_PRIVATE_KEY")
	accountPublicKey := os.Getenv("STARKNET_DEPLOYER_PUBLIC_KEY")

	if accountAddress == "" || accountPrivateKey == "" || accountPublicKey == "" {
		fmt.Println("❌ Missing required environment variables:")
		fmt.Println("   STARKNET_DEPLOYER_ADDRESS: Your Starknet account address")
		fmt.Println("   STARKNET_DEPLOYER_PRIVATE_KEY: Your private key")
		fmt.Println("   STARKNET_DEPLOYER_PUBLIC_KEY: Your public key")
		os.Exit(1)
	}

	fmt.Printf("📋 Network: %s\n", networkName)
	fmt.Printf("📋 RPC URL: %s\n", networkConfig.RPCURL)
	fmt.Printf("📋 Chain ID: %d\n", networkConfig.ChainID)
	fmt.Printf("📋 Account: %s\n", accountAddress)

	// Initialize connection to RPC provider
	client, err := rpc.NewProvider(networkConfig.RPCURL)
	if err != nil {
		panic(fmt.Sprintf("❌ Error connecting to RPC provider: %s", err))
	}

	// Initialize the account memkeyStore (public and private keys)
	ks := account.NewMemKeystore()
	privKeyBI, ok := new(big.Int).SetString(accountPrivateKey, 0)
	if !ok {
		panic("❌ Failed to convert private key to big.Int")
	}
	ks.Put(accountPublicKey, privKeyBI)

	// Convert account address to felt
	accountAddressInFelt, err := utils.HexToFelt(accountAddress)
	if err != nil {
		fmt.Println("❌ Failed to transform the account address, did you give the hex address?")
		panic(err)
	}

	// Initialize the account
	accnt, err := account.NewAccount(client, accountAddressInFelt, accountPublicKey, ks, account.CairoV2)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to initialize account: %s", err))
	}

	fmt.Println("✅ Connected to Starknet RPC")

	fmt.Printf("📋 Loading contract files:\n")
	fmt.Printf("   Sierra: %s\n", sierraContractFilePath)
	fmt.Printf("   Casm: %s\n", casmContractFilePath)

	// Unmarshalling the casm contract class from a JSON file.
	casmClass, err := utils.UnmarshalJSONFileToType[contracts.CasmClass](casmContractFilePath, "")
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to parse casm contract: %s", err))
	}

	// Unmarshalling the sierra contract class from a JSON file.
	contractClass, err := utils.UnmarshalJSONFileToType[contracts.ContractClass](sierraContractFilePath, "")
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to parse sierra contract: %s", err))
	}

	// Building and sending the Broadcast Invoke Txn.
	resp, err := accnt.BuildAndSendDeclareTxn(
		context.Background(),
		casmClass,
		contractClass,
		nil,
	)
	if err != nil {
		if strings.Contains(err.Error(), "is already declared") {
			fmt.Println("✅ Contract already declared")
			return
		}
		panic(fmt.Sprintf("❌ Failed to declare contract: %s", err))
	}

	// Building and sending the declare transaction
	fmt.Println("📤 Declaring contract...")
	_, err = accnt.WaitForTransactionReceipt(context.Background(), resp.Hash, time.Second)
	if err != nil {
		panic(fmt.Sprintf("❌ Declare txn failed: %s", err))
	}

	fmt.Printf("✅ Contract declaration completed!\n")
	fmt.Printf("   Class Hash: %s\n", resp.ClassHash)

	// Save declaration info
	saveDeclarationInfo(resp.Hash.String(), resp.ClassHash.String(), networkName)
}

// saveDeclarationInfo saves declaration information to a file
func saveDeclarationInfo(txHash, classHash, networkName string) {
	declarationInfo := map[string]string{
		"networkName":     networkName,
		"classHash":       classHash,
		"transactionHash": txHash,
		"declarationTime": time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(declarationInfo, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Failed to marshal declaration info: %s\n", err)
		return
	}

	// Ensure canonical state directory exists (local go/state)
	stateDir := filepath.Clean(filepath.Join("state", "network_state"))
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		fmt.Printf("⚠️  Failed to create state directory: %s\n", err)
		return
	}

	filename := filepath.Join(stateDir, fmt.Sprintf("%s-mock-erc20-declaration.json", sanitizeNetworkName(networkName)))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("⚠️  Failed to save declaration info: %s\n", err)
		return
	}

	fmt.Printf("💾 Declaration info saved to %s\n", filename)
}

// sanitizeNetworkName converts a human network name to a safe slug
func sanitizeNetworkName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
