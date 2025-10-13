package base

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
)

func TestSolverContext(t *testing.T) {
	t.Run("SolverContext creation", func(t *testing.T) {
		metadata := map[string]interface{}{
			"chain":   "base",
			"version": "1.0",
		}

		context := &SolverContext{
			Metadata: metadata,
		}

		assert.NotNil(t, context)
		assert.Equal(t, metadata, context.Metadata)

		// Test metadata access
		metadataMap, ok := context.Metadata.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "base", metadataMap["chain"])
		assert.Equal(t, "1.0", metadataMap["version"])
	})
}

func TestRuleInterface(t *testing.T) {
	t.Run("Rule function signature", func(t *testing.T) {
		// Test that Rule type is properly defined
		var rule Rule
		// Rule is a function type, so nil is valid
		assert.Nil(t, rule)

		// Test rule execution
		rule = func(args *types.ParsedArgs, context *SolverContext) error {
			return nil
		}

		args := types.ParsedArgs{}
		context := &SolverContext{}

		err := rule(&args, context)
		assert.NoError(t, err)
	})

	t.Run("Rule with error", func(t *testing.T) {
		expectedError := assert.AnError
		rule := func(_ *types.ParsedArgs, _ *SolverContext) error {
			return expectedError
		}

		args := types.ParsedArgs{}
		context := &SolverContext{}

		err := rule(&args, context)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestSolverImpl(t *testing.T) {
	t.Run("NewSolver creation", func(t *testing.T) {
		metadata := map[string]interface{}{
			"chain": "ethereum",
		}

		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{},
			BlockList: []types.AllowBlockListItem{},
		}

		solver := NewSolver(allowBlockLists, metadata)
		assert.NotNil(t, solver)

		assert.Equal(t, metadata, solver.metadata)
		assert.NotNil(t, solver.rules)
		assert.Equal(t, allowBlockLists, solver.allowBlockLists)
	})

	t.Run("AddRule", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		initialRuleCount := len(solver.rules)

		rule := func(args *types.ParsedArgs, context *SolverContext) error {
			return nil
		}

		solver.AddRule(rule)
		assert.Equal(t, initialRuleCount+1, len(solver.rules))
	})

	t.Run("GetRules", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		rule1 := func(args *types.ParsedArgs, context *SolverContext) error {
			return nil
		}
		rule2 := func(args *types.ParsedArgs, context *SolverContext) error {
			return nil
		}

		solver.AddRule(rule1)
		solver.AddRule(rule2)

		rules := solver.GetRules()
		assert.Len(t, rules, 2)
	})
}

func TestPrepareIntent(t *testing.T) {
	t.Run("Intent blocked by allow/block lists", func(t *testing.T) {
		// Set up block list
		blockList := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{},
			BlockList: []types.AllowBlockListItem{
				{
					SenderAddress:     "0x1234567890123456789012345678901234567890",
					DestinationDomain: "Base",
					RecipientAddress:  "*", // Wildcard
				},
			},
		}
		solver := NewSolver(blockList, map[string]interface{}{})

		// Create args that should be blocked
		args := types.ParsedArgs{
			SenderAddress: "0x1234567890123456789012345678901234567890",
			Recipients: []types.Recipient{
				{
					DestinationChainName: "Base",
					RecipientAddress:     "0x0987654321098765432109876543210987654321",
				},
			},
		}

		result, err := solver.PrepareIntent(context.Background(), &args)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Intent blocked by allow/block lists")
	})

	t.Run("Intent allowed by allow list", func(t *testing.T) {
		// Set up allow list
		allowList := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{
				{
					SenderAddress:     "0x1234567890123456789012345678901234567890",
					DestinationDomain: "Base",
					RecipientAddress:  "*", // Wildcard
				},
			},
			BlockList: []types.AllowBlockListItem{},
		}
		solver := NewSolver(allowList, map[string]interface{}{})

		// Create args that should be allowed
		args := types.ParsedArgs{
			SenderAddress: "0x1234567890123456789012345678901234567890",
			Recipients: []types.Recipient{
				{
					DestinationChainName: "Base",
					RecipientAddress:     "0x0987654321098765432109876543210987654321",
				},
			},
		}

		result, err := solver.PrepareIntent(context.Background(), &args)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
	})

	t.Run("Rule failure", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		// Add a rule that always fails
		failingRule := func(args *types.ParsedArgs, context *SolverContext) error {
			return assert.AnError
		}
		solver.AddRule(failingRule)

		args := types.ParsedArgs{}
		result, err := solver.PrepareIntent(context.Background(), &args)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Equal(t, assert.AnError.Error(), result.Error)
	})

	t.Run("All rules pass", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		// Add a rule that always passes
		passingRule := func(args *types.ParsedArgs, context *SolverContext) error {
			return nil
		}
		solver.AddRule(passingRule)

		// Create args with resolved order
		args := types.ParsedArgs{
			ResolvedOrder: types.ResolvedCrossChainOrder{
				FillInstructions: []types.FillInstruction{
					{
						DestinationChainID: big.NewInt(1),
						DestinationSettler: "0x1234567890123456789012345678901234567890",
						OriginData:         []byte("test data"),
					},
				},
			},
		}

		result, err := solver.PrepareIntent(context.Background(), &args)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.NotNil(t, result.Data)
		assert.Equal(t, args.ResolvedOrder.FillInstructions, result.Data.FillInstructions)
	})
}

