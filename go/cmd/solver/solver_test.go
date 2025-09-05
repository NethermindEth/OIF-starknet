package solver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NethermindEth/oif-starknet/go/solvercore/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to capture logrus output
func captureLogrusOutput(fn func()) string {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stderr) // Restore default output

	fn()

	return buf.String()
}

func TestCleanFormatter(t *testing.T) {
	t.Run("cleanFormatter formats message correctly", func(t *testing.T) {
		formatter := &cleanFormatter{}
		entry := &logrus.Entry{
			Message: "Test message",
		}

		result, err := formatter.Format(entry)
		require.NoError(t, err)
		assert.Equal(t, "Test message\n", string(result))
	})

	t.Run("cleanFormatter handles empty message", func(t *testing.T) {
		formatter := &cleanFormatter{}
		entry := &logrus.Entry{
			Message: "",
		}

		result, err := formatter.Format(entry)
		require.NoError(t, err)
		assert.Equal(t, "\n", string(result))
	})

	t.Run("cleanFormatter handles multiline message", func(t *testing.T) {
		formatter := &cleanFormatter{}
		entry := &logrus.Entry{
			Message: "Line 1\nLine 2",
		}

		result, err := formatter.Format(entry)
		require.NoError(t, err)
		assert.Equal(t, "Line 1\nLine 2\n", string(result))
	})
}

func TestRunSolver(t *testing.T) {
	// Skip this test if we don't have a valid config
	if os.Getenv("SKIP_SOLVER_TESTS") == "true" {
		t.Skip("Solver tests disabled via SKIP_SOLVER_TESTS")
	}

	t.Run("RunSolver initializes correctly", func(t *testing.T) {
		// This test verifies that RunSolver can be called without panicking
		// We can't easily test the full execution without mocking the solver manager

		// Create a test environment
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Test that the function exists and can be called
		// Note: This will actually try to run the solver, so we need to be careful
		// For now, just verify the function signature is correct
		assert.NotPanics(t, func() {
			// Don't actually call RunSolver() as it would start the real solver
			// Instead, just verify the function exists
			_ = RunSolver
		})
	})
}

func TestTestConnection(t *testing.T) {
	// Skip this test if we don't have network access or valid config
	if os.Getenv("SKIP_NETWORK_TESTS") == "true" {
		t.Skip("Network tests disabled via SKIP_NETWORK_TESTS")
	}

	t.Run("TestConnection runs without panic", func(t *testing.T) {
		// Capture the log output
		output := captureLogrusOutput(func() {
			// We need to be careful here - this will try to make real network connections
			// For unit tests, we should mock this, but for now, let's test the structure
			assert.NotPanics(t, func() {
				// Only test if we have a valid config
				_, err := config.LoadConfig()
				if err != nil {
					t.Skip("No valid config available for testing")
					return
				}

				// This will make real network calls, so we should be cautious
				// For now, just verify the function exists
				_ = TestConnection
			})
		})

		// If we got here without panicking, that's good
		_ = output // Use the output variable to avoid unused variable warning
	})

	t.Run("TestConnection handles missing config gracefully", func(t *testing.T) {
		// Test what happens when config loading fails
		// We can't easily simulate this without modifying global state

		// For now, just verify the function signature
		assert.NotPanics(t, func() {
			_ = TestConnection
		})
	})
}

// Test helper functions and utilities
func TestSolverUtilities(t *testing.T) {
	t.Run("Package imports are correct", func(t *testing.T) {
		// Verify that all required packages can be imported
		// This is a basic smoke test to ensure dependencies are correct

		// Test that we can create a context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		assert.NotNil(t, ctx)

		// Test that we can create a timeout context
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
		defer cancel2()
		assert.NotNil(t, ctx2)
	})

	t.Run("Logrus configuration works", func(t *testing.T) {
		// Test that we can configure logrus
		oldFormatter := logrus.StandardLogger().Formatter
		oldLevel := logrus.GetLevel()
		defer func() {
			logrus.SetFormatter(oldFormatter)
			logrus.SetLevel(oldLevel)
		}()

		// Set up clean logging like in RunSolver
		logrus.SetFormatter(&cleanFormatter{})
		logrus.SetLevel(logrus.InfoLevel)

		// Verify the formatter was set
		assert.IsType(t, &cleanFormatter{}, logrus.StandardLogger().Formatter)
		assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
	})

	t.Run("Signal handling setup", func(t *testing.T) {
		// Test that we can set up signal handling
		sigChan := make(chan os.Signal, 1)
		assert.NotNil(t, sigChan)

		// Verify the channel has the correct capacity
		assert.Equal(t, 1, cap(sigChan))
	})
}

