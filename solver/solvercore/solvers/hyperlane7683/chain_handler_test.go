package hyperlane7683

import (
	"context"
	"testing"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChainHandlerInterface tests the ChainHandler interface
func TestChainHandlerInterface(t *testing.T) {
	t.Run("ChainHandler_interface_methods", func(t *testing.T) {
		// Test that our mock handler implements all required methods
		handler := &mockChainHandler{}

		// Test Fill method
		args := types.ParsedArgs{
			OrderID:       "test-order-123",
			SenderAddress: "0x1234567890123456789012345678901234567890",
		}

		action, err := handler.Fill(context.Background(), &args)
		require.NoError(t, err)
		assert.Equal(t, OrderActionComplete, action)

		// Test Settle method
		err = handler.Settle(context.Background(), &args)
		require.NoError(t, err)
	})

	t.Run("ChainHandler_different_actions", func(t *testing.T) {
		// Test different OrderAction values
		testCases := []struct {
			name   string
			action OrderAction
		}{
			{"Settle", OrderActionSettle},
			{"Complete", OrderActionComplete},
			{"Error", OrderActionError},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				handler := &mockChainHandlerWithAction{action: tc.action}

				action, err := handler.Fill(context.Background(), &types.ParsedArgs{})
				require.NoError(t, err)
				assert.Equal(t, tc.action, action)
			})
		}
	})
}

// TestOrderActionEnum tests the OrderAction enum values
func TestOrderActionEnum(t *testing.T) {
	t.Run("OrderAction_values", func(t *testing.T) {
		assert.Equal(t, OrderAction(0), OrderActionSettle)
		assert.Equal(t, OrderAction(1), OrderActionComplete)
		assert.Equal(t, OrderAction(2), OrderActionError)
	})

	t.Run("OrderAction_comparison", func(t *testing.T) {
		// Test that we can compare OrderAction values
		assert.True(t, OrderActionSettle < OrderActionComplete)
		assert.True(t, OrderActionComplete < OrderActionError)
		assert.True(t, OrderActionSettle < OrderActionError)
	})

	t.Run("OrderAction_equality", func(t *testing.T) {
		// Test equality
		assert.Equal(t, OrderActionSettle, OrderAction(0))
		assert.Equal(t, OrderActionComplete, OrderAction(1))
		assert.Equal(t, OrderActionError, OrderAction(2))

		// Test inequality
		assert.NotEqual(t, OrderActionSettle, OrderActionComplete)
		assert.NotEqual(t, OrderActionComplete, OrderActionError)
		assert.NotEqual(t, OrderActionSettle, OrderActionError)
	})
}

// TestChainHandlerErrorHandling tests error handling in chain handlers
func TestChainHandlerErrorHandling(t *testing.T) {
	t.Run("handler_with_error", func(t *testing.T) {
		handler := &mockChainHandlerWithError{shouldError: true}

		action, err := handler.Fill(context.Background(), &types.ParsedArgs{})
		assert.Error(t, err)
		assert.Equal(t, OrderActionError, action)
	})

	t.Run("handler_without_error", func(t *testing.T) {
		handler := &mockChainHandlerWithError{shouldError: false}

		action, err := handler.Fill(context.Background(), &types.ParsedArgs{})
		assert.NoError(t, err)
		assert.Equal(t, OrderActionComplete, action)
	})
}

// TestChainHandlerContextHandling tests context handling
func TestChainHandlerContextHandling(t *testing.T) {
	t.Run("canceled_context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		handler := &mockChainHandler{}

		action, err := handler.Fill(ctx, &types.ParsedArgs{})
		// The mock handler doesn't check context, but real handlers should
		assert.NoError(t, err)
		assert.Equal(t, OrderActionComplete, action)
	})

	t.Run("timeout_context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()

		handler := &mockChainHandler{}

		action, err := handler.Fill(ctx, &types.ParsedArgs{})
		// The mock handler doesn't check context, but real handlers should
		assert.NoError(t, err)
		assert.Equal(t, OrderActionComplete, action)
	})
}

// TestChainHandlerParsedArgs tests handling of different ParsedArgs
func TestChainHandlerParsedArgs(t *testing.T) {
	t.Run("empty_parsed_args", func(t *testing.T) {
		handler := &mockChainHandler{}

		action, err := handler.Fill(context.Background(), &types.ParsedArgs{})
		require.NoError(t, err)
		assert.Equal(t, OrderActionComplete, action)
	})

	t.Run("parsed_args_with_order_id", func(t *testing.T) {
		handler := &mockChainHandler{}

		args := types.ParsedArgs{
			OrderID: "test-order-456",
		}

		action, err := handler.Fill(context.Background(), &args)
		require.NoError(t, err)
		assert.Equal(t, OrderActionComplete, action)
	})

	t.Run("parsed_args_with_sender_address", func(t *testing.T) {
		handler := &mockChainHandler{}

		args := types.ParsedArgs{
			SenderAddress: "0x1234567890123456789012345678901234567890",
		}

		action, err := handler.Fill(context.Background(), &args)
		require.NoError(t, err)
		assert.Equal(t, OrderActionComplete, action)
	})
}

// Mock implementations for testing

// mockChainHandler implements ChainHandler for basic testing
type mockChainHandler struct{}

func (m *mockChainHandler) Fill(ctx context.Context, args *types.ParsedArgs) (OrderAction, error) {
	return OrderActionComplete, nil
}

func (m *mockChainHandler) Settle(ctx context.Context, args *types.ParsedArgs) error {
	return nil
}

// mockChainHandlerWithAction implements ChainHandler with configurable action
type mockChainHandlerWithAction struct {
	action OrderAction
}

func (m *mockChainHandlerWithAction) Fill(ctx context.Context, args *types.ParsedArgs) (OrderAction, error) {
	return m.action, nil
}

func (m *mockChainHandlerWithAction) Settle(ctx context.Context, args *types.ParsedArgs) error {
	return nil
}

// mockChainHandlerWithError implements ChainHandler with configurable error
type mockChainHandlerWithError struct {
	shouldError bool
}

func (m *mockChainHandlerWithError) Fill(ctx context.Context, args *types.ParsedArgs) (OrderAction, error) {
	if m.shouldError {
		return OrderActionError, assert.AnError
	}
	return OrderActionComplete, nil
}

func (m *mockChainHandlerWithError) Settle(ctx context.Context, args *types.ParsedArgs) error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}
