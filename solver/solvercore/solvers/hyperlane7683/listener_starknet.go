package hyperlane7683

// Module: Starknet Open event listener for Hyperlane7683
// - Polls/backfills block ranges on Starknet
// - Parses Cairo Open events and reconstructs EVM-compatible ResolvedCrossChainOrder
// - Invokes the filler with parsed args
// - Persists last processed block via deployment state

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/common"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/base"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/logutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
)

// Open event topic
var openEventSelector, _ = utils.HexToFelt("0x35D8BA7F4BF26B6E2E2060E5BD28107042BE35460FBD828C9D29A2D8AF14445")

// starknetListener implements listener.Listener for Starknet chains
type starknetListener struct {
	config             *base.ListenerConfig
	provider           *rpc.Provider
	contractAddress    *felt.Felt
	lastProcessedBlock uint64
	stopChan           chan struct{}
	mu                 sync.RWMutex
}

// NewStarknetListener creates a new Starknet listener
func NewStarknetListener(listenerConfig *base.ListenerConfig, rpcURL string) (base.Listener, error) {
	provider, err := rpc.NewProvider(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Starknet RPC: %w", err)
	}

	addrFelt, err := types.ToStarknetAddress(listenerConfig.ContractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid Starknet contract address: %w", err)
	}

	// Use the start block from config, but check if deployment state has a higher value
	var lastProcessedBlock uint64
	configStartBlock := listenerConfig.InitialBlock.Int64()

	// Handle different start block scenarios
	var resolvedStartBlock uint64
	if configStartBlock >= 0 {
		// Positive number or zero - use as-is
		resolvedStartBlock = uint64(configStartBlock)
		if configStartBlock == 0 {
			// Zero means start at current block
			ctx := context.Background()
			currentBlock, err := provider.BlockNumber(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get current block for start block 0: %w", err)
			}
			resolvedStartBlock = currentBlock
			fmt.Printf("%s📚 Start block was 0, using current block: %d\n",
				logutil.Prefix(listenerConfig.ChainName), resolvedStartBlock)
		}
	} else {
		// Negative number - start N blocks before current block
		ctx := context.Background()
		currentBlock, err := provider.BlockNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current block for negative start block: %w", err)
		}

		// Calculate start block: current - abs(configStartBlock)
		resolvedStartBlock = currentBlock - uint64(-configStartBlock)

		// Ensure we don't go below block 0
		if resolvedStartBlock > currentBlock {
			resolvedStartBlock = 0
		}

		fmt.Printf("%s📚 Start block was %d, using current block %d - %d = %d\n",
			logutil.Prefix(listenerConfig.ChainName), configStartBlock, currentBlock, -configStartBlock, resolvedStartBlock)
	}

	state, err := config.GetSolverState()
	if err != nil {
		return nil, fmt.Errorf("failed to get solver state: %w", err)
	}

	if networkState, exists := state.Networks[listenerConfig.ChainName]; exists {
		deploymentStateBlock := networkState.LastIndexedBlock

		// Use the HIGHER of the two values - this respects updated .env values
		// while also respecting any actual progress that's been saved
		if deploymentStateBlock > resolvedStartBlock {
			lastProcessedBlock = deploymentStateBlock
			fmt.Printf("%s📚 Using saved progress LastIndexedBlock: %d (config wants %d)\n",
				logutil.Prefix(listenerConfig.ChainName), lastProcessedBlock, resolvedStartBlock)
		} else {
			lastProcessedBlock = resolvedStartBlock
			fmt.Printf("%s📚 Using config SolverStartBlock: %d (saved state was %d)\n",
				logutil.Prefix(listenerConfig.ChainName), lastProcessedBlock, deploymentStateBlock)
		}
	} else {
		return nil, fmt.Errorf("network %s not found in solver state", listenerConfig.ChainName)
	}

	return &starknetListener{
		config:             listenerConfig,
		provider:           provider,
		contractAddress:    addrFelt,
		lastProcessedBlock: lastProcessedBlock,
		stopChan:           make(chan struct{}),
	}, nil
}

