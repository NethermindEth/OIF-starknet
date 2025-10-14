# Hyperlane7683 Solver - Go Implementation

This (Golang) solver is an extension to BootNodeDev's Hyperlane7683 (Typescript) solver adding support for Starknet. This codebase should be used as a reference for protocols to implement or extend.

## Overview

The solver listens for `Open` events from Hyperlane7683 contracts on Starknet and multiple EVM chains, then fills the intents based on configurable rules.

## Order Lifecycle

1. **Opened on origin**: Alice locks input tokens into the origin chain's hyperlane contract
   - Alice calls `OriginChainHyperlane7683::open(...)`
2. **Fill on destination**: Solver sends output tokens to Alice's destination chain wallet
   - Solver calls `DestinationChainHyperlane7683::fill(...)`
3. **Settle on destination**: Solver sends this followup txn to prevent double-filling, triggers settlement dispatch
   - Solver calls `DestinationChainHyperlane7683::settle(...)`
4. **Hyperlane dispatch**: Releases locked input tokens to solver (handled by Hyperlane protocol)
   - Hyperlane protocol detects the settlement and then routes the settlement throught the `OriginChainHyperlane7683` contract

**Note:** Step 4 is out of scope for this repo (as well as the original BootNodeDev implementation)

## 🚀 Current Status

**🎉 (Local Sepolia Forks) solves all 3 order types**: Opens, Fills, and Settles EVM->EVM, EVM->Starknet & Starknet->EVM orders. Requires spoofing a call to each EVM Hyperlane7683 contract to register the Starknet domain.

**🎉 (Live Sepolia) fully solves 2/3 order types on live Sepolia:**: For Starknet-EVM orders, only Opens & Fills can be called. The Settle call needs Hyperlane to register the Starknet domain on each EVM contract.

## Initial Setup

1. **Clone and navigate to the solver directory:**

   ```bash
   git clone https://github.com/NethermindEth/oif-starknet.git
   cd oif-starknet/solver
   ```

2. **Install Go dependencies:**

   ```bash
   go mod tidy
   ```

## Required Dependencies

- [[Golang 1.25.1+]](https://go.dev/) - Repo is tested with 1.25.1

   ```bash
   # Verify installation
   go version  # Should show 1.25.1 or later
   ```

- [[Foundry 1.4.0+]](https://getfoundry.sh/) - Repo is tested with 1.4.0

   ```bash
   # Verify installation
   anvil --version  # Should show 1.4.0 or later
   ```

- [[Dojo 1.7.1]](https://book.dojoengine.org/installation) - Repo is tested with 1.7.1
   - Comes with Katana v1.7.0
   - [asdf](https://asdf-vm.com/) reccomended

   ```bash
   # Verify installation
   katana --version  # Should show 1.7.0 or later
   ```

- [Alechemy API key](https://book.dojoengine.org/installation)
   - App configured for Sepolia: Ethereum, Arbitrum, Base, Optimism & Starknet
   - Used in [`.env`](example.env)

## Configuration

The solver is configured via environment variables. First, copy `example.env` to `.env` and fill in your Alchemy API key/URLs. The LOCAL addresses are already configured to anvil/katana defaults, but you must fill out the other vars to use/test the solver on live Sepolia networks.

```bash
cp example.env .env
```

### Running the Solver Locally

For local runs, you'll need 3 terminals. All commands should be run from the `solver/` directory.

**Terminal 1: Start local network forks**

```bash
cd solver
make start-networks                   # Start local forked networks (EVMs + Starknet)
# This runs continuously...
```

**Terminal 2: Setup and run solver**

```bash
cd solver
make rebuild                          # Rebuild the solver and ensure a clean starting state (do not do this to let the solver backfill from where it left off if stopped and restarted)
make register-starknet-on-evm-local   # Register Starknet domain on EVM contracts
make fund-accounts-local              # Fund accounts with test ERC-20 tokens
make run-local                        # Start the solver
# This runs continuously...
```

**Terminal 3: Create test orders**

```bash
cd solver
# Can do these back to back or one at a time and watch the logs in the other 2 terminals
make open-random-evm-order-local      # EVM → EVM order
make open-random-evm-sn-order-local   # EVM → Starknet order
make open-random-sn-order-local       # Starknet → EVM order

# First you'll see the order being created in the network logs. 
# Shortly after this you'll see the solver detect the order and begin completing it.
```

## Running the Solver Live

For live Sepolia testing, ensure you have funded accounts with testnet ETH. For live runs, you'll need 2 terminals.

**Terminal 1: Setup and run solver**

```bash
cd solver
make fund-accounts-live               # Fund your (deployer & alice) accounts with test ERC-20 tokens
make run-live                         # Start the solver
# This runs continuously...
```

**Terminal 2: Create test orders**

```bash
cd solver
# Can do these back to back or one at a time and watch the logs in the other terminal
make open-random-evm-order-live       # EVM → EVM order
make open-random-evm-sn-order-live    # EVM → Starknet order
make open-random-sn-order-live        # Starknet → EVM order

# You'll see the solver detect the order and begin completing it shortly after creation.
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
make test-integration-local    # Basic integration tests (opening orders)
make test-solver-local         # Full solver integration tests (opening orders and completing them)
```

### Live Network Tests

```bash
make test-rpc-live             # RPC connectivity tests
make test-integration-live     # Basic integration tests
make test-solver-live          # Full solver integration tests
```

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

## Extending

To add support for a new blockchain (e.g., Solana):

1. **Create listener**: `listener_solana.go` implementing `Listener`
2. **Create operations**: `hyperlane_solana.go` with Solana-specific fill logic
3. **Update routing**: Add Solana case in `solver.go` destination routing
4. **Add config**: Network configuration in `solvercore/config/networks.go`

## License

Apache-2.0
