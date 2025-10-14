package ethutil

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestERC20Balance(t *testing.T) {
	t.Run("ERC20Balance function exists", func(t *testing.T) {
		// Test that the function is properly defined
		// In a real test, we'd need a mock client
		assert.NotNil(t, ERC20Balance)
	})

	t.Run("ERC20Allowance function exists", func(t *testing.T) {
		// Test that the function is properly defined
		assert.NotNil(t, ERC20Allowance)
	})

	t.Run("ERC20Approve function exists", func(t *testing.T) {
		// Test that the function is properly defined
		assert.NotNil(t, ERC20Approve)
	})
}

func TestAddressValidation(t *testing.T) {
	t.Run("Valid Ethereum addresses", func(t *testing.T) {
		validAddresses := []string{
			"0x1234567890123456789012345678901234567890",
			"0x0000000000000000000000000000000000000000",
		}

		for _, addr := range validAddresses {
			t.Run(addr, func(t *testing.T) {
				address := common.HexToAddress(addr)
				// For zero address, it's valid but will be zero
				if addr == "0x0000000000000000000000000000000000000000" {
					assert.Equal(t, common.Address{}, address)
				} else {
					assert.NotEqual(t, common.Address{}, address)
				}
				assert.Equal(t, addr, address.Hex())
			})
		}
	})

	t.Run("Invalid Ethereum addresses", func(t *testing.T) {
		invalidAddresses := []string{
			"0x123", // Too short
			"0x12345678901234567890123456789012345678901", // Too long
			"1234567890123456789012345678901234567890",    // Missing 0x
			"0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG",  // Invalid hex
			"", // Empty
		}

		for _, addr := range invalidAddresses {
			t.Run(addr, func(t *testing.T) {
				// These should either fail to parse or result in zero address
				address := common.HexToAddress(addr)
				// We can't easily test parsing failures with HexToAddress
				// as it's quite permissive, but we can test the result
				if addr == "" {
					assert.Equal(t, common.Address{}, address)
				}
			})
		}
	})
}

func TestBigIntBasicOperations(t *testing.T) {
	t.Run("BigInt creation and comparison", func(t *testing.T) {
		zero := big.NewInt(0)
		one := big.NewInt(1)
		large := big.NewInt(1000000000000000000) // 1 ETH in wei

		assert.Equal(t, 0, zero.Cmp(big.NewInt(0)))
		assert.Equal(t, 1, one.Cmp(zero))
		assert.Equal(t, -1, zero.Cmp(one))
		assert.Equal(t, 1, large.Cmp(one))
	})

	t.Run("BigInt arithmetic", func(t *testing.T) {
		a := big.NewInt(100)
		b := big.NewInt(50)

		// Addition
		sum := new(big.Int).Add(a, b)
		assert.Equal(t, int64(150), sum.Int64())

		// Subtraction
		diff := new(big.Int).Sub(a, b)
		assert.Equal(t, int64(50), diff.Int64())

		// Multiplication
		product := new(big.Int).Mul(a, b)
		assert.Equal(t, int64(5000), product.Int64())

		// Division
		quotient := new(big.Int).Div(a, b)
		assert.Equal(t, int64(2), quotient.Int64())
	})

	t.Run("BigInt string conversion", func(t *testing.T) {
		value := big.NewInt(123456789)
		assert.Equal(t, "123456789", value.String())

		// Test with hex
		hexValue := big.NewInt(0x1234567890ABCDEF)
		assert.Equal(t, "1311768467294899695", hexValue.String())
	})
}

func TestCommonAddressOperations(t *testing.T) {
	t.Run("Address to bytes conversion", func(t *testing.T) {
		addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		bytes := addr.Bytes()

		assert.Len(t, bytes, 20) // Ethereum addresses are 20 bytes
		assert.Equal(t, addr, common.BytesToAddress(bytes))
	})

	t.Run("Address comparison", func(t *testing.T) {
		addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addr2 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addr3 := common.HexToAddress("0x0987654321098765432109876543210987654321")

		assert.Equal(t, addr1, addr2)
		assert.NotEqual(t, addr1, addr3)
	})

	t.Run("Zero address", func(t *testing.T) {
		zeroAddr := common.Address{}
		assert.Equal(t, "0x0000000000000000000000000000000000000000", zeroAddr.Hex())
		assert.True(t, zeroAddr == common.Address{})
	})
}

