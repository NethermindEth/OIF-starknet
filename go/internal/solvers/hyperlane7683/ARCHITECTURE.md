# Hyperlane7683 Solver Architecture

This document explains the clean, extensible architecture for the Hyperlane7683 solver.

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────┐
│  ChainHandler Interface                     │  
│  - Fill(ctx, args) (OrderAction, error)     │
│  - Settle(ctx, args) error                  │  
│  - GetOrderStatus(ctx, args) (string, error)│
└─────────────────┬───────────────────────────┘
                  │ implements
┌─────────────────▼───────────────────────────┐
│  Hyperlane7683Solver (orchestrator)         │
│  - Routes to appropriate ChainHandler       │
│  - Chain detection & handler management     │
└─────────────────┬───────────────────────────┘
                  │ uses
        ┌─────────┼─────────┐─────────┐
        ▼         ▼         ▼         ▼
   ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
   │   EVM   │ │Starknet │ │ Cosmos  │ │   ???   │
   │Handler  │ │Handler  │ │Handler  │ │Handler  │  
   └─────────┘ └─────────┘ └─────────┘ └─────────┘
```

## 📁 File Structure

- **`chain_handler.go`** - Interface definition & types
- **`solver.go`** - Main orchestrator/router  
- **`hyperlane_evm.go`** - EVM implementation
- **`hyperlane_starknet.go`** - Starknet implementation
- **`listener_evm.go`** - EVM event listening
- **`listener_starknet.go`** - Starknet event listening

## 🔌 Adding New Chain Support

To add support for a new blockchain (e.g., Cosmos), follow these simple steps:

### 1. Create the Chain Handler

```go
// File: hyperlane_cosmos.go
package hyperlane7683

type HyperlaneCosmos struct {
    client *cosmos.Client  // Your cosmos client
    signer *cosmos.Signer  // Your cosmos signer
    mu     sync.Mutex
}

func NewHyperlaneCosmos(rpcURL string) ChainHandler {
    // Initialize cosmos client & signer
    return &HyperlaneCosmos{...}
}

// Implement ChainHandler interface
func (h *HyperlaneCosmos) Fill(ctx context.Context, args types.ParsedArgs) (OrderAction, error) {
    // Your cosmos-specific fill logic
}

func (h *HyperlaneCosmos) Settle(ctx context.Context, args types.ParsedArgs) error {
    // Your cosmos-specific settle logic  
}

func (h *HyperlaneCosmos) GetOrderStatus(ctx context.Context, args types.ParsedArgs) (string, error) {
    // Your cosmos-specific status check
}
```

### 2. Create the Event Listener

```go
// File: listener_cosmos.go  
package hyperlane7683

func NewCosmosListener(config *base.ListenerConfig, rpcURL string) (base.BaseListener, error) {
    // Your cosmos event listener implementation
    // Follow the same pattern as listener_evm.go and listener_starknet.go
}
```

### 3. Add Chain Detection

In `solver.go`, add chain detection logic:

```go
func (f *Hyperlane7683Solver) isCosmosChain(chainID *big.Int) bool {
    config.InitializeNetworks()
    
    for networkName, network := range config.Networks {
        if network.ChainID == chainID.Uint64() {
            return strings.Contains(strings.ToLower(networkName), "cosmos")
        }
    }
    return false
}
```

### 4. Add Handler Creation

In `solver.go`, add handler creation:

```go
func (f *Hyperlane7683Solver) getCosmosHandler(chainID *big.Int) (ChainHandler, error) {
    if f.hyperlaneCosmos != nil {
        return f.hyperlaneCosmos, nil
    }
    
    chainConfig, err := f.getNetworkConfigByChainID(chainID)
    if err != nil {
        return nil, fmt.Errorf("cosmos network not found: %w", err)
    }
    
    f.hyperlaneCosmos = NewHyperlaneCosmos(chainConfig.RPCURL)
    return f.hyperlaneCosmos, nil
}
```

### 5. Update Router Logic

In `solver.go`, add cosmos case to Fill() and SettleOrder():

```go
case f.isCosmosChain(instruction.DestinationChainID):
    handler, err := f.getCosmosHandler(instruction.DestinationChainID)
    if err != nil {
        return OrderActionError, fmt.Errorf("failed to get Cosmos handler: %w", err)
    }
    return handler.Fill(ctx, args)
```

### 6. Register in Config

Add cosmos networks to your `.env` and `config/networks.go`.

That's it! 🎉 The solver will automatically route cosmos orders to your new handler.

## 🎯 Key Benefits

1. **Clean Interface** - All chains implement the same 3 methods
2. **Easy Extension** - Just implement `ChainHandler` interface  
3. **No Base Class Complexity** - Simple, focused architecture
4. **Consistent Patterns** - Fill → Settle → Status workflow
5. **Proper Separation** - Each chain handles its own complexity

## 🔄 Migration from Old Architecture

The old architecture had:
- ❌ Confusing `BaseSolver` interface mismatch
- ❌ Complex inheritance with `BaseSolverImpl`  
- ❌ Inconsistent method signatures
- ❌ Hard to extend

The new architecture has:
- ✅ Simple `ChainHandler` interface
- ✅ Clean orchestrator pattern
- ✅ Consistent method signatures
- ✅ Easy to extend with new chains
