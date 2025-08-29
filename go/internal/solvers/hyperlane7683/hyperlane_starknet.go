package hyperlane7683

// Module: Starknet chain handler for Hyperlane7683
// - Executes fill/settle/status calls against EVM Hyperlane7683 contracts
// - Manages ERC20 approvals and gas/value handling for calls
//
// Interface Contract:
// - Fill(): Must acquire mutex, setup approvals, execute fill, return OrderAction
// - Settle(): Must acquire mutex, quote gas, ensure ETH approval, execute settle
// - getOrderStatus(): Must check order status and return human-readable status
// - All methods should use consistent logging patterns and error handling

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/types"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

// HyperlaneStarknet contains all Starknet-specific logic for the Hyperlane 7683 protocol
type HyperlaneStarknet struct {
	// Client
	provider *rpc.Provider
	// Signer
	account    *account.Account
	solverAddr *felt.Felt

	//hyperlaneAddr *felt.Felt
	mu sync.Mutex // Serialize operations to prevent nonce conflicts
}

// NewHyperlaneStarknet creates a new Starknet handler for Hyperlane operations
func NewHyperlaneStarknet(rpcURL string) *HyperlaneStarknet {
	provider, err := rpc.NewProvider(rpcURL)
	if err != nil {
		fmt.Printf("failed to create Starknet provider: %v", err)
		return nil
	}

	pub := os.Getenv("STARKNET_SOLVER_PUBLIC_KEY")
	addrHex := os.Getenv("STARKNET_SOLVER_ADDRESS")
	priv := os.Getenv("STARKNET_SOLVER_PRIVATE_KEY")
	if pub == "" || addrHex == "" || priv == "" {
		fmt.Printf("missing STARKNET_SOLVER_* env vars for Starknet signer")
		return nil
	}

	addrF, err := utils.HexToFelt(addrHex)
	if err != nil {
		fmt.Printf("invalid STARKNET_SOLVER_ADDRESS: %v", err)
		return nil
	}

	ks := account.NewMemKeystore()
	privBI, ok := new(big.Int).SetString(priv, 0)
	if !ok {
		fmt.Printf("failed to parse STARKNET_SOLVER_PRIVATE_KEY")
		return nil
	}

	ks.Put(pub, privBI)
	acct, err := account.NewAccount(provider, addrF, pub, ks, account.CairoV2)
	if err != nil {
		fmt.Printf("failed to create Starknet account: %v", err)
		return nil
	}

	return &HyperlaneStarknet{
		account:    acct,
		provider:   provider,
		solverAddr: addrF,
	}
}

// Fill executes a fill operation on Starknet
func (h *HyperlaneStarknet) Fill(ctx context.Context, args types.ParsedArgs) (OrderAction, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return OrderActionError, fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	// Use the order ID from the event
	orderID := args.OrderID

	// Convert destination settler string to Starknet address (felt) for contract operations
	destinationSettlerAddr, err := types.ToStarknetAddress(instruction.DestinationSettler)
	if err != nil {
		return OrderActionError, fmt.Errorf("failed to convert destination settler to felt: %w", err)
	}

	// Pre-check: skip if order is already filled or settled
	status, err := h.getOrderStatus(ctx, args)
	if err != nil {
		return OrderActionError, err
	}
	if status == "FILLED" {
		return OrderActionSettle, nil
	}
	if status == "SETTLED" {
		return OrderActionSettle, nil
	}

	// Handle max spent approvals if needed
	if err := h.setupApprovals(ctx, args, destinationSettlerAddr); err != nil {
		return OrderActionError, fmt.Errorf("failed to setup approvals: %w", err)
	}

	// Prepare calldata; has a capacity of 6 + len(words)
	// - Order ID: 2 felts (u256)
	// - Origin data: 1 felt for size (usize), 1 felt for length (usize), 1 felt for each element
	// - Filler data: 1 felt for size (usize), 1 felt for length (usize), 0 elements
	originData := instruction.OriginData
	words := bytesToU128Felts(originData)

	// Convert bytes32 representation of orderID to u256 (2 felts)
	orderIDLow, orderIDHigh, err := convertSolidityOrderIDForStarknet(orderID)
	if err != nil {
		return OrderActionError, fmt.Errorf("failed to convert solidity order ID for starknet: %w", err)
	}

	calldata := make([]*felt.Felt, 0, 6+len(words))
	calldata = append(calldata, orderIDLow, orderIDHigh)
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(originData))))
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(words))))
	calldata = append(calldata, words...)
	calldata = append(calldata, utils.Uint64ToFelt(0), utils.Uint64ToFelt(0)) // empty (size=0, len=0)

	// Execute the fill transaction
	invoke := rpc.InvokeFunctionCall{ContractAddress: destinationSettlerAddr, FunctionName: "fill", CallData: calldata}
	tx, err := h.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return OrderActionError, fmt.Errorf("starknet fill send failed: %w", err)
	}
	fmt.Printf("   🚀 Starknet fill transaction sent: %s\n", tx.Hash.String())

	// Wait for confirmation
	_, waitErr := h.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return OrderActionError, fmt.Errorf("starknet fill wait failed: %w", waitErr)
	}
	fmt.Printf("   ✅ Starknet fill transaction confirmed\n")

	return OrderActionSettle, nil
}

