package envutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConditionalEnv(t *testing.T) {
	t.Run("IS_DEVNET=true uses LOCAL_ variables", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_TEST_VAR", "local_value")
		t.Setenv("TEST_VAR", "regular_value")

		result := GetConditionalEnv("TEST_VAR", "default")
		assert.Equal(t, "local_value", result)
	})

	t.Run("IS_DEVNET=false uses regular variables", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("LOCAL_TEST_VAR", "local_value")
		t.Setenv("TEST_VAR", "regular_value")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_TEST_VAR")
			os.Unsetenv("TEST_VAR")
		}()

		result := GetConditionalEnv("TEST_VAR", "default")
		assert.Equal(t, "regular_value", result)
	})

	t.Run("IS_DEVNET unset defaults to regular variables", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		t.Setenv("LOCAL_TEST_VAR", "local_value")
		t.Setenv("TEST_VAR", "regular_value")
		defer func() {
			os.Unsetenv("LOCAL_TEST_VAR")
			os.Unsetenv("TEST_VAR")
		}()

		result := GetConditionalEnv("TEST_VAR", "default")
		assert.Equal(t, "regular_value", result)
	})

	t.Run("Missing variables fall back to default", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		defer os.Unsetenv("IS_DEVNET")

		result := GetConditionalEnv("NONEXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})
}

func TestGetConditionalUint64(t *testing.T) {
	t.Run("Get conditional uint64 with IS_DEVNET=true", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_TEST_VAR", "123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_TEST_VAR")
		}()

		value := GetConditionalUint64("TEST_VAR", 456, 789)
		assert.Equal(t, uint64(123), value)
	})

	t.Run("Get conditional uint64 with IS_DEVNET=false", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("TEST_VAR", "789")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("TEST_VAR")
		}()

		value := GetConditionalUint64("TEST_VAR", 456, 789)
		assert.Equal(t, uint64(789), value)
	})

	t.Run("Get conditional uint64 with invalid value", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_TEST_VAR", "invalid")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_TEST_VAR")
		}()

		value := GetConditionalUint64("TEST_VAR", 456, 789)
		assert.Equal(t, uint64(789), value) // Should return local default
	})
}

func TestIsForking(t *testing.T) {
	t.Run("Returns true when IS_DEVNET=true", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		defer os.Unsetenv("IS_DEVNET")

		assert.True(t, IsForking())
	})

	t.Run("Returns false when IS_DEVNET=false", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		defer os.Unsetenv("IS_DEVNET")

		assert.False(t, IsForking())
	})

	t.Run("Returns false when IS_DEVNET unset", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")

		assert.False(t, IsForking())
	})
}

func TestGetStarknetAliceAddress(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_STARKNET_ALICE_ADDRESS", "0xlocal123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_STARKNET_ALICE_ADDRESS")
		}()

		result := GetStarknetAliceAddress()
		assert.Equal(t, "0xlocal123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("STARKNET_ALICE_ADDRESS", "0xregular123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("STARKNET_ALICE_ADDRESS")
		}()

		result := GetStarknetAliceAddress()
		assert.Equal(t, "0xregular123", result)
	})

	t.Run("Falls back to default when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("STARKNET_ALICE_ADDRESS")

		result := GetStarknetAliceAddress()
		assert.Equal(t, "0x13d9ee239f33fea4f8785b9e3870ade909e20a9599ae7cd62c1c292b73af1b7", result)
	})
}

func TestGetStarknetAlicePrivateKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_STARKNET_ALICE_PRIVATE_KEY", "0xlocalpriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_STARKNET_ALICE_PRIVATE_KEY")
		}()

		result := GetStarknetAlicePrivateKey()
		assert.Equal(t, "0xlocalpriv123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("STARKNET_ALICE_PRIVATE_KEY", "0xregularpriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("STARKNET_ALICE_PRIVATE_KEY")
		}()

		result := GetStarknetAlicePrivateKey()
		assert.Equal(t, "0xregularpriv123", result)
	})

	t.Run("Returns empty when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("STARKNET_ALICE_PRIVATE_KEY")

		result := GetStarknetAlicePrivateKey()
		assert.Equal(t, "", result)
	})
}

func TestGetStarknetAlicePublicKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_STARKNET_ALICE_PUBLIC_KEY", "0xlocalpub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_STARKNET_ALICE_PUBLIC_KEY")
		}()

		result := GetStarknetAlicePublicKey()
		assert.Equal(t, "0xlocalpub123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("STARKNET_ALICE_PUBLIC_KEY", "0xregularpub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("STARKNET_ALICE_PUBLIC_KEY")
		}()

		result := GetStarknetAlicePublicKey()
		assert.Equal(t, "0xregularpub123", result)
	})

	t.Run("Returns empty when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("STARKNET_ALICE_PUBLIC_KEY")

		result := GetStarknetAlicePublicKey()
		assert.Equal(t, "", result)
	})
}

