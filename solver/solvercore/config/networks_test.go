package config

import (
	"os"
	"testing"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/stretchr/testify/assert"
)

func TestConditionalEnvironment(t *testing.T) {
	t.Run("IS_DEVNET=true uses LOCAL_ variables", func(t *testing.T) {
		// Set up test environment
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_ETHEREUM_RPC_URL", "http://localhost:8545")
		t.Setenv("ETHEREUM_RPC_URL", "https://eth-sepolia.g.alchemy.com/v2/test")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_ETHEREUM_RPC_URL")
			os.Unsetenv("ETHEREUM_RPC_URL")
		}()

		// Test that LOCAL_ version is used when IS_DEVNET=true
		result := envutil.GetConditionalEnv("ETHEREUM_RPC_URL", "default")
		assert.Equal(t, "http://localhost:8545", result)
	})

	t.Run("IS_DEVNET=false uses regular variables", func(t *testing.T) {
		// Set up test environment
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("LOCAL_ETHEREUM_RPC_URL", "http://localhost:8545")
		t.Setenv("ETHEREUM_RPC_URL", "https://eth-sepolia.g.alchemy.com/v2/test")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_ETHEREUM_RPC_URL")
			os.Unsetenv("ETHEREUM_RPC_URL")
		}()

		// Test that regular version is used when IS_DEVNET=false
		result := envutil.GetConditionalEnv("ETHEREUM_RPC_URL", "default")
		assert.Equal(t, "https://eth-sepolia.g.alchemy.com/v2/test", result)
	})

	t.Run("IS_DEVNET unset defaults to regular variables", func(t *testing.T) {
		// Set up test environment
		os.Unsetenv("IS_DEVNET")
		t.Setenv("LOCAL_ETHEREUM_RPC_URL", "http://localhost:8545")
		t.Setenv("ETHEREUM_RPC_URL", "https://eth-sepolia.g.alchemy.com/v2/test")
		defer func() {
			os.Unsetenv("LOCAL_ETHEREUM_RPC_URL")
			os.Unsetenv("ETHEREUM_RPC_URL")
		}()

		// Test that regular version is used when IS_DEVNET is unset
		result := envutil.GetConditionalEnv("ETHEREUM_RPC_URL", "default")
		assert.Equal(t, "https://eth-sepolia.g.alchemy.com/v2/test", result)
	})

	t.Run("Missing variables fall back to default", func(t *testing.T) {
		// Set up test environment
		t.Setenv("IS_DEVNET", "true")
		defer os.Unsetenv("IS_DEVNET")

		// Test that default is used when neither LOCAL_ nor regular version exists
		result := envutil.GetConditionalEnv("NONEXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})
}

func TestGetConditionalAccountEnv(t *testing.T) {
	t.Run("Account variables work with conditional logic", func(t *testing.T) {
		// Set up test environment
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_SOLVER_PUB_KEY", "0x1234567890123456789012345678901234567890")
		t.Setenv("SOLVER_PUB_KEY", "0x0987654321098765432109876543210987654321")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_SOLVER_PUB_KEY")
			os.Unsetenv("SOLVER_PUB_KEY")
		}()

		// Test that LOCAL_ version is used when IS_DEVNET=true
		result := GetConditionalAccountEnv("SOLVER_PUB_KEY")
		assert.Equal(t, "0x1234567890123456789012345678901234567890", result)
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("getEnvUint64", func(t *testing.T) {
		// Set up environment
		t.Setenv("TEST_UINT", "123")
		defer os.Unsetenv("TEST_UINT")

		result := envutil.GetEnvUint64("TEST_UINT", 456)
		assert.Equal(t, uint64(123), result)
	})

	t.Run("getEnvUint64 with missing key", func(t *testing.T) {
		result := envutil.GetEnvUint64("MISSING_UINT", 456)
		assert.Equal(t, uint64(456), result)
	})

	t.Run("getEnvUint64 with invalid value", func(t *testing.T) {
		// Set up environment
		t.Setenv("INVALID_UINT", "not_a_number")
		defer os.Unsetenv("INVALID_UINT")

		result := envutil.GetEnvUint64("INVALID_UINT", 456)
		assert.Equal(t, uint64(456), result)
	})

	t.Run("getEnvUint64Any", func(t *testing.T) {
		// Set up environment
		t.Setenv("KEY1", "100")
		t.Setenv("KEY2", "200")
		defer func() {
			os.Unsetenv("KEY1")
			os.Unsetenv("KEY2")
		}()

		result := envutil.GetEnvUint64Any([]string{"KEY1", "KEY2"}, 300)
		assert.Equal(t, uint64(100), result) // Should return first valid value
	})

	t.Run("getEnvUint64Any with no valid keys", func(t *testing.T) {
		result := envutil.GetEnvUint64Any([]string{"MISSING1", "MISSING2"}, 300)
		assert.Equal(t, uint64(300), result)
	})

	t.Run("getEnvInt", func(t *testing.T) {
		// Set up environment
		t.Setenv("TEST_INT", "123")
		defer os.Unsetenv("TEST_INT")

		result := envutil.GetEnvInt("TEST_INT", 456)
		assert.Equal(t, 123, result)
	})

	t.Run("getEnvInt with missing key", func(t *testing.T) {
		result := envutil.GetEnvInt("MISSING_INT", 456)
		assert.Equal(t, 456, result)
	})

	t.Run("getEnvInt with invalid value", func(t *testing.T) {
		// Set up environment
		t.Setenv("INVALID_INT", "not_a_number")
		defer os.Unsetenv("INVALID_INT")

		result := envutil.GetEnvInt("INVALID_INT", 456)
		assert.Equal(t, 456, result)
	})

	// Note: parseUint64 is now internal to envutil package, so we test it indirectly
	// through the public functions that use it
}
