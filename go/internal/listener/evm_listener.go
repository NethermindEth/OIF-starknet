package listener

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

)

// Open event topic: Open(bytes32,ResolvedCrossChainOrder)
var openEventTopic = common.HexToHash("0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d")

// EVMListener implements BaseListener for EVM chains
type EVMListener struct {
	config             *ListenerConfig
	client             *ethclient.Client
	contractAddress    common.Address
	logger             interface{}
	lastProcessedBlock uint64
	stopChan           chan struct{}
	mu                 sync.RWMutex
}

// NewEVMListener creates a new EVM listener
func NewEVMListener(config *ListenerConfig, rpcURL string, logger interface{}) (*EVMListener, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	// Initialize lastProcessedBlock safely, handling nil/zero InitialBlock
	var lastProcessedBlock uint64
	if config.InitialBlock == nil || config.InitialBlock.Sign() <= 0 {
		// If no initial block specified, start from current block
		currentBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get current block number: %w", err)
		}
		lastProcessedBlock = currentBlock
	} else {
		// Start from specified initial block - 1 to ensure we process from InitialBlock
		lastProcessedBlock = config.InitialBlock.Uint64() - 1
	}

	return &EVMListener{
		config:             config,
		client:             client,
		contractAddress:    common.HexToAddress(config.ContractAddress),
		logger:             logger,
		lastProcessedBlock: lastProcessedBlock,
		stopChan:           make(chan struct{}),
	}, nil
}

// Start begins listening for events
func (l *EVMListener) Start(ctx context.Context, handler EventHandler) (ShutdownFunc, error) {
	// Start real event listening
	go l.realEventLoop(ctx, handler)

	// Return shutdown function
	return func() {
		close(l.stopChan)
	}, nil
}

// Stop gracefully stops the listener
func (l *EVMListener) Stop() error {
	fmt.Printf("Stopping EVM listener...\n")

	// Close stop channel
	close(l.stopChan)

	// ethclient.Client doesn't have a Close method
	return nil
}

// GetLastProcessedBlock returns the last processed block number
func (l *EVMListener) GetLastProcessedBlock() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastProcessedBlock
}

// MarkBlockFullyProcessed marks a block as fully processed and updates LastIndexedBlock
// This should be called after all events in a block have been processed/filled
func (l *EVMListener) MarkBlockFullyProcessed(blockNumber uint64) error {
	// Only update if this is the next block in sequence
	if blockNumber != l.lastProcessedBlock+1 {
		return fmt.Errorf("cannot mark block %d as processed, expected %d", blockNumber, l.lastProcessedBlock+1)
	}
	
	// Update the last processed block
	l.lastProcessedBlock = blockNumber
	
	// TODO: This method should be called by the solver manager after all events in a block are processed
	// The solver manager will handle updating LastIndexedBlock via deployer.UpdateLastIndexedBlock
	// This ensures proper coordination between event processing and block indexing
	
	fmt.Printf("✅ Block %d marked as fully processed for %s\n", blockNumber, l.config.ChainName)
	return nil
}

// realEventLoop implements simple polling for local forks (which don't support eth_subscribe)
func (l *EVMListener) realEventLoop(ctx context.Context, handler EventHandler) {
	fmt.Printf("⚙️  Starting (%s) event listener...\n", l.config.ChainName)

	// Step 1: Catch up on historical blocks (MUST complete before polling starts)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		fmt.Printf("❌ Failed to catch up on (%s) historical blocks: %v\n", l.config.ChainName, err)
		// Continue anyway, we can still listen to new events
	}

	// Small delay to ensure blockchain state is stable after backfill
	fmt.Printf("🔄 Backfill complete (%s)\n", l.config.ChainName)
	time.Sleep(1 * time.Second)

	// Step 2: Start polling for new events (only after backfill is complete)
	l.startPolling(ctx, handler)
}

// processCurrentBlockRange processes the current block range for events
func (l *EVMListener) processCurrentBlockRange(ctx context.Context, handler EventHandler) error {
	// Get current block
	currentBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}

	// Only process new blocks since last processed
	if currentBlock <= l.lastProcessedBlock {
		return nil // Silent when no new blocks
	}

	// Process new blocks from last processed + 1 to current
	fromBlock := l.lastProcessedBlock + 1
	toBlock := currentBlock

	// Defensive check: ensure we have a valid range
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range for %s: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return nil
	}

	// Process the block range
	if err := l.processBlockRange(ctx, fromBlock, toBlock, handler); err != nil {
		return fmt.Errorf("failed to process blocks %d-%d: %v", fromBlock, toBlock, err)
	}

	// Update the last processed block (but don't persist to deployment state yet)
	// TODO: We only persist after all events in the block have been fully processed/filled
	// This means:
	// 1. Process all events in the block
	// 2. For each event: check if already filled, if not then fill it
	// 3. Only after ALL events in the block are processed/filled, update LastIndexedBlock
	// 4. This ensures we never skip a block with unprocessed events
	l.lastProcessedBlock = toBlock

	return nil
}

