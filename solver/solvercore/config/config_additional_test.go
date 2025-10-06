package config

import (
	"os"
	"testing"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("LoadConfig with valid environment", func(t *testing.T) {
		// Set up test environment variables
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		config, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify that config was loaded successfully
		assert.NotNil(t, config)
	})

	t.Run("LoadConfig with missing environment variables", func(t *testing.T) {
		// Clear all environment variables
		os.Clearenv()

		config, err := LoadConfig()
		// Should not error, but may have empty networks
		assert.NoError(t, err)
		assert.NotNil(t, config)
	})
}

func TestIsSolverEnabled(t *testing.T) {
	t.Run("Solver enabled", func(t *testing.T) {
		config := &Config{
			Solvers: map[string]SolverConfig{
				"hyperlane7683": {Enabled: true},
			},
		}

		enabled := config.IsSolverEnabled("hyperlane7683")
		assert.True(t, enabled)
	})

	t.Run("Solver disabled", func(t *testing.T) {
		config := &Config{
			Solvers: map[string]SolverConfig{
				"hyperlane7683": {Enabled: false},
			},
		}

		enabled := config.IsSolverEnabled("hyperlane7683")
		assert.False(t, enabled)
	})

	t.Run("Solver not found", func(t *testing.T) {
		config := &Config{
			Solvers: map[string]SolverConfig{},
		}

		enabled := config.IsSolverEnabled("nonexistent")
		assert.False(t, enabled)
	})
}

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("Get existing environment variable", func(t *testing.T) {
		t.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		value := envutil.GetEnvWithDefault("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", value)
	})

	t.Run("Get non-existing environment variable", func(t *testing.T) {
		os.Unsetenv("NON_EXISTING_VAR")

		value := envutil.GetEnvWithDefault("NON_EXISTING_VAR", "default_value")
		assert.Equal(t, "default_value", value)
	})
}

func TestGetConditionalUint64(t *testing.T) {
	t.Run("Get conditional uint64 with DEVNET=true", func(t *testing.T) {
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_TEST_VAR", "123")
		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_TEST_VAR")
		}()

		value := envutil.GetConditionalUint64("TEST_VAR", 456, 789)
		assert.Equal(t, uint64(123), value)
	})

	t.Run("Get conditional uint64 with DEVNET=false", func(t *testing.T) {
		t.Setenv("DEVNET", "false")
		t.Setenv("TEST_VAR", "789")
		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("TEST_VAR")
		}()

		value := envutil.GetConditionalUint64("TEST_VAR", 456, 789)
		assert.Equal(t, uint64(789), value)
	})

	t.Run("Get conditional uint64 with invalid value", func(t *testing.T) {
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_TEST_VAR", "invalid")
		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_TEST_VAR")
		}()

		value := envutil.GetConditionalUint64("TEST_VAR", 456, 789)
		assert.Equal(t, uint64(789), value) // Should return local default
	})
}

func TestInitializeNetworks(t *testing.T) {
	t.Run("Initialize networks", func(t *testing.T) {
		// Set up test environment
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		// Verify that networks were initialized
		// Note: networks is a package-level variable, so we can't directly test it
		// But we can test that InitializeNetworks() doesn't panic
		assert.NotPanics(t, func() {
			InitializeNetworks()
		})
	})
}

func TestGetNetworkConfig(t *testing.T) {
	t.Run("Get existing network config", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		config, err := GetNetworkConfig("Base")
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, "Base", config.Name)
		assert.Equal(t, "http://localhost:8545", config.RPCURL)
		assert.Equal(t, uint64(84532), config.ChainID)
	})

	t.Run("Get non-existing network config", func(t *testing.T) {
		config, err := GetNetworkConfig("non_existing")
		assert.Error(t, err)
		assert.Equal(t, NetworkConfig{}, config)
	})
}

func TestGetRPCURL(t *testing.T) {
	t.Run("Get RPC URL for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		url, err := GetRPCURL("Base")
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8545", url)
	})

	t.Run("Get RPC URL for non-existing network", func(t *testing.T) {
		url, err := GetRPCURL("non_existing")
		assert.Error(t, err)
		assert.Empty(t, url)
	})
}

func TestGetChainID(t *testing.T) {
	t.Run("Get Chain ID for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		chainID, err := GetChainID("Base")
		require.NoError(t, err)
		assert.Equal(t, uint64(84532), chainID)
	})

	t.Run("Get Chain ID for non-existing network", func(t *testing.T) {
		chainID, err := GetChainID("non_existing")
		assert.Error(t, err)
		assert.Equal(t, uint64(0), chainID)
	})
}

func TestGetHyperlaneAddress(t *testing.T) {
	t.Run("Get Hyperlane address for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		address, err := GetHyperlaneAddress("Base")
		require.NoError(t, err)
		assert.Equal(t, "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3", address.String())
	})

	t.Run("Get Hyperlane address for non-existing network", func(t *testing.T) {
		address, err := GetHyperlaneAddress("non_existing")
		assert.Error(t, err)
		assert.Empty(t, address)
	})
}

func TestGetHyperlaneDomain(t *testing.T) {
	t.Run("Get Hyperlane domain for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		domain, err := GetHyperlaneDomain("Base")
		require.NoError(t, err)
		assert.Equal(t, uint64(84532), domain)
	})

	t.Run("Get Hyperlane domain for non-existing network", func(t *testing.T) {
		domain, err := GetHyperlaneDomain("non_existing")
		assert.Error(t, err)
		assert.Equal(t, uint64(0), domain)
	})
}

