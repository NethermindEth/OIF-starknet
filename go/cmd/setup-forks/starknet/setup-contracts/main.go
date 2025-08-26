package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	central "github.com/NethermindEth/oif-starknet/go/internal/deployer"
)

// getEnvWithDefault gets an environment variable with a default fallback
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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
	DeploymentFilePath = "state/network_state/starknet-sepolia-mock-erc20-deployment.json"
)

// loadCentralAddresses loads Hyperlane, OrcaCoin, DogCoin from centralized deployment state
func loadCentralAddresses(networkName string) (hyperlane string, orca string, dog string, err error) {
	state, e := central.GetDeploymentState()
	if e != nil {
		return "", "", "", fmt.Errorf("failed to read centralized deployment-state.json: %w", e)
	}
	net, ok := state.Networks[networkName]
	if !ok {
		return "", "", "", fmt.Errorf("network %s not found in centralized state", networkName)
	}
	return net.HyperlaneAddress, net.OrcaCoinAddress, net.DogCoinAddress, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	// Check if we just want to test allowance reading
	if len(os.Args) > 1 && os.Args[1] == "test-allowance" {
		testAllowanceReading()
		return
	}

	fmt.Println("🚀 Setting up Starknet contracts: funding users and setting allowances...")

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
	hyperlaneAddr, orcaAddr, dogAddr, err := loadCentralAddresses(networkName)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to load centralized addresses: %s", err))
	}

	// Prepare TokenInfo based on centralized state
	orcaCoin := TokenInfo{Name: "OrcaCoin", Symbol: "ORCA", Address: orcaAddr}
	dogCoin := TokenInfo{Name: "DogCoin", Symbol: "DOG", Address: dogAddr}

	fmt.Printf("📋 OrcaCoin: %s\n", orcaCoin.Address)
	fmt.Printf("📋 DogCoin: %s\n", dogCoin.Address)

	// Fund test users
	fmt.Println("\n💰 Funding test users...")
	if err := fundUsers(accnt, orcaCoin, dogCoin, aliceAddress, solverAddress); err != nil {
		panic(fmt.Sprintf("❌ Failed to fund users: %s", err))
	}

	// Set allowances for Hyperlane7683
	fmt.Println("\n🔐 Setting allowances for Hyperlane7683...")
	fmt.Printf("   📋 Found Hyperlane7683 at: %s\n", hyperlaneAddr)
	fmt.Println("   🔐 Setting allowances for Hyperlane7683...")
	if err := setAllowances(accnt, orcaCoin, dogCoin, hyperlaneAddr, aliceAddress); err != nil {
		panic(fmt.Sprintf("❌ Failed to set allowances: %s", err))
	}

	// Verify balances and allowances after everything is set
	fmt.Printf("\n🔍 Verifying balances and allowances...\n")
	if err := verifyBalancesAndAllowances(accnt, orcaCoin, dogCoin, hyperlaneAddr, aliceAddress, solverAddress); err != nil {
		fmt.Printf("❌ Verification failed: %v\n", err)
	} else {
		fmt.Printf("✅ All verifications passed!\n")
	}

	fmt.Printf("\n🎯 Starknet contract setup completed successfully!\n")
	fmt.Printf("   • Users funded with tokens\n")
	fmt.Printf("   • Allowances set for Hyperlane7683\n")
	fmt.Printf("   • Ready for cross-chain operations!\n")
}

// loadTokenDeploymentInfo loads token deployment information from file
func loadTokenDeploymentInfo(networkName string) ([]TokenInfo, error) {
	// Try to read from deployment file
	deploymentFile := fmt.Sprintf("mock_erc20_deployment_%s.json", networkName)

	// Check if deployment file exists
	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
		// Try the default deployment file
		if _, err := os.Stat(DeploymentFilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("no deployment file found. Please run deploy-mock-erc20 first")
		}
		deploymentFile = DeploymentFilePath
	}

	// Read and parse deployment file
	data, err := os.ReadFile(deploymentFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployment file %s: %w", deploymentFile, err)
	}

	var deployment TokenDeploymentInfo
	if err := json.Unmarshal(data, &deployment); err != nil {
		return nil, fmt.Errorf("failed to parse deployment file %s: %w", deploymentFile, err)
	}

	return deployment.Tokens, nil
}

