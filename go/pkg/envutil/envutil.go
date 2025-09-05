package envutil

import (
	"fmt"
	"os"
)

const (
	trueValue = "true"
)

// GetConditionalEnv gets an environment variable based on FORKING flag
// If FORKING=true, uses LOCAL_* version, otherwise uses regular version
func GetConditionalEnv(key, defaultValue string) string {
	forking := os.Getenv("FORKING")

	var targetKey string
	if forking == trueValue {
		targetKey = "LOCAL_" + key
	} else {
		targetKey = key
	}

	if value := os.Getenv(targetKey); value != "" {
		return value
	}
	return defaultValue
}

// GetConditionalAccountEnv gets account-related environment variables based on FORKING flag
// This is a convenience function for account keys and addresses
func GetConditionalAccountEnv(key string) string {
	return GetConditionalEnv(key, "")
}

// GetConditionalUint64 gets a uint64 environment variable based on FORKING flag
// If FORKING=true, uses LOCAL_* version with local defaults, otherwise uses regular version with live defaults
func GetConditionalUint64(key string, liveDefault, localDefault uint64) uint64 {
	forking := os.Getenv("FORKING")

	var targetKey string
	var defaultValue uint64
	if forking == trueValue {
		targetKey = "LOCAL_" + key
		defaultValue = localDefault
	} else {
		targetKey = key
		defaultValue = liveDefault
	}

	if value := os.Getenv(targetKey); value != "" {
		if parsed, err := parseUint64(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetEnvWithDefault gets an environment variable with a default fallback
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvUint64 gets an environment variable as uint64 with a default fallback
func GetEnvUint64(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := parseUint64(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetEnvUint64Any returns the first present uint64 among keys, or defaultValue
func GetEnvUint64Any(keys []string, defaultValue uint64) uint64 {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			if parsed, err := parseUint64(v); err == nil {
				return parsed
			}
		}
	}
	return defaultValue
}

// GetEnvInt gets an environment variable as int with a default fallback
func GetEnvInt(key string, defaultValue int) int {
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

// IsForking returns true if FORKING environment variable is set to "true"
func IsForking() bool {
	return os.Getenv("FORKING") == "true"
}

// Common environment variable getters with FORKING logic
// These provide type-safe access to commonly used environment variables

// GetStarknetAliceAddress returns Alice's Starknet address based on FORKING flag
func GetStarknetAliceAddress() string {
	return GetConditionalEnv("STARKNET_ALICE_ADDRESS", "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7")
}

// GetStarknetAlicePrivateKey returns Alice's Starknet private key based on FORKING flag
func GetStarknetAlicePrivateKey() string {
	return GetConditionalAccountEnv("STARKNET_ALICE_PRIVATE_KEY")
}

// GetStarknetAlicePublicKey returns Alice's Starknet public key based on FORKING flag
func GetStarknetAlicePublicKey() string {
	return GetConditionalAccountEnv("STARKNET_ALICE_PUBLIC_KEY")
}

// GetStarknetSolverAddress returns solver's Starknet address based on FORKING flag
func GetStarknetSolverAddress() string {
	return GetConditionalEnv("STARKNET_SOLVER_ADDRESS", "0x2af9427c5a277474c079a1283c880ee8a6f0f8fbf73ce969c08d88befec1bba")
}

// GetStarknetSolverPrivateKey returns solver's Starknet private key based on FORKING flag
func GetStarknetSolverPrivateKey() string {
	return GetConditionalAccountEnv("STARKNET_SOLVER_PRIVATE_KEY")
}

// GetStarknetSolverPublicKey returns solver's Starknet public key based on FORKING flag
func GetStarknetSolverPublicKey() string {
	return GetConditionalAccountEnv("STARKNET_SOLVER_PUBLIC_KEY")
}

// GetStarknetRPCURL returns Starknet RPC URL based on FORKING flag
func GetStarknetRPCURL() string {
	return GetConditionalEnv("STARKNET_RPC_URL", "http://localhost:5050")
}

// GetAlicePublicKey returns Alice's EVM public key based on FORKING flag
func GetAlicePublicKey() string {
	return GetConditionalEnv("ALICE_PUB_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
}

// GetAlicePrivateKey returns Alice's EVM private key based on FORKING flag
func GetAlicePrivateKey() string {
	return GetConditionalAccountEnv("ALICE_PRIVATE_KEY")
}

// GetSolverPublicKey returns solver's EVM public key based on FORKING flag
func GetSolverPublicKey() string {
	return GetConditionalAccountEnv("SOLVER_PUB_KEY")
}

// GetSolverPrivateKey returns solver's EVM private key based on FORKING flag
func GetSolverPrivateKey() string {
	return GetConditionalAccountEnv("SOLVER_PRIVATE_KEY")
}
