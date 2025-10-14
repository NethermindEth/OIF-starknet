package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
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

	// Fund each recipient
	for _, recipient := range recipients {
		fmt.Printf("   💸 Funding %s (%s)...\n", recipient.Name, recipient.Address)

		// Check current balance
		currentBalance, err := starknetutil.ERC20Balance(client, tokenAddress, recipient.Address)
		if err == nil {
			fmt.Printf("     📊 Current balance: %s\n", starknetutil.FormatTokenAmount(currentBalance, tokenDecimals))
		}

		// Convert addresses to felts for the mint call
		tokenFelt, err := utils.HexToFelt(tokenAddress)
		if err != nil {
			log.Printf("     ❌ Failed to convert token address to felt: %v", err)
			continue
		}

		recipientFelt, err := utils.HexToFelt(recipient.Address)
		if err != nil {
			log.Printf("     ❌ Failed to convert recipient address to felt: %v", err)
			continue
		}

		// Convert amount to two felts (low, high) for u256
		amountLow, amountHigh := starknetutil.ConvertBigIntToU256Felts(amount)

		// Build mint calldata: mint(to: ContractAddress, amount: u256)
		mintCalldata := []*felt.Felt{recipientFelt, amountLow, amountHigh}

		// Create mint transaction
		mintCall := rpc.InvokeFunctionCall{
			ContractAddress: tokenFelt,
			FunctionName:    "mint",
			CallData:        mintCalldata,
		}

		// Send mint transaction
		mintTx, err := minterAccount.BuildAndSendInvokeTxn(context.Background(), []rpc.InvokeFunctionCall{mintCall}, nil)
		if err != nil {
			log.Printf("     ❌ Failed to send mint transaction for %s: %v", recipient.Name, err)
			continue
		}

		fmt.Printf("     🚀 Mint transaction: %s\n", mintTx.Hash.String())

		// Wait for confirmation
		_, err = minterAccount.WaitForTransactionReceipt(context.Background(), mintTx.Hash, 2*time.Second)
		if err != nil {
			log.Printf("     ❌ Failed to wait for transaction confirmation: %v", err)
			continue
		}

		fmt.Printf("     ✅ Minted %s tokens\n", starknetutil.FormatTokenAmount(amount, tokenDecimals))

		// Verify new balance
		newBalance, err := starknetutil.ERC20Balance(client, tokenAddress, recipient.Address)
		if err == nil {
			fmt.Printf("     💰 New balance: %s\n", starknetutil.FormatTokenAmount(newBalance, tokenDecimals))
		}
	}
}

type StarknetRecipient struct {
	Name    string
	Address string
}

func getStarknetRecipients() []StarknetRecipient {
	var recipients []StarknetRecipient

	// Alice and Solver
	recipients = append(recipients,
		StarknetRecipient{
			Name:    "Alice",
			Address: envutil.GetStarknetAliceAddress(),
		},
		StarknetRecipient{
			Name:    "Solver",
			Address: envutil.GetStarknetSolverAddress(),
		},
	)

	return recipients
}
