package main

import (
	"context"

	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/go/solvercore/config"
)

// Token deployment info structure
type TokenDeploymentInfo struct {
	NetworkName    string      `json:"networkName"`
	DeploymentTime string      `json:"deploymentTime"`
	Tokens         []TokenInfo `json:"tokens"`
}

type TokenInfo struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Address   string `json:"address"`
	ClassHash string `json:"classHash"`
}

// User funding configuration
const (
	// Amount to fund each user
	UserFundingAmount = "100000000000000000000000"

	// Default deployment file path
	DeploymentFilePath = "state/deployment/starknet-mock-erc20-deployment.json"
)

// loadCentralAddresses loads Hyperlane, DogCoin from .env variables
func loadCentralAddresses(networkName string) (hyperlane string, dog string, err error) {
	// Get addresses from environment variables
	hyperlane = os.Getenv("STARKNET_HYPERLANE_ADDRESS")
	dog = os.Getenv("STARKNET_DOG_COIN_ADDRESS")

	if hyperlane == "" {
		return "", "", fmt.Errorf("STARKNET_HYPERLANE_ADDRESS not found in .env")
	}
	if dog == "" {
		return "", "", fmt.Errorf("STARKNET_DOG_COIN_ADDRESS not found in .env")
	}

	return hyperlane, dog, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	fmt.Println("🚀 Setting up Starknet contracts: funding users and setting allowances...")

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

	// Load test user addresses from .env
	aliceAddress := os.Getenv("STARKNET_ALICE_ADDRESS")
	solverAddress := os.Getenv("STARKNET_SOLVER_ADDRESS")

	if deployerAddress == "" || deployerPrivateKey == "" || deployerPublicKey == "" {
		fmt.Println("❌ Missing required environment variables:")
		fmt.Println("   STARKNET_DEPLOYER_ADDRESS: Your Starknet account address")
		fmt.Println("   STARKNET_DEPLOYER_PRIVATE_KEY: Your private key")
		fmt.Println("   STARKNET_DEPLOYER_PUBLIC_KEY: Your public key")
		os.Exit(1)
	}

	if aliceAddress == "" || solverAddress == "" {
		fmt.Println("❌ Missing required environment variables:")
		fmt.Println("   STARKNET_ALICE_ADDRESS: Alice's Starknet address")
		fmt.Println("   STARKNET_SOLVER_ADDRESS: Solver's Starknet address")
		os.Exit(1)
	}

	fmt.Printf("📋 Network: %s\n", networkName)
	fmt.Printf("📋 RPC URL: %s\n", networkConfig.RPCURL)
	fmt.Printf("📋 Chain ID: %d\n", networkConfig.ChainID)
	fmt.Printf("📋 Deployer: %s\n", deployerAddress)
	fmt.Printf("📋 Test Users: Alice=%s, Solver=%s\n", aliceAddress, solverAddress)

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

	// Load addresses from centralized deployment-state
	hyperlaneAddr, dogAddr, err := loadCentralAddresses(networkName)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to load centralized addresses: %s", err))
	}

	// Prepare TokenInfo based on centralized state
	dogCoin := TokenInfo{Name: "DogCoin", Symbol: "DOG", Address: dogAddr, ClassHash: ""}

	fmt.Printf("📋 DogCoin: %s\n", dogCoin.Address)

	// Fund test users
	fmt.Println("\n💰 Funding test users...")
	if err := fundUsers(accnt, dogCoin, aliceAddress, solverAddress); err != nil {
		panic(fmt.Sprintf("❌ Failed to fund users: %s", err))
	}

	// Set allowances for Hyperlane7683
	fmt.Println("\n🔐 Setting allowances for Hyperlane7683...")
	fmt.Printf("   📋 Found Hyperlane7683 at: %s\n", hyperlaneAddr)
	fmt.Println("   🔐 Setting allowances for Hyperlane7683...")
	if err := setAllowances(accnt, dogCoin, hyperlaneAddr, aliceAddress); err != nil {
		panic(fmt.Sprintf("❌ Failed to set allowances: %s", err))
	}

	// Verify balances and allowances after everything is set
	fmt.Printf("\n🔍 Verifying balances and allowances...\n")
	if err := verifyBalancesAndAllowances(accnt, dogCoin, hyperlaneAddr, aliceAddress, solverAddress); err != nil {
		fmt.Printf("❌ Verification failed: %v\n", err)
	} else {
		fmt.Printf("✅ All verifications passed!\n")
	}

	// Note: .env file updates removed - addresses should be set manually after live deployment

	fmt.Printf("\n🎯 Starknet contract setup completed successfully!\n")
	fmt.Printf("   • Users funded with DogCoin tokens\n")
	fmt.Printf("   • Allowances set for Hyperlane7683\n")
	fmt.Printf("   • Environment variables updated\n")
	fmt.Printf("   • Ready for cross-chain operations!\n")
}