// Settle executes settlement on Starknet
func (h *HyperlaneStarknet) Settle(ctx context.Context, args types.ParsedArgs) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	// Use the order ID from the event
	orderID := args.OrderID

	// Convert destination settler string to Starknet address (felt) for contract operations
	destinationSettler, err := types.ToStarknetAddress(instruction.DestinationSettler)
	if err != nil {
		return fmt.Errorf("failed to convert destination settler to felt: %w", err)
	}

	// Pre-settle check: ensure order is FILLED
	status, err := h.getOrderStatus(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}
	if status != "FILLED" {
		return fmt.Errorf("order status must be filled in order to settle, got: %s", status)
	}

	// Get gas payment (protocol fee) that must be sent with settlement
	originDomain, err := h.getOriginDomain(args)
	if err != nil {
		return fmt.Errorf("failed to get origin domain: %w", err)
	}

	fmt.Printf("   💰 Quoting gas payment for origin domain: %d\n", originDomain)
	gasPayment, err := h.quoteGasPayment(ctx, originDomain, destinationSettler)
	if err != nil {
		return fmt.Errorf("failed to quote gas payment: %w", err)
	}
	fmt.Printf("   💰 Gas payment quoted: %s wei\n", gasPayment.String())

	// Approve ETH for the quoted gas amount
	if err := h.ensureETHApproval(ctx, gasPayment, destinationSettler); err != nil {
		return fmt.Errorf("ETH approval failed for settlement gas: %w", err)
	}
	fmt.Printf("   ✅ ETH approved for settlement gas payment: %s wei\n", gasPayment.String())

	// Prepare calldata
	orderIDLow, orderIDHigh, err := convertSolidityOrderIDForStarknet(orderID)
	gasLow, gasHigh := convertBigIntToU256Felts(gasPayment)
	calldata := []*felt.Felt{
		utils.Uint64ToFelt(1),   // order ID array length
		orderIDLow, orderIDHigh, // order ID (u256) low and high
		gasLow, gasHigh, // gas amount (u256) low and high
	}

	// Execute the settle transaction
	invoke := rpc.InvokeFunctionCall{
		ContractAddress: destinationSettler,
		FunctionName:    "settle",
		CallData:        calldata,
	}

	// Wait for confirmation
	tx, err := h.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return fmt.Errorf("starknet settle send failed: %w", err)
	}

	fmt.Printf("   🔄 Starknet settle tx sent: %s\n", tx.Hash.String())
	_, waitErr := h.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet settle wait failed: %w", waitErr)
	}

	fmt.Printf("   ✅ Starknet settle transaction confirmed\n")
	return nil
}

// getOrderStatus returns the current status of an order
func (h *HyperlaneStarknet) getOrderStatus(ctx context.Context, args types.ParsedArgs) (string, error) {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return "UNKNOWN", fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	// Convert destination settler string to Starknet address for contract call
	destinationSettlerAddr, err := types.ToStarknetAddress(instruction.DestinationSettler)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to convert hex Hyperlane address to felt: %w", err)
	}

	// Convert order ID to cairo u256
	orderIDLow, orderIDHigh, err := convertSolidityOrderIDForStarknet(args.OrderID)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to convert solidity order id for cairo: %w", err)
	}

	call := rpc.FunctionCall{ContractAddress: destinationSettlerAddr, EntryPointSelector: utils.GetSelectorFromNameFelt("order_status"), Calldata: []*felt.Felt{orderIDLow, orderIDHigh}}
	resp, err := h.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil || len(resp) == 0 {
		return "UNKNOWN", err
	}
	status := resp[0].String()

	return h.interpretStarknetStatus(status), nil
}

// getOriginDomain returns the hyperlane domain of the order's origin chain
func (h *HyperlaneStarknet) getOriginDomain(args types.ParsedArgs) (uint32, error) {
	if args.ResolvedOrder.OriginChainID == nil {
		return 0, fmt.Errorf("no origin chain ID in resolved order")
	}

	chainID := args.ResolvedOrder.OriginChainID.Uint64()

	// Use the config system (.env) to find the domain for this chain ID
	for _, network := range config.Networks {
		if network.ChainID == chainID {
			return uint32(network.HyperlaneDomain), nil
		}
	}

	return 0, fmt.Errorf("no domain found for chain ID %d in config (check your .env file)", chainID)
}

