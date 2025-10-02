package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/NethermindEth/oif-starknet/solver/pkg/starknetutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

func fundStarknet(amount *big.Int) {
	fmt.Printf("📡 Funding Starknet network...\n")

	// Load network configuration
	config.InitializeNetworks()

	starknetConfig, exists := config.Networks["Starknet"]
	if !exists {
		log.Fatalf("Starknet network not found in config")
	}

	// Get MockERC20 address from environment
	tokenAddress := os.Getenv("STARKNET_DOG_COIN_ADDRESS")
	if tokenAddress == "" {
		log.Fatalf("STARKNET_DOG_COIN_ADDRESS not found in environment")
	}

	// Connect to Starknet
	client, err := rpc.NewProvider(starknetConfig.RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to Starknet: %v", err)
	}

	fmt.Printf("   📍 Network: Starknet (Chain ID: %d)\n", starknetConfig.ChainID)
	fmt.Printf("   🪙 MockERC20: %s\n", tokenAddress)

	// Get minter account (use Alice as minter)
	minterPrivateKey := envutil.GetStarknetAlicePrivateKey()
	minterPublicKey := envutil.GetStarknetAlicePublicKey()
	minterAddress := envutil.GetStarknetAliceAddress()

	if minterPrivateKey == "" || minterPublicKey == "" {
		log.Fatalf("Starknet minter credentials not found (Alice's keys)")
	}

	// Create minter account
	minterAddrFelt, err := utils.HexToFelt(minterAddress)
	if err != nil {
		log.Fatalf("Failed to convert minter address to felt: %v", err)
	}

	minterKs := account.NewMemKeystore()
	minterPrivKeyBI, ok := new(big.Int).SetString(minterPrivateKey, 0)
	if !ok {
		log.Fatalf("Failed to parse minter private key")
	}
	minterKs.Put(minterPublicKey, minterPrivKeyBI)

	minterAccount, err := account.NewAccount(client, minterAddrFelt, minterPublicKey, minterKs, account.CairoV2)
	if err != nil {
		log.Fatalf("Failed to create minter account: %v", err)
	}

	// Get recipient addresses
	recipients := getStarknetRecipients()

	// Convert token address to felt
	tokenFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		log.Fatalf("Failed to convert token address to felt: %v", err)
	}

	// Convert amount to two felts (low, high) for u256
	amountLow, amountHigh := starknetutil.ConvertBigIntToU256Felts(amount)

	// Build all mint calls for multi-call transaction
	var calls []rpc.InvokeFunctionCall
	var callDescriptions []string

	fmt.Printf("   💸 Building multi-call to fund both accounts...\n")

	// Check current balances and build mint calls
	for _, recipient := range recipients {
		fmt.Printf("   📊 Checking %s (%s)...\n", recipient.Name, recipient.Address)

		// Check current balance
		currentBalance, err := starknetutil.ERC20Balance(client, tokenAddress, recipient.Address)
		if err == nil {
			fmt.Printf("     📊 Current balance: %s\n", starknetutil.FormatTokenAmount(currentBalance, tokenDecimals))
		}

		// Convert recipient address to felt
		recipientFelt, err := utils.HexToFelt(recipient.Address)
		if err != nil {
			log.Printf("     ❌ Failed to convert recipient address to felt: %v", err)
			continue
		}

		// Build mint calldata: mint(to: ContractAddress, amount: u256)
		mintCalldata := []*felt.Felt{recipientFelt, amountLow, amountHigh}

		// Create mint call
		mintCall := rpc.InvokeFunctionCall{
			ContractAddress: tokenFelt,
			FunctionName:    "mint",
			CallData:        mintCalldata,
		}

		calls = append(calls, mintCall)
		callDescriptions = append(callDescriptions, fmt.Sprintf("mint(%s)", recipient.Name))
	}

	if len(calls) == 0 {
		log.Fatalf("No valid recipients found for funding")
	}

	// Log the multi-call composition
	fmt.Printf("   📝 Executing multi-call with [%s]...\n", strings.Join(callDescriptions, ", "))

	// Send multi-call transaction
	mintTx, err := minterAccount.BuildAndSendInvokeTxn(context.Background(), calls, nil)
	if err != nil {
		log.Fatalf("Failed to send multi-call mint transaction: %v", err)
	}

	fmt.Printf("   🚀 Multi-call mint transaction: %s\n", mintTx.Hash.String())

	// Wait for confirmation
	_, err = minterAccount.WaitForTransactionReceipt(context.Background(), mintTx.Hash, 2*time.Second)
	if err != nil {
		log.Fatalf("Failed to wait for transaction confirmation: %v", err)
	}

	fmt.Printf("   ✅ Multi-call transaction confirmed - funded all accounts!\n")

	// Verify new balances
	for _, recipient := range recipients {
		newBalance, err := starknetutil.ERC20Balance(client, tokenAddress, recipient.Address)
		if err == nil {
			fmt.Printf("   💰 %s new balance: %s\n", recipient.Name, starknetutil.FormatTokenAmount(newBalance, tokenDecimals))
		}
	}
}

type StarknetRecipient struct {
	Name    string
	Address string
}

func getStarknetRecipients() []StarknetRecipient {
	var recipients []StarknetRecipient

	// Alice
	recipients = append(recipients, StarknetRecipient{
		Name:    "Alice",
		Address: envutil.GetStarknetAliceAddress(),
	})

	// Solver
	recipients = append(recipients, StarknetRecipient{
		Name:    "Solver",
		Address: envutil.GetStarknetSolverAddress(),
	})

	return recipients
}
