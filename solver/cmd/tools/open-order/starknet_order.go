package openorder

// Starknet order creation logic - extracted from the original open-order/starknet/main.go
// This handles creating orders on Starknet chains

import (
	"context"
	"crypto/rand"
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

// getAliceAddressForNetwork gets Alice's address for a specific network using IS_DEVNET logic
func getAliceAddressForNetwork(networkName string) (string, error) {
	if strings.Contains(strings.ToLower(networkName), "starknet") {
		// Use conditional environment variable for Starknet
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

// NetworkConfig represents a single network configuration for Starknet
type StarknetNetworkConfig struct {
	name             string
	url              string
	chainID          uint64
	hyperlaneAddress string
	dogCoinAddress   string
}

// OrderConfig represents order configuration for Starknet
type StarknetOrderConfig struct {
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

// OrderData struct matching the Cairo OrderData
type StarknetOrderData struct {
	Sender             *felt.Felt
	Recipient          *felt.Felt
	InputToken         *felt.Felt
	OutputToken        *felt.Felt
	AmountIn           *big.Int
	AmountOut          *big.Int
	SenderNonce        *felt.Felt
	OriginDomain       uint32
	DestinationDomain  uint32
	DestinationSettler *felt.Felt
	OpenDeadline       uint64
	FillDeadline       uint64
	Data               []*felt.Felt
}

// StarknetOnchainCrossChainOrder struct matching the Cairo interface
type StarknetOnchainCrossChainOrder struct {
	FillDeadline      uint64
	OrderDataTypeLow  *felt.Felt
	OrderDataTypeHigh *felt.Felt
	OrderData         []*felt.Felt
}

// Test user configuration for Starknet
var starknetTestUsers []struct {
	name       string
	privateKey string
	address    string
}

// initializeStarknetTestUsers initializes the test user configuration after .env is loaded
func initializeStarknetTestUsers() {
	// Use envutil for conditional environment variable access
	aliceAddr := envutil.GetStarknetAliceAddress()
	solverAddr := envutil.GetStarknetSolverAddress()
	alicePrivateKeyVar := envutil.GetConditionalAccountEnv("STARKNET_ALICE_PRIVATE_KEY")
	solverPrivateKeyVar := envutil.GetConditionalAccountEnv("STARKNET_SOLVER_PRIVATE_KEY")

	starknetTestUsers = []struct {
		name       string
		privateKey string
		address    string
	}{
		{"Alice", alicePrivateKeyVar, aliceAddr},
		{"Solver", solverPrivateKeyVar, solverAddr},
	}
}

// loadStarknetNetworks loads network configuration from centralized config and environment variables
func loadStarknetNetworks() []StarknetNetworkConfig {
	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	// Build networks from centralized config
	networkNames := config.GetNetworkNames()
	networks := make([]StarknetNetworkConfig, 0, len(networkNames))

	for _, networkName := range networkNames {
		// Only include Starknet networks
		if networkName != "Starknet" {
			continue
		}

		networkConfig := config.Networks[networkName]

		// Load addresses from environment variables
		hyperlaneAddr := getEnvWithDefault("STARKNET_HYPERLANE_ADDRESS", "")
		dogAddr := getEnvWithDefault("STARKNET_DOG_COIN_ADDRESS", "")

		if hyperlaneAddr == "" || dogAddr == "" {
			log.Fatalf("missing STARKNET_HYPERLANE_ADDRESS or STARKNET_DOG_COIN_ADDRESS in .env")
		}

		// fmt.Printf("   🔍 Loaded %s DogCoin from env: %s\n", networkName, dogAddr)

		networks = append(networks, StarknetNetworkConfig{
			name:             networkConfig.Name,
			url:              networkConfig.RPCURL,
			chainID:          networkConfig.ChainID,
			hyperlaneAddress: hyperlaneAddr,
			dogCoinAddress:   dogAddr,
		})
	}

	return networks
}

// RunStarknetOrder creates a Starknet order based on the command
func RunStarknetOrder(command string) {
	fmt.Println("🎯 Running Starknet order creation...")

	// Load configuration (this loads .env and initializes networks)
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize test users after .env is loaded
	initializeStarknetTestUsers()

	// Load network configuration
	networks := loadStarknetNetworks()

	switch command {
	case "random":
		openRandomStarknetOrder(networks)
	case "default":
		openDefaultStarknetToEvm(networks)
	default:
		// Default to random Starknet order
		openRandomStarknetOrder(networks)
	}
}

func openRandomStarknetOrder(networks []StarknetNetworkConfig) {
	fmt.Println("🎲 Opening Random Starknet Test Order...")

	// Use configured Starknet network as origin
	originChain := "Starknet"

	// Get available destination networks from config
	destinationChain := getRandomDestinationChain(originChain)

	// Get Alice's address for the destination chain
	user, err := getAliceAddressForNetwork(destinationChain)
	if err != nil {
		log.Fatalf("Failed to get Alice address for %s: %v", destinationChain, err)
	}

	// Random amounts
	inputAmount := CreateTokenAmount(int64(secureRandomInt(9901)+100), 18) // 100-10000 tokens
	delta := CreateTokenAmount(int64(secureRandomInt(10)+1), 18)           // 1-10 tokens
	outputAmount := new(big.Int).Sub(inputAmount, delta)             // slightly less to ensure it's fillable

	order := StarknetOrderConfig{
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

	executeStarknetOrder(order, networks)
}

func openDefaultStarknetToEvm(networks []StarknetNetworkConfig) {
	fmt.Println("🎯 Opening Default Starknet → EVM Test Order...")

	// Use configured networks instead of hardcoded names
	originChain := "Starknet"
	destinationChain := getEnvWithDefault("DEFAULT_EVM_DESTINATION", "Ethereum")

	// Get Alice's address for the destination chain
	aliceAddress, err := getAliceAddressForNetwork(destinationChain)
	if err != nil {
		log.Fatalf("Failed to get Alice address for %s: %v", destinationChain, err)
	}

	order := StarknetOrderConfig{
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

	executeStarknetOrder(order, networks)
}

func executeStarknetOrder(order StarknetOrderConfig, networks []StarknetNetworkConfig) {
	fmt.Printf("\n📋 Executing Order: %s → %s\n", order.OriginChain, order.DestinationChain)

	// Find origin network (should be Starknet)
	var originNetwork *StarknetNetworkConfig
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

	// Always use Alice's Starknet credentials for signing orders on Starknet
	// The order.User field contains the recipient address (destination chain), not the signer
	userKey := envutil.GetStarknetAlicePrivateKey()
	userPublicKey := envutil.GetStarknetAlicePublicKey()

	if userKey == "" || userPublicKey == "" {
		fmt.Printf("❌ Missing Alice's Starknet credentials (IS_DEVNET=%v)\n", envutil.IsDevnet())
		if envutil.IsDevnet() {
			fmt.Printf("   Required: LOCAL_STARKNET_ALICE_PRIVATE_KEY and LOCAL_STARKNET_ALICE_PUBLIC_KEY\n")
		} else {
			fmt.Printf("   Required: STARKNET_ALICE_PRIVATE_KEY and STARKNET_ALICE_PUBLIC_KEY\n")
		}
		os.Exit(1)
	}

	// Always use Alice's Starknet address for signing (order signer)
	var userAddr string
	for _, user := range starknetTestUsers {
		if user.name == "Alice" {
			userAddr = user.address
			break
		}
	}

	// Get domains from config
	var originDomain, destinationDomain uint32
	if originConfig, err := config.GetHyperlaneDomain(order.OriginChain); err == nil {
		originDomain = uint32(originConfig)
	} else {
		fmt.Printf("   ⚠️  Warning: Could not get origin domain from config, using chain ID\n")
		originDomain = uint32(originNetwork.chainID)
	}

	if destConfig, err := config.GetHyperlaneDomain(order.DestinationChain); err == nil {
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

	// Store initial balance for comparison
	// initialBalance := initialUserBalance

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

	// Build the order data
	orderData := buildStarknetOrderData(order, originNetwork, originDomain, destinationDomain, senderNonce, order.DestinationChain)

	// Build the StarknetOnchainCrossChainOrder with u256 order_data_type (low, high)
	lowHash, highHash := getOrderDataTypeHashU256()
	crossChainOrder := StarknetOnchainCrossChainOrder{
		FillDeadline:      order.FillDeadline,
		OrderDataTypeLow:  lowHash,
		OrderDataTypeHigh: highHash,
		OrderData:         encodeStarknetOrderData(orderData),
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

func buildStarknetOrderData(order StarknetOrderConfig, originNetwork *StarknetNetworkConfig, originDomain uint32, destinationDomain uint32, senderNonce *big.Int, destChainName string) StarknetOrderData {
	// Get the actual user address for the specified user
	var userAddr string
	for _, user := range starknetTestUsers {
		if user.name == order.User {
			userAddr = user.address
			break
		}
	}

	// Convert addresses to felt
	userAddrFelt, _ := utils.HexToFelt(userAddr)
	inputTokenFelt, _ := utils.HexToFelt(originNetwork.dogCoinAddress)

	// For Starknet→EVM orders, recipient is always Alice's EVM address
	// Use conditional environment variable based on IS_DEVNET
	evmUserAddr := envutil.GetAlicePublicKey()
	if evmUserAddr == "" {
		log.Fatalf("Alice public key not set")
	}

	var recipientFelt *felt.Felt

	// Pad EVM address to 32 bytes for Cairo ContractAddress
	evmAddr := common.HexToAddress(evmUserAddr)
	paddedAddr := common.LeftPadBytes(evmAddr.Bytes(), 32)
	recipientFelt, _ = utils.HexToFelt(hex.EncodeToString(paddedAddr))
	// fmt.Printf("   🔍 EVM recipient address for %s: %s (padded to 32 bytes: %s)\n", order.User, evmUserAddr, hex.EncodeToString(paddedAddr))

	// Output token should be from the destination network, not origin
	var outputTokenFelt *felt.Felt
	if isStarknetNetwork(destChainName) {
		// If destination is Starknet, use Starknet's DogCoin
		outputTokenFelt, _ = utils.HexToFelt(originNetwork.dogCoinAddress)
	} else {
		// If destination is EVM, get DogCoin address from destination network config (.env)
		if _, exists := config.Networks[destChainName]; exists {
			dogCoinAddr := getEnvWithDefault(strings.ToUpper(destChainName)+"_DOG_COIN_ADDRESS", "")
			if dogCoinAddr != "" {
				// For EVM addresses, we need to left-pad to 32 bytes for Cairo ContractAddress
				evmAddr := common.HexToAddress(dogCoinAddr)
				paddedAddr := common.LeftPadBytes(evmAddr.Bytes(), 32)
				outputTokenFelt, _ = utils.HexToFelt(hex.EncodeToString(paddedAddr))
				// fmt.Printf("   🔍 EVM output token address padded to 32 bytes: %s\n", hex.EncodeToString(paddedAddr))
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
	if staticAddr, err := config.GetHyperlaneAddress(destChainName); err == nil {
		destSettlerHex = staticAddr.Hex()
	} else if destNetwork, exists := config.Networks[destChainName]; exists {
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
		// fmt.Printf("   🔍 EVM destination settler address padded to 32 bytes: %s\n", hex.EncodeToString(paddedAddr))
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

func encodeStarknetOrderData(orderData StarknetOrderData) []*felt.Felt {
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
	writeWord(&raw, make([]byte, 0)) // length = 0 -> becomes 32 zero bytes

	// Now wrap into Cairo Bytes: size, words_len, then 16-byte words as felts
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

	//	// Log the first few words to see the structure
	//	for i := 0; i < len(words) && i < 5; i++ {
	//		fmt.Printf("     • word[%d]: %s\n", i, words[i].String())
	//	}
	//	if len(words) > 5 {
	//		fmt.Printf("     • ... and %d more words\n", len(words)-5)
	//	}

	return bytesStruct
}

// getRandomDestinationChain gets a random destination chain from available networks
func getRandomDestinationChain(originChain string) string {
	// Get all available networks from the internal config
	allNetworks := config.GetNetworkNames()

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
	destIdx := secureRandomInt(len(evmDestinations))
	return evmDestinations[destIdx]
}

// isStarknetNetwork checks if a network name represents a Starknet network
func isStarknetNetwork(networkName string) bool {
	// Check if network name contains "starknet" (case insensitive)
	return strings.Contains(strings.ToLower(networkName), "starknet")
}

// getEnvWithDefault gets an environment variable with a default fallback
// TODO: Remove this once all usages are migrated to envutil
func getEnvWithDefault(key, defaultValue string) string {
	return envutil.GetEnvWithDefault(key, defaultValue)
}

// secureRandomInt generates a secure random integer in the range [0, max)
func secureRandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	
	// Generate random bytes
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to time-based seed if crypto/rand fails
		return int(time.Now().UnixNano()) % max
	}
	
	// Convert bytes to int and mod by max
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3]) % max
}