// setupApprovals ensures each MaxSpent token allowances are set
func (h *HyperlaneStarknet) setupApprovals(ctx context.Context, args types.ParsedArgs, destinationSettler *felt.Felt) error {
	if len(args.ResolvedOrder.MaxSpent) == 0 {
		return nil
	}

	fmt.Printf("   🔍 Setting up Starknet ERC20 approvals before fill\n")

	for i, maxSpent := range args.ResolvedOrder.MaxSpent {
		// Skip native ETH (empty string)
		if maxSpent.Token == "" {
			fmt.Printf("   ⏭️  Skipping approval for native ETH (index %d)\n", i)
			continue
		}

		fmt.Printf("   📊 MaxSpent[%d] Token: %s, Amount: %s\n", i, maxSpent.Token, maxSpent.Amount.String())

		// Convert token address to Starknet format

		fmt.Printf("   🎯 TOKEN[%d] APPROVAL CALL:\n", i)
		fmt.Printf("     • Token address: %s\n", maxSpent.Token)
		fmt.Printf("     • Amount to approve: %s\n", maxSpent.Amount.String())

		if err := h.ensureTokenApproval(ctx, maxSpent.Token, maxSpent.Amount, destinationSettler); err != nil {
			return fmt.Errorf("starknet approval failed for token %s: %w", maxSpent.Token, err)
		}

		fmt.Printf("   ✅ TOKEN[%d] approval completed\n", i)
	}

	return nil
}

// interpretStarknetStatus returns the string representation of the order status
func (h *HyperlaneStarknet) interpretStarknetStatus(status string) string {
	switch status {
	case "0x0", "0":
		return "UNKNOWN"
	case "0x46494c4c4544":
		return "FILLED"
	case "0x534554544c4544":
		return "SETTLED"
	default:
		return fmt.Sprintf("%s", status)
	}
}

// quoteGasPayment calls the Starknet contract's quote_gas_payment function
func (f *HyperlaneStarknet) quoteGasPayment(ctx context.Context, originDomain uint32, hyperlaneAddress *felt.Felt) (*big.Int, error) {
	// Convert origin domain to felt
	domainFelt := utils.BigIntToFelt(big.NewInt(int64(originDomain)))

	// Call quote_gas_payment(origin_domain: u32) -> u256
	call := rpc.FunctionCall{
		ContractAddress:    hyperlaneAddress,
		EntryPointSelector: utils.GetSelectorFromNameFelt("quote_gas_payment"),
		Calldata:           []*felt.Felt{domainFelt},
	}

	resp, err := f.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("starknet quote_gas_payment call failed: %w", err)
	}

	if len(resp) < 2 {
		return nil, fmt.Errorf("starknet quote_gas_payment returned insufficient data: expected 2 felts, got %d", len(resp))
	}

	// Convert two felts (low, high) back to u256
	low := utils.FeltToBigInt(resp[0])
	high := utils.FeltToBigInt(resp[1])

	// Combine low and high into u256: (high << 128) | low
	result := new(big.Int).Lsh(high, 128)
	result.Or(result, low)

	return result, nil
}

// EnsureETHApproval ensures the solver has approved the ETH address for settlement
func (h *HyperlaneStarknet) ensureETHApproval(ctx context.Context, amount *big.Int, hyperlaneAddress *felt.Felt) error {
	// Hard-coded ETH address on Starknet
	ethAddress := "0x49d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	ethFelt, err := utils.HexToFelt(ethAddress)
	if err != nil {
		return fmt.Errorf("failed to convert ETH address to felt: %w", err)
	}

	// Check current allowance
	call := rpc.FunctionCall{
		ContractAddress:    ethFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{h.solverAddr, hyperlaneAddress},
	}

	resp, err := h.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil {
		return fmt.Errorf("starknet ETH allowance call failed: %w", err)
	}

	if len(resp) < 2 {
		return fmt.Errorf("starknet ETH allowance returned insufficient data: expected 2 felts, got %d", len(resp))
	}

	// Convert two felts (low, high) back to u256
	low := utils.FeltToBigInt(resp[0])
	high := utils.FeltToBigInt(resp[1])
	currentAllowance := new(big.Int).Lsh(high, 128)
	currentAllowance.Or(currentAllowance, low)

	// If allowance is sufficient, no need to approve
	if currentAllowance.Cmp(amount) >= 0 {
		fmt.Printf("   ✅ ETH allowance sufficient: %s >= %s\n", currentAllowance.String(), amount.String())
		return nil
	}

	// Need to approve - convert amount to two felts (low, high)
	low128 := new(big.Int).And(amount, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
	high128 := new(big.Int).Rsh(amount, 128)

	lowFelt := utils.BigIntToFelt(low128)
	highFelt := utils.BigIntToFelt(high128)

	// Build approve calldata: approve(spender: felt, amount: u256)
	approveCalldata := []*felt.Felt{hyperlaneAddress, lowFelt, highFelt}

	invoke := rpc.InvokeFunctionCall{
		ContractAddress: ethFelt,
		FunctionName:    "approve",
		CallData:        approveCalldata,
	}

	tx, err := h.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return fmt.Errorf("starknet ETH approve send failed: %w", err)
	}

	fmt.Printf("   🔄 Starknet ETH approve tx sent: %s\n", tx.Hash.String())
	_, waitErr := h.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet ETH approve wait failed: %w", waitErr)
	}

	fmt.Printf("   ✅ Starknet ETH approval confirmed\n")
	return nil
}