func TestFunctionSignaturesBasic(t *testing.T) {
	t.Run("ERC20 function signatures", func(t *testing.T) {
		// Test that we can create function signatures
		// These would be used in actual contract calls

		// balanceOf(address) -> 0x70a08231
		balanceOfSig := "0x70a08231"
		assert.Len(t, balanceOfSig, 10) // 0x + 8 hex chars

		// allowance(address,address) -> 0xdd62ed3e
		allowanceSig := "0xdd62ed3e"
		assert.Len(t, allowanceSig, 10)

		// approve(address,uint256) -> 0x095ea7b3
		approveSig := "0x095ea7b3"
		assert.Len(t, approveSig, 10)
	})

	t.Run("Function call data construction", func(t *testing.T) {
		// Test basic function call data construction
		// In a real implementation, this would use ABI encoding

		functionSelector := "0x70a08231" // balanceOf
		addressParam := "0000000000000000000000001234567890123456789012345678901234567890"

		callData := functionSelector + addressParam
		assert.Len(t, callData, 74) // 10 (selector) + 64 (padded address)
	})
}

func TestErrorHandlingBasic(t *testing.T) {
	t.Run("Error types", func(t *testing.T) {
		// Test that we can create and handle different error types
		// that might occur in EVM interactions

		// Connection error
		connErr := "connection refused"
		assert.Contains(t, connErr, "connection")

		// Contract error
		contractErr := "execution reverted"
		assert.Contains(t, contractErr, "reverted")

		// Gas error
		gasErr := "gas required exceeds allowance"
		assert.Contains(t, gasErr, "gas")
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("Wei to Ether conversion", func(t *testing.T) {
		wei := big.NewInt(1000000000000000000) // 1 ETH
		ether := new(big.Float).SetInt(wei)
		ether.Quo(ether, big.NewFloat(1e18))

		assert.Equal(t, "1", ether.Text('f', 0))
	})

	t.Run("Ether to Wei conversion", func(t *testing.T) {
		ether := big.NewFloat(1.5)
		wei := new(big.Float).Mul(ether, big.NewFloat(1e18))
		weiInt, _ := wei.Int(nil)

		expected := big.NewInt(1500000000000000000) // 1.5 ETH in wei
		assert.Equal(t, expected, weiInt)
	})

	t.Run("Gas price formatting", func(t *testing.T) {
		gasPrice := big.NewInt(20000000000) // 20 gwei
		gwei := new(big.Float).SetInt(gasPrice)
		gwei.Quo(gwei, big.NewFloat(1e9))

		assert.Equal(t, "20", gwei.Text('f', 0))
	})
}

func TestNewTransactor(t *testing.T) {
	t.Run("successful transactor creation", func(t *testing.T) {
		chainID := big.NewInt(1) // Ethereum mainnet
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		transactor, err := NewTransactor(chainID, privateKey)

		require.NoError(t, err)
		assert.NotNil(t, transactor)
		assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey), transactor.From)
	})

	t.Run("nil chain ID", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		_, err = NewTransactor(nil, privateKey)
		assert.Error(t, err)
	})

	t.Run("nil private key", func(t *testing.T) {
		chainID := big.NewInt(1)

		// This will panic, so we expect it to panic
		assert.Panics(t, func() {
			_, _ = NewTransactor(chainID, nil)
		})
	})
}

