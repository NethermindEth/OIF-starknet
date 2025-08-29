package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"

	githubConfig "github.com/NethermindEth/oif-starknet/go/internal/config"
)

// Network configuration - will be loaded from deployment state
var networks []NetworkConfig

// NetworkConfig represents a single network configuration
type NetworkConfig struct {
	name             string
	url              string
	chainID          uint64
	hyperlaneAddress string
	dogCoinAddress   string
}

// loadNetworks loads network configuration from deployment state
func loadNetworks() error {
	// Base network entry - completely configurable via env vars
	networks = []NetworkConfig{
		{
			name:             "Starknet",
			url:              getEnvWithDefault("STARKNET_RPC_URL", "http://localhost:5050"),
			chainID:          getEnvUint64("STARKNET_CHAIN_ID", 23448591),
			hyperlaneAddress: "",
			dogCoinAddress:   "",
		},
	}

	// Check FORKING mode for address loading
	//forkingStr := strings.ToLower(getEnvWithDefault("FORKING", "true"))
	//isForking, _ := strconv.ParseBool(forkingStr)

	//	if isForking {
	//		// Local forks: Use deployment state
	//		fmt.Printf("🔄 FORKING=true: Loading addresses from deployment state (fork mode)\n")
	//		state, err := githubDeployer.GetDeploymentState()
	//		if err == nil {
	//			if sn, ok := state.Networks[starknetNetworkName]; ok {
	//				for i := range networks {
	//					if networks[i].name == starknetNetworkName {
	//						networks[i].hyperlaneAddress = sn.HyperlaneAddress
	//						networks[i].dogCoinAddress = sn.DogCoinAddress
	//						fmt.Printf("   🔍 Loaded centralized state for %s\n", networks[i].name)
	//						fmt.Printf("   🔍 Hyperlane7683: %s\n", networks[i].hyperlaneAddress)
	//						fmt.Printf("   🔍 DogCoin: %s\n", networks[i].dogCoinAddress)
	//						return nil
	//					}
	//				}
	//			}
	//		}
	//	} else {
	// Live networks: Use .env addresses
	for i := range networks {
		if networks[i].name == "Starknet" {
			hyperlaneAddr := getEnvWithDefault("STARKNET_HYPERLANE_ADDRESS", "")
			dogAddr := getEnvWithDefault("STARKNET_DOG_COIN_ADDRESS", "")

			if hyperlaneAddr == "" || dogAddr == "" {
				return fmt.Errorf("missing STARKNET_HYPERLANE_ADDRESS or STARKNET_DOG_COIN_ADDRESS in .env")
			}

			networks[i].hyperlaneAddress = hyperlaneAddr
			networks[i].dogCoinAddress = dogAddr
			if dogAddr != "" {
				fmt.Printf("   🔄 Using %s DogCoin from .env: %s\n", networks[i].name, dogAddr)
			}
			return nil
		}
		//		}
	}

	//// If neither mode loaded addresses successfully, try fallback to legacy files
	//if networks[0].hyperlaneAddress == "" {
	//	fmt.Printf("   ⚠️ Fallback: Trying legacy per-file state\n")
	//	for i, network := range networks {
	//		if network.name == starknetNetworkName {
	//			if hyperlaneAddr, err := loadHyperlaneAddress(); err == nil && hyperlaneAddr != "" {
	//				networks[i].hyperlaneAddress = hyperlaneAddr
	//				fmt.Printf("   🔍 Loaded %s Hyperlane7683: %s\n", network.name, hyperlaneAddr)
	//			}
	//			if tokens, err := loadTokenAddresses(); err == nil {
	//				for _, token := range tokens {
	//					if token.Name == "DogCoin" {
	//						networks[i].dogCoinAddress = token.Address
	//						fmt.Printf("   🔍 Loaded %s DogCoin: %s\n", network.name, token.Address)
	//					}
	//				}
	//			}
	//		}
	//	}
	//}

	return nil
}