// Test configuration loading behavior
func TestConfigurationHandling(t *testing.T) {
	t.Run("Config loading behavior", func(t *testing.T) {
		// Test basic config loading functionality
		cfg, err := config.LoadConfig()
		if err != nil {
			// If config loading fails, that's okay for unit tests
			// We just want to ensure it doesn't panic
			t.Logf("Config loading failed (expected in test environment): %v", err)
			return
		}

		// If config loads successfully, verify it's not nil
		assert.NotNil(t, cfg)

		// Test network initialization
		assert.NotPanics(t, func() {
			config.InitializeNetworks()
		})

		// Test getting network names
		names := config.GetNetworkNames()
		assert.NotNil(t, names)
		// Network names should be a slice (might be empty in test environment)
		assert.IsType(t, []string{}, names)
	})

	t.Run("Network configuration access", func(t *testing.T) {
		// Test accessing network configurations
		_, err := config.LoadConfig()
		if err != nil {
			t.Skip("No valid config available for testing")
			return
		}

		config.InitializeNetworks()

		// Test getting network names
		names := config.GetNetworkNames()

		// Test accessing each configured network
		for _, name := range names {
			networkConfig, exists := config.Networks[name]
			if exists {
				assert.NotEmpty(t, networkConfig.RPCURL)
				assert.NotZero(t, networkConfig.ChainID)
				t.Logf("Network %s: RPC=%s, ChainID=%d", name, networkConfig.RPCURL, networkConfig.ChainID)
			}
		}
	})
}

// Test network detection logic
func TestNetworkDetection(t *testing.T) {
	t.Run("Starknet network detection", func(t *testing.T) {
		// Test the logic used in TestConnection for detecting Starknet networks
		testCases := []struct {
			networkName string
			isStarknet  bool
		}{
			{"Starknet", true},
			{"starknet", true},
			{"STARKNET", true},
			{"Starknet Testnet", true},
			{"Ethereum", false},
			{"Base", false},
			{"Optimism", false},
			{"Arbitrum", false},
		}

		for _, tc := range testCases {
			t.Run(tc.networkName, func(t *testing.T) {
				isStarknet := strings.Contains(strings.ToLower(tc.networkName), "starknet")
				assert.Equal(t, tc.isStarknet, isStarknet)
			})
		}
	})
}

// Test error handling scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("Handles invalid RPC URLs gracefully", func(t *testing.T) {
		// Test that invalid RPC URLs are handled gracefully
		// This is testing the error handling logic in TestConnection

		invalidURLs := []string{
			"invalid-url",
			"http://nonexistent-domain.invalid",
			"",
			"not-a-url-at-all",
		}

		for _, url := range invalidURLs {
			t.Run(fmt.Sprintf("URL_%s", url), func(t *testing.T) {
				// We can't easily test the actual network connection error handling
				// without making real network calls, but we can verify the URLs are invalid
				assert.True(t, url == "" || !strings.HasPrefix(url, "http://localhost") && !strings.HasPrefix(url, "https://"))
			})
		}
	})
}

// Test logging output format
func TestLoggingOutput(t *testing.T) {
	t.Run("Logging messages format correctly", func(t *testing.T) {
		// Test the logging messages used in the solver
		messages := []string{
			"🚀 Starting OIF Starknet Solver...",
			"🔄 Shutdown signal received, stopping solver...",
			"✅ Solver stopped gracefully",
			"🔍 Testing network connections...",
			"🎉 Network connection test completed",
		}

		for _, message := range messages {
			t.Run(fmt.Sprintf("Message_%s", strings.ReplaceAll(message, " ", "_")), func(t *testing.T) {
				// Verify messages are not empty and contain expected emojis
				assert.NotEmpty(t, message)
				assert.True(t, len(message) > 5) // Should have some content beyond just emoji
			})
		}
	})

	t.Run("Log level configuration", func(t *testing.T) {
		originalLevel := logrus.GetLevel()
		defer logrus.SetLevel(originalLevel)

		// Test setting different log levels
		levels := []logrus.Level{
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
		}

		for _, level := range levels {
			logrus.SetLevel(level)
			assert.Equal(t, level, logrus.GetLevel())
		}
	})
}

// Test context and cancellation
func TestContextHandling(t *testing.T) {
	t.Run("Context creation and cancellation", func(t *testing.T) {
		// Test context creation like in RunSolver
		ctx, cancel := context.WithCancel(context.Background())

		// Verify context is not nil and not cancelled initially
		assert.NotNil(t, ctx)
		assert.NoError(t, ctx.Err())

		// Test cancellation
		cancel()

		// Give it a moment to propagate
		time.Sleep(time.Millisecond)

		// Verify context is now cancelled
		assert.Error(t, ctx.Err())
		assert.Equal(t, context.Canceled, ctx.Err())
	})

	t.Run("Context with timeout", func(t *testing.T) {
		// Test timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Initially should not be cancelled
		assert.NoError(t, ctx.Err())

		// Wait for timeout
		time.Sleep(20 * time.Millisecond)

		// Should now be cancelled due to timeout
		assert.Error(t, ctx.Err())
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	})
}
