package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/pkg/ethutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

/// Deploys mock erc20 tokens, funds accounts, and sets allowances

// Network configuration - built from centralized config
var networks = func() []struct {
	name string
	url  string
} {
	networkNames := config.GetNetworkNames()
	networks := make([]struct {
		name string
		url  string
	}, 0, len(networkNames))

	for _, networkName := range networkNames {
		// Skip non-EVM networks (like Starknet)
		if networkName == "Starknet Sepolia" {
			continue
		}

		networkConfig := config.Networks[networkName]
		networks = append(networks, struct {
			name string
			url  string
		}{
			name: networkConfig.Name,
			url:  networkConfig.RPCURL,
		})
	}
	return networks
}()

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	deployerKeyHex := os.Getenv("DEPLOYER_PRIVATE_KEY")
	if deployerKeyHex == "" {
		log.Fatal("DEPLOYER_PRIVATE_KEY environment variable is required")
	}

	aliceKeyHex := os.Getenv("ALICE_PRIVATE_KEY")
	solverKeyHex := os.Getenv("SOLVER_PRIVATE_KEY")

	if aliceKeyHex == "" || solverKeyHex == "" {
		log.Fatal("ALICE_PRIVATE_KEY and SOLVER_PRIVATE_KEY environment variables are required")
	}

	// Parse deployer private key
	deployerKey, err := ethutil.ParsePrivateKey(deployerKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse deployer private key: %v", err)
	}

	// Parse private keys for Alice (opens orders) and Solver (fills orders)
	aliceKey, err := ethutil.ParsePrivateKey(aliceKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Alice private key: %v", err)
	}

	solverKey, err := ethutil.ParsePrivateKey(solverKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Solver private key: %v", err)
	}

	// Deploy tokens to all networks
	for _, network := range networks {
		fmt.Printf("\n🚀 Deploying tokens to %s...\n", network.name)
		fmt.Printf("   URL: %s\n", network.url)

		client, err := ethclient.Dial(network.url)
		if err != nil {
			fmt.Printf("   ❌ Failed to connect: %v\n", err)
			continue
		}

		// Deploy OrcaCoin (origin chain token)
		orcaCoinAddress, err := deployERC20(client, deployerKey, "OrcaCoin", network.name)
		if err != nil {
			fmt.Printf("   ❌ Failed to deploy OrcaCoin: %v\n", err)
			continue
		}
		fmt.Printf("   ✅ OrcaCoin deployed at: %s\n", orcaCoinAddress)

		// Deploy DogCoin (destination chain token)
		dogCoinAddress, err := deployERC20(client, deployerKey, "DogCoin", network.name)
		if err != nil {
			fmt.Printf("   ❌ Failed to deploy DogCoin: %v\n", err)
			continue
		}
		fmt.Printf("   ✅ DogCoin deployed at: %s\n", dogCoinAddress)

		// Save deployment state for this network
		if err := deployer.UpdateNetworkState(network.name, orcaCoinAddress.Hex(), dogCoinAddress.Hex()); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to save deployment state: %v\n", err)
		} else {
			fmt.Printf("   💾 Deployment state saved for %s\n", network.name)
		}

		// Fund both Alice (order creator) and Solver (order filler)
		if err := fundUsers(client, deployerKey, aliceKey, solverKey, orcaCoinAddress, dogCoinAddress, network.name); err != nil {
			fmt.Printf("   ❌ Failed to fund users: %v\n", err)
			continue
		}

		// Set allowances for both Alice and Solver
		if err := setAllowances(client, aliceKey, orcaCoinAddress, dogCoinAddress, network.name); err != nil {
			fmt.Printf("   ❌ Failed to set allowances: %v\n", err)
			continue
		}

		// Verify balances and allowances for both users
		fmt.Printf("   🔍 Verifying balances and allowances...\n")
		if err := verifyBalancesAndAllowances(client, aliceKey, solverKey, orcaCoinAddress, dogCoinAddress, network.name); err != nil {
			fmt.Printf("   ❌ Verification failed: %v\n", err)
			continue
		}
		fmt.Printf("   ✅ All verifications passed!\n")

		client.Close()
		fmt.Printf("   🎉 %s setup complete!\n", network.name)
	}

	fmt.Printf("\n🎯 All networks configured!\n")
	fmt.Printf("   • OrcaCoin and DogCoin deployed to all networks\n")
	fmt.Printf("   • Alice funded with tokens (to open orders)\n")
	fmt.Printf("   • Solver funded with tokens (to fill orders)\n")
	fmt.Printf("   • Allowances set for Hyperlane7683\n")
	fmt.Printf("   • Ready for Alice to open orders and Solver to fill them!\n")
}

