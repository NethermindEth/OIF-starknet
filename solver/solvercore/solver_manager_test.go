package solvercore

import (
	"os"
	"testing"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
	"github.com/stretchr/testify/assert"
)

func TestNewSolverManager(t *testing.T) {
	cfg := &config.Config{
		Solvers: map[string]config.SolverConfig{
			"hyperlane7683": {Enabled: true},
		},
	}

	sm := NewSolverManager(cfg)

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.evmClients)
	assert.NotNil(t, sm.activeShutdowns)
	assert.NotNil(t, sm.solverRegistry)
	assert.NotNil(t, sm.allowBlockLists)
	assert.Equal(t, 0, len(sm.evmClients))
	assert.Equal(t, 0, len(sm.activeShutdowns))
	assert.Equal(t, 1, len(sm.solverRegistry))
	assert.True(t, sm.solverRegistry["hyperlane7683"].Enabled)
}

func TestSetAllowBlockLists(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	allowBlockLists := types.AllowBlockLists{
		AllowList: []types.AllowBlockListItem{
			{SenderAddress: "0x123", DestinationDomain: "Ethereum", RecipientAddress: "*"},
		},
		BlockList: []types.AllowBlockListItem{
			{SenderAddress: "0x456", DestinationDomain: "Optimism", RecipientAddress: "*"},
		},
	}

	sm.SetAllowBlockLists(allowBlockLists)

	result := sm.GetAllowBlockLists()
	assert.Equal(t, allowBlockLists, result)
}

func TestGetSolverStatus(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	status := sm.GetSolverStatus()
	assert.NotNil(t, status)
	assert.True(t, status["hyperlane7683"])
}

func TestAddSolver(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	newSolver := SolverConfig{
		Enabled: true,
		Options: map[string]interface{}{
			"test": "value",
		},
	}

	sm.AddSolver("test_solver", newSolver)

	status := sm.GetSolverStatus()
	assert.True(t, status["test_solver"])
}

func TestEnableSolver(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Add a disabled solver
	sm.AddSolver("test_solver", SolverConfig{Enabled: false})

	// Enable it
	err := sm.EnableSolver("test_solver")
	assert.NoError(t, err)

	status := sm.GetSolverStatus()
	assert.True(t, status["test_solver"])
}

func TestDisableSolver(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Disable the default solver
	err := sm.DisableSolver("hyperlane7683")
	assert.NoError(t, err)

	status := sm.GetSolverStatus()
	assert.False(t, status["hyperlane7683"])
}

func TestEnableSolverNotFound(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	err := sm.EnableSolver("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solver nonexistent not found")
}

func TestDisableSolverNotFound(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	err := sm.DisableSolver("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solver nonexistent not found")
}

func TestGetEVMClientNotFound(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	client, err := sm.GetEVMClient(999)
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "EVM client not found for chain ID 999")
}

func TestGetStarknetClientNotInitialized(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	client, err := sm.GetStarknetClient()
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "starknet client not initialized")
}

