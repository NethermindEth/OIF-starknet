package starknetutil

import (
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaintainableWithNetworks tests maintainable functions that require RPC connections
func TestMaintainableWithNetworks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RPC coverage test in short mode")
	}

	// Check if we should run RPC tests
	if os.Getenv("SKIP_RPC_TESTS") == "true" {
		t.Skip("RPC tests disabled via SKIP_RPC_TESTS")
	}

	// Check if we should run integration tests (same flag as integration tests)
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("RPC tests disabled via SKIP_INTEGRATION_TESTS")
	}

	useLocalForks := os.Getenv("IS_DEVNET") == "true"
	t.Logf("Running Starknet maintainable code tests with networks (IS_DEVNET=%t)", useLocalForks)

	// Test Starknet RPC functions
	t.Run("Starknet_RPC_Functions", func(t *testing.T) {
		testStarknetRPCFunctions(t, useLocalForks)
	})
}

// testStarknetRPCFunctions tests Starknet RPC-dependent functions
func testStarknetRPCFunctions(t *testing.T, useLocalForks bool) {
	// Get Starknet network configuration
	networkConfig, err := config.GetNetworkConfig("Starknet")
	if err != nil {
		t.Skipf("Skipping Starknet: %v", err)
		return
	}

	// Test provider creation
	provider, err := rpc.NewProvider(networkConfig.RPCURL)
	if err != nil {
		t.Skipf("Skipping Starknet: failed to create provider: %v", err)
		return
	}

	// Test Starknet ERC20 functions if we have token addresses
	testStarknetERC20Functions(t, provider, useLocalForks)
}

// testStarknetERC20Functions tests Starknet ERC20 functions
func testStarknetERC20Functions(t *testing.T, provider *rpc.Provider, useLocalForks bool) {
	// Get Alice's address
	aliceAddress, err := getAliceAddressForNetwork("Starknet", useLocalForks)
	if err != nil {
		t.Skipf("Skipping Starknet ERC20 tests: %v", err)
		return
	}

	// Get DogCoin token address
	tokenAddress, err := getDogCoinAddressForNetwork("Starknet")
	if err != nil {
		t.Skipf("Skipping Starknet ERC20 tests: %v", err)
		return
	}

	t.Run("ERC20Balance", func(t *testing.T) {
		balance, err := ERC20Balance(provider, tokenAddress, aliceAddress)
		require.NoError(t, err)
		assert.True(t, balance.Cmp(big.NewInt(0)) >= 0, "Balance should be non-negative")
	})

	t.Run("ERC20Allowance", func(t *testing.T) {
		// Get Hyperlane address for allowance check
		hyperlaneAddress := os.Getenv("STARKNET_HYPERLANE_ADDRESS")
		if hyperlaneAddress == "" {
			t.Skip("Skipping allowance test: STARKNET_HYPERLANE_ADDRESS not set")
			return
		}

		allowance, err := ERC20Allowance(provider, tokenAddress, aliceAddress, hyperlaneAddress)
		require.NoError(t, err)
		assert.True(t, allowance.Cmp(big.NewInt(0)) >= 0, "Allowance should be non-negative")
	})
}

// Helper functions
func getAliceAddressForNetwork(_ string, useLocalForks bool) (string, error) {
	if useLocalForks {
		address := os.Getenv("LOCAL_STARKNET_ALICE_ADDRESS")
		if address == "" {
			return "", fmt.Errorf("LOCAL_STARKNET_ALICE_ADDRESS not set")
		}
		return address, nil
	}
	address := os.Getenv("STARKNET_ALICE_ADDRESS")
	if address == "" {
		return "", fmt.Errorf("STARKNET_ALICE_ADDRESS not set")
	}
	return address, nil
}

func getDogCoinAddressForNetwork(networkName string) (string, error) {
	envVarName := fmt.Sprintf("%s_DOG_COIN_ADDRESS", strings.ToUpper(networkName))
	address := os.Getenv(envVarName)
	if address == "" {
		return "", fmt.Errorf("no DogCoin address found for %s (env var: %s)", networkName, envVarName)
	}
	return address, nil
}
