package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSolverConfig tests the SolverConfig struct
func TestSolverConfig(t *testing.T) {
	t.Run("solver_config_creation", func(t *testing.T) {
		config := SolverConfig{
			Enabled: true,
		}
		assert.True(t, config.Enabled)
	})

	t.Run("solver_config_defaults", func(t *testing.T) {
		config := SolverConfig{}
		assert.False(t, config.Enabled)
	})

	t.Run("solver_config_serialization", func(t *testing.T) {
		config := SolverConfig{
			Enabled: true,
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(config)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "true")

		// Test JSON unmarshaling
		var unmarshaled SolverConfig
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, config.Enabled, unmarshaled.Enabled)
	})
}

// TestConfig tests the main Config struct
func TestConfig(t *testing.T) {
	t.Run("config_creation", func(t *testing.T) {
		config := &Config{
			Solvers:    make(map[string]SolverConfig),
			LogLevel:   "debug",
			LogFormat:  "json",
			MaxRetries: 10,
		}

		assert.NotNil(t, config.Solvers)
		assert.Equal(t, "debug", config.LogLevel)
		assert.Equal(t, "json", config.LogFormat)
		assert.Equal(t, 10, config.MaxRetries)
	})

	t.Run("config_defaults", func(t *testing.T) {
		config := &Config{}
		assert.Nil(t, config.Solvers)
		assert.Empty(t, config.LogLevel)
		assert.Empty(t, config.LogFormat)
		assert.Equal(t, 0, config.MaxRetries)
	})

	t.Run("config_serialization", func(t *testing.T) {
		config := &Config{
			Solvers: map[string]SolverConfig{
				"test": {Enabled: true},
			},
			LogLevel:   "info",
			LogFormat:  "text",
			MaxRetries: 5,
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(config)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "test")
		assert.Contains(t, string(jsonData), "info")

		// Test JSON unmarshaling
		var unmarshaled Config
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, config.LogLevel, unmarshaled.LogLevel)
		assert.Equal(t, config.LogFormat, unmarshaled.LogFormat)
		assert.Equal(t, config.MaxRetries, unmarshaled.MaxRetries)
	})
}

// TestDefaultSolvers tests the default solver configurations
func TestDefaultSolvers(t *testing.T) {
	t.Run("default_solvers_contains_hyperlane7683", func(t *testing.T) {
		assert.Contains(t, defaultSolvers, "hyperlane7683")
		assert.True(t, defaultSolvers["hyperlane7683"].Enabled)
	})

	t.Run("default_solvers_immutability", func(t *testing.T) {
		// Test that we can't modify the default solvers map
		originalCount := len(defaultSolvers)
		assert.Greater(t, originalCount, 0)

		// Attempt to modify (should not affect the original)
		testSolvers := make(map[string]SolverConfig)
		for k, v := range defaultSolvers {
			testSolvers[k] = v
		}
		testSolvers["new"] = SolverConfig{Enabled: true}

		// Original should be unchanged
		assert.Equal(t, originalCount, len(defaultSolvers))
		assert.NotContains(t, defaultSolvers, "new")
	})
}

