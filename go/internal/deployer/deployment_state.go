package deployer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DeploymentState holds the addresses of deployed contracts across all networks
type DeploymentState struct {
	Networks map[string]NetworkState `json:"networks"`
}

// NetworkState holds the contract addresses for a specific network
type NetworkState struct {
	ChainID           uint64 `json:"chainId"`
	HyperlaneAddress  string `json:"hyperlaneAddress"`
	OrcaCoinAddress   string `json:"orcaCoinAddress"`
	DogCoinAddress    string `json:"dogCoinAddress"`
	LastUpdated       string `json:"lastUpdated"`
}

// Default deployment state with known Hyperlane addresses
var defaultDeploymentState = DeploymentState{
	Networks: map[string]NetworkState{
		"Sepolia": {
			ChainID:          31337,
			HyperlaneAddress: "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastUpdated:      "",
		},
		"Optimism Sepolia": {
			ChainID:          31338,
			HyperlaneAddress: "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastUpdated:      "",
		},
		"Arbitrum Sepolia": {
			ChainID:          31339,
			HyperlaneAddress: "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastUpdated:      "",
		},
		"Base Sepolia": {
			ChainID:          31340,
			HyperlaneAddress: "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3",
			OrcaCoinAddress:  "",
			DogCoinAddress:   "",
			LastUpdated:      "",
		},
	},
}

// GetDeploymentState loads the current deployment state from file
func GetDeploymentState() (*DeploymentState, error) {
	stateFile := getStateFilePath()
	
	// If file doesn't exist, create it with defaults
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		if err := SaveDeploymentState(&defaultDeploymentState); err != nil {
			return nil, fmt.Errorf("failed to create default state file: %w", err)
		}
		return &defaultDeploymentState, nil
	}
	
	// Read existing state
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	
	var state DeploymentState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	
	return &state, nil
}

// SaveDeploymentState saves the deployment state to file
func SaveDeploymentState(state *DeploymentState) error {
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
	
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	
	return nil
}

// UpdateNetworkState updates the state for a specific network
func UpdateNetworkState(networkName string, orcaCoinAddr, dogCoinAddr string) error {
	state, err := GetDeploymentState()
	if err != nil {
		return err
	}
	
	if network, exists := state.Networks[networkName]; exists {
		network.OrcaCoinAddress = orcaCoinAddr
		network.DogCoinAddress = dogCoinAddr
		network.LastUpdated = "now" // TODO: Add proper timestamp
		state.Networks[networkName] = network
	}
	
	return SaveDeploymentState(state)
}

// getStateFilePath returns the path to the deployment state file
func getStateFilePath() string {
	// Store in the go directory for easy access
	return "deployment-state.json"
}
