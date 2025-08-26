package hyperlane7683

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/filler"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Rule implementations for Hyperlane7683 protocol
// Following the modular structure from TypeScript reference

// EnoughBalanceOnDestination validates that the filler has sufficient token balances
// before attempting to fill orders (prevents failed fills due to insufficient funds)
func (f *Hyperlane7683Filler) enoughBalanceOnDestination(args types.ParsedArgs, ctx *filler.FillerContext) error {
	fmt.Printf("   🔍 Validating filler token balances across chains...\n")

	// Group amounts by chain and token
	amountByTokenByChain := make(map[uint64]map[common.Address]*big.Int)

	for _, output := range args.ResolvedOrder.MaxSpent {
		chainID := output.ChainID.Uint64()
		
		// Check if this is a Starknet chain using dynamic detection
		if f.isStarknetChain(output.ChainID) {
			// For Starknet, skip balance validation for now
			// TODO: Implement proper Starknet RPC balance checking
			fmt.Printf("   ⚠️  Skipping Starknet balance check for chain %d (not implemented yet)\n", chainID)
			continue
		}
		
		// Handle EVM chains normally
		tokenAddr := output.Token

		if amountByTokenByChain[chainID] == nil {
			amountByTokenByChain[chainID] = make(map[common.Address]*big.Int)
		}

		if amountByTokenByChain[chainID][tokenAddr] == nil {
			amountByTokenByChain[chainID][tokenAddr] = big.NewInt(0)
		}

		amountByTokenByChain[chainID][tokenAddr].Add(
			amountByTokenByChain[chainID][tokenAddr],
			output.Amount,
		)
	}

	// Check balances for each EVM chain and token
	for chainID, tokenAmounts := range amountByTokenByChain {
		client, err := f.getClientForChain(big.NewInt(int64(chainID)))
		if err != nil {
			return fmt.Errorf("failed to get client for chain %d: %w", chainID, err)
		}

		signer, err := f.getSignerForChain(big.NewInt(int64(chainID)))
		if err != nil {
			return fmt.Errorf("failed to get signer for chain %d: %w", chainID, err)
		}

		fillerAddress := signer.From

		for tokenAddr, requiredAmount := range tokenAmounts {
			balance, err := f.getTokenBalance(client, tokenAddr, fillerAddress)
			if err != nil {
				return fmt.Errorf("failed to get balance for token %s on chain %d: %w", tokenAddr.Hex(), chainID, err)
			}

			if balance.Cmp(requiredAmount) < 0 {
				return fmt.Errorf("insufficient balance on chain %d for token %s: have %s, need %s",
					chainID, tokenAddr.Hex(), balance.String(), requiredAmount.String())
			}

			fmt.Printf("   ✅ Chain %d Token %s: Balance %s >= Required %s\n",
				chainID, tokenAddr.Hex(), balance.String(), requiredAmount.String())
		}
	}

	fmt.Printf("   ✅ All token balance validations passed\n")
	return nil
}

// FilterByTokenAndAmount validates that tokens and amounts are within allowed limits
// Supports configurable per-chain, per-token limits (following TypeScript structure)
func (f *Hyperlane7683Filler) filterByTokenAndAmount(args types.ParsedArgs, ctx *filler.FillerContext) error {
	// TODO: Make this configurable via metadata CustomRules
	// For now, implement basic profitability check like TypeScript version

	if len(args.ResolvedOrder.MinReceived) == 0 || len(args.ResolvedOrder.MaxSpent) == 0 {
		return fmt.Errorf("invalid order: missing minReceived or maxSpent")
	}

	minReceived := args.ResolvedOrder.MinReceived[0].Amount
	maxSpent := args.ResolvedOrder.MaxSpent[0].Amount

	// Basic profitability check - we should receive more than we spend
	if minReceived.Cmp(maxSpent) <= 0 {
		return fmt.Errorf("intent is not profitable: minReceived %s <= maxSpent %s",
			minReceived.String(), maxSpent.String())
	}

	fmt.Printf("   ✅ Profitability check passed: profit = %s\n",
		new(big.Int).Sub(minReceived, maxSpent).String())

	return nil
}

// getTokenBalance retrieves the token balance for an address
func (f *Hyperlane7683Filler) getTokenBalance(client *ethclient.Client, tokenAddr, holderAddr common.Address) (*big.Int, error) {
	// Handle native token (ETH)
	if tokenAddr == (common.Address{}) {
		return client.BalanceAt(context.Background(), holderAddr, nil)
	}

	// Handle ERC20 tokens
	balanceOfABI := `[{"type":"function","name":"balanceOf","inputs":[{"type":"address","name":"account"}],"outputs":[{"type":"uint256","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(balanceOfABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse balanceOf ABI: %w", err)
	}

	callData, err := parsedABI.Pack("balanceOf", holderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf call: %w", err)
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{To: &tokenAddr, Data: callData}, nil)
	if err != nil {
		return nil, fmt.Errorf("balanceOf call failed: %w", err)
	}

	if len(result) < 32 {
		return nil, fmt.Errorf("invalid balanceOf result length: %d", len(result))
	}

	return new(big.Int).SetBytes(result), nil
}