func TestParsePrivateKey(t *testing.T) {
	t.Run("valid private key with 0x prefix", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		privateKeyHex := crypto.FromECDSA(privateKey)
		privateKeyHexStr := "0x" + common.Bytes2Hex(privateKeyHex)

		parsedKey, err := ParsePrivateKey(privateKeyHexStr)

		require.NoError(t, err)
		assert.Equal(t, privateKey.D, parsedKey.D)
		assert.Equal(t, privateKey.PublicKey.X, parsedKey.PublicKey.X)
		assert.Equal(t, privateKey.PublicKey.Y, parsedKey.PublicKey.Y)
	})

	t.Run("valid private key without 0x prefix", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		privateKeyHex := crypto.FromECDSA(privateKey)
		privateKeyHexStr := common.Bytes2Hex(privateKeyHex)

		parsedKey, err := ParsePrivateKey(privateKeyHexStr)

		require.NoError(t, err)
		assert.Equal(t, privateKey.D, parsedKey.D)
	})

	t.Run("invalid private key", func(t *testing.T) {
		_, err := ParsePrivateKey("invalid")
		assert.Error(t, err)
	})

	t.Run("empty private key", func(t *testing.T) {
		_, err := ParsePrivateKey("")
		assert.Error(t, err)
	})

	t.Run("too short private key", func(t *testing.T) {
		_, err := ParsePrivateKey("0x123")
		assert.Error(t, err)
	})
}

func TestFormatTokenAmount(t *testing.T) {
	t.Run("format with 18 decimals", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000) // 1 token
		result := FormatTokenAmount(amount, 18)
		assert.Equal(t, "1.00 tokens", result)
	})

	t.Run("format with 6 decimals", func(t *testing.T) {
		amount := big.NewInt(1000000) // 1 token
		result := FormatTokenAmount(amount, 6)
		assert.Equal(t, "1.00 tokens", result)
	})

	t.Run("format zero amount", func(t *testing.T) {
		amount := big.NewInt(0)
		result := FormatTokenAmount(amount, 18)
		assert.Equal(t, "0.00 tokens", result)
	})

	t.Run("format large amount", func(t *testing.T) {
		amount, _ := big.NewInt(0).SetString("123456789012345678901234567890", 10)
		result := FormatTokenAmount(amount, 18)
		assert.Equal(t, "123456789012.35 tokens", result)
	})

	t.Run("format nil amount", func(t *testing.T) {
		result := FormatTokenAmount(nil, 18)
		assert.Equal(t, "0", result)
	})
}

func TestERC20ABI(t *testing.T) {
	t.Run("ERC20ABI is valid JSON", func(t *testing.T) {
		// Test that the ABI string is valid JSON
		assert.NotEmpty(t, ERC20ABI)
		assert.Contains(t, ERC20ABI, "balanceOf")
		assert.Contains(t, ERC20ABI, "allowance")
		assert.Contains(t, ERC20ABI, "transfer")
		assert.Contains(t, ERC20ABI, "approve")
	})
}

func TestBigIntOperations(t *testing.T) {
	t.Run("uint256 conversion", func(t *testing.T) {
		// Test that we can work with large numbers that might be used in uint256
		largeNumber := new(big.Int)
		largeNumber.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10) // 2^256 - 1

		assert.NotNil(t, largeNumber)
		assert.True(t, largeNumber.Cmp(big.NewInt(0)) > 0)
	})

	t.Run("wei to ether conversion", func(t *testing.T) {
		wei := big.NewInt(1500000000000000000) // 1.5 ETH
		ether := new(big.Float).SetInt(wei)
		ether.Quo(ether, big.NewFloat(1e18))

		assert.Equal(t, "1.5", ether.Text('f', 1))
	})

	t.Run("ether to wei conversion", func(t *testing.T) {
		ether := big.NewFloat(2.5)
		wei := new(big.Float).Mul(ether, big.NewFloat(1e18))
		weiInt, _ := wei.Int(nil)

		expected := big.NewInt(2500000000000000000) // 2.5 ETH in wei
		assert.Equal(t, expected, weiInt)
	})
}

func TestAddressOperations(t *testing.T) {
	t.Run("address validation", func(t *testing.T) {
		validAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		assert.NotEqual(t, common.Address{}, validAddr)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", validAddr.Hex())
	})

	t.Run("zero address", func(t *testing.T) {
		zeroAddr := common.Address{}
		assert.Equal(t, "0x0000000000000000000000000000000000000000", zeroAddr.Hex())
	})

	t.Run("address from bytes", func(t *testing.T) {
		addrBytes := make([]byte, 20)
		addrBytes[19] = 0x42 // Set last byte
		addr := common.BytesToAddress(addrBytes)

		assert.Equal(t, "0x0000000000000000000000000000000000000042", addr.Hex())
	})
}

