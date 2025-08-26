package hyperlane7683

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	contracts "github.com/NethermindEth/oif-starknet/go/internal/contracts"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/listener"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Open event topic: Open(bytes32,ResolvedCrossChainOrder)
var openEventTopic = common.HexToHash("0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d")

// evmListener implements listener.BaseListener for EVM chains for Hyperlane7683
type evmListener struct {
	config             *listener.ListenerConfig
	client             *ethclient.Client
	contractAddress    common.Address
	lastProcessedBlock uint64
	stopChan           chan struct{}
	mu                 sync.RWMutex
	// Add cooldown tracking for failed blocks
	failedBlocks map[uint64]time.Time
	failedMu     sync.RWMutex
}

func NewEVMListener(config *listener.ListenerConfig, rpcURL string) (listener.BaseListener, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	// Always use the last processed block from deployment state
	var lastProcessedBlock uint64
	state, err := deployer.GetDeploymentState()
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment state: %w", err)
	}

	if networkState, exists := state.Networks[config.ChainName]; exists {
		lastProcessedBlock = networkState.LastIndexedBlock
		fmt.Printf("📚 %s: Using persisted LastIndexedBlock: %d\n", config.ChainName, lastProcessedBlock)
	} else {
		return nil, fmt.Errorf("network %s not found in deployment state", config.ChainName)
	}

	return &evmListener{
		config:             config,
		client:             client,
		contractAddress:    common.HexToAddress(config.ContractAddress),
		lastProcessedBlock: lastProcessedBlock,
		stopChan:           make(chan struct{}),
		failedBlocks:       make(map[uint64]time.Time),
	}, nil
}

// Start begins listening for events
func (l *evmListener) Start(ctx context.Context, handler listener.EventHandler) (listener.ShutdownFunc, error) {
	go l.realEventLoop(ctx, handler)
	return func() { close(l.stopChan) }, nil
}

// Stop gracefully stops the listener
func (l *evmListener) Stop() error {
	fmt.Printf("Stopping EVM listener...\n")
	close(l.stopChan)
	return nil
}

// GetLastProcessedBlock returns the last processed block number
func (l *evmListener) GetLastProcessedBlock() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastProcessedBlock
}

// MarkBlockFullyProcessed marks a block as fully processed and updates LastIndexedBlock
func (l *evmListener) MarkBlockFullyProcessed(blockNumber uint64) error {
	if blockNumber != l.lastProcessedBlock+1 {
		return fmt.Errorf("cannot mark block %d as processed, expected %d", blockNumber, l.lastProcessedBlock+1)
	}
	l.lastProcessedBlock = blockNumber
	fmt.Printf("✅ Block %d marked as fully processed for %s\n", blockNumber, l.config.ChainName)
	return nil
}

func (l *evmListener) realEventLoop(ctx context.Context, handler listener.EventHandler) {
	fmt.Printf("⚙️  Starting (%s) event listener...\n", l.config.ChainName)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		fmt.Printf("❌ Failed to catch up on (%s) historical blocks: %v\n", l.config.ChainName, err)
	}
	fmt.Printf("🔄 Backfill complete (%s)\n", l.config.ChainName)
	time.Sleep(1 * time.Second)
	l.startPolling(ctx, handler)
}

// cleanupFailedBlocks removes expired cooldown entries
func (l *evmListener) cleanupFailedBlocks() {
	l.failedMu.Lock()
	defer l.failedMu.Unlock()
	
	cutoff := time.Now().Add(-10 * time.Minute) // Remove entries older than 10 minutes
	for block, failTime := range l.failedBlocks {
		if failTime.Before(cutoff) {
			delete(l.failedBlocks, block)
		}
	}
}

func (l *evmListener) processCurrentBlockRange(ctx context.Context, handler listener.EventHandler) error {
	// Clean up old failed blocks periodically
	l.cleanupFailedBlocks()
	
	currentBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}
	// Apply confirmations window if configured
	safeBlock := currentBlock
	if l.config.ConfirmationBlocks > 0 && currentBlock > l.config.ConfirmationBlocks {
		safeBlock = currentBlock - l.config.ConfirmationBlocks
	}
	if safeBlock <= l.lastProcessedBlock {
		return nil
	}
	fromBlock := l.lastProcessedBlock + 1
	toBlock := safeBlock
	fmt.Printf("🧭 %s EVM range: from=%d to=%d (current=%d, conf=%d)\n", l.config.ChainName, fromBlock, toBlock, currentBlock, l.config.ConfirmationBlocks)
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range for %s: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return nil
	}
	newLast, err := l.processBlockRange(ctx, fromBlock, toBlock, handler)
	if err != nil {
		return fmt.Errorf("failed to process blocks %d-%d: %v", fromBlock, toBlock, err)
	}
	
	fmt.Printf("🔍 DEBUG %s: processBlockRange returned newLast=%d, current lastProcessedBlock=%d\n", l.config.ChainName, newLast, l.lastProcessedBlock)
	
	l.lastProcessedBlock = newLast
	if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, newLast); err != nil {
		fmt.Printf("⚠️  Failed to persist LastIndexedBlock for %s: %v\n", l.config.ChainName, err)
	} else {
		fmt.Printf("💾 Persisted LastIndexedBlock=%d for %s\n", newLast, l.config.ChainName)
	}
	return nil
}

