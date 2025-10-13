package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/NethermindEth/oif-starknet/solver/pkg/ethutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// Default funding amount (420,690,000,000 tokens)
	defaultFundingAmount = 420_690_000_000
	// Token decimals (18 for most ERC20 tokens)
	tokenDecimals = 18
	// Gas limit for transactions
	defaultGasLimit = 300000
	// Base 10 for string parsing
	base10 = 10
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("🏦 MockERC20 Token Funding Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  fund-accounts <network|all> [amount]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fund-accounts all           # Fund Alice & Solver on all networks with 10000 tokens")
		fmt.Println("  fund-accounts ethereum      # Fund Alice & Solver on Ethereum with 10000 tokens")
		fmt.Println("  fund-accounts starknet      # Fund Alice & Solver on Starknet with 10000 tokens")
		fmt.Println("  fund-accounts all 50000     # Fund Alice & Solver on all networks with 50000 tokens")
		fmt.Println()
		fmt.Println("Networks: ethereum, optimism, arbitrum, base, starknet, all")
		os.Exit(1)
	}

	networkArg := strings.ToLower(os.Args[1])

	// Default funding amount (420,690,000,000 tokens with 18 decimals)
	fundingAmount := createTokenAmount(defaultFundingAmount, tokenDecimals)

	// Parse custom amount if provided
	if len(os.Args) >= 3 {
		if customAmount, ok := new(big.Int).SetString(os.Args[2], base10); ok {
			fundingAmount = createTokenAmount(customAmount.Int64(), tokenDecimals)
		} else {
			log.Fatalf("Invalid amount: %s", os.Args[2])
		}
	}

	// Load configuration
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("🏦 Funding Alice and Solver accounts with %s tokens each\n", ethutil.FormatTokenAmount(fundingAmount, tokenDecimals))
	fmt.Printf("💰 Using conditional environment variables (IS_DEVNET=%s)\n", os.Getenv("IS_DEVNET"))
	fmt.Println()

	if networkArg == "all" {
		fundAllNetworks(fundingAmount)
		fmt.Println()
		fundStarknet(fundingAmount)
	} else if networkArg == "starknet" {
		fundStarknet(fundingAmount)
	} else {
		fundNetwork(networkArg, fundingAmount)
	}

	fmt.Println("🎉 Funding completed!")
}

func fundAllNetworks(amount *big.Int) {
	networks := []string{"ethereum", "optimism", "arbitrum", "base"}

	for _, network := range networks {
		fmt.Printf("📡 Funding %s network...\n", strings.ToTitle(network))
		fundNetwork(network, amount)
		fmt.Println()
	}
}

