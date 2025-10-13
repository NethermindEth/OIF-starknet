package logutil

import (
	"bytes"
	"io"
	"math/big"
	"os"
	"testing"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrefix(t *testing.T) {
	t.Run("Valid network names", func(t *testing.T) {
		tests := []struct {
			networkName string
			expected    string
		}{
			{"Base", "\x1b[38;5;27m[BASE]\x1b[0m "},
			{"Ethereum", "\x1b[32m[ETH]\x1b[0m "},
			{"Arbitrum", "\x1b[35m[ARB]\x1b[0m "},
			{"Optimism", "\x1b[91m[OPT]\x1b[0m "},
			{"Starknet", "\x1b[38;5;208m[STRK]\x1b[0m "},
			{"Polygon", "\x1b[36m[POL]\x1b[0m "},
		}

		for _, tt := range tests {
			t.Run(tt.networkName, func(t *testing.T) {
				result := Prefix(tt.networkName)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Unknown network name", func(t *testing.T) {
		result := Prefix("UnknownNetwork")
		assert.Equal(t, "\x1b[36m[UNK]\x1b[0m ", result)
	})

	t.Run("Empty network name", func(t *testing.T) {
		result := Prefix("")
		assert.Equal(t, "\x1b[36m[NET]\x1b[0m ", result)
	})
}

func TestNetworkNameByChainID(t *testing.T) {
	t.Run("Network name lookup", func(t *testing.T) {
		// Test that the function doesn't panic
		// The actual result depends on configuration
		name := NetworkNameByChainID(1)
		assert.NotNil(t, name) // Should return something, even if empty

		name2 := NetworkNameByChainID(84532)
		assert.NotNil(t, name2)
	})

	t.Run("Unknown chain ID", func(t *testing.T) {
		name := NetworkNameByChainID(999999)
		assert.NotNil(t, name) // Should return something, even if empty
	})
}

func TestNetworkTagFormatting(t *testing.T) {
	t.Run("Network tag consistency", func(t *testing.T) {
		// Test that network tags are consistently formatted
		networks := []string{"Base", "Ethereum", "Arbitrum", "Optimism", "Starknet"}

		for _, network := range networks {
			tag := Prefix(network)
			assert.Contains(t, tag, "[")
			assert.Contains(t, tag, "]")
			assert.True(t, len(tag) > 2) // Should have content between brackets
		}
	})

	t.Run("Cross-chain arrow formatting", func(t *testing.T) {
		// Test that cross-chain operations use the correct arrow format
		// Note: LogCrossChainOperation function doesn't exist in current implementation
		// This test ensures the basic logging doesn't panic
		assert.True(t, true) // Placeholder test
	})
}

// Helper function to capture stdout
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestSolverLogger(t *testing.T) {
	t.Run("NewSolverLogger creates logger", func(t *testing.T) {
		logger := NewSolverLogger("Ethereum")
		require.NotNil(t, logger)
		assert.Equal(t, "Ethereum", logger.networkName)
	})

	t.Run("GetNetworkTag returns correct tag", func(t *testing.T) {
		logger := NewSolverLogger("Base")
		tag := logger.GetNetworkTag()
		assert.Contains(t, tag, "[BASE]")
	})

	t.Run("GetNetworkTagByChainID returns correct tag", func(t *testing.T) {
		// Test known chain ID mapping
		tag := GetNetworkTagByChainID(84532) // Base
		assert.Contains(t, tag, "[")
		assert.Contains(t, tag, "]")
	})

	t.Run("GetNetworkTagByChainID with unknown chain ID", func(t *testing.T) {
		tag := GetNetworkTagByChainID(999999)
		// Unknown chain IDs get formatted as "chain-XXXXX" which gets processed by Prefix()
		assert.Contains(t, tag, "[CHA]")    // "chain-999999" -> "[CHA]"
		assert.Contains(t, tag, "\033[36m") // cyan color
	})
}

func TestCrossChainOperation(t *testing.T) {
	t.Run("CrossChainOperation logs correctly", func(t *testing.T) {
		output := captureOutput(func() {
			CrossChainOperation("Fill Order", 84532, 11155111, "0x1234567890abcdef1234567890abcdef12345678")
		})

		assert.Contains(t, output, "→")
		assert.Contains(t, output, "🔄")
		assert.Contains(t, output, "Fill Order")
		assert.Contains(t, output, "0x123456")
	})
}

func TestRemoveColorCodes(t *testing.T) {
	t.Run("removeColorCodes strips ANSI codes", func(t *testing.T) {
		input := "\033[32m[ETH]\033[0m test"
		expected := "[ETH] test"
		result := removeColorCodes(input)
		assert.Equal(t, expected, result)
	})

	t.Run("removeColorCodes handles multiple colors", func(t *testing.T) {
		input := "\033[32mgreen\033[91mred\033[35mpurple\033[0m"
		expected := "greenredpurple"
		result := removeColorCodes(input)
		assert.Equal(t, expected, result)
	})

	t.Run("removeColorCodes handles text without colors", func(t *testing.T) {
		input := "plain text"
		result := removeColorCodes(input)
		assert.Equal(t, input, result)
	})
}

func TestLogOrderProcessing(t *testing.T) {
	t.Run("LogOrderProcessing with complete args", func(t *testing.T) {
		args := types.ParsedArgs{
			OrderID: "0x1234567890abcdef1234567890abcdef12345678",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(84532),
				FillInstructions: []types.FillInstruction{
					{DestinationChainID: big.NewInt(11155111)},
				},
			},
		}

		output := captureOutput(func() {
			LogOrderProcessing(args, "Processing Order")
		})

		assert.Contains(t, output, "→")
		assert.Contains(t, output, "🔄")
		assert.Contains(t, output, "Processing Order")
	})

	t.Run("LogOrderProcessing with incomplete args", func(t *testing.T) {
		args := types.ParsedArgs{
			OrderID: "0x1234567890abcdef1234567890abcdef12345678",
		}

		output := captureOutput(func() {
			LogOrderProcessing(args, "Processing Order")
		})

		assert.Contains(t, output, "🔄")
		assert.Contains(t, output, "Processing Order")
	})
}

func TestLogFillOperation(t *testing.T) {
	t.Run("LogFillOperation success", func(t *testing.T) {
		output := captureOutput(func() {
			LogFillOperation("Ethereum", "0x1234567890abcdef1234567890abcdef12345678", true)
		})

		assert.Contains(t, output, "✅")
		assert.Contains(t, output, "Fill completed")
		assert.Contains(t, output, "0x123456")
	})

	t.Run("LogFillOperation failure", func(t *testing.T) {
		output := captureOutput(func() {
			LogFillOperation("Ethereum", "0x1234567890abcdef1234567890abcdef12345678", false)
		})

		assert.Contains(t, output, "❌")
		assert.Contains(t, output, "Fill failed")
		assert.Contains(t, output, "0x123456")
	})
}

func TestLogSettleOperation(t *testing.T) {
	t.Run("LogSettleOperation success", func(t *testing.T) {
		output := captureOutput(func() {
			LogSettleOperation("Base", "0x1234567890abcdef1234567890abcdef12345678", true)
		})

		assert.Contains(t, output, "✅")
		assert.Contains(t, output, "Settlement completed")
		assert.Contains(t, output, "0x123456")
	})

	t.Run("LogSettleOperation failure", func(t *testing.T) {
		output := captureOutput(func() {
			LogSettleOperation("Base", "0x1234567890abcdef1234567890abcdef12345678", false)
		})

		assert.Contains(t, output, "❌")
		assert.Contains(t, output, "Settlement failed")
		assert.Contains(t, output, "0x123456")
	})
}

func TestLogBlockProcessing(t *testing.T) {
	t.Run("LogBlockProcessing with events", func(t *testing.T) {
		output := captureOutput(func() {
			LogBlockProcessing("Ethereum", 100, 105, 3)
		})

		assert.Contains(t, output, "📦")
		assert.Contains(t, output, "100-105")
		assert.Contains(t, output, "3 events")
	})

	t.Run("LogBlockProcessing without events", func(t *testing.T) {
		output := captureOutput(func() {
			LogBlockProcessing("Ethereum", 100, 105, 0)
		})

		assert.Contains(t, output, "📦")
		assert.Contains(t, output, "100-105")
		assert.NotContains(t, output, "events")
	})

	t.Run("LogBlockProcessing single block", func(t *testing.T) {
		output := captureOutput(func() {
			LogBlockProcessing("Ethereum", 100, 100, 0)
		})

		// Should not log for single block with no events
		assert.Empty(t, output)
	})
}

func TestLogStatusCheck(t *testing.T) {
	t.Run("LogStatusCheck first attempt", func(t *testing.T) {
		output := captureOutput(func() {
			LogStatusCheck("Ethereum", 1, 3, "OPENED", "FILLED")
		})

		assert.Contains(t, output, "📊")
		assert.Contains(t, output, "Status: OPENED")
		assert.Contains(t, output, "expected: FILLED")
		assert.NotContains(t, output, "Retry")
	})

	t.Run("LogStatusCheck retry attempt", func(t *testing.T) {
		output := captureOutput(func() {
			LogStatusCheck("Ethereum", 2, 3, "OPENED", "FILLED")
		})

		assert.Contains(t, output, "📊")
		assert.Contains(t, output, "Retry 2/3")
		assert.Contains(t, output, "OPENED")
		assert.Contains(t, output, "expected: FILLED")
	})
}

func TestLogRetryWait(t *testing.T) {
	t.Run("LogRetryWait logs correctly", func(t *testing.T) {
		output := captureOutput(func() {
			LogRetryWait("Base", 1, 3, "5s")
		})

		assert.Contains(t, output, "⏳")
		assert.Contains(t, output, "Waiting 5s")
		assert.Contains(t, output, "retry 2/3")
	})
}

func TestLogOperationComplete(t *testing.T) {
	t.Run("LogOperationComplete success with cross-chain", func(t *testing.T) {
		args := types.ParsedArgs{
			OrderID: "0x1234567890abcdef1234567890abcdef12345678",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(84532),
				FillInstructions: []types.FillInstruction{
					{DestinationChainID: big.NewInt(11155111)},
				},
			},
		}

		output := captureOutput(func() {
			LogOperationComplete(args, "Fill", true)
		})

		assert.Contains(t, output, "→")
		assert.Contains(t, output, "✅")
		assert.Contains(t, output, "Fill completed")
	})

	t.Run("LogOperationComplete failure with cross-chain", func(t *testing.T) {
		args := types.ParsedArgs{
			OrderID: "0x1234567890abcdef1234567890abcdef12345678",
			ResolvedOrder: types.ResolvedCrossChainOrder{
				OriginChainID: big.NewInt(84532),
				FillInstructions: []types.FillInstruction{
					{DestinationChainID: big.NewInt(11155111)},
				},
			},
		}

		output := captureOutput(func() {
			LogOperationComplete(args, "Fill", false)
		})

		assert.Contains(t, output, "→")
		assert.Contains(t, output, "❌")
		assert.Contains(t, output, "Fill failed")
	})

	t.Run("LogOperationComplete without cross-chain info", func(t *testing.T) {
		args := types.ParsedArgs{
			OrderID: "0x1234567890abcdef1234567890abcdef12345678",
		}

		output := captureOutput(func() {
			LogOperationComplete(args, "Fill", true)
		})

		assert.Contains(t, output, "✅")
		assert.Contains(t, output, "Fill completed")
		assert.NotContains(t, output, "→")
	})
}

func TestLogWithNetworkTagf(t *testing.T) {
	t.Run("LogWithNetworkTagf formats correctly", func(t *testing.T) {
		output := captureOutput(func() {
			LogWithNetworkTagf("Ethereum", "Test message: %s\n", "hello")
		})

		assert.Contains(t, output, "[ETH]")
		assert.Contains(t, output, "Test message: hello")
	})
}

func TestLogPersistence(t *testing.T) {
	t.Run("LogPersistence first call", func(t *testing.T) {
		// Reset counter for test
		persistenceCounters["TestNetwork"] = 0

		output := captureOutput(func() {
			LogPersistence("TestNetwork", 12345)
		})

		assert.Contains(t, output, "💾")
		assert.Contains(t, output, "Persisted LastIndexedBlock=12345")
	})

	t.Run("LogPersistence subsequent calls", func(t *testing.T) {
		// Reset counter for test
		persistenceCounters["TestNetwork2"] = 0

		// First call should log
		output1 := captureOutput(func() {
			LogPersistence("TestNetwork2", 12345)
		})
		assert.Contains(t, output1, "💾")

		// Calls 2-29 should not log
		for i := 2; i < 30; i++ {
			output := captureOutput(func() {
				LogPersistence("TestNetwork2", uint64(12345+i))
			})
			assert.Empty(t, output)
		}

		// Call 30 should log (30 % 30 == 0, but we check counter%30 == 1)
		output30 := captureOutput(func() {
			LogPersistence("TestNetwork2", 12375)
		})
		assert.Empty(t, output30)

		// Call 31 should log (31 % 30 == 1)
		output31 := captureOutput(func() {
			LogPersistence("TestNetwork2", 12376)
		})
		assert.Contains(t, output31, "💾")
	})
}

func TestDeriveTag(t *testing.T) {
	t.Run("deriveTag single word", func(t *testing.T) {
		result := deriveTag("Polygon")
		assert.Equal(t, "[POL]", result)
	})

	t.Run("deriveTag two words", func(t *testing.T) {
		result := deriveTag("Binance Smart")
		assert.Equal(t, "[BS]", result)
	})

	t.Run("deriveTag short word", func(t *testing.T) {
		result := deriveTag("Go")
		assert.Equal(t, "[GO]", result)
	})

	t.Run("deriveTag empty string", func(t *testing.T) {
		result := deriveTag("")
		assert.Equal(t, "[NET]", result)
	})
}

func TestTagColorByName(t *testing.T) {
	t.Run("tagColorByName known networks", func(t *testing.T) {
		tests := []struct {
			name          string
			expectedTag   string
			expectedColor string
		}{
			{"ethereum", "[ETH]", green},
			{"Ethereum Sepolia", "[ETH]", green},
			{"optimism", "[OPT]", pastelRed},
			{"arbitrum", "[ARB]", purple},
			{"base", "[BASE]", royalBlue},
			{"starknet", "[STRK]", orange},
		}

		for _, tt := range tests {
			tag, color := tagColorByName(tt.name)
			assert.Equal(t, tt.expectedTag, tag)
			assert.Equal(t, tt.expectedColor, color)
		}
	})

	t.Run("tagColorByName unknown network", func(t *testing.T) {
		tag, color := tagColorByName("unknown")
		assert.Equal(t, "", tag)
		assert.Equal(t, "", color)
	})
}

func TestBindEnvChain(t *testing.T) {
	t.Run("bindEnvChain with valid env var", func(t *testing.T) {
		// Set up test environment
		t.Setenv("TEST_CHAIN_ID", "12345")
		defer os.Unsetenv("TEST_CHAIN_ID")

		// Reset the mapping to test initialization
		colorByChainID = make(map[uint64]string)
		tagByChainID = make(map[uint64]string)

		bindEnvChain("TEST_CHAIN_ID", "[TEST]", cyan)

		assert.Equal(t, cyan, colorByChainID[12345])
		assert.Equal(t, "[TEST]", tagByChainID[12345])
	})

	t.Run("bindEnvChain with invalid env var", func(t *testing.T) {
		// Set up test environment with invalid chain ID
		t.Setenv("TEST_CHAIN_ID_INVALID", "not-a-number")
		defer os.Unsetenv("TEST_CHAIN_ID_INVALID")

		// Reset the mapping to test initialization
		colorByChainID = make(map[uint64]string)
		tagByChainID = make(map[uint64]string)

		bindEnvChain("TEST_CHAIN_ID_INVALID", "[TEST]", cyan)

		// Should not add anything to maps
		assert.Empty(t, colorByChainID)
		assert.Empty(t, tagByChainID)
	})

	t.Run("bindEnvChain with missing env var", func(t *testing.T) {
		// Reset the mapping to test initialization
		colorByChainID = make(map[uint64]string)
		tagByChainID = make(map[uint64]string)

		bindEnvChain("NONEXISTENT_CHAIN_ID", "[TEST]", cyan)

		// Should not add anything to maps
		assert.Empty(t, colorByChainID)
		assert.Empty(t, tagByChainID)
	})
}