func TestGetStarknetSolverAddress(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_STARKNET_SOLVER_ADDRESS", "0xlocalsolver123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_STARKNET_SOLVER_ADDRESS")
		}()

		result := GetStarknetSolverAddress()
		assert.Equal(t, "0xlocalsolver123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("STARKNET_SOLVER_ADDRESS", "0xregularsolver123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("STARKNET_SOLVER_ADDRESS")
		}()

		result := GetStarknetSolverAddress()
		assert.Equal(t, "0xregularsolver123", result)
	})

	t.Run("Falls back to default when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("STARKNET_SOLVER_ADDRESS")

		result := GetStarknetSolverAddress()
		assert.Equal(t, "0x2af9427c5a277474c079a1283c880ee8a6f0f8fbf73ce969c08d88befec1bba", result)
	})
}

func TestGetStarknetSolverPrivateKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_STARKNET_SOLVER_PRIVATE_KEY", "0xlocalsolverpriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_STARKNET_SOLVER_PRIVATE_KEY")
		}()

		result := GetStarknetSolverPrivateKey()
		assert.Equal(t, "0xlocalsolverpriv123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("STARKNET_SOLVER_PRIVATE_KEY", "0xregularsolverpriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("STARKNET_SOLVER_PRIVATE_KEY")
		}()

		result := GetStarknetSolverPrivateKey()
		assert.Equal(t, "0xregularsolverpriv123", result)
	})

	t.Run("Returns empty when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("STARKNET_SOLVER_PRIVATE_KEY")

		result := GetStarknetSolverPrivateKey()
		assert.Equal(t, "", result)
	})
}

func TestGetStarknetSolverPublicKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_STARKNET_SOLVER_PUBLIC_KEY", "0xlocalsolverpub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_STARKNET_SOLVER_PUBLIC_KEY")
		}()

		result := GetStarknetSolverPublicKey()
		assert.Equal(t, "0xlocalsolverpub123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("STARKNET_SOLVER_PUBLIC_KEY", "0xregularsolverpub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("STARKNET_SOLVER_PUBLIC_KEY")
		}()

		result := GetStarknetSolverPublicKey()
		assert.Equal(t, "0xregularsolverpub123", result)
	})

	t.Run("Returns empty when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("STARKNET_SOLVER_PUBLIC_KEY")

		result := GetStarknetSolverPublicKey()
		assert.Equal(t, "", result)
	})
}

func TestGetStarknetRPCURL(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_STARKNET_RPC_URL", "http://localhost:5051")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_STARKNET_RPC_URL")
		}()

		result := GetStarknetRPCURL()
		assert.Equal(t, "http://localhost:5051", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("STARKNET_RPC_URL", "https://starknet-sepolia.public.blastapi.io")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("STARKNET_RPC_URL")
		}()

		result := GetStarknetRPCURL()
		assert.Equal(t, "https://starknet-sepolia.public.blastapi.io", result)
	})

	t.Run("Falls back to default when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("STARKNET_RPC_URL")

		result := GetStarknetRPCURL()
		assert.Equal(t, "http://localhost:5050", result)
	})
}

func TestGetAlicePublicKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_ALICE_PUB_KEY", "0xlocalalicepub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_ALICE_PUB_KEY")
		}()

		result := GetAlicePublicKey()
		assert.Equal(t, "0xlocalalicepub123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("ALICE_PUB_KEY", "0xregularalicepub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("ALICE_PUB_KEY")
		}()

		result := GetAlicePublicKey()
		assert.Equal(t, "0xregularalicepub123", result)
	})

	t.Run("Falls back to default when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("ALICE_PUB_KEY")

		result := GetAlicePublicKey()
		assert.Equal(t, "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", result)
	})
}

func TestGetAlicePrivateKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_ALICE_PRIVATE_KEY", "0xlocalalicepriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_ALICE_PRIVATE_KEY")
		}()

		result := GetAlicePrivateKey()
		assert.Equal(t, "0xlocalalicepriv123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("ALICE_PRIVATE_KEY", "0xregularalicepriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("ALICE_PRIVATE_KEY")
		}()

		result := GetAlicePrivateKey()
		assert.Equal(t, "0xregularalicepriv123", result)
	})

	t.Run("Returns empty when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("ALICE_PRIVATE_KEY")

		result := GetAlicePrivateKey()
		assert.Equal(t, "", result)
	})
}

func TestGetSolverPublicKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_SOLVER_PUB_KEY", "0xlocalsolverpub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_SOLVER_PUB_KEY")
		}()

		result := GetSolverPublicKey()
		assert.Equal(t, "0xlocalsolverpub123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("SOLVER_PUB_KEY", "0xregularsolverpub123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("SOLVER_PUB_KEY")
		}()

		result := GetSolverPublicKey()
		assert.Equal(t, "0xregularsolverpub123", result)
	})

	t.Run("Returns empty when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("SOLVER_PUB_KEY")

		result := GetSolverPublicKey()
		assert.Equal(t, "", result)
	})
}

