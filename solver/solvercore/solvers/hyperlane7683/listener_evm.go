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
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/base"
	contracts "github.com/NethermindEth/oif-starknet/solver/solvercore/contracts"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/logutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Open event topic
var openEventTopic = common.HexToHash("0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d")

// evmListener implements listener.Listener for EVM chains
type evmListener struct {
	config             *base.ListenerConfig
	client             *ethclient.Client
	contractAddress    common.Address
	lastProcessedBlock uint64
	stopChan           chan struct{}
	mu                 sync.RWMutex
	baseListener       *BaseListener
}

func NewEVMListener(listenerConfig *base.ListenerConfig, rpcURL string) (base.Listener, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	address, err := types.ToEVMAddress(listenerConfig.ContractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid EVM contract address: %w", err)
	}

	ctx := context.Background()
	commonConfig, err := ResolveCommonListenerConfig(ctx, listenerConfig, client)
	if err != nil {
		return nil, err
	}

	baseListener := NewBaseListener(*listenerConfig, client, "EVM")
	baseListener.SetLastProcessedBlock(commonConfig.LastProcessedBlock)
	
	return &evmListener{
		config:             listenerConfig,
		client:             client,
		contractAddress:    address,
		lastProcessedBlock: commonConfig.LastProcessedBlock,
		stopChan:           make(chan struct{}),
		mu:                 sync.RWMutex{},
		baseListener:       baseListener,
	}, nil
}

// Start begins listening for events
func (l *evmListener) Start(ctx context.Context, handler base.EventHandler) (base.ShutdownFunc, error) {
	go l.startEventLoop(ctx, handler)
	return func() { close(l.stopChan) }, nil
}

// Stop gracefully stops the listener
func (l *evmListener) Stop() error {
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

func (l *evmListener) startEventLoop(ctx context.Context, handler base.EventHandler) {
	p := logutil.Prefix(l.config.ChainName)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		fmt.Printf("%s❌ backfill failed: %v\n", p, err)
	}
	fmt.Printf("%s🔄 backfill complete\n", p)
	l.startPolling(ctx, handler)
}

func (l *evmListener) catchUpHistoricalBlocks(ctx context.Context, handler base.EventHandler) error {
	return l.baseListener.CatchUpHistoricalBlocks(ctx, handler, l.processBlockRange)
}

func (l *evmListener) startPolling(ctx context.Context, handler base.EventHandler) {
	fmt.Printf("%s📭 Starting event polling...\n", logutil.Prefix(l.config.ChainName))

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("🔄 Context canceled, stopping event polling\n")
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

func (l *evmListener) processCurrentBlockRange(ctx context.Context, handler base.EventHandler) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return ProcessCurrentBlockRangeCommon(ctx, handler, l.client, l.config, &l.lastProcessedBlock, "EVM", l.processBlockRange)
}

// processBlockRange processes logs in [fromBlock, toBlock] and returns the highest contiguous block fully processed
func (l *evmListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler base.EventHandler) (uint64, error) {
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range (%s) in processBlockRange: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return l.lastProcessedBlock, nil
	}

	// Fetch events for the block range
	query := ethereum.FilterQuery{
		FromBlock:  big.NewInt(int64(fromBlock)),
		ToBlock:    big.NewInt(int64(toBlock)),
		Addresses:  []common.Address{l.contractAddress},
		Topics:     [][]common.Hash{{openEventTopic}},
		BlockHash:  nil,
	}

	logs, err := l.client.FilterLogs(ctx, query)
	if err != nil {
		return l.lastProcessedBlock, fmt.Errorf("failed to filter logs: %w", err)
	}

	// Use the new logging system for reduced verbosity
	logutil.LogBlockProcessing(l.config.ChainName, fromBlock, toBlock, len(logs))

	// Group logs by block
	byBlock := make(map[uint64][]ethtypes.Log)
	for i := range logs {
		event := &logs[i]
		byBlock[event.BlockNumber] = append(byBlock[event.BlockNumber], *event)
	}

	// Process blocks in order
	newLast := l.lastProcessedBlock
	for b := fromBlock; b <= toBlock; b++ {
		events := byBlock[b]

		// Process each event in this block
		for i := range events {
			logEvent := &events[i]
			// Use generated binding to parse Open events
			filterer, err := contracts.NewHyperlane7683Filterer(l.contractAddress, l.client)
			if err != nil {
				fmt.Printf("❌ Failed to bind filterer: %v\n", err)
				continue
			}

			// Parse Open event
			event, err := filterer.ParseOpen(*logEvent)
			if err != nil {
				fmt.Printf("❌ Failed to parse Open event: %v\n", err)
				continue
			}

			// Handle the event
			_, err = l.handleParsedOpenEvent(event, handler)
			if err != nil {
				fmt.Printf("❌ Failed to handle Open event: %v\n", err)
				continue
			}
		}

		// Mark block as processed
		newLast = b

		// Only log individual blocks if there are events
		if len(events) > 0 {
			logutil.LogWithNetworkTagf(l.config.ChainName, "   ✅ Block %d processed: %d events\n", b, len(events))
		}
	}

	return newLast, nil
}

// handleParsedOpenEvent converts a typed binding event into our internal ParsedArgs and dispatches the handler
func (l *evmListener) handleParsedOpenEvent(ev *contracts.Hyperlane7683Open, handler base.EventHandler) (bool, error) {
	p := logutil.Prefix(l.config.ChainName)

	// Parse to ResolvedCrossChainOrder
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