// fundUsers funds test users with tokens using the mint function
func fundUsers(accnt *account.Account, orcaCoin, dogCoin TokenInfo, aliceAddr, solverAddr string) error {
	users := []struct {
		name    string
		address string
	}{
		{"Alice", aliceAddr},
		{"Solver", solverAddr},
	}

	// Fund each user with both tokens
	for _, user := range users {
		fmt.Printf("   💸 Funding %s...\n", user.name)

		// Check balance before minting
		orcaBalanceBefore, err := getTokenBalance(accnt, orcaCoin.Address, user.address, "OrcaCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's OrcaCoin balance before minting: %w", user.name, err)
		}

		dogBalanceBefore, err := getTokenBalance(accnt, dogCoin.Address, user.address, "DogCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin balance before minting: %w", user.name, err)
		}

		fmt.Printf("     📊 %s balances before: OrcaCoin=%s, DogCoin=%s\n", user.name, orcaBalanceBefore, dogBalanceBefore)

		// Fund with OrcaCoin
		if err := mintTokens(accnt, orcaCoin.Address, user.address, UserFundingAmount, "OrcaCoin"); err != nil {
			return fmt.Errorf("failed to fund %s with OrcaCoin: %w", user.name, err)
		}

		// Fund with DogCoin
		if err := mintTokens(accnt, dogCoin.Address, user.address, UserFundingAmount, "DogCoin"); err != nil {
			return fmt.Errorf("failed to fund %s with DogCoin: %w", user.name, err)
		}

		// Check balance after minting
		orcaBalanceAfter, err := getTokenBalance(accnt, orcaCoin.Address, user.address, "OrcaCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's OrcaCoin balance after minting: %w", user.name, err)
		}

		dogBalanceAfter, err := getTokenBalance(accnt, dogCoin.Address, user.address, "DogCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin balance after minting: %w", user.name, err)
		}

		fmt.Printf("     📊 %s balances after: OrcaCoin=%s, DogCoin=%s\n", user.name, orcaBalanceAfter, dogBalanceAfter)

		// Verify the minting actually worked
		expectedAmount := new(big.Int)
		expectedAmount.SetString(UserFundingAmount, 10)

		orcaIncrease := new(big.Int).Sub(orcaBalanceAfter, orcaBalanceBefore)
		dogIncrease := new(big.Int).Sub(dogBalanceAfter, dogBalanceBefore)

		if orcaIncrease.Cmp(expectedAmount) != 0 {
			return fmt.Errorf("OrcaCoin minting failed for %s: expected increase %s, got %s", user.name, expectedAmount.String(), orcaIncrease.String())
		}

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

// getHyperlaneAddress gets the Hyperlane contract address for the network
func getHyperlaneAddress(networkName string) (string, error) {
	// Try to read from deployment file
	deploymentFile := fmt.Sprintf("state/network_state/starknet-sepolia-deployment.json")

	// Check if deployment file exists
	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
		return "", fmt.Errorf("Hyperlane7683 deployment file not found: %s", deploymentFile)
	}

	// Read and parse deployment file
	data, err := os.ReadFile(deploymentFile)
	if err != nil {
		return "", fmt.Errorf("failed to read deployment file %s: %w", deploymentFile, err)
	}

	var deployment struct {
		DeployedAddress string `json:"deployedAddress"`
	}
	if err := json.Unmarshal(data, &deployment); err != nil {
		return "", fmt.Errorf("failed to parse deployment file %s: %w", deploymentFile, err)
	}

	if deployment.DeployedAddress == "" {
		return "", fmt.Errorf("no deployed address found in deployment file")
	}

	fmt.Printf("   📋 Found Hyperlane7683 at: %s\n", deployment.DeployedAddress)
	return deployment.DeployedAddress, nil
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