func TestIsAllowedIntent(t *testing.T) {
	t.Run("Empty allow/block lists - should allow", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		args := types.ParsedArgs{
			SenderAddress: "0x1234567890123456789012345678901234567890",
			Recipients: []types.Recipient{
				{
					DestinationChainName: "Base",
					RecipientAddress:     "0x0987654321098765432109876543210987654321",
				},
			},
		}

		allowed := solver.isAllowedIntent(&args)
		assert.True(t, allowed)
	})

	t.Run("Blocked by block list", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{},
			BlockList: []types.AllowBlockListItem{
				{
					SenderAddress:     "0x1234567890123456789012345678901234567890",
					DestinationDomain: "Base",
					RecipientAddress:  "*",
				},
			},
		}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		args := types.ParsedArgs{
			SenderAddress: "0x1234567890123456789012345678901234567890",
			Recipients: []types.Recipient{
				{
					DestinationChainName: "Base",
					RecipientAddress:     "0x0987654321098765432109876543210987654321",
				},
			},
		}

		allowed := solver.isAllowedIntent(&args)
		assert.False(t, allowed)
	})

	t.Run("Allowed by allow list", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{
				{
					SenderAddress:     "0x1234567890123456789012345678901234567890",
					DestinationDomain: "Base",
					RecipientAddress:  "*",
				},
			},
			BlockList: []types.AllowBlockListItem{},
		}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		args := types.ParsedArgs{
			SenderAddress: "0x1234567890123456789012345678901234567890",
			Recipients: []types.Recipient{
				{
					DestinationChainName: "Base",
					RecipientAddress:     "0x0987654321098765432109876543210987654321",
				},
			},
		}

		allowed := solver.isAllowedIntent(&args)
		assert.True(t, allowed)
	})

	t.Run("Wildcard matching", func(t *testing.T) {
		allowBlockLists := types.AllowBlockLists{
			AllowList: []types.AllowBlockListItem{
				{
					SenderAddress:     "*",
					DestinationDomain: "Base",
					RecipientAddress:  "*",
				},
			},
			BlockList: []types.AllowBlockListItem{},
		}
		solver := NewSolver(allowBlockLists, map[string]interface{}{})

		args := types.ParsedArgs{
			SenderAddress: "0x1234567890123456789012345678901234567890",
			Recipients: []types.Recipient{
				{
					DestinationChainName: "Base",
					RecipientAddress:     "0x0987654321098765432109876543210987654321",
				},
			},
		}

		allowed := solver.isAllowedIntent(&args)
		assert.True(t, allowed)
	})
}