//// loadHyperlaneAddress loads the Hyperlane7683 address from deployment file
//func loadHyperlaneAddress() (string, error) {
//	// Try multiple possible paths
//	paths := []string{
//		"state/network_state/starknet-sepolia-deployment.json",
//		"../state/network_state/starknet-sepolia-deployment.json",
//		"../../state/network_state/starknet-sepolia-deployment.json",
//	}
//
//	for _, path := range paths {
//		data, err := os.ReadFile(path)
//		if err == nil {
//			fmt.Printf("   🔍 Loaded Hyperlane address from: %s\n", path)
//			var deployment struct {
//				DeployedAddress string `json:"deployedAddress"`
//			}
//			if err := json.Unmarshal(data, &deployment); err != nil {
//				continue
//			}
//			return deployment.DeployedAddress, nil
//		}
//	}
//
//	return "", fmt.Errorf("could not find Hyperlane deployment file in any of the expected paths")
//}

// loadTokenAddresses loads token addresses from deployment file
//func loadTokenAddresses() ([]TokenInfo, error) {
//	// Try multiple possible paths
//	paths := []string{
//		"state/network_state/starknet-mock-erc20-deployment.json",
//		"../state/network_state/starknet-mock-erc20-deployment.json",
//		"../../state/network_state/deployment-state.json",
//	}
//
//	for _, path := range paths {
//		data, err := os.ReadFile(path)
//		if err == nil {
//			fmt.Printf("   🔍 Loaded token addresses from: %s\n", path)
//			var deployment struct {
//				Tokens []TokenInfo `json:"tokens"`
//			}
//			if err := json.Unmarshal(data, &deployment); err != nil {
//				continue
//			}
//			return deployment.Tokens, nil
//		}
//	}
//
//	return nil, fmt.Errorf("could not find token deployment file in any of the expected paths")
//}

// TokenInfo represents token deployment information
type TokenInfo struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Address   string `json:"address"`
	ClassHash string `json:"classHash"`
}

// Test user configuration (Alice-only for orders, Solver for fills)
// NOTE: This gets initialized after .env is loaded to ensure addresses come from environment
var testUsers []struct {
	name       string
	privateKey string
	address    string
}

// initializeTestUsers initializes the test user configuration after .env is loaded
func initializeTestUsers() {
	testUsers = []struct {
		name       string
		privateKey string
		address    string
	}{
		{"Alice", "STARKNET_ALICE_PRIVATE_KEY", getEnvWithDefault("STARKNET_ALICE_ADDRESS", "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7")},
		{"Solver", "STARKNET_SOLVER_PRIVATE_KEY", getEnvWithDefault("STARKNET_SOLVER_ADDRESS", "0x2af9427c5a277474c079a1283c880ee8a6f0f8fbf73ce969c08d88befec1bba")},
	}
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
	OpenDeadline     uint64 // Changed from uint32 to uint64
	FillDeadline     uint64 // Changed from uint32 to uint64
}

// OrderData struct matching the Cairo OrderData
type OrderData struct {
	Sender             *felt.Felt
	Recipient          *felt.Felt
	InputToken         *felt.Felt
	OutputToken        *felt.Felt
	AmountIn           *big.Int // Changed from *felt.Felt to *big.Int for u256 splitting
	AmountOut          *big.Int // Changed from *felt.Felt to *big.Int for u256 splitting
	SenderNonce        *felt.Felt
	OriginDomain       uint32
	DestinationDomain  uint32
	DestinationSettler *felt.Felt
	OpenDeadline       uint64       // Added missing field
	FillDeadline       uint64       // Changed from uint32 to uint64
	Data               []*felt.Felt // Empty for basic orders
}

