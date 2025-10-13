package hyperlane7683

// Module: Rules system for Hyperlane7683 solver
// - Provides pluggable validation rules for order processing
// - Supports both EVM and Starknet chain-specific logic
// - Designed to be easily extensible for custom solver implementations

import (
	"context"
	"fmt"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/NethermindEth/oif-starknet/solver/pkg/ethutil"
	"github.com/NethermindEth/oif-starknet/solver/pkg/starknetutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/logutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
)

const (
	starknetNetworkName = "Starknet"
	// Profit margin calculation (100 = 100%)
	profitMarginMultiplier = 100
)

// RuleResult represents the result of a rule evaluation
type RuleResult struct {
	Passed bool
	Reason string
}

// Rule defines the interface for validation rules
// This interface is designed for plugin architecture - anyone can implement custom rules
type Rule interface {
	Name() string
	Evaluate(ctx context.Context, args *types.ParsedArgs) RuleResult
}

// RulesEngine coordinates rule evaluation
type RulesEngine struct {
	rules []Rule
}

// NewRulesEngine creates a new rules engine with default rules
func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		rules: []Rule{
			&BalanceRule{},
			&ProfitabilityRule{},
		},
	}
}

// AddRule adds a custom rule to the engine
func (re *RulesEngine) AddRule(rule Rule) {
	re.rules = append(re.rules, rule)
}

// EvaluateAll runs all rules and returns the first failure, or success if all pass
func (re *RulesEngine) EvaluateAll(ctx context.Context, args *types.ParsedArgs) RuleResult {
	// Get chain IDs for cross-chain logging
	originChainID := args.ResolvedOrder.OriginChainID.Uint64()
	destChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID.Uint64()

	for _, rule := range re.rules {
		result := rule.Evaluate(ctx, args)
		if !result.Passed {
			logutil.CrossChainOperation(fmt.Sprintf("Rule '%s' failed: %s", rule.Name(), result.Reason), originChainID, destChainID, args.OrderID)
			return result
		}
		logutil.CrossChainOperation(fmt.Sprintf("Rule '%s' passed", rule.Name()), originChainID, destChainID, args.OrderID)
	}
	return RuleResult{Passed: true, Reason: "All rules passed"}
}

// BalanceRule validates that the solver has sufficient balance for the order
type BalanceRule struct{}

func (br *BalanceRule) Name() string {
	return "BalanceCheck"
}

func (br *BalanceRule) Evaluate(ctx context.Context, args *types.ParsedArgs) RuleResult {
	if len(args.ResolvedOrder.MaxSpent) == 0 {
		return RuleResult{Passed: true, Reason: "No tokens to spend"}
	}

	// Get destination chain ID for routing
	destinationChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID.Uint64()

	// Switch based on destination chain type
	switch {
	case isStarknetChain(destinationChainID):
		return br.checkStarknetBalance(ctx, args)
	default:
		return br.checkEVMBalance(ctx, args)
	}
}

func (br *BalanceRule) checkStarknetBalance(_ context.Context, args *types.ParsedArgs) RuleResult {
	// Get solver's Starknet address from environment (conditional based on IS_DEVNET)
	solverAddrHex := envutil.GetStarknetSolverAddress()
	if solverAddrHex == "" {
		return RuleResult{Passed: false, Reason: "Starknet solver address not set"}
	}

	// Get Starknet RPC URL (conditional based on IS_DEVNET)
	starknetRPC := envutil.GetStarknetRPCURL()
	if starknetRPC == "" {
		return RuleResult{Passed: false, Reason: "STARKNET_RPC_URL not set"}
	}

	provider, err := rpc.NewProvider(starknetRPC)
	if err != nil {
		return RuleResult{Passed: false, Reason: fmt.Sprintf("Failed to create Starknet provider: %v", err)}
	}

	// Check balance for each token in MaxSpent (what solver needs to provide on Starknet)
	for _, maxSpent := range args.ResolvedOrder.MaxSpent {
		// Skip native ETH (empty string)
		if maxSpent.Token == "" || maxSpent.Token == "0x0" {
			continue
		}

		// Use starknetutil for balance check (assume valid input from order creation)
		balance, err := starknetutil.ERC20Balance(provider, maxSpent.Token, solverAddrHex)
		if err != nil {
			return RuleResult{Passed: false, Reason: fmt.Sprintf("Failed to check balance for token %s: %v", maxSpent.Token, err)}
		}

		if balance.Cmp(maxSpent.Amount) < 0 {
			return RuleResult{Passed: false, Reason: fmt.Sprintf("Insufficient balance for token %s: have %s, need %s",
				maxSpent.Token, balance.String(), maxSpent.Amount.String())}
		}
	}

	return RuleResult{Passed: true, Reason: "Starknet balance check passed"}
}