// Start begins listening for events
func (l *starknetListener) Start(ctx context.Context, handler base.EventHandler) (base.ShutdownFunc, error) {
	go l.startEventLoop(ctx, handler)
	return func() { close(l.stopChan) }, nil
}

// Stop gracefully stops the listener
func (l *starknetListener) Stop() error {
	close(l.stopChan)
	return nil
}

// GetLastProcessedBlock returns the last processed block number
func (l *starknetListener) GetLastProcessedBlock() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastProcessedBlock
}

// MarkBlockFullyProcessed marks a block as fully processed
func (l *starknetListener) MarkBlockFullyProcessed(blockNumber uint64) error {
	if blockNumber != l.lastProcessedBlock+1 {
		return fmt.Errorf("cannot mark block %d as processed, expected %d", blockNumber, l.lastProcessedBlock+1)
	}
	l.lastProcessedBlock = blockNumber
	fmt.Printf("%s✅ block %d processed\n", logutil.Prefix(l.config.ChainName), blockNumber)
	return nil
}

func (l *starknetListener) startEventLoop(ctx context.Context, handler base.EventHandler) {
	p := logutil.Prefix(l.config.ChainName)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		fmt.Printf("%s❌ backfill failed: %v\n", p, err)
	}
	fmt.Printf("%s🔄 backfill complete\n", p)
	l.startPolling(ctx, handler)
}

func (l *starknetListener) catchUpHistoricalBlocks(ctx context.Context, handler base.EventHandler) error {
	p := logutil.Prefix(l.config.ChainName)
	fmt.Printf("%s🔄 Catching up on historical blocks...\n", p)

	currentBlock, err := l.provider.BlockNumber(ctx)
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
		if err := config.UpdateLastIndexedBlock(l.config.ChainName, newLast); err != nil {
			fmt.Printf("%s⚠️  Failed to persist LastIndexedBlock: %v\n", p, err)
		}
	}
	fmt.Printf("%s✅ Historical block processing complete\n", p)
	return nil
}

func (l *starknetListener) startPolling(ctx context.Context, handler base.EventHandler) {
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

// getMapKeys returns the keys of a map as a slice

func (l *starknetListener) processCurrentBlockRange(ctx context.Context, handler base.EventHandler) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return ProcessCurrentBlockRangeCommon(ctx, handler, l.provider, l.config, &l.lastProcessedBlock, "Starknet", l.processBlockRange)
}

