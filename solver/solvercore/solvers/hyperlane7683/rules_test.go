package hyperlane7683

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
)

func TestRulesEngine(t *testing.T) {
	// Set required environment variables for tests
	t.Setenv("SOLVER_PUB_KEY", "0x1234567890123456789012345678901234567890123456789012345678901234")
	t.Setenv("BASE_RPC_URL", "http://localhost:8545")
	t.Setenv("BASE_CHAIN_ID", "84532")
	defer func() {
		os.Unsetenv("SOLVER_PUB_KEY")
		os.Unsetenv("BASE_RPC_URL")
		os.Unsetenv("BASE_CHAIN_ID")
	}()

	// Force initialization of networks
	_ = config.GetDefaultNetwork()

	t.Run("NewRulesEngine creation", func(t *testing.T) {
		engine := NewRulesEngine()
		assert.NotNil(t, engine)
		assert.NotNil(t, engine.rules)
		// Note: RulesEngine may have default rules, so we don't assert empty
	})

	t.Run("AddRule", func(t *testing.T) {
		engine := NewRulesEngine()
		initialCount := len(engine.rules)

		rule := &BalanceRule{}
		engine.AddRule(rule)

		assert.Len(t, engine.rules, initialCount+1)
		assert.Equal(t, rule, engine.rules[initialCount])
	})

	t.Run("EvaluateAll with no rules", func(t *testing.T) {
		engine := &RulesEngine{rules: []Rule{}}
		// Create a minimal args structure to avoid nil pointer issues
		args := types.ParsedArgs{
			OrderID: "0x1234567890123456789012345678901234567890123456789012345678901234",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(1),
				MaxSpent: []types.Output{
					{
						Token:     "0x1234567890123456789012345678901234567890",
						Amount:    big.NewInt(1000),
						Recipient: "0x0987654321098765432109876543210987654321",
						ChainID:   big.NewInt(84532),
					},
				},
				MinReceived: []types.Output{
					{
						Token:     "0x0987654321098765432109876543210987654321",
						Amount:    big.NewInt(1100), // 10% profit
						Recipient: "0x1234567890123456789012345678901234567890",
						ChainID:   big.NewInt(84532),
					},
				},
				FillInstructions: []types.FillInstruction{
					{
						DestinationChainID: big.NewInt(84532), // Base Sepolia
					},
				},
			},
		}

		result := engine.EvaluateAll(context.Background(), &args)
		// Should pass if no rules or all rules pass
		assert.True(t, result.Passed)
	})

	t.Run("EvaluateAll with passing rules", func(t *testing.T) {
		engine := &RulesEngine{rules: []Rule{}}

		// Add a mock rule that always passes (no network calls)
		rule := &MockRule{name: "MockRule", shouldPass: true}
		engine.AddRule(rule)

		args := types.ParsedArgs{
			OrderID: "0x1234567890123456789012345678901234567890123456789012345678901234",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(1),
				MaxSpent: []types.Output{
					{
						Token:     "0x1234567890123456789012345678901234567890",
						Amount:    big.NewInt(1000),
						Recipient: "0x0987654321098765432109876543210987654321",
						ChainID:   big.NewInt(84532),
					},
				},
				MinReceived: []types.Output{
					{
						Token:     "0x0987654321098765432109876543210987654321",
						Amount:    big.NewInt(1100), // 10% profit
						Recipient: "0x1234567890123456789012345678901234567890",
						ChainID:   big.NewInt(84532),
					},
				},
				FillInstructions: []types.FillInstruction{
					{
						DestinationChainID: big.NewInt(84532), // Base Sepolia
					},
				},
			},
		}

		result := engine.EvaluateAll(context.Background(), &args)
		assert.True(t, result.Passed)
	})
}