func fundNetwork(networkName string, amount *big.Int) {
	// Load network configuration
	config.InitializeNetworks()

	var networkConfig *config.NetworkConfig
	for name, cfg := range config.Networks {
		if strings.EqualFold(name, networkName) {
			networkConfig = &cfg
			break
		}
	}

	if networkConfig == nil {
		log.Fatalf("Network not found: %s", networkName)
	}

	// Get MockERC20 address from environment
	envVarName := strings.ToUpper(networkName) + "_DOG_COIN_ADDRESS"
	tokenAddress := os.Getenv(envVarName)
	if tokenAddress == "" {
		log.Fatalf("MockERC20 address not found in environment: %s", envVarName)
	}

	// Connect to network
	client, err := ethclient.Dial(networkConfig.RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", networkName, err)
	}
	defer client.Close()

	fmt.Printf("   📍 Network: %s (Chain ID: %d)\n", networkConfig.Name, networkConfig.ChainID)
	fmt.Printf("   🪙 MockERC20: %s\n", tokenAddress)

	// Get funding account (the one that can mint tokens)
	// Use Alice's account as the minter for simplicity
	var minterPrivateKey string
	isDevnet := os.Getenv("IS_DEVNET") == "true"
	if isDevnet {
		minterPrivateKey = os.Getenv("LOCAL_ALICE_PRIVATE_KEY")
	} else {
		minterPrivateKey = os.Getenv("ALICE_PRIVATE_KEY")
	}

	if minterPrivateKey == "" {
		log.Fatalf("Minter private key not found (Alice's key)")
	}

	// Parse private key and create auth
	privateKey, err := ethutil.ParsePrivateKey(minterPrivateKey)
	if err != nil {
		log.Fatalf("Failed to parse minter private key: %v", err)
	}

	auth, err := ethutil.NewTransactor(big.NewInt(int64(networkConfig.ChainID)), privateKey)
	if err != nil {
		log.Fatalf("Failed to create transactor: %v", err)
	}

	// Set gas price
	gasPrice, err := ethutil.SuggestGas(client)
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}
	auth.GasPrice = gasPrice

	// Get recipient addresses
	recipients := getRecipients(isDevnet)

	// Fund each recipient using direct contract call (since Go bindings don't have Mint yet)
	for _, recipient := range recipients {
		fmt.Printf("   💸 Funding %s (%s)...\n", recipient.Name, recipient.Address.Hex())

		// Check current balance
		currentBalance, err := ethutil.ERC20Balance(client, common.HexToAddress(tokenAddress), recipient.Address)
		if err == nil {
			fmt.Printf("     📊 Current balance: %s\n", ethutil.FormatTokenAmount(currentBalance, tokenDecimals))
		}

		// Call mint function directly using raw transaction
		err = mintTokensRaw(client, auth, tokenAddress, recipient.Address, amount)
		if err != nil {
			log.Printf("     ❌ Failed to mint tokens for %s: %v", recipient.Name, err)
			continue
		}

		fmt.Printf("     ✅ Minted %s tokens\n", ethutil.FormatTokenAmount(amount, tokenDecimals))

		// Verify new balance
		newBalance, err := ethutil.ERC20Balance(client, common.HexToAddress(tokenAddress), recipient.Address)
		if err == nil {
			fmt.Printf("     💰 New balance: %s\n", ethutil.FormatTokenAmount(newBalance, tokenDecimals))
		}
	}
}

type Recipient struct {
	Name    string
	Address common.Address
}

func getRecipients(isDevnet bool) []Recipient {
	var recipients []Recipient

	// Alice
	var aliceAddr string
	if isDevnet {
		aliceAddr = envutil.GetEnvWithDefault("LOCAL_ALICE_PUB_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	} else {
		aliceAddr = envutil.GetEnvWithDefault("ALICE_PUB_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	}
	recipients = append(recipients, Recipient{
		Name:    "Alice",
		Address: common.HexToAddress(aliceAddr),
	})

	// Solver
	var solverAddr string
	if isDevnet {
		solverAddr = envutil.GetEnvWithDefault("LOCAL_SOLVER_PUB_KEY", "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC")
	} else {
		solverAddr = envutil.GetEnvWithDefault("SOLVER_PUB_KEY", "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC")
	}
	recipients = append(recipients, Recipient{
		Name:    "Solver",
		Address: common.HexToAddress(solverAddr),
	})

	return recipients
}

func createTokenAmount(tokens int64, decimals int) *big.Int {
	amount := big.NewInt(tokens)
	multiplier := big.NewInt(base10)
	multiplier.Exp(multiplier, big.NewInt(int64(decimals)), nil)
	return amount.Mul(amount, multiplier)
}

// mintTokensRaw calls the mint function directly using raw transaction
func mintTokensRaw(client *ethclient.Client, auth *bind.TransactOpts, tokenAddress string, recipient common.Address, amount *big.Int) error {
	// mint(address to, uint256 amount) function signature
	mintABI := `[{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(mintABI))
	if err != nil {
		return fmt.Errorf("failed to parse mint ABI: %w", err)
	}

	// Pack the function call data
	data, err := parsedABI.Pack("mint", recipient, amount)
	if err != nil {
		return fmt.Errorf("failed to pack mint call: %w", err)
	}

	// Get current nonce
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(tokenAddress),
		big.NewInt(0),   // No ETH value
		defaultGasLimit, // Gas limit for mint
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return fmt.Errorf("failed to sign mint transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send mint transaction: %w", err)
	}

	fmt.Printf("     🚀 Mint transaction: %s\n", signedTx.Hash().Hex())

	// Wait for confirmation
	receipt, err := bind.WaitMined(context.Background(), client, signedTx)
	if err != nil {
		return fmt.Errorf("failed to wait for mint confirmation: %w", err)
	}

	if receipt.Status != 1 {
		return fmt.Errorf("mint transaction failed")
	}

	fmt.Printf("     ⛽ Gas used: %d\n", receipt.GasUsed)
	return nil
}