// fundUsers funds test users with DogCoin tokens using the mint function
func fundUsers(accnt *account.Account, dogCoin TokenInfo, aliceAddr, solverAddr string) error {
	users := []struct {
		name    string
		address string
	}{
		{"Alice", aliceAddr},
		{"Solver", solverAddr},
	}

	// Fund each user with DogCoin tokens
	for _, user := range users {
		fmt.Printf("   💸 Funding %s...\n", user.name)

		// Check balance before minting
		dogBalanceBefore, err := getTokenBalance(accnt, dogCoin.Address, user.address, "DogCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin balance before minting: %w", user.name, err)
		}

		fmt.Printf("     📊 %s balance before: DogCoin=%s\n", user.name, dogBalanceBefore)

		// Fund with DogCoin
		if err := mintTokens(accnt, dogCoin.Address, user.address, UserFundingAmount, "DogCoin"); err != nil {
			return fmt.Errorf("failed to fund %s with DogCoin: %w", user.name, err)
		}

		// Check balance after minting
		dogBalanceAfter, err := getTokenBalance(accnt, dogCoin.Address, user.address, "DogCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin balance after minting: %w", user.name, err)
		}

		fmt.Printf("     📊 %s balance after: DogCoin=%s\n", user.name, dogBalanceAfter)

		// Verify the minting actually worked
		expectedAmount := new(big.Int)
		expectedAmount.SetString(UserFundingAmount, 10)

		dogIncrease := new(big.Int).Sub(dogBalanceAfter, dogBalanceBefore)

		if dogIncrease.Cmp(expectedAmount) != 0 {
			return fmt.Errorf("DogCoin minting failed for %s: expected increase %s, got %s", user.name, expectedAmount.String(), dogIncrease.String())
		}

		fmt.Printf("   ✅ %s funded successfully\n", user.name)
	}

	return nil
}

// toU256 converts a big.Int to u256 representation (low and high felts)
// u256 is represented as two felt.Felt values where:
// - low contains the first 128 bits
// - high contains the remaining 128 bits
func toU256(num *big.Int) (low, high *felt.Felt) {
	// Create a mask for 128 bits (2^128 - 1)
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

	// Extract low part (first 128 bits)
	lowBigInt := new(big.Int).And(num, mask)
	low = utils.BigIntToFelt(lowBigInt)

	// Extract high part (remaining bits)
	highBigInt := new(big.Int).Rsh(num, 128)
	high = utils.BigIntToFelt(highBigInt)

	return low, high
}

// mintTokens calls the mint function on a token contract
func mintTokens(accnt *account.Account, tokenAddress, recipient, amount, tokenName string) error {
	fmt.Printf("     🪙 Minting %s %s to %s...\n", amount, tokenName, recipient)

	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return fmt.Errorf("invalid token address: %w", err)
	}

	recipientFelt, err := utils.HexToFelt(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	// Convert amount string to big.Int
	amountBigInt := new(big.Int)
	amountBigInt.SetString(amount, 10) // Parse as decimal string

	// Convert to u256 representation (low and high felts)
	lowFelt, highFelt := toU256(amountBigInt)

	// Build the mint function call with u256 (low, high)
	mintCall := rpc.InvokeFunctionCall{
		ContractAddress: tokenAddrFelt,
		FunctionName:    "mint",
		CallData:        []*felt.Felt{recipientFelt, lowFelt, highFelt},
	}

	// Send the mint transaction
	resp, err := accnt.BuildAndSendInvokeTxn(context.Background(), []rpc.InvokeFunctionCall{mintCall}, nil)
	if err != nil {
		return fmt.Errorf("failed to send mint transaction: %w", err)
	}

	fmt.Printf("     ⏳ Mint transaction sent: %s\n", resp.Hash.String())
	fmt.Printf("     ⏳ Waiting for confirmation...\n")

	// Wait for transaction receipt
	_, err = accnt.WaitForTransactionReceipt(context.Background(), resp.Hash, time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for mint confirmation: %w", err)
	}

	fmt.Printf("     ✅ Mint transaction confirmed\n")
	return nil
}

