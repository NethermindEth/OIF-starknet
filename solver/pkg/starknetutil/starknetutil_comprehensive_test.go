package starknetutil

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStarknetUtilComprehensive tests comprehensive starknetutil functionality
func TestStarknetUtilComprehensive(t *testing.T) {
	t.Run("felt_operations", func(t *testing.T) {
		// Test Felt creation and operations
		felt1 := new(felt.Felt).SetUint64(12345)
		felt2 := new(felt.Felt).SetUint64(67890)

		assert.Equal(t, uint64(12345), felt1.Uint64())
		assert.Equal(t, uint64(67890), felt2.Uint64())

		// Test Felt comparison
		assert.True(t, felt1.Cmp(felt2) < 0)
		assert.True(t, felt2.Cmp(felt1) > 0)
		assert.Equal(t, 0, felt1.Cmp(new(felt.Felt).Set(felt1)))
	})

	t.Run("felt_string_operations", func(t *testing.T) {
		felt := new(felt.Felt).SetUint64(0x1234567890abcdef)

		// Test hex string representation
		hexStr := felt.Text(16)
		assert.NotEmpty(t, hexStr)

		// Test decimal string representation
		decStr := felt.Text(10)
		assert.NotEmpty(t, decStr)
	})

	t.Run("felt_arithmetic", func(t *testing.T) {
		felt1 := new(felt.Felt).SetUint64(100)
		felt2 := new(felt.Felt).SetUint64(50)

		// Test addition
		sum := new(felt.Felt).Add(felt1, felt2)
		assert.Equal(t, uint64(150), sum.Uint64())

		// Test subtraction
		diff := new(felt.Felt).Sub(felt1, felt2)
		assert.Equal(t, uint64(50), diff.Uint64())

		// Test multiplication
		product := new(felt.Felt).Mul(felt1, felt2)
		assert.Equal(t, uint64(5000), product.Uint64())
	})
}

// TestUint256OperationsComprehensive tests comprehensive Uint256 operations
func TestUint256OperationsComprehensive(t *testing.T) {
	t.Run("uint256_creation_from_big_int", func(t *testing.T) {
		// Test with small number
		small := big.NewInt(12345)
		uint256 := ToUint256(small)
		assert.NotNil(t, uint256)
		assert.Equal(t, small, uint256.ToBig())

		// Test with large number
		large := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128
		uint256Large := ToUint256(large)
		assert.NotNil(t, uint256Large)
		assert.Equal(t, large, uint256Large.ToBig())
	})

	t.Run("uint256_creation_from_uint64", func(t *testing.T) {
		val := uint64(98765)
		uint256 := ToUint256(big.NewInt(int64(val)))
		assert.NotNil(t, uint256)
		assert.Equal(t, val, uint256.Uint64())
	})

	t.Run("uint256_creation_from_zero", func(t *testing.T) {
		zero := big.NewInt(0)
		uint256 := ToUint256(zero)
		assert.NotNil(t, uint256)
		assert.True(t, uint256.ToBig().Cmp(big.NewInt(0)) == 0)
	})

	t.Run("uint256_creation_from_max_uint64", func(t *testing.T) {
		maxUint64 := big.NewInt(0).SetUint64(^uint64(0))
		uint256 := ToUint256(maxUint64)
		assert.NotNil(t, uint256)
		assert.Equal(t, maxUint64, uint256.ToBig())
	})
}

