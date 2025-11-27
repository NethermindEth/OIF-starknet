package openorder

// Ztarknet order creation logic - similar to Starknet but uses Ztarknet network
// This handles creating orders on Ztarknet chains (Cairo-based, identical to Starknet)

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/common"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/NethermindEth/oif-starknet/solver/pkg/starknetutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
)

// getAliceAddressForZtarknetNetwork gets Alice's address for a specific network
// For Ztarknet origin, always use Ztarknet Alice
// For destination, use appropriate network's Alice
func getAliceAddressForZtarknetNetwork(networkName string) (string, error) {
	if strings.Contains(strings.ToLower(networkName), "ztarknet") {
		// Use Ztarknet Alice address
		address := envutil.GetZtarknetAliceAddress()
		if address == "" {
			return "", fmt.Errorf("ztarknet Alice address not set")
		}
		return address, nil
	} else if strings.Contains(strings.ToLower(networkName), "starknet") {
		// Use Starknet Alice address for Starknet destination
		address := envutil.GetStarknetAliceAddress()
		if address == "" {
			return "", fmt.Errorf("starknet Alice address not set")
		}
		return address, nil
	} else {
		// Use conditional environment variable for EVM networks
		address := envutil.GetAlicePublicKey()
		if address == "" {
			return "", fmt.Errorf("alice public key not set")
		}
		return address, nil
	}
}

// ZtarknetNetworkConfig represents a single network configuration for Ztarknet
type ZtarknetNetworkConfig struct {
	name             string
	url              string
	chainID          uint64
	hyperlaneAddress string
	dogCoinAddress   string
}

// OrderConfig represents order configuration for Ztarknet (reusing StarknetOrderConfig structure)
type ZtarknetOrderConfig struct {
	OriginChain      string
	DestinationChain string
	InputToken       string
	OutputToken      string
	InputAmount      *big.Int
	OutputAmount     *big.Int
	User             string
	OpenDeadline     uint64
	FillDeadline     uint64
}

// Test user configuration for Ztarknet
var ztarknetTestUsers []struct {
	name       string
	privateKey string
	address    string
}

// initializeZtarknetTestUsers initializes the test user configuration after .env is loaded
func initializeZtarknetTestUsers() {
	// Use ztarknet environment variables (testnet-only, no LOCAL_ variants)
	aliceAddr := envutil.GetZtarknetAliceAddress()
	solverAddr := envutil.GetZtarknetSolverAddress()
	alicePrivateKeyVar := envutil.GetZtarknetAlicePrivateKey()
	solverPrivateKeyVar := envutil.GetZtarknetSolverPrivateKey()

	ztarknetTestUsers = []struct {
		name       string
		privateKey string
		address    string
	}{
		{"Alice", alicePrivateKeyVar, aliceAddr},
		{"Solver", solverPrivateKeyVar, solverAddr},
	}
}

// loadZtarknetNetworks loads network configuration from centralized config and environment variables
func loadZtarknetNetworks() []ZtarknetNetworkConfig {
	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	// Build networks from centralized config
	networkNames := config.GetNetworkNames()
	networks := make([]ZtarknetNetworkConfig, 0, len(networkNames))

	for _, networkName := range networkNames {
		// Only include Ztarknet networks
		if networkName != "Ztarknet" {
			continue
		}

		networkConfig := config.Networks[networkName]

		// Load addresses from environment variables
		hyperlaneAddr := getEnvWithDefault("ZTARKNET_HYPERLANE_ADDRESS", "")
		dogAddr := getEnvWithDefault("ZTARKNET_DOG_COIN_ADDRESS", "")

		if hyperlaneAddr == "" || dogAddr == "" {
			log.Fatalf("missing ZTARKNET_HYPERLANE_ADDRESS or ZTARKNET_DOG_COIN_ADDRESS in .env")
		}

		networks = append(networks, ZtarknetNetworkConfig{
			name:             networkConfig.Name,
			url:              networkConfig.RPCURL,
			chainID:          networkConfig.ChainID,
			hyperlaneAddress: hyperlaneAddr,
			dogCoinAddress:   dogAddr,
		})
	}

	return networks
}

