package listener

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

// Open event topic: Open(bytes32,ResolvedCrossChainOrder)
var openEventTopic = common.HexToHash("0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d")

// EVMListener implements BaseListener for EVM chains
type EVMListener struct {
	config         *ListenerConfig
	client         *ethclient.Client
	contractAddress common.Address
	logger         *logrus.Logger
	lastProcessedBlock uint64
	stopChan       chan struct{}
	mu             sync.RWMutex
}

// NewEVMListener creates a new EVM listener
func NewEVMListener(config *ListenerConfig, rpcURL string, logger *logrus.Logger) (*EVMListener, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	// Initialize lastProcessedBlock to InitialBlock - 1 to ensure we process from InitialBlock
	initialBlock := config.InitialBlock.Uint64()
	lastProcessedBlock := initialBlock - 1

	return &EVMListener{
		config:         config,
		client:         client,
		contractAddress: common.HexToAddress(config.ContractAddress),
		logger:         logger,
		lastProcessedBlock: lastProcessedBlock,
		stopChan:       make(chan struct{}),
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
	l.logger.Info("Stopping EVM listener...")
	
	// Close stop channel
	close(l.stopChan)
	
	// ethclient.Client doesn't have a Close method
	return nil
}

// GetLastProcessedBlock returns the last processed block number
func (l *EVMListener) GetLastProcessedBlock() uint64 {
	return l.lastProcessedBlock
}

// realEventLoop implements simple polling for local forks (which don't support eth_subscribe)
func (l *EVMListener) realEventLoop(ctx context.Context, handler EventHandler) {
	l.logger.Info("🚀 Starting polling-based event listener...")
	
	// Step 1: Catch up on historical blocks (MUST complete before polling starts)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		l.logger.Errorf("Failed to catch up on historical blocks: %v", err)
		// Continue anyway, we can still listen to new events
	}
	
	// Small delay to ensure blockchain state is stable after backfill
	l.logger.Info("⏳ Backfill complete, waiting for blockchain state to stabilize...")
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
	
	l.logger.Debugf("🔍 processCurrentBlockRange: currentBlock=%d, lastProcessedBlock=%d", currentBlock, l.lastProcessedBlock)
	
	// Only process new blocks since last processed
	if currentBlock <= l.lastProcessedBlock {
		l.logger.Debugf("✅ No new blocks to process: currentBlock=%d <= lastProcessedBlock=%d", currentBlock, l.lastProcessedBlock)
		return nil // Silent when no new blocks
	}
	
	// Process new blocks from last processed + 1 to current
	fromBlock := l.lastProcessedBlock + 1
	toBlock := currentBlock
	
	l.logger.Debugf("📊 Processing new block range: %d-%d", fromBlock, toBlock)
	
	// Defensive check: ensure we have a valid range
	if fromBlock > toBlock {
		l.logger.Warnf("Invalid block range: fromBlock (%d) > toBlock (%d), skipping", fromBlock, toBlock)
		return nil
	}
	
	// Process the block range
	if err := l.processBlockRange(ctx, fromBlock, toBlock, handler); err != nil {
		return fmt.Errorf("failed to process blocks %d-%d: %v", fromBlock, toBlock, err)
	}
	
	// Update the last processed block
	oldLastProcessed := l.lastProcessedBlock
	l.lastProcessedBlock = toBlock
	l.logger.Debugf("✅ Updated lastProcessedBlock: %d → %d", oldLastProcessed, l.lastProcessedBlock)
	
	// Persist the updated lastProcessedBlock to deployment state
	if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, toBlock); err != nil {
		l.logger.Warnf("⚠️  Failed to persist lastProcessedBlock: %v", err)
	} else {
		l.logger.Debugf("💾 Persisted lastProcessedBlock %d to deployment state", toBlock)
	}
	
	return nil
}

