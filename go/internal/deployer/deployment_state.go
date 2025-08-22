package deployer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
)

// DeploymentState holds the addresses of deployed contracts across all networks
type DeploymentState struct {
	Networks map[string]NetworkState `json:"networks"`
}

// NetworkState holds the contract addresses for a specific network
type NetworkState struct {
	ChainID          uint64 `json:"chainId"`
	HyperlaneAddress string `json:"hyperlaneAddress"`
	OrcaCoinAddress  string `json:"orcaCoinAddress"`
	DogCoinAddress   string `json:"dogCoinAddress"`
	LastIndexedBlock uint64 `json:"lastIndexedBlock"`
	LastUpdated      string `json:"lastUpdated"`
}

// Default deployment state with known Hyperlane addresses
var defaultDeploymentState = DeploymentState{
	Networks: map[string]NetworkState{
		"Sepolia": {
			ChainID:          config.Networks["Sepolia"].ChainID,
			HyperlaneAddress: config.Networks["Sepolia"].HyperlaneAddress.Hex(),
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastIndexedBlock: config.Networks["Sepolia"].SolverStartBlock,
			LastUpdated:      "",
		},
		"Optimism Sepolia": {
			ChainID:          config.Networks["Optimism Sepolia"].ChainID,
			HyperlaneAddress: config.Networks["Optimism Sepolia"].HyperlaneAddress.Hex(),
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastIndexedBlock: config.Networks["Optimism Sepolia"].SolverStartBlock,
			LastUpdated:      "",
		},
		"Arbitrum Sepolia": {
			ChainID:          config.Networks["Arbitrum Sepolia"].ChainID,
			HyperlaneAddress: config.Networks["Arbitrum Sepolia"].HyperlaneAddress.Hex(),
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastIndexedBlock: config.Networks["Arbitrum Sepolia"].SolverStartBlock,
			LastUpdated:      "",
		},
		"Base Sepolia": {
			ChainID:          config.Networks["Base Sepolia"].ChainID,
			HyperlaneAddress: config.Networks["Base Sepolia"].HyperlaneAddress.Hex(),
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastIndexedBlock: config.Networks["Base Sepolia"].SolverStartBlock,
			LastUpdated:      "",
		},
		"Starknet Sepolia": {
			ChainID:          config.Networks["Starknet Sepolia"].ChainID,
			HyperlaneAddress: "", // Will be populated after deployment
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastIndexedBlock: config.Networks["Starknet Sepolia"].SolverStartBlock,
			LastUpdated:      "",
		},
	},
}

// process-local lock to serialize state file access
var stateMu sync.Mutex

// GetDeploymentState loads the current deployment state from file
func GetDeploymentState() (*DeploymentState, error) {
	stateMu.Lock()
	defer stateMu.Unlock()
	return readStateLocked()
}

// SaveDeploymentState saves the deployment state to file
func SaveDeploymentState(state *DeploymentState) error {
	stateMu.Lock()
	defer stateMu.Unlock()
	return saveStateLocked(state)
}

// UpdateNetworkState updates the state for a specific network
func UpdateNetworkState(networkName string, orcaCoinAddr, dogCoinAddr string) error {
	stateMu.Lock()
	defer stateMu.Unlock()

	state, err := readStateLocked()
	if err != nil {
		return err
	}
	if network, exists := state.Networks[networkName]; exists {
		network.OrcaCoinAddress = orcaCoinAddr
		network.DogCoinAddress = dogCoinAddr
		network.LastUpdated = time.Now().Format(time.RFC3339)
		state.Networks[networkName] = network
	}
	return saveStateLocked(state)
}

// UpdateLastIndexedBlock updates the LastIndexedBlock for a specific network and saves to file
func UpdateLastIndexedBlock(networkName string, newBlockNumber uint64) error {
	stateMu.Lock()
	defer stateMu.Unlock()

	state, err := readStateLocked()
	if err != nil {
		return fmt.Errorf("failed to get deployment state: %w", err)
	}

	network, exists := state.Networks[networkName]
	if !exists {
		return fmt.Errorf("network %s not found in deployment state", networkName)
	}

	oldBlock := network.LastIndexedBlock
	network.LastIndexedBlock = newBlockNumber
	network.LastUpdated = time.Now().Format(time.RFC3339)
	state.Networks[networkName] = network

	if err := saveStateLocked(state); err != nil {
		return fmt.Errorf("failed to save deployment state: %w", err)
	}

	if oldBlock != newBlockNumber {
		fmt.Printf("✅ Updated %s LastIndexedBlock: %d → %d\n", networkName, oldBlock, newBlockNumber)
	}

	return nil
}

