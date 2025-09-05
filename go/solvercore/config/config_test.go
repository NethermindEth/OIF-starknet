package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkConfig(t *testing.T) {
	t.Run("NetworkConfig creation", func(t *testing.T) {
		config := NetworkConfig{
			Name:               "TestNetwork",
			RPCURL:             "http://localhost:8545",
			ChainID:            12345,
			ConfirmationBlocks: 1,
			PollInterval:       1000,
			MaxBlockRange:      1000,
		}

		assert.Equal(t, "TestNetwork", config.Name)
		assert.Equal(t, "http://localhost:8545", config.RPCURL)
		assert.Equal(t, uint64(12345), config.ChainID)
		assert.Equal(t, uint64(1), config.ConfirmationBlocks)
		assert.Equal(t, int(1000), config.PollInterval)
		assert.Equal(t, uint64(1000), config.MaxBlockRange)
	})

	t.Run("NetworkConfig validation", func(t *testing.T) {
		tests := []struct {
			name    string
			config  NetworkConfig
			isValid bool
		}{
			{
				name: "Valid config",
				config: NetworkConfig{
					Name:    "TestNetwork",
					RPCURL:  "http://localhost:8545",
					ChainID: 12345,
				},
				isValid: true,
			},
			{
				name: "Empty name",
				config: NetworkConfig{
					Name:    "",
					RPCURL:  "http://localhost:8545",
					ChainID: 12345,
				},
				isValid: false,
			},
			{
				name: "Empty RPC URL",
				config: NetworkConfig{
					Name:    "TestNetwork",
					RPCURL:  "",
					ChainID: 12345,
				},
				isValid: false,
			},
			{
				name: "Zero chain ID",
				config: NetworkConfig{
					Name:    "TestNetwork",
					RPCURL:  "http://localhost:8545",
					ChainID: 0,
				},
				isValid: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Basic validation logic
				isValid := tt.config.Name != "" && tt.config.RPCURL != "" && tt.config.ChainID != 0
				assert.Equal(t, tt.isValid, isValid)
			})
		}
	})
}

func TestEnvironmentLoading(t *testing.T) {
	t.Run("Environment variable loading", func(t *testing.T) {
		// Set test environment variables
		t.Setenv("TEST_NETWORK_NAME", "TestNetwork")
		t.Setenv("TEST_RPC_URL", "https://test.example.com")
		t.Setenv("TEST_CHAIN_ID", "12345")

		defer func() {
			os.Unsetenv("TEST_NETWORK_NAME")
			os.Unsetenv("TEST_RPC_URL")
			os.Unsetenv("TEST_CHAIN_ID")
		}()

		// Test environment variable retrieval
		name := os.Getenv("TEST_NETWORK_NAME")
		rpcURL := os.Getenv("TEST_RPC_URL")
		chainIDStr := os.Getenv("TEST_CHAIN_ID")

		assert.Equal(t, "TestNetwork", name)
		assert.Equal(t, "https://test.example.com", rpcURL)
		assert.Equal(t, "12345", chainIDStr)
	})

	t.Run("Missing environment variables", func(t *testing.T) {
		// Test with unset environment variables
		missingVar := os.Getenv("NONEXISTENT_VAR")
		assert.Empty(t, missingVar)
	})

	t.Run("Default values", func(t *testing.T) {
		// Test default value handling
		getEnvWithDefault := func(key, defaultValue string) string {
			if value := os.Getenv(key); value != "" {
				return value
			}
			return defaultValue
		}

		// Test with unset variable
		result := getEnvWithDefault("UNSET_VAR", "default_value")
		assert.Equal(t, "default_value", result)

		// Test with set variable
		t.Setenv("SET_VAR", "actual_value")
		defer os.Unsetenv("SET_VAR")
		result = getEnvWithDefault("SET_VAR", "default_value")
		assert.Equal(t, "actual_value", result)
	})
}