// setAllowances sets unlimited allowances for users on tokens
func setAllowances(accnt *account.Account, orcaCoin, dogCoin TokenInfo, hyperlaneAddress, aliceAddr string) error {
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

	// Set unlimited allowance for each user on both tokens
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

		// Set unlimited allowance for OrcaCoin
		fmt.Printf("       🪙 Approving OrcaCoin unlimited allowance...\n")
		if err := approveUnlimited(userAccnt, orcaCoin.Address, userAddrFelt, hyperlaneAddrFelt, "OrcaCoin"); err != nil {
			return fmt.Errorf("failed to approve OrcaCoin for %s: %w", user.name, err)
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
func verifyBalancesAndAllowances(accnt *account.Account, orcaCoin, dogCoin TokenInfo, hyperlaneAddress, aliceAddr, solverAddr string) error {
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

		// Check OrcaCoin balance
		orcaBalance, err := getTokenBalance(accnt, orcaCoin.Address, user.addr, "OrcaCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's OrcaCoin balance: %w", user.name, err)
		}

		// Check DogCoin balance
		dogBalance, err := getTokenBalance(accnt, dogCoin.Address, user.addr, "DogCoin")
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin balance: %w", user.name, err)
		}

		// Verify that balances are at least the expected amount (they might have had existing tokens)
		if orcaBalance.Cmp(expectedIncrease) < 0 {
			return fmt.Errorf("%s's OrcaCoin balance too low: expected at least %s, got %s", user.name, expectedIncrease.String(), orcaBalance.String())
		}
		fmt.Printf("       ✅ OrcaCoin: %s (at least %s)\n", formatTokenAmount(orcaBalance), formatTokenAmount(expectedIncrease))

		if dogBalance.Cmp(expectedIncrease) < 0 {
			return fmt.Errorf("%s's DogCoin balance too low: expected at least %s, got %s", user.name, expectedIncrease.String(), dogBalance.String())
		}
		fmt.Printf("       ✅ DogCoin: %s (at least %s)\n", formatTokenAmount(dogBalance), formatTokenAmount(expectedIncrease))

		// Check allowances if Hyperlane address is available and user is Alice
		if hyperlaneAddress != "" && user.name == "Alice" {
			// Check OrcaCoin allowance
			orcaAllowance, err := getTokenAllowance(accnt, orcaCoin.Address, user.addr, hyperlaneAddress, "OrcaCoin")
			if err != nil {
				return fmt.Errorf("failed to get %s's OrcaCoin allowance: %w", user.name, err)
			}

			// Debug: Show the actual allowance value
			if orcaAllowance.Cmp(big.NewInt(0)) == 0 {
				fmt.Printf("       ⚠️  OrcaCoin allowance: %s (this might indicate an issue)\n", formatTokenAmount(orcaAllowance))
			} else {
				fmt.Printf("       ✅ OrcaCoin allowance: %s\n", formatTokenAmount(orcaAllowance))
			}

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

// testAllowanceReading is a simple test function to debug allowance reading
func testAllowanceReading() {
	fmt.Println("🧪 Testing allowance reading...")

	// Load token addresses from deployment file
	tokens, err := loadTokenDeploymentInfo("Starknet Sepolia")
	if err != nil {
		fmt.Printf("❌ Failed to load token addresses: %v\n", err)
		os.Exit(1)
	}

	// Load Hyperlane address from deployment file
	data, err := os.ReadFile("state/network_state/starknet-sepolia-deployment.json")
	if err != nil {
		fmt.Printf("❌ Failed to load Hyperlane address: %v\n", err)
		os.Exit(1)
	}

	var deployment struct {
		DeployedAddress string `json:"deployedAddress"`
	}
	if err := json.Unmarshal(data, &deployment); err != nil {
		fmt.Printf("❌ Failed to parse Hyperlane deployment: %v\n", err)
		os.Exit(1)
	}

	hyperlaneAddr := deployment.DeployedAddress

	fmt.Printf("📋 Hyperlane address: %s\n", hyperlaneAddr)
	fmt.Printf("📋 OrcaCoin: %s\n", tokens[0].Address)
	fmt.Printf("📋 DogCoin: %s\n", tokens[1].Address)

	// Test user (Alice)
	// Alice's address (Katana default account)
	aliceAddr := getEnvWithDefault("STARKNET_ALICE_ADDRESS", "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7")

	// Connect to Starknet RPC
	client, err := rpc.NewProvider(os.Getenv("STARKNET_RPC_URL"))
	if err != nil {
		fmt.Printf("❌ Failed to connect to Starknet: %v\n", err)
		os.Exit(1)
	}

	// Test allowance reading for OrcaCoin
	fmt.Printf("\n🔍 Testing OrcaCoin allowance for Alice -> Hyperlane7683...\n")

	// Convert addresses to felt
	tokenAddrFelt, _ := utils.HexToFelt(tokens[0].Address)
	ownerAddrFelt, _ := utils.HexToFelt(aliceAddr)
	spenderAddrFelt, _ := utils.HexToFelt(hyperlaneAddr)

	// Build the allowance function call
	allowanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{ownerAddrFelt, spenderAddrFelt},
	}

	fmt.Printf("   🔍 Calling allowance(owner=%s, spender=%s)\n", ownerAddrFelt.String(), spenderAddrFelt.String())

	// Call the contract to get allowance
	resp, err := client.Call(context.Background(), allowanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		fmt.Printf("   ❌ Failed to call allowance: %v\n", err)
		return
	}

	// Show the full response
	fmt.Printf("   🔍 Full allowance response: %d values\n", len(resp))
	for i, val := range resp {
		fmt.Printf("   🔍 Response[%d]: %s\n", i, val.String())
	}

	if len(resp) >= 2 {
		// Convert low and high felts to big.Ints
		lowBigInt := utils.FeltToBigInt(resp[0])
		highBigInt := utils.FeltToBigInt(resp[1])

		// Combine low and high into a single u256 value
		shiftedHigh := new(big.Int).Lsh(highBigInt, 128)
		totalAllowance := new(big.Int).Add(shiftedHigh, lowBigInt)

		fmt.Printf("   ✅ Parsed allowance: %s\n", totalAllowance.String())
	} else {
		fmt.Printf("   ⚠️  Unexpected response format\n")
	}

	// Now let's try to set a small allowance and see if it works
	fmt.Printf("\n🧪 Testing setting a small allowance...\n")

	// Try to set allowance to 1000 tokens
	smallAmount := new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	lowFelt := utils.BigIntToFelt(smallAmount)
	highFelt := new(felt.Felt) // 0 for small amounts

	fmt.Printf("   🔍 Setting allowance to: low=%s, high=%s\n", lowFelt.String(), highFelt.String())

	// Build the approve function call
	fmt.Printf("   🔍 Calling approve(spender=%s, amount_low=%s, amount_high=%s)\n",
		spenderAddrFelt.String(), lowFelt.String(), highFelt.String())

	// Note: We can't actually send the transaction here without an account, but we can show the call data
	fmt.Printf("   📝 Approve call data prepared (would need account to send)\n")
}