// UpdateHyperlaneAddress updates the HyperlaneAddress for a specific network and saves to file
func UpdateHyperlaneAddress(networkName string, newAddress string) error {
	stateMu.Lock()
	defer stateMu.Unlock()

	state, err := readStateLocked()
	if err != nil {
		return fmt.Errorf("failed to get deployment state: %w", err)
	}

	network, exists := state.Networks[networkName]
	if !exists {
		return fmt.Errorf("network %s not found in deployment state", networkName)
	}

	network.HyperlaneAddress = newAddress
	network.LastUpdated = time.Now().Format(time.RFC3339)
	state.Networks[networkName] = network

	if err := saveStateLocked(state); err != nil {
		return fmt.Errorf("failed to save deployment state: %w", err)
	}

	fmt.Printf("✅ Updated %s HyperlaneAddress: %s\n", networkName, newAddress)
	return nil
}

// DisplayDeploymentState prints the current deployment state to stdout
func DisplayDeploymentState() error {
	state, err := GetDeploymentState()
	if err != nil {
		return fmt.Errorf("failed to get deployment state: %w", err)
	}

	fmt.Printf("📊 Current Deployment State:\n")
	fmt.Printf("============================\n")
	for networkName, networkState := range state.Networks {
		fmt.Printf("🌐 %s:\n", networkName)
		fmt.Printf("   • Chain ID: %d\n", networkState.ChainID)
		fmt.Printf("   • Hyperlane Address: %s\n", networkState.HyperlaneAddress)
		fmt.Printf("   • Last Indexed Block: %d\n", networkState.LastIndexedBlock)
		fmt.Printf("   • Last Updated: %s\n", networkState.LastUpdated)
		fmt.Printf("   • Orca Coin: %s\n", networkState.OrcaCoinAddress)
		fmt.Printf("   • Dog Coin: %s\n", networkState.DogCoinAddress)
		fmt.Printf("\n")
	}

	return nil
}

// readStateLocked reads state with retry while holding stateMu
func readStateLocked() (*DeploymentState, error) {
	stateFile := getStateFilePath()

	// If file doesn't exist, create it with defaults
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		if err := saveStateLocked(&defaultDeploymentState); err != nil {
			return nil, fmt.Errorf("failed to create default state file: %w", err)
		}
		return &defaultDeploymentState, nil
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		data, err := os.ReadFile(stateFile)
		if err != nil {
			lastErr = fmt.Errorf("failed to read state file: %w", err)
			time.Sleep(25 * time.Millisecond)
			continue
		}
		var state DeploymentState
		if err := json.Unmarshal(data, &state); err != nil {
			lastErr = fmt.Errorf("failed to parse state file: %w", err)
			time.Sleep(25 * time.Millisecond)
			continue
		}
		return &state, nil
	}
	return nil, lastErr
}

// saveStateLocked writes the state atomically while holding stateMu
func saveStateLocked(state *DeploymentState) error {
	stateFile := getStateFilePath()

	// Ensure directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Atomic write: temp file -> fsync -> rename
	tmp, err := os.CreateTemp(dir, "deployment-state-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp state file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp state file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp state file: %w", err)
	}
	if err := os.Rename(tmpPath, stateFile); err != nil {
		return fmt.Errorf("failed to atomically replace state file: %w", err)
	}
	return nil
}

// getStateFilePath returns the path to the deployment state file
func getStateFilePath() string {
	// Allow override via environment variable
	if custom := os.Getenv("STATE_FILE"); custom != "" {
		return custom
	}

	// Prefer local go directory state path
	candidates := []string{
		"state/network_state/deployment-state.json",
		"deployment-state.json",
	}
	for _, p := range candidates {
		dir := filepath.Dir(p)
		if _, err := os.Stat(dir); err == nil {
			return p
		}
	}

	// Fallback to local state path
	return "state/network_state/deployment-state.json"
}