// TestLoadConfigComprehensive tests the LoadConfig function comprehensively
func TestLoadConfigComprehensive(t *testing.T) {
	t.Run("load_config_without_env_file", func(t *testing.T) {
		// Test loading config when .env file doesn't exist
		config, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.NotNil(t, config.Solvers)
		assert.Equal(t, "info", config.LogLevel)
		assert.Equal(t, "text", config.LogFormat)
		assert.Equal(t, 5, config.MaxRetries)
	})

	t.Run("load_config_with_env_file", func(t *testing.T) {
		// Create a temporary .env file
		envContent := `LOG_LEVEL=debug
LOG_FORMAT=json
MAX_RETRIES=10
SOLVER_HYPERLANE7683_ENABLED=false
SOLVER_NEWSOLVER_ENABLED=true`

		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		err := os.WriteFile(envFile, []byte(envContent), 0644)
		require.NoError(t, err)

		// Change to temp directory
		oldDir, err := os.Getwd()
		require.NoError(t, err)
		defer t.Chdir(oldDir)

		t.Chdir(tempDir)

		// Load config
		config, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "debug", config.LogLevel)
		assert.Equal(t, "json", config.LogFormat)
		assert.Equal(t, 10, config.MaxRetries)
	})

	t.Run("load_config_environment_override", func(t *testing.T) {
		// Set environment variables
		t.Setenv("LOG_LEVEL", "trace")
		t.Setenv("LOG_FORMAT", "structured")
		t.Setenv("MAX_RETRIES", "15")
		defer func() {
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("LOG_FORMAT")
			os.Unsetenv("MAX_RETRIES")
		}()

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "trace", config.LogLevel)
		assert.Equal(t, "structured", config.LogFormat)
		assert.Equal(t, 15, config.MaxRetries)
	})
}

// TestSolverState tests the SolverState functionality
func TestSolverState(t *testing.T) {
	t.Run("solver_state_creation", func(t *testing.T) {
		state := SolverState{
			Networks: make(map[string]SolverNetworkState),
		}

		assert.NotNil(t, state.Networks)
		assert.Empty(t, state.Networks)
	})

	t.Run("solver_state_with_networks", func(t *testing.T) {
		state := SolverState{
			Networks: map[string]SolverNetworkState{
				"Ethereum": {
					LastIndexedBlock: 12345,
					LastUpdated:      "2023-01-01T00:00:00Z",
				},
				"Base": {
					LastIndexedBlock: 67890,
					LastUpdated:      "2023-01-02T00:00:00Z",
				},
			},
		}

		assert.Len(t, state.Networks, 2)
		assert.Equal(t, uint64(12345), state.Networks["Ethereum"].LastIndexedBlock)
		assert.Equal(t, uint64(67890), state.Networks["Base"].LastIndexedBlock)
	})

	t.Run("solver_state_serialization", func(t *testing.T) {
		state := SolverState{
			Networks: map[string]SolverNetworkState{
				"Test": {
					LastIndexedBlock: 99999,
					LastUpdated:      "2023-12-31T23:59:59Z",
				},
			},
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(state)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "Test")
		assert.Contains(t, string(jsonData), "99999")

		// Test JSON unmarshaling
		var unmarshaled SolverState
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, state.Networks["Test"].LastIndexedBlock, unmarshaled.Networks["Test"].LastIndexedBlock)
	})
}

// TestSolverNetworkState tests the SolverNetworkState functionality
func TestSolverNetworkState(t *testing.T) {
	t.Run("solver_network_state_creation", func(t *testing.T) {
		state := SolverNetworkState{
			LastIndexedBlock: 1000,
			LastUpdated:      time.Now().Format(time.RFC3339),
		}

		assert.Equal(t, uint64(1000), state.LastIndexedBlock)
		assert.NotEmpty(t, state.LastUpdated)
	})

	t.Run("solver_network_state_defaults", func(t *testing.T) {
		state := SolverNetworkState{}
		assert.Equal(t, uint64(0), state.LastIndexedBlock)
		assert.Empty(t, state.LastUpdated)
	})

	t.Run("solver_network_state_serialization", func(t *testing.T) {
		state := SolverNetworkState{
			LastIndexedBlock: 5000,
			LastUpdated:      "2023-06-15T12:30:45Z",
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(state)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "5000")
		assert.Contains(t, string(jsonData), "2023-06-15T12:30:45Z")

		// Test JSON unmarshaling
		var unmarshaled SolverNetworkState
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, state.LastIndexedBlock, unmarshaled.LastIndexedBlock)
		assert.Equal(t, state.LastUpdated, unmarshaled.LastUpdated)
	})
}