// processBlockRange processes events in [fromBlock, toBlock] and returns the highest contiguous block fully processed
func (l *starknetListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler base.EventHandler) (uint64, error) {
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range (%s) in processBlockRange: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return l.lastProcessedBlock, nil
	}

	// Fetch events for the block range
	filter := rpc.EventFilter{
		FromBlock: rpc.BlockID{Number: &fromBlock},
		ToBlock:   rpc.BlockID{Number: &toBlock},
		Address:   l.contractAddress,
		Keys:      [][]*felt.Felt{{openEventSelector}},
	}

	query := rpc.EventsInput{
		EventFilter:       filter,
		ResultPageRequest: rpc.ResultPageRequest{ChunkSize: 128, ContinuationToken: ""},
	}

	logs, err := l.provider.Events(ctx, query)
	if err != nil {
		return l.lastProcessedBlock, fmt.Errorf("failed to filter events: %w", err)
	}

	logutil.LogWithNetworkTagf(l.config.ChainName, "📩 events found: %d\n", len(logs.Events))
	if len(logs.Events) > 0 {
		fmt.Printf("📩 Found %d Open events on %s\n", len(logs.Events), l.config.ChainName)
	}

	// Group logs by block
	byBlock := make(map[uint64][]rpc.EmittedEvent)
	for _, event := range logs.Events {
		byBlock[event.BlockNumber] = append(byBlock[event.BlockNumber], event)
	}

	// Process blocks in order
	newLast := l.lastProcessedBlock
	for b := fromBlock; b <= toBlock; b++ {
		events := byBlock[b]

		// Process each event in this block
		for _, event := range events {
			// Ensure each event is the correct type
			isOpen := false
			if len(event.Event.Keys) >= 1 {
				actual := event.Event.Keys[0].Bytes()
				expected := openEventSelector.Bytes()
				if bytes.Equal(actual[:], expected[:]) {
					isOpen = true
				}
			}
			if !isOpen {
				continue
			}

			// Parse Open event
			ro := decodeResolvedOrderFromFelts(event.Event.Data)
			parsedArgs := types.ParsedArgs{
				OrderID:       common.BytesToHash(ro.OrderID[:]).Hex(),
				SenderAddress: ro.User,
				Recipients:    []types.Recipient{{DestinationChainName: l.config.ChainName, RecipientAddress: "*"}},
				ResolvedOrder: ro,
			}

			// Handle the event
			_, herr := handler(parsedArgs, l.config.ChainName, b)
			if herr != nil {
				fmt.Printf("❌ Failed to handle event: %v\n", herr)
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

// --- Decoders ---

func decodeResolvedOrderFromFelts(data []*felt.Felt) types.ResolvedCrossChainOrder {
	decoder := newFeltDecoder(data)
	
	ro := types.ResolvedCrossChainOrder{}
	ro.User = decoder.readAddress()
	ro.OriginChainID = new(big.Int).SetUint64(uint64(decoder.readU32()))
	ro.OpenDeadline = uint32(decoder.readU64())
	ro.FillDeadline = uint32(decoder.readU64())

	orderID := decoder.readU256()
	var orderArr [32]byte
	orderBytes := orderID.Bytes()
	copy(orderArr[32-len(orderBytes):], orderBytes)

	ro.OrderID = orderArr
	ro.MaxSpent = decoder.readOutputs()
	ro.MinReceived = decoder.readOutputs()
	ro.FillInstructions = decoder.readFillInstructions()
	return ro
}

// feltDecoder handles decoding of felt data
type feltDecoder struct {
	data []*felt.Felt
	idx  int
}

func newFeltDecoder(data []*felt.Felt) *feltDecoder {
	return &feltDecoder{data: data, idx: 0}
}

func (d *feltDecoder) readFelt() *felt.Felt {
	f := d.data[d.idx]
	d.idx++
	return f
}

func (d *feltDecoder) readU32() uint32 {
	bi := utils.FeltToBigInt(d.readFelt())
	return uint32(bi.Uint64())
}

func (d *feltDecoder) readU64() uint64 {
	bi := utils.FeltToBigInt(d.readFelt())
	return bi.Uint64()
}

func (d *feltDecoder) readU256() *big.Int {
	low := utils.FeltToBigInt(d.readFelt())
	high := utils.FeltToBigInt(d.readFelt())
	return new(big.Int).Add(low, new(big.Int).Lsh(high, 128))
}

func (d *feltDecoder) readAddress() string {
	feltBytes := d.readFelt().Bytes()
	// Convert to slice to handle consistently
	b := feltBytes[:]
	// Pad to 32 bytes if shorter (for consistent bytes32 format)
	if len(b) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(b):], b)
		return "0x" + hex.EncodeToString(padded)
	}
	return "0x" + hex.EncodeToString(b)
}

func (d *feltDecoder) readOutput() types.Output {
	out := types.Output{}
	out.Token = d.readAddress()
	out.Amount = d.readU256()
	out.Recipient = d.readAddress()
	chainDomain := d.readU32()
	// Map domain to actual chain ID using config
	if chainID, err := domainToChainID(chainDomain); err == nil {
		out.ChainID = chainID
	} else {
		fmt.Printf("   ⚠️  Warning: Could not map domain %d to chain ID for output, using domain as chain ID\n", chainDomain)
		out.ChainID = new(big.Int).SetUint64(uint64(chainDomain))
	}
	return out
}

func (d *feltDecoder) readOutputs() []types.Output {
	length := utils.FeltToBigInt(d.readFelt()).Uint64()
	outs := make([]types.Output, 0, length)
	for i := uint64(0); i < length; i++ {
		outs = append(outs, d.readOutput())
	}
	return outs
}

