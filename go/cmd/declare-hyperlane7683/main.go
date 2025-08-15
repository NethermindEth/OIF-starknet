package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/contracts"
	"github.com/NethermindEth/starknet.go/hash"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
)

const (
	casmContractFilePath = "../cairo/target/dev/oif_starknet_Hyperlane7683.contract_class.json"
	sierraContractFilePath   = "../cairo/target/dev/oif_starknet_Hyperlane7683.compiled_contract_class.json"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	fmt.Println("📋 Declaring Hyperlane7683 contract on Starknet...")

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
	accountAddress := os.Getenv("SN_DEPLOYER_ADDRESS")
	privateKey := os.Getenv("SN_DEPLOYER_PRIVATE_KEY")
	publicKey := os.Getenv("SN_DEPLOYER_PUBLIC_KEY")

	if accountAddress == "" || privateKey == "" || publicKey == "" {
		fmt.Println("❌ Missing required environment variables:")
		fmt.Println("   SN_DEPLOYER_ADDRESS: Your Starknet account address")
		fmt.Println("   SN_DEPLOYER_PRIVATE_KEY: Your private key")
		fmt.Println("   SN_DEPLOYER_PUBLIC_KEY: Your public key")
		os.Exit(1)
	}

	fmt.Printf("📋 Network: %s\n", networkName)
	fmt.Printf("📋 RPC URL: %s\n", networkConfig.RPCURL)
	fmt.Printf("📋 Chain ID: %d\n", networkConfig.ChainID)
	fmt.Printf("📋 Account: %s\n", accountAddress)

	// Initialise connection to RPC provider
	client, err := rpc.NewProvider(networkConfig.RPCURL)
	if err != nil {
		panic(fmt.Sprintf("❌ Error connecting to RPC provider: %s", err))
	}

	// Initialise the account memkeyStore (set public and private keys)
	ks := account.NewMemKeystore()
	privKeyBI, ok := new(big.Int).SetString(privateKey, 0)
	if !ok {
		panic("❌ Failed to convert private key to big.Int")
	}
	ks.Put(publicKey, privKeyBI)

	// Here we are converting the account address to felt
	accountAddressInFelt, err := utils.HexToFelt(accountAddress)
	if err != nil {
		fmt.Println("❌ Failed to transform the account address, did you give the hex address?")
		panic(err)
	}

	// Initialise the account (use Cairo v0 for v0.7.3 compatibility)
	accnt, err := account.NewAccount(client, accountAddressInFelt, publicKey, ks, 1) // Cairo v0
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to initialize account: %s", err))
	}

	fmt.Println("✅ Connected to Starknet RPC")

	// Check if contract files exist
	if _, err := os.Stat(sierraContractFilePath); os.IsNotExist(err) {
		panic(fmt.Sprintf("❌ Sierra contract file not found: %s", sierraContractFilePath))
	}

	if _, err := os.Stat(casmContractFilePath); os.IsNotExist(err) {
		panic(fmt.Sprintf("❌ Casm contract file not found: %s", casmContractFilePath))
	}

	fmt.Printf("📋 Loading contract files:\n")
	fmt.Printf("   Sierra: %s\n", sierraContractFilePath)
	fmt.Printf("   Casm: %s\n", casmContractFilePath)

	// Read and parse the casm contract file manually
	casmData, err := os.ReadFile(casmContractFilePath)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to read casm contract file: %s", err))
	}

	var casmClass contracts.CasmClass
	if err := json.Unmarshal(casmData, &casmClass); err != nil {
		panic(fmt.Sprintf("❌ Failed to parse casm contract: %s", err))
	}

	// Read and parse the sierra contract file manually
	sierraData, err := os.ReadFile(sierraContractFilePath)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to read sierra contract file: %s", err))
	}

	var contractClass rpc.ContractClass
	if err := json.Unmarshal(sierraData, &contractClass); err != nil {
		panic(fmt.Sprintf("❌ Failed to parse sierra contract: %s", err))
	}

	// Calculate class hash from Sierra program using the proper hash function
	classHash := hash.ClassHash(contractClass)
	fmt.Printf("📋 Calculated class hash: %s\n", classHash)

	// Calculate compiled class hash from Casm bytecode using the proper hash function
	compiledClassHash := hash.CompiledClassHash(casmClass)
	fmt.Printf("📋 Calculated compiled class hash: %s\n", compiledClassHash)

	// Building and sending the declare transaction
	fmt.Println("📤 Declaring contract...")

	// Add some debugging info
	fmt.Printf("   📋 Casm class size: %d bytes\n", len(casmClass.ByteCode))
	fmt.Printf("   📋 Sierra program length: %d entries\n", len(contractClass.SierraProgram))

	// Get the current nonce
	nonce, err := client.Nonce(context.Background(), rpc.BlockID{Tag: "latest"}, accountAddressInFelt)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to get nonce: %s", err))
	}

	// Create the declare transaction manually
	version, err := utils.HexToFelt("0x2")
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to convert version to felt: %s", err))
	}
	maxFee, err := utils.HexToFelt("0x100000000000000")
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to convert maxFee to felt: %s", err))
	}

	declareTxn := rpc.DeclareTxnV2{
		Version:             rpc.TransactionVersion(version.String()),
		MaxFee:             maxFee,
		Signature:          []*felt.Felt{},
		Nonce:              nonce,
		ClassHash:          classHash,
		CompiledClassHash:  compiledClassHash,
		SenderAddress:      accountAddressInFelt,
	}

	// Sign the transaction
	if err := accnt.SignDeclareTransaction(context.Background(), &declareTxn); err != nil {
		panic(fmt.Sprintf("❌ Failed to sign declare transaction: %s", err))
	}

	// Send the transaction
	resp, err := accnt.SendTransaction(context.Background(), &declareTxn)
	if err != nil {
		if strings.Contains(err.Error(), "is already declared") {
			fmt.Println("")
			fmt.Println("✅ Contract is already declared!")
			fmt.Printf("Class hash: %s\n", classHash)
			fmt.Println("💡 You can now use this class hash for deployment!")
			return
		}

		// Enhanced error handling
		fmt.Printf("❌ Declaration failed with error: %s\n", err)
		fmt.Println("")
		fmt.Println("🔍 Troubleshooting tips:")
		fmt.Println("   1. Check if your local Starknet node supports contract declaration")
		fmt.Println("   2. Verify the contract files are valid and complete")
		fmt.Println("   3. Ensure your account has sufficient balance for declaration fees")
		fmt.Println("   4. Try using a different RPC endpoint (e.g., Sepolia testnet)")
		fmt.Println("")
		fmt.Println("💡 For local development, you might need to:")
		fmt.Println("   - Use a different Starknet node version")
		fmt.Println("   - Or deploy to a testnet instead")

		panic(fmt.Sprintf("❌ Failed to declare contract: %s", err))
	}

	fmt.Printf("⏳ Contract declaration sent! Hash: %s\n", resp.TransactionHash)
	fmt.Println("⏳ Waiting for declaration confirmation...")

	// Wait for transaction receipt
	txReceipt, err := accnt.WaitForTransactionReceipt(context.Background(), resp.TransactionHash, time.Second)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to get transaction receipt: %s", err))
	}

	fmt.Printf("✅ Contract declaration completed!\n")
	fmt.Printf("   Transaction Hash: %s\n", resp.TransactionHash)
	fmt.Printf("   Class Hash: %s\n", classHash)
	fmt.Printf("   Execution Status: %s\n", txReceipt.ExecutionStatus)
	fmt.Printf("   Finality Status: %s\n", txReceipt.FinalityStatus)
	fmt.Printf("💡 Use this class hash for deployment: %s\n", classHash)

	// Save declaration info
	saveDeclarationInfo(resp.TransactionHash.String(), classHash.String(), networkName)
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

	filename := fmt.Sprintf("hyperlane7683_declaration_%s.json", networkName)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("⚠️  Failed to save declaration info: %s\n", err)
		return
	}

	fmt.Printf("💾 Declaration info saved to %s\n", filename)
}