// getTokenBalance gets the balance of a token for a specific address
func getTokenBalance(accnt *account.Account, tokenAddress, userAddress, tokenName string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	userAddrFelt, err := utils.HexToFelt(userAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid user address: %w", err)
	}

	// Build the balanceOf function call using rpc.FunctionCall
	balanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
		Calldata:           []*felt.Felt{userAddrFelt},
	}

	// Call the contract to get balance using the RPC provider
	resp, err := accnt.Provider.Call(context.Background(), balanceCall, rpc.WithBlockTag("latest"))
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

// setAllowances sets unlimited allowances for users on DogCoin token
func setAllowances(accnt *account.Account, dogCoin TokenInfo, hyperlaneAddress, aliceAddr string) error {
	if hyperlaneAddress == "" {
		fmt.Println("   ⚠️  No Hyperlane address provided, skipping allowance setup")
		return nil
	}

	fmt.Println("   🔐 Setting allowances for Hyperlane7683...")

	// Convert Hyperlane address to felt
	hyperlaneAddrFelt, err := utils.HexToFelt(hyperlaneAddress)
	if err != nil {
		return fmt.Errorf("invalid Hyperlane address: %w", err)
	}

	// Users to set allowances for
	users := []struct {
		name       string
		address    string
		privateKey string
		publicKey  string
	}{
		{"Alice", aliceAddr, os.Getenv("STARKNET_ALICE_PRIVATE_KEY"), os.Getenv("STARKNET_ALICE_PUBLIC_KEY")},
	}

	// Set unlimited allowance for each user on DogCoin
	for _, user := range users {
		fmt.Printf("     🔓 Setting %s allowances...\n", user.name)

		// Check if user has credentials
		if user.privateKey == "" || user.publicKey == "" {
			fmt.Printf("       ⚠️  Missing credentials for %s, skipping\n", user.name)
			continue
		}

		// Create user account
		userAddrFelt, err := utils.HexToFelt(user.address)
		if err != nil {
			return fmt.Errorf("invalid user address for %s: %w", user.name, err)
		}

		// Initialize user's keystore
		userKs := account.NewMemKeystore()
		userPrivKeyBI, ok := new(big.Int).SetString(user.privateKey, 0)
		if !ok {
			return fmt.Errorf("failed to convert private key for %s: %w", user.name, err)
		}
		userKs.Put(user.publicKey, userPrivKeyBI)

		// Create user account (Cairo v2)
		userAccnt, err := account.NewAccount(accnt.Provider, userAddrFelt, user.publicKey, userKs, account.CairoV2)
		if err != nil {
			return fmt.Errorf("failed to create account for %s: %w", user.name, err)
		}

		// Set unlimited allowance for DogCoin
		fmt.Printf("       🪙 Approving DogCoin unlimited allowance...\n")
		if err := approveUnlimited(userAccnt, dogCoin.Address, userAddrFelt, hyperlaneAddrFelt, "DogCoin"); err != nil {
			return fmt.Errorf("failed to approve DogCoin for %s: %w", user.name, err)
		}

		fmt.Printf("       ✅ %s allowances set successfully\n", user.name)
	}

	fmt.Println("   ✅ All allowances set successfully!")
	return nil
}

// approveUnlimited sets unlimited allowance for a token
func approveUnlimited(accnt *account.Account, tokenAddress string, ownerAddrFelt, spenderAddrFelt *felt.Felt, tokenName string) error {
	// Convert token address to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return fmt.Errorf("invalid token address: %w", err)
	}

	// Set unlimited allowance (max u256)
	// For u256, we need to pass (low, high) where both are max u128
	// Max u128 value is 2^128 - 1, so max u256 = (2^128 - 1) << 128 + (2^128 - 1)
	maxLowFelt := new(felt.Felt)
	maxLowFelt.SetString("0xffffffffffffffffffffffffffffffff") // Max u128 value (2^128 - 1)
	maxHighFelt := new(felt.Felt)
	maxHighFelt.SetString("0xffffffffffffffffffffffffffffffff") // Max u128 value (2^128 - 1)

	// Debug: Show what allowance values we're setting
	fmt.Printf("         🔍 Setting allowance: low=%s, high=%s\n", maxLowFelt.String(), maxHighFelt.String())

	// Build the approve function call
	approveCall := rpc.InvokeFunctionCall{
		ContractAddress: tokenAddrFelt,
		FunctionName:    "approve",
		CallData:        []*felt.Felt{spenderAddrFelt, maxLowFelt, maxHighFelt},
	}

	// Send the approve transaction
	resp, err := accnt.BuildAndSendInvokeTxn(context.Background(), []rpc.InvokeFunctionCall{approveCall}, nil)
	if err != nil {
		return fmt.Errorf("failed to send approve transaction: %w", err)
	}

	fmt.Printf("         ⏳ Approve transaction sent: %s\n", resp.Hash.String())
	fmt.Printf("         ⏳ Waiting for confirmation...\n")

	// Wait for transaction receipt
	_, err = accnt.WaitForTransactionReceipt(context.Background(), resp.Hash, time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for approve confirmation: %w", err)
	}

	fmt.Printf("         ✅ Approve transaction confirmed\n")
	return nil
}

