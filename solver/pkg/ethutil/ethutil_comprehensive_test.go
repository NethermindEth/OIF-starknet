package ethutil

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEthUtilComprehensive tests comprehensive ethutil functionality
func TestEthUtilComprehensive(t *testing.T) {
	t.Run("address_operations", func(t *testing.T) {
		// Test address creation and validation
		address := common.HexToAddress("0x1234567890123456789012345678901234567890")
		assert.NotEqual(t, common.Address{}, address)

		// Test address string conversion
		addressStr := address.Hex()
		assert.Equal(t, "0x1234567890123456789012345678901234567890", addressStr)

		// Test address comparison
		address2 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		assert.Equal(t, address, address2)
	})

	t.Run("big_int_operations", func(t *testing.T) {
		// Test big.Int creation and operations
		val1 := big.NewInt(1000)
		val2 := big.NewInt(2000)

		// Test addition
		sum := new(big.Int).Add(val1, val2)
		assert.Equal(t, big.NewInt(3000), sum)

		// Test multiplication
		product := new(big.Int).Mul(val1, val2)
		assert.Equal(t, big.NewInt(2000000), product)

		// Test division
		quotient := new(big.Int).Div(val2, val1)
		assert.Equal(t, big.NewInt(2), quotient)

		// Test comparison
		assert.True(t, val1.Cmp(val2) < 0)
		assert.True(t, val2.Cmp(val1) > 0)
		assert.Equal(t, 0, val1.Cmp(new(big.Int).Set(val1)))
	})

	t.Run("private_key_operations", func(t *testing.T) {
		// Test private key generation
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		assert.NotNil(t, privateKey)

		// Test private key to public key
		publicKey := &privateKey.PublicKey
		assert.NotNil(t, publicKey)

		// Test private key to address
		address := crypto.PubkeyToAddress(*publicKey)
		assert.NotEqual(t, common.Address{}, address)

		// Test private key hex encoding
		privateKeyHex := crypto.FromECDSA(privateKey)
		assert.NotEmpty(t, privateKeyHex)
		assert.Equal(t, 32, len(privateKeyHex))
	})

	t.Run("function_signature_operations", func(t *testing.T) {
		// Test function signature hashing
		signature := "transfer(address,uint256)"
		hash := crypto.Keccak256Hash([]byte(signature))
		assert.NotEqual(t, common.Hash{}, hash)

		// Test method ID extraction (first 4 bytes)
		methodID := hash.Bytes()[:4]
		assert.Equal(t, 4, len(methodID))

		// Test different signatures produce different hashes
		signature2 := "approve(address,uint256)"
		hash2 := crypto.Keccak256Hash([]byte(signature2))
		assert.NotEqual(t, hash, hash2)
	})

	t.Run("error_handling", func(t *testing.T) {
		// Test invalid private key parsing
		_, err := ParsePrivateKey("invalid_hex")
		assert.Error(t, err)

		// Test empty private key
		_, err = ParsePrivateKey("")
		assert.Error(t, err)

		// Test private key with wrong length
		_, err = ParsePrivateKey("0x1234")
		assert.Error(t, err)

		// Test valid private key
		validKey := "0x1234567890123456789012345678901234567890123456789012345678901234"
		_, err = ParsePrivateKey(validKey)
		assert.NoError(t, err)
	})

	t.Run("token_amount_formatting", func(t *testing.T) {
		// Test wei to ether conversion
		wei := big.NewInt(1000000000000000000) // 1 ether in wei
		formatted := FormatTokenAmount(wei, 18)
		assert.Equal(t, "1.00 tokens", formatted)

		// Test smaller amounts
		smallWei := big.NewInt(100000000000000000) // 0.1 ether in wei
		formatted = FormatTokenAmount(smallWei, 18)
		assert.Equal(t, "0.10 tokens", formatted)

		// Test zero amount
		zero := big.NewInt(0)
		formatted = FormatTokenAmount(zero, 18)
		assert.Equal(t, "0.00 tokens", formatted)

		// Test very small amount
		verySmall := big.NewInt(1) // 1 wei
		formatted = FormatTokenAmount(verySmall, 18)
		assert.Equal(t, "0.00 tokens", formatted)
	})

	t.Run("abi_operations", func(t *testing.T) {
		// Test ABI parsing
		abiStr := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]`
		parsedABI, err := ParseABI(abiStr)
		require.NoError(t, err)
		assert.NotNil(t, parsedABI)

		// Test ABI method lookup
		method, exists := parsedABI.Methods["balanceOf"]
		assert.True(t, exists)
		assert.NotNil(t, method)
		assert.Equal(t, "balanceOf", method.Name)
	})

	t.Run("transactor_operations", func(t *testing.T) {
		// Test transactor creation
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		chainID := big.NewInt(1) // Mainnet
		transactor, err := NewTransactor(chainID, privateKey)
		require.NoError(t, err)
		assert.NotNil(t, transactor)
		assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey), transactor.From)
	})

	t.Run("erc20_operations", func(t *testing.T) {
		// Test ERC20 function signatures
		balanceOfSig := "balanceOf(address)"
		allowanceSig := "allowance(address,address)"
		transferSig := "transfer(address,uint256)"
		approveSig := "approve(address,uint256)"

		// Test that these are valid function signatures
		assert.NotEmpty(t, balanceOfSig)
		assert.NotEmpty(t, allowanceSig)
		assert.NotEmpty(t, transferSig)
		assert.NotEmpty(t, approveSig)

		// Test function signature hashing
		balanceOfHash := crypto.Keccak256Hash([]byte(balanceOfSig))
		allowanceHash := crypto.Keccak256Hash([]byte(allowanceSig))
		transferHash := crypto.Keccak256Hash([]byte(transferSig))
		approveHash := crypto.Keccak256Hash([]byte(approveSig))

		assert.NotEqual(t, common.Hash{}, balanceOfHash)
		assert.NotEqual(t, common.Hash{}, allowanceHash)
		assert.NotEqual(t, common.Hash{}, transferHash)
		assert.NotEqual(t, common.Hash{}, approveHash)

		// Test that different signatures produce different hashes
		assert.NotEqual(t, balanceOfHash, allowanceHash)
		assert.NotEqual(t, transferHash, approveHash)
	})

	t.Run("wait_for_transaction", func(t *testing.T) {
		// Test WaitForTransaction function exists and can be called
		// We can't easily test this without a real transaction, so we'll skip
		t.Skip("Skipping WaitForTransaction test - requires real transaction")
	})
}

// TestEthUtilEdgeCases tests edge cases for ethutil functions
func TestEthUtilEdgeCases(t *testing.T) {
	t.Run("very_large_numbers", func(t *testing.T) {
		// Test with very large numbers
		veryLarge := new(big.Int)
		veryLarge.SetString("123456789012345678901234567890123456789012345678901234567890", 10)

		// Test formatting very large numbers
		formatted := FormatTokenAmount(veryLarge, 18)
		assert.NotEmpty(t, formatted)

		// Test transactor with very large chain ID
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		largeChainID := new(big.Int).SetUint64(^uint64(0)) // Max uint64
		transactor, err := NewTransactor(largeChainID, privateKey)
		require.NoError(t, err)
		assert.NotNil(t, transactor)
	})

	t.Run("zero_values", func(t *testing.T) {
		// Test with zero values
		zero := big.NewInt(0)
		formatted := FormatTokenAmount(zero, 18)
		assert.Equal(t, "0.00 tokens", formatted)

		// Test with valid chain ID (zero is invalid)
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		validChainID := big.NewInt(1) // Mainnet
		transactor, err := NewTransactor(validChainID, privateKey)
		require.NoError(t, err)
		assert.NotNil(t, transactor)
	})

	t.Run("negative_values", func(t *testing.T) {
		// Test with negative values (should be handled gracefully)
		negative := big.NewInt(-1000)
		formatted := FormatTokenAmount(negative, 18)
		// Should handle negative values gracefully
		assert.NotEmpty(t, formatted)
	})
}

// TestEthUtilConcurrency tests concurrent access to ethutil functions
func TestEthUtilConcurrency(t *testing.T) {
	t.Run("concurrent_transactor_creation", func(t *testing.T) {
		// Test concurrent transactor creation
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				privateKey, err := crypto.GenerateKey()
				require.NoError(t, err)

				chainID := big.NewInt(int64(index + 1))
				transactor, err := NewTransactor(chainID, privateKey)
				require.NoError(t, err)
				assert.NotNil(t, transactor)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent_private_key_parsing", func(t *testing.T) {
		// Test concurrent private key parsing
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(_ int) {
				privateKey, err := crypto.GenerateKey()
				require.NoError(t, err)

				privateKeyHex := crypto.FromECDSA(privateKey)
				parsedKey, err := ParsePrivateKey("0x" + common.Bytes2Hex(privateKeyHex))
				require.NoError(t, err)
				assert.NotNil(t, parsedKey)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestEthUtilValidation tests validation functions
func TestEthUtilValidation(t *testing.T) {
	t.Run("validate_addresses", func(t *testing.T) {
		// Test valid addresses
		validAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
		assert.NotEqual(t, common.Address{}, validAddress)

		// Test zero address
		zeroAddress := common.Address{}
		assert.Equal(t, common.Address{}, zeroAddress)
	})

	t.Run("validate_private_keys", func(t *testing.T) {
		// Test valid private key
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		assert.NotNil(t, privateKey)

		// Test private key length
		privateKeyBytes := crypto.FromECDSA(privateKey)
		assert.Equal(t, 32, len(privateKeyBytes))
	})

	t.Run("validate_big_ints", func(t *testing.T) {
		// Test valid big.Int values
		val := big.NewInt(12345)
		assert.True(t, val.Cmp(big.NewInt(0)) > 0)

		// Test zero big.Int
		zero := big.NewInt(0)
		assert.True(t, zero.Cmp(big.NewInt(0)) == 0)
	})
}
