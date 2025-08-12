package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Test user addresses (Alice, Bob, Charlie)
var testUsers = []string{
	"0x70997970C51812dc3A010C7d01b50e0d17dc79C8", // Alice (Account 1)
	"0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC", // Bob (Account 2)
	"0x90F79bf6EB2c4f870365E785982E1f101E93b906", // Charlie (Account 3)
}

// Network configuration
var networks = []struct {
	name string
	url  string
}{
	{"Sepolia", "http://localhost:8545"},
	{"Optimism Sepolia", "http://localhost:8546"},
	{"Arbitrum Sepolia", "http://localhost:8547"},
	{"Base Sepolia", "http://localhost:8548"},
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	deployerKeyHex := os.Getenv("DEPLOYER_PRIVATE_KEY")
	if deployerKeyHex == "" {
		log.Fatal("DEPLOYER_PRIVATE_KEY environment variable is required")
	}

	aliceKeyHex := os.Getenv("ALICE_PRIVATE_KEY")
	bobKeyHex := os.Getenv("BOB_PRIVATE_KEY")
	charlieKeyHex := os.Getenv("CHARLIE_PRIVATE_KEY")
	
	if aliceKeyHex == "" || bobKeyHex == "" || charlieKeyHex == "" {
		log.Fatal("ALICE_PRIVATE_KEY, BOB_PRIVATE_KEY, and CHARLIE_PRIVATE_KEY environment variables are required")
	}

	// Parse deployer private key
	deployerKeyHex = strings.TrimPrefix(deployerKeyHex, "0x")
	deployerKey, err := crypto.HexToECDSA(deployerKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse deployer private key: %v", err)
	}

	// Parse test user private keys
	aliceKeyHex = strings.TrimPrefix(aliceKeyHex, "0x")
	aliceKey, err := crypto.HexToECDSA(aliceKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Alice private key: %v", err)
	}

	bobKeyHex = strings.TrimPrefix(bobKeyHex, "0x")
	bobKey, err := crypto.HexToECDSA(bobKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Bob private key: %v", err)
	}

	charlieKeyHex = strings.TrimPrefix(charlieKeyHex, "0x")
	charlieKey, err := crypto.HexToECDSA(charlieKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Charlie private key: %v", err)
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

		// Fund test users
		if err := fundUsers(client, deployerKey, aliceKey, bobKey, charlieKey, orcaCoinAddress, dogCoinAddress, network.name); err != nil {
			fmt.Printf("   ❌ Failed to fund users: %v\n", err)
			continue
		}

		// Set allowances for Hyperlane7683
		if err := setAllowances(client, aliceKey, bobKey, charlieKey, orcaCoinAddress, dogCoinAddress, network.name); err != nil {
			fmt.Printf("   ❌ Failed to set allowances: %v\n", err)
			continue
		}

		client.Close()
		fmt.Printf("   🎉 %s setup complete!\n", network.name)
	}

	fmt.Printf("\n🎯 All networks configured!\n")
	fmt.Printf("   • OrcaCoin and DogCoin deployed to all networks\n")
	fmt.Printf("   • Test users funded with tokens\n")
	fmt.Printf("   • Allowances set for Hyperlane7683\n")
	fmt.Printf("   • Ready to open orders!\n")
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
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to create auth: %w", err)
	}
	
	// Get current gas price from network
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get gas price: %w", err)
	}
	
	// Set gas price and limit
	auth.GasPrice = gasPrice
	auth.GasLimit = uint64(5000000) // 5M gas
	
	// Deploy the contract
	address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(contract.Bytecode), client)
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
		return common.Address{}, fmt.Errorf("deployment transaction failed")
	}
	
	fmt.Printf("   ✅ %s deployed successfully at: %s\n", symbol, address.Hex())
	return address, nil
}

func fundUsers(client *ethclient.Client, deployerKey, aliceKey, bobKey, charlieKey *ecdsa.PrivateKey, orcaCoinAddress, dogCoinAddress common.Address, networkName string) error {
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
	deployerAuth, err := bind.NewKeyedTransactorWithChainID(deployerKey, chainID)
	if err != nil {
		return fmt.Errorf("failed to create deployer auth: %w", err)
	}
	
	// Deployer already has initial supply (420,690,000,000,000 * 10^decimals)
	// Amount to distribute per user (100,000 tokens with 18 decimals)
	userAmount := new(big.Int).Mul(big.NewInt(100000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	
	fmt.Printf("     💰 Deployer has initial supply, distributing to users...\n")
	
	// Distribute tokens to test users
	users := []struct {
		name string
		key  *ecdsa.PrivateKey
	}{
		{"Alice", aliceKey},
		{"Bob", bobKey},
		{"Charlie", charlieKey},
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
	
	userAuth, err := bind.NewKeyedTransactorWithChainID(userKey, chainID)
	if err != nil {
		return fmt.Errorf("failed to create user auth: %w", err)
	}
	
	// Get current nonce for deployer
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}
	
	// Get current gas price from network
	gasPrice, err := client.SuggestGasPrice(context.Background())
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

func setAllowances(client *ethclient.Client, aliceKey, bobKey, charlieKey *ecdsa.PrivateKey, orcaCoinAddress, dogCoinAddress common.Address, networkName string) error {
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
	
	// Hyperlane7683 contract address (pre-deployed on testnets)
	hyperlaneAddress := common.HexToAddress("0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")
	
	// Users to set allowances for
	users := []struct {
		name string
		key  *ecdsa.PrivateKey
	}{
		{"Alice", aliceKey},
		{"Bob", bobKey},
		{"Charlie", charlieKey},
	}
	
	// Set unlimited allowance for each user
	for _, user := range users {
		fmt.Printf("     🔓 Setting %s allowances...\n", user.name)
		
		// Create user auth
		userAuth, err := bind.NewKeyedTransactorWithChainID(user.key, chainID)
		if err != nil {
			return fmt.Errorf("failed to create auth for %s: %w", user.name, err)
		}
		
		// Get current gas price
		gasPrice, err := client.SuggestGasPrice(context.Background())
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
		
		// Set unlimited allowance for DogCoin
		fmt.Printf("       🪙 Approving DogCoin unlimited allowance...\n")
		if err := approveUnlimited(client, userAuth, dogCoinAddress, hyperlaneAddress, parsedABI, nonce+1, gasPrice); err != nil {
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
		100000,
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
		return fmt.Errorf("approve transaction failed")
	}
	
	return nil
}