func TestGetForkStartBlock(t *testing.T) {
	t.Run("Get fork start block for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		block, err := GetForkStartBlock("Base")
		require.NoError(t, err)
		assert.Equal(t, uint64(1000), block)
	})

	t.Run("Get fork start block for non-existing network", func(t *testing.T) {
		block, err := GetForkStartBlock("non_existing")
		assert.Error(t, err)
		assert.Equal(t, uint64(0), block)
	})
}

func TestGetSolverStartBlock(t *testing.T) {
	t.Run("Get solver start block for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		block, err := GetSolverStartBlock("Base")
		require.NoError(t, err)
		assert.Equal(t, int64(1000), block)
	})

	t.Run("Get solver start block for non-existing network", func(t *testing.T) {
		block, err := GetSolverStartBlock("non_existing")
		assert.Error(t, err)
		assert.Equal(t, int64(0), block)
	})
}

func TestGetListenerConfig(t *testing.T) {
	t.Run("Get listener config for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		pollInterval, confirmationBlocks, maxBlockRange, err := GetListenerConfig("Base")
		require.NoError(t, err)
		assert.Equal(t, 1000, pollInterval)
		assert.Equal(t, uint64(0), confirmationBlocks)
		assert.Equal(t, uint64(10), maxBlockRange)
	})

	t.Run("Get listener config for non-existing network", func(t *testing.T) {
		pollInterval, confirmationBlocks, maxBlockRange, err := GetListenerConfig("non_existing")
		assert.Error(t, err)
		assert.Equal(t, 0, pollInterval)
		assert.Equal(t, uint64(0), confirmationBlocks)
		assert.Equal(t, uint64(0), maxBlockRange)
	})
}

func TestGetRPCURLByChainID(t *testing.T) {
	t.Run("Get RPC URL by chain ID for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		url, err := GetRPCURLByChainID(84532)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8545", url)
	})

	t.Run("Get RPC URL by chain ID for non-existing network", func(t *testing.T) {
		url, err := GetRPCURLByChainID(99999)
		assert.Error(t, err)
		assert.Empty(t, url)
	})
}

func TestGetHyperlaneAddressByChainID(t *testing.T) {
	t.Run("Get Hyperlane address by chain ID for existing network", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		address, err := GetHyperlaneAddressByChainID(84532)
		require.NoError(t, err)
		assert.Equal(t, "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3", address.String())
	})

	t.Run("Get Hyperlane address by chain ID for non-existing network", func(t *testing.T) {
		address, err := GetHyperlaneAddressByChainID(99999)
		assert.Error(t, err)
		assert.Empty(t, address)
	})
}

func TestGetNetworkNames(t *testing.T) {
	t.Run("Get network names", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		names := GetNetworkNames()
		assert.NotEmpty(t, names)
		assert.Contains(t, names, "Base")
	})
}

func TestValidateNetworkName(t *testing.T) {
	t.Run("Validate existing network name", func(t *testing.T) {
		// Initialize networks first
		t.Setenv("DEVNET", "true")
		t.Setenv("LOCAL_BASE_RPC_URL", "http://localhost:8545")
		t.Setenv("LOCAL_BASE_CHAIN_ID", "84532")
		t.Setenv("LOCAL_BASE_HYPERLANE_ADDRESS", "0x1234567890123456789012345678901234567890")
		t.Setenv("LOCAL_BASE_HYPERLANE_DOMAIN", "84532")
		t.Setenv("LOCAL_BASE_FORK_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_SOLVER_START_BLOCK", "1000")
		t.Setenv("LOCAL_BASE_CONFIRMATION_BLOCKS", "1")
		t.Setenv("LOCAL_BASE_POLL_INTERVAL", "1000")
		t.Setenv("LOCAL_BASE_MAX_BLOCK_RANGE", "1000")

		defer func() {
			os.Unsetenv("DEVNET")
			os.Unsetenv("LOCAL_BASE_RPC_URL")
			os.Unsetenv("LOCAL_BASE_CHAIN_ID")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_ADDRESS")
			os.Unsetenv("LOCAL_BASE_HYPERLANE_DOMAIN")
			os.Unsetenv("LOCAL_BASE_FORK_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_SOLVER_START_BLOCK")
			os.Unsetenv("LOCAL_BASE_CONFIRMATION_BLOCKS")
			os.Unsetenv("LOCAL_BASE_POLL_INTERVAL")
			os.Unsetenv("LOCAL_BASE_MAX_BLOCK_RANGE")
		}()

		InitializeNetworks()

		valid := ValidateNetworkName("Base")
		assert.True(t, valid)
	})

	t.Run("Validate non-existing network name", func(t *testing.T) {
		valid := ValidateNetworkName("non_existing")
		assert.False(t, valid)
	})
}

func TestGetDefaultNetwork(t *testing.T) {
	t.Run("Get default network", func(t *testing.T) {
		network := GetDefaultNetwork()
		assert.NotEmpty(t, network)
	})
}

func TestGetDefaultRPCURL(t *testing.T) {
	t.Run("Get default RPC URL", func(t *testing.T) {
		url := GetDefaultRPCURL()
		assert.NotEmpty(t, url)
	})
}