// processBlockRange processes a range of blocks for Open events only
func (l *EVMListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler EventHandler) error {
	// Defensive check: ensure we have a valid range
	if fromBlock > toBlock {
		l.logger.Warnf("Invalid block range in processBlockRange: fromBlock (%d) > toBlock (%d), skipping", fromBlock, toBlock)
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
		l.logger.Infof("🔍 Found %d Open events in blocks %d-%d", len(logs), fromBlock, toBlock)
		
		// Debug: Log each event's details
		for i, log := range logs {
			l.logger.Infof("📋 Event %d: Block=%d, TxHash=%s, Topics=%v", 
				i+1, log.BlockNumber, log.TxHash.Hex(), log.Topics)
		}
	}
	
	// Process each Open event directly
	for _, log := range logs {
		if err := l.processOpenEvent(log, handler); err != nil {
			l.logger.Errorf("Failed to process Open event: %v", err)
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
				DestinationChainName: l.config.ChainName, // We'll need to map this properly
				RecipientAddress:     "0x0000000000000000000000000000000000000000", // Placeholder
			},
		},
		ResolvedOrder: types.ResolvedOrder{
			User:             common.BytesToAddress(log.Topics[1][:20]).Hex(),
			MinReceived:      []types.TokenAmount{}, // TODO: Parse from data
			MaxSpent:         []types.TokenAmount{}, // TODO: Parse from data
			FillInstructions: log.Data, // Raw data for now
		},
	}
	
	l.logger.Infof("🎯 Open event detected: OrderID=%s, Chain=%s, Block=%d", 
		orderID.Hex(), l.config.ChainName, log.BlockNumber)
	
	// Call the handler
	return handler(parsedArgs, l.config.ChainName, log.BlockNumber)
}

// catchUpHistoricalBlocks processes all historical blocks to catch up on missed events
func (l *EVMListener) catchUpHistoricalBlocks(ctx context.Context, handler EventHandler) error {
	l.logger.Info("🔄 Catching up on historical blocks...")
	
	// Get current block
	currentBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}
	
	// Process from initial block to current block
	fromBlock := l.config.InitialBlock.Uint64()
	toBlock := currentBlock
	
	l.logger.Infof("🔧 Backfill: InitialBlock=%d, CurrentBlock=%d, LastProcessedBlock=%d", fromBlock, toBlock, l.lastProcessedBlock)
	
	if fromBlock > toBlock {
		l.logger.Info("✅ Already up to date, no historical blocks to process")
		return nil
	}
	
	// Ensure we start from the correct block (should be InitialBlock)
	if l.lastProcessedBlock != fromBlock - 1 {
		l.logger.Warnf("⚠️  lastProcessedBlock mismatch: expected %d, got %d, correcting...", fromBlock - 1, l.lastProcessedBlock)
		l.lastProcessedBlock = fromBlock - 1
	}
	
	l.logger.Infof("Processing historical blocks %d-%d", fromBlock, toBlock)
	
	// Process in chunks to avoid overwhelming the node
	chunkSize := l.config.MaxBlockRange
	l.logger.Debugf("🔧 Backfill chunk size: %d", chunkSize)
	
	for start := fromBlock; start < toBlock; start += chunkSize {
		end := start + chunkSize
		if end > toBlock {
			end = toBlock
		}
		
		l.logger.Debugf("📦 Processing backfill chunk: %d-%d", start, end)
		
		if err := l.processBlockRange(ctx, start, end, handler); err != nil {
			return fmt.Errorf("failed to process historical blocks %d-%d: %v", start, end, err)
		}
		
		l.logger.Debugf("✅ Completed backfill chunk: %d-%d", start, end)
		
		// Don't update lastProcessedBlock here - wait until the end
	}
	
	// Update last processed block only after ALL historical blocks are processed
	oldLastProcessed := l.lastProcessedBlock
	l.lastProcessedBlock = toBlock
	l.logger.Infof("🔄 Backfill complete: updated lastProcessedBlock from %d to %d", oldLastProcessed, l.lastProcessedBlock)
	
	// Persist the updated lastProcessedBlock to deployment state
	if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, toBlock); err != nil {
		l.logger.Warnf("⚠️  Failed to persist lastProcessedBlock after backfill: %v", err)
	} else {
		l.logger.Infof("💾 Persisted lastProcessedBlock %d to deployment state after backfill", toBlock)
	}
	
	l.logger.Info("✅ Historical block processing completed")
	return nil
}

// startPolling continuously polls for new Open events
func (l *EVMListener) startPolling(ctx context.Context, handler EventHandler) {
	l.logger.Info("🔄 Starting event polling...")
	
	for {
		select {
		case <-ctx.Done():
			l.logger.Info("Context cancelled, stopping event polling")
			return
		case <-l.stopChan:
			l.logger.Info("Stop signal received, stopping event polling")
			return
		default:
			// Process current block range
			if err := l.processCurrentBlockRange(ctx, handler); err != nil {
				l.logger.Errorf("Failed to process current block range: %v", err)
			}
			
			// Wait for next poll interval (5 seconds)
			time.Sleep(5 * time.Second)
		}
	}
}
