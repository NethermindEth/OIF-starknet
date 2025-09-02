package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/config"

	"github.com/NethermindEth/oif-starknet/go/pkg/ethutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"

	contracts "github.com/NethermindEth/oif-starknet/go/internal/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Network configuration - will be loaded from deployment state
var networks []NetworkConfig

// NetworkConfig represents a single network configuration
type NetworkConfig struct {
	name             string
	url              string
	chainID          uint64
	hyperlaneAddress common.Address
	dogCoinAddress   common.Address
}

// loadNetworks loads network configuration from centralized config and environment variables
func loadNetworks() error {
	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	// Build networks from centralized config
	networkNames := config.GetNetworkNames()
	networks = make([]NetworkConfig, 0, len(networkNames))

	for _, networkName := range networkNames {
		// Skip non-EVM networks (like Starknet)
		if networkName == "Starknet" {
			continue
		}

		networkConfig := config.Networks[networkName]

		// Load DogCoin address from environment variable
		var envVarName string
		switch networkName {
		case "Ethereum":
			envVarName = "ETHEREUM_DOG_COIN_ADDRESS"
		case "Optimism":
			envVarName = "OPTIMISM_DOG_COIN_ADDRESS"
		case "Arbitrum":
			envVarName = "ARBITRUM_DOG_COIN_ADDRESS"
		case "Base":
			envVarName = "BASE_DOG_COIN_ADDRESS"
		default:
			fmt.Printf("   ⚠️  Unknown network: %s\n", networkName)
			continue
		}

		dogCoinAddr := os.Getenv(envVarName)
		var dogCoinAddress common.Address
		if dogCoinAddr != "" {
			dogCoinAddress = common.HexToAddress(dogCoinAddr)
			fmt.Printf("   🔍 Loaded %s DogCoin from env: %s\n", networkName, dogCoinAddr)
		} else {
			fmt.Printf("   ⚠️  No DogCoin address found for %s (env var: %s)\n", networkName, envVarName)
		}

		networks = append(networks, NetworkConfig{
			name:             networkConfig.Name,
			url:              networkConfig.RPCURL,
			chainID:          networkConfig.ChainID,
			hyperlaneAddress: networkConfig.HyperlaneAddress,
			dogCoinAddress:   dogCoinAddress,
		})
	}

	return nil
}

// getEnvWithDefault gets an environment variable with a default fallback
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getHyperlaneDomain returns the Hyperlane domain ID for a given network
func getHyperlaneDomain(networkName string) uint32 {
	domain, err := config.GetHyperlaneDomain(networkName)
	if err != nil {
		return 0
	}
	return uint32(domain)
}

// Test user configuration (Alice-only for orders, Solver for fills)
var testUsers = []struct {
	name       string
	privateKey string
	address    string
}{
	{"Alice", "ALICE_PRIVATE_KEY", getEnvWithDefault("ALICE_PUB_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")},
	{"Solver", "SOLVER_PRIVATE_KEY", getEnvWithDefault("SOLVER_PUB_KEY", "0x90F79bf6EB2c4f870365E785982E1f101E93b906")},
}

// Order configuration
type OrderConfig struct {
	OriginChain      string
	DestinationChain string
	InputToken       string
	OutputToken      string
	InputAmount      *big.Int
	OutputAmount     *big.Int
	User             string
	OpenDeadline     uint32
	FillDeadline     uint32
}

// OrderData struct matching the Solidity OrderData
type OrderData struct {
	Sender             [32]byte
	Recipient          [32]byte
	InputToken         [32]byte
	OutputToken        [32]byte
	AmountIn           *big.Int
	AmountOut          *big.Int
	SenderNonce        *big.Int
	OriginDomain       uint32
	DestinationDomain  uint32
	DestinationSettler [32]byte
	FillDeadline       uint32
	Data               []byte
}

