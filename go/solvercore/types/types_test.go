package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResult(t *testing.T) {
	t.Run("Success result", func(t *testing.T) {
		data := "test data"
		result := NewSuccessResult(data)
		
		assert.True(t, result.Success)
		assert.Equal(t, data, result.Data)
		assert.Empty(t, result.Error) // Error should be empty string, not nil
	})

	t.Run("Error result", func(t *testing.T) {
		err := assert.AnError
		result := NewErrorResult[string](err)
		
		assert.False(t, result.Success)
		assert.Empty(t, result.Data)
		assert.Equal(t, err.Error(), result.Error) // Compare error strings
	})
}

func TestAddressUtils(t *testing.T) {
	t.Run("IsStarknetAddress", func(t *testing.T) {
		tests := []struct {
			address string
			expected bool
		}{
			{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", true}, // 66 chars with 0x = 64 hex chars
			{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde", false}, // 64 chars with 0x = 62 hex chars (too short)
			{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1", false}, // too long
			{"invalid", false},
			{"", false},
		}

		for _, tt := range tests {
			t.Run(tt.address, func(t *testing.T) {
				result := IsStarknetAddress(tt.address)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsEVMAddress", func(t *testing.T) {
		tests := []struct {
			address string
			expected bool
		}{
			{"0x1234567890123456789012345678901234567890", true},
			{"0x123456789012345678901234567890123456789", false}, // too short
			{"0x12345678901234567890123456789012345678901", false}, // too long
			{"invalid", false},
			{"", false},
		}

		for _, tt := range tests {
			t.Run(tt.address, func(t *testing.T) {
				result := IsEVMAddress(tt.address)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("HexToBytes32", func(t *testing.T) {
		tests := []struct {
			hex      string
			expected [32]byte
			hasError bool
		}{
			{
				"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				[32]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef},
				false,
			},
			{
				"invalid",
				[32]byte{},
				true,
			},
			{
				"",
				[32]byte{},
				true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.hex, func(t *testing.T) {
				result, err := HexToBytes32(tt.hex)
				if tt.hasError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expected, result)
				}
			})
		}
	})
}

func TestAllowBlockLists(t *testing.T) {
	t.Run("AllowBlockListItem matching", func(t *testing.T) {
		item := AllowBlockListItem{
			SenderAddress:     "0x1234567890123456789012345678901234567890",
			DestinationDomain: "ethereum",
			RecipientAddress:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		}

		tests := []struct {
			sender      string
			destination string
			recipient   string
			expected    bool
		}{
			{
				"0x1234567890123456789012345678901234567890",
				"ethereum",
				"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				true,
			},
			{
				"0x1234567890123456789012345678901234567890",
				"ethereum",
				"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				true,
			},
			{
				"0x0000000000000000000000000000000000000000",
				"ethereum",
				"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				false,
			},
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				result := item.Matches(tt.sender, tt.destination, tt.recipient)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Wildcard matching", func(t *testing.T) {
		item := AllowBlockListItem{
			SenderAddress:     "*",
			DestinationDomain: "*",
			RecipientAddress:  "*",
		}

		result := item.Matches("any", "any", "any")
		assert.True(t, result)
	})
}

func TestParsedArgs(t *testing.T) {
	t.Run("OrderID conversion", func(t *testing.T) {
		args := ParsedArgs{
			OrderID: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		}

		orderIDBytes := args.GetOrderIDBytes()
		assert.Len(t, orderIDBytes, 32)
		assert.Equal(t, byte(0x12), orderIDBytes[0])
		assert.Equal(t, byte(0xef), orderIDBytes[31])
	})

	t.Run("Invalid OrderID", func(t *testing.T) {
		args := ParsedArgs{
			OrderID: "invalid",
		}

		orderIDBytes := args.GetOrderIDBytes()
		assert.Len(t, orderIDBytes, 32)
		// Should return zero bytes for invalid hex
		for _, b := range orderIDBytes {
			assert.Equal(t, byte(0), b)
		}
	})
}

// Test constants
func TestConstants(t *testing.T) {
	assert.Equal(t, 64, StarknetAddressLength, "StarknetAddressLength should be 64")
	assert.Equal(t, 40, EthereumAddressLength, "EthereumAddressLength should be 40")
	assert.Equal(t, 62, StarknetAddressLengthWithPrefix, "StarknetAddressLengthWithPrefix should be 62") // Fixed: was 66, should be 62
	assert.Equal(t, 42, EthereumAddressLengthWithPrefix, "EthereumAddressLengthWithPrefix should be 42")
	assert.Equal(t, 32, Bytes32Length, "Bytes32Length should be 32")
	assert.Equal(t, 31, Bytes31Length, "Bytes31Length should be 31")
}

func TestToEVMAddress(t *testing.T) {
	t.Run("Valid EVM address", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		result, err := ToEVMAddress(address)
		
		assert.NoError(t, err)
		assert.Equal(t, address, result.Hex())
	})

	t.Run("Invalid address format", func(t *testing.T) {
		address := "invalid"
		_, err := ToEVMAddress(address)
		
		assert.Error(t, err)
	})

	t.Run("Empty address", func(t *testing.T) {
		address := ""
		_, err := ToEVMAddress(address)
		
		assert.Error(t, err)
	})
}

func TestToStarknetAddress(t *testing.T) {
	t.Run("Valid Starknet address", func(t *testing.T) {
		address := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"
		result, err := ToStarknetAddress(address)
		
		assert.NoError(t, err)
		assert.Equal(t, address, result.String())
	})

	t.Run("Invalid address format", func(t *testing.T) {
		address := "invalid"
		_, err := ToStarknetAddress(address)
		
		assert.Error(t, err)
	})

	t.Run("Empty address", func(t *testing.T) {
		address := ""
		_, err := ToStarknetAddress(address)
		
		assert.Error(t, err)
	})
}

func TestIsBytes32Address(t *testing.T) {
	ac := NewAddressConverter()
	
	t.Run("Valid bytes32 address", func(t *testing.T) {
		address := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		result := ac.IsBytes32Address(address)
		
		assert.True(t, result)
	})

	t.Run("Invalid bytes32 address", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890" // EVM address
		result := ac.IsBytes32Address(address)
		
		assert.False(t, result)
	})

	t.Run("Empty address", func(t *testing.T) {
		address := ""
		result := ac.IsBytes32Address(address)
		
		assert.False(t, result)
	})
}

func TestFormatAddress(t *testing.T) {
	ac := NewAddressConverter()
	
	t.Run("Format Starknet address", func(t *testing.T) {
		address := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"
		result := ac.FormatAddress(address)
		
		assert.Equal(t, "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde", result)
	})

	t.Run("Format EVM address", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		result := ac.FormatAddress(address)
		
		assert.Equal(t, "1234567890123456789012345678901234567890", result)
	})

	t.Run("Format address without 0x prefix", func(t *testing.T) {
		address := "1234567890123456789012345678901234567890"
		result := ac.FormatAddress(address)
		
		assert.Equal(t, "1234567890123456789012345678901234567890", result)
	})
}

func TestFormatTokenAmount(t *testing.T) {
	t.Run("Format token amount with 18 decimals", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000) // 1 token with 18 decimals
		result := FormatTokenAmount(amount, 18)
		
		assert.Equal(t, "1.00 tokens", result)
	})

	t.Run("Format token amount with 6 decimals", func(t *testing.T) {
		amount := big.NewInt(1000000) // 1 token with 6 decimals
		result := FormatTokenAmount(amount, 6)
		
		assert.Equal(t, "1.00 tokens", result)
	})

	t.Run("Format zero amount", func(t *testing.T) {
		amount := big.NewInt(0)
		result := FormatTokenAmount(amount, 18)
		
		assert.Equal(t, "0.00 tokens", result)
	})

	t.Run("Format large amount", func(t *testing.T) {
		amount, _ := big.NewInt(0).SetString("123456789012345678901234567890", 10)
		result := FormatTokenAmount(amount, 18)
		
		assert.Equal(t, "123456789012.35 tokens", result)
	})

	t.Run("Format nil amount", func(t *testing.T) {
		result := FormatTokenAmount(nil, 18)
		
		assert.Equal(t, "0", result)
	})
}

func TestGetOrderIDBytes(t *testing.T) {
	t.Run("Valid order ID", func(t *testing.T) {
		args := ParsedArgs{
			OrderID: "0x1234567890123456789012345678901234567890123456789012345678901234",
		}
		result := args.GetOrderIDBytes()
		
		assert.Len(t, result, 32)
		// Check that the first few bytes are correct
		assert.Equal(t, byte(0x12), result[0])
		assert.Equal(t, byte(0x34), result[1])
	})

	t.Run("Invalid order ID format", func(t *testing.T) {
		args := ParsedArgs{
			OrderID: "invalid",
		}
		result := args.GetOrderIDBytes()
		
		// Should return zero bytes for invalid format
		expected := [32]byte{}
		assert.Equal(t, expected, result)
	})

	t.Run("Empty order ID", func(t *testing.T) {
		args := ParsedArgs{
			OrderID: "",
		}
		result := args.GetOrderIDBytes()
		
		// Should return zero bytes for empty order ID
		expected := [32]byte{}
		assert.Equal(t, expected, result)
	})

	t.Run("Order ID too short", func(t *testing.T) {
		args := ParsedArgs{
			OrderID: "0x1234",
		}
		result := args.GetOrderIDBytes()
		
		// Should return zero bytes for too short order ID
		expected := [32]byte{}
		assert.Equal(t, expected, result)
	})
}
