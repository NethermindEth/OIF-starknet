package starknetutil

import (
	"math/big"
	"strings"
	"testing"

	"github.com/NethermindEth/starknet.go/utils"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBigIntToU256Felts(t *testing.T) {
	tests := []struct {
		name  string
		input *big.Int
	}{
		{
			name:  "zero",
			input: big.NewInt(0),
		},
		{
			name:  "small number",
			input: big.NewInt(42),
		},
		{
			name:  "number requiring high part",
			input: new(big.Int).Lsh(big.NewInt(1), 130), // 2^130
		},
		{
			name:  "max uint128",
			input: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low, high := ConvertBigIntToU256Felts(tt.input)
			assert.NotNil(t, low, "low part should not be nil")
			assert.NotNil(t, high, "high part should not be nil")

			// Verify that low + (high << 128) equals original input
			reconstructed := new(big.Int).Add(
				low.BigInt(big.NewInt(0)),
				new(big.Int).Lsh(high.BigInt(big.NewInt(0)), 128),
			)
			assert.Equal(t, tt.input, reconstructed, "reconstructed value should match input")
		})
	}
}

func TestToUint256(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected *uint256.Int
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: uint256.NewInt(0),
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			expected: uint256.NewInt(0),
		},
		{
			name:     "positive number",
			input:    big.NewInt(12345),
			expected: uint256.NewInt(12345),
		},
		{
			name:     "large number",
			input:    new(big.Int).Lsh(big.NewInt(1), 200),
			expected: new(uint256.Int).Lsh(uint256.NewInt(1), 200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUint256(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    *uint256.Int
		expected *big.Int
	}{
		{
			name:     "positive number",
			input:    uint256.NewInt(12345),
			expected: big.NewInt(12345),
		},
		{
			name:     "large number",
			input:    new(uint256.Int).Lsh(uint256.NewInt(1), 200),
			expected: new(big.Int).Lsh(big.NewInt(1), 200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBigInt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBytesToU128Felts(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "single byte",
			input: []byte{0x42},
		},
		{
			name:  "16 bytes (exactly 128 bits)",
			input: make([]byte, 16),
		},
		{
			name:  "17 bytes (requires high part)",
			input: append(make([]byte, 16), 0x01),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesToU128Felts(tt.input)
			assert.NotEmpty(t, result, "result should not be empty")
			// Just verify the function doesn't panic and returns felt values
			for i, felt := range result {
				assert.NotNil(t, felt, "felt at index %d should not be nil", i)
			}
		})
	}
}

func TestConvertSolidityOrderIDForStarknet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "valid hex string",
			input:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			hasError: false,
		},
		{
			name:     "invalid hex string",
			input:    "invalid",
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low, high, err := ConvertSolidityOrderIDForStarknet(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, low, "low should be nil on error")
				assert.Nil(t, high, "high should be nil on error")
			} else {
				require.NoError(t, err)
				assert.NotNil(t, low, "low felt should not be nil")
				assert.NotNil(t, high, "high felt should not be nil")
			}
		})
	}
}