// OnchainCrossChainOrder struct matching the Solidity interface
type OnchainCrossChainOrder struct {
	FillDeadline  uint32
	OrderDataType [32]byte
	OrderData     []byte
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load network configuration from deployment state
	if err := loadNetworks(); err != nil {
		log.Fatalf("Failed to load networks: %v", err)
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: open-order <command>")
		fmt.Println("Commands:")
		fmt.Println("  random-to-evm        - Open a randomly generated order (EVM→EVM)")
		fmt.Println("  random-to-sn         - Open a random EVM→Starknet order")
		fmt.Println("  default-evm-evm      - Open default EVM→EVM order (Nonce: 1)")
		fmt.Println("  default-evm-sn       - Open default EVM→Starknet order (Nonce: 2)")
		fmt.Println("  batch                - Open multiple random orders (future)")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "random-to-evm":
		openRandomToEvm()
	case "random-to-sn":
		openRandomToStarknet()
	case "default-evm-evm":
		openDefaultEvmToEvm()
	case "default-evm-sn":
		openDefaultEvmToStarknet()
		//	case "default-sn-evm":
		//		openDefaultStarknetToEvm()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func openRandomToEvm() {
	fmt.Println("🎲 Opening Random Test Order...")

	// Seed random number generator
	// rand.Seed(time.Now().UnixNano())

	// Random origin and destination chains (exclude Starknet from origins)
	var evmNetworks []NetworkConfig
	for _, n := range networks {
		if n.name != "Starknet" {
			evmNetworks = append(evmNetworks, n)
		}
	}
	if len(evmNetworks) == 0 {
		log.Fatalf("no EVM networks configured")
	}

	originIdx := rand.Intn(len(evmNetworks))
	destIdx := rand.Intn(len(evmNetworks)) // Only pick from EVM networks for EVM-EVM orders
	for destIdx == originIdx {
		destIdx = (originIdx + 1) % len(evmNetworks) // Ensure different EVM chain
	}

	// Always use Alice for orders
	user := "Alice"

	// Random amounts (100-10000 tokens)
	inputAmount :=
		new(big.Int).Mul(big.NewInt(int64(rand.Intn(9901)+100)), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)) // 100-10000 tokens
	delta := big.NewInt(int64(rand.Intn(90) + 1))        // 1-90
	outputAmount := new(big.Int).Sub(inputAmount, delta) // slightly less to ensure it's fillable

	order := OrderConfig{
		OriginChain:      evmNetworks[originIdx].name,
		DestinationChain: evmNetworks[destIdx].name,
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      inputAmount,
		OutputAmount:     outputAmount,
		User:             user,
		OpenDeadline:     uint32(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint32(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func openRandomToStarknet() {
	fmt.Println("🎲 Opening Random EVM → Starknet Test Order...")
	// rand.Seed(time.Now().UnixNano())

	// Pick random EVM origin (exclude Starknet)
	var evmNetworks []NetworkConfig
	for _, n := range networks {
		if n.name != "Starknet" {
			evmNetworks = append(evmNetworks, n)
		}
	}
	if len(evmNetworks) == 0 {
		log.Fatalf("no EVM networks configured")
	}
	origin := evmNetworks[rand.Intn(len(evmNetworks))]

	inputAmount := new(big.Int).Mul(big.NewInt(int64(rand.Intn(9901)+100)), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)) // 100-10000 tokens
	delta := big.NewInt(int64(rand.Intn(90) + 1))                                                                                  // 1-90
	outputAmount := new(big.Int).Sub(inputAmount, delta)                                                                           // slightly less to ensure it's fillable

	order := OrderConfig{
		OriginChain:      origin.name,
		DestinationChain: "Starknet",
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      inputAmount,
		OutputAmount:     outputAmount,
		User:             "Alice",
		OpenDeadline:     uint32(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint32(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func openDefaultEvmToEvm() {
	fmt.Println("🎯 Opening Default EVM → EVM Test Order...")

	order := OrderConfig{
		OriginChain:      "Ethereum",
		DestinationChain: "Optimism",
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 tokens
		OutputAmount:     new(big.Int).Mul(big.NewInt(999), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),  // 999 tokens
		User:             "Alice",
		OpenDeadline:     uint32(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint32(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func openDefaultEvmToStarknet() {
	fmt.Println("🎯 Opening Default EVM → Starknet Test Order...")

	order := OrderConfig{
		OriginChain:      "Ethereum",
		DestinationChain: "Starknet",
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 tokens
		OutputAmount:     new(big.Int).Mul(big.NewInt(999), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),  // 999 tokens
		User:             "Alice",
		OpenDeadline:     uint32(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint32(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func executeOrder(order OrderConfig) {
	fmt.Printf("\n📋 Executing Order: %s → %s\n", order.OriginChain, order.DestinationChain)

	// Find origin network
	var originNetwork *NetworkConfig

	for _, network := range networks {
		if network.name == order.OriginChain {
			originNetwork = &network
			break
		}
	}

	if originNetwork == nil {
		log.Fatalf("Origin network not found: %s", order.OriginChain)
	}

	// Connect to origin network
	client, err := ethclient.Dial(originNetwork.url)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", order.OriginChain, err)
	}
	defer client.Close()

	// Get user private key
	userKey := os.Getenv(fmt.Sprintf("%s_PRIVATE_KEY", strings.ToUpper(order.User)))
	if userKey == "" {
		log.Fatalf("Private key not found for user: %s", order.User)
	}

	// Parse private key
	privateKey, err := ethutil.ParsePrivateKey(userKey)
	if err != nil {
		log.Fatalf("Failed to parse private key for %s: %v", order.User, err)
	}

	// Create auth
	auth, err := ethutil.NewTransactor(big.NewInt(int64(originNetwork.chainID)), privateKey)
	if err != nil {
		log.Fatalf("Failed to create auth: %v", err)
	}

	// Get current gas price
	gasPrice, err := ethutil.SuggestGas(client)
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}
	auth.GasPrice = gasPrice

	// Find destination network (check all networks, including Starknet)
	var destinationNetwork *NetworkConfig

	// First check EVM networks
	for _, network := range networks {
		if network.name == order.DestinationChain {
			destinationNetwork = &network
			break
		}
	}

	// If not found in EVM networks, check if it's Starknet
	if destinationNetwork == nil && order.DestinationChain == "Starknet" {
		// Create NetworkConfig for Starknet destination
		starknetConfig := config.Networks["Starknet"]
		destinationNetwork = &NetworkConfig{
			name:             "Starknet",
			url:              starknetConfig.RPCURL,
			chainID:          starknetConfig.ChainID,
			hyperlaneAddress: starknetConfig.HyperlaneAddress,
			dogCoinAddress:   common.HexToAddress(os.Getenv("STARKNET_DOG_COIN_ADDRESS")), // From env
		}
	}

	if destinationNetwork == nil {
		log.Fatalf("Destination network not found: %s", order.DestinationChain)
	}

	// Read localDomain from the origin Hyperlane contract to guarantee it matches on-chain
	localDomain, err := getLocalDomain(client, originNetwork.hyperlaneAddress)
	if err != nil {
		log.Fatalf("Failed to read localDomain from origin contract: %v", err)
	}

	// Preflight: balances and allowances on origin for input token
	inputToken := originNetwork.dogCoinAddress
	owner := auth.From
	spender := originNetwork.hyperlaneAddress

	// Get initial balances
	initialUserBalance, err := ethutil.ERC20Balance(client, inputToken, owner)
	if err == nil {
		fmt.Printf("   🔍 Initial InputToken balance(owner): %s\n", initialUserBalance.String())
	} else {
		fmt.Printf("   ⚠️  Could not read initial balance: %v\n", err)
	}

	initialHyperlaneBalance, err := ethutil.ERC20Balance(client, inputToken, spender)
	if err == nil {
		fmt.Printf("   🔍 Initial InputToken balance(hyperlane): %s\n", initialHyperlaneBalance.String())
	} else {
		fmt.Printf("   ⚠️  Could not read initial hyperlane balance: %v\n", err)
	}

	// Store initial balances for comparison
	initialBalances := struct {
		userBalance      *big.Int
		hyperlaneBalance *big.Int
	}{
		userBalance:      initialUserBalance,
		hyperlaneBalance: initialHyperlaneBalance,
	}

	// Pick a fresh senderNonce recognized by the contract to avoid InvalidNonce
	senderNonce, err := pickValidSenderNonce(client, originNetwork.hyperlaneAddress, auth.From)
	if err != nil {
		log.Fatalf("Failed to pick a valid sender nonce: %v", err)
	}

	// Build the order data
	orderData := buildOrderData(order, originNetwork, destinationNetwork, localDomain, senderNonce)

	// Build the OnchainCrossChainOrder
	crossChainOrder := OnchainCrossChainOrder{
		FillDeadline:  order.FillDeadline,
		OrderDataType: getOrderDataTypeHash(),
		OrderData:     encodeOrderData(orderData),
	}

	// Use generated bindings for open()
	contract, err := contracts.NewHyperlane7683(originNetwork.hyperlaneAddress, client)
	if err != nil {
		log.Fatalf("Failed to bind Hyperlane7683: %v", err)
	}
	tx, err := contract.Open(auth, contracts.OnchainCrossChainOrder{
		FillDeadline:  crossChainOrder.FillDeadline,
		OrderDataType: crossChainOrder.OrderDataType,
		OrderData:     crossChainOrder.OrderData,
	})
	if err != nil {
		log.Fatalf("Failed to send open transaction: %v", err)
	}

	fmt.Printf("   🚀 Transaction sent: %s\n", tx.Hash().Hex())
	fmt.Printf("   ⏳ Waiting for confirmation...\n")

	// Wait for transaction confirmation
	receipt, err := ethutil.WaitForTransaction(client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction confirmation: %v", err)
	}

	if receipt.Status == 1 {
		fmt.Printf("✅ Order opened successfully!\n")
		fmt.Printf("📊 Gas used: %d\n", receipt.GasUsed)
		fmt.Printf("🎯 Order ID: %s\n", calculateOrderId(orderData))

		// Verify that balances actually changed as expected
		fmt.Printf("   🔍 Verifying input tokens were transferred...\n")
		if err := verifyBalanceChanges(client, inputToken, owner, spender, initialBalances, order.InputAmount); err != nil {
			fmt.Printf("⚠️  Balance verification failed: %v\n", err)
		} else {
			fmt.Printf("👍 Balance changes verified\n")
		}
	} else {
		fmt.Printf("❌ Order opening failed\n")
		fmt.Printf("🔍 Transaction hash: %s\n", tx.Hash().Hex())
		fmt.Printf("📊 Gas used: %d\n", receipt.GasUsed)

		// Try to get more details about the failure
		fmt.Printf("   🔍 Checking transaction details...\n")
		txDetails, _, err := client.TransactionByHash(context.Background(), tx.Hash())
		if err != nil {
			fmt.Printf("❌ Could not retrieve transaction details: %v\n", err)
		} else {
			fmt.Printf("📝 Transaction data: 0x%x\n", txDetails.Data())
		}
	}

	fmt.Printf("\n🎉 Order execution completed!\n")
}

func buildOrderData(order OrderConfig, originNetwork *NetworkConfig, destinationNetwork *NetworkConfig, originDomain uint32, senderNonce *big.Int) OrderData {
	// Get the actual user address for the specified user
	var userAddr common.Address
	for _, user := range testUsers {
		if user.name == order.User {
			userAddr = common.HexToAddress(user.address)
			break
		}
	}

	// Input token from origin network, output token from destination network
	inputTokenAddr := originNetwork.dogCoinAddress
	outputTokenAddr := destinationNetwork.dogCoinAddress

	// Convert destination settler depending on destination chain type
	var destSettlerBytes [32]byte
	if destinationNetwork.name == "Starknet" {
		// Use Starknet Hyperlane address from config (.env) as raw 32 bytes (felt)
		if snNetwork, exists := config.Networks["Starknet"]; exists {
			starknetHyperlaneAddr := os.Getenv("STARKNET_HYPERLANE_ADDRESS")
			if starknetHyperlaneAddr != "" {
				destSettlerBytes = hexToBytes32(starknetHyperlaneAddr)
			} else if snNetwork.HyperlaneAddress.Hex() != "" {
				destSettlerBytes = hexToBytes32(snNetwork.HyperlaneAddress.Hex())
			}
		}
		if destSettlerBytes == ([32]byte{}) {
			log.Printf("⚠️  Starknet Hyperlane address not found in config; destinationSettler will be zero")
		}
	} else {
		// EVM router is 20-byte address left-padded to 32
		destSettler := destinationNetwork.hyperlaneAddress
		copy(destSettlerBytes[12:], destSettler.Bytes())
	}

	// Convert addresses to bytes32 (left-pad)
	var userBytes, inputTokenBytes, outputTokenBytes [32]byte
	copy(userBytes[12:], userAddr.Bytes())
	copy(inputTokenBytes[12:], inputTokenAddr.Bytes())

	// For recipient, we need to determine if it should be EVM or Starknet address
	var recipientBytes [32]byte
	if destinationNetwork.name == "Starknet" {
		// If destination is Starknet, recipient should be the Starknet address of the same user
		var starknetUserAddr string
		switch order.User {
		case "Alice":
			starknetUserAddr = getEnvWithDefault("STARKNET_ALICE_ADDRESS", "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7")
		case "Solver":
			starknetUserAddr = getEnvWithDefault("STARKNET_SOLVER_ADDRESS", "0x02af9427c5a277474c079a1283c880ee8a6f0f8fbf73ce969c08d88befec1bba")
		default:
			// Fallback to Alice address if unknown user (should only be Alice now)
			starknetUserAddr = getEnvWithDefault("STARKNET_ALICE_ADDRESS", "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7")
		}
		recipientBytes = hexToBytes32(starknetUserAddr)
	} else {
		// If destination is EVM, recipient is the same EVM user
		recipientBytes = userBytes
	}
	if destinationNetwork.name == "Starknet" {
		// Get Starknet DogCoin address from config (.env)
		starknetDogCoinAddr := os.Getenv("STARKNET_DOG_COIN_ADDRESS")
		if starknetDogCoinAddr != "" {
			outputTokenBytes = hexToBytes32(starknetDogCoinAddr)
		} else {
			log.Fatalf("missing STARKNET_DOG_COIN_ADDRESS in .env for destination; set this variable")
		}
	} else {
		copy(outputTokenBytes[12:], outputTokenAddr.Bytes())
	}

	// Get the destination chain ID (Hyperlane domain)
	destinationChainID := getHyperlaneDomain(destinationNetwork.name)

	return OrderData{
		Sender:             userBytes,
		Recipient:          recipientBytes,
		InputToken:         inputTokenBytes,
		OutputToken:        outputTokenBytes,
		AmountIn:           order.InputAmount,
		AmountOut:          order.OutputAmount,
		SenderNonce:        senderNonce,
		OriginDomain:       originDomain,
		DestinationDomain:  destinationChainID,
		DestinationSettler: destSettlerBytes,
		FillDeadline:       order.FillDeadline,
		Data:               []byte{},
	}
}

func hexToBytes32(hexStr string) [32]byte {
	s := strings.TrimPrefix(hexStr, "0x")
	if len(s)%2 == 1 {
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		log.Fatalf("invalid hex for bytes32: %s (%v)", hexStr, err)
	}
	if len(b) > 32 {
		b = b[len(b)-32:]
	}
	var out [32]byte
	copy(out[32-len(b):], b)
	return out
}

// getLocalDomain reads the `localDomain()` from the Hyperlane7683 contract on the connected chain
func getLocalDomain(client *ethclient.Client, contractAddress common.Address) (uint32, error) {
	abiStr := `[{"inputs":[],"name":"localDomain","outputs":[{"internalType":"uint32","name":"","type":"uint32"}],"stateMutability":"view","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return 0, err
	}

	data, err := parsedABI.Pack("localDomain")
	if err != nil {
		return 0, err
	}

	msg := ethereum.CallMsg{To: &contractAddress, Data: data}
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}

	var domain uint32
	if err := parsedABI.UnpackIntoInterface(&domain, "localDomain", result); err != nil {
		return 0, err
	}

	return domain, nil
}

// isValidNonce calls the contract to check whether a nonce is usable for a given address
func isValidNonce(client *ethclient.Client, contractAddress, from common.Address, nonce *big.Int) (bool, error) {
	abiStr := `[{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"uint256","name":"_nonce","type":"uint256"}],"name":"isValidNonce","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return false, err
	}

	data, err := parsedABI.Pack("isValidNonce", from, nonce)
	if err != nil {
		return false, err
	}

	msg := ethereum.CallMsg{To: &contractAddress, Data: data}
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	var valid bool
	if err := parsedABI.UnpackIntoInterface(&valid, "isValidNonce", result); err != nil {
		return false, err
	}
	return valid, nil
}

// pickValidSenderNonce finds a nonce that the contract reports as valid for the provided sender
func pickValidSenderNonce(client *ethclient.Client, contractAddress, from common.Address) (*big.Int, error) {
	// Start with a pseudo-random seed and probe upward
	seed := time.Now().Unix() % 1_000_000
	if seed < 1 {
		seed = 1
	}
	nonce := big.NewInt(seed)
	for i := 0; i < 1000; i++ {
		valid, err := isValidNonce(client, contractAddress, from, nonce)
		if err != nil {
			return nil, err
		}
		if valid {
			return new(big.Int).Set(nonce), nil
		}
		nonce = new(big.Int).Add(nonce, big.NewInt(1))
	}
	return nil, fmt.Errorf("could not find a valid sender nonce after 1000 attempts starting from %d", seed)
}

func getOrderDataTypeHash() [32]byte {
	// This should match OrderEncoder.orderDataType() from Solidity EXACTLY
	// Including field names and spacing
	orderDataType := "OrderData(bytes32 sender,bytes32 recipient,bytes32 inputToken,bytes32 outputToken,uint256 amountIn,uint256 amountOut,uint256 senderNonce,uint32 originDomain,uint32 destinationDomain,bytes32 destinationSettler,uint32 fillDeadline,bytes data)"
	hash := crypto.Keccak256Hash([]byte(orderDataType))

	//// Debug: Print the hash we're generating
	// fmt.Printf("   🔍 Generated orderDataType hash: %x\n", hash)
	//fmt.Printf("   �� Expected hash from contract: 08d75650babf4de09c9273d48ef647876057ed91d4323f8a2e3ebc2cd8a63b5e\n")

	return hash
}

func encodeOrderData(orderData OrderData) []byte {
	// Pack a single tuple that matches Solidity's OrderData and abi.encode(order)
	tupleT, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "sender", Type: "bytes32"},
		{Name: "recipient", Type: "bytes32"},
		{Name: "inputToken", Type: "bytes32"},
		{Name: "outputToken", Type: "bytes32"},
		{Name: "amountIn", Type: "uint256"},
		{Name: "amountOut", Type: "uint256"},
		{Name: "senderNonce", Type: "uint256"},
		{Name: "originDomain", Type: "uint32"},
		{Name: "destinationDomain", Type: "uint32"},
		{Name: "destinationSettler", Type: "bytes32"},
		{Name: "fillDeadline", Type: "uint32"},
		{Name: "data", Type: "bytes"},
	})
	if err != nil {
		log.Fatalf("Failed to define OrderData tuple type: %v", err)
	}

	args := abi.Arguments{{Type: tupleT}}

	encoded, err := args.Pack(orderData)
	if err != nil {
		log.Fatalf("Failed to ABI-pack OrderData tuple: %v", err)
	}

	return encoded
}





// func sendOpenTransaction(client *ethclient.Client, auth *bind.TransactOpts, contractAddr common.Address, calldata []byte) (*types.Transaction, error) {
//	// Get current nonce
//	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get nonce: %v", err)
//	}
//
//	// Create transaction
//	tx := types.NewTransaction(
//		nonce,
//		contractAddr,
//		big.NewInt(0),     // No ETH value
//		uint64(2_000_000), // Gas limit
//		auth.GasPrice,
//		calldata,
//	)
//
//	// Sign transaction
//	signedTx, err := auth.Signer(auth.From, tx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to sign transaction: %v", err)
//	}
//
//	// Send transaction
//	err = client.SendTransaction(context.Background(), signedTx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to send transaction: %v", err)
//	}
//
//	return signedTx, nil
//}

func calculateOrderId(orderData OrderData) string {
	// Match Solidity: keccak256(abi.encode(order))
	encoded := encodeOrderData(orderData)
	hash := crypto.Keccak256Hash(encoded)
	return hash.Hex()
}

// verifyBalanceChanges verifies that opening an order actually transferred tokens
func verifyBalanceChanges(client *ethclient.Client, tokenAddress, userAddress, hyperlaneAddress common.Address, initialBalances struct {
	userBalance      *big.Int
	hyperlaneBalance *big.Int
}, expectedTransferAmount *big.Int) error {
	// Wait a moment for the transaction to be fully processed
	time.Sleep(2 * time.Second)

	// Get final balances
	finalUserBalance, err := ethutil.ERC20Balance(client, tokenAddress, userAddress)
	if err != nil {
		return fmt.Errorf("failed to get final user balance: %w", err)
	}

	finalHyperlaneBalance, err := ethutil.ERC20Balance(client, tokenAddress, hyperlaneAddress)
	if err != nil {
		return fmt.Errorf("failed to get final hyperlane balance: %w", err)
	}

	// Calculate actual changes
	userBalanceChange := new(big.Int).Sub(initialBalances.userBalance, finalUserBalance)
	hyperlaneBalanceChange := new(big.Int).Sub(finalHyperlaneBalance, initialBalances.hyperlaneBalance)

	// Print balance changes
	// fmt.Printf("     💰 User balance change: %s → %s (Δ: %s)\n",
	//	ethutil.FormatTokenAmount(initialBalances.userBalance, 18),
	//	ethutil.FormatTokenAmount(finalUserBalance, 18),
	//	ethutil.FormatTokenAmount(userBalanceChange, 18))

	// fmt.Printf("     💰 Hyperlane balance change: %s → %s (Δ: %s)\n",
	//	ethutil.FormatTokenAmount(initialBalances.hyperlaneBalance, 18),
	//	ethutil.FormatTokenAmount(finalHyperlaneBalance, 18),
	//	ethutil.FormatTokenAmount(hyperlaneBalanceChange, 18))

	// Verify the changes match expectations
	if userBalanceChange.Cmp(expectedTransferAmount) != 0 {
		return fmt.Errorf("user balance decreased by %s, expected %s",
			ethutil.FormatTokenAmount(userBalanceChange, 18),
			ethutil.FormatTokenAmount(expectedTransferAmount, 18))
	}

	if hyperlaneBalanceChange.Cmp(expectedTransferAmount) != 0 {
		return fmt.Errorf("hyperlane balance increased by %s, expected %s",
			ethutil.FormatTokenAmount(hyperlaneBalanceChange, 18),
			ethutil.FormatTokenAmount(expectedTransferAmount, 18))
	}

	// Verify total supply is preserved (user decrease = hyperlane increase)
	if userBalanceChange.Cmp(hyperlaneBalanceChange) != 0 {
		return fmt.Errorf("balance changes don't match: user decreased by %s, hyperlane increased by %s",
			ethutil.FormatTokenAmount(userBalanceChange, 18),
			ethutil.FormatTokenAmount(hyperlaneBalanceChange, 18))
	}

	return nil
}

//// simulateAndDecodeRevert runs an eth_call with the same calldata and decodes common custom errors
// func simulateAndDecodeRevert(client *ethclient.Client, to, from common.Address, data []byte) error {
//	msg := ethereum.CallMsg{To: &to, From: from, Data: data, Gas: 2_000_000}
//	_, err := client.CallContract(context.Background(), msg, nil)
//	if err == nil {
//		fmt.Printf("   ✅ eth_call succeeded (unexpected for a failing tx)\n")
//		return nil
//	}
//	// Expecting a "execution reverted"-style error; extract data if present
//	// go-ethereum attaches the revert data to the error string; we can do a second call with a custom tracer,
//	// or parse with the built-in helper. Simpler: do CallContract with a big gas and then fetch debug via error string.
//	// As a fallback, attempt to decode if the provider returns the revert data hex in err.Error().
//
//	errStr := err.Error()
//	// Look for "data: 0x..." substring
//	idx := strings.Index(errStr, "data: 0x")
//	if idx == -1 {
//		fmt.Printf("   ⚠️  No revert data found in error string: %s\n", errStr)
//		return nil
//	}
//	hexData := errStr[idx+6:]
//	// Trim trailing context
//	if sp := strings.Index(hexData, "]"); sp != -1 {
//		hexData = hexData[:sp]
//	}
//	hexData = strings.TrimSpace(hexData)
//	hexData = strings.TrimSuffix(hexData, ",")
//	hexData = strings.TrimSuffix(hexData, "0x")
//
//	revertData, decErr := hex.DecodeString(hexData)
//	if decErr != nil || len(revertData) < 4 {
//		fmt.Printf("   ⚠️  Could not parse revert data from error: %v\n", decErr)
//		return nil
//	}
//
//	// Known custom errors
//	errorABI := `[
//        {"type":"error","name":"InvalidOrderType","inputs":[{"type":"bytes32","name":"orderType"}]},
//        {"type":"error","name":"InvalidOriginDomain","inputs":[{"type":"uint32","name":"originDomain"}]},
//        {"type":"error","name":"InvalidOrderId","inputs":[]},
//        {"type":"error","name":"OrderFillExpired","inputs":[]},
//        {"type":"error","name":"InvalidOrderDomain","inputs":[]},
//        {"type":"error","name":"InvalidDomain","inputs":[]},
//        {"type":"error","name":"InvalidSender","inputs":[]},
//        {"type":"error","name":"InvalidNonce","inputs":[]},
//        {"type":"error","name":"InvalidNativeAmount","inputs":[]}
//    ]`
//	parsed, err := abi.JSON(strings.NewReader(errorABI))
//	if err != nil {
//		return err
//	}
//
//	sel := revertData[:4]
//	for name, def := range parsed.Errors {
//		if bytes.Equal(sel, def.ID.Bytes()[:4]) {
//			// Decode args
//			vals, decErr := def.Inputs.Unpack(revertData[4:])
//			if decErr != nil {
//				fmt.Printf("   🔍 Revert: %s (failed to unpack args: %v)\n", name, decErr)
//			} else {
//				fmt.Printf("   🔍 Revert: %s %v\n", name, vals)
//			}
//			return nil
//		}
//	}
//	fmt.Printf("   ⚠️  Unknown revert selector: 0x%x\n", sel)
//	return nil
//}
