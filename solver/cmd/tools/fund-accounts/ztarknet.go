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
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

func fundZtarknet(amount *big.Int) {
	fmt.Printf("📡 Funding Ztarknet network...\n")

	// Get RPC URL from environment
	rpcURL := envutil.GetZtarknetRPCURL()
	if rpcURL == "" {
		log.Fatalf("ZTARKNET_RPC_URL not found in environment")
	}

	// Get chain ID from environment
	chainIDStr := os.Getenv("ZTARKNET_CHAIN_ID")
	if chainIDStr == "" {
		log.Fatalf("ZTARKNET_CHAIN_ID not found in environment")
	}

	// Get MockERC20 address from environment
	tokenAddress := os.Getenv("ZTARKNET_DOG_COIN_ADDRESS")
	if tokenAddress == "" {
		log.Fatalf("ZTARKNET_DOG_COIN_ADDRESS not found in environment")
	}

	// Connect to Ztarknet
	client, err := rpc.NewProvider(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ztarknet: %v", err)
	}

	fmt.Printf("   📍 Network: Ztarknet (Chain ID: %s)\n", chainIDStr)
	fmt.Printf("   🪙 MockERC20: %s\n", tokenAddress)

	// Get minter account (use Alice as minter)
	minterPrivateKey := envutil.GetZtarknetAlicePrivateKey()
	minterPublicKey := envutil.GetZtarknetAlicePublicKey()
	minterAddress := envutil.GetZtarknetAliceAddress()

	if minterPrivateKey == "" || minterPublicKey == "" || minterAddress == "" {
		log.Fatalf("Ztarknet minter credentials not found (Alice's keys)")
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
	recipients := getZtarknetRecipients()

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

type ZtarknetRecipient struct {
	Name    string
	Address string
}

func getZtarknetRecipients() []ZtarknetRecipient {
	var recipients []ZtarknetRecipient

	// Alice and Solver
	recipients = append(recipients,
		ZtarknetRecipient{
			Name:    "Alice",
			Address: envutil.GetZtarknetAliceAddress(),
		},
		ZtarknetRecipient{
			Name:    "Solver",
			Address: envutil.GetZtarknetSolverAddress(),
		},
	)

	return recipients
}