func (br *BalanceRule) checkEVMBalance(_ context.Context, args *types.ParsedArgs) RuleResult {
	// Get solver's EVM address from environment (conditional based on IS_DEVNET)
	solverAddrHex := envutil.GetSolverPublicKey()
	if solverAddrHex == "" {
		return RuleResult{Passed: false, Reason: "Solver public key not set"}
	}

	solverAddr := common.HexToAddress(solverAddrHex)
	destinationChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID.Uint64()

	// Find the network config for destination chain
	var networkConfig *config.NetworkConfig
	for _, network := range config.Networks {
		if network.ChainID == destinationChainID {
			networkConfig = &network
			break
		}
	}

	if networkConfig == nil {
		return RuleResult{Passed: false, Reason: fmt.Sprintf("No network config found for chain ID %d", destinationChainID)}
	}

	// Connect to destination chain RPC
	client, err := ethclient.Dial(networkConfig.RPCURL)
	if err != nil {
		return RuleResult{Passed: false, Reason: fmt.Sprintf("Failed to connect to EVM RPC: %v", err)}
	}
	defer client.Close()

	// Check balance for each token in MaxSpent (what solver needs to provide on destination chain)
	for _, maxSpent := range args.ResolvedOrder.MaxSpent {
		// Skip native ETH (empty string)
		if maxSpent.Token == "" {
			continue
		}

		// Convert token address using address_utils (assume valid input from order creation)
		tokenAddr, err := types.ToEVMAddress(maxSpent.Token)
		if err != nil {
			return RuleResult{Passed: false, Reason: fmt.Sprintf("Failed to convert token address %s: %v", maxSpent.Token, err)}
		}

		// Use ethutil for balance check
		balance, err := ethutil.ERC20Balance(client, tokenAddr, solverAddr)
		if err != nil {
			return RuleResult{Passed: false, Reason: fmt.Sprintf("Failed to check balance for token %s: %v", maxSpent.Token, err)}
		}

		if balance.Cmp(maxSpent.Amount) < 0 {
			return RuleResult{Passed: false, Reason: fmt.Sprintf("Insufficient balance for token %s: have %s, need %s",
				maxSpent.Token, balance.String(), maxSpent.Amount.String())}
		}
	}

	return RuleResult{Passed: true, Reason: "EVM balance check passed"}
}

// ProfitabilityRule validates that the order is profitable for the solver
type ProfitabilityRule struct{}

func (pr *ProfitabilityRule) Name() string {
	return "ProfitabilityCheck"
}

