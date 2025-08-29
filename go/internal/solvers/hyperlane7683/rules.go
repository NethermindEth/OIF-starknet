package hyperlane7683

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/oif-starknet/go/internal/base"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Rule implementations for Hyperlane7683 protocol
// Following the modular structure from TypeScript reference

// EnoughBalanceOnDestination validates that the solver has sufficient token balances
// before attempting to fill orders (prevents failed fills due to insufficient funds)
func (f *Hyperlane7683Solver) enoughBalanceOnDestination(args types.ParsedArgs, ctx *base.SolverContext) error {
	fmt.Printf("   🔍 Validating solver token balances across chains...\n")

	// Group amounts by chain and token
	amountByTokenByChain := make(map[uint64]map[common.Address]*big.Int)

	for _, output := range args.ResolvedOrder.MaxSpent {
		chainID := output.ChainID.Uint64()

		// Check if this is a Starknet chain using dynamic detection
		if f.isStarknetChain(output.ChainID) {
			// For Starknet, implement balance validation using Starknet RPC
			if err := f.validateStarknetBalance(output); err != nil {
				return fmt.Errorf("Starknet balance validation failed: %w", err)
			}
			continue
		}

		// Handle EVM chains normally
		tokenAddr := output.Token

		// Convert string address to EVM address for map operations
		tokenAddrEVM, err := types.ToEVMAddress(tokenAddr)
		if err != nil {
			return fmt.Errorf("failed to convert token address %s to EVM format: %w", tokenAddr, err)
		}

		if amountByTokenByChain[chainID] == nil {
			amountByTokenByChain[chainID] = make(map[common.Address]*big.Int)
		}

		if amountByTokenByChain[chainID][tokenAddrEVM] == nil {
			amountByTokenByChain[chainID][tokenAddrEVM] = big.NewInt(0)
		}

		amountByTokenByChain[chainID][tokenAddrEVM].Add(
			amountByTokenByChain[chainID][tokenAddrEVM],
			output.Amount,
		)
	}

	// Check balances for each EVM chain and token
	for chainID, tokenAmounts := range amountByTokenByChain {
		client, err := f.getEVMClient(chainID)
		if err != nil {
			return fmt.Errorf("failed to get client for chain %d: %w", chainID, err)
		}

		signer, err := f.getEVMSigner(chainID)
		if err != nil {
			return fmt.Errorf("failed to get signer for chain %d: %w", chainID, err)
		}

		solverAddress := signer.From

		for tokenAddr, requiredAmount := range tokenAmounts {
			balance, err := f.getTokenBalance(client, tokenAddr, solverAddress)
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

// validateStarknetBalance checks if the solver has sufficient token balance on Starknet
func (f *Hyperlane7683Solver) validateStarknetBalance(output types.Output) error {
	// Get Starknet network config
	chainConfig, err := f.getNetworkConfigByChainID(output.ChainID)
	if err != nil {
		return fmt.Errorf("failed to get Starknet network config: %w", err)
	}

	// Convert token address to Starknet format
	tokenAddressHex := f.getStarknetTokenAddress(output)

	// Get Starknet solver address from environment
	starknetSolverAddr := os.Getenv("STARKNET_SOLVER_ADDRESS")
	if starknetSolverAddr == "" {
		return fmt.Errorf("STARKNET_SOLVER_ADDRESS environment variable not set")
	}

	// Check token balance on Starknet using direct RPC call
	balance, err := f.getStarknetTokenBalance(chainConfig.RPCURL, tokenAddressHex, starknetSolverAddr)
	if err != nil {
		return fmt.Errorf("failed to get Starknet token balance: %w", err)
	}

	// Compare balance with required amount
	if balance.Cmp(output.Amount) < 0 {
		return fmt.Errorf("insufficient Starknet balance for token %s: have %s, need %s",
			tokenAddressHex, balance.String(), output.Amount.String())
	}

	fmt.Printf("   ✅ Starknet Chain %d Token %s: Balance %s >= Required %s\n",
		output.ChainID.Uint64(), tokenAddressHex, balance.String(), output.Amount.String())

	return nil
}

// getStarknetTokenBalance retrieves token balance directly using RPC (without full StarknetSolver)
func (f *Hyperlane7683Solver) getStarknetTokenBalance(rpcURL, tokenAddressHex, holderAddressHex string) (*big.Int, error) {
	// Create a minimal provider just for balance checking
	provider, err := rpc.NewProvider(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Starknet provider: %w", err)
	}

	// Convert addresses to felt format
	tokenAddr, err := utils.HexToFelt(tokenAddressHex)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	holderAddr, err := utils.HexToFelt(holderAddressHex)
	if err != nil {
		return nil, fmt.Errorf("invalid holder address: %w", err)
	}

	// Call balanceOf function on the ERC20 contract
	// balanceOf(address) -> uint256
	// Starknet uses function name selectors, not hardcoded hex
	balanceOfSelector := utils.GetSelectorFromNameFelt("balanceOf")

	call := rpc.FunctionCall{
		ContractAddress:    tokenAddr,
		EntryPointSelector: balanceOfSelector,
		Calldata:           []*felt.Felt{holderAddr},
	}

	result, err := provider.Call(context.Background(), call, rpc.BlockID{Tag: "latest"})
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("balanceOf returned no results")
	}

	// Convert felt result to big.Int
	balance := utils.FeltToBigInt(result[0])
	return balance, nil
}

// getStarknetTokenAddress converts EVM token address to Starknet format
func (f *Hyperlane7683Solver) getStarknetTokenAddress(output types.Output) string {
	// For Starknet destinations, use the token address directly
	if f.isStarknetChain(output.ChainID) {
		fmt.Printf("   🎯 Using Starknet token address: %s\n", output.Token)
		return output.Token
	}

	// For EVM destinations, convert to Starknet format if needed
	fmt.Printf("   ⚠️  Using token address as-is: %s\n", output.Token)
	return output.Token
}

// FilterByTokenAndAmount validates that tokens and amounts are within allowed limits
// Supports configurable per-chain, per-token limits (following TypeScript structure)
func (f *Hyperlane7683Solver) filterByTokenAndAmount(args types.ParsedArgs, ctx *base.SolverContext) error {
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
func (f *Hyperlane7683Solver) getTokenBalance(client *ethclient.Client, tokenAddr, holderAddr common.Address) (*big.Int, error) {
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