func TestGetEVMSigner(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Set up test environment
	t.Setenv("DEVNET", "false")
	t.Setenv("SOLVER_PRIVATE_KEY", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	defer func() {
		os.Unsetenv("DEVNET")
		os.Unsetenv("SOLVER_PRIVATE_KEY")
	}()

	signer, err := sm.GetEVMSigner(1)
	assert.NoError(t, err)
	assert.NotNil(t, signer)
	assert.NotNil(t, signer.From)
}

func TestGetEVMSignerMissingKey(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Don't set the private key
	t.Setenv("DEVNET", "false")
	defer os.Unsetenv("DEVNET")

	signer, err := sm.GetEVMSigner(1)
	assert.Nil(t, signer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SOLVER_PRIVATE_KEY environment variable not set")
}

func TestGetEVMSignerInvalidKey(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Set invalid private key
	t.Setenv("DEVNET", "false")
	t.Setenv("SOLVER_PRIVATE_KEY", "invalid_key")
	defer func() {
		os.Unsetenv("DEVNET")
		os.Unsetenv("SOLVER_PRIVATE_KEY")
	}()

	signer, err := sm.GetEVMSigner(1)
	assert.Nil(t, signer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse solver private key")
}

func TestGetStarknetSignerNotInitialized(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	signer, err := sm.GetStarknetSigner()
	assert.Nil(t, signer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "starknet client not initialized")
}

func TestGetStarknetSignerMissingKeys(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Don't set the keys
	t.Setenv("DEVNET", "false")
	defer os.Unsetenv("DEVNET")

	signer, err := sm.GetStarknetSigner()
	assert.Nil(t, signer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "starknet client not initialized")
}

func TestGetStarknetSignerInvalidAddress(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Set invalid address
	t.Setenv("DEVNET", "false")
	t.Setenv("STARKNET_SOLVER_PUBLIC_KEY", "0x123")
	t.Setenv("STARKNET_SOLVER_ADDRESS", "invalid_address")
	t.Setenv("STARKNET_SOLVER_PRIVATE_KEY", "0x123")
	defer func() {
		os.Unsetenv("DEVNET")
		os.Unsetenv("STARKNET_SOLVER_PUBLIC_KEY")
		os.Unsetenv("STARKNET_SOLVER_ADDRESS")
		os.Unsetenv("STARKNET_SOLVER_PRIVATE_KEY")
	}()

	signer, err := sm.GetStarknetSigner()
	assert.Nil(t, signer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "starknet client not initialized")
}

func TestGetStarknetSignerInvalidPrivateKey(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Set invalid private key
	t.Setenv("DEVNET", "false")
	t.Setenv("STARKNET_SOLVER_PUBLIC_KEY", "0x123")
	t.Setenv("STARKNET_SOLVER_ADDRESS", "0x123")
	t.Setenv("STARKNET_SOLVER_PRIVATE_KEY", "invalid_private_key")
	defer func() {
		os.Unsetenv("DEVNET")
		os.Unsetenv("STARKNET_SOLVER_PUBLIC_KEY")
		os.Unsetenv("STARKNET_SOLVER_ADDRESS")
		os.Unsetenv("STARKNET_SOLVER_PRIVATE_KEY")
	}()

	signer, err := sm.GetStarknetSigner()
	assert.Nil(t, signer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "starknet client not initialized")
}

func TestGetStarknetHyperlaneAddress(t *testing.T) {
	// Test with environment variable set
	t.Setenv("DEVNET", "false")
	t.Setenv("STARKNET_HYPERLANE_ADDRESS", "0x1234567890abcdef")
	defer func() {
		os.Unsetenv("DEVNET")
		os.Unsetenv("STARKNET_HYPERLANE_ADDRESS")
	}()

	networkConfig := config.NetworkConfig{}
	addr, err := getStarknetHyperlaneAddress(networkConfig)
	assert.NoError(t, err)
	assert.Equal(t, "0x1234567890abcdef", addr)
}

func TestGetStarknetHyperlaneAddressMissing(t *testing.T) {
	// Test with no environment variable set
	os.Unsetenv("DEVNET")
	os.Unsetenv("STARKNET_HYPERLANE_ADDRESS")

	networkConfig := config.NetworkConfig{}
	addr, err := getStarknetHyperlaneAddress(networkConfig)
	assert.Error(t, err)
	assert.Equal(t, "", addr)
	assert.Contains(t, err.Error(), "no STARKNET_HYPERLANE_ADDRESS set in .env")
}

func TestShutdown(t *testing.T) {
	sm := NewSolverManager(&config.Config{})

	// Add some mock shutdown functions
	shutdownCount := 0
	sm.activeShutdowns = append(sm.activeShutdowns, func() { shutdownCount++ })
	sm.activeShutdowns = append(sm.activeShutdowns, func() { shutdownCount++ })

	sm.Shutdown()

	assert.Equal(t, 2, shutdownCount)
	assert.Equal(t, 0, len(sm.activeShutdowns))
}
