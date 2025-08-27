package hyperlane7683

// Module: Starknet chain handler for Hyperlane7683
// - Coordinates StarknetFiller to perform fill/settle on Starknet
// - Resolves correct Hyperlane contract address and origin domain

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/types"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

// HyperlaneStarknet contains all Starknet-specific logic for the Hyperlane 7683 protocol
type HyperlaneStarknet struct {
	rpcURL string
	mu     sync.Mutex // Serialize operations to prevent nonce conflicts
}

// NewHyperlaneStarknet creates a new Starknet handler for Hyperlane operations
func NewHyperlaneStarknet(rpcURL string) *HyperlaneStarknet {
	return &HyperlaneStarknet{
		rpcURL: rpcURL,
	}
}

// Fill executes a fill operation on Starknet
func (h *HyperlaneStarknet) Fill(ctx context.Context, args types.ParsedArgs, originChainName string) error {
	fmt.Printf("   🔒 Acquiring Starknet mutex for order %s\n", args.OrderID)
	h.mu.Lock()
	defer func() {
		h.mu.Unlock()
		fmt.Printf("   🔓 Released Starknet mutex for order %s\n", args.OrderID)
	}()

	orderID := args.OrderID

	// Extract origin data from the first fill instruction
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]
	originData := instruction.OriginData

	//fmt.Printf("🔵 Starknet Fill: %s\n", orderID)

	// Get the proper Hyperlane address
	hyperlaneAddressHex, err := h.getHyperlaneAddress(args)
	if err != nil {
		return fmt.Errorf("failed to get Hyperlane address: %w", err)
	}

	// Create StarknetFiller instance
	sf, err := NewStarknetFiller(h.rpcURL, hyperlaneAddressHex)
	if err != nil {
		return fmt.Errorf("failed to create StarknetFiller: %w", err)
	}

	// Set up ERC20 approvals before filling (inside mutex to prevent concurrent approvals)
	if err := h.setupApprovals(ctx, sf, args); err != nil {
		return fmt.Errorf("failed to setup approvals: %w", err)
	}

	// Execute the fill
	fmt.Printf("   🚀 Proceeding with Starknet fill after approvals\n")
	return sf.Fill(ctx, orderID, originData)
}

// Settle executes settlement on Starknet
func (h *HyperlaneStarknet) Settle(ctx context.Context, args types.ParsedArgs) error {
	fmt.Printf("   🔒 Acquiring Starknet mutex for settlement of order %s\n", args.OrderID)
	h.mu.Lock()
	defer func() {
		h.mu.Unlock()
		fmt.Printf("   🔓 Released Starknet mutex for settlement of order %s\n", args.OrderID)
	}()

	orderID := args.OrderID

	fmt.Printf("🔵 Starknet Settle: %s\n", orderID)

	// Get the proper Hyperlane address
	hyperlaneAddressHex, err := h.getHyperlaneAddress(args)
	if err != nil {
		return fmt.Errorf("failed to get Hyperlane address: %w", err)
	}

	// Create StarknetFiller instance
	sf, err := NewStarknetFiller(h.rpcURL, hyperlaneAddressHex)
	if err != nil {
		return fmt.Errorf("failed to create StarknetFiller for settle: %w", err)
	}

	// Quote gas payment from Hyperlane contract
	originDomain, err := h.getOriginDomain(args)
	if err != nil {
		return fmt.Errorf("failed to get origin domain: %w", err)
	}
	fmt.Printf("   💰 Quoting gas payment for origin domain: %d\n", originDomain)

	gasPayment, err := sf.QuoteGasPayment(ctx, originDomain)
	if err != nil {
		return fmt.Errorf("failed to quote gas payment: %w", err)
	}

	fmt.Printf("   💰 Gas payment quoted: %s wei\n", gasPayment.String())
	// Approve ETH for the quoted gas amount
	if err := sf.EnsureETHApproval(ctx, gasPayment); err != nil {
		return fmt.Errorf("ETH approval failed for settlement gas: %w", err)
	}

	fmt.Printf("   ✅ ETH approved for settlement gas payment: %s wei\n", gasPayment.String())

	// Execute settlement
	if err := sf.Settle(ctx, orderID, gasPayment); err != nil {
		return fmt.Errorf("starknet settle send failed: %w", err)
	}

	fmt.Printf("   ✅ Starknet settlement completed successfully\n")
	return nil
}