// ensureTokenApproval ensures the solver has approved an arbitrary ERC20 token for the Hyperlane contract
func (h *HyperlaneStarknet) ensureTokenApproval(ctx context.Context, tokenHex string, amount *big.Int, hyperlaneAddress *felt.Felt) error {
	tokenFelt, err := utils.HexToFelt(tokenHex)
	if err != nil {
		return fmt.Errorf("invalid Starknet token address: %w", err)
	}

	// allowance(owner=solverAddr, spender=hyperlaneAddr) -> (low, high)
	call := rpc.FunctionCall{
		ContractAddress:    tokenFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{h.solverAddr, hyperlaneAddress},
	}

	resp, err := h.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil {
		return fmt.Errorf("starknet allowance call failed: %w", err)
	}
	if len(resp) < 2 {
		return fmt.Errorf("starknet allowance response too short: %d", len(resp))
	}

	low := utils.FeltToBigInt(resp[0])
	high := utils.FeltToBigInt(resp[1])
	current := new(big.Int).Add(low, new(big.Int).Lsh(high, 128))
	if current.Cmp(amount) >= 0 {
		return nil
	}

	// Approve exact amount: approve(spender: felt, amount: u256)
	low128 := new(big.Int).And(amount, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
	high128 := new(big.Int).Rsh(amount, 128)
	lowF := utils.BigIntToFelt(low128)
	highF := utils.BigIntToFelt(high128)

	invoke := rpc.InvokeFunctionCall{
		ContractAddress: tokenFelt,
		FunctionName:    "approve",
		CallData:        []*felt.Felt{hyperlaneAddress, lowF, highF},
	}

	tx, err := h.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return fmt.Errorf("starknet token approve send failed: %w", err)
	}

	_, waitErr := h.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet token approve wait failed: %w", waitErr)
	}
	return nil
}

// convertSolidityOrderIDForStarknet converts a Solidity-style orderID (bytes32) into the low and high felts of a Starknet u256 orderID
// Note: Assigns the left 16 bytes to the high felt and the right 16 bytes to the low felt
func convertSolidityOrderIDForStarknet(orderID string) (low *felt.Felt, high *felt.Felt, err error) {
	orderBytes := utils.HexToBN(orderID).Bytes()
	if len(orderBytes) < 32 {
		pad := make([]byte, 32-len(orderBytes))
		orderBytes = append(pad, orderBytes...)
	}

	left16 := utils.BigIntToFelt(new(big.Int).SetBytes(orderBytes[0:16]))
	right16 := utils.BigIntToFelt(new(big.Int).SetBytes(orderBytes[16:32]))

	low = right16
	high = left16

	return low, high, nil
}

// convertBigIntToU256Felts converts a big.Int to two felts, one for the low 128 bits and one for the high 128 bits
func convertBigIntToU256Felts(value *big.Int) (low *felt.Felt, high *felt.Felt) {
	lowerMask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	low = utils.BigIntToFelt(new(big.Int).And(value, lowerMask))
	high = utils.BigIntToFelt(new(big.Int).Rsh(value, 128))
	return low, high
}

// bytesToU128Felts converts bytes to u128 felts for Cairo
func bytesToU128Felts(b []byte) []*felt.Felt {
	words := make([]*felt.Felt, 0, (len(b)+15)/16)
	for i := 0; i < len(b); i += 16 {
		end := i + 16
		chunk := make([]byte, 16)
		if end > len(b) {
			copy(chunk, b[i:])
		} else {
			copy(chunk, b[i:end])
		}
		// Keep big-endian u128 words; Cairo decoders reconstruct bytes in order
		words = append(words, utils.BigIntToFelt(new(big.Int).SetBytes(chunk)))
	}
	return words
}
