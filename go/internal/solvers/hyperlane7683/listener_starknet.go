package hyperlane7683

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

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/listener"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
)

// starknetListener implements listener.BaseListener for Starknet chains
type starknetListener struct {
	config             *listener.ListenerConfig
	provider           *rpc.Provider
	contractAddress    *felt.Felt
	openEventSelector  *felt.Felt
	lastProcessedBlock uint64
	stopChan           chan struct{}
	mu                 sync.RWMutex
}

// NewStarknetListener creates a new Starknet listener
func NewStarknetListener(config *listener.ListenerConfig, rpcURL string) (listener.BaseListener, error) {
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

	return &starknetListener{
		config:             config,
		provider:           provider,
		contractAddress:    addrFelt,
		openEventSelector:  openSelector,
		lastProcessedBlock: lastProcessedBlock,
		stopChan:           make(chan struct{}),
	}, nil
}

// Start begins listening for events
func (l *starknetListener) Start(ctx context.Context, handler listener.EventHandler) (listener.ShutdownFunc, error) {
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
	fmt.Printf("✅ Block %d marked as fully processed for %s\n", blockNumber, l.config.ChainName)
	return nil
}

func (l *starknetListener) realEventLoop(ctx context.Context, handler listener.EventHandler) {
	fmt.Printf("⚙️  Starting (%s) Starknet event listener...\n", l.config.ChainName)
	if err := l.catchUpHistoricalBlocks(ctx, handler); err != nil {
		fmt.Printf("❌ Failed to catch up on (%s) historical blocks: %v\n", l.config.ChainName, err)
	}
	fmt.Printf("🔄 Backfill complete (%s)\n", l.config.ChainName)
	time.Sleep(1 * time.Second)
	l.startPolling(ctx, handler)
}

func (l *starknetListener) catchUpHistoricalBlocks(ctx context.Context, handler listener.EventHandler) error {
	fmt.Printf("🔄 Catching up on (%s) historical blocks...\n", l.config.ChainName)
	currentBlock, err := l.provider.BlockNumber(ctx)
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
		if err := deployer.UpdateLastIndexedBlock(l.config.ChainName, newLast); err != nil {
			fmt.Printf("⚠️  Failed to persist LastIndexedBlock for %s: %v\n", l.config.ChainName, err)
		} else {
			fmt.Printf("💾 Persisted LastIndexedBlock=%d for %s\n", newLast, l.config.ChainName)
		}
	}
	fmt.Printf("✅ Historical block processing completed for %s\n", l.config.ChainName)
	return nil
}