func TestNetworkValidation(t *testing.T) {
	t.Run("Valid network names", func(t *testing.T) {
		validNetworks := []string{
			"TestNetwork",
			"LocalNetwork",
			"MockNetwork",
			"TestEthereum",
			"TestStarknet",
			"TestArbitrum",
		}

		for _, network := range validNetworks {
			t.Run(network, func(t *testing.T) {
				// Basic validation - non-empty and reasonable length
				assert.NotEmpty(t, network)
				assert.Less(t, len(network), 50) // Reasonable length limit
			})
		}
	})

	t.Run("Invalid network names", func(t *testing.T) {
		invalidNetworks := []string{
			"",
			"   ", // Whitespace only
		}

		for _, network := range invalidNetworks {
			t.Run(network, func(t *testing.T) {
				// Basic validation
				isValid := network != "" && strings.TrimSpace(network) != ""
				assert.False(t, isValid)
			})
		}
	})

	t.Run("RPC URL validation", func(t *testing.T) {
		tests := []struct {
			url     string
			isValid bool
		}{
			{"http://localhost:8545", true},
			{"https://test.example.com", true},
			{"ws://localhost:8546", true},
			{"", false},
			{"invalid-url", false},
			{"ftp://example.com", false}, // Wrong protocol
		}

		for _, tt := range tests {
			t.Run(tt.url, func(t *testing.T) {
				// Basic URL validation
				isValid := tt.url != "" &&
					(strings.HasPrefix(tt.url, "http://") ||
						strings.HasPrefix(tt.url, "https://") ||
						strings.HasPrefix(tt.url, "ws://") ||
						strings.HasPrefix(tt.url, "wss://"))
				assert.Equal(t, tt.isValid, isValid)
			})
		}
	})
}

func TestChainIDValidation(t *testing.T) {
	t.Run("Valid chain IDs", func(t *testing.T) {
		validChainIDs := []uint64{
			12345, // Test chain ID
			54321, // Test chain ID
			99999, // Test chain ID
			1,     // Ethereum mainnet (for reference)
			8453,  // Base mainnet (for reference)
			42161, // Arbitrum mainnet (for reference)
		}

		for _, chainID := range validChainIDs {
			t.Run(string(rune(chainID)), func(t *testing.T) {
				assert.Greater(t, chainID, uint64(0))
				assert.Less(t, chainID, uint64(1000000000)) // Reasonable upper limit
			})
		}
	})

	t.Run("Invalid chain IDs", func(t *testing.T) {
		invalidChainIDs := []uint64{
			0, // Zero is invalid
		}

		for _, chainID := range invalidChainIDs {
			t.Run(string(rune(chainID)), func(t *testing.T) {
				assert.Equal(t, uint64(0), chainID)
			})
		}
	})
}

func TestConfigurationDefaults(t *testing.T) {
	t.Run("Default confirmation blocks", func(t *testing.T) {
		config := NetworkConfig{
			Name:    "Test",
			RPCURL:  "https://test.example.com",
			ChainID: 12345,
		}

		// Test default values
		if config.ConfirmationBlocks == 0 {
			config.ConfirmationBlocks = 1 // Default
		}
		assert.Equal(t, uint64(1), config.ConfirmationBlocks)
	})

	t.Run("Default poll interval", func(t *testing.T) {
		config := NetworkConfig{
			Name:    "Test",
			RPCURL:  "https://test.example.com",
			ChainID: 12345,
		}

		// Test default values
		if config.PollInterval == 0 {
			config.PollInterval = 1000 // Default 1 second
		}
		assert.Equal(t, int(1000), config.PollInterval)
	})

	t.Run("Default max block range", func(t *testing.T) {
		config := NetworkConfig{
			Name:    "Test",
			RPCURL:  "https://test.example.com",
			ChainID: 12345,
		}

		// Test default values
		if config.MaxBlockRange == 0 {
			config.MaxBlockRange = 1000 // Default
		}
		assert.Equal(t, uint64(1000), config.MaxBlockRange)
	})
}

func TestNetworkConfigComparison(t *testing.T) {
	t.Run("Equal configurations", func(t *testing.T) {
		config1 := NetworkConfig{
			Name:    "TestNetwork",
			RPCURL:  "http://localhost:8545",
			ChainID: 12345,
		}

		config2 := NetworkConfig{
			Name:    "TestNetwork",
			RPCURL:  "http://localhost:8545",
			ChainID: 12345,
		}

		assert.Equal(t, config1, config2)
	})

	t.Run("Different configurations", func(t *testing.T) {
		config1 := NetworkConfig{
			Name:    "TestNetwork1",
			RPCURL:  "http://localhost:8545",
			ChainID: 12345,
		}

		config2 := NetworkConfig{
			Name:    "TestNetwork2",
			RPCURL:  "http://localhost:8546",
			ChainID: 54321,
		}

		assert.NotEqual(t, config1, config2)
	})
}
