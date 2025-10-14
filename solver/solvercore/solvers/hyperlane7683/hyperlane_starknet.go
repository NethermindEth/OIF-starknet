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
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/NethermindEth/oif-starknet/solver/pkg/starknetutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/logutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

const (
	// Calldata constants
	calldataBaseSize = 6
	// EVM origin data size (bytes)
	evmOriginDataSize = 448
)

// HyperlaneStarknet contains all Starknet-specific logic for the Hyperlane 7683 protocol
type HyperlaneStarknet struct {
	// Client
	provider *rpc.Provider
	// Signer
	account    *account.Account
	solverAddr *felt.Felt
	chainID    uint64

	// hyperlaneAddr *felt.Felt
	mu sync.Mutex // Serialize operations to prevent nonce conflicts
}

// NewHyperlaneStarknet creates a new Starknet handler for Hyperlane operations
func NewHyperlaneStarknet(rpcURL string, chainID uint64) *HyperlaneStarknet {
	provider, err := rpc.NewProvider(rpcURL)
	if err != nil {
		fmt.Printf("failed to create Starknet provider: %v", err)
		return nil
	}

	// Use conditional environment variables based on IS_DEVNET
	pub := envutil.GetStarknetSolverPublicKey()
	addrHex := envutil.GetStarknetSolverAddress()
	priv := envutil.GetStarknetSolverPrivateKey()

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
		chainID:    chainID,
		mu:         sync.Mutex{},
	}
}

// Fill executes a fill operation on Starknet
func (h *HyperlaneStarknet) Fill(ctx context.Context, args *types.ParsedArgs) (OrderAction, error) {
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
	status, err := h.GetOrderStatus(ctx, args)
	if err != nil {
		fmt.Printf("   ⚠️  Status check failed: %v\n", err)
		return OrderActionError, err
	}
	networkName := logutil.NetworkNameByChainID(h.chainID)
	logutil.LogStatusCheck(networkName, 1, 1, status, orderStatusUnknown)
	if status == orderStatusFilled {
		fmt.Printf("⏭️  Order already filled, proceeding to settlement\n")
		return OrderActionSettle, nil
	}
	if status == orderStatusSettled {
		fmt.Printf("🎉  Order already settled, nothing to do\n")
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
	words := starknetutil.BytesToU128Felts(originData)

	// Convert bytes32 representation of orderID to u256 (2 felts)
	orderIDLow, orderIDHigh, err := starknetutil.ConvertSolidityOrderIDForStarknet(orderID)
	if err != nil {
		return OrderActionError, fmt.Errorf("failed to convert solidity order ID for starknet: %w", err)
	}

	calldata := make([]*felt.Felt, 0, calldataBaseSize+len(words))
	calldata = append(calldata, 
		orderIDLow, orderIDHigh,
		utils.Uint64ToFelt(uint64(len(originData))),
		utils.Uint64ToFelt(uint64(len(words))),
	)
	calldata = append(calldata, words...)
	calldata = append(calldata, utils.Uint64ToFelt(0), utils.Uint64ToFelt(0)) // empty (size=0, len=0)

	// Execute the fill transaction
	invoke := rpc.InvokeFunctionCall{ContractAddress: destinationSettlerAddr, FunctionName: "fill", CallData: calldata}
	tx, err := h.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return OrderActionError, fmt.Errorf("starknet fill send failed: %w", err)
	}
	// Get chain IDs for cross-chain logging
	originChainID := args.ResolvedOrder.OriginChainID.Uint64()
	destChainID := instruction.DestinationChainID.Uint64()
	logutil.CrossChainOperation(fmt.Sprintf("Fill transaction sent: %s", tx.Hash.String()), originChainID, destChainID, orderID)

	// Wait for confirmation
	_, waitErr := h.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return OrderActionError, fmt.Errorf("starknet fill wait failed: %w", waitErr)
	}
	logutil.CrossChainOperation("Fill transaction confirmed", originChainID, destChainID, orderID)

	return OrderActionSettle, nil
}

// Settle executes settlement on Starknet
func (h *HyperlaneStarknet) Settle(ctx context.Context, args *types.ParsedArgs) error {
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

	// Pre-settle check: ensure order is FILLED with retry logic
	status, err := h.waitForOrderStatus(ctx, args, orderStatusFilled, 5, 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get order status after retries: %w", err)
	}
	if status != orderStatusFilled {
		return fmt.Errorf("order status must be filled in order to settle, got: %s", status)
	}

	// Get gas payment (protocol fee) that must be sent with settlement
	originDomain, err := h.getOriginDomain(args)
	if err != nil {
		return fmt.Errorf("failed to get origin domain: %w", err)
	}

	// Get chain IDs for cross-chain logging
	originChainID := args.ResolvedOrder.OriginChainID.Uint64()
	destChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID.Uint64()
	logutil.CrossChainOperation(fmt.Sprintf("Quoting gas payment for origin domain: %d", originDomain), originChainID, destChainID, args.OrderID)
	gasPayment, err := h.quoteGasPayment(ctx, originDomain, destinationSettler)
	if err != nil {
		return fmt.Errorf("failed to quote gas payment: %w", err)
	}

	// Approve ETH for the quoted gas amount
	if err := h.ensureETHApproval(ctx, gasPayment, destinationSettler); err != nil {
		return fmt.Errorf("ETH approval failed for settlement gas: %w", err)
	}
	logutil.CrossChainOperation(fmt.Sprintf("ETH approved for settlement gas payment: %s wei", gasPayment.String()), originChainID, destChainID, args.OrderID)

	// Prepare calldata
	orderIDLow, orderIDHigh, err := starknetutil.ConvertSolidityOrderIDForStarknet(orderID)
	if err != nil {
		return fmt.Errorf("failed to convert solidity order ID for starknet: %w", err)
	}
	gasLow, gasHigh := starknetutil.ConvertBigIntToU256Felts(gasPayment)
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

	logutil.CrossChainOperation(fmt.Sprintf("Starknet settle tx sent: %s", tx.Hash.String()), originChainID, destChainID, args.OrderID)
	_, waitErr := h.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet settle wait failed: %w", waitErr)
	}

	logutil.CrossChainOperation("Starknet settle transaction confirmed", originChainID, destChainID, args.OrderID)
	return nil
}

