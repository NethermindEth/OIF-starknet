# Hyperlane7683 Solver - Go Implementation

This (Golang) solver is an extension to BootNodeDev's Hyperlane7683 (Typescript) solver adding support for Starknet. This codebase should be used as a reference for protocols to implement or extend.

## Overview

The solver listens for `Open` events from Hyperlane7683 contracts on Starknet and multiple EVM chains (orders are opened on the "origin" chain), then fills and settles* them on the "destination" chain (origin chain != destination chain). 

## Order Lifecycle

1. **Opened on origin**: Alice locks input tokens into the origin chain's hyperlane contract
2. **Fill on destination**: Solver sends output tokens to Alice's destination chain wallet (using the destination chain hyperlane contract)
3. **Settle on destination**: A simple txn sent after filling to prevent double-filling, triggers dispatch for settlement
4. **Hyperlane dispatch settlement on origin**: Releases locked input tokens to solver on origin chain (handled by Hyperlane protocol; not in the scope of the solver)

## 🚀 Current Status

**🎉 (Local Sepolia) solves all 3 order types on local forks:** Opens, Fills, and Settles EVM->EVM, EVM->Starknet & Starknet->EVM orders. Requires spoofing a call to each EVM Hyperlane7683 contract to register the Starknet domain.

**🎉 (Live Sepolia) fully solves 2/3 order types on live Sepolia:** Opens, Fills, and Settles EVM->EVM & EVM->Starknet orders, only Opens & Fills Starknet->EVM orders. The Settle call is awaiting Hyperlane to register the Starknet domain on each EVM contract.

