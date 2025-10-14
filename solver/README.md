# Hyperlane7683 Solver - Go Implementation

This (Golang) solver is an extension to BootNodeDev's Hyperlane7683 (Typescript) solver adding support for Starknet. This codebase should be used as a reference for protocols to implement or extend.

## Overview

The solver listens for `Open` events from Hyperlane7683 contracts on Starknet and multiple EVM chains, then fills the intents based on configurable rules.

## 🚀 Current Status

**🎉 (Local Sepolia) solves all 3 order types on local forks**: Opens, Fills, and Settles EVM->EVM, EVM->Starknet & Starknet->EVM orders. Requires spoofing a call to each EVM Hyperlane7683 contract to register the Starknet domain

**🎉 (Live Sepolia) fully solves 2/3 order types on live Sepolia:**: Only Opens & Fills Starknet->EVM orders. The Settle call is awaiting Hyperlane to register the Starknet domain on each EVM contract.

## Prerequisites

### Required Dependencies

#### 1. Go (Version 1.25.1+)

```bash
# Install Go 1.25.1 or later
# Visit https://golang.org/dl/ or use your package manager
go version  # Should show 1.25.1 or later
```

#### 2. Foundry (Includes Anvil)

Foundry is a complete toolkit that includes `forge`, `anvil`, `cast`, and `chisel`. When you install Foundry, you get all these tools including Anvil.

```bash
# Install Foundry (includes Anvil)
curl -L https://foundry.paradigm.xyz | bash
foundryup

# Verify installation
forge --version  # Should show 1.3.5-stable or later
anvil --version  # Should show 1.3.5-stable or later
```

#### 3. Dojo/Katana (Version 1.6.2/1.6.3)

Due to contract dependencies requiring spec v0_8 of the Starknet RPC, version 1.6.2 is required for dojo (this comes with katana version 1.6.3).

