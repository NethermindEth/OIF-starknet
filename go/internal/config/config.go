package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// SolverConfig represents configuration for a single solver
type SolverConfig struct {
	Enabled bool `json:"enabled"`
}

// Config holds all configuration
type Config struct {
	Solvers    map[string]SolverConfig `json:"solvers"`
	LogLevel   string                  `json:"logLevel"`
	LogFormat  string                  `json:"logFormat"`
	MaxRetries int                     `json:"maxRetries"`
}

// Default solver configurations
var defaultSolvers = map[string]SolverConfig{
	"hyperlane7683": {
		Enabled: true,
	},
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file first
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist, just log a warning
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Create config with defaults
	config := &Config{
		Solvers:    make(map[string]SolverConfig),
		LogLevel:   "info",
		LogFormat:  "text",
		MaxRetries: 5,
	}

	// Copy default solvers
	for name, solver := range defaultSolvers {
		config.Solvers[name] = solver
	}

	// Override with environment variables
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		config.LogFormat = logFormat
	}

	if mr := os.Getenv("MAX_RETRIES"); mr != "" {
		if n, err := strconv.Atoi(mr); err == nil {
			config.MaxRetries = n
		}
	}

	// Allow environment variable override for solver enable/disable
	// Format: SOLVER_HYPERLANE7683_ENABLED=true/false
	for solverName := range config.Solvers {
		envKey := fmt.Sprintf("SOLVER_%s_ENABLED", strings.ToUpper(solverName))
		if enabled := os.Getenv(envKey); enabled != "" {
			solver := config.Solvers[solverName]
			if enabled == "true" {
				solver.Enabled = true
			} else if enabled == "false" {
				solver.Enabled = false
			}
			config.Solvers[solverName] = solver
		}
	}

	return config, nil
}

// IsSolverEnabled checks if a solver is enabled
func (c *Config) IsSolverEnabled(solverName string) bool {
	solver, exists := c.Solvers[solverName]
	if !exists {
		return false
	}
	return solver.Enabled
}
