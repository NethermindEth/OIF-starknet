# Hyperlane7683 Solver - Go Implementation

This (Golang) solver is an extension to BootNodeDev's Hyperlane7683 (Typescript) solver adding support for Starknet. This codebase should be used as a reference for protocols to implement or extend.

## Overview

The solver listens for `Open` events from Hyperlane7683 contracts on Starknet and multiple EVM chains, then fills the intents based on configurable rules.

## 🚀 Current Status

**🎉 (Local Sepolia) solves all 3 order types on local forks**: Opens, Fills, and Settles EVM->EVM, EVM->Starknet & Starknet->EVM orders. Requires spoofing a call to each EVM Hyperlane7683 contract to register the Starknet domain

**🎉 (Live Sepolia) fully solves 2/3 order types on live Sepolia:**: Only Opens & Fills Starknet->EVM orders. The Settle call is awaiting Hyperlane to register the Starknet domain on each EVM contract.

## Quick Start

1. Install dependencies:

   ```bash
   go mod tidy
   ```

2. Configure your environment:

   ```bash
   cp example.env .env
   # Edit .env with your configuration
   ```

3. Run the solver:
   ```bash
   make run
   ```

## Running on Local Forks

For an efficient setup, open 3 terminals and move each to the `go/` directory. Make sure your `IS_DEVNET` env var is set to true in your `.env` file.

**Terminal 1: Start networks (runs continuously)**

```bash
make rebuild                    # Rebuild contracts & binaries
make start-networks             # Start local forked networks (EVM + Starknet)
```

**Terminal 2: Setup and run solver**

```bash
make register-starknet-on-evm   # Spoof call to register Starknet domain on EVM contracts
make fund-accounts              # Fund Dog coins to Alice and the Solver on all networks
make run                        # Start the solver
```

**Terminal 3: Create orders**

```bash
make open-random-evm-order    # EVM → EVM order
make open-random-evm-sn-order # EVM → Starknet order
make open-random-sn-order     # Starknet → EVM order
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

## Testing

```bash
# Unit tests (no RPC required)
make test-unit

# RPC tests (requires networks)
make start-networks  # Terminal 1
make test-rpc        # Terminal 2

# Integration tests (requires full setup)
make start-networks  # Terminal 1
make fund-accounts register-starknet-on-evm  # Terminal 2
make test-integration
```

## License

Apache-2.0
