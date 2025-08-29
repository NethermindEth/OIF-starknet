package hyperlane7683

// Module: EVM Open event listener for Hyperlane7683
// - Polls/backfills block ranges on EVM networks
// - Parses Hyperlane7683 Open events via abigen bindings
// - Translates to types.ParsedArgs and invokes the solver
// - Persists last processed block via deployment state

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/base"
	"github.com/NethermindEth/oif-starknet/go/internal/config"
	contracts "github.com/NethermindEth/oif-starknet/go/internal/contracts"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/logutil"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Open event topic: Open(bytes32,ResolvedCrossChainOrder)
var openEventTopic = common.HexToHash("0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d")

// evmListener implements listener.BaseListener for EVM chains for Hyperlane7683
type evmListener struct {
	config             *base.ListenerConfig
	client             *ethclient.Client
	contractAddress    common.Address
	lastProcessedBlock uint64
	stopChan           chan struct{}
	mu                 sync.RWMutex
}

func NewEVMListener(config *base.ListenerConfig, rpcURL string) (base.BaseListener, error) {
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
		fmt.Printf("%s📚 Using LastIndexedBlock: %d\n", logutil.Prefix(config.ChainName), lastProcessedBlock)
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
func (l *evmListener) Start(ctx context.Context, handler base.EventHandler) (base.ShutdownFunc, error) {
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
	fmt.Printf("%s✅ block %d processed\n", logutil.Prefix(l.config.ChainName), blockNumber)
	return nil
}

func (l *evmListener) realEventLoop(ctx context.Context, handler base.EventHandler) {
	p := logutil.Prefix(l.config.ChainName)
	//fmt.Printf("%s⚙️  starting listener...\n", p)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		fmt.Printf("%s❌ backfill failed: %v\n", p, err)
	}
	fmt.Printf("%s🔄 backfill complete\n", p)
	time.Sleep(1 * time.Second)
	l.startPolling(ctx, handler)
}

func (l *evmListener) processCurrentBlockRange(ctx context.Context, handler base.EventHandler) error {

	currentBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %v", err)
	}
	// Apply confirmations window if configured
	safeBlock := currentBlock
	if l.config.ConfirmationBlocks > 0 && currentBlock > l.config.ConfirmationBlocks {
		safeBlock = currentBlock - l.config.ConfirmationBlocks
	}
	fromBlock := l.lastProcessedBlock + 1
	toBlock := safeBlock

	// Check if we have any new blocks to process
	if fromBlock > toBlock {
		// No new blocks to process, we're up to date
		// Only log this occasionally to avoid spam
		// if time.Now().Unix()%30 == 0 { // Log every 30 seconds max
		// 	fmt.Printf("🧭 %s EVM range: from=%d to=%d (current=%d, conf=%d) - ✅ Already up to date\n", l.config.ChainName, fromBlock, toBlock, currentBlock, l.config.ConfirmationBlocks)
		// }
		return nil
	}

	fmt.Printf("🧭 %s EVM range: from=%d to=%d (current=%d, conf=%d)\n", l.config.ChainName, fromBlock, toBlock, currentBlock, l.config.ConfirmationBlocks)
	newLast, err := l.processBlockRange(ctx, fromBlock, toBlock, handler)
	if err != nil {
		return fmt.Errorf("failed to process blocks %d-%d: %v", fromBlock, toBlock, err)
	}

	// Block processing complete

	l.lastProcessedBlock = newLast
	if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, newLast); err != nil {
		fmt.Printf("⚠️  Failed to persist LastIndexedBlock for %s: %v\n", l.config.ChainName, err)
	} else {
		fmt.Printf("💾 Persisted LastIndexedBlock=%d for %s\n", newLast, l.config.ChainName)
	}
	return nil
}