func (d *feltDecoder) readFillInstruction() types.FillInstruction {
	fi := types.FillInstruction{}
	destinationDomain := d.readU32()
	// Map destination domain to actual chain ID using config
	if chainID, err := domainToChainID(destinationDomain); err == nil {
		fi.DestinationChainID = chainID
	} else {
		fmt.Printf("   ⚠️  Warning: Could not map domain %d to chain ID, using domain as chain ID\n", destinationDomain)
		fi.DestinationChainID = new(big.Int).SetUint64(uint64(destinationDomain))
	}
	fi.DestinationSettler = d.readAddress()

	// Parse the origin_data bytes (OrderData struct) from the event data
	fi.OriginData = d.parseOriginData()
	return fi
}

func (d *feltDecoder) parseOriginData() []byte {
	// Read size and u128 array length from the event data (absolute indices)
	size := utils.FeltToBigInt(d.data[21]).Uint64()
	u128ArrayLength := utils.FeltToBigInt(d.data[22]).Uint64()
	_ = size
	_ = u128ArrayLength

	// Parse each bytes32 field from the u128 array
	orderDataFields := make([][]byte, 0)
	for i := uint64(0); i < u128ArrayLength && (23+int(i)+1) < len(d.data); i += 2 {
		// Read two u128 felts and combine into bytes32
		lowFelt := d.data[23+int(i)]
		highFelt := d.data[23+int(i)+1]
		lowBytes := lowFelt.Bytes()
		highBytes := highFelt.Bytes()
		lowU128 := lowBytes[16:]
		highU128 := highBytes[16:]
		bytes32 := make([]byte, 32)
		copy(bytes32[0:16], lowU128)
		copy(bytes32[16:32], highU128)
		orderDataFields = append(orderDataFields, bytes32)
	}

	// Build EVM origin_data bytes (ABI-compatible, 448 bytes total)
	evmOriginData := make([]byte, 0, evmOriginDataSize)
	firstWord := make([]byte, 32)
	firstWord[31] = 0x20
	evmOriginData = append(evmOriginData, firstWord...)
	evmOriginData = append(evmOriginData, orderDataFields[1]...)  // sender
	evmOriginData = append(evmOriginData, orderDataFields[2]...)  // recipient
	evmOriginData = append(evmOriginData, orderDataFields[3]...)  // input_token
	evmOriginData = append(evmOriginData, orderDataFields[4]...)  // output_token
	evmOriginData = append(evmOriginData, orderDataFields[5]...)  // amount_in
	evmOriginData = append(evmOriginData, orderDataFields[6]...)  // amount_out
	evmOriginData = append(evmOriginData, orderDataFields[7]...)  // sender_nonce
	evmOriginData = append(evmOriginData, orderDataFields[8]...)  // origin_domain
	evmOriginData = append(evmOriginData, orderDataFields[9]...)  // destination_domain
	evmOriginData = append(evmOriginData, orderDataFields[10]...) // destination_settler
	evmOriginData = append(evmOriginData, orderDataFields[11]...) // fill_deadline
	dataOffset := make([]byte, 32)
	dataOffset[31] = 0x80
	dataOffset[30] = 0x01
	evmOriginData = append(evmOriginData, dataOffset...)
	dataSize := make([]byte, 32)
	dataSize[31] = 0x00
	evmOriginData = append(evmOriginData, dataSize...)
	if len(evmOriginData) != evmOriginDataSize {
		fmt.Printf("   ⚠️  origin_data unexpected length: %d\n", len(evmOriginData))
	}
	return evmOriginData
}

func (d *feltDecoder) readFillInstructions() []types.FillInstruction {
	length := utils.FeltToBigInt(d.readFelt()).Uint64()
	arr := make([]types.FillInstruction, 0, length)
	for i := uint64(0); i < length; i++ {
		arr = append(arr, d.readFillInstruction())
	}
	return arr
}

// domainToChainID maps a Hyperlane domain ID to its corresponding chain ID
func domainToChainID(domain uint32) (*big.Int, error) {
	// Search through all networks to find the one with matching HyperlaneDomain
	for _, network := range config.Networks {
		if network.HyperlaneDomain == uint64(domain) {
			return big.NewInt(int64(network.ChainID)), nil
		}
	}
	return nil, fmt.Errorf("no chain found for domain %d", domain)
}