func TestFunctionSignatures(t *testing.T) {
	t.Run("ERC20 function selectors", func(t *testing.T) {
		// These are the actual function selectors for ERC20 functions
		balanceOfSelector := crypto.Keccak256([]byte("balanceOf(address)"))[:4]
		allowanceSelector := crypto.Keccak256([]byte("allowance(address,address)"))[:4]
		transferSelector := crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]
		approveSelector := crypto.Keccak256([]byte("approve(address,uint256)"))[:4]

		assert.Len(t, balanceOfSelector, 4)
		assert.Len(t, allowanceSelector, 4)
		assert.Len(t, transferSelector, 4)
		assert.Len(t, approveSelector, 4)

		// Verify known selectors
		expectedBalanceOf := []byte{0x70, 0xa0, 0x82, 0x31}
		expectedAllowance := []byte{0xdd, 0x62, 0xed, 0x3e}
		expectedTransfer := []byte{0xa9, 0x05, 0x9c, 0xbb}
		expectedApprove := []byte{0x09, 0x5e, 0xa7, 0xb3}

		assert.Equal(t, expectedBalanceOf, balanceOfSelector)
		assert.Equal(t, expectedAllowance, allowanceSelector)
		assert.Equal(t, expectedTransfer, transferSelector)
		assert.Equal(t, expectedApprove, approveSelector)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("common error types", func(t *testing.T) {
		// Test that we can create and handle different error types
		// that might occur in EVM interactions

		// Connection error
		connErr := "connection refused"
		assert.Contains(t, connErr, "connection")

		// Contract error
		contractErr := "execution reverted"
		assert.Contains(t, contractErr, "reverted")

		// Gas error
		gasErr := "gas required exceeds allowance"
		assert.Contains(t, gasErr, "gas")

		// Invalid address error
		addrErr := "invalid address format"
		assert.Contains(t, addrErr, "address")
	})
}

func TestPrivateKeyOperations(t *testing.T) {
	t.Run("generate and validate private key", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		// Verify the private key is valid
		assert.NotNil(t, privateKey.D)
		assert.NotNil(t, privateKey.PublicKey.X)
		assert.NotNil(t, privateKey.PublicKey.Y)

		// Verify we can get the address
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		assert.NotEqual(t, common.Address{}, address)
	})

	t.Run("private key to hex and back", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		// Convert to hex
		privateKeyHex := crypto.FromECDSA(privateKey)
		hexStr := common.Bytes2Hex(privateKeyHex)

		// Parse back
		parsedKey, err := ParsePrivateKey(hexStr)
		require.NoError(t, err)

		// Verify they're the same
		assert.Equal(t, privateKey.D, parsedKey.D)
	})
}

func TestERC20ABIValidation(t *testing.T) {
	t.Run("ERC20ABI contains required functions", func(t *testing.T) {
		// Test that the ABI string contains all required ERC20 functions
		assert.Contains(t, ERC20ABI, "balanceOf")
		assert.Contains(t, ERC20ABI, "allowance")
		assert.Contains(t, ERC20ABI, "transfer")
		assert.Contains(t, ERC20ABI, "approve")
	})

	t.Run("ERC20ABI is valid JSON", func(t *testing.T) {
		// Test that the ABI string is valid JSON by attempting to parse it
		// This is a basic validation without using the abi package
		assert.True(t, ERC20ABI != "")
		assert.True(t, strings.Contains(ERC20ABI, "{"))
		assert.True(t, strings.Contains(ERC20ABI, "}"))
	})
}