// processBlockRange processes logs in [fromBlock, toBlock] and returns the highest contiguous block fully processed
func (l *evmListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler listener.EventHandler) (uint64, error) {
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range (%s) in processBlockRange: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return l.lastProcessedBlock, nil
	}
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{l.contractAddress},
		Topics:    [][]common.Hash{{openEventTopic}},
	}
	fmt.Printf("🔎 %s filter: addr=%s, topic0=%s, from=%d, to=%d\n", l.config.ChainName, l.contractAddress.Hex(), openEventTopic.Hex(), fromBlock, toBlock)
	logs, err := l.client.FilterLogs(ctx, query)
	if err != nil {
		return l.lastProcessedBlock, fmt.Errorf("failed to filter logs: %v", err)
	}
	fmt.Printf("📩 %s logs found: %d\n", l.config.ChainName, len(logs))
	if len(logs) > 0 {
		fmt.Printf("📩 Found %d Open events on %s\n", len(logs), l.config.ChainName)
	}

	// group logs by block
	byBlock := make(map[uint64][]gethtypes.Log)
	for _, lg := range logs {
		byBlock[lg.BlockNumber] = append(byBlock[lg.BlockNumber], lg)
	}

	// iterate blocks in order
	newLast := l.lastProcessedBlock
	maxRetries := config.GetDefaultNetwork() // placeholder to avoid unused import
	_ = maxRetries
	for b := fromBlock; b <= toBlock; b++ {
		// Check if this block is in cooldown
		l.failedMu.RLock()
		if failTime, exists := l.failedBlocks[b]; exists {
			if time.Since(failTime) < 5*time.Minute { // 5 minute cooldown
				fmt.Printf("   ⏸️  Block %d in cooldown (failed at %v), skipping\n", b, failTime)
				l.failedMu.RUnlock()
				continue
			} else {
				// Remove expired cooldown
				l.failedMu.RUnlock()
				l.failedMu.Lock()
				delete(l.failedBlocks, b)
				l.failedMu.Unlock()
			}
		} else {
			l.failedMu.RUnlock()
		}
		
		retryCount := 0
		failed := false
		events := byBlock[b]
		
		// Track which orders in this block are settled
		blockOrders := len(events)
		settledOrders := 0
		
		for {
			blockFailed := false
			settledOrders = 0 // Reset for retry
			
			for _, lg := range events {
				// Use generated binding to parse Open events
				filterer, ferr := contracts.NewHyperlane7683Filterer(l.contractAddress, l.client)
				if ferr != nil {
					return newLast, fmt.Errorf("failed to bind filterer: %w", ferr)
				}
				event, perr := filterer.ParseOpen(lg)
				if perr != nil {
					fmt.Printf("❌ Failed to parse Open event: %v\n", perr)
					blockFailed = true
					continue
				}
				
				// Handle the event and track if it was settled
				settled, herr := l.handleParsedOpenEvent(*event, handler)
				if herr != nil {
					fmt.Printf("❌ Failed to handle Open event: %v\n", herr)
					blockFailed = true
					continue
				}
				
							// Track settlement status
			if settled {
				settledOrders++
			} else {
				// Log why the order wasn't settled to help debug
				fmt.Printf("   ⚠️  Order %s not settled (rules may have rejected it)\n", common.BytesToHash(event.OrderId[:]).Hex())
			}
			}
			
			if !blockFailed {
				break
			}
			retryCount++
			if retryCount >= configObj().MaxRetries {
				fmt.Printf("⏭️  Giving up on block %d after %d retries, adding to cooldown\n", b, retryCount)
				// Add block to cooldown
				l.failedMu.Lock()
				l.failedBlocks[b] = time.Now()
				l.failedMu.Unlock()
				failed = true
				break
			}
			fmt.Printf("🔁 Retry %d for block %d\n", retryCount, b)
			time.Sleep(500 * time.Millisecond)
		}
		
		if failed {
			break
		}
		
		// Only advance to this block if all orders were processed
		if settledOrders == blockOrders {
			newLast = b
			fmt.Printf("   ✅ Block %d fully processed: %d/%d orders settled\n", b, settledOrders, blockOrders)
		} else {
			fmt.Printf("   ⚠️  Block %d partially processed: %d/%d orders settled, stopping here\n", b, settledOrders, blockOrders)
			break
		}
	}
	return newLast, nil
}