// RunZtarknetOrder creates a Ztarknet order based on the command
func RunZtarknetOrder(command string) {
	fmt.Println("🎯 Running Ztarknet order creation...")

	// Load configuration (this loads .env and initializes networks)
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize test users after .env is loaded
	initializeZtarknetTestUsers()

	// Load network configuration
	networks := loadZtarknetNetworks()

	switch command {
	case "random":
		openRandomZtarknetOrder(networks)
	case "to-starknet":
		openZtarknetToStarknet(networks)
	case "default":
		openDefaultZtarknetToStarknet(networks)
	default:
		// Default to Ztarknet -> Starknet order
		openDefaultZtarknetToStarknet(networks)
	}
}

func openRandomZtarknetOrder(networks []ZtarknetNetworkConfig) {
	fmt.Println("🎲 Opening Random Ztarknet Test Order...")

	// Use configured Ztarknet network as origin
	originChain := "Ztarknet"

	// For now, default to Starknet as destination (can be extended later)
	destinationChain := "Starknet"

	// Get Alice's address for the destination chain
	user, err := getAliceAddressForZtarknetNetwork(destinationChain)
	if err != nil {
		log.Fatalf("Failed to get Alice address for %s: %v", destinationChain, err)
	}

	// Random amounts
	inputAmount := CreateTokenAmount(int64(secureRandomInt(maxTokenAmount-minTokenAmount)+minTokenAmount), 18) // 100-10000 tokens
	delta := CreateTokenAmount(int64(secureRandomInt(maxDeltaAmount-minDeltaAmount)+minDeltaAmount), 18)       // 1-10 tokens
	outputAmount := new(big.Int).Sub(inputAmount, delta)                                                       // slightly less to ensure it's fillable

	order := ZtarknetOrderConfig{
		OriginChain:      originChain,
		DestinationChain: destinationChain,
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      inputAmount,
		OutputAmount:     outputAmount,
		User:             user, // Recipient address on destination chain
		OpenDeadline:     uint64(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint64(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeZtarknetOrder(&order, networks)
}

func openDefaultZtarknetToStarknet(networks []ZtarknetNetworkConfig) {
	fmt.Println("🎯 Opening Default Ztarknet → Starknet Test Order...")

	// Use configured networks
	originChain := "Ztarknet"
	destinationChain := "Starknet"

	// Get Alice's address for the destination chain (Starknet)
	aliceAddress, err := getAliceAddressForZtarknetNetwork(destinationChain)
	if err != nil {
		log.Fatalf("Failed to get Alice address for %s: %v", destinationChain, err)
	}

	order := ZtarknetOrderConfig{
		OriginChain:      originChain,
		DestinationChain: destinationChain,
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      CreateTokenAmount(1000, 18),                                // 1000 tokens
		OutputAmount:     CreateTokenAmount(testOutputAmountStarknet, tokenDecimals), // 999 tokens
		User:             aliceAddress,                                               // Recipient address on destination chain
		OpenDeadline:     uint64(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint64(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeZtarknetOrder(&order, networks)
}

func openZtarknetToStarknet(networks []ZtarknetNetworkConfig) {
	// Alias for default
	openDefaultZtarknetToStarknet(networks)
}

func executeZtarknetOrder(order *ZtarknetOrderConfig, networks []ZtarknetNetworkConfig) {
	fmt.Printf("\n📋 Executing Order: %s → %s\n", order.OriginChain, order.DestinationChain)

	// Find origin network (should be Ztarknet)
	var originNetwork *ZtarknetNetworkConfig
	for _, network := range networks {
		if network.name == order.OriginChain {
			originNetwork = &network
			break
		}
	}

	if originNetwork == nil {
		fmt.Printf("❌ Origin network not found: %s\n", order.OriginChain)
		os.Exit(1)
	}

	// Connect to Ztarknet RPC
	client, err := rpc.NewProvider(originNetwork.url)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s: %v\n", order.OriginChain, err)
		os.Exit(1)
	}

	// Always use Alice's Ztarknet credentials for signing orders on Ztarknet
	// The order.User field contains the recipient address (destination chain), not the signer
	userKey := envutil.GetZtarknetAlicePrivateKey()
	userPublicKey := envutil.GetZtarknetAlicePublicKey()

	if userKey == "" || userPublicKey == "" {
		fmt.Printf("❌ Missing Alice's Ztarknet credentials\n")
		fmt.Printf("   Required: ZTARKNET_ALICE_PRIVATE_KEY and ZTARKNET_ALICE_PUBLIC_KEY\n")
		os.Exit(1)
	}

	// Always use Alice's Ztarknet address for signing (order signer)
	var userAddr string
	for _, user := range ztarknetTestUsers {
		if user.name == "Alice" {
			userAddr = user.address
			break
		}
	}

	// Get domains from config
	// For Ztarknet, domain is 0x999999 (10076175 in decimal), not 999999
	var originDomain, destinationDomain uint32
	if order.OriginChain == "Ztarknet" {
		// Ztarknet domain is 0x999999 = 10066329 in decimal
		originDomain = 10066329
	} else if originConfig, err := config.GetHyperlaneDomain(order.OriginChain); err == nil {
		originDomain = uint32(originConfig)
	} else {
		fmt.Printf("   ⚠️  Warning: Could not get origin domain from config, using chain ID\n")
		originDomain = uint32(originNetwork.chainID)
	}

	if order.DestinationChain == "Ztarknet" {
		// Ztarknet domain is 0x999999 = 10066329 in decimal
		destinationDomain = 10066329	
		} else if destConfig, err := config.GetHyperlaneDomain(order.DestinationChain); err == nil {
		destinationDomain = uint32(destConfig)
	} else {
		fmt.Printf("   ⚠️  Warning: Could not get destination domain from config\n")
		os.Exit(1)
	}

	// Preflight: check balances and allowances
	inputToken := originNetwork.dogCoinAddress
	owner := userAddr
	spender := originNetwork.hyperlaneAddress

	// Get initial balances
	initialUserBalance, err := starknetutil.ERC20Balance(client, inputToken, owner)
	if err == nil {
		fmt.Printf("   🔍 Initial InputToken balance(owner): %s\n", starknetutil.FormatTokenAmount(initialUserBalance, 18))
	} else {
		fmt.Printf("   ⚠️  Could not read initial balance: %v\n", err)
	}

	// Check if Alice has sufficient tokens for the order
	requiredAmount := order.InputAmount
	if initialUserBalance == nil || initialUserBalance.Cmp(requiredAmount) < 0 {
		fmt.Printf("   ⚠️  Insufficient balance! Alice needs %s tokens but has %s\n",
			starknetutil.FormatTokenAmount(requiredAmount, 18),
			starknetutil.FormatTokenAmount(initialUserBalance, 18))
		fmt.Printf("   💡 Please mint tokens manually using the MockERC20 contract's mint() function\n")
		fmt.Printf("   📝 Contract address: %s\n", inputToken)
		fmt.Printf("❌ Insufficient token balance for order creation\n")
		os.Exit(1)
	} else {
		fmt.Printf("   ✅ Alice has sufficient tokens (%s)\n", starknetutil.FormatTokenAmount(initialUserBalance, 18))
	}

	// Create user account for transaction signing (needed for approval)
	userAddrFelt, err := utils.HexToFelt(userAddr)
	if err != nil {
		fmt.Printf("❌ Failed to convert user address to felt: %v\n", err)
		os.Exit(1)
	}

	// Initialize user's keystore
	userKs := account.NewMemKeystore()
	userPrivKeyBI, ok := new(big.Int).SetString(userKey, 0)
	if !ok {
		fmt.Printf("❌ Failed to convert private key for %s: %v\n", order.User, err)
		os.Exit(1)
	}
	userKs.Put(userPublicKey, userPrivKeyBI)

	// Create user account (Cairo v2)
	userAccnt, err := account.NewAccount(client, userAddrFelt, userPublicKey, userKs, account.CairoV2)
	if err != nil {
		fmt.Printf("❌ Failed to create account for %s: %v\n", order.User, err)
		os.Exit(1)
	}

	// Check allowance
	allowance, err := starknetutil.ERC20Allowance(client, inputToken, owner, spender)
	if err == nil {
		fmt.Printf("   🔍 Current allowance(owner->hyperlane): %s\n", starknetutil.FormatTokenAmount(allowance, 18))
	} else {
		fmt.Printf("   ⚠️  Could not read allowance: %v\n", err)
	}

	// If allowance is insufficient, approve the Hyperlane contract
	requiredAmount = order.InputAmount
	if allowance == nil || allowance.Cmp(requiredAmount) < 0 {
		fmt.Printf("   🔄 Insufficient allowance, approving %s tokens...\n", starknetutil.FormatTokenAmount(requiredAmount, 18))

		// Create approval transaction
		approveCall, err := starknetutil.ERC20Approve(inputToken, spender, requiredAmount)
		if err != nil {
			fmt.Printf("❌ Failed to create approve transaction: %v\n", err)
			os.Exit(1)
		}

		// Send approval transaction
		approveTx, err := userAccnt.BuildAndSendInvokeTxn(context.Background(), []rpc.InvokeFunctionCall{*approveCall}, nil)
		if err != nil {
			fmt.Printf("❌ Failed to send approval transaction: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("   🚀 Approval transaction sent: %s\n", approveTx.Hash.String())
		fmt.Printf("   ⏳ Waiting for approval confirmation...\n")

		// Wait for approval transaction to be mined
		_, err = userAccnt.WaitForTransactionReceipt(context.Background(), approveTx.Hash, 2*time.Second)
		if err != nil {
			fmt.Printf("❌ Failed to wait for approval transaction: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("   ✅ Approval confirmed!\n")
	} else {
		fmt.Printf("   ✅ Sufficient allowance already exists\n")
	}

	// Generate a random nonce for the order
	senderNonce := big.NewInt(time.Now().UnixNano())

	// Build the order data (reuse Starknet order data structure since it's identical)
	orderData := buildZtarknetOrderData(order, originNetwork, originDomain, destinationDomain, senderNonce, order.DestinationChain)

	// Build the StarknetOnchainCrossChainOrder with u256 order_data_type (low, high)
	lowHash, highHash := getOrderDataTypeHashU256()
	crossChainOrder := StarknetOnchainCrossChainOrder{
		FillDeadline:      order.FillDeadline,
		OrderDataTypeLow:  lowHash,
		OrderDataTypeHigh: highHash,
		OrderData:         encodeStarknetOrderData(&orderData),
	}

	// Use generated bindings for open()
	fmt.Printf("   📝 Calling open() function...\n")

	// Get Hyperlane7683 contract address
	hyperlaneAddrFelt, err := utils.HexToFelt(originNetwork.hyperlaneAddress)
	if err != nil {
		fmt.Printf("❌ Failed to convert Hyperlane7683 address to felt: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   📝 Sending open transaction...\n")

	// Build the transaction calldata for open(fill_deadline: u64, order_data_type: u256, order_data: Bytes)
	calldata := []*felt.Felt{
		utils.Uint64ToFelt(crossChainOrder.FillDeadline),
		crossChainOrder.OrderDataTypeLow,
		crossChainOrder.OrderDataTypeHigh,
	}
	calldata = append(calldata, crossChainOrder.OrderData...)

	tx, err := userAccnt.BuildAndSendInvokeTxn(
		context.Background(),
		[]rpc.InvokeFunctionCall{{
			ContractAddress: hyperlaneAddrFelt,
			FunctionName:    "open",
			CallData:        calldata,
		}},
		nil,
	)
	if err != nil {
		fmt.Printf("❌ Failed to send open transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   🚀 Transaction sent: %s\n", tx.Hash.String())
	fmt.Printf("   ⏳ Waiting for confirmation...\n")

	// Wait for transaction receipt
	_, err = userAccnt.WaitForTransactionReceipt(context.Background(), tx.Hash, time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to wait for transaction confirmation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   ✅ Order opened successfully!\n")

	fmt.Printf("\n🎉 Order execution completed!\n")
	fmt.Printf("📊 Order Summary:\n")
	fmt.Printf("   Input Amount: %s\n", order.InputAmount.String())
	fmt.Printf("   Output Amount: %s\n", order.OutputAmount.String())
	fmt.Printf("   Origin Chain: %s\n", order.OriginChain)
	fmt.Printf("   Destination Chain: %s\n", order.DestinationChain)
}

func buildZtarknetOrderData(order *ZtarknetOrderConfig, originNetwork *ZtarknetNetworkConfig, originDomain, destinationDomain uint32, senderNonce *big.Int, destChainName string) StarknetOrderData {
	// Get the actual user address for the specified user (Alice on Ztarknet)
	var userAddr string
	for _, user := range ztarknetTestUsers {
		if user.name == "Alice" {
			userAddr = user.address
			break
		}
	}

	// Convert addresses to felt
	userAddrFelt, _ := utils.HexToFelt(userAddr)
	inputTokenFelt, _ := utils.HexToFelt(originNetwork.dogCoinAddress)

	// Determine recipient based on destination chain
	var recipientFelt *felt.Felt
	var outputTokenFelt *felt.Felt

	if isStarknetNetwork(destChainName) {
		// If destination is Starknet, use Starknet's Alice address and DogCoin
		starknetAliceAddr := envutil.GetStarknetAliceAddress()
		if starknetAliceAddr == "" {
			log.Fatalf("Starknet Alice address not set")
		}
		recipientFelt, _ = utils.HexToFelt(starknetAliceAddr)

		// Get Starknet DogCoin address
		starknetDogCoin := getEnvWithDefault("STARKNET_DOG_COIN_ADDRESS", "")
		if starknetDogCoin == "" {
			log.Fatalf("STARKNET_DOG_COIN_ADDRESS not set")
		}
		outputTokenFelt, _ = utils.HexToFelt(starknetDogCoin)
	} else {
		// If destination is EVM, get Alice's EVM address and pad it
		evmUserAddr := envutil.GetAlicePublicKey()
		if evmUserAddr == "" {
			log.Fatalf("Alice public key not set")
		}

		// Pad EVM address to 32 bytes for Cairo ContractAddress
		evmAddr := common.HexToAddress(evmUserAddr)
		paddedAddr := common.LeftPadBytes(evmAddr.Bytes(), 32)
		recipientFelt, _ = utils.HexToFelt(hex.EncodeToString(paddedAddr))

		// Get DogCoin address from destination network config (.env)
		if _, exists := config.Networks[destChainName]; exists {
			dogCoinAddr := getEnvWithDefault(strings.ToUpper(destChainName)+"_DOG_COIN_ADDRESS", "")
			if dogCoinAddr != "" {
				// For EVM addresses, we need to left-pad to 32 bytes for Cairo ContractAddress
				evmAddr := common.HexToAddress(dogCoinAddr)
				paddedAddr := common.LeftPadBytes(evmAddr.Bytes(), 32)
				outputTokenFelt, _ = utils.HexToFelt(hex.EncodeToString(paddedAddr))
			} else {
				log.Fatalf("No %s_DOG_COIN_ADDRESS in .env", strings.ToUpper(destChainName))
			}
		} else {
			log.Fatalf("Destination network %s not found in config", destChainName)
		}
	}

	// Destination settler must be the Hyperlane address for the destination network
	destSettlerHex := ""
	if staticAddr, err := config.GetHyperlaneAddress(destChainName); err == nil {
		destSettlerHex = staticAddr.Hex()
	} else if destNetwork, exists := config.Networks[destChainName]; exists {
		destSettlerHex = destNetwork.HyperlaneAddress.Hex()
	}
	if destSettlerHex == "" {
		log.Fatalf("Could not get destination settler address for %s", destChainName)
	}

	// Ensure destination settler is properly padded to 32 bytes for Cairo ContractAddress
	var destSettlerFelt *felt.Felt
	if isStarknetNetwork(destChainName) {
		// If destination is Starknet, use Starknet address directly
		destSettlerFelt, _ = utils.HexToFelt(destSettlerHex)
	} else {
		// If destination is EVM, pad the EVM address to 32 bytes
		evmAddr := common.HexToAddress(destSettlerHex)
		paddedAddr := common.LeftPadBytes(evmAddr.Bytes(), 32)
		destSettlerFelt, _ = utils.HexToFelt(hex.EncodeToString(paddedAddr))
	}

	return StarknetOrderData{
		Sender:             userAddrFelt,
		Recipient:          recipientFelt,
		InputToken:         inputTokenFelt,
		OutputToken:        outputTokenFelt,
		AmountIn:           order.InputAmount,
		AmountOut:          order.OutputAmount,
		SenderNonce:        utils.BigIntToFelt(senderNonce),
		OriginDomain:       originDomain,
		DestinationDomain:  destinationDomain,
		DestinationSettler: destSettlerFelt,
		OpenDeadline:       order.OpenDeadline,
		FillDeadline:       order.FillDeadline,
		Data:               []*felt.Felt{},
	}
}

