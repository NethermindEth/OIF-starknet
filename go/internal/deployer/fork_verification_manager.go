package deployer

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ForkVerificationConfig represents the configuration for verifying a forked network
type ForkVerificationConfig struct {
	RPCURL     string
	PrivateKey string
	ChainName  string
	ChainID    int64
}

// ForkVerificationManager manages the verification of pre-deployed contracts on forked networks
type ForkVerificationManager struct {
	configs []ForkVerificationConfig
}

// NewForkVerificationManager creates a new fork verification manager
func NewForkVerificationManager(configs []ForkVerificationConfig) *ForkVerificationManager {
	return &ForkVerificationManager{
		configs: configs,
	}
}

// VerifyPreDeployedContracts verifies that all pre-deployed contracts exist on the forked networks
func (fvm *ForkVerificationManager) VerifyPreDeployedContracts() error {
	for _, config := range fvm.configs {
		if err := fvm.verifyNetwork(config); err != nil {
			return fmt.Errorf("failed to verify network %s: %w", config.ChainName, err)
		}
	}
	return nil
}

// verifyNetwork verifies a single network
func (fvm *ForkVerificationManager) verifyNetwork(config ForkVerificationConfig) error {
	log.Printf("🔍 Verifying network: %s (Chain ID: %d)", config.ChainName, config.ChainID)

	// Connect to the network
	client, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", config.RPCURL, err)
	}
	defer client.Close()

	// Check if the network is running
	ctx := context.Background()
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block number from %s: %w", config.RPCURL, err)
	}

	log.Printf("   ✅ Network is running (Block: %d)", blockNumber)

	// Verify chain ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID from %s: %w", config.RPCURL, err)
	}

	if chainID.Cmp(big.NewInt(config.ChainID)) != 0 {
		return fmt.Errorf("chain ID mismatch: expected %d, got %s", config.ChainID, chainID.String())
	}

	log.Printf("   ✅ Chain ID verified: %s", chainID.String())

	return nil
}

// GetContractAddresses returns the contract addresses for each network
func (fvm *ForkVerificationManager) GetContractAddresses() map[string]common.Address {
	addresses := make(map[string]common.Address)
	
	// Load deployment state to get actual contract addresses
	state, err := GetDeploymentState()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to load deployment state: %v", err)
		// Fall back to zero addresses if we can't load state
		for _, config := range fvm.configs {
			addresses[config.ChainName] = common.Address{}
		}
		return addresses
	}
	
	// Map chain names to deployment state network names
	chainNameToNetwork := map[string]string{
		"ethereum": "Sepolia",
		"optimism": "Optimism Sepolia", 
		"arbitrum": "Arbitrum Sepolia",
		"base":     "Base Sepolia",
	}
	
	for _, config := range fvm.configs {
		networkName, exists := chainNameToNetwork[config.ChainName]
		if !exists {
			log.Printf("⚠️  Warning: Unknown chain name: %s", config.ChainName)
			addresses[config.ChainName] = common.Address{}
			continue
		}
		
		networkState, exists := state.Networks[networkName]
		if !exists {
			log.Printf("⚠️  Warning: No deployment state found for network: %s", networkName)
			addresses[config.ChainName] = common.Address{}
			continue
		}
		
		// Use the Hyperlane contract address
		addresses[config.ChainName] = common.HexToAddress(networkState.HyperlaneAddress)
	}
	
	return addresses
}