// configObj fetches the loaded config (simple singleton)
var cfgSingleton *config.Config
var cfgOnce sync.Once

func configObj() *config.Config {
	cfgOnce.Do(func() {
		c, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("⚠️  Failed to load config: %v (using defaults)\n", err)
			c = &config.Config{MaxRetries: 5}
		}
		cfgSingleton = c
	})
	return cfgSingleton
}

// handleParsedOpenEvent converts a typed binding event into our internal ParsedArgs and dispatches the handler
func (l *evmListener) handleParsedOpenEvent(ev contracts.Hyperlane7683Open, handler listener.EventHandler) (bool, error) {
	// Map ResolvedCrossChainOrder
	ro := types.ResolvedCrossChainOrder{
		User:             ev.ResolvedOrder.User,
		OriginChainID:    ev.ResolvedOrder.OriginChainId,
		OpenDeadline:     ev.ResolvedOrder.OpenDeadline,
		FillDeadline:     ev.ResolvedOrder.FillDeadline,
		OrderID:          ev.ResolvedOrder.OrderId,
		MaxSpent:         make([]types.Output, 0, len(ev.ResolvedOrder.MaxSpent)),
		MinReceived:      make([]types.Output, 0, len(ev.ResolvedOrder.MinReceived)),
		FillInstructions: make([]types.FillInstruction, 0, len(ev.ResolvedOrder.FillInstructions)),
	}

	for _, o := range ev.ResolvedOrder.MaxSpent {
		// For Starknet destinations, store the original 32-byte addresses
		if l.isStarknetChain(o.ChainId) {
			fmt.Printf("   🔍 Original Starknet addresses (32 bytes):\n")
			fmt.Printf("     • Token: 0x%x\n", o.Token)
			fmt.Printf("     • Recipient: 0x%x\n", o.Recipient)
		}
		
		ro.MaxSpent = append(ro.MaxSpent, types.Output{
			Token:            bytes32ToAddress(o.Token),
			Amount:           o.Amount,
			Recipient:        bytes32ToAddress(o.Recipient),
			ChainID:          o.ChainId,
			OriginalToken:    o.Token,     // Store original 32-byte address
			OriginalRecipient: o.Recipient, // Store original 32-byte address
		})
	}
	for _, o := range ev.ResolvedOrder.MinReceived {
		ro.MinReceived = append(ro.MinReceived, types.Output{
			Token:     bytes32ToAddress(o.Token),
			Amount:    o.Amount,
			Recipient: bytes32ToAddress(o.Recipient),
			ChainID:   o.ChainId,
		})
	}
	for _, fi := range ev.ResolvedOrder.FillInstructions {
		ro.FillInstructions = append(ro.FillInstructions, types.FillInstruction{
			DestinationChainID:         fi.DestinationChainId,
			DestinationSettler:         bytes32ToAddress(fi.DestinationSettler),
			OriginData:                 fi.OriginData,
			OriginalDestinationSettler: fi.DestinationSettler, // ✅ Store original 32-byte address
		})
	}

	parsedArgs := types.ParsedArgs{
		OrderID:       common.BytesToHash(ev.OrderId[:]).Hex(),
		SenderAddress: ro.User.Hex(),
		Recipients: []types.Recipient{{
			DestinationChainName: l.config.ChainName,
			RecipientAddress:     "*",
		}},
		ResolvedOrder: ro,
	}

	fmt.Printf("📜 Open order: OrderID=%s, Chain=%s\n", parsedArgs.OrderID, l.config.ChainName)
	fmt.Printf("   📊 Order details: User=%s, OriginChainID=%s, FillDeadline=%d\n", ro.User.Hex(), ro.OriginChainID.String(), ro.FillDeadline)
	fmt.Printf("   📦 Arrays: MaxSpent=%d, MinReceived=%d, FillInstructions=%d\n", len(ro.MaxSpent), len(ro.MinReceived), len(ro.FillInstructions))

	return handler(parsedArgs, l.config.ChainName, ev.Raw.BlockNumber)
}

