package types

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/common"
)

// AddressConverter handles conversion between different address formats
type AddressConverter struct{}

// NewAddressConverter creates a new address converter
func NewAddressConverter() *AddressConverter {
	return &AddressConverter{}
}

// ToEVMAddress converts a string address to EVM common.Address for operations like allowances
func (ac *AddressConverter) ToEVMAddress(address string) (common.Address, error) {
	// Remove 0x prefix if present
	cleanAddr := strings.TrimPrefix(address, "0x")
	
	// If it's a 62-character hex string (Starknet felt - 31 bytes), convert to EVM
	if len(cleanAddr) == 62 {
		// Convert Starknet felt to EVM address by taking last 20 bytes
		bytes, err := hex.DecodeString(cleanAddr)
		if err != nil {
			return common.Address{}, fmt.Errorf("failed to decode Starknet address: %w", err)
		}
		if len(bytes) != 31 {
			return common.Address{}, fmt.Errorf("invalid Starknet address length: %d", len(bytes))
		}
		// Take last 20 bytes for EVM address (right-aligned)
		// If Starknet address is less than 20 bytes, it will be left-padded with zeros
		evmBytes := make([]byte, 20)
		copy(evmBytes[20-len(bytes):], bytes)
		return common.BytesToAddress(evmBytes), nil
	}
	
	// If it's a 64-character hex string (EVM bytes32 - 32 bytes), extract the address
	if len(cleanAddr) == 64 {
		bytes, err := hex.DecodeString(cleanAddr)
		if err != nil {
			return common.Address{}, fmt.Errorf("failed to decode bytes32 address: %w", err)
		}
		if len(bytes) != 32 {
			return common.Address{}, fmt.Errorf("invalid bytes32 address length: %d", len(bytes))
		}
		// Take last 20 bytes for EVM address (right-aligned)
		return common.BytesToAddress(bytes[12:]), nil
	}
	
	// If it's a 40-character hex string (EVM address), use directly
	if len(cleanAddr) == 40 {
		return common.HexToAddress(address), nil
	}
	
	return common.Address{}, fmt.Errorf("unsupported address format: %s", address)
}

// ToStarknetAddress converts a string address to Starknet felt for operations like allowances
func (ac *AddressConverter) ToStarknetAddress(address string) (*felt.Felt, error) {
	// Remove 0x prefix if present
	cleanAddr := strings.TrimPrefix(address, "0x")
	
	// If it's already a 62-character hex string (Starknet felt - 31 bytes), convert directly
	if len(cleanAddr) == 62 {
		return utils.HexToFelt(address)
	}
	
	// If it's a 64-character hex string (EVM bytes32 - 32 bytes), convert to Starknet felt
	if len(cleanAddr) == 64 {
		bytes, err := hex.DecodeString(cleanAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode bytes32 address: %w", err)
		}
		if len(bytes) != 32 {
			return nil, fmt.Errorf("invalid bytes32 address length: %d", len(bytes))
		}
		// Convert bytes32 to Starknet felt (take last 31 bytes)
		starknetBytes := bytes[1:] // Remove first byte, keep last 31
		return utils.HexToFelt("0x" + hex.EncodeToString(starknetBytes))
	}
	
	// If it's a 40-character hex string (EVM address), convert to Starknet felt
	if len(cleanAddr) == 40 {
		evmAddr := common.HexToAddress(address)
		// Convert EVM address to Starknet felt (no padding needed)
		return utils.HexToFelt("0x" + hex.EncodeToString(evmAddr.Bytes()))
	}
	
	return nil, fmt.Errorf("unsupported address format: %s", address)
}

// ToBytes32 converts a string address to bytes32 for order hashing and contract calls
func (ac *AddressConverter) ToBytes32(address string) ([32]byte, error) {
	// Remove 0x prefix if present
	cleanAddr := strings.TrimPrefix(address, "0x")
	
	// If it's already a 64-character hex string (bytes32), decode directly
	if len(cleanAddr) == 64 {
		bytes, err := hex.DecodeString(cleanAddr)
		if err != nil {
			return [32]byte{}, fmt.Errorf("failed to decode bytes32 address: %w", err)
		}
		if len(bytes) != 32 {
			return [32]byte{}, fmt.Errorf("invalid bytes32 address length: %d", len(bytes))
		}
		var result [32]byte
		copy(result[:], bytes)
		return result, nil
	}
	
	// If it's a 62-character hex string (Starknet felt), pad to 32 bytes
	if len(cleanAddr) == 62 {
		bytes, err := hex.DecodeString(cleanAddr)
		if err != nil {
			return [32]byte{}, fmt.Errorf("failed to decode Starknet address: %w", err)
		}
		if len(bytes) != 31 {
			return [32]byte{}, fmt.Errorf("invalid Starknet address length: %d", len(bytes))
		}
		var result [32]byte
		copy(result[1:], bytes) // Left-pad with one zero byte
		return result, nil
	}
	
	// If it's a 40-character hex string (EVM address), left-pad to 32 bytes
	if len(cleanAddr) == 40 {
		evmAddr := common.HexToAddress(address)
		var result [32]byte
		copy(result[12:], evmAddr.Bytes()) // Left-pad with 12 zero bytes
		return result, nil
	}
	
	return [32]byte{}, fmt.Errorf("unsupported address format: %s", address)
}

// IsStarknetAddress checks if an address string represents a Starknet address
func (ac *AddressConverter) IsStarknetAddress(address string) bool {
	cleanAddr := strings.TrimPrefix(address, "0x")
	return len(cleanAddr) == 62
}

// IsEVMAddress checks if an address string represents an EVM address
func (ac *AddressConverter) IsEVMAddress(address string) bool {
	cleanAddr := strings.TrimPrefix(address, "0x")
	return len(cleanAddr) == 40
}

// IsBytes32Address checks if an address string represents a bytes32 address
func (ac *AddressConverter) IsBytes32Address(address string) bool {
	cleanAddr := strings.TrimPrefix(address, "0x")
	return len(cleanAddr) == 64
}

// FormatAddress formats an address string consistently
func (ac *AddressConverter) FormatAddress(address string) string {
	return strings.ToLower(strings.TrimPrefix(address, "0x"))
}