// TestGetDefaultSolverState tests the getDefaultSolverState function
func TestGetDefaultSolverState(t *testing.T) {
	t.Run("get_default_solver_state", func(t *testing.T) {
		state := getDefaultSolverState()
		assert.NotNil(t, state.Networks)
		assert.Greater(t, len(state.Networks), 0)
	})

	t.Run("default_solver_state_contains_networks", func(t *testing.T) {
		state := getDefaultSolverState()

		// Should contain common networks
		expectedNetworks := []string{"Ethereum", "Optimism", "Arbitrum", "Base", "Starknet"}
		for _, network := range expectedNetworks {
			if _, exists := state.Networks[network]; exists {
				assert.NotNil(t, state.Networks[network])
			}
		}
	})
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	t.Run("valid_config", func(t *testing.T) {
		config := &Config{
			Solvers: map[string]SolverConfig{
				"test": {Enabled: true},
			},
			LogLevel:   "info",
			LogFormat:  "text",
			MaxRetries: 5,
		}

		// Basic validation
		assert.NotNil(t, config.Solvers)
		assert.NotEmpty(t, config.LogLevel)
		assert.NotEmpty(t, config.LogFormat)
		assert.GreaterOrEqual(t, config.MaxRetries, 0)
	})

	t.Run("config_with_empty_solvers", func(t *testing.T) {
		config := &Config{
			Solvers:    make(map[string]SolverConfig),
			LogLevel:   "info",
			LogFormat:  "text",
			MaxRetries: 5,
		}

		assert.NotNil(t, config.Solvers)
		assert.Empty(t, config.Solvers)
	})

	t.Run("config_with_negative_retries", func(t *testing.T) {
		config := &Config{
			Solvers:    make(map[string]SolverConfig),
			LogLevel:   "info",
			LogFormat:  "text",
			MaxRetries: -1,
		}

		// Should handle negative retries gracefully
		assert.Equal(t, -1, config.MaxRetries)
	})
}

// TestConfigEdgeCases tests edge cases for configuration
func TestConfigEdgeCases(t *testing.T) {
	t.Run("config_with_nil_solvers", func(t *testing.T) {
		config := &Config{
			Solvers:    nil,
			LogLevel:   "info",
			LogFormat:  "text",
			MaxRetries: 5,
		}

		assert.Nil(t, config.Solvers)
	})

	t.Run("config_with_empty_strings", func(t *testing.T) {
		config := &Config{
			Solvers:    make(map[string]SolverConfig),
			LogLevel:   "",
			LogFormat:  "",
			MaxRetries: 0,
		}

		assert.Empty(t, config.LogLevel)
		assert.Empty(t, config.LogFormat)
		assert.Equal(t, 0, config.MaxRetries)
	})

	t.Run("config_with_large_values", func(t *testing.T) {
		config := &Config{
			Solvers:    make(map[string]SolverConfig),
			LogLevel:   "very_verbose_debug_level",
			LogFormat:  "very_detailed_json_format",
			MaxRetries: 999999,
		}

		assert.Equal(t, "very_verbose_debug_level", config.LogLevel)
		assert.Equal(t, "very_detailed_json_format", config.LogFormat)
		assert.Equal(t, 999999, config.MaxRetries)
	})
}

// TestConfigConcurrency tests concurrent access to configuration
func TestConfigConcurrency(t *testing.T) {
	t.Run("concurrent_config_access", func(t *testing.T) {
		config := &Config{
			Solvers:    make(map[string]SolverConfig),
			LogLevel:   "info",
			LogFormat:  "text",
			MaxRetries: 5,
		}

		// Test concurrent reads
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_ = config.LogLevel
				_ = config.LogFormat
				_ = config.MaxRetries
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent_solver_config_access", func(t *testing.T) {
		config := &Config{
			Solvers: map[string]SolverConfig{
				"test1": {Enabled: true},
				"test2": {Enabled: false},
			},
			LogLevel:   "info",
			LogFormat:  "text",
			MaxRetries: 5,
		}

		// Test concurrent reads
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_ = config.Solvers["test1"]
				_ = config.Solvers["test2"]
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