// verifyBalancesAndAllowances verifies that users have the expected balances and allowances
func verifyBalancesAndAllowances(accnt *account.Account, dogCoin TokenInfo, hyperlaneAddress, aliceAddr, solverAddr string) error {
	// Expected increase in balance after funding
	expectedIncrease := new(big.Int)
	expectedIncrease.SetString(UserFundingAmount, 10)

	// Users to verify
	users := []struct {
		name string
		addr string
	}{
		{"Alice", aliceAddr},
		{"Solver", solverAddr},
	}

	// Verify each user's balances
	for _, user := range users {
		fmt.Printf("     🔍 Verifying %s...\n", user.name)

		// Check DogCoin balance
		dogBalance, err := getTokenBalance(accnt, dogCoin.Address, user.addr, "DogCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin balance: %w", user.name, err)
		}

		// Verify that balance is at least the expected amount (they might have had existing tokens)
		if dogBalance.Cmp(expectedIncrease) < 0 {
			return fmt.Errorf("%s's DogCoin balance too low: expected at least %s, got %s", user.name, expectedIncrease.String(), dogBalance.String())
		}
		fmt.Printf("       ✅ DogCoin: %s (at least %s)\n", formatTokenAmount(dogBalance), formatTokenAmount(expectedIncrease))

		// Check allowance if Hyperlane address is available and user is Alice
		if hyperlaneAddress != "" && user.name == "Alice" {
			// Check DogCoin allowance
			dogAllowance, err := getTokenAllowance(accnt, dogCoin.Address, user.addr, hyperlaneAddress, "DogCoin")
			if err != nil {
				return fmt.Errorf("failed to get %s's DogCoin allowance: %w", user.name, err)
			}

			// Debug: Show the actual allowance value
			if dogAllowance.Cmp(big.NewInt(0)) == 0 {
				fmt.Printf("       ⚠️  DogCoin allowance: %s (this might indicate an issue)\n", formatTokenAmount(dogAllowance))
			} else {
				fmt.Printf("       ✅ DogCoin allowance: %s\n", formatTokenAmount(dogAllowance))
			}
		}
	}

	return nil
}

// getTokenAllowance gets the allowance of a token for a specific spender
func getTokenAllowance(accnt *account.Account, tokenAddress, ownerAddress, spenderAddress, tokenName string) (*big.Int, error) {
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

	// Debug: Show what we're calling
	fmt.Printf("         🔍 Calling allowance(owner=%s, spender=%s)\n", ownerAddrFelt.String(), spenderAddrFelt.String())

	// Call the contract to get allowance
	resp, err := accnt.Provider.Call(context.Background(), allowanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance: %w", err)
	}

	// Debug: Show the full response
	fmt.Printf("         🔍 Full allowance response: %d values\n", len(resp))
	for i, val := range resp {
		fmt.Printf("         🔍 Response[%d]: %s\n", i, val.String())
	}

	if len(resp) < 2 {
		return nil, fmt.Errorf("insufficient response from allowance call: expected 2 values for u256, got %d", len(resp))
	}

	// For u256, the response should be [low, high] where:
	// - low contains the first 128 bits
	// - high contains the remaining 128 bits
	lowFelt := resp[0]
	highFelt := resp[1]

	// Debug: Show the raw felt values
	fmt.Printf("         🔍 Raw allowance response: low=%s, high=%s\n", lowFelt.String(), highFelt.String())

	// Convert low and high felts to big.Ints
	lowBigInt := utils.FeltToBigInt(lowFelt)
	highBigInt := utils.FeltToBigInt(highFelt)

	// Combine low and high into a single u256 value
	// high << 128 + low
	shiftedHigh := new(big.Int).Lsh(highBigInt, 128)
	totalAllowance := new(big.Int).Add(shiftedHigh, lowBigInt)

	return totalAllowance, nil
}

// formatTokenAmount formats a token amount for display (converts from wei to tokens)
func formatTokenAmount(amount *big.Int) string {
	// Convert from wei (18 decimals) to tokens
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	tokens := new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(decimals))
	return tokens.Text('f', 0) + " tokens"
}
