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
IntentData → EVM Fill Transaction (hyperlane_evm.go)
IntentData → Starknet Fill Transaction (hyperlane_starknet.go)
IntentData → XYZ Fill Transaction (hyperlane_xyz.go - easy to add)
```

#### 3. **Concurrent Multi-Network Processing**

- Each network runs its own goroutine-based listener
- All events flow through a unified handler for consistent processing
- Context-based cancellation enables graceful shutdown across all networks

#### 4. **Extensibility for New VMs**

To add support for a new blockchain (e.g., Solana):

1. **Create listener**: `listener_solana.go` implementing `BaseListener`
2. **Create operations**: `hyperlane_solana.go` with Solana-specific fill logic
3. **Update routing**: Add Solana case in `filler.go` destination routing
4. **Add config**: Network configuration in `internal/config/networks.go`

**The core orchestration code doesn't need to change** - this is the power of the interface-based design.

### Concurrency Architecture

The solver uses **Go concurrency patterns** for high-performance multi-chain processing:

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

### Recommendations for Improvement

#### **Architecture Enhancements**

1. **Database integration**: Persist order state for crash recovery

#### **Testing Infrastructure**

1. **Integration tests**: End-to-end testing across multiple chains
2. **Load testing**: Validate performance under high intent volumes

## 🚀 Current Status

**✅ SOLVING ORDERS ON LOCAL FORKS**

## Quick Start

1. Install dependencies:

   ```bash
   go mod tidy
   ```

2. Configure your environment:

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. Run the solver:
   ```bash
   go run cmd/solver/main.go
   ```

## Configuration

The solver uses environment variables to manage:

- RPC endpoints for different chains
- Private keys for transaction signing
- Contract addresses
- Rule parameters
- Allow/block lists
- Solver enable/disable flags

## Extending

This implementation is designed to be easily extensible:

- Add new rules in `internal/solvers/hyperlane7683/rules.go`
- Support new chains in `internal/config/networks.go`
- Implement custom fillers in `internal/solvers/hyperlane7683/`
- Add new solvers in `internal/solvers/`

## License

Apache-2.0