// GetOrderStatus returns the current status of an order
func (h *HyperlaneStarknet) GetOrderStatus(ctx context.Context, args types.ParsedArgs) (string, error) {
	orderID := args.OrderID

	// Get the proper Hyperlane address
	hyperlaneAddressHex, err := h.getHyperlaneAddress(args)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to get Hyperlane address: %w", err)
	}

	// Use StarknetFiller's status helper
	sf, err := NewStarknetFiller(h.rpcURL, hyperlaneAddressHex)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to create StarknetFiller: %w", err)
	}

	processed, status, err := sf.isOrderProcessed(ctx, orderID)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to check order status: %w", err)
	}
	if !processed {
		return "UNKNOWN", nil
	}
	return h.interpretStarknetStatus(status), nil
}

// Helper methods
func (h *HyperlaneStarknet) getHyperlaneAddress(args types.ParsedArgs) (string, error) {
	// Use the destination settler address from the instruction
	if len(args.ResolvedOrder.FillInstructions) > 0 {
		instruction := args.ResolvedOrder.FillInstructions[0]
		if h.isStarknetChain(instruction.DestinationChainID) {
			// Use the destination settler address (already in correct format)
			fmt.Printf("   🎯 Using destination settler address from instruction: %s\n", instruction.DestinationSettler)
			return instruction.DestinationSettler, nil
		}
	}

	// Fallback to deployment state
	ds, err := deployer.GetDeploymentState()
	if err != nil {
		return "", fmt.Errorf("failed to load deployment state: %w", err)
	}

	if networkState, exists := ds.Networks["Starknet"]; exists && networkState.HyperlaneAddress != "" {
		hyperlaneAddressHex := networkState.HyperlaneAddress
		fmt.Printf("   🎯 Using deployment state Hyperlane address: %s\n", hyperlaneAddressHex)
		return hyperlaneAddressHex, nil
	}

	return "", fmt.Errorf("no Hyperlane address found for Starknet")
}

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

func (h *HyperlaneStarknet) setupApprovals(ctx context.Context, sf *StarknetFiller, args types.ParsedArgs) error {
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

		fmt.Printf("   📊 MaxSpent[%d] Token: %s, Amount: %s, Recipient: %s, ChainID: %s\n",
			i, maxSpent.Token, maxSpent.Amount.String(), maxSpent.Recipient, maxSpent.ChainID.String())

		// Convert token address to Starknet format
		tokenAddressHex := h.getTokenAddress(maxSpent)

		fmt.Printf("   🎯 TOKEN[%d] APPROVAL CALL:\n", i)
		fmt.Printf("     • Token address: %s\n", tokenAddressHex)
		fmt.Printf("     • Amount to approve: %s\n", maxSpent.Amount.String())

		if err := sf.EnsureTokenApproval(ctx, tokenAddressHex, maxSpent.Amount); err != nil {
			return fmt.Errorf("starknet approval failed for token %s: %w", tokenAddressHex, err)
		}

		fmt.Printf("   ✅ TOKEN[%d] approval completed\n", i)
	}

	return nil
}

func (h *HyperlaneStarknet) getTokenAddress(maxSpent types.Output) string {
	// For Starknet destinations, use the token address directly
	if h.isStarknetChain(maxSpent.ChainID) {
		fmt.Printf("   🎯 Using Starknet token address: %s\n", maxSpent.Token)
		return maxSpent.Token
	}

	// For EVM destinations, convert to Starknet format if needed
	fmt.Printf("   ⚠️  Using token address as-is: %s\n", maxSpent.Token)
	return maxSpent.Token
}

func (h *HyperlaneStarknet) interpretStarknetStatus(status string) string {
	switch status {
	case "0x0", "0":
		return "UNKNOWN"
	case "0x1", "1":
		return "FILLED"
	case "0x2", "2":
		return "SETTLED"
	default:
		return fmt.Sprintf("CUSTOM_%s", status)
	}
}

// DeriveOrderID creates an order ID for Starknet (for compatibility)
func (h *HyperlaneStarknet) DeriveOrderID(originData []byte) string {
	// For Starknet, we need to apply the same keccak256 but format for u256
	orderHash := utils.Keccak256(originData)
	return "0x" + hex.EncodeToString(orderHash)
}