// GetOrderStatus returns the current status of an order
func (h *HyperlaneStarknet) GetOrderStatus(ctx context.Context, args *types.ParsedArgs) (string, error) {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return orderStatusUnknown, fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	// Convert destination settler string to Starknet address for contract call
	destinationSettlerAddr, err := types.ToStarknetAddress(instruction.DestinationSettler)
	if err != nil {
		return orderStatusUnknown, fmt.Errorf("failed to convert hex Hyperlane address to felt: %w", err)
	}

	// Convert order ID to cairo u256
	orderIDLow, orderIDHigh, err := starknetutil.ConvertSolidityOrderIDForStarknet(args.OrderID)
	if err != nil {
		return orderStatusUnknown, fmt.Errorf("failed to convert solidity order id for cairo: %w", err)
	}

	call := rpc.FunctionCall{
		ContractAddress:      destinationSettlerAddr,
		EntryPointSelector:   utils.GetSelectorFromNameFelt("order_status"),
		Calldata:             []*felt.Felt{orderIDLow, orderIDHigh},
	}
	resp, err := h.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil || len(resp) == 0 {
		return orderStatusUnknown, err
	}
	status := resp[0].String()

	return h.interpretStarknetStatus(status), nil
}

// getOriginDomain returns the hyperlane domain of the order's origin chain
func (h *HyperlaneStarknet) getOriginDomain(args *types.ParsedArgs) (uint32, error) {
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
func (h *HyperlaneStarknet) setupApprovals(ctx context.Context, args *types.ParsedArgs, destinationSettler *felt.Felt) error {
	if len(args.ResolvedOrder.MaxSpent) == 0 {
		return nil
	}

	// Get destination chain ID from fill instruction
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}
	destinationChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID.Uint64()

	// Get origin chain ID for cross-chain logging
	originChainID := args.ResolvedOrder.OriginChainID.Uint64()

	for _, maxSpent := range args.ResolvedOrder.MaxSpent {
		// Skip native ETH (empty string)
		if maxSpent.Token == "" {
			continue
		}

		// Only approve tokens that belong to this chain (destination chain)
		if maxSpent.ChainID.Uint64() != destinationChainID {
			fmt.Printf("   ⚠️  Skipping approval for token %s on chain %d (this handler is for chain %d)\n",
				maxSpent.Token, maxSpent.ChainID.Uint64(), destinationChainID)
			continue
		}

		// Convert token address to Starknet format
		if err := h.ensureTokenApproval(ctx, maxSpent.Token, maxSpent.Amount, destinationSettler); err != nil {
			return fmt.Errorf("starknet approval failed for token %s: %w", maxSpent.Token, err)
		}
	}

	logutil.CrossChainOperation("Set token approvals", originChainID, destinationChainID, args.OrderID)

	// Add a small delay to ensure blockchain state is updated after approvals
	time.Sleep(1 * time.Second)

	return nil
}

// interpretStarknetStatus returns the string representation of the order status
func (h *HyperlaneStarknet) interpretStarknetStatus(status string) string {
	switch status {
	case "0x0", "0":
		return orderStatusUnknown
	case "0x46494c4c4544":
		return orderStatusFilled
	case "0x534554544c4544":
		return orderStatusSettled
	default:
		return status
	}
}

// quoteGasPayment calls the Starknet contract's quote_gas_payment function
func (h *HyperlaneStarknet) quoteGasPayment(ctx context.Context, originDomain uint32, hyperlaneAddress *felt.Felt) (*big.Int, error) {
	// Convert origin domain to felt
	domainFelt := utils.BigIntToFelt(big.NewInt(int64(originDomain)))

	// Call quote_gas_payment(origin_domain: u32) -> u256
	call := rpc.FunctionCall{
		ContractAddress:    hyperlaneAddress,
		EntryPointSelector: utils.GetSelectorFromNameFelt("quote_gas_payment"),
		Calldata:           []*felt.Felt{domainFelt},
	}

	resp, err := h.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
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

// waitForOrderStatus waits for the order status to become the expected value with retry logic
func (h *HyperlaneStarknet) waitForOrderStatus(
	ctx context.Context,
	args *types.ParsedArgs,
	expectedStatus string,
	maxRetries int,
	initialDelay time.Duration,
) (string, error) {
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		status, err := h.GetOrderStatus(ctx, args)
		if err != nil {
			fmt.Printf("   ⚠️  Status check attempt %d failed: %v\n", attempt, err)
		} else {
			fmt.Printf("   📊 Status check attempt %d: %s (expected: %s)\n", attempt, status, expectedStatus)
			if status == expectedStatus {
				return status, nil
			}
		}

		// Don't wait after the last attempt
		if attempt < maxRetries {
			fmt.Printf("   ⏳ Waiting %v before retry %d/%d...\n", delay, attempt+1, maxRetries)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
				// Exponential backoff: double the delay for next attempt
				delay *= 2
			}
		}
	}

	// Final attempt to get the current status
	finalStatus, err := h.GetOrderStatus(ctx, args)
	if err != nil {
		return orderStatusUnknown, fmt.Errorf("final status check failed after %d attempts: %w", maxRetries, err)
	}

	return finalStatus, nil
}
