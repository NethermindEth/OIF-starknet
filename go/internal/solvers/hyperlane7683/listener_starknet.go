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

	"github.com/NethermindEth/oif-starknet/go/internal/base"
	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/logutil"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
)

// starknetListener implements listener.BaseListener for Starknet chains
type starknetListener struct {
	config             *base.ListenerConfig
	provider           *rpc.Provider
	contractAddress    *felt.Felt
	openEventSelector  *felt.Felt
	lastProcessedBlock uint64
	stopChan           chan struct{}
	mu                 sync.RWMutex
}

// NewStarknetListener creates a new Starknet listener
func NewStarknetListener(config *base.ListenerConfig, rpcURL string) (base.BaseListener, error) {
	provider, err := rpc.NewProvider(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Starknet RPC: %w", err)
	}

	addrFelt, err := utils.HexToFelt(config.ContractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid Starknet contract address: %w", err)
	}

	// Open event selector for Cairo event "Open"
	openSelector, err := utils.HexToFelt("0x35D8BA7F4BF26B6E2E2060E5BD28107042BE35460FBD828C9D29A2D8AF14445")
	if err != nil {
		return nil, fmt.Errorf("invalid Open event selector: %w", err)
	}

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

	return &starknetListener{config: config, provider: provider, contractAddress: addrFelt, openEventSelector: openSelector, lastProcessedBlock: lastProcessedBlock, stopChan: make(chan struct{})}, nil
}

// Start begins listening for events
func (l *starknetListener) Start(ctx context.Context, handler base.EventHandler) (base.ShutdownFunc, error) {
	go l.realEventLoop(ctx, handler)
	return func() { close(l.stopChan) }, nil
}

// Stop gracefully stops the listener
func (l *starknetListener) Stop() error {
	fmt.Printf("Stopping Starknet listener...\n")
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

func (l *starknetListener) realEventLoop(ctx context.Context, handler base.EventHandler) {
	p := logutil.Prefix(l.config.ChainName)
	//fmt.Printf("%s⚙️  starting listener...\n", p)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		fmt.Printf("%s❌ backfill failed: %v\n", p, err)
	}
	fmt.Printf("%s🔄 backfill complete\n", p)
	time.Sleep(1 * time.Second)
	l.startPolling(ctx, handler)
}

func (l *starknetListener) catchUpHistoricalBlocks(ctx context.Context, handler base.EventHandler) error {
	p := logutil.Prefix(l.config.ChainName)
	fmt.Printf("%s🔄 Catching up on historical blocks...\n", p)
	currentBlock, err := l.provider.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("%sfailed to get current block number%v", p, err)
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
		if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, newLast); err != nil {
			fmt.Printf("%s⚠️  Failed to persist LastIndexedBlock: %v\n", p, err)
		} else {
			fmt.Printf("%s💾 Persisted LastIndexedBlock=%d\n", p, newLast)
		}
	}
	fmt.Printf("%s✅ Historical block processing completed\n", p)
	return nil
}

func (l *starknetListener) startPolling(ctx context.Context, handler base.EventHandler) {
	fmt.Printf("📭 Starting event polling...\n")
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("📭 Context cancelled, stopping polling for %s\n", l.config.ChainName)
			return
		case <-l.stopChan:
			fmt.Printf("📭 Stop signal received, stopping polling for %s\n", l.config.ChainName)
			return
		default:
			if err := l.processCurrentBlockRange(ctx, handler); err != nil {
				fmt.Printf("❌ Failed to process current block range: %v\n", err)
			}
			time.Sleep(time.Duration(l.config.PollInterval) * time.Millisecond)
		}
	}
}

func (l *starknetListener) processCurrentBlockRange(ctx context.Context, handler base.EventHandler) error {
	currentBlock, err := l.provider.BlockNumber(ctx)
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
		// 	fmt.Printf("🧭 %s Starknet range: from=%d to=%d (current=%d, conf=%d) - ✅ Already up to date\n", l.config.ChainName, fromBlock, toBlock, currentBlock, l.config.ConfirmationBlocks)
		// }
		return nil
	}
	newLast, err := l.processBlockRange(ctx, fromBlock, toBlock, handler)
	if err != nil {
		return fmt.Errorf("failed to process blocks %d-%d: %v", fromBlock, toBlock, err)
	}
	l.lastProcessedBlock = newLast
	if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, newLast); err != nil {
		fmt.Printf("⚠️  Failed to persist LastIndexedBlock for %s: %v\n", l.config.ChainName, err)
	} else {
		fmt.Printf("💾 Persisted LastIndexedBlock=%d for %s\n", newLast, l.config.ChainName)
	}
	return nil
}