// isStarknetChain checks if the given chain ID is a Starknet chain
func (h *HyperlaneStarknet) isStarknetChain(chainID *big.Int) bool {
	// Find any network with "Starknet" in the name that matches this chain ID
	for networkName, network := range config.Networks {
		if network.ChainID == chainID.Uint64() {
			// Check if network name contains "Starknet" (case insensitive)
			return strings.Contains(strings.ToLower(networkName), "starknet")
		}
	}
	return false
}

// StarknetFiller handles Starknet-specific filling logic
type StarknetFiller struct {
	provider      *rpc.Provider
	account       *account.Account
	hyperlaneAddr *felt.Felt
	solverAddr    *felt.Felt
}

// NewStarknetFiller creates a new StarknetFiller instance
func NewStarknetFiller(rpcURL string, hyperlaneAddressHex string) (*StarknetFiller, error) {
	provider, err := rpc.NewProvider(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Starknet provider: %w", err)
	}

	addrFelt, err := utils.HexToFelt(hyperlaneAddressHex)
	if err != nil {
		return nil, fmt.Errorf("invalid Starknet Hyperlane address: %w", err)
	}

	pub := os.Getenv("STARKNET_SOLVER_PUBLIC_KEY")
	addrHex := os.Getenv("STARKNET_SOLVER_ADDRESS")
	priv := os.Getenv("STARKNET_SOLVER_PRIVATE_KEY")
	if pub == "" || addrHex == "" || priv == "" {
		return nil, fmt.Errorf("missing STARKNET_SOLVER_* env vars for Starknet signer")
	}
	addrF, err := utils.HexToFelt(addrHex)
	if err != nil {
		return nil, fmt.Errorf("invalid STARKNET_SOLVER_ADDRESS: %w", err)
	}
	ks := account.NewMemKeystore()
	privBI, ok := new(big.Int).SetString(priv, 0)
	if !ok {
		return nil, fmt.Errorf("failed to parse STARKNET_SOLVER_PRIVATE_KEY")
	}
	ks.Put(pub, privBI)
	acct, err := account.NewAccount(provider, addrF, pub, ks, account.CairoV2)
	if err != nil {
		return nil, fmt.Errorf("failed to create Starknet account: %w", err)
	}

	return &StarknetFiller{provider: provider, account: acct, hyperlaneAddr: addrFelt, solverAddr: addrF}, nil
}

// GetTokenBalance retrieves the token balance for an address on Starknet
func (sf *StarknetFiller) GetTokenBalance(ctx context.Context, tokenAddressHex string, holderAddressHex string) (*big.Int, error) {
	// Convert addresses to felt format
	tokenAddr, err := utils.HexToFelt(tokenAddressHex)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	holderAddr, err := utils.HexToFelt(holderAddressHex)
	if err != nil {
		return nil, fmt.Errorf("invalid holder address: %w", err)
	}

	// Call balanceOf function on the ERC20 contract
	// balanceOf(address) -> uint256
	balanceOfSelector, err := utils.HexToFelt("0x2e17de78") // balanceOf selector
	if err != nil {
		return nil, fmt.Errorf("failed to create balanceOf selector: %w", err)
	}

	call := rpc.FunctionCall{
		ContractAddress:    tokenAddr,
		EntryPointSelector: balanceOfSelector,
		Calldata:           []*felt.Felt{holderAddr},
	}

	result, err := sf.provider.Call(ctx, call, rpc.BlockID{Tag: "latest"})
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("balanceOf returned no results")
	}

	// Convert felt result to big.Int
	balance := utils.FeltToBigInt(result[0])
	return balance, nil
}