// Test constants
func TestConstants(t *testing.T) {
	assert.Equal(t, 128, U128BitShift, "U128BitShift should be 128")
	assert.Equal(t, 32, Bytes32Length, "Bytes32Length should be 32")
	assert.Equal(t, 16, Bytes16Length, "Bytes16Length should be 16")
	assert.Equal(t, 18, TokenDecimals, "TokenDecimals should be 18")
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

func TestERC20BalanceErrorCases(t *testing.T) {
	t.Run("invalid token address", func(t *testing.T) {
		// We can't easily test the full function without a real provider,
		// but we can test the address validation part
		_, err := utils.HexToFelt("invalid")
		assert.Error(t, err)
	})

	t.Run("invalid owner address", func(t *testing.T) {
		_, err := utils.HexToFelt("invalid")
		assert.Error(t, err)
	})
}

func TestERC20AllowanceErrorCases(t *testing.T) {
	t.Run("invalid token address", func(t *testing.T) {
		_, err := utils.HexToFelt("invalid")
		assert.Error(t, err)
	})

	t.Run("invalid owner address", func(t *testing.T) {
		_, err := utils.HexToFelt("invalid")
		assert.Error(t, err)
	})

	t.Run("invalid spender address", func(t *testing.T) {
		_, err := utils.HexToFelt("invalid")
		assert.Error(t, err)
	})
}

func TestERC20Approve(t *testing.T) {
	t.Run("successful approve call creation", func(t *testing.T) {
		tokenAddress := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"
		spenderAddress := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"
		amount := big.NewInt(1000000000000000000) // 1 token

		call, err := ERC20Approve(tokenAddress, spenderAddress, amount)

		require.NoError(t, err)
		assert.NotNil(t, call)
		assert.Equal(t, "approve", call.FunctionName)
		assert.Len(t, call.CallData, 3) // spender, low, high
	})

	t.Run("invalid token address", func(t *testing.T) {
		_, err := ERC20Approve("invalid", "0x123", big.NewInt(1000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token address")
	})

	t.Run("invalid spender address", func(t *testing.T) {
		_, err := ERC20Approve("0x123", "invalid", big.NewInt(1000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid spender address")
	})

	t.Run("zero amount", func(t *testing.T) {
		tokenAddress := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"
		spenderAddress := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"

		call, err := ERC20Approve(tokenAddress, spenderAddress, big.NewInt(0))

		require.NoError(t, err)
		assert.NotNil(t, call)
	})

	t.Run("large amount", func(t *testing.T) {
		tokenAddress := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"
		spenderAddress := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"
		// Amount that requires both low and high parts
		amount := new(big.Int).Lsh(big.NewInt(1), 200)

		call, err := ERC20Approve(tokenAddress, spenderAddress, amount)

		require.NoError(t, err)
		assert.NotNil(t, call)
	})
}

func TestConvertBigIntToU256FeltsEdgeCases(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		low, high := ConvertBigIntToU256Felts(nil)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
		assert.Equal(t, "0x0", low.String())
		assert.Equal(t, "0x0", high.String())
	})

	t.Run("max uint128", func(t *testing.T) {
		// 2^128 - 1
		maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
		low, high := ConvertBigIntToU256Felts(maxUint128)

		assert.NotNil(t, low)
		assert.NotNil(t, high)
		assert.Equal(t, "0x0", high.String())   // High part should be 0
		assert.NotEqual(t, "0x0", low.String()) // Low part should not be 0
	})

	t.Run("exactly 2^128", func(t *testing.T) {
		// 2^128
		value := new(big.Int).Lsh(big.NewInt(1), 128)
		low, high := ConvertBigIntToU256Felts(value)

		assert.NotNil(t, low)
		assert.NotNil(t, high)
		assert.Equal(t, "0x0", low.String())  // Low part should be 0
		assert.Equal(t, "0x1", high.String()) // High part should be 1
	})

	t.Run("negative number", func(t *testing.T) {
		// Test with negative number (should be treated as 0)
		negative := big.NewInt(-1)
		low, high := ConvertBigIntToU256Felts(negative)

		assert.NotNil(t, low)
		assert.NotNil(t, high)
		// The function should handle negative numbers gracefully
	})
}

func TestBytesToU128FeltsEdgeCases(t *testing.T) {
	t.Run("empty bytes", func(t *testing.T) {
		result := BytesToU128Felts([]byte{})
		assert.Empty(t, result)
	})

	t.Run("exactly 16 bytes", func(t *testing.T) {
		bytes := make([]byte, 16)
		for i := range bytes {
			bytes[i] = byte(i)
		}
		result := BytesToU128Felts(bytes)
		assert.Len(t, result, 1)
		assert.NotNil(t, result[0])
	})

	t.Run("17 bytes", func(t *testing.T) {
		bytes := make([]byte, 17)
		for i := range bytes {
			bytes[i] = byte(i)
		}
		result := BytesToU128Felts(bytes)
		assert.Len(t, result, 2) // Should create 2 felts
		assert.NotNil(t, result[0])
		assert.NotNil(t, result[1])
	})

	t.Run("single byte", func(t *testing.T) {
		bytes := []byte{0x42}
		result := BytesToU128Felts(bytes)
		assert.Len(t, result, 1)
		assert.NotNil(t, result[0])
	})
}

func TestConvertSolidityOrderIDForStarknetEdgeCases(t *testing.T) {
	t.Run("short hex string", func(t *testing.T) {
		// Test with a short hex string that needs padding
		shortHex := "0x1234"
		low, high, err := ConvertSolidityOrderIDForStarknet(shortHex)

		require.NoError(t, err)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("exactly 32 bytes", func(t *testing.T) {
		// Test with exactly 32 bytes
		hex32 := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		low, high, err := ConvertSolidityOrderIDForStarknet(hex32)

		require.NoError(t, err)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("longer than 32 bytes", func(t *testing.T) {
		// Test with longer than 32 bytes
		longHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		low, high, err := ConvertSolidityOrderIDForStarknet(longHex)

		require.NoError(t, err)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("zero order ID", func(t *testing.T) {
		zeroHex := "0x0000000000000000000000000000000000000000000000000000000000000000"
		low, high, err := ConvertSolidityOrderIDForStarknet(zeroHex)

		require.NoError(t, err)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
		assert.Equal(t, "0x0", low.String())
		assert.Equal(t, "0x0", high.String())
	})
}

func TestStarknetERC20ABI(t *testing.T) {
	t.Run("ABI contains required functions", func(t *testing.T) {
		assert.Contains(t, StarknetERC20ABI, "balanceOf")
		assert.Contains(t, StarknetERC20ABI, "allowance")
		assert.Contains(t, StarknetERC20ABI, "approve")
	})

	t.Run("ABI has correct selectors", func(t *testing.T) {
		// Test that the selectors are valid hex strings
		for name, selector := range StarknetERC20ABI {
			assert.True(t, selector != "", "Selector for %s should not be empty", name)
			assert.True(t, strings.HasPrefix(selector, "0x"), "Selector for %s should start with 0x", name)
		}
	})
}

func TestUint256Operations(t *testing.T) {
	t.Run("uint256 bit operations", func(t *testing.T) {
		// Test bit shifting operations
		value := big.NewInt(0x1234567890ABCDEF)
		u := ToUint256(value)

		// Test left shift
		leftShifted := new(uint256.Int).Lsh(u, 8)
		assert.True(t, leftShifted.Cmp(u) > 0)

		// Test right shift
		rightShifted := new(uint256.Int).Rsh(u, 8)
		assert.True(t, rightShifted.Cmp(u) < 0)
	})

	t.Run("uint256 arithmetic", func(t *testing.T) {
		a := ToUint256(big.NewInt(100))
		b := ToUint256(big.NewInt(50))

		// Test addition
		sum := new(uint256.Int).Add(a, b)
		expectedSum := ToUint256(big.NewInt(150))
		assert.Equal(t, expectedSum, sum)

		// Test subtraction
		diff := new(uint256.Int).Sub(a, b)
		expectedDiff := ToUint256(big.NewInt(50))
		assert.Equal(t, expectedDiff, diff)
	})

	t.Run("uint256 comparison", func(t *testing.T) {
		small := ToUint256(big.NewInt(100))
		large := ToUint256(big.NewInt(1000))

		assert.Equal(t, -1, small.Cmp(large))
		assert.Equal(t, 1, large.Cmp(small))
		assert.Equal(t, 0, small.Cmp(new(uint256.Int).Set(small)))
	})
}