// OnchainCrossChainOrder struct matching the Cairo interface
// NOTE: order_data_type is now u256 → two felts (low, high)
type OnchainCrossChainOrder struct {
	FillDeadline      uint64
	OrderDataTypeLow  *felt.Felt
	OrderDataTypeHigh *felt.Felt
	OrderData         []*felt.Felt
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	// Initialize network configuration after loading .env
	githubConfig.InitializeNetworks()

	// Initialize test users after .env is loaded
	initializeTestUsers()

	// Load network configuration
	if err := loadNetworks(); err != nil {
		fmt.Printf("❌ Failed to load networks: %v\n", err)
		os.Exit(1)
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: open-starknet-order <command>")
		fmt.Println("Commands:")
		fmt.Println("  default       - Open default Starknet→EVM order")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "random":
		openRandomOrder()
	case "default":
		openDefaultStarknetToEvm()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func openRandomOrder() {
	fmt.Println("🎲 Opening Random Starknet Test Order...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Use configured Starknet network as origin
	originChain := getEnvWithDefault("STARKNET_NETWORK_NAME", "Starknet")

	// Get available destination networks from config
	destinationChain := getRandomDestinationChain(originChain)

	// Always use Alice for orders
	user := "Alice"

	// Random amounts
	inputAmount :=
		new(big.Int).Mul(big.NewInt(int64(rand.Intn(9901)+100)), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)) // 100-10000 tokens
	delta := big.NewInt(int64(rand.Intn(90) + 1))        // 1-90
	outputAmount := new(big.Int).Sub(inputAmount, delta) // slightly less to ensure it's fillable

	order := OrderConfig{
		OriginChain:      originChain,
		DestinationChain: destinationChain,
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      inputAmount,
		OutputAmount:     outputAmount,
		User:             user,
		OpenDeadline:     uint64(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint64(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func openDefaultStarknetToEvm() {
	fmt.Println("🎯 Opening Default Starknet → EVM Test Order (Nonce: 3)...")

	// Use configured networks instead of hardcoded names
	originChain := getEnvWithDefault("STARKNET_NETWORK_NAME", "Starknet")
	destinationChain := getEnvWithDefault("DEFAULT_EVM_DESTINATION", "Ethereum")

	order := OrderConfig{
		OriginChain:      originChain,
		DestinationChain: destinationChain,
		InputToken:       "DogCoin",
		OutputToken:      "DogCoin",
		InputAmount:      new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 tokens
		OutputAmount:     new(big.Int).Mul(big.NewInt(999), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),  // 999 tokens
		User:             "Alice",
		OpenDeadline:     uint64(time.Now().Add(1 * time.Hour).Unix()),
		FillDeadline:     uint64(time.Now().Add(24 * time.Hour).Unix()),
	}

	executeOrder(order)
}

func executeOrder(order OrderConfig) {
	fmt.Printf("\n📋 Executing Order: %s → %s\n", order.OriginChain, order.DestinationChain)

	// Find origin network (should be Starknet)
	var originNetwork *NetworkConfig
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

	// Connect to Starknet RPC
	client, err := rpc.NewProvider(originNetwork.url)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s: %v\n", order.OriginChain, err)
		os.Exit(1)
	}

	// Get user private key and public key
	userKey := os.Getenv(fmt.Sprintf("STARKNET_%s_PRIVATE_KEY", strings.ToUpper(order.User)))
	userPublicKey := os.Getenv(fmt.Sprintf("STARKNET_%s_PUBLIC_KEY", strings.ToUpper(order.User)))
	if userKey == "" || userPublicKey == "" {
		fmt.Printf("❌ Missing credentials for user: %s\n", order.User)
		fmt.Printf("   Required: STARKNET_%s_PRIVATE_KEY and STARKNET_%s_PUBLIC_KEY\n", strings.ToUpper(order.User), strings.ToUpper(order.User))
		os.Exit(1)
	}

	// Get user address
	var userAddr string
	for _, user := range testUsers {
		if user.name == order.User {
			userAddr = user.address
			break
		}
	}

	// Get domains from config
	var originDomain, destinationDomain uint32
	if originConfig, err := githubConfig.GetHyperlaneDomain(order.OriginChain); err == nil {
		originDomain = uint32(originConfig)
	} else {
		fmt.Printf("   ⚠️  Warning: Could not get origin domain from config, using chain ID\n")
		originDomain = uint32(originNetwork.chainID)
	}

	if destConfig, err := githubConfig.GetHyperlaneDomain(order.DestinationChain); err == nil {
		destinationDomain = uint32(destConfig)
	} else {
		fmt.Printf("   ⚠️  Warning: Could not get destination domain from config\n")
		os.Exit(1)
	}

	// Preflight: check balances and allowances
	inputToken := originNetwork.dogCoinAddress
	owner := userAddr

	// Get initial balances
	initialUserBalance, err := getTokenBalanceFromRPC(client, inputToken, owner)
	if err == nil {
		fmt.Printf("   🔍 Initial InputToken balance(owner): %s\n", formatTokenAmount(initialUserBalance))
	} else {
		fmt.Printf("   ⚠️  Could not read initial balance: %v\n", err)
	}

	// Store initial balance for comparison
	initialBalance := initialUserBalance

	// Generate a random nonce for the order
	senderNonce := big.NewInt(time.Now().UnixNano())

	// Build the order data
	orderData := buildOrderData(order, originNetwork, originDomain, destinationDomain, senderNonce, order.DestinationChain)

	// Build the OnchainCrossChainOrder with u256 order_data_type (low, high)
	lowHash, highHash := getOrderDataTypeHashU256()
	crossChainOrder := OnchainCrossChainOrder{
		FillDeadline:      order.FillDeadline,
		OrderDataTypeLow:  lowHash,
		OrderDataTypeHigh: highHash,
		OrderData:         encodeOrderData(orderData),
	}

	// Use generated bindings for open()
	fmt.Printf("   📝 Calling open() function...\n")

	// Create user account for transaction signing
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
	fmt.Printf("   🎯 Order ID: %s\n", calculateOrderId(orderData))

	// Verify that balances actually changed as expected
	fmt.Printf("   🔍 Verifying balance changes...\n")
	if err := verifyBalanceChangesFromRPC(client, inputToken, owner, initialBalance, order.InputAmount); err != nil {
		fmt.Printf("   ⚠️  Balance verification failed: %v\n", err)
	} else {
		fmt.Printf("   👍 Balance changes verified!\n")
	}

	fmt.Printf("\n🎉 Order execution completed!\n")
}

func buildOrderData(order OrderConfig, originNetwork *NetworkConfig, originDomain uint32, destinationDomain uint32, senderNonce *big.Int, destChainName string) OrderData {
	// Get the actual user address for the specified user
	var userAddr string
	for _, user := range testUsers {
		if user.name == order.User {
			userAddr = user.address
			break
		}
	}

	// Convert addresses to felt
	userAddrFelt, _ := utils.HexToFelt(userAddr)
	inputTokenFelt, _ := utils.HexToFelt(originNetwork.dogCoinAddress)

	// For recipient, since this is a Starknet order opener, destination is always EVM
	// We need to map the Starknet user to their corresponding EVM address
	var recipientFelt *felt.Felt
	var evmUserAddr string

	// Map Starknet users to their EVM addresses
	switch order.User {
	case "Alice":
		evmUserAddr = getEnvWithDefault("ALICE_PUB_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	case "Solver":
		evmUserAddr = getEnvWithDefault("SOLVER_PUB_KEY", "0x90F79bf6EB2c4f870365E785982E1f101E93b906")
	default:
		// Fallback to Alice address if unknown user (should only be Alice now)
		evmUserAddr = getEnvWithDefault("ALICE_PUB_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
		fmt.Printf("   ⚠️  Warning: Unknown user %s, using Alice EVM address as recipient\n", order.User)
	}

	// Pad EVM address to 32 bytes for Cairo ContractAddress
	evmAddr := common.HexToAddress(evmUserAddr)
	paddedAddr := common.LeftPadBytes(evmAddr.Bytes(), 32)
	recipientFelt, _ = utils.HexToFelt(hex.EncodeToString(paddedAddr))
	fmt.Printf("   🔍 EVM recipient address for %s: %s (padded to 32 bytes: %s)\n", order.User, evmUserAddr, hex.EncodeToString(paddedAddr))

	// Output token should be from the destination network, not origin
	var outputTokenFelt *felt.Felt
	if isStarknetNetwork(destChainName) {
		// If destination is Starknet, use Starknet's DogCoin
		outputTokenFelt, _ = utils.HexToFelt(originNetwork.dogCoinAddress)
	} else {
		// If destination is EVM, get DogCoin address from destination network config (.env)
		if _, exists := githubConfig.Networks[destChainName]; exists {
			dogCoinAddr := getEnvWithDefault(strings.ToUpper(destChainName)+"_DOG_COIN_ADDRESS", "")
			if dogCoinAddr != "" {
				// For EVM addresses, we need to left-pad to 32 bytes for Cairo ContractAddress
				evmAddr := common.HexToAddress(dogCoinAddr)
				paddedAddr := common.LeftPadBytes(evmAddr.Bytes(), 32)
				outputTokenFelt, _ = utils.HexToFelt(hex.EncodeToString(paddedAddr))
				fmt.Printf("   🔍 EVM output token address padded to 32 bytes: %s\n", hex.EncodeToString(paddedAddr))
			} else {
				// Last resort - use origin network (this is wrong but prevents crash)
				outputTokenFelt, _ = utils.HexToFelt(originNetwork.dogCoinAddress)
				fmt.Printf("   ⚠️  Warning: No %s_DOG_COIN_ADDRESS in .env, using origin network DogCoin as fallback\n", strings.ToUpper(destChainName))
			}
		} else {
			// Fallback to origin network if destination network not found
			outputTokenFelt, _ = utils.HexToFelt(originNetwork.dogCoinAddress)
			fmt.Printf("   ⚠️  Warning: Destination network %s not found in config, using origin network DogCoin as fallback\n", destChainName)
		}
	}

	// Destination settler must be the EVM Hyperlane address for the destination network
	destSettlerHex := ""
	if staticAddr, err := githubConfig.GetHyperlaneAddress(destChainName); err == nil {
		destSettlerHex = staticAddr.Hex()
	} else if destNetwork, exists := githubConfig.Networks[destChainName]; exists {
		destSettlerHex = destNetwork.HyperlaneAddress.Hex()
	}
	if destSettlerHex == "" {
		// As a last resort, keep previous behavior (but this is likely wrong for cross-chain)
		destSettlerHex = originNetwork.hyperlaneAddress
		fmt.Printf("   ⚠️  Warning: Using origin Hyperlane address as destination settler (may be incorrect)\n")
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
		fmt.Printf("   🔍 EVM destination settler address padded to 32 bytes: %s\n", hex.EncodeToString(paddedAddr))
	}

	return OrderData{
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

func getOrderDataTypeHashU256() (low *felt.Felt, high *felt.Felt) {
	// Solidity ORDER_DATA_TYPE_HASH (32 bytes)
	hashHex := getEnvWithDefault("ORDER_DATA_TYPE_HASH", "0x08d75650babf4de09c9273d48ef647876057ed91d4323f8a2e3ebc2cd8a63b5e")
	bi, ok := new(big.Int).SetString(hashHex, 0)
	if !ok {
		panic("Failed to parse ORDER_DATA_TYPE_HASH")
	}
	// Split into low/high 128-bit felts: low = lower 128, high = upper 128
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	lowBI := new(big.Int).And(bi, mask)
	highBI := new(big.Int).Rsh(bi, 128)
	return utils.BigIntToFelt(lowBI), utils.BigIntToFelt(highBI)
}

func encodeOrderData(orderData OrderData) []*felt.Felt {
	//// Log the input values for debugging
	//fmt.Printf("   🧪 OrderData Input Debug:\n")
	//fmt.Printf("     • sender: %s\n", orderData.Sender.String())
	//fmt.Printf("     • recipient: %s\n", orderData.Recipient.String())
	//fmt.Printf("     • input_token: %s\n", orderData.InputToken.String())
	//fmt.Printf("     • output_token: %s\n", orderData.OutputToken.String())
	//fmt.Printf("     • amount_in: %s\n", orderData.AmountIn.String())
	//fmt.Printf("     • amount_out: %s\n", orderData.AmountOut.String())
	//fmt.Printf("     • sender_nonce: %s\n", orderData.SenderNonce.String())
	//fmt.Printf("     • origin_domain: %d (u32)\n", orderData.OriginDomain)
	//fmt.Printf("     • destination_domain: %d (u32)\n", orderData.DestinationDomain)
	//fmt.Printf("     • destination_settler: %s\n", orderData.DestinationSettler.String())
	//fmt.Printf("     • fill_deadline: %d (u64)\n", orderData.FillDeadline)
	//fmt.Printf("     • data: %d elements\n", len(orderData.Data))

	// Build Solidity ABI-compatible bytes for OrderData
	// Head (12 words of 32 bytes), then tail for bytes data (length + data padded)

	leftPad := func(src []byte, size int) []byte {
		if len(src) >= size {
			return src[len(src)-size:]
		}
		out := make([]byte, size)
		copy(out[size-len(src):], src)
		return out
	}

	feltToBytes32 := func(f *felt.Felt) []byte {
		b := f.Bytes()
		// Ensure we always return exactly 32 bytes
		result := make([]byte, 32)

		if len(b) >= 32 {
			// Take the last 32 bytes if longer
			start := len(b) - 32
			for i := 0; i < 32; i++ {
				result[i] = b[start+i]
			}
		} else {
			// Pad with zeros if shorter
			start := 32 - len(b)
			for i := 0; i < len(b); i++ {
				result[start+i] = b[i]
			}
		}

		return result
	}

	u256ToBytes32 := func(n *big.Int) []byte {
		if n == nil {
			return make([]byte, 32)
		}
		return leftPad(n.Bytes(), 32)
	}

	u32Word := func(v uint32) []byte {
		b := make([]byte, 32)
		b[28] = byte(v >> 24)
		b[29] = byte(v >> 16)
		b[30] = byte(v >> 8)
		b[31] = byte(v)
		return b
	}

	writeWord := func(dst *[]byte, word []byte) {
		*dst = append(*dst, leftPad(word, 32)...)
	}

	var raw []byte

	// 0) Dynamic data offset (32 bytes) - this is the offset to where the data field starts
	// In Solidity ABI encoding, this is the first field when you have dynamic data
	writeWord(&raw, u32Word(32))

	// 1) sender, recipient, input_token, output_token (bytes32)
	writeWord(&raw, feltToBytes32(orderData.Sender))
	writeWord(&raw, feltToBytes32(orderData.Recipient))
	writeWord(&raw, feltToBytes32(orderData.InputToken))
	writeWord(&raw, feltToBytes32(orderData.OutputToken))
	// 2) amount_in, amount_out (uint256)
	writeWord(&raw, u256ToBytes32(orderData.AmountIn))
	writeWord(&raw, u256ToBytes32(orderData.AmountOut))
	// 3) sender_nonce (uint256 from felt)
	writeWord(&raw, feltToBytes32(orderData.SenderNonce))
	// 4) origin_domain (uint32), destination_domain (uint32) as 32-byte words
	writeWord(&raw, u32Word(orderData.OriginDomain))
	writeWord(&raw, u32Word(orderData.DestinationDomain))
	// 5) destination_settler (bytes32)
	writeWord(&raw, feltToBytes32(orderData.DestinationSettler))
	// 6) fill_deadline (uint32) in 32-byte word (clamp lower 32 bits)
	fill32 := uint32(orderData.FillDeadline)
	writeWord(&raw, u32Word(fill32))
	// 7) data offset (32 * 12 = 384)
	writeWord(&raw, u32Word(384))

	// Tail: data length (32 bytes) then data padded to 32
	// Currently no dynamic data, so encode zero length
	writeWord(&raw, make([]byte, 0)) // length = 0 -> becomes 32 zero bytes

	// Now wrap into Cairo Bytes: size, words_len, then 16-byte words as felts
	// split into u128 words (16 bytes)
	numElements := (len(raw) + 15) / 16
	words := make([]*felt.Felt, 0, numElements)
	for i := 0; i < len(raw); i += 16 {
		end := i + 16
		if end > len(raw) {
			chunk := make([]byte, 16)
			copy(chunk, raw[i:])
			words = append(words, utils.BigIntToFelt(new(big.Int).SetBytes(chunk)))
		} else {
			words = append(words, utils.BigIntToFelt(new(big.Int).SetBytes(raw[i:end])))
		}
	}

	bytesStruct := make([]*felt.Felt, 0, 2+len(words))
	bytesStruct = append(bytesStruct, utils.Uint64ToFelt(uint64(len(raw))))
	bytesStruct = append(bytesStruct, utils.Uint64ToFelt(uint64(len(words))))
	bytesStruct = append(bytesStruct, words...)

	// Log the first few words to see the structure
	for i := 0; i < len(words) && i < 5; i++ {
		fmt.Printf("     • word[%d]: %s\n", i, words[i].String())
	}
	if len(words) > 5 {
		fmt.Printf("     • ... and %d more words\n", len(words)-5)
	}

	return bytesStruct
}

// toU256 converts a big.Int to low and high felt values for u256 representation
func toU256(num *big.Int) (low, high *felt.Felt) {
	// Create a mask for the lower 128 bits
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

	// Extract low and high parts
	lowBigInt := new(big.Int).And(num, mask)
	highBigInt := new(big.Int).Rsh(num, 128)

	// Convert to felt
	lowFelt := utils.BigIntToFelt(lowBigInt)
	highFelt := utils.BigIntToFelt(highBigInt)

	return lowFelt, highFelt
}

func calculateOrderId(orderData OrderData) string {
	// Generate a simple order ID for now
	// In production, this should match the contract's order ID generation
	return fmt.Sprintf("sn_order_%d", time.Now().UnixNano())
}

// getTokenBalance gets the balance of a token for a specific address using RPC
func getTokenBalanceFromRPC(client rpc.RpcProvider, tokenAddress, userAddress string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	userAddrFelt, err := utils.HexToFelt(userAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid user address: %w", err)
	}

	// Build the balanceOf function call
	balanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
		Calldata:           []*felt.Felt{userAddrFelt},
	}

	// Call the contract to get balance
	resp, err := client.Call(context.Background(), balanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no response from balanceOf call")
	}

	// Convert felt response to big.Int
	balanceFelt := resp[0]
	balanceBigInt := utils.FeltToBigInt(balanceFelt)

	return balanceBigInt, nil
}

// getTokenAllowance gets the allowance of a token for a specific spender using RPC
func getTokenAllowanceFromRPC(client rpc.RpcProvider, tokenAddress, ownerAddress, spenderAddress string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	ownerAddrFelt, err := utils.HexToFelt(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	spenderAddrFelt, err := utils.HexToFelt(spenderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid spender address: %w", err)
	}

	// Build the allowance function call
	allowanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{ownerAddrFelt, spenderAddrFelt},
	}

	// Call the contract to get allowance
	resp, err := client.Call(context.Background(), allowanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance: %w", err)
	}

	if len(resp) < 2 {
		return nil, fmt.Errorf("insufficient response from allowance call: expected 2 values for u256, got %d", len(resp))
	}

	// For u256, the response should be [low, high] where:
	// - low contains the first 128 bits
	// - high contains the remaining 128 bits
	lowFelt := resp[0]
	highFelt := resp[1]

	// Convert low and high felts to big.Ints
	lowBigInt := utils.FeltToBigInt(lowFelt)
	highBigInt := utils.FeltToBigInt(highFelt)

	// Combine low and high into a single u256 value
	// high << 128 + low
	shiftedHigh := new(big.Int).Lsh(highBigInt, 128)
	totalAllowance := new(big.Int).Add(shiftedHigh, lowBigInt)

	return totalAllowance, nil
}

// verifyBalanceChanges verifies that opening an order actually transferred tokens using RPC
func verifyBalanceChangesFromRPC(client rpc.RpcProvider, tokenAddress, userAddress string, initialBalance, expectedTransferAmount *big.Int) error {
	// Wait a moment for the transaction to be fully processed
	time.Sleep(2 * time.Second)

	// Get final balance
	finalUserBalance, err := getTokenBalanceFromRPC(client, tokenAddress, userAddress)
	if err != nil {
		return fmt.Errorf("failed to get final user balance: %w", err)
	}

	// Calculate actual change
	userBalanceChange := new(big.Int).Sub(initialBalance, finalUserBalance)

	// Print balance changes
	fmt.Printf("     💰 User balance change: %s → %s (Δ: %s)\n",
		formatTokenAmount(initialBalance),
		formatTokenAmount(finalUserBalance),
		formatTokenAmount(userBalanceChange))

	// Verify the change matches expectations
	if userBalanceChange.Cmp(expectedTransferAmount) != 0 {
		return fmt.Errorf("user balance decreased by %s, expected %s",
			formatTokenAmount(userBalanceChange),
			formatTokenAmount(expectedTransferAmount))
	}

	return nil
}

// formatTokenAmount formats a token amount for display (converts from wei to tokens)
func formatTokenAmount(amount *big.Int) string {
	// Convert from wei (18 decimals) to tokens
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	tokens := new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(decimals))
	return tokens.Text('f', 0) + " tokens"
}

// getEnvWithDefault gets an environment variable with a default fallback
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvUint64 gets an environment variable as uint64 with a default fallback
func getEnvUint64(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		bi, ok := new(big.Int).SetString(value, 10)
		if ok {
			return bi.Uint64()
		}
		fmt.Printf("⚠️  Environment variable %s has invalid value: %s. Using default %d.\n", key, value, defaultValue)
	}
	return defaultValue
}

// getRandomDestinationChain gets a random destination chain from available networks
func getRandomDestinationChain(originChain string) string {
	// Get all available networks from the internal config
	allNetworks := githubConfig.GetNetworkNames()

	// Filter out the origin chain and non-EVM networks
	var evmDestinations []string
	for _, networkName := range allNetworks {
		if networkName != originChain && !isStarknetNetwork(networkName) {
			evmDestinations = append(evmDestinations, networkName)
		}
	}

	// If no EVM networks found, use fallback
	if len(evmDestinations) == 0 {
		fmt.Printf("   ⚠️ No EVM networks found in config, using fallback destination\n")
		return getEnvWithDefault("DEFAULT_EVM_DESTINATION", "Sepolia")
	}

	// Select random destination
	destIdx := rand.Intn(len(evmDestinations))
	return evmDestinations[destIdx]
}

// isStarknetNetwork checks if a network name represents a Starknet network
func isStarknetNetwork(networkName string) bool {
	// Check if network name contains "starknet" (case insensitive)
	return strings.Contains(strings.ToLower(networkName), "starknet")
}
