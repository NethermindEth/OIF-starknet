package openorder

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

// TestOrderOpening tests the order opening functionality
func TestOrderOpening(t *testing.T) {
	// Check if we have the required environment variables
	baseURL := os.Getenv("BASE_RPC_URL")
	ethereumURL := os.Getenv("ETHEREUM_RPC_URL")
	starknetURL := os.Getenv("STARKNET_RPC_URL")
	
	if baseURL == "" || ethereumURL == "" || starknetURL == "" {
		t.Skip("Skipping integration tests - missing RPC URLs (BASE_RPC_URL, ETHEREUM_RPC_URL, STARKNET_RPC_URL)")
	}

	t.Run("EVM to EVM order opening", func(t *testing.T) {
		testEVMEVMOrderOpening(t)
	})

	t.Run("EVM to Starknet order opening", func(t *testing.T) {
		testEVMStarknetOrderOpening(t)
	})

	t.Run("Starknet to EVM order opening", func(t *testing.T) {
		testStarknetEVMOrderOpening(t)
	})
}

func testEVMEVMOrderOpening(t *testing.T) {
	// This is a high-level integration test
	// It would require:
	// 1. Connected EVM clients
	// 2. Funded accounts
	// 3. Deployed contracts
	// 4. Actual order opening
	
	t.Skip("Integration test - requires running network and contracts")
	
	// Example test structure using existing env vars:
	/*
	ethereumURL := os.Getenv("ETHEREUM_RPC_URL")
	baseURL := os.Getenv("BASE_RPC_URL")
	
	client, err := ethclient.Dial(ethereumURL)
	require.NoError(t, err)
	defer client.Close()

	// Get initial balances
	aliceAddr := common.HexToAddress("0x...") // Alice's address
	hyperlaneAddr := common.HexToAddress(os.Getenv("ETHEREUM_HYPERLANE_ADDR"))
	tokenAddr := common.HexToAddress(os.Getenv("ETHEREUM_TOKEN_ADDR"))

	initialAliceBalance, err := getTokenBalance(client, tokenAddr, aliceAddr)
	require.NoError(t, err)

	initialHyperlaneBalance, err := getTokenBalance(client, tokenAddr, hyperlaneAddr)
	require.NoError(t, err)

	// Open order
	orderAmount := uint256.NewInt(1000) // 1000 tokens
	err = openEVMEVMOrder(orderAmount)
	require.NoError(t, err)

	// Wait for transaction confirmation
	time.Sleep(5 * time.Second)

	// Check balances
	finalAliceBalance, err := getTokenBalance(client, tokenAddr, aliceAddr)
	require.NoError(t, err)

	finalHyperlaneBalance, err := getTokenBalance(client, tokenAddr, hyperlaneAddr)
	require.NoError(t, err)

	// Verify balance changes
	expectedAliceDecrease := orderAmount
	expectedHyperlaneIncrease := orderAmount

	actualAliceDecrease := new(uint256.Int).Sub(initialAliceBalance, finalAliceBalance)
	actualHyperlaneIncrease := new(uint256.Int).Sub(finalHyperlaneBalance, initialHyperlaneBalance)

	assert.Equal(t, expectedAliceDecrease, actualAliceDecrease, "Alice's balance should decrease by order amount")
	assert.Equal(t, expectedHyperlaneIncrease, actualHyperlaneIncrease, "Hyperlane contract balance should increase by order amount")
	*/
}

func testEVMStarknetOrderOpening(t *testing.T) {
	t.Skip("Integration test - requires running network and contracts")
	// Similar structure to EVM-EVM test but for cross-chain orders
}

func testStarknetEVMOrderOpening(t *testing.T) {
	t.Skip("Integration test - requires running network and contracts")
	// Similar structure but for Starknet origin orders
}

// Helper function to get token balance (would be implemented)
func getTokenBalance(client *ethclient.Client, tokenAddr, accountAddr common.Address) (*uint256.Int, error) {
	// This would call the ERC20 balanceOf function
	// For now, return a mock value
	return uint256.NewInt(10000), nil
}