// processBlockRange processes a range of blocks for Open events only
func (l *EVMListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler EventHandler) error {
	// Defensive check: ensure we have a valid range
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range (%s) in processBlockRange: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return nil
	}

	// Query for Open events only
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{l.contractAddress},
		Topics: [][]common.Hash{
			{openEventTopic}, // Only Open events
		},
	}

	logs, err := l.client.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to filter logs: %v", err)
	}

	// Only log if we found events
	if len(logs) > 0 {
		fmt.Printf("📩 Found %d Open events on %s\n", len(logs), l.config.ChainName)

		// // Debug: Log each event's details
		// for i, log := range logs {
		// 	l.logger.Infof("📋 Event %d: Block=%d, TxHash=%s, Topics=%v",
		// 		i+1, log.BlockNumber, log.TxHash.Hex(), log.Topics)
		// }
	}

	// Process each Open event directly
	for _, log := range logs {
		if err := l.processOpenEvent(log, handler); err != nil {
			fmt.Printf("❌ Failed to process Open event: %v\n", err)
			continue
		}
	}

	return nil
}

// processOpenEvent processes a single Open event
func (l *EVMListener) processOpenEvent(log ethtypes.Log, handler EventHandler) error {
	// Parse the Open event
	// Event structure: Open(bytes32 indexed orderId, ResolvedCrossChainOrder resolvedOrder)

	// Extract orderId from indexed topic
	if len(log.Topics) < 2 {
		return fmt.Errorf("invalid Open event: missing orderId topic")
	}
	orderID := log.Topics[1] // orderId is the first indexed parameter

	// Parse the resolvedOrder from the data field
	// This is complex as it contains nested structs, so we'll extract basic info for now
	parsedArgs := types.ParsedArgs{
		OrderID:       orderID.Hex(),
		SenderAddress: common.BytesToAddress(log.Topics[1][:20]).Hex(), // Use orderId as sender for now
		Recipients: []types.Recipient{
			{
				DestinationChainName: l.config.ChainName,                           // We'll need to map this properly
				RecipientAddress:     "0x0000000000000000000000000000000000000000", // Placeholder
			},
		},
		ResolvedOrder: types.ResolvedOrder{
			User:             common.BytesToAddress(log.Topics[1][:20]).Hex(),
			MinReceived:      []types.TokenAmount{}, // TODO: Parse from data
			MaxSpent:         []types.TokenAmount{}, // TODO: Parse from data
			FillInstructions: log.Data,              // Raw data for now
		},
	}

	fmt.Printf("📜 Open order: OrderID=%s, Chain=%s\n",
		orderID.Hex(), l.config.ChainName)

	// Call the handler
	return handler(parsedArgs, l.config.ChainName, log.BlockNumber)
}

// catchUpHistoricalBlocks processes all historical blocks to catch up on missed events
func (l *EVMListener) catchUpHistoricalBlocks(ctx context.Context, handler EventHandler) error {
	fmt.Printf("🔄 Catching up on (%s) historical blocks...\n", l.config.ChainName)

	// Get current block
	currentBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}

	// Process from initial block to current block, handling nil/zero InitialBlock
	var fromBlock uint64
	if l.config.InitialBlock == nil || l.config.InitialBlock.Sign() <= 0 {
		// If no initial block specified, start from current block (no historical processing needed)
		fromBlock = currentBlock
	} else {
		fromBlock = l.config.InitialBlock.Uint64()
	}
	toBlock := currentBlock

	if fromBlock >= toBlock {
		fmt.Printf("✅ Already up to date, no historical blocks to process\n")
		return nil
	}

	// Ensure we start from the correct block (should be InitialBlock)
	if l.lastProcessedBlock != fromBlock-1 {
		fmt.Printf("⚠️  lastProcessedBlock mismatch: expected %d, got %d, correcting...\n", fromBlock-1, l.lastProcessedBlock)
		l.lastProcessedBlock = fromBlock - 1
	}

	// Process in chunks to avoid overwhelming the node
	chunkSize := l.config.MaxBlockRange

	for start := fromBlock; start < toBlock; start += chunkSize {
		end := start + chunkSize
		if end > toBlock {
			end = toBlock
		}

		if err := l.processBlockRange(ctx, start, end, handler); err != nil {
			return fmt.Errorf("failed to process historical blocks %d-%d: %v", start, end, err)
		}
	}

	// Update last processed block only after ALL historical blocks are processed
	l.lastProcessedBlock = toBlock

	fmt.Printf("✅ Historical block processing completed for %s\n", l.config.ChainName)
	return nil
}

// startPolling continuously polls for new Open events
func (l *EVMListener) startPolling(ctx context.Context, handler EventHandler) {
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
			// Process current block range
			if err := l.processCurrentBlockRange(ctx, handler); err != nil {
				fmt.Printf("❌ Failed to process current block range: %v\n", err)
			}

			// Wait for next poll interval using configured value
			time.Sleep(time.Duration(l.config.PollInterval) * time.Millisecond)
		}
	}
}