func deployERC20(client *ethclient.Client, privateKey *ecdsa.PrivateKey, symbol, networkName string) (common.Address, error) {
	fmt.Printf("   📝 Deploying %s...\n", symbol)

	// Get the ERC20 contract configuration
	contract := deployer.GetERC20Contract()

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(contract.ABI))
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Get chain ID for transaction signing
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Create auth for transaction signing
	auth, err := ethutil.NewTransactor(chainID, privateKey)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to create auth: %w", err)
	}

	// Get current gas price from network
	gasPrice, err := ethutil.SuggestGas(client)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Set gas price and limit
	auth.GasPrice = gasPrice
	auth.GasLimit = uint64(5000000) // 5M gas

	// Deploy the contract with constructor parameters: name, symbol, decimals, initialSupply
	// For OrcaCoin: "OrcaCoin", "ORCA", 18, 420690000000000 * 10^18
	// For DogCoin: "DogCoin", "DOG", 18, 420690000000000 * 10^18
	var tokenName, tokenSymbol string
	var decimals uint8
	var initialSupply *big.Int

	if symbol == "OrcaCoin" {
		tokenName = "OrcaCoin"
		tokenSymbol = "ORCA"
		decimals = 18
		initialSupply = new(big.Int).Mul(big.NewInt(420690000000000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	} else {
		tokenName = "DogCoin"
		tokenSymbol = "DOG"
		decimals = 18
		initialSupply = new(big.Int).Mul(big.NewInt(420690000000000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	}

	address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(contract.Bytecode), client, tokenName, tokenSymbol, decimals, initialSupply)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to deploy contract: %w", err)
	}

	fmt.Printf("   📡 Deployment transaction: %s\n", tx.Hash().Hex())
	fmt.Printf("   ⏳ Waiting for confirmation...\n")

	// Wait for transaction confirmation
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to wait for confirmation: %w", err)
	}

	if receipt.Status == 0 {
		// Get more details about the failed transaction
		tx, _, err := client.TransactionByHash(context.Background(), tx.Hash())
		if err != nil {
			return common.Address{}, fmt.Errorf("deployment transaction failed, and failed to get transaction details: %w", err)
		}

		// Try to get the reason for failure
		msg := ethereum.CallMsg{
			From:  auth.From,
			To:    nil, // Contract creation
			Value: big.NewInt(0),
			Data:  tx.Data(),
		}

		_, err = client.CallContract(context.Background(), msg, receipt.BlockNumber)
		if err != nil {
			return common.Address{}, fmt.Errorf("deployment transaction failed: %w", err)
		}

		return common.Address{}, fmt.Errorf("deployment transaction failed with status 0")
	}

	fmt.Printf("   ✅ %s deployed successfully at: %s\n", symbol, address.Hex())
	return address, nil
}