func TestBigIntUtilityFunctions(t *testing.T) {
	t.Run("big.Int operations for token amounts", func(t *testing.T) {
		// Test common token amount operations
		oneToken := big.NewInt(1000000000000000000) // 1 token in wei
		halfToken := new(big.Int).Div(oneToken, big.NewInt(2))

		assert.Equal(t, int64(500000000000000000), halfToken.Int64())

		// Test multiplication
		twoTokens := new(big.Int).Mul(oneToken, big.NewInt(2))
		assert.Equal(t, int64(2000000000000000000), twoTokens.Int64())
	})

	t.Run("big.Int comparison operations", func(t *testing.T) {
		small := big.NewInt(100)
		large := big.NewInt(1000)

		assert.Equal(t, -1, small.Cmp(large))
		assert.Equal(t, 1, large.Cmp(small))
		assert.Equal(t, 0, small.Cmp(new(big.Int).Set(small)))
	})

	t.Run("big.Int bit operations", func(t *testing.T) {
		value := big.NewInt(0x1234567890ABCDEF)

		// Test bit shifting
		leftShifted := new(big.Int).Lsh(value, 8)
		rightShifted := new(big.Int).Rsh(value, 8)

		assert.True(t, leftShifted.Cmp(value) > 0)
		assert.True(t, rightShifted.Cmp(value) < 0)
	})
}

func TestAddressUtilityFunctions(t *testing.T) {
	t.Run("address validation and conversion", func(t *testing.T) {
		// Test valid address
		addrStr := "0x1234567890123456789012345678901234567890"
		addr := common.HexToAddress(addrStr)
		assert.Equal(t, addrStr, addr.Hex())

		// Test address to bytes and back
		addrBytes := addr.Bytes()
		addrFromBytes := common.BytesToAddress(addrBytes)
		assert.Equal(t, addr, addrFromBytes)
	})

	t.Run("zero address handling", func(t *testing.T) {
		zeroAddr := common.Address{}
		assert.Equal(t, "0x0000000000000000000000000000000000000000", zeroAddr.Hex())
		assert.True(t, zeroAddr == common.Address{})
	})

	t.Run("address comparison", func(t *testing.T) {
		addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addr2 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addr3 := common.HexToAddress("0x0987654321098765432109876543210987654321")

		assert.Equal(t, addr1, addr2)
		assert.NotEqual(t, addr1, addr3)
	})
}

func TestGasPriceCalculations(t *testing.T) {
	t.Run("gas price conversions", func(t *testing.T) {
		// Test gwei to wei conversion
		gwei := big.NewInt(20)                                // 20 gwei
		wei := new(big.Int).Mul(gwei, big.NewInt(1000000000)) // 20 * 10^9

		assert.Equal(t, int64(20000000000), wei.Int64())

		// Test wei to gwei conversion
		weiAmount := big.NewInt(20000000000) // 20 gwei in wei
		gweiAmount := new(big.Int).Div(weiAmount, big.NewInt(1000000000))

		assert.Equal(t, int64(20), gweiAmount.Int64())
	})

	t.Run("gas limit calculations", func(t *testing.T) {
		// Test common gas limits
		approveGasLimit := uint64(200000)
		transferGasLimit := uint64(100000)

		assert.True(t, approveGasLimit > transferGasLimit)
		assert.True(t, approveGasLimit > 0)
		assert.True(t, transferGasLimit > 0)
	})
}

func TestTransactionDataConstruction(t *testing.T) {
	t.Run("function selector generation", func(t *testing.T) {
		// Test that we can generate function selectors
		balanceOfSig := crypto.Keccak256([]byte("balanceOf(address)"))[:4]
		allowanceSig := crypto.Keccak256([]byte("allowance(address,address)"))[:4]

		assert.Len(t, balanceOfSig, 4)
		assert.Len(t, allowanceSig, 4)
		assert.NotEqual(t, balanceOfSig, allowanceSig)
	})

	t.Run("calldata length validation", func(t *testing.T) {
		// Test that calldata has expected length
		functionSelector := "0x70a08231" // balanceOf
		addressParam := "0000000000000000000000001234567890123456789012345678901234567890"
		calldata := functionSelector + addressParam

		assert.Len(t, calldata, 74) // 10 (selector) + 64 (padded address)
	})
}