**Note**: We recommend using [asdf](https://asdf-vm.com/) for version management for Starknet tooling to ensure exact versions.

```bash
# Install asdf (see https://asdf-vm.com/guide/getting-started.html for other installation methods)
brew install asdf

# Install dojo/katana
asdf plugin add dojo
asdf install dojo 1.6.2

# Verify installation
katana --version  # Should show 1.6.3
```

**Important Notes:**

- **Foundry**: Use stable version 1.3.5+, not nightly (nightly can have breaking changes). This repo is tested with 1.3.5.
- **Dojo**: Must be exactly 1.6.2 (includes Katana 1.6.3) due to Starknet RPC spec requirements
- **Go**: Must be 1.25.1 or later for module compatibility. This repo is tested with 1.25.1.

### Setup

1. **Clone and navigate to the solver directory:**

   ```bash
   git clone <repository-url>
   cd oif-starknet/solver
   ```

2. **Install Go dependencies:**

   ```bash
   go mod tidy
   ```

3. **Configure environment:**

   ```bash
   cp example.env .env
   # Edit .env with your configuration (see Configuration section below)
   ```

4. **Build the solver with clean state:**
   ```bash
   make rebuild
   ```

## Configuration

The solver is configured via environment variables. Copy `example.env` to `.env` and edit as needed:

- **For local testing**: The example.env is pre-configured for local development
- **For live Sepolia**: Fill in the non-local variables (RPC URLs, API keys, PKs, etc.)

```bash
cp example.env .env
```

### Running Locally

For local runs, you'll need 3 terminals. All commands should be run from the `solver/` directory.

**Terminal 1: Start local network forks**

```bash
cd solver
make start-networks             # Start local forked networks (EVM + Starknet)
# This runs continuously - keep this terminal open
```

**Terminal 2: Setup and run solver**

```bash
cd solver
make register-starknet-on-evm   # Register Starknet domain on EVM contracts
make fund-accounts              # Fund test accounts with tokens
make run-local                  # Start the solver
```

**Terminal 3: Create test orders**

```bash
cd solver
make open-random-evm-order      # EVM → EVM order
make open-random-evm-sn-order   # EVM → Starknet order
make open-random-sn-order       # Starknet → EVM order
```

### Live Network Testing

For live Sepolia testing, ensure you have:

1. Valid RPC endpoints (Alchemy recommended)
2. Funded accounts with testnet ETH
3. Set `IS_DEVNET=false` in your `.env` file

```bash
# Single terminal setup for live networks
make run-live                   # Run solver on live networks
```

## Testing

### Unit Tests (No RPC Required)

```bash
make test-unit
```

### Local Network Tests

```bash
# Terminal 1: Start networks
make start-networks

# Terminal 2: Run tests
make test-rpc-local            # RPC connectivity tests
make test-integration-local    # Basic integration tests
make test-solver-local         # Full solver integration tests
```

### Live Network Tests

```bash
make test-rpc-live             # RPC connectivity tests
make test-integration-live     # Basic integration tests
make test-solver-live          # Full solver integration tests
```

## Order Lifecycle

1. **Opened on origin**: Alice locks input tokens into the origin chain's hyperlane contract
2. **Fill on destination**: Solver sends output tokens to Alice's destination chain wallet
3. **Settle on destination**: Prevents double-filling, triggers dispatch
4. **Hyperlane dispatch**: Releases locked input tokens to solver (handled by Hyperlane protocol)

## Architecture

```js
solver/
├── cmd/                              # CLI entry points
│   ├── open-order/                   # Create orders (EVM & Starknet)
│   ├── setup-forks/                  # Setup local testnet forks
│   └── solver/                       # Main solver binary
├── solvercore/                       # Core solver logic
│   ├── base/                         # Core interfaces (listener & solver)
│   ├── config/                       # Configuration management
│   ├── contracts/                    # Contract bindings & deployments
│   ├── logutil/                      # Logging utilities
│   ├── solvers/hyperlane7683/        # Hyperlane7683 solver implementation
│   │   ├── chain_handler.go          # Chain handler interface definition
│   │   ├── hyperlane_evm.go          # EVM chain operations (fill/settle)
│   │   ├── hyperlane_starknet.go     # Starknet chain operations (fill/settle)
│   │   ├── listener_base.go          # Common listener logic & block processing
│   │   ├── listener_evm.go           # EVM event listener & processing
│   │   ├── listener_starknet.go      # Starknet event listener & processing
│   │   ├── rules.go                  # Intent validation rules & profitability
│   ├── types/                        # Cross-chain data structures
│   │   └── solver.go                 # Main solver orchestration & chain routing
│   └── solver_manager.go             # Solver orchestration & lifecycle
├── pkg/                              # Public utilities
│   ├── envutil/                      # Environment variable utilities
│   ├── ethutil/                      # Ethereum utilities
│   └── starknetutil/                 # Starknet utilities
└── state/                            # Persistent state storage
```

## Key Files in `solvers/hyperlane7683/`

### Core Orchestration

- **`solver.go`** - Main solver orchestration, chain routing, and multi-instruction support
- **`chain_handler.go`** - Defines the `ChainHandler` interface for chain-specific operations

### Chain-Specific Operations

- **`hyperlane_evm.go`** - EVM chain operations (fill orders, settle orders, balance checks)
- **`hyperlane_starknet.go`** - Starknet chain operations (fill orders, settle orders, balance checks)

### Event Processing

- **`listener_evm.go`** - EVM event listener, processes `Open` events from EVM chains
- **`listener_starknet.go`** - Starknet event listener, processes `Open` events from Starknet
- **`listener_base.go`** - Common listener logic, block range processing, eliminates duplication

### Validation & Rules

- **`rules.go`** - Intent validation rules, profitability analysis, balance checks, allow/block lists

### Key Design Patterns

#### Interface-Based Multi-Chain Architecture

- `Listener` interface enables any blockchain to plug into the system
- `ChainHandler` interface provides common intent processing pipeline
- Chain-specific implementations handle translation between common types and native operations

#### Translation Layer Strategy

**Level 1: Chain Events → Common Format**

```
EVM Open Event → ParsedArgs
Starknet Open Event → ParsedArgs
```

**Level 2: Common Format → Chain Operations**

```
ParsedArgs → EVM Fill Transaction (hyperlane_evm.go)
ParsedArgs → Starknet Fill Transaction (hyperlane_starknet.go)
```

#### Concurrent Multi-Network Processing

- Each network runs its own goroutine-based listener
- All events flow through a unified handler for consistent processing
- Context-based cancellation enables graceful shutdown across all networks

## Configuration

The solver uses environment variables to manage:

- RPC endpoints for different chains
- Private keys for transaction signing
- Contract addresses
- Operational parameters (polling intervals, gas limits, starting block numbers, etc.)

## Extending

To add support for a new blockchain (e.g., Solana):

1. **Create listener**: `listener_solana.go` implementing `Listener`
2. **Create operations**: `hyperlane_solana.go` with Solana-specific fill logic
3. **Update routing**: Add Solana case in `solver.go` destination routing
4. **Add config**: Network configuration in `solvercore/config/networks.go`

## Troubleshooting

### Common Issues

#### 1. Go Version Mismatch

```bash
# If you get Go version errors, ensure you have Go 1.25.1+
go version
# If needed, update Go or use asdf for version management
```

#### 2. Foundry/Anvil Not Found

```bash
# Ensure Foundry is properly installed and in PATH
foundryup
source ~/.bashrc  # or ~/.zshrc
anvil --version
```

#### 2a. Switching from Foundry Nightly to Stable

```bash
# If you're on nightly and want to switch to stable
foundryup  # This will install the latest stable (1.3.5+)
# Or specify exact version: foundryup --version 1.3.5
# Or use asdf for better version management (see Version Management section)
```

#### 3. Katana Not Found

```bash
# Ensure Katana 1.6.3 is installed
katana --version
# If missing, reinstall with the exact version
curl -L https://github.com/dojoengine/dojo/releases/download/v1.6.2/katana-installer.sh | bash
```

#### 4. Network Connection Issues

```bash
# For local testing, ensure all networks are running
make kill-all  # Clean up any stuck processes
make start-networks  # Restart networks

# For live testing, verify RPC endpoints in .env
# Check ALCHEMY_API_KEY is set correctly
```

#### 5. Solver State Issues

```bash
# Clean solver state if experiencing issues
make clean-solver
# This resets the solver to start from configured blocks
```

### Getting Help

- Check the logs for specific error messages
- Ensure all dependencies are at the correct versions
- Verify your `.env` configuration matches your setup (local vs live)
- For live networks, ensure you have sufficient testnet ETH

## License

Apache-2.0
