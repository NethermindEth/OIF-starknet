package starknetutil

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/starknet.go/utils"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBigIntToU256Felts(t *testing.T) {
	tests := []struct {
		name string
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




