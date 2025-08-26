package config

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

// NetworkConfig represents a single network configuration
type NetworkConfig struct {
	Name             string
	RPCURL           string
	ChainID          uint64
	HyperlaneAddress common.Address
	HyperlaneDomain  uint64 // Changed to uint64 to match new_code
	ForkStartBlock   uint64
	SolverStartBlock uint64 // Block number where solver should start listening (fork block + 1)
	// Listener-specific configuration
	PollInterval       int    // milliseconds, 0 = use default
	ConfirmationBlocks uint64 // 0 = use default
	MaxBlockRange      uint64 // 0 = use default
}

// getEnvWithDefault gets an environment variable with a default fallback
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvUint64 gets an environment variable as uint64 with a default fallback
func getEnvUint64(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := parseUint64(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt gets an environment variable as int with a default fallback
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

// parseUint64 parses a string to uint64
func parseUint64(s string) (uint64, error) {
	var result uint64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// Networks contains all network configurations
var Networks = map[string]NetworkConfig{
	"Sepolia": {
		Name:               "Sepolia",
		RPCURL:             getEnvWithDefault("SEPOLIA_RPC_URL", "http://localhost:8545"),
		ChainID:            getEnvUint64("SEPOLIA_CHAIN_ID", 11155111),
		HyperlaneAddress:   common.HexToAddress(getEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
		HyperlaneDomain:    getEnvUint64("SEPOLIA_DOMAIN_ID", 11155111),
		ForkStartBlock:     getEnvUint64("SEPOLIA_SOLVER_START_BLOCK", 8319000),
		SolverStartBlock:   getEnvUint64("SEPOLIA_SOLVER_START_BLOCK", 8319000),
		PollInterval:       getEnvInt("POLL_INTERVAL_MS", 1000),
		ConfirmationBlocks: getEnvUint64("CONFIRMATION_BLOCKS", 0), // Read from environment
		MaxBlockRange:      getEnvUint64("MAX_BLOCK_RANGE", 500),
	},
	"Optimism Sepolia": {
		Name:               "Optimism Sepolia",
		RPCURL:             getEnvWithDefault("OPTIMISM_RPC_URL", "http://localhost:8546"),
		ChainID:            getEnvUint64("OPTIMISM_CHAIN_ID", 11155420),
		HyperlaneAddress:   common.HexToAddress(getEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
		HyperlaneDomain:    getEnvUint64("OPTIMISM_DOMAIN_ID", 11155420),
		ForkStartBlock:     getEnvUint64("OPTIMISM_SOLVER_START_BLOCK", 27370000),
		SolverStartBlock:   getEnvUint64("OPTIMISM_SOLVER_START_BLOCK", 27370000),
		PollInterval:       getEnvInt("POLL_INTERVAL_MS", 1000),
		ConfirmationBlocks: getEnvUint64("CONFIRMATION_BLOCKS", 0), // Read from environment
		MaxBlockRange:      getEnvUint64("MAX_BLOCK_RANGE", 500),
	},
	"Arbitrum Sepolia": {
		Name:               "Arbitrum Sepolia",
		RPCURL:             getEnvWithDefault("ARBITRUM_RPC_URL", "http://localhost:8547"),
		ChainID:            getEnvUint64("ARBITRUM_CHAIN_ID", 421614),
		HyperlaneAddress:   common.HexToAddress(getEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
		HyperlaneDomain:    getEnvUint64("ARBITRUM_DOMAIN_ID", 421614),
		ForkStartBlock:     getEnvUint64("ARBITRUM_SOLVER_START_BLOCK", 138020000),
		SolverStartBlock:   getEnvUint64("ARBITRUM_SOLVER_START_BLOCK", 138020000),
		PollInterval:       getEnvInt("POLL_INTERVAL_MS", 1000),
		ConfirmationBlocks: getEnvUint64("CONFIRMATION_BLOCKS", 0), // Read from environment
		MaxBlockRange:      getEnvUint64("MAX_BLOCK_RANGE", 500),
	},
	"Base Sepolia": {
		Name:               "Base Sepolia",
		RPCURL:             getEnvWithDefault("BASE_RPC_URL", "http://localhost:8548"),
		ChainID:            getEnvUint64("BASE_CHAIN_ID", 84532),
		HyperlaneAddress:   common.HexToAddress(getEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
		HyperlaneDomain:    getEnvUint64("BASE_DOMAIN_ID", 84532),
		ForkStartBlock:     getEnvUint64("BASE_SOLVER_START_BLOCK", 25380000),
		SolverStartBlock:   getEnvUint64("BASE_SOLVER_START_BLOCK", 25380000),
		PollInterval:       getEnvInt("POLL_INTERVAL_MS", 1000),
		ConfirmationBlocks: getEnvUint64("CONFIRMATION_BLOCKS", 0), // Read from environment
		MaxBlockRange:      getEnvUint64("MAX_BLOCK_RANGE", 500),
	},
	"Starknet Sepolia": {
		Name:               "Starknet Sepolia",
		RPCURL:             getEnvWithDefault("STARKNET_RPC_URL", "http://localhost:5050"),
		ChainID:            getEnvUint64("STARKNET_CHAIN_ID", 23448591),
		HyperlaneAddress:   common.HexToAddress(getEnvWithDefault("STARKNET_HYPERLANE_ADDRESS", "")), // Can be set manually via .env, but deployment state takes precedence
		HyperlaneDomain:    getEnvUint64("STARKNET_DOMAIN_ID", 23448591),
		ForkStartBlock:     getEnvUint64("STARKNET_SOLVER_START_BLOCK", 1530000),
		SolverStartBlock:   getEnvUint64("STARKNET_SOLVER_START_BLOCK", 1530000),
		PollInterval:       getEnvInt("STARKNET_POLL_INTERVAL_MS", getEnvInt("POLL_INTERVAL_MS", 2000)),    // Starknet-specific override or global
		ConfirmationBlocks: getEnvUint64("STARKNET_CONFIRMATION_BLOCKS", 0),                                // Starknet-specific override
		MaxBlockRange:      getEnvUint64("STARKNET_MAX_BLOCK_RANGE", getEnvUint64("MAX_BLOCK_RANGE", 100)), // Starknet-specific override or global
	},
}

// GetNetworkConfig returns the configuration for a given network name
func GetNetworkConfig(networkName string) (NetworkConfig, error) {
	if config, exists := Networks[networkName]; exists {
		return config, nil
	}
	return NetworkConfig{}, fmt.Errorf("network not found: %s", networkName)
}

// GetRPCURL returns the RPC URL for a given network name
func GetRPCURL(networkName string) (string, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return "", err
	}
	return config.RPCURL, nil
}

// GetChainID returns the chain ID for a given network name
func GetChainID(networkName string) (uint64, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, err
	}
	return config.ChainID, nil
}

// GetHyperlaneAddress returns the Hyperlane contract address for a given network name
func GetHyperlaneAddress(networkName string) (common.Address, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return common.Address{}, err
	}
	return config.HyperlaneAddress, nil
}

// GetHyperlaneDomain returns the Hyperlane domain ID for a given network name
func GetHyperlaneDomain(networkName string) (uint64, error) { // Changed to uint64 to match new_code
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, err
	}
	return config.HyperlaneDomain, nil
}

// GetForkStartBlock returns the fork start block for a given network name
func GetForkStartBlock(networkName string) (uint64, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, err
	}
	return config.ForkStartBlock, nil
}

// GetSolverStartBlock returns the solver start block for a given network name
func GetSolverStartBlock(networkName string) (uint64, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, err
	}
	return config.SolverStartBlock, nil
}

// GetListenerConfig returns the listener configuration for a given network name
func GetListenerConfig(networkName string) (int, uint64, uint64, error) {
	config, err := GetNetworkConfig(networkName)
	if err != nil {
		return 0, 0, 0, err
	}
	return config.PollInterval, config.ConfirmationBlocks, config.MaxBlockRange, nil
}

// GetRPCURLByChainID returns the RPC URL for a given chain ID
func GetRPCURLByChainID(chainID uint64) (string, error) {
	for _, network := range Networks {
		if network.ChainID == chainID {
			return network.RPCURL, nil
		}
	}
	return "", fmt.Errorf("network not found for chain ID: %d", chainID)
}

// GetHyperlaneAddressByChainID returns the Hyperlane address for a given chain ID
func GetHyperlaneAddressByChainID(chainID uint64) (common.Address, error) {
	for _, network := range Networks {
		if network.ChainID == chainID {
			return network.HyperlaneAddress, nil
		}
	}
	return common.Address{}, fmt.Errorf("network not found for chain ID: %d", chainID)
}

// GetNetworkNames returns all available network names
func GetNetworkNames() []string {
	names := make([]string, 0, len(Networks))
	for name := range Networks {
		names = append(names, name)
	}
	return names
}

// ValidateNetworkName checks if a network name is valid
func ValidateNetworkName(networkName string) bool {
	_, exists := Networks[networkName]
	return exists
}

// GetDefaultNetwork returns the default network (Sepolia for now)
func GetDefaultNetwork() NetworkConfig {
	return Networks["Sepolia"]
}

// GetDefaultRPCURL returns the default RPC URL
func GetDefaultRPCURL() string {
	return Networks["Sepolia"].RPCURL
}
