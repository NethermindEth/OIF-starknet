# Hyperlane7683 Solver - Go Implementation

This (Golang) solver is an extension to BootNodeDev's Hyperlane7683 (Typescript) solver adding support for Starknet. This codebase should be used as a reference for protocols to implement or extend.

## Overview

The solver listens for `Open` events from Hyperlane7683 contracts on Starknet and multiple EVM chains, then fills the intents based on configurable rules.

## Architecture

```js
go/
├── cmd/                              # CLI entry points
│   ├── open-order/                   # Create orders (EVM & Starknet)
│   │   ├── evm/                      # EVM order creation utilities
│   │   └── starknet/                 # Starknet order creation utilities
│   ├── setup-forks/                  # Setup local testnet forks
│   │   ├── evm/                      # EVM fork setup (Anvil)
│   │   └── starknet/                 # Starknet fork setup (Katana)
│   └── solver/                       # Main solver binary
├── internal/                         # Core solver logic
│   ├── config/                       # Configuration management
│   │   ├── config.go                 # Solver configuration
│   │   └── networks.go               # Multi-chain network configs
│   ├── contracts/                    # Go bindings for smart contracts
│   │   ├── erc20_contract.go         # ERC20 contract bindings
│   │   └── hyperlane7683.go          # Hyperlane7683 contract bindings
│   ├── deployer/                     # Deployment state management
│   │   └── deployment_state.go       # Contract deployment tracking
│   ├── filler/                       # Intent filling interface
│   │   └── base_filler.go            # Base filler interface
│   ├── listener/                     # Event listening interface
│   │   └── base_listener.go          # Base listener interface
│   ├── logutil/                      # Terminal logging utilities
│   ├── solvers/                      # Solver implementations
│   │   └── hyperlane7683/            # Hyperlane7683 solver
│   │       ├── filler.go             # Main orchestrator - routes intents to chain-specific handlers
│   │       ├── filler_starknet.go    # Low-level Starknet operations (build/send transactions)
│   │       ├── hyperlane_evm.go      # EVM chain handler (fill/settle/approvals)
│   │       ├── hyperlane_starknet.go # Starknet chain handler (coordinates StarknetFiller)
│   │       ├── listener_evm.go       # EVM Open event listener (polls blocks, parses events)
│   │       ├── listener_starknet.go  # Starknet Open event listener (Cairo event parsing)
│   │       └── rules.go              # Intent validation rules (balance checks, allowlists)
│   ├── types/                        # Cross-chain data structures
│   │   ├── address_utils.go          # Address conversion utilities
│   │   └── types.go                  # Core type definitions
│   └── solver_manager.go             # Solver orchestration & lifecycle
├── pkg/                              # Public utilities
│   └── ethutil/                      # Ethereum utilities (signing, gas, ERC20)
├── state/                            # Persistent state storage
│   └── network_state/                # Network deployment states
├── bin/                              # Built binaries
├── env.example                       # Environment configuration template
├── Makefile                          # Build & deployment automation
├── start-networks.sh                 # Multi-network startup script
└── go.mod                            # Go module dependencies
```

### Key Design Patterns

#### 1. **Interface-Based Multi-Chain Architecture**

- `BaseListener` interface enables any blockchain to plug into the system
- `BaseFiller` interface provides a common intent processing pipeline
- Chain-specific implementations handle translation between common types and native operations

#### 2. **Translation Layer Strategy**

The system uses **multiple translation layers** for maximum extensibility:

**Level 1: Chain Events → Common Format**

```
EVM Open Event → ParsedArgs
Starknet Open Event → ParsedArgs
XYZ Chain Event → ParsedArgs (easy to add)
```

**Level 2: Common Format → Chain Operations**

```
ParsedArgs → EVM Fill Transaction (hyperlane_evm.go)
ParsedArgs → Starknet Fill Transaction (hyperlane_starknet.go)
ParsedArgs → XYZ Fill Transaction (hyperlane_xyz.go - easy to add)
```

#### 3. **Concurrent Multi-Network Processing**

- Each network runs its own goroutine-based listener
- All events flow through a unified handler for consistent processing
- Context-based cancellation enables graceful shutdown across all networks

#### 4. **Extensibility for New VMs**

To add support for a new blockchain (e.g., Solana):

1. **Create listener**: `listener_solana.go` implementing `BaseListener`
2. **Create operations**: `hyperlane_solana.go` with Solana-specific fill logic
3. **Update routing**: Add Solana case in `solver.go` destination routing
4. **Add config**: Network configuration in `internal/config/networks.go`

#### **Context-Based Lifecycle Management**

```go
ctx, cancel := context.WithCancel(context.Background())
// All goroutines respect this context for graceful shutdown
```

#### **Coordinated Goroutine Management**

```go
sm.shutdownWg.Add(1)
go func() {
    defer sm.shutdownWg.Done()
    <-sm.ctx.Done()
    shutdownFunc()  // Clean shutdown per network
}()
```

#### **Multi-Network Concurrent Event Processing**

- Each blockchain network runs in its own goroutine
- Events from all chains feed through the same `EventHandler` function
- Maintains **order integrity** while enabling **parallel processing**
- No blocking between networks - if one network is slow, others continue processing

## 🚀 Current Status

**✅ (LOCAL SEPOLIA) SOLVING ALL (3) ORDER TYPES ON LOCAL FORKS (EVM->EVM, EVM->SN, SN->EVM)**
**✅ (LIVE SEPOLIA) SOLVING MOST (2) ORDER TYPES ON LOCAL FORKS (EVM->EVM, EVM->SN)**: Awaiting Hyperlane contract to register Starknet domain

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

## Configuration

The solver uses environment variables to manage:

- RPC endpoints for different chains
- Private keys for transaction signing
- Contract addresses
- Operational parameters (polling intervals, gas limits, starting block numbers, etc.)

## Running on Local Forks

Besides an Alchemy API key, the `example.env` file has all of the values needed to run the solver locally on forks of Sepolia networks, just make sure you copy them over to a `.env` file. Make sure you have katana and anvil installed before continuing.

For an efficient setup, it is recommended that you open 3 terminals and move each to the `go/` directory.

In the first terminal, run the following command to make sure the state file is clean and the binaries are built:

```bash
make rebuild
```

After this is finished, run the following command to start local (Sepolia) forks of Ethereum, Optimism, Arbitrum, Base, and Starknet. You can leave this terminal running and watch transaction logs come in.

```bash
make start-networks
```

In the second terminal, run this command to deploy a mock ERC-20 token onto each network, fund the accounts on each network, and register the Starknet domain on each EVM Hyperlane7683 contract:

```bash
make setup-forks
```

Once this is finished, start the solver by running:

```bash
make run
```

We will use the third terminal to create orders. There are 3 order commands to choose from for each of the different order types. Run these at will.

```bash
make open-random-evm-order    # Opens a random order from one EVM chain to another

make open-random-evm-sn-order # Opens a random order from an EVM chain to Starknet

make open-random-sn-order     # Opens a random order from Starknet to an EVM chain

```

## Extending

This implementation is designed to be easily extensible:

- Support new chains in `internal/config/networks.go` & `internal/solvers/hyperlane7683/`
- Add new solvers (Eco, Polymer) in `internal/solvers/`

## License

Apache-2.0

```

```