// bytes32ToAddress converts a left-padded bytes32 address into common.Address
func bytes32ToAddress(b [32]byte) common.Address { return common.BytesToAddress(b[12:]) }

// addressToBytes32 converts a common.Address to [32]byte (left-padded)
func addressToBytes32(addr common.Address) [32]byte {
	var result [32]byte
	copy(result[12:], addr.Bytes())
	return result
}

// chainAwareBytes32ToAddress converts bytes32 to address based on chain type
// For Starknet chains, it preserves the full 32-byte address in a special format
func chainAwareBytes32ToAddress(b [32]byte, chainID *big.Int) common.Address {
	// For Starknet chains, we need to handle the full 32-byte address differently  
	if isStarknetChainByID(chainID) {
		// For Starknet, we'll encode the full 32-byte address into the 20-byte field
		// by using a special encoding that can be decoded later
		// We'll use the first 20 bytes of the Starknet address
		var result [20]byte
		copy(result[:], b[:20])
		return common.BytesToAddress(result[:])
	}
	
	// For EVM chains, use the standard left-padded conversion
	return common.BytesToAddress(b[12:])
}

// getOriginalBytes32Address retrieves the original 32-byte address for Starknet chains
// This is a temporary workaround until we fix the type system
func getOriginalBytes32Address(encodedAddr common.Address, chainID *big.Int) [32]byte {
	if isStarknetChainByID(chainID) {
		// For Starknet, we need to reconstruct the full address
		// This is a placeholder - in the real solution, we'd store the original bytes32
		var result [32]byte
		copy(result[:20], encodedAddr.Bytes())
		// The remaining 12 bytes would come from the original event data
		// For now, we'll use zeros as a placeholder
		return result
	}
	
	// For EVM chains, convert back to bytes32 (left-padded)
	var result [32]byte
	copy(result[12:], encodedAddr.Bytes())
	return result
}

// Chain detection helper functions for listener
func (l *evmListener) isStarknetChain(chainID *big.Int) bool {
	return isStarknetChainByID(chainID)
}

// Global helper function for chain detection (used by multiple files)
func isStarknetChainByID(chainID *big.Int) bool {
	// Find any network with "Starknet" in the name that matches this chain ID
	for networkName, network := range config.Networks {
		if network.ChainID == chainID.Uint64() {
			// Check if network name contains "Starknet" (case insensitive)
			return strings.Contains(strings.ToLower(networkName), "starknet")
		}
	}
	return false
}

func (l *evmListener) catchUpHistoricalBlocks(ctx context.Context, handler listener.EventHandler) error {
	fmt.Printf("🔄 Catching up on (%s) historical blocks...\n", l.config.ChainName)
	currentBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}
	// Apply confirmations during backfill as well
	safeBlock := currentBlock
	if l.config.ConfirmationBlocks > 0 && currentBlock > l.config.ConfirmationBlocks {
		safeBlock = currentBlock - l.config.ConfirmationBlocks
	}

	// Start from the last processed block + 1 (which should be the solver start block)
	fromBlock := l.lastProcessedBlock + 1
	toBlock := safeBlock
	if fromBlock >= toBlock {
		fmt.Printf("✅ Already up to date, no historical blocks to process\n")
		return nil
	}

	chunkSize := l.config.MaxBlockRange
	for start := fromBlock; start < toBlock; start += chunkSize {
		end := start + chunkSize
		if end > toBlock {
			end = toBlock
		}
		newLast, err := l.processBlockRange(ctx, start, end, handler)
		if err != nil {
			return fmt.Errorf("failed to process historical blocks %d-%d: %v", start, end, err)
		}
		l.lastProcessedBlock = newLast
	}
	fmt.Printf("✅ Historical block processing completed for %s\n", l.config.ChainName)
	return nil
}

func (l *evmListener) startPolling(ctx context.Context, handler listener.EventHandler) {
	fmt.Printf("📭 Starting event polling...\n")
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("🔄 Context cancelled, stopping event polling\n")
			return
		case <-l.stopChan:
			fmt.Printf("🔄 Stop signal received, stopping event polling\n")
			return
		default:
			if err := l.processCurrentBlockRange(ctx, handler); err != nil {
				fmt.Printf("❌ Failed to process current block range: %v\n", err)
			}
			time.Sleep(time.Duration(l.config.PollInterval) * time.Millisecond)
		}
	}
}