// processBlockRange processes events in [fromBlock, toBlock] and returns the highest contiguous block fully processed
func (l *starknetListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler base.EventHandler) (uint64, error) {
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range (%s) in processBlockRange: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
		return l.lastProcessedBlock, nil
	}

	pageSize := 128
	cursor := ""
	newLast := l.lastProcessedBlock

	retryCount := 0
	for {
		fb := fromBlock
		tb := toBlock
		filter := rpc.EventFilter{
			FromBlock: rpc.BlockID{Number: &fb},
			ToBlock:   rpc.BlockID{Number: &tb},
			Address:   l.contractAddress,
			// Filter by first key = Open selector
			Keys: [][]*felt.Felt{{l.openEventSelector}},
		}

		input := rpc.EventsInput{
			EventFilter:       filter,
			ResultPageRequest: rpc.ResultPageRequest{ChunkSize: pageSize, ContinuationToken: cursor},
		}

		res, err := l.provider.Events(ctx, input)
		if err != nil {
			return newLast, fmt.Errorf("failed to fetch events: %w", err)
		}

		if len(res.Events) > 0 {
			fmt.Printf("📩 Found %d events on %s (blocks %d-%d)\n", len(res.Events), l.config.ChainName, fromBlock, toBlock)
		}

		// group by block
		byBlock := make(map[uint64][]rpc.EmittedEvent)
		for _, ev := range res.Events {
			byBlock[ev.BlockNumber] = append(byBlock[ev.BlockNumber], ev)
		}

		// Iterate blocks in range
		blockFailed := false
		for b := fromBlock; b <= toBlock; b++ {
			if evs, ok := byBlock[b]; ok {
				for _, ev := range evs {
					// Only handle Open events (first key == Open selector)
					isOpen := false
					if len(ev.Event.Keys) >= 1 {
						k0 := ev.Event.Keys[0].Bytes()
						openSel := l.openEventSelector.Bytes()
						k0b := k0[:]
						openb := openSel[:]
						if bytes.Equal(k0b, openb) {
							isOpen = true
						}
					}
					if !isOpen {
						continue
					}

					ro, derr := decodeResolvedOrderFromFelts(ev.Event.Data)
					if derr != nil {
						fmt.Printf("❌ Failed to decode ResolvedCrossChainOrder: %v\n", derr)
						blockFailed = true
						continue
					}

					parsedArgs := types.ParsedArgs{
						OrderID:       common.BytesToHash(ro.OrderID[:]).Hex(),
						SenderAddress: ro.User,
						Recipients:    []types.Recipient{{DestinationChainName: l.config.ChainName, RecipientAddress: "*"}},
						ResolvedOrder: ro,
					}

					settled, herr := handler(parsedArgs, l.config.ChainName, b)
					if herr != nil {
						fmt.Printf("❌ Failed to handle event: %v\n", herr)
						blockFailed = true
						continue
					}

					// Track settlement status (for now, assume all events are processed)
					// In a more sophisticated implementation, we'd use the actual settlement status
					_ = settled
				}
			}
			if blockFailed {
				break
			}
			newLast = b
		}

		if !blockFailed {
			break
		}
		retryCount++
		// Get max retries from config
		cfg, err := config.LoadConfig()
		maxRetries := 5 // fallback default
		if err == nil {
			maxRetries = cfg.MaxRetries
		}
		if retryCount >= maxRetries {
			fmt.Printf("⏭️  Giving up after %d retries for range %d-%d\n", retryCount, fromBlock, toBlock)
			break
		}
		fmt.Printf("🔁 Retry %d for range %d-%d\n", retryCount, fromBlock, toBlock)
		time.Sleep(500 * time.Millisecond)
		cursor = res.ContinuationToken
	}

	return newLast, nil
}

// --- Decoders ---