// processBlockRange processes logs in [fromBlock, toBlock] and returns the highest contiguous block fully processed
func (l *evmListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler base.EventHandler) (uint64, error) {
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

	// Group logs by block
	byBlock := make(map[uint64][]ethtypes.Log)
	for _, lg := range logs {
		byBlock[lg.BlockNumber] = append(byBlock[lg.BlockNumber], lg)
	}

	// Process blocks in order
	newLast := l.lastProcessedBlock
	for b := fromBlock; b <= toBlock; b++ {
		events := byBlock[b]

		// Process each event in this block
		for _, lg := range events {
			// Use generated binding to parse Open events
			filterer, err := contracts.NewHyperlane7683Filterer(l.contractAddress, l.client)
			if err != nil {
				fmt.Printf("❌ Failed to bind filterer: %v\n", err)
				continue
			}

			event, err := filterer.ParseOpen(lg)
			if err != nil {
				fmt.Printf("❌ Failed to parse Open event: %v\n", err)
				continue
			}

			// Handle the event
			_, err = l.handleParsedOpenEvent(*event, handler)
			if err != nil {
				fmt.Printf("❌ Failed to handle Open event: %v\n", err)
				continue
			}
		}

		// Mark block as processed
		newLast = b
		fmt.Printf("   ✅ Block %d processed: %d events\n", b, len(events))
	}

	return newLast, nil
}

// handleParsedOpenEvent converts a typed binding event into our internal ParsedArgs and dispatches the handler
func (l *evmListener) handleParsedOpenEvent(ev contracts.Hyperlane7683Open, handler base.EventHandler) (bool, error) {
	p := logutil.Prefix(l.config.ChainName)

	// Map ResolvedCrossChainOrder
	ro := types.ResolvedCrossChainOrder{
		User:             ev.ResolvedOrder.User.Hex(),
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
			Token:     bytes32ToHexString(o.Token),
			Amount:    o.Amount,
			Recipient: bytes32ToHexString(o.Recipient),
			ChainID:   o.ChainId,
		})
	}
	for _, o := range ev.ResolvedOrder.MinReceived {
		ro.MinReceived = append(ro.MinReceived, types.Output{
			Token:     bytes32ToHexString(o.Token),
			Amount:    o.Amount,
			Recipient: bytes32ToHexString(o.Recipient),
			ChainID:   o.ChainId,
		})
	}
	for _, fi := range ev.ResolvedOrder.FillInstructions {
		ro.FillInstructions = append(ro.FillInstructions, types.FillInstruction{
			DestinationChainID: fi.DestinationChainId,
			DestinationSettler: bytes32ToHexString(fi.DestinationSettler),
			OriginData:         fi.OriginData,
		})
	}

	parsedArgs := types.ParsedArgs{
		OrderID:       common.BytesToHash(ev.OrderId[:]).Hex(),
		SenderAddress: ro.User,
		Recipients: []types.Recipient{{
			DestinationChainName: l.config.ChainName,
			RecipientAddress:     "*",
		}},
		ResolvedOrder: ro,
	}

	fmt.Printf("%s📜 Open order: OrderID=%s\n", p, parsedArgs.OrderID)
	fmt.Printf("%s📊 Order details: User=%s\n", p, ro.User)

	// Just pass to handler, let the solver decide what to do
	return handler(parsedArgs, l.config.ChainName, ev.Raw.BlockNumber)
}

// bytes32ToHexString converts a bytes32 address to a hex string
func bytes32ToHexString(b [32]byte) string {
	return "0x" + hex.EncodeToString(b[:])
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

func (l *evmListener) catchUpHistoricalBlocks(ctx context.Context, handler base.EventHandler) error {
	p := logutil.Prefix(l.config.ChainName)
	fmt.Printf("%s🔄 Catching up on historical blocks...\n", p)
	currentBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("%sfailed to get current block number: %v", p, err)
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
		fmt.Printf("%s✅ Already up to date, no historical blocks to process\n", p)
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
			return fmt.Errorf("%sfailed to process historical blocks %d-%d: %v", p, start, end, err)
		}
		l.lastProcessedBlock = newLast
	}
	fmt.Printf("%s✅ Historical block processing complete\n", p)
	return nil
}

func (l *evmListener) startPolling(ctx context.Context, handler base.EventHandler) {
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
				fmt.Printf("%s❌ Failed to process current block range: %v\n", logutil.Prefix(l.config.ChainName), err)
			}
			time.Sleep(time.Duration(l.config.PollInterval) * time.Millisecond)
		}
	}
}