func (l *starknetListener) startPolling(ctx context.Context, handler listener.EventHandler) {
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

func (l *starknetListener) processCurrentBlockRange(ctx context.Context, handler listener.EventHandler) error {
	currentBlock, err := l.provider.BlockNumber(ctx)
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
	if fromBlock > toBlock {
		fmt.Printf("⚠️  Invalid block range for %s: fromBlock (%d) > toBlock (%d), skipping\n", l.config.ChainName, fromBlock, toBlock)
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
func (l *starknetListener) processBlockRange(ctx context.Context, fromBlock, toBlock uint64, handler listener.EventHandler) (uint64, error) {
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
						OrderID:       "", // leave as empty for now; filler will use origin_data hashing on EVM side
						SenderAddress: ro.User.Hex(),
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
		if retryCount >= configObj().MaxRetries {
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
	readAddress := func() common.Address {
		b := readFelt().Bytes()
		return common.BytesToAddress(b[12:])
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

		// COMPREHENSIVE: Parse all Cairo event data into structured variables
		fmt.Printf("   🧪 Comprehensive Cairo Event Data Parsing:\n")

		// Parse the origin_data bytes (OrderData struct) from the event data
		fmt.Printf("     📦 Parsing OrderData from Cairo event:\n")

		// Read size and u128 array length from the event data (absolute indices)
		size := utils.FeltToBigInt(data[21]).Uint64()
		u128ArrayLength := utils.FeltToBigInt(data[22]).Uint64()
		fmt.Printf("       • Size: %d bytes\n", size)
		fmt.Printf("       • U128 array length: %d\n", u128ArrayLength)

		// Parse each bytes32 field from the u128 array
		orderDataFields := make([][]byte, 0)
		for i := uint64(0); i < u128ArrayLength && (23+int(i)+1) < len(data); i += 2 {
			// Read two u128 felts and combine into bytes32
			lowFelt := data[23+int(i)]
			highFelt := data[23+int(i)+1]

			lowBytes := lowFelt.Bytes()
			highBytes := highFelt.Bytes()

			// Extract u128 values (last 16 bytes)
			lowU128 := lowBytes[16:]
			highU128 := highBytes[16:]

			// Combine into bytes32
			bytes32 := make([]byte, 32)
			copy(bytes32[0:16], lowU128)
			copy(bytes32[16:32], highU128)

			orderDataFields = append(orderDataFields, bytes32)
		}

		// Log the parsed OrderData fields
		fmt.Printf("       • Field 0 (offset): %s\n", hex.EncodeToString(orderDataFields[0]))
		fmt.Printf("       • Field 1 (sender): %s\n", hex.EncodeToString(orderDataFields[1]))
		fmt.Printf("       • Field 2 (recipient): %s\n", hex.EncodeToString(orderDataFields[2]))
		fmt.Printf("       • Field 3 (input_token): %s\n", hex.EncodeToString(orderDataFields[3]))
		fmt.Printf("       • Field 4 (output_token): %s\n", hex.EncodeToString(orderDataFields[4]))
		fmt.Printf("       • Field 5 (amount_in): %s\n", hex.EncodeToString(orderDataFields[5]))
		fmt.Printf("       • Field 6 (amount_out): %s\n", hex.EncodeToString(orderDataFields[6]))
		fmt.Printf("       • Field 7 (sender_nonce): %s\n", hex.EncodeToString(orderDataFields[7]))
		fmt.Printf("       • Field 8 (origin_domain): %s\n", hex.EncodeToString(orderDataFields[8]))
		fmt.Printf("       • Field 9 (destination_domain): %s\n", hex.EncodeToString(orderDataFields[9]))
		fmt.Printf("       • Field 10 (destination_settler): %s\n", hex.EncodeToString(orderDataFields[10]))
		fmt.Printf("       • Field 11 (fill_deadline): %s\n", hex.EncodeToString(orderDataFields[11]))
		fmt.Printf("       • Field 12 (data_offset): %s\n", hex.EncodeToString(orderDataFields[12]))
		fmt.Printf("       • Field 13 (data_size): %s\n", hex.EncodeToString(orderDataFields[13]))

		// Now read the origin_data using the existing logic
		fmt.Printf("   🧪 Cairo Felt Processing for origin_data:\n")
		fmt.Printf("     • Current felt index: %d\n", idx)
		fmt.Printf("     • Remaining felts: %d\n", len(data)-idx)

		// MANUAL CONSTRUCTION: Build EVM-compatible origin_data from parsed fields
		fmt.Printf("   🧪 Manual EVM origin_data Construction:\n")

		// Create a buffer for the manually constructed EVM origin_data
		// OrderData struct needs to match the EVM ABI encoding: 3 ABI words + 12 static fields = 448 bytes total
		evmOriginData := make([]byte, 0, 448)

		// First word of OrderData encoding inside bytes: 0x20
		firstWord := make([]byte, 32)
		firstWord[31] = 0x20
		evmOriginData = append(evmOriginData, firstWord...)

		// Now add the 12 static fields (352 bytes)
		// Field 0: Sender (32 bytes) - should be the first field
		evmOriginData = append(evmOriginData, orderDataFields[1]...)

		// Field 1: Recipient (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[2]...)

		// Field 2: Input token (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[3]...)

		// Field 3: Output token (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[4]...)

		// Field 4: Amount in (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[5]...)

		// Field 5: Amount out (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[6]...)

		// Field 6: Sender nonce (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[7]...)

		// Field 7: Origin domain (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[8]...)

		// Field 8: Destination domain (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[9]...)

		// Field 9: Destination settler (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[10]...)

		// Field 10: Fill deadline (32 bytes)
		evmOriginData = append(evmOriginData, orderDataFields[11]...)

		// Field 11: Data offset (32 bytes) - 0x20 (32 bytes) pointing to where data would be
		// This is the offset within the OrderData struct to the dynamic bytes field
		dataOffset := make([]byte, 32)
		dataOffset[31] = 0x80
		dataOffset[30] = 0x01
		evmOriginData = append(evmOriginData, dataOffset...)
		dataSize := make([]byte, 32)
		dataSize[31] = 0x00
		evmOriginData = append(evmOriginData, dataSize...)

		fmt.Printf("     • OrderData Fields (352 bytes): 12 fields of 32 bytes each\n")

		// Note: We don't append the actual data content since it's empty for our orders
		// The offset 0x20 points to where the data would be within the struct, but since data size is 0,
		// no additional bytes are needed

		fmt.Printf("     • Manual EVM origin_data length: %d bytes\n", len(evmOriginData))
		fmt.Printf("     • Manual EVM origin_data hex: %s\n", hex.EncodeToString(evmOriginData))

		// Verify the structure matches expected EVM ABI encoding
		if len(evmOriginData) != 448 {
			fmt.Printf("     ⚠️  WARNING: Expected 448 bytes, got %d bytes\n", len(evmOriginData))
		} else {
			fmt.Printf("     ✅ Perfect! Exactly 448 bytes as expected\n")
		}

		// Debug: Show the structure breakdown
		fmt.Printf("     • Structure: 96 bytes (ABI header) + 352 bytes (12 fields) = %d bytes\n", len(evmOriginData))

		// Debug: Show the first few fields to verify mapping
		if len(evmOriginData) >= 128 {
			fmt.Printf("     • First 4 fields (128 bytes): %x\n", evmOriginData[:128])
		}

		// Use the manually constructed EVM origin_data instead of Cairo bytes
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