func decodeResolvedOrderFromFelts(data []*felt.Felt) (types.ResolvedCrossChainOrder, error) {
	idx := 0
	readFelt := func() *felt.Felt {
		f := data[idx]
		idx++
		return f
	}
	readU32 := func() uint32 {
		bi := utils.FeltToBigInt(readFelt())
		return uint32(bi.Uint64())
	}
	readU64 := func() uint64 {
		bi := utils.FeltToBigInt(readFelt())
		return bi.Uint64()
	}
	readU256 := func() *big.Int {
		low := utils.FeltToBigInt(readFelt())
		high := utils.FeltToBigInt(readFelt())
		return new(big.Int).Add(low, new(big.Int).Lsh(high, 128))
	}
	readAddress := func() string {
		b := readFelt().Bytes()
		return "0x" + hex.EncodeToString(b[:])
	}

	readOutput := func() types.Output {
		out := types.Output{}
		out.Token = readAddress()
		out.Amount = readU256()
		out.Recipient = readAddress()
		chainDomain := readU32()
		// Map domain to actual chain ID using config
		if chainID, err := domainToChainID(chainDomain); err == nil {
			out.ChainID = chainID
		} else {
			fmt.Printf("   ⚠️  Warning: Could not map domain %d to chain ID for output, using domain as chain ID\n", chainDomain)
			out.ChainID = new(big.Int).SetUint64(uint64(chainDomain))
		}
		return out
	}
	readOutputs := func() []types.Output {
		length := utils.FeltToBigInt(readFelt()).Uint64()
		outs := make([]types.Output, 0, length)
		for i := uint64(0); i < length; i++ {
			outs = append(outs, readOutput())
		}
		return outs
	}
	readFillInstruction := func() types.FillInstruction {
		fi := types.FillInstruction{}
		destinationDomain := readU32()
		// Map destination domain to actual chain ID using config
		if chainID, err := domainToChainID(destinationDomain); err == nil {
			fi.DestinationChainID = chainID
		} else {
			fmt.Printf("   ⚠️  Warning: Could not map domain %d to chain ID, using domain as chain ID\n", destinationDomain)
			fi.DestinationChainID = new(big.Int).SetUint64(uint64(destinationDomain))
		}
		fi.DestinationSettler = readAddress()

		// Parsing Cairo event data

		// Parse the origin_data bytes (OrderData struct) from the event data
		fmt.Printf("     📦 Parsing OrderData from Cairo event:\n")

		// Construct EVM-compatible origin_data from Cairo event data

		// Read size and u128 array length from the event data (absolute indices)
		size := utils.FeltToBigInt(data[21]).Uint64()
		u128ArrayLength := utils.FeltToBigInt(data[22]).Uint64()
		_ = size
		_ = u128ArrayLength

		// Parse each bytes32 field from the u128 array
		orderDataFields := make([][]byte, 0)
		for i := uint64(0); i < u128ArrayLength && (23+int(i)+1) < len(data); i += 2 {
			// Read two u128 felts and combine into bytes32
			lowFelt := data[23+int(i)]
			highFelt := data[23+int(i)+1]
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
		evmOriginData := make([]byte, 0, 448)
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
		if len(evmOriginData) != 448 {
			fmt.Printf("   ⚠️  origin_data unexpected length: %d\n", len(evmOriginData))
		}
		fi.OriginData = evmOriginData
		return fi
	}
	readFillInstructions := func() []types.FillInstruction {
		length := utils.FeltToBigInt(readFelt()).Uint64()
		arr := make([]types.FillInstruction, 0, length)
		for i := uint64(0); i < length; i++ {
			arr = append(arr, readFillInstruction())
		}
		return arr
	}

	ro := types.ResolvedCrossChainOrder{}
	ro.User = readAddress()
	ro.OriginChainID = new(big.Int).SetUint64(uint64(readU32()))
	ro.OpenDeadline = uint32(readU64())
	ro.FillDeadline = uint32(readU64())

	orderID := readU256()
	var orderArr [32]byte
	orderBytes := orderID.Bytes()
	copy(orderArr[32-len(orderBytes):], orderBytes)

	ro.OrderID = orderArr
	ro.MaxSpent = readOutputs()
	ro.MinReceived = readOutputs()
	ro.FillInstructions = readFillInstructions()
	return ro, nil
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
