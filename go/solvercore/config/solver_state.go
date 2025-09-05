// Package config manages solver state persistence across networks.
//
// SolverState tracks only the last processed blocks for solver listeners
// across all networks (Ethereum, Optimism, Arbitrum, Base, Starknet).
// All contract addresses and network config come from .env files.
//
// Key Features:
// - Minimal persistent storage of last indexed blocks only
// - Thread-safe file operations with atomic writes
// - Automatic fallback to .env start blocks if file doesn't exist
// - Special handling: start block 0 → use current block
//
// Usage:
//
//	state, err := config.GetSolverState()
//	if err := config.UpdateLastIndexedBlock("Ethereum", 12345); err != nil { ... }
//
// This package is actively used by:
// - Solvers (for block tracking between restarts)
// - Listeners (for resuming from last processed block)
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SolverState holds only the solver persistence data across all networks
type SolverState struct {
	Networks map[string]SolverNetworkState `json:"networks"`
}

// SolverNetworkState holds only the last indexed block for solver listeners
// All addresses and config come from .env files via config package
type SolverNetworkState struct {
	LastIndexedBlock uint64 `json:"lastIndexedBlock"`
	LastUpdated      string `json:"lastUpdated"`
}

// getDefaultSolverState creates default solver state with start blocks from .env
func getDefaultSolverState() SolverState {
	// Ensure config is loaded before accessing Networks
	InitializeNetworks()

	return SolverState{
		Networks: map[string]SolverNetworkState{
			"Ethereum": {
				LastIndexedBlock: resolveSolverStartBlock(Networks["Ethereum"].SolverStartBlock),
				LastUpdated:      "",
			},
			"Optimism": {
				LastIndexedBlock: resolveSolverStartBlock(Networks["Optimism"].SolverStartBlock),
				LastUpdated:      "",
			},
			"Arbitrum": {
				LastIndexedBlock: resolveSolverStartBlock(Networks["Arbitrum"].SolverStartBlock),
				LastUpdated:      "",
			},
			"Base": {
				LastIndexedBlock: resolveSolverStartBlock(Networks["Base"].SolverStartBlock),
				LastUpdated:      "",
			},
			"Starknet": {
				LastIndexedBlock: resolveSolverStartBlock(Networks["Starknet"].SolverStartBlock),
				LastUpdated:      "",
			},
		},
	}
}

// resolveSolverStartBlock resolves a solver start block to a valid uint64
// - Positive numbers: use as-is
// - Zero: use 0 (will be resolved to current block by listeners)
// - Negative numbers: use 0 (will be resolved to current block - N by listeners)
func resolveSolverStartBlock(solverStartBlock int64) uint64 {
	if solverStartBlock >= 0 {
		return uint64(solverStartBlock)
	}
	// For negative values, return 0 - the actual resolution will happen in the listener
	// when it has access to the current block number
	return 0
}

// process-local lock to serialize state file access
var solverStateMu sync.Mutex

// GetSolverState loads the current solver state from file
func GetSolverState() (*SolverState, error) {
	solverStateMu.Lock()
	defer solverStateMu.Unlock()
	return readSolverStateLocked()
}

// SaveSolverState saves the solver state to file
func SaveSolverState(state *SolverState) error {
	solverStateMu.Lock()
	defer solverStateMu.Unlock()
	return saveSolverStateLocked(state)
}

// UpdateLastIndexedBlock updates the LastIndexedBlock for a specific network and saves to file
func UpdateLastIndexedBlock(networkName string, newBlockNumber uint64) error {
	solverStateMu.Lock()
	defer solverStateMu.Unlock()

	state, err := readSolverStateLocked()
	if err != nil {
		return fmt.Errorf("failed to get solver state: %w", err)
	}

	network, exists := state.Networks[networkName]
	if !exists {
		return fmt.Errorf("network %s not found in solver state", networkName)
	}

	network.LastIndexedBlock = newBlockNumber
	network.LastUpdated = time.Now().Format(time.RFC3339)
	state.Networks[networkName] = network

	if err := saveSolverStateLocked(state); err != nil {
		return fmt.Errorf("failed to save solver state: %w", err)
	}

	return nil
}

// DisplaySolverState prints the current solver persistence state to stdout
func DisplaySolverState() error {
	state, err := GetSolverState()
	if err != nil {
		return fmt.Errorf("failed to get solver state: %w", err)
	}

	fmt.Printf("📊 Solver State (Last Indexed Blocks):\n")
	fmt.Printf("======================================\n")
	for networkName, networkState := range state.Networks {
		fmt.Printf("🌐 %s: block %d", networkName, networkState.LastIndexedBlock)
		if networkState.LastUpdated != "" {
			fmt.Printf(" (updated: %s)", networkState.LastUpdated)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n💡 All addresses and config come from .env file\n")
	return nil
}

// readSolverStateLocked reads state with retry while holding solverStateMu
func readSolverStateLocked() (*SolverState, error) {
	stateFile := getSolverStateFilePath()
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		defaultState := getDefaultSolverState()
		if err := saveSolverStateLocked(&defaultState); err != nil {
			return nil, fmt.Errorf("failed to create default solver state file: %w", err)
		}
		return &defaultState, nil
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		data, err := os.ReadFile(stateFile)
		if err != nil {
			lastErr = fmt.Errorf("failed to read solver state file: %w", err)
			time.Sleep(25 * time.Millisecond)
			continue
		}

		// Handle empty/corrupted file - fall back to default state
		if len(data) == 0 {
			fmt.Printf("⚠️  Solver state file is empty, creating default state\n")
			defaultState := getDefaultSolverState()
			if err := saveSolverStateLocked(&defaultState); err != nil {
				return nil, fmt.Errorf("failed to create default solver state for empty file: %w", err)
			}
			return &defaultState, nil
		}

		var state SolverState
		if err := json.Unmarshal(data, &state); err != nil {
			// If JSON parsing fails, treat as corrupted and fall back to default
			if i == 2 { // Last retry attempt
				fmt.Printf("⚠️  Solver state file corrupted, creating fresh default state\n")
				defaultState := getDefaultSolverState()
				if err := saveSolverStateLocked(&defaultState); err != nil {
					return nil, fmt.Errorf("failed to create default solver state for corrupted file: %w", err)
				}
				return &defaultState, nil
			}
			lastErr = fmt.Errorf("failed to parse solver state file: %w", err)
			time.Sleep(25 * time.Millisecond)
			continue
		}
		return &state, nil
	}
	return nil, lastErr
}

// saveSolverStateLocked writes the state atomically while holding solverStateMu
func saveSolverStateLocked(state *SolverState) error {
	stateFile := getSolverStateFilePath()
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create solver state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal solver state: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "solver-state-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp solver state file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { tmp.Close(); os.Remove(tmpPath) }()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("failed to write temp solver state file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp solver state file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp solver state file: %w", err)
	}
	if err := os.Rename(tmpPath, stateFile); err != nil {
		return fmt.Errorf("failed to atomically replace solver state file: %w", err)
	}
	return nil
}

// getSolverStateFilePath returns the path to the solver state file
func getSolverStateFilePath() string {
	if custom := os.Getenv("SOLVER_STATE_FILE"); custom != "" {
		return custom
	}
	candidates := []string{"state/solver_state/solver-state.json", "solver-state.json"}
	for _, p := range candidates {
		dir := filepath.Dir(p)
		if _, err := os.Stat(dir); err == nil {
			return p
		}
	}
	return "state/solver_state/solver-state.json"
}