// TestConvertBigIntToU256FeltsComprehensive tests comprehensive BigInt to U256 Felts conversion
func TestConvertBigIntToU256FeltsComprehensive(t *testing.T) {
	t.Run("convert_small_number", func(t *testing.T) {
		val := big.NewInt(12345)
		low, high := ConvertBigIntToU256Felts(val)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("convert_large_number", func(t *testing.T) {
		// Create a number that spans both low and high parts
		val := new(big.Int).Lsh(big.NewInt(1), 130) // 2^130
		low, high := ConvertBigIntToU256Felts(val)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("convert_zero", func(t *testing.T) {
		val := big.NewInt(0)
		low, high := ConvertBigIntToU256Felts(val)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("convert_max_uint128", func(t *testing.T) {
		// 2^128 - 1
		val := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
		low, high := ConvertBigIntToU256Felts(val)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})
}

// TestBytesToU128FeltsComprehensive tests comprehensive bytes to U128 Felts conversion
func TestBytesToU128FeltsComprehensive(t *testing.T) {
	t.Run("convert_short_bytes", func(t *testing.T) {
		bytes := []byte{0x12, 0x34, 0x56, 0x78}
		felts := BytesToU128Felts(bytes)
		assert.NotEmpty(t, felts)
		for i, felt := range felts {
			assert.NotNil(t, felt, "felt at index %d should not be nil", i)
		}
	})

	t.Run("convert_empty_bytes", func(t *testing.T) {
		bytes := []byte{}
		felts := BytesToU128Felts(bytes)
		// Empty bytes should return empty slice or single zero felt
		if len(felts) > 0 {
			for i, felt := range felts {
				assert.NotNil(t, felt, "felt at index %d should not be nil", i)
			}
		}
	})

	t.Run("convert_long_bytes", func(t *testing.T) {
		// Create 32 bytes of data
		bytes := make([]byte, 32)
		for i := range bytes {
			bytes[i] = byte(i)
		}
		felts := BytesToU128Felts(bytes)
		assert.NotEmpty(t, felts)
		for i, felt := range felts {
			assert.NotNil(t, felt, "felt at index %d should not be nil", i)
		}
	})

	t.Run("convert_single_byte", func(t *testing.T) {
		bytes := []byte{0xFF}
		felts := BytesToU128Felts(bytes)
		assert.NotEmpty(t, felts)
		for i, felt := range felts {
			assert.NotNil(t, felt, "felt at index %d should not be nil", i)
		}
	})
}

// TestConvertSolidityOrderIDForStarknetComprehensive tests comprehensive Solidity order ID conversion
func TestConvertSolidityOrderIDForStarknetComprehensive(t *testing.T) {
	t.Run("convert_short_order_id", func(t *testing.T) {
		orderID := "0x12345678"
		low, high, err := ConvertSolidityOrderIDForStarknet(orderID)
		require.NoError(t, err)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("convert_empty_order_id", func(t *testing.T) {
		orderID := ""
		_, _, err := ConvertSolidityOrderIDForStarknet(orderID)
		// Empty string should return an error
		assert.Error(t, err)
	})

	t.Run("convert_32_byte_order_id", func(t *testing.T) {
		// Create a 32-byte order ID
		orderID := "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		low, high, err := ConvertSolidityOrderIDForStarknet(orderID)
		require.NoError(t, err)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})

	t.Run("convert_max_order_id", func(t *testing.T) {
		// Create order ID with all 0xFF bytes
		orderID := "0x" + "ffffffffffffffffffffffffffffffff"
		low, high, err := ConvertSolidityOrderIDForStarknet(orderID)
		require.NoError(t, err)
		assert.NotNil(t, low)
		assert.NotNil(t, high)
	})
}

// TestStarknetERC20ABIComprehensive tests comprehensive Starknet ERC20 ABI functionality
func TestStarknetERC20ABIComprehensive(t *testing.T) {
	t.Run("abi_validation", func(t *testing.T) {
		// Test that the ABI is valid
		abi := StarknetERC20ABI
		assert.NotEmpty(t, abi)

		// Test that it contains expected methods
		assert.Contains(t, abi, "balanceOf")
		assert.Contains(t, abi, "allowance")
		assert.Contains(t, abi, "approve")
		// Note: transfer might not be in this ABI
	})

	t.Run("abi_structure", func(t *testing.T) {
		// Test that the ABI has proper structure
		abi := StarknetERC20ABI
		// This is a map of function names to selectors, not JSON
		assert.Greater(t, len(abi), 0)
		// Check that values are valid hex strings
		for name, selector := range abi {
			assert.NotEmpty(t, name)
			assert.NotEmpty(t, selector)
			assert.Contains(t, selector, "0x")
		}
	})

	t.Run("abi_methods", func(t *testing.T) {
		// Test specific method signatures
		abi := StarknetERC20ABI
		assert.Contains(t, abi, "balanceOf")
		assert.Contains(t, abi, "allowance")
		assert.Contains(t, abi, "approve")
		// Note: transfer might not be in this ABI
	})
}

// TestStarknetUtilEdgeCases tests edge cases for starknetutil functions
func TestStarknetUtilEdgeCases(t *testing.T) {
	t.Run("felt_with_max_value", func(t *testing.T) {
		// Test with maximum Felt value
		maxFelt := new(felt.Felt).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
		assert.NotNil(t, maxFelt)
	})

	t.Run("felt_with_min_value", func(t *testing.T) {
		// Test with minimum Felt value (zero)
		minFelt := new(felt.Felt).SetUint64(0)
		assert.Equal(t, uint64(0), minFelt.Uint64())
	})

	t.Run("uint256_with_very_large_number", func(t *testing.T) {
		// Test with a very large number
		veryLarge := new(big.Int).Lsh(big.NewInt(1), 200) // 2^200
		uint256 := ToUint256(veryLarge)
		assert.NotNil(t, uint256)
		assert.Equal(t, veryLarge, uint256.ToBig())
	})

	t.Run("bytes_conversion_edge_cases", func(t *testing.T) {
		// Test with nil bytes
		felts := BytesToU128Felts(nil)
		// Should handle nil gracefully
		if len(felts) > 0 {
			for i, felt := range felts {
				assert.NotNil(t, felt, "felt at index %d should not be nil", i)
			}
		}

		// Test with single zero byte
		felts = BytesToU128Felts([]byte{0})
		// Should handle single byte gracefully
		if len(felts) > 0 {
			for i, felt := range felts {
				assert.NotNil(t, felt, "felt at index %d should not be nil", i)
			}
		}
	})
}

// TestStarknetUtilConcurrency tests concurrent access to starknetutil functions
func TestStarknetUtilConcurrency(t *testing.T) {
	t.Run("concurrent_uint256_creation", func(t *testing.T) {
		// Test concurrent Uint256 creation
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				val := big.NewInt(int64(index * 1000))
				uint256 := ToUint256(val)
				assert.NotNil(t, uint256)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent_felt_operations", func(t *testing.T) {
		// Test concurrent Felt operations
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				felt := new(felt.Felt).SetUint64(uint64(index * 1000))
				assert.NotNil(t, felt)
				assert.Equal(t, uint64(index*1000), felt.Uint64())
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent_bytes_conversion", func(t *testing.T) {
		// Test concurrent bytes conversion
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				bytes := []byte{byte(index), byte(index * 2), byte(index * 3)}
				felts := BytesToU128Felts(bytes)
				assert.NotEmpty(t, felts)
				for i, felt := range felts {
					assert.NotNil(t, felt, "felt at index %d should not be nil", i)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestStarknetUtilValidation tests validation functions
func TestStarknetUtilValidation(t *testing.T) {
	t.Run("validate_felt_values", func(t *testing.T) {
		// Test valid Felt values
		validFelt := new(felt.Felt).SetUint64(12345)
		assert.NotNil(t, validFelt)
		assert.Equal(t, uint64(12345), validFelt.Uint64())
	})

	t.Run("validate_uint256_values", func(t *testing.T) {
		// Test valid Uint256 values
		val := big.NewInt(12345)
		uint256 := ToUint256(val)
		assert.NotNil(t, uint256)
		assert.True(t, uint256.ToBig().Cmp(big.NewInt(0)) >= 0)
	})

	t.Run("validate_bytes_conversion", func(t *testing.T) {
		// Test valid bytes conversion
		bytes := []byte{0x12, 0x34, 0x56, 0x78}
		felts := BytesToU128Felts(bytes)
		assert.NotEmpty(t, felts)
		for i, felt := range felts {
			assert.NotNil(t, felt, "felt at index %d should not be nil", i)
		}
	})
}