func fundUsers(client *ethclient.Client, deployerKey, aliceKey, solverKey *ecdsa.PrivateKey, orcaCoinAddress, dogCoinAddress common.Address, networkName string) error {
	fmt.Printf("   💰 Funding test users...\n")

	// Get the ERC20 contract configuration
	contract := deployer.GetERC20Contract()

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(contract.ABI))
	if err != nil {
		return fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Create deployer auth for minting
	deployerAuth, err := ethutil.NewTransactor(chainID, deployerKey)
	if err != nil {
		return fmt.Errorf("failed to create deployer auth: %w", err)
	}

	// Deployer already has initial supply (420,690,000,000,000 * 10^decimals)
	// Amount to distribute per user (100,000 tokens with 18 decimals)
	userAmount := new(big.Int).Mul(big.NewInt(100000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	fmt.Printf("     💰 Deployer has initial supply, distributing to users...\n")

	// Distribute tokens to Alice (opens orders) and Solver (fills orders)
	users := []struct {
		name string
		key  *ecdsa.PrivateKey
	}{
		{"Alice", aliceKey},
		{"Solver", solverKey},
	}

	for _, user := range users {
		fmt.Printf("     💸 Funding %s with OrcaCoins...\n", user.name)
		if err := transferTokens(client, deployerAuth, orcaCoinAddress, parsedABI, user.key, userAmount); err != nil {
			return fmt.Errorf("failed to fund %s with OrcaCoins: %w", user.name, err)
		}

		fmt.Printf("     💸 Funding %s with DogCoins...\n", user.name)
		if err := transferTokens(client, deployerAuth, dogCoinAddress, parsedABI, user.key, userAmount); err != nil {
			return fmt.Errorf("failed to fund %s with DogCoins: %w", user.name, err)
		}
	}

	fmt.Printf("   ✅ All users funded successfully!\n")
	return nil
}

// transferTokens transfers tokens from deployer to a user
func transferTokens(client *ethclient.Client, auth *bind.TransactOpts, tokenAddress common.Address, parsedABI abi.ABI, userKey *ecdsa.PrivateKey, amount *big.Int) error {
	// Get user address
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	userAuth, err := ethutil.NewTransactor(chainID, userKey)
	if err != nil {
		return fmt.Errorf("failed to create user auth: %w", err)
	}

	// Get current nonce for deployer
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get current gas price from network
	gasPrice, err := ethutil.SuggestGas(client)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	// Encode transfer function call
	data, err := parsedABI.Pack("transfer", userAuth.From, amount)
	if err != nil {
		return fmt.Errorf("failed to encode transfer call: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0),
		100000,
		gasPrice,
		data,
	)

	// Sign and send transaction
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return fmt.Errorf("failed to sign transfer transaction: %w", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transfer transaction: %w", err)
	}

	// Wait for confirmation
	receipt, err := bind.WaitMined(context.Background(), client, signedTx)
	if err != nil {
		return fmt.Errorf("failed to wait for transfer confirmation: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transfer transaction failed")
	}

	return nil
}

func setAllowances(client *ethclient.Client, aliceKey *ecdsa.PrivateKey, orcaCoinAddress, dogCoinAddress common.Address, networkName string) error {
	fmt.Printf("   🔐 Setting allowances for Hyperlane7683...\n")

	// Get the ERC20 contract configuration
	contract := deployer.GetERC20Contract()

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(contract.ABI))
	if err != nil {
		return fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Get Hyperlane address from centralized config for the current network
	hyperlaneAddress, err := config.GetHyperlaneAddress(networkName)
	if err != nil {
		return fmt.Errorf("failed to get Hyperlane address: %w", err)
	}

	// Users to set allowances for (Alice opens orders, Solver fills orders)
	users := []struct {
		name string
		key  *ecdsa.PrivateKey
	}{
		{"Alice", aliceKey},
	}

	// Set unlimited allowance for each user
	for _, user := range users {
		fmt.Printf("     🔓 Setting %s allowances...\n", user.name)

		// Create user auth
		userAuth, err := ethutil.NewTransactor(chainID, user.key)
		if err != nil {
			return fmt.Errorf("failed to create auth for %s: %w", user.name, err)
		}

		// Get current gas price
		gasPrice, err := ethutil.SuggestGas(client)
		if err != nil {
			return fmt.Errorf("failed to get gas price for %s: %w", user.name, err)
		}

		// Get current nonce
		nonce, err := client.PendingNonceAt(context.Background(), userAuth.From)
		if err != nil {
			return fmt.Errorf("failed to get nonce for %s: %w", user.name, err)
		}

		// Set unlimited allowance for OrcaCoin
		fmt.Printf("       🪙 Approving OrcaCoin unlimited allowance...\n")
		if err := approveUnlimited(client, userAuth, orcaCoinAddress, hyperlaneAddress, parsedABI, nonce, gasPrice); err != nil {
			return fmt.Errorf("failed to approve OrcaCoin for %s: %w", user.name, err)
		}

		// Get fresh nonce for second transaction
		nonce, err = client.PendingNonceAt(context.Background(), userAuth.From)
		if err != nil {
			return fmt.Errorf("failed to get fresh nonce for %s: %w", user.name, err)
		}

		// Set unlimited allowance for DogCoin
		fmt.Printf("       🪙 Approving DogCoin unlimited allowance...\n")
		if err := approveUnlimited(client, userAuth, dogCoinAddress, hyperlaneAddress, parsedABI, nonce, gasPrice); err != nil {
			return fmt.Errorf("failed to approve DogCoin for %s: %w", user.name, err)
		}

		fmt.Printf("       ✅ %s allowances set successfully\n", user.name)
	}

	fmt.Printf("   ✅ All allowances set successfully!\n")
	return nil
}

// approveUnlimited sets unlimited allowance for a token
func approveUnlimited(client *ethclient.Client, auth *bind.TransactOpts, tokenAddress, spenderAddress common.Address, parsedABI abi.ABI, nonce uint64, gasPrice *big.Int) error {
	// Encode approve function call with max uint256 allowance
	maxAllowance := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)) // 2^256 - 1

	data, err := parsedABI.Pack("approve", spenderAddress, maxAllowance)
	if err != nil {
		return fmt.Errorf("failed to encode approve call: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0),
		200000,
		gasPrice,
		data,
	)

	// Sign and send transaction
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return fmt.Errorf("failed to sign approve transaction: %w", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send approve transaction: %w", err)
	}

	// Wait for confirmation
	receipt, err := bind.WaitMined(context.Background(), client, signedTx)
	if err != nil {
		return fmt.Errorf("failed to wait for approve confirmation: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("approve transaction failed at block %d", receipt.BlockNumber)
	}

	return nil
}

// verifyBalancesAndAllowances verifies that users have the expected balances and allowances
func verifyBalancesAndAllowances(client *ethclient.Client, aliceKey, solverKey *ecdsa.PrivateKey, orcaCoinAddress, dogCoinAddress common.Address, networkName string) error {

	// Expected balance after funding
	expectedBalance := new(big.Int).Mul(big.NewInt(100000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)) // 100,000 tokens

	// Get Hyperlane address from centralized config for the current network
	hyperlaneAddress, err := config.GetHyperlaneAddress(networkName)
	if err != nil {
		return fmt.Errorf("failed to get Hyperlane address: %w", err)
	}

	// Users to verify (Alice opens orders, Solver fills orders)
	users := []struct {
		name string
		key  *ecdsa.PrivateKey
	}{
		{"Alice", aliceKey},
		{"Solver", solverKey},
	}

	// Verify each user's balances
	for _, user := range users {
		fmt.Printf("     🔍 Verifying %s...\n", user.name)

		// Check OrcaCoin balance
		userAddr := crypto.PubkeyToAddress(user.key.PublicKey)
		orcaBalance, err := ethutil.ERC20Balance(client, orcaCoinAddress, userAddr)
		if err != nil {
			return fmt.Errorf("failed to get %s's OrcaCoin balance: %w", user.name, err)
		}

		if orcaBalance.Cmp(expectedBalance) != 0 {
			return fmt.Errorf("%s's OrcaCoin balance mismatch: expected %s, got %s", user.name, expectedBalance.String(), orcaBalance.String())
		}
		fmt.Printf("       ✅ OrcaCoin: %s tokens\n", new(big.Float).Quo(new(big.Float).SetInt(orcaBalance), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))).Text('f', 0))

		// Check DogCoin balance
		dogBalance, err := ethutil.ERC20Balance(client, dogCoinAddress, userAddr)
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin balance: %w", user.name, err)
		}
		if dogBalance.Cmp(expectedBalance) != 0 {
			return fmt.Errorf("%s's DogCoin balance mismatch: expected %s, got %s", user.name, expectedBalance.String(), dogBalance.String())
		}
		fmt.Printf("       ✅ DogCoin: %s tokens\n", new(big.Float).Quo(new(big.Float).SetInt(dogBalance), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))).Text('f', 0))

		// Check OrcaCoin allowance
		orcaAllowance, err := ethutil.ERC20Allowance(client, orcaCoinAddress, userAddr, hyperlaneAddress)
		if err != nil {
			return fmt.Errorf("failed to get %s's OrcaCoin allowance: %w", user.name, err)
		}
		maxAllowance := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)) // 2^256 - 1
		if orcaAllowance.Cmp(maxAllowance) != 0 {
			return fmt.Errorf("%s's OrcaCoin allowance mismatch: expected unlimited, got %s", user.name, orcaAllowance.String())
		}
		fmt.Printf("       ✅ OrcaCoin allowance: Unlimited\n")

		// Check DogCoin allowance
		dogAllowance, err := ethutil.ERC20Allowance(client, dogCoinAddress, userAddr, hyperlaneAddress)
		if err != nil {
			return fmt.Errorf("failed to get %s's DogCoin allowance: %w", user.name, err)
		}
		if dogAllowance.Cmp(maxAllowance) != 0 {
			return fmt.Errorf("%s's DogCoin allowance mismatch: expected unlimited, got %s", user.name, dogAllowance.String())
		}
		fmt.Printf("       ✅ DogCoin allowance: Unlimited\n")
	}

	return nil
}
