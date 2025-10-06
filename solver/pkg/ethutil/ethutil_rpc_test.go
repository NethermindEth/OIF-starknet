package ethutil

import (
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaintainableWithNetworks tests maintainable functions that require RPC connections
// These tests will only run when networks are available and SKIP_RPC_TESTS is not set
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
	t.Logf("Running maintainable code tests with networks (IS_DEVNET=%t)", useLocalForks)

	// Test EVM RPC functions
	t.Run("EVM_RPC_Functions", func(t *testing.T) {
		testEVMRPCFunctions(t, useLocalForks)
	})
}

// testEVMRPCFunctions tests EVM RPC-dependent functions
func testEVMRPCFunctions(t *testing.T, useLocalForks bool) {
	// Test networks (prioritize local forks if available)
	networks := []string{"Base", "Ethereum", "Optimism", "Arbitrum"}
	if useLocalForks {
		networks = []string{"Base", "Ethereum"} // Local forks are typically faster
	}

	for _, networkName := range networks {
		t.Run(networkName, func(t *testing.T) {
			// Get network configuration
			networkConfig, err := config.GetNetworkConfig(networkName)
			if err != nil {
				t.Skipf("Skipping %s: %v", networkName, err)
				return
			}

			// Test client creation and basic RPC functions
			client, err := ethclient.Dial(networkConfig.RPCURL)
			if err != nil {
				t.Skipf("Skipping %s: failed to connect to RPC: %v", networkName, err)
				return
			}
			defer client.Close()

			// Test GetChainID
			t.Run("GetChainID", func(t *testing.T) {
				chainID, err := GetChainID(networkConfig.RPCURL)
				require.NoError(t, err)
				assert.Equal(t, big.NewInt(int64(networkConfig.ChainID)), chainID)
			})

			// Test GetBlockNumber
			t.Run("GetBlockNumber", func(t *testing.T) {
				blockNumber, err := GetBlockNumber(networkConfig.RPCURL)
				require.NoError(t, err)
				assert.True(t, blockNumber > 0, "Block number should be positive")
			})

			// Test SuggestGas
			t.Run("SuggestGas", func(t *testing.T) {
				gasPrice, err := SuggestGas(client)
				require.NoError(t, err)
				assert.True(t, gasPrice.Cmp(big.NewInt(0)) > 0, "Gas price should be positive")
			})

			// Test ERC20 functions if we have token addresses
			testERC20Functions(t, client, networkName, useLocalForks)
		})
	}
}

// testERC20Functions tests EVM ERC20 functions
func testERC20Functions(t *testing.T, client *ethclient.Client, networkName string, useLocalForks bool) {
	// Get Alice's address
	aliceAddress, err := getAliceAddressForNetwork(networkName, useLocalForks)
	if err != nil {
		t.Skipf("Skipping ERC20 tests for %s: %v", networkName, err)
		return
	}

	// Get DogCoin token address
	tokenAddress, err := getDogCoinAddressForNetwork(networkName)
	if err != nil {
		t.Skipf("Skipping ERC20 tests for %s: %v", networkName, err)
		return
	}

	t.Run("ERC20Balance", func(t *testing.T) {
		balance, err := ERC20Balance(client, common.HexToAddress(tokenAddress), common.HexToAddress(aliceAddress))
		require.NoError(t, err)
		assert.True(t, balance.Cmp(big.NewInt(0)) >= 0, "Balance should be non-negative")
	})

	t.Run("ERC20Allowance", func(t *testing.T) {
		// Get Hyperlane address for allowance check
		networkConfig, err := config.GetNetworkConfig(networkName)
		require.NoError(t, err)

		allowance, err := ERC20Allowance(client, common.HexToAddress(tokenAddress), common.HexToAddress(aliceAddress), networkConfig.HyperlaneAddress)
		require.NoError(t, err)
		assert.True(t, allowance.Cmp(big.NewInt(0)) >= 0, "Allowance should be non-negative")
	})
}

// Helper functions
func getAliceAddressForNetwork(_ string, useLocalForks bool) (string, error) {
	if useLocalForks {
		address := os.Getenv("LOCAL_ALICE_PUB_KEY")
		if address == "" {
			return "", fmt.Errorf("LOCAL_ALICE_PUB_KEY not set")
		}
		return address, nil
	}
	address := os.Getenv("ALICE_PUB_KEY")
	if address == "" {
		return "", fmt.Errorf("ALICE_PUB_KEY not set")
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