> **NOTE:** The Starknet Hyperlane7683 contract was deployed [here](https://sepolia.voyager.online/contract/0x002369427e2142db4dfac3a61f5ea7f084e3a74f4c444b5c4e6192a12e49a349) August 29th (2025) in-house; allowing the same deploying address (the owner) to also register the EVM domains on Starknet. This is why EVM->Starknet orders can fully solve, but Starknet->EVM cannot on live Sepolia (on local forks, we can spoof the EVM calls necessary to register the Starknet domain).

## Quick Start

1. Ensure .env is sourced:

   ```bash
   source .env
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Configure your environment:

   ```bash
   cp example.env .env
   # Edit .env with your configuration
   ```

4. Run the solver:
   ```bash
   make run
   ```

## Testing

### Unit Tests (No Networks Required)

- **What it tests:** Core package functionality without network dependencies

```bash
make test-unit
```

### RPC Tests (Requires Networks)

- **What it tests:** Basic RPC connections
- **Requirements:** For local tests, the forked networks must be running in the background (`make start-networks`). 

```bash
# Test with local devnet
make test-rpc

# Test with live networks
make test-rpc-live
```

### Integration Tests (Requires Networks)

- **What it tests:** Order opening and creation
- **Requirements:** For local tests, the forked networks must be running in the background (`make start-networks`). For live tests, accounts must be funded (see [Running the Solver](#running-the-solver))

```bash
# Test with local devnet
make test-integration

# Test with live testnets
make test-integration-live
```

### Solver Integration Tests

- **What it tests:** Solver completes an order's lifecycle (Open → Fill → Settle)
- **Requirements:** For local tests, the forked networks must be running in the background (`make start-networks`). For live tests, accounts must be funded (see [Running the Solver](#running-the-solver))

```bash
# Test with local devnet
make test-solver

# Test with live testnets
make test-solver-live
```

## Running the Solver

### Using Local Forked Networks

For an efficient setup, open 3 terminals and move each to the `solver/` directory. Make sure your `FORKING` env var is set to `true` in your `.env` file.

**Terminal 1: Start networks (runs continuously)**

```bash
make kill-all                   # Ensures no background solvers are running and cleans the solver's state
make rebuild                    # Rebuild binaries
make start-networks             # Start local forked networks (Starknet + EVMs)
```

**Terminal 2: Setup and run solver**

```bash
make register-starknet-on-evm   # Spoof call to register Starknet domain on each EVM contract
make fund-accounts              # Fund Dog coins to Alice and the Solver on all networks
make run                        # Start the solver
```

**Terminal 3: Create orders**

```bash
make open-random-evm-order    # EVM → EVM order
make open-random-evm-sn-order # EVM → Starknet order
make open-random-sn-order     # Starknet → EVM order
```

### Using Live Networks

Before running or testing the solver on Sepolia, you must first fund the order opening wallet (Alice) and the Solver wallet with test tokens. Each ERC-20 contract has a `mint` function that can be used to fund accounts.

> **(Sepolia) Token Addresses:** [Starknet](https://sepolia.voyager.online/token/0x0312be4cb8416dda9e192d7b4d42520e3365f71414aefad7ccd837595125f503), [Ethereum](https://sepolia.etherscan.io/token/0x76878654a2d96dddf8cf0cfe8fa608ab4ce0d499), [Arbitrum](https://sepolia.arbiscan.io/token/0x1083b934abb0be83aae6579c6d5fd974d94e8ea5), [Base](https://sepolia.basescan.org/token/0xb844eed1581f3fb810ffb6dd6c5e30c049cf23f4), [Optimism](https://sepolia-optimism.etherscan.io/token/0xe2f9c9ecab8ae246455be4810cac8fc7c5009150),

Once your accounts are funded (Alice and Solver on all networks), you can run the solver and start opening orders. For this setup, open 2 terminals and move both to the `solver/` directory. Make sure your `FORKING` env var is set to `false` in your `.env` file.

**Terminal 1: Run the solver (runs continuously)**

```bash
make run                        # Starts the solver
```

**Terminal 2: Create orders**

```bash
make open-random-evm-order    # EVM → EVM order
make open-random-evm-sn-order # EVM → Starknet order
make open-random-sn-order     # Starknet → EVM order
```



## Architecture

```js
solver/
├── cmd/                              # CLI entry points
│   ├── main.go                       # Main solver binary entry point
│   ├── solver/                       # Solver-specific CLI commands
│   └── tools/                        # Utility tools and scripts
│       ├── additional-helpers/       # Starknet contract deployment helpers
│       │   ├── declare-sn-hyperlane7683/
│       │   ├── declare-sn-mock-erc20/
│       │   ├── deploy-sn-hyperlane7683/
│       │   ├── deploy-sn-mock-erc20/
│       │   ├── register-evm-routers/
│       │   ├── register-sn-routers/
│       │   ├── setup-starknet-contracts/
│       │   └── verify-hyperlane7683/
│       ├── deploy-forge-mock-erc20/  # Deploy MockERC20 via Forge
│       ├── fund-accounts/            # Fund test accounts with tokens
│       ├── open-order/               # Create orders (EVM & Starknet)
│       └── start-networks.sh         # Start local testnet forks
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
│   │   └── solver.go                 # Main solver orchestration & chain routing
│   ├── types/                        # Cross-chain data structures
│   └── solver_manager.go             # Solver orchestration & lifecycle
├── pkg/                              # Public utilities
│   ├── envutil/                      # Environment variable utilities
│   ├── ethutil/                      # Ethereum utilities
│   └── starknetutil/                 # Starknet utilities
└── state/                            # Persistent state storage
    ├── deployment/                   # Contract deployment artifacts
    └── solver_state/                 # Solver state persistence
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

### Test Files

- **`*_test.go`** - Comprehensive test suites for all components

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

1. **Create listener**: `listener_solana.go` implementing `Listener` in `solvercore/solvers/hyperlane7683/`
2. **Create operations**: `hyperlane_solana.go` with Solana-specific fill logic in `solvercore/solvers/hyperlane7683/`
3. **Update routing**: Add Solana case in `solver.go` destination routing
4. **Add config**: Network configuration in `solvercore/config/networks.go`
5. **Add tools**: Create order creation tools in `cmd/tools/open-order/` if needed

## License

Apache-2.0