// Fill executes the fill operation on Starknet
func (sf *StarknetFiller) Fill(ctx context.Context, orderIDHex string, originData []byte) error {
	// Skip if already processed (status != 0)
	if processed, status, err := sf.isOrderProcessed(ctx, orderIDHex); err == nil && processed {
		fmt.Printf("   ⏩ Skipping Starknet fill: order status=%s (non-zero)\n", status)
		return nil
	}

	// Build calldata for fill(order_id: u256, origin_data: Bytes, filler_data: Bytes)
	// order_id u256 -> two felts (low, high)
	orderBytes := utils.HexToBN(orderIDHex).Bytes()
	// pad to 32 bytes
	if len(orderBytes) < 32 {
		pad := make([]byte, 32-len(orderBytes))
		orderBytes = append(pad, orderBytes...)
	}
	// Cairo's OrderEncoder.id applies u256_reverse_endian to keccak bytes.
	// Map bytes32 (big-endian) to Cairo u256 as:
	// low = reverse(bytes[0:16]); high = reverse(bytes[16:32])
	rev := func(in []byte) []byte {
		out := make([]byte, len(in))
		for i := 0; i < len(in); i++ {
			out[i] = in[len(in)-1-i]
		}
		return out
	}
	lowBytes := rev(orderBytes[0:16])
	highBytes := rev(orderBytes[16:32])
	low := new(big.Int).SetBytes(lowBytes)
	high := new(big.Int).SetBytes(highBytes)
	lowF := utils.BigIntToFelt(low)
	highF := utils.BigIntToFelt(high)

	// origin_data Bytes: [size, words_len, words...], words are u128 (16-byte) chunks
	words := bytesToU128Felts(originData)

	calldata := make([]*felt.Felt, 0, 2+2+len(words)+2) // order_id(2) + size + len + data + empty filler_data
	calldata = append(calldata, lowF, highF)
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(originData))))
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(words))))
	calldata = append(calldata, words...)
	// filler_data: empty Bytes (size=0, len=0)
	calldata = append(calldata, utils.Uint64ToFelt(0), utils.Uint64ToFelt(0))

	// Calldata ready for Starknet contract
	invoke := rpc.InvokeFunctionCall{ContractAddress: sf.hyperlaneAddr, FunctionName: "fill", CallData: calldata}
	tx, err := sf.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return fmt.Errorf("starknet fill send failed: %w", err)
	}
	fmt.Printf("   🔄 Starknet fill tx sent: %s\n", tx.Hash.String())
	_, waitErr := sf.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet fill wait failed: %w", waitErr)
	}
	fmt.Printf("   ✅ Starknet fill transaction confirmed\n")

	return nil
}

// QuoteGasPayment calls the Starknet contract's quote_gas_payment function
func (sf *StarknetFiller) QuoteGasPayment(ctx context.Context, originDomain uint32) (*big.Int, error) {
	// Convert origin domain to felt
	domainFelt := utils.BigIntToFelt(big.NewInt(int64(originDomain)))

	// Call quote_gas_payment(origin_domain: u32) -> u256
	call := rpc.FunctionCall{
		ContractAddress:    sf.hyperlaneAddr,
		EntryPointSelector: utils.GetSelectorFromNameFelt("quote_gas_payment"),
		Calldata:           []*felt.Felt{domainFelt},
	}

	resp, err := sf.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
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
func (sf *StarknetFiller) EnsureETHApproval(ctx context.Context, amount *big.Int) error {
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
		Calldata:           []*felt.Felt{sf.solverAddr, sf.hyperlaneAddr},
	}

	resp, err := sf.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
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
	approveCalldata := []*felt.Felt{sf.hyperlaneAddr, lowFelt, highFelt}

	invoke := rpc.InvokeFunctionCall{
		ContractAddress: ethFelt,
		FunctionName:    "approve",
		CallData:        approveCalldata,
	}

	tx, err := sf.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return fmt.Errorf("starknet ETH approve send failed: %w", err)
	}

	fmt.Printf("   🔄 Starknet ETH approve tx sent: %s\n", tx.Hash.String())
	_, waitErr := sf.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet ETH approve wait failed: %w", waitErr)
	}

	fmt.Printf("   ✅ Starknet ETH approval confirmed\n")
	return nil
}

// EnsureTokenApproval ensures the solver has approved an arbitrary ERC20 token for the Hyperlane contract
func (sf *StarknetFiller) EnsureTokenApproval(ctx context.Context, tokenHex string, amount *big.Int) error {
	tokenFelt, err := utils.HexToFelt(tokenHex)
	if err != nil {
		return fmt.Errorf("invalid Starknet token address: %w", err)
	}

	// allowance(owner=solverAddr, spender=hyperlaneAddr) -> (low, high)
	call := rpc.FunctionCall{
		ContractAddress:    tokenFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{sf.solverAddr, sf.hyperlaneAddr},
	}

	resp, err := sf.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
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
		CallData:        []*felt.Felt{sf.hyperlaneAddr, lowF, highF},
	}

	tx, err := sf.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return fmt.Errorf("starknet token approve send failed: %w", err)
	}

	_, waitErr := sf.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet token approve wait failed: %w", waitErr)
	}
	return nil
}