func TestGetSolverPrivateKey(t *testing.T) {
	t.Run("Uses LOCAL version when forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "true")
		t.Setenv("LOCAL_SOLVER_PRIVATE_KEY", "0xlocalsolverpriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("LOCAL_SOLVER_PRIVATE_KEY")
		}()

		result := GetSolverPrivateKey()
		assert.Equal(t, "0xlocalsolverpriv123", result)
	})

	t.Run("Uses regular version when not forking", func(t *testing.T) {
		t.Setenv("IS_DEVNET", "false")
		t.Setenv("SOLVER_PRIVATE_KEY", "0xregularsolverpriv123")
		defer func() {
			os.Unsetenv("IS_DEVNET")
			os.Unsetenv("SOLVER_PRIVATE_KEY")
		}()

		result := GetSolverPrivateKey()
		assert.Equal(t, "0xregularsolverpriv123", result)
	})

	t.Run("Returns empty when not set", func(t *testing.T) {
		os.Unsetenv("IS_DEVNET")
		os.Unsetenv("SOLVER_PRIVATE_KEY")

		result := GetSolverPrivateKey()
		assert.Equal(t, "", result)
	})
}

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("Returns env var value when set", func(t *testing.T) {
		t.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := GetEnvWithDefault("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", result)
	})

	t.Run("Returns default when env var not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR")

		result := GetEnvWithDefault("TEST_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("Returns default when env var is empty", func(t *testing.T) {
		t.Setenv("TEST_VAR", "")
		defer os.Unsetenv("TEST_VAR")

		result := GetEnvWithDefault("TEST_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})
}

func TestGetEnvUint64(t *testing.T) {
	t.Run("Returns parsed uint64 when valid", func(t *testing.T) {
		t.Setenv("TEST_UINT64", "12345")
		defer os.Unsetenv("TEST_UINT64")

		result := GetEnvUint64("TEST_UINT64", 999)
		assert.Equal(t, uint64(12345), result)
	})

	t.Run("Returns default when invalid", func(t *testing.T) {
		t.Setenv("TEST_UINT64", "invalid")
		defer os.Unsetenv("TEST_UINT64")

		result := GetEnvUint64("TEST_UINT64", 999)
		assert.Equal(t, uint64(999), result)
	})

	t.Run("Returns default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_UINT64")

		result := GetEnvUint64("TEST_UINT64", 999)
		assert.Equal(t, uint64(999), result)
	})
}

func TestGetEnvUint64Any(t *testing.T) {
	t.Run("Returns first valid uint64", func(t *testing.T) {
		t.Setenv("VAR1", "invalid")
		t.Setenv("VAR2", "12345")
		t.Setenv("VAR3", "67890")
		defer func() {
			os.Unsetenv("VAR1")
			os.Unsetenv("VAR2")
			os.Unsetenv("VAR3")
		}()

		result := GetEnvUint64Any([]string{"VAR1", "VAR2", "VAR3"}, 999)
		assert.Equal(t, uint64(12345), result)
	})

	t.Run("Returns default when none valid", func(t *testing.T) {
		t.Setenv("VAR1", "invalid")
		t.Setenv("VAR2", "also_invalid")
		defer func() {
			os.Unsetenv("VAR1")
			os.Unsetenv("VAR2")
		}()

		result := GetEnvUint64Any([]string{"VAR1", "VAR2"}, 999)
		assert.Equal(t, uint64(999), result)
	})

	t.Run("Returns default when none set", func(t *testing.T) {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")

		result := GetEnvUint64Any([]string{"VAR1", "VAR2"}, 999)
		assert.Equal(t, uint64(999), result)
	})
}

func TestGetEnvInt(t *testing.T) {
	t.Run("Returns parsed int when valid", func(t *testing.T) {
		t.Setenv("TEST_INT", "12345")
		defer os.Unsetenv("TEST_INT")

		result := GetEnvInt("TEST_INT", 999)
		assert.Equal(t, 12345, result)
	})

	t.Run("Returns default when invalid", func(t *testing.T) {
		t.Setenv("TEST_INT", "invalid")
		defer os.Unsetenv("TEST_INT")

		result := GetEnvInt("TEST_INT", 999)
		assert.Equal(t, 999, result)
	})

	t.Run("Returns default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_INT")

		result := GetEnvInt("TEST_INT", 999)
		assert.Equal(t, 999, result)
	})
}

func TestParseUint64(t *testing.T) {
	t.Run("Parses valid uint64", func(t *testing.T) {
		result, err := parseUint64("12345")
		assert.NoError(t, err)
		assert.Equal(t, uint64(12345), result)
	})

	t.Run("Parses zero", func(t *testing.T) {
		result, err := parseUint64("0")
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), result)
	})

	t.Run("Parses large number", func(t *testing.T) {
		result, err := parseUint64("18446744073709551615") // max uint64
		assert.NoError(t, err)
		assert.Equal(t, uint64(18446744073709551615), result)
	})

	t.Run("Returns error for invalid string", func(t *testing.T) {
		_, err := parseUint64("invalid")
		assert.Error(t, err)
	})

	t.Run("Returns error for negative number", func(t *testing.T) {
		_, err := parseUint64("-1")
		assert.Error(t, err)
	})

	t.Run("Returns error for empty string", func(t *testing.T) {
		_, err := parseUint64("")
		assert.Error(t, err)
	})
}
