package config

import (
	"fmt"

	"github.com/NethermindEth/oif-starknet/go/pkg/envutil"
	"github.com/ethereum/go-ethereum/common"
)

// Network configuration constants
const (
	// Chain IDs
	EthereumSepoliaChainID = 11155111
	OptimismSepoliaChainID = 11155420
	ArbitrumSepoliaChainID = 421614
	BaseSepoliaChainID     = 84532
	StarknetSepoliaChainID = 23448591

	// Default block numbers (0 = latest block for live networks)
	EthereumDefaultStartBlock = 0
	OptimismDefaultStartBlock = 0
	ArbitrumDefaultStartBlock = 0
	BaseDefaultStartBlock     = 0
	StarknetDefaultStartBlock = 0

	// Local fork block numbers (latest blocks for local development)
	EthereumLocalStartBlock = 9121214
	OptimismLocalStartBlock = 32529526
	ArbitrumLocalStartBlock = 190326088
	BaseLocalStartBlock     = 30546661
	StarknetLocalStartBlock = 1850850

	// Default intervals
	DefaultPollIntervalMs         = 1000
	StarknetDefaultPollIntervalMs = 2000
	DefaultMaxBlockRange          = 10
	StarknetDefaultMaxBlockRange  = 100
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

// GetConditionalAccountEnv gets account-related environment variables based on FORKING flag
// This is a convenience function for account keys and addresses
// Deprecated: Use envutil.GetConditionalAccountEnv instead
func GetConditionalAccountEnv(key string) string {
	return envutil.GetConditionalAccountEnv(key)
}

// networksInitialized tracks whether networks have been initialized from env vars
var networksInitialized = false

// InitializeNetworks must be called after loading .env file to ensure proper config
func InitializeNetworks() {
	if !networksInitialized {
		initializeNetworks()
	}
}

// ensureInitialized initializes networks if not already done (fallback for legacy usage)
func ensureInitialized() {
	if !networksInitialized {
		initializeNetworks()
	}
}

// Networks contains all network configurations
var Networks map[string]NetworkConfig

// initializeNetworks initializes the network configurations from environment variables
func initializeNetworks() {
	Networks = map[string]NetworkConfig{
		"Ethereum": {
			Name:               "Ethereum",
			RPCURL:             envutil.GetConditionalEnv("ETHEREUM_RPC_URL", "http://localhost:8545"),
			ChainID:            envutil.GetEnvUint64Any([]string{"ETHEREUM_CHAIN_ID", "SEPOLIA_CHAIN_ID"}, EthereumSepoliaChainID),
			HyperlaneAddress:   common.HexToAddress(envutil.GetEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
			HyperlaneDomain:    envutil.GetEnvUint64Any([]string{"ETHEREUM_DOMAIN_ID", "SEPOLIA_DOMAIN_ID"}, EthereumSepoliaChainID),
			ForkStartBlock:     envutil.GetConditionalUint64("ETHEREUM_SOLVER_START_BLOCK", EthereumDefaultStartBlock, EthereumLocalStartBlock),
			SolverStartBlock:   envutil.GetConditionalUint64("ETHEREUM_SOLVER_START_BLOCK", EthereumDefaultStartBlock, EthereumLocalStartBlock),
			PollInterval:       envutil.GetEnvInt("POLL_INTERVAL_MS", DefaultPollIntervalMs),
			ConfirmationBlocks: envutil.GetEnvUint64("CONFIRMATION_BLOCKS", 0),
			MaxBlockRange:      envutil.GetEnvUint64("MAX_BLOCK_RANGE", DefaultMaxBlockRange),
		},
		"Optimism": {
			Name:               "Optimism",
			RPCURL:             envutil.GetConditionalEnv("OPTIMISM_RPC_URL", "http://localhost:8546"),
			ChainID:            envutil.GetEnvUint64("OPTIMISM_CHAIN_ID", OptimismSepoliaChainID),
			HyperlaneAddress:   common.HexToAddress(envutil.GetEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
			HyperlaneDomain:    envutil.GetEnvUint64("OPTIMISM_DOMAIN_ID", OptimismSepoliaChainID),
			ForkStartBlock:     envutil.GetConditionalUint64("OPTIMISM_SOLVER_START_BLOCK", OptimismDefaultStartBlock, OptimismLocalStartBlock),
			SolverStartBlock:   envutil.GetConditionalUint64("OPTIMISM_SOLVER_START_BLOCK", OptimismDefaultStartBlock, OptimismLocalStartBlock),
			PollInterval:       envutil.GetEnvInt("POLL_INTERVAL_MS", DefaultPollIntervalMs),
			ConfirmationBlocks: envutil.GetEnvUint64("CONFIRMATION_BLOCKS", 0),
			MaxBlockRange:      envutil.GetEnvUint64("MAX_BLOCK_RANGE", DefaultMaxBlockRange),
		},
		"Arbitrum": {
			Name:               "Arbitrum",
			RPCURL:             envutil.GetConditionalEnv("ARBITRUM_RPC_URL", "http://localhost:8547"),
			ChainID:            envutil.GetEnvUint64("ARBITRUM_CHAIN_ID", ArbitrumSepoliaChainID),
			HyperlaneAddress:   common.HexToAddress(envutil.GetEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
			HyperlaneDomain:    envutil.GetEnvUint64("ARBITRUM_DOMAIN_ID", ArbitrumSepoliaChainID),
			ForkStartBlock:     envutil.GetConditionalUint64("ARBITRUM_SOLVER_START_BLOCK", ArbitrumDefaultStartBlock, ArbitrumLocalStartBlock),
			SolverStartBlock:   envutil.GetConditionalUint64("ARBITRUM_SOLVER_START_BLOCK", ArbitrumDefaultStartBlock, ArbitrumLocalStartBlock),
			PollInterval:       envutil.GetEnvInt("POLL_INTERVAL_MS", DefaultPollIntervalMs),
			ConfirmationBlocks: envutil.GetEnvUint64("CONFIRMATION_BLOCKS", 0),
			MaxBlockRange:      envutil.GetEnvUint64("MAX_BLOCK_RANGE", DefaultMaxBlockRange),
		},
		"Base": {
			Name:               "Base",
			RPCURL:             envutil.GetConditionalEnv("BASE_RPC_URL", "http://localhost:8548"),
			ChainID:            envutil.GetEnvUint64("BASE_CHAIN_ID", BaseSepoliaChainID),
			HyperlaneAddress:   common.HexToAddress(envutil.GetEnvWithDefault("EVM_HYPERLANE_ADDRESS", "0xf614c6bF94b022E16BEF7dBecF7614FFD2b201d3")),
			HyperlaneDomain:    envutil.GetEnvUint64("BASE_DOMAIN_ID", BaseSepoliaChainID),
			ForkStartBlock:     envutil.GetConditionalUint64("BASE_SOLVER_START_BLOCK", BaseDefaultStartBlock, BaseLocalStartBlock),
			SolverStartBlock:   envutil.GetConditionalUint64("BASE_SOLVER_START_BLOCK", BaseDefaultStartBlock, BaseLocalStartBlock),
			PollInterval:       envutil.GetEnvInt("POLL_INTERVAL_MS", DefaultPollIntervalMs),
			ConfirmationBlocks: envutil.GetEnvUint64("CONFIRMATION_BLOCKS", 0),
			MaxBlockRange:      envutil.GetEnvUint64("MAX_BLOCK_RANGE", DefaultMaxBlockRange),
		},
		"Starknet": {
			Name:               "Starknet",
			RPCURL:             envutil.GetConditionalEnv("STARKNET_RPC_URL", "http://localhost:5050"),
			ChainID:            envutil.GetEnvUint64("STARKNET_CHAIN_ID", StarknetSepoliaChainID),
			HyperlaneAddress:   common.HexToAddress(envutil.GetEnvWithDefault("STARKNET_HYPERLANE_ADDRESS", "")),
			HyperlaneDomain:    envutil.GetEnvUint64("STARKNET_DOMAIN_ID", StarknetSepoliaChainID),
			ForkStartBlock:     envutil.GetConditionalUint64("STARKNET_SOLVER_START_BLOCK", StarknetDefaultStartBlock, StarknetLocalStartBlock),
			SolverStartBlock:   envutil.GetConditionalUint64("STARKNET_SOLVER_START_BLOCK", StarknetDefaultStartBlock, StarknetLocalStartBlock),
			PollInterval:       envutil.GetEnvInt("STARKNET_POLL_INTERVAL_MS", envutil.GetEnvInt("POLL_INTERVAL_MS", StarknetDefaultPollIntervalMs)),
			ConfirmationBlocks: envutil.GetEnvUint64("STARKNET_CONFIRMATION_BLOCKS", 0),
			MaxBlockRange: envutil.GetEnvUint64("STARKNET_MAX_BLOCK_RANGE",
				envutil.GetEnvUint64("MAX_BLOCK_RANGE", StarknetDefaultMaxBlockRange)),
		},
	}
	networksInitialized = true
}

// GetNetworkConfig returns the configuration for a given network name
func GetNetworkConfig(networkName string) (NetworkConfig, error) {
	ensureInitialized()
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
	ensureInitialized()
	names := make([]string, 0, len(Networks))
	for name := range Networks {
		names = append(names, name)
	}
	return names
}

// ValidateNetworkName checks if a network name is valid
func ValidateNetworkName(networkName string) bool {
	ensureInitialized()
	_, exists := Networks[networkName]
	return exists
}

// GetDefaultNetwork returns the default network (Ethereum)
func GetDefaultNetwork() NetworkConfig {
	ensureInitialized()
	return Networks["Ethereum"]
}

// GetDefaultRPCURL returns the default RPC URL
func GetDefaultRPCURL() string {
	ensureInitialized()
	return Networks["Ethereum"].RPCURL
}