func (pr *ProfitabilityRule) Evaluate(ctx context.Context, args *types.ParsedArgs) RuleResult {
	// Calculate expected profit from the order
	// This involves comparing MaxSpent vs MinReceived

	if len(args.ResolvedOrder.MaxSpent) == 0 || len(args.ResolvedOrder.MinReceived) == 0 {
		return RuleResult{Passed: false, Reason: "Missing MaxSpent or MinReceived data"}
	}

	// Simple profitability check: ensure MaxSpent > MinReceived
	// In a real implementation, this would be more sophisticated
	// considering gas costs, slippage, etc.

	// Get chain IDs for cross-chain logging
	originChainID := args.ResolvedOrder.OriginChainID.Uint64()
	destChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID.Uint64()
	logutil.CrossChainOperation("Checking order profitability", originChainID, destChainID, args.OrderID)

	// Basic profitability check: ensure MinReceived > MaxSpent + expectedFees
	// NOTE: This is a simplified check that assumes same token types and doesn't account for:
	// - Token price differences (would need oracles)
	// - Slippage and market conditions
	// - Cross-chain value differences
	//
	// For production use, consider:
	// - Integrating price oracles (Chainlink, etc.)
	// - Using approved token allow lists with known price feeds
	// - Implementing more sophisticated profit margin calculations
	// - Accounting for actual gas costs and protocol fees
	// - Adding minimum profit thresholds based on risk tolerance

	// Expected fees and minimum profit threshold
	// TODO: Calculate actual gas costs for fill + settle operations
	// TODO: Add protocol fees (Hyperlane, etc.)
	// TODO: Consider minimum profit threshold based on risk/reward
	expectedFees := uint256.NewInt(0)       // Placeholder - should be calculated based on gas costs
	minProfitThreshold := uint256.NewInt(0) // Placeholder - minimum profit to consider order worthwhile

	// Calculate total MaxSpent (what we're spending)
	totalMaxSpent := uint256.NewInt(0)
	for _, maxSpent := range args.ResolvedOrder.MaxSpent {
		maxSpentU256, _ := uint256.FromBig(maxSpent.Amount)
		totalMaxSpent.Add(totalMaxSpent, maxSpentU256)
	}

	// Calculate total MinReceived (what we expect to receive)
	totalMinReceived := uint256.NewInt(0)
	for _, minReceived := range args.ResolvedOrder.MinReceived {
		minReceivedU256, _ := uint256.FromBig(minReceived.Amount)
		totalMinReceived.Add(totalMinReceived, minReceivedU256)
	}

	// Calculate total costs (MaxSpent + expected fees)
	totalCosts := new(uint256.Int).Add(totalMaxSpent, expectedFees)

	// Basic check: MinReceived should be greater than TotalCosts (solver profit > 0)
	if totalMinReceived.Cmp(totalCosts) <= 0 {
		return RuleResult{
			Passed: false,
			Reason: fmt.Sprintf("Order not profitable: MinReceived (%s) <= TotalCosts (%s + %s fees)",
				totalMinReceived.Dec(), totalMaxSpent.Dec(), expectedFees.Dec()),
		}
	}

	// Calculate gross profit (before fees) and net profit (after fees)
	grossProfit := new(uint256.Int).Sub(totalMinReceived, totalMaxSpent)
	netProfit := new(uint256.Int).Sub(totalMinReceived, totalCosts)

	// Check if net profit meets minimum threshold
	if netProfit.Cmp(minProfitThreshold) < 0 {
		return RuleResult{
			Passed: false,
			Reason: fmt.Sprintf("Order profit below threshold: NetProfit (%s) < MinThreshold (%s)",
				netProfit.Dec(), minProfitThreshold.Dec()),
		}
	}

	// Calculate profit margin for logging (based on gross profit vs MaxSpent)
	profitMargin := new(uint256.Int).Mul(grossProfit, uint256.NewInt(profitMarginMultiplier))
	profitMargin.Div(profitMargin, totalMaxSpent)

	logutil.CrossChainOperation(fmt.Sprintf("Profitability check passed: NetProfit=%s, GrossProfit=%s (%.2f%% margin)",
		netProfit.Dec(), grossProfit.Dec(), float64(profitMargin.Uint64())), originChainID, destChainID, args.OrderID)

	return RuleResult{Passed: true, Reason: fmt.Sprintf("Order profitable: NetProfit=%s, GrossProfit=%s (%.2f%% margin)",
		netProfit.Dec(), grossProfit.Dec(), float64(profitMargin.Uint64()))}
}

// Helper function to determine if a chain ID is Starknet
func isStarknetChain(chainID uint64) bool {
	for _, network := range config.Networks {
		if network.ChainID == chainID && network.Name == starknetNetworkName {
			return true
		}
	}
	return false
}

// Helper function to get chain type (EVM or Starknet)
// func getChainType(chainID uint64) string {
//	if isStarknetChain(chainID) {
//		return "Starknet"
//	}
//	return "EVM"
// }
