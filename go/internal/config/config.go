package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// SolverConfig represents configuration for a single solver
type SolverConfig struct {
	Enabled bool `json:"enabled"`
}

// Config holds all configuration
type Config struct {
	Solvers     map[string]SolverConfig `json:"solvers"`
	PrivateKey  string                  `json:"privateKey"`
	LogLevel    string                  `json:"logLevel"`
	LogFormat   string                  `json:"logFormat"`
	MaxRetries  int                     `json:"maxRetries"`
}

// Default solver configurations
var defaultSolvers = map[string]SolverConfig{
	"hyperlane7683": {
		Enabled: true,
	},
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig() (*Config, error) {
	// Load .env file first
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist, just log a warning
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("solvers", defaultSolvers)
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("logFormat", "text")
	viper.SetDefault("maxRetries", 5)

	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Override with environment variables
	if privateKey := os.Getenv("PRIVATE_KEY"); privateKey != "" {
		viper.Set("privateKey", privateKey)
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		viper.Set("logLevel", logLevel)
	}

	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		viper.Set("logFormat", logFormat)
	}

	if mr := os.Getenv("MAX_RETRIES"); mr != "" {
		if n, err := strconv.Atoi(mr); err == nil {
			viper.Set("maxRetries", n)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// IsSolverEnabled checks if a solver is enabled
func (c *Config) IsSolverEnabled(solverName string) bool {
	solver, exists := c.Solvers[solverName]
	if !exists {
		return false
	}
	return solver.Enabled
}