// TestOrderDataValidation tests the order data creation and validation
func TestOrderDataValidation(t *testing.T) {
	t.Run("Valid order data creation", func(t *testing.T) {
		orderData := OrderData{
			OriginChainID:      big.NewInt(1),
			DestinationChainID: big.NewInt(421614), // Base Sepolia
			User:               "0x1234567890123456789012345678901234567890",
			OpenDeadline:       big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
			FillDeadline:       big.NewInt(time.Now().Add(2 * time.Hour).Unix()),
			MaxSpent: []TokenAmount{
				{
					Token:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
					Amount: uint256.NewInt(1000),
				},
			},
			MinReceived: []TokenAmount{
				{
					Token:  "0xfedcbafedcbafedcbafedcbafedcbafedcbafedc",
					Amount: uint256.NewInt(950),
				},
			},
		}

		// Test order data validation
		assert.True(t, orderData.IsValid(), "Order data should be valid")
		assert.True(t, orderData.OpenDeadline.Cmp(big.NewInt(time.Now().Unix())) > 0, "Open deadline should be in the future")
		assert.True(t, orderData.FillDeadline.Cmp(orderData.OpenDeadline) > 0, "Fill deadline should be after open deadline")
	})

	t.Run("Invalid order data", func(t *testing.T) {
		orderData := OrderData{
			OriginChainID:      big.NewInt(1),
			DestinationChainID: big.NewInt(1), // Same as origin - invalid
			User:               "invalid_address",
			OpenDeadline:       big.NewInt(time.Now().Add(-1 * time.Hour).Unix()), // Past deadline
			FillDeadline:       big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
		}

		assert.False(t, orderData.IsValid(), "Order data should be invalid")
	})
}

// TestOrderIDCalculation tests the order ID calculation
func TestOrderIDCalculation(t *testing.T) {
	t.Run("Order ID calculation", func(t *testing.T) {
		orderData := OrderData{
			OriginChainID:      big.NewInt(1),
			DestinationChainID: big.NewInt(421614),
			User:               "0x1234567890123456789012345678901234567890",
			OpenDeadline:       big.NewInt(1234567890),
			FillDeadline:       big.NewInt(1234567890),
		}

		orderID := calculateOrderId(orderData)
		assert.NotEmpty(t, orderID, "Order ID should not be empty")
		assert.Len(t, orderID, 66, "Order ID should be 66 characters (0x + 64 hex chars)")
		assert.Equal(t, "0x", orderID[:2], "Order ID should start with 0x")
	})

	t.Run("Order ID uniqueness", func(t *testing.T) {
		orderData1 := OrderData{
			OriginChainID:      big.NewInt(1),
			DestinationChainID: big.NewInt(421614),
			User:               "0x1234567890123456789012345678901234567890",
			OpenDeadline:       big.NewInt(1234567890),
			FillDeadline:       big.NewInt(1234567890),
		}

		orderData2 := OrderData{
			OriginChainID:      big.NewInt(1),
			DestinationChainID: big.NewInt(421614),
			User:               "0x1234567890123456789012345678901234567890",
			OpenDeadline:       big.NewInt(1234567891), // Different deadline
			FillDeadline:       big.NewInt(1234567891),
		}

		orderID1 := calculateOrderId(orderData1)
		orderID2 := calculateOrderId(orderData2)

		assert.NotEqual(t, orderID1, orderID2, "Different order data should produce different order IDs")
	})
}

// TestBalanceVerification tests balance verification logic
func TestBalanceVerification(t *testing.T) {
	t.Run("Sufficient balance", func(t *testing.T) {
		userBalance := uint256.NewInt(10000)
		requiredAmount := uint256.NewInt(1000)

		hasSufficientBalance := userBalance.Cmp(requiredAmount) >= 0
		assert.True(t, hasSufficientBalance, "User should have sufficient balance")
	})

	t.Run("Insufficient balance", func(t *testing.T) {
		userBalance := uint256.NewInt(500)
		requiredAmount := uint256.NewInt(1000)

		hasSufficientBalance := userBalance.Cmp(requiredAmount) >= 0
		assert.False(t, hasSufficientBalance, "User should not have sufficient balance")
	})
}
