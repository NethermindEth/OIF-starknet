package hyperlane7683

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	contracts "github.com/NethermindEth/oif-starknet/go/internal/contracts"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/listener"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
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

func (l *evmListener) processCurrentBlockRange(ctx context.Context, handler listener.EventHandler) error {
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
	if err := l.processBlockRange(ctx, fromBlock, toBlock, handler); err != nil {
		return fmt.Errorf("failed to process blocks %d-%d: %v", fromBlock, toBlock, err)
	}
	l.lastProcessedBlock = toBlock
	if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, toBlock); err != nil {
		fmt.Printf("⚠️  Failed to persist LastIndexedBlock for %s: %v\n", l.config.ChainName, err)
	} else {
		fmt.Printf("💾 Persisted LastIndexedBlock=%d for %s\n", toBlock, l.config.ChainName)
	}
	return nil
}

func (l *evmListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler listener.EventHandler) error {
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range (%s) in processBlockRange: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return nil
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
		return fmt.Errorf("failed to filter logs: %v", err)
	}
	fmt.Printf("📩 %s logs found: %d\n", l.config.ChainName, len(logs))
	if len(logs) > 0 {
		fmt.Printf("📩 Found %d Open events on %s\n", len(logs), l.config.ChainName)
	}

	// Use generated binding to parse Open events
	filterer, err := contracts.NewHyperlane7683Filterer(l.contractAddress, l.client)
	if err != nil {
		return fmt.Errorf("failed to bind filterer: %w", err)
	}

	for _, lg := range logs {
		event, err := filterer.ParseOpen(lg)
		if err != nil {
			fmt.Printf("❌ Failed to parse Open event: %v\n", err)
			continue
		}
		if err := l.handleParsedOpenEvent(*event, handler); err != nil {
			fmt.Printf("❌ Failed to handle Open event: %v\n", err)
			continue
		}
	}
	return nil
}

// handleParsedOpenEvent converts a typed binding event into our internal ParsedArgs and dispatches the handler
func (l *evmListener) handleParsedOpenEvent(ev contracts.Hyperlane7683Open, handler listener.EventHandler) error {
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
		ro.MaxSpent = append(ro.MaxSpent, types.Output{
			Token:     bytes32ToAddress(o.Token),
			Amount:    o.Amount,
			Recipient: bytes32ToAddress(o.Recipient),
			ChainID:   o.ChainId,
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
			DestinationChainID: fi.DestinationChainId,
			DestinationSettler: bytes32ToAddress(fi.DestinationSettler),
			OriginData:         fi.OriginData,
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
		if err := l.processBlockRange(ctx, start, end, handler); err != nil {
			return fmt.Errorf("failed to process historical blocks %d-%d: %v", start, end, err)
		}
	}
	l.lastProcessedBlock = toBlock
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