func TestNewListenerConfig(t *testing.T) {
	t.Run("NewListenerConfig with all parameters", func(t *testing.T) {
		config := NewListenerConfig(
			"0x1234567890123456789012345678901234567890",
			"Base",
			big.NewInt(1000),
			5000,
			2,
			10,
		)

		assert.NotNil(t, config)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", config.ContractAddress)
		assert.Equal(t, "Base", config.ChainName)
		assert.Equal(t, big.NewInt(1000), config.InitialBlock)
		assert.Equal(t, 5000, config.PollInterval)
		assert.Equal(t, uint64(2), config.ConfirmationBlocks)
		assert.Equal(t, uint64(10), config.MaxBlockRange)
	})

	t.Run("NewListenerConfig with default poll interval", func(t *testing.T) {
		config := NewListenerConfig(
			"0x1234567890123456789012345678901234567890",
			"Base",
			big.NewInt(1000),
			0, // Should default to 10000
			2,
			10,
		)

		assert.NotNil(t, config)
		assert.Equal(t, 10000, config.PollInterval) // Should use default
	})

	t.Run("NewListenerConfig with default max block range", func(t *testing.T) {
		config := NewListenerConfig(
			"0x1234567890123456789012345678901234567890",
			"Base",
			big.NewInt(1000),
			5000,
			2,
			0, // Should default to 9
		)

		assert.NotNil(t, config)
		assert.Equal(t, uint64(9), config.MaxBlockRange) // Should use default
	})
}

// MockSolver for testing ProcessIntent
type MockSolver struct {
	*solverImpl
	fillError   error
	settleError error
}

func (m *MockSolver) Fill(ctx context.Context, args *types.ParsedArgs, data types.IntentData, originChainName string, blockNumber uint64) error {
	return m.fillError
}

func (m *MockSolver) SettleOrder(ctx context.Context, args *types.ParsedArgs, data types.IntentData, originChainName string) error {
	return m.settleError
}

func TestProcessIntent(t *testing.T) {
	t.Run("ProcessIntent with rule failure", func(t *testing.T) {
		solver := NewSolver(types.AllowBlockLists{}, nil)

		// Add a rule that always fails
		solver.AddRule(func(args *types.ParsedArgs, context *SolverContext) error {
			return assert.AnError
		})

		args := types.ParsedArgs{
			OrderID: "test-order",
		}

		success, err := solver.ProcessIntent(context.Background(), &args, "Base", 1000)

		assert.NoError(t, err)
		assert.False(t, success) // Should return false when rules fail
	})

	t.Run("ProcessIntent with successful rules", func(t *testing.T) {
		solver := NewSolver(types.AllowBlockLists{}, nil)

		// Add a rule that always passes
		solver.AddRule(func(args *types.ParsedArgs, context *SolverContext) error {
			return nil
		})

		args := types.ParsedArgs{
			OrderID: "test-order",
		}

		// This will succeed because rules pass and base Fill/SettleOrder return nil
		success, err := solver.ProcessIntent(context.Background(), &args, "Base", 1000)

		assert.NoError(t, err)
		assert.True(t, success)
	})
}

func TestFill(t *testing.T) {
	t.Run("Fill method exists and can be called", func(t *testing.T) {
		solver := NewSolver(types.AllowBlockLists{}, nil)

		// Test that Fill method exists and can be called
		// This is a base implementation that should be overridden by concrete solvers
		args := types.ParsedArgs{
			OrderID: "test-order",
		}
		data := types.IntentData{}

		err := solver.Fill(context.Background(), &args, data, "Base", 1000)

		// The base implementation returns nil (no error) - it's a placeholder
		assert.NoError(t, err)
	})
}

func TestSettleOrder(t *testing.T) {
	t.Run("SettleOrder method exists and can be called", func(t *testing.T) {
		solver := NewSolver(types.AllowBlockLists{}, nil)

		// Test that SettleOrder method exists and can be called
		// This is a base implementation that should be overridden by concrete solvers
		args := types.ParsedArgs{
			OrderID: "test-order",
		}
		data := types.IntentData{}

		err := solver.SettleOrder(context.Background(), &args, data, "Base")

		// The base implementation returns nil (no error) - it's a placeholder
		assert.NoError(t, err)
	})
}
