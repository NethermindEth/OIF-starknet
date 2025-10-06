package envutil

import (
	"fmt"
	"os"
)

const (
	trueValue = "true"
)

// GetConditionalEnv gets an environment variable based on IS_DEVNET flag
// If IS_DEVNET=true, uses LOCAL_* version, otherwise uses regular version
func GetConditionalEnv(key, defaultValue string) string {
	isDevnet := os.Getenv("IS_DEVNET")

	var targetKey string
	if isDevnet == trueValue {
		targetKey = "LOCAL_" + key
	} else {
		targetKey = key
	}

	if value := os.Getenv(targetKey); value != "" {
		return value
	}
	return defaultValue
}

// GetConditionalAccountEnv gets account-related environment variables based on IS_DEVNET flag
// This is a convenience function for account keys and addresses
func GetConditionalAccountEnv(key string) string {
	return GetConditionalEnv(key, "")
}

// GetConditionalUint64 gets a uint64 environment variable based on IS_DEVNET flag
// If IS_DEVNET=true, uses LOCAL_* version with local defaults, otherwise uses regular version with live defaults
func GetConditionalUint64(key string, liveDefault, localDefault uint64) uint64 {
	isDevnet := os.Getenv("IS_DEVNET")

	var targetKey string
	var defaultValue uint64
	if isDevnet == trueValue {
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

// GetConditionalInt64 gets an int64 environment variable based on IS_DEVNET flag
// If IS_DEVNET=true, uses LOCAL_* version with local defaults, otherwise uses regular version with live defaults
func GetConditionalInt64(key string, liveDefault, localDefault int64) int64 {
	isDevnet := os.Getenv("IS_DEVNET")

	var targetKey string
	var defaultValue int64
	if isDevnet == trueValue {
		targetKey = "LOCAL_" + key
		defaultValue = localDefault
	} else {
		targetKey = key
		defaultValue = liveDefault
	}

	if value := os.Getenv(targetKey); value != "" {
		if parsed, err := parseInt64(value); err == nil {
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

// parseInt64 parses a string to int64
func parseInt64(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// IsDevnet returns true if IS_DEVNET environment variable is set to "true"
func IsDevnet() bool {
	return os.Getenv("IS_DEVNET") == "true"
}

// Common environment variable getters with IS_DEVNET logic
// These provide type-safe access to commonly used environment variables

// GetStarknetAliceAddress returns Alice's Starknet address based on IS_DEVNET flag
func GetStarknetAliceAddress() string {
	return GetConditionalEnv("STARKNET_ALICE_ADDRESS", "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7")
}

// GetStarknetAlicePrivateKey returns Alice's Starknet private key based on IS_DEVNET flag
func GetStarknetAlicePrivateKey() string {
	return GetConditionalAccountEnv("STARKNET_ALICE_PRIVATE_KEY")
}

// GetStarknetAlicePublicKey returns Alice's Starknet public key based on IS_DEVNET flag
func GetStarknetAlicePublicKey() string {
	return GetConditionalAccountEnv("STARKNET_ALICE_PUBLIC_KEY")
}

// GetStarknetSolverAddress returns solver's Starknet address based on IS_DEVNET flag
func GetStarknetSolverAddress() string {
	return GetConditionalEnv("STARKNET_SOLVER_ADDRESS", "0x2af9427c5a277474c079a1283c880ee8a6f0f8fbf73ce969c08d88befec1bba")
}

// GetStarknetSolverPrivateKey returns solver's Starknet private key based on IS_DEVNET flag
func GetStarknetSolverPrivateKey() string {
	return GetConditionalAccountEnv("STARKNET_SOLVER_PRIVATE_KEY")
}

// GetStarknetSolverPublicKey returns solver's Starknet public key based on IS_DEVNET flag
func GetStarknetSolverPublicKey() string {
	return GetConditionalAccountEnv("STARKNET_SOLVER_PUBLIC_KEY")
}

// GetStarknetRPCURL returns Starknet RPC URL based on IS_DEVNET flag
func GetStarknetRPCURL() string {
	return GetConditionalEnv("STARKNET_RPC_URL", "http://localhost:5050")
}

// GetAlicePublicKey returns Alice's EVM public key based on IS_DEVNET flag
func GetAlicePublicKey() string {
	return GetConditionalEnv("ALICE_PUB_KEY", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
}

// GetAlicePrivateKey returns Alice's EVM private key based on IS_DEVNET flag
func GetAlicePrivateKey() string {
	return GetConditionalAccountEnv("ALICE_PRIVATE_KEY")
}

// GetSolverPublicKey returns solver's EVM public key based on IS_DEVNET flag
func GetSolverPublicKey() string {
	return GetConditionalAccountEnv("SOLVER_PUB_KEY")
}

// GetSolverPrivateKey returns solver's EVM private key based on IS_DEVNET flag
func GetSolverPrivateKey() string {
	return GetConditionalAccountEnv("SOLVER_PRIVATE_KEY")
}