// Settle calls the Starknet contract's settle function
func (sf *StarknetFiller) Settle(ctx context.Context, orderIDHex string, gasPayment *big.Int) error {
	// Convert order ID to two felts (low, high) for u256
	idBytes := utils.HexToBN(orderIDHex).Bytes()
	if len(idBytes) < 32 {
		idBytes = append(make([]byte, 32-len(idBytes)), idBytes...)
	}
	rev := func(in []byte) []byte {
		out := make([]byte, len(in))
		for i := 0; i < len(in); i++ {
			out[i] = in[len(in)-1-i]
		}
		return out
	}
	low := utils.BigIntToFelt(new(big.Int).SetBytes(rev(idBytes[0:16])))
	high := utils.BigIntToFelt(new(big.Int).SetBytes(rev(idBytes[16:32])))

	// Convert gas payment to two felts (low, high) for u256
	low128 := new(big.Int).And(gasPayment, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
	high128 := new(big.Int).Rsh(gasPayment, 128)
	gasLow := utils.BigIntToFelt(low128)
	gasHigh := utils.BigIntToFelt(high128)

	// Build calldata for settle(order_ids: Array<u256>, value: u256)
	// order_ids: [Array<u256>] -> [size, len, low, high]
	// value: u256 -> [low, high]
	calldata := []*felt.Felt{
		utils.Uint64ToFelt(1), // Array size = 1
		low, high,             // order_id u256
		gasLow, gasHigh,       // value u256
	}

	// Starknet settle transaction
	fmt.Printf("     • orderID: %s\n", orderIDHex)
	fmt.Printf("     • orderID.low: %s\n", low.String())
	fmt.Printf("     • orderID.high: %s\n", high.String())
	fmt.Printf("     • gasPayment: %s wei\n", gasPayment.String())
	fmt.Printf("     • gasPayment.low: %s\n", gasLow.String())
	fmt.Printf("     • gasPayment.high: %s\n", gasHigh.String())
	fmt.Printf("     • total calldata felts: %d\n", len(calldata))

	invoke := rpc.InvokeFunctionCall{
		ContractAddress: sf.hyperlaneAddr,
		FunctionName:    "settle",
		CallData:        calldata,
	}

	tx, err := sf.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		return fmt.Errorf("starknet settle send failed: %w", err)
	}

	fmt.Printf("   🔄 Starknet settle tx sent: %s\n", tx.Hash.String())
	_, waitErr := sf.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if waitErr != nil {
		return fmt.Errorf("starknet settle wait failed: %w", waitErr)
	}

	fmt.Printf("   ✅ Starknet settle transaction confirmed\n")
	return nil
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

// isOrderProcessed checks if an order has already been processed
func (sf *StarknetFiller) isOrderProcessed(ctx context.Context, orderIDHex string) (bool, string, error) {
	idBytes := utils.HexToBN(orderIDHex).Bytes()
	if len(idBytes) < 32 {
		idBytes = append(make([]byte, 32-len(idBytes)), idBytes...)
	}
	rev := func(in []byte) []byte {
		out := make([]byte, len(in))
		for i := 0; i < len(in); i++ {
			out[i] = in[len(in)-1-i]
		}
		return out
	}
	low := utils.BigIntToFelt(new(big.Int).SetBytes(rev(idBytes[0:16])))
	high := utils.BigIntToFelt(new(big.Int).SetBytes(rev(idBytes[16:32])))
	call := rpc.FunctionCall{ContractAddress: sf.hyperlaneAddr, EntryPointSelector: utils.GetSelectorFromNameFelt("order_status"), Calldata: []*felt.Felt{low, high}}
	resp, err := sf.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil || len(resp) == 0 {
		return false, "", err
	}
	status := resp[0].String()
	return status != "0x0" && status != "0", status, nil
}