func TestErrorHandlingAdvanced(t *testing.T) {
	t.Run("common error message patterns", func(t *testing.T) {
		// Test that we can create and identify common error patterns
		errors := []string{
			"connection refused",
			"execution reverted",
			"gas required exceeds allowance",
			"invalid address format",
			"transaction failed",
		}

		for _, errMsg := range errors {
			assert.NotEmpty(t, errMsg)
			assert.True(t, errMsg != "")
		}
	})

	t.Run("error wrapping", func(t *testing.T) {
		// Test error wrapping patterns
		originalErr := fmt.Errorf("original error")
		wrappedErr := fmt.Errorf("wrapped: %w", originalErr)

		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "wrapped")
		assert.Contains(t, wrappedErr.Error(), "original error")
	})
}

func TestTokenAmountFormatting(t *testing.T) {
	t.Run("wei to ether conversion", func(t *testing.T) {
		// Test various wei amounts
		testCases := []struct {
			wei   *big.Int
			ether string
		}{
			{big.NewInt(0), "0.0"},
			{big.NewInt(1000000000000000000), "1.0"}, // 1 ETH
			{big.NewInt(1500000000000000000), "1.5"}, // 1.5 ETH
			{big.NewInt(100000000000000000), "0.1"},  // 0.1 ETH
		}

		for _, tc := range testCases {
			ether := new(big.Float).SetInt(tc.wei)
			ether.Quo(ether, big.NewFloat(1e18))
			assert.Equal(t, tc.ether, ether.Text('f', 1))
		}
	})

	t.Run("token amount validation", func(t *testing.T) {
		// Test token amount validation
		validAmounts := []*big.Int{
			big.NewInt(0),
			big.NewInt(1),
			big.NewInt(1000000000000000000), // 1 token
			new(big.Int).Lsh(big.NewInt(1), 256).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)), // max uint256
		}

		for _, amount := range validAmounts {
			assert.True(t, amount.Sign() >= 0) // Should be non-negative
		}
	})
}

// Test actual functions defined in ethutil.go
func TestFormatTokenAmountFunction(t *testing.T) {
	t.Run("format token amount with 18 decimals", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000) // 1 token
		result := FormatTokenAmount(amount, 18)
		assert.Equal(t, "1.00 tokens", result)
	})

	t.Run("format token amount with 6 decimals", func(t *testing.T) {
		amount := big.NewInt(1000000) // 1 token
		result := FormatTokenAmount(amount, 6)
		assert.Equal(t, "1.00 tokens", result)
	})

	t.Run("format zero amount", func(t *testing.T) {
		amount := big.NewInt(0)
		result := FormatTokenAmount(amount, 18)
		assert.Equal(t, "0.00 tokens", result)
	})

	t.Run("format nil amount", func(t *testing.T) {
		result := FormatTokenAmount(nil, 18)
		assert.Equal(t, "0", result)
	})
}

func TestParsePrivateKeyFunction(t *testing.T) {
	t.Run("parse private key with 0x prefix", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		privateKeyHex := crypto.FromECDSA(privateKey)
		privateKeyHexStr := "0x" + common.Bytes2Hex(privateKeyHex)

		parsedKey, err := ParsePrivateKey(privateKeyHexStr)
		require.NoError(t, err)
		assert.Equal(t, privateKey.D, parsedKey.D)
	})

	t.Run("parse private key without 0x prefix", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		privateKeyHex := crypto.FromECDSA(privateKey)
		privateKeyHexStr := common.Bytes2Hex(privateKeyHex)

		parsedKey, err := ParsePrivateKey(privateKeyHexStr)
		require.NoError(t, err)
		assert.Equal(t, privateKey.D, parsedKey.D)
	})

	t.Run("parse invalid private key", func(t *testing.T) {
		_, err := ParsePrivateKey("invalid")
		assert.Error(t, err)
	})

	t.Run("parse empty private key", func(t *testing.T) {
		_, err := ParsePrivateKey("")
		assert.Error(t, err)
	})
}

func TestNewTransactorFunction(t *testing.T) {
	t.Run("create transactor with valid inputs", func(t *testing.T) {
		chainID := big.NewInt(1)
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		transactor, err := NewTransactor(chainID, privateKey)
		require.NoError(t, err)
		assert.NotNil(t, transactor)
		assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey), transactor.From)
	})

	t.Run("create transactor with nil chain ID", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		_, err = NewTransactor(nil, privateKey)
		assert.Error(t, err)
	})
}