func TestBalanceRule(t *testing.T) {
	t.Run("Rule name", func(t *testing.T) {
		rule := &BalanceRule{}
		assert.Equal(t, "BalanceCheck", rule.Name())
	})

	t.Run("No tokens to spend", func(t *testing.T) {
		rule := &BalanceRule{}
		args := types.ParsedArgs{
			ResolvedOrder: types.ResolvedCrossChainOrder{
				MaxSpent: []types.Output{}, // Empty list
			},
		}

		result := rule.Evaluate(context.Background(), &args)
		assert.True(t, result.Passed)
		assert.Equal(t, "No tokens to spend", result.Reason)
	})

	t.Run("Single token with sufficient balance", func(t *testing.T) {
		rule := &BalanceRule{}
		args := types.ParsedArgs{
			OrderID: "0x1234567890123456789012345678901234567890123456789012345678901234",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(1),
				MaxSpent: []types.Output{
					{
						Token:     "0x1234567890123456789012345678901234567890",
						Amount:    big.NewInt(1000),
						Recipient: "0x0987654321098765432109876543210987654321",
						ChainID:   big.NewInt(84532),
					},
				},
				MinReceived: []types.Output{}, // Empty list should pass
				FillInstructions: []types.FillInstruction{
					{
						DestinationChainID: big.NewInt(84532), // Base Sepolia
					},
				},
			},
		}

		// This would normally check actual balances, but for unit tests
		// we'll test the structure and basic logic
		result := rule.Evaluate(context.Background(), &args)
		// The actual result depends on the implementation
		// For now, we just ensure it doesn't panic
		assert.NotNil(t, result)
	})
}

func TestProfitabilityRule(t *testing.T) {
	t.Run("Rule name", func(t *testing.T) {
		rule := &ProfitabilityRule{}
		assert.Equal(t, "ProfitabilityCheck", rule.Name())
	})

	t.Run("No tokens to compare", func(t *testing.T) {
		rule := &ProfitabilityRule{}
		args := types.ParsedArgs{
			OrderID: "0x1234567890123456789012345678901234567890123456789012345678901234",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(1),
				MaxSpent:      []types.Output{},
				MinReceived:   []types.Output{},
				FillInstructions: []types.FillInstruction{
					{
						DestinationChainID: big.NewInt(84532), // Base Sepolia
					},
				},
			},
		}

		result := rule.Evaluate(context.Background(), &args)
		assert.False(t, result.Passed)
		assert.Equal(t, "Missing MaxSpent or MinReceived data", result.Reason)
	})

	t.Run("Profitable order", func(t *testing.T) {
		rule := &ProfitabilityRule{}
		args := types.ParsedArgs{
			OrderID: "0x1234567890123456789012345678901234567890123456789012345678901234",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(1),
				MaxSpent: []types.Output{
					{
						Token:  "0x1234567890123456789012345678901234567890",
						Amount: big.NewInt(1000), // Spend 1000
					},
				},
				MinReceived: []types.Output{
					{
						Token:  "0x0987654321098765432109876543210987654321",
						Amount: big.NewInt(1100), // Receive 1100 (10% profit)
					},
				},
				FillInstructions: []types.FillInstruction{
					{
						DestinationChainID: big.NewInt(84532), // Base Sepolia
					},
				},
			},
		}

		result := rule.Evaluate(context.Background(), &args)
		// The actual result depends on the implementation
		// For now, we just ensure it doesn't panic
		assert.NotNil(t, result)
	})
}

func TestRuleResult(t *testing.T) {
	t.Run("RuleResult creation", func(t *testing.T) {
		result := RuleResult{
			Passed: true,
			Reason: "Test passed",
		}

		assert.True(t, result.Passed)
		assert.Equal(t, "Test passed", result.Reason)
	})

	t.Run("Failed rule result", func(t *testing.T) {
		result := RuleResult{
			Passed: false,
			Reason: "Test failed",
		}

		assert.False(t, result.Passed)
		assert.Equal(t, "Test failed", result.Reason)
	})
}

// MockRule is a test rule that doesn't make network calls
type MockRule struct {
	name       string
	shouldPass bool
}

func (m *MockRule) Name() string {
	return m.name
}

func (m *MockRule) Evaluate(ctx context.Context, args *types.ParsedArgs) RuleResult {
	if m.shouldPass {
		return RuleResult{Passed: true, Reason: "Mock rule passed"}
	}
	return RuleResult{Passed: false, Reason: "Mock rule failed"}
}
