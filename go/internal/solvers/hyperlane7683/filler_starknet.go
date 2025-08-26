package hyperlane7683

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/crypto"
)

// StarknetFiller handles Starknet-specific filling logic
type StarknetFiller struct {
	provider      *rpc.Provider
	account       *account.Account
	hyperlaneAddr *felt.Felt
	solverAddr    *felt.Felt
}

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

	// Debug: log order id and origin data
	keccak := crypto.Keccak256(originData)
	fmt.Printf("   🧪 StarknetFill Debug\n")
	fmt.Printf("     • orderID: %s\n", orderIDHex)
	fmt.Printf("     • orderID.low: 0x%s\n", hex.EncodeToString(low.Bytes()))
	fmt.Printf("     • orderID.high: 0x%s\n", hex.EncodeToString(high.Bytes()))
	fmt.Printf("     • origin_data_len: %d bytes\n", len(originData))
	fmt.Printf("     • origin_data_hex: %s\n", hex.EncodeToString(originData))
	fmt.Printf("     • keccak256(origin_data): 0x%s\n", hex.EncodeToString(keccak))

	// origin_data Bytes: [size, words_len, words...], words are u128 (16-byte) chunks
	words := bytesToU128Felts(originData)

	// Balance checking removed - can be validated manually if needed

	calldata := make([]*felt.Felt, 0, 2+2+len(words)+2) // order_id(2) + size + len + data + empty filler_data
	calldata = append(calldata, lowF, highF)
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(originData))))
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(words))))
	calldata = append(calldata, words...)
	// filler_data: empty Bytes (size=0, len=0)
	calldata = append(calldata, utils.Uint64ToFelt(0), utils.Uint64ToFelt(0))

	// Log the calldata being sent to the Starknet contract
	fmt.Printf("   🧪 StarknetFill Calldata Debug\n")
	fmt.Printf("     • fill function: fill(order_id: u256, origin_data: Bytes, filler_data: Bytes)\n")
	fmt.Printf("     • order_id.low: %s\n", lowF.String())
	fmt.Printf("     • order_id.high: %s\n", highF.String())
	fmt.Printf("     • origin_data.size: %d\n", len(originData))
	fmt.Printf("     • origin_data.words_len: %d\n", len(words))
	fmt.Printf("     • filler_data.size: 0\n")
	fmt.Printf("     • filler_data.words_len: 0\n")
	fmt.Printf("     • total calldata felts: %d\n", len(calldata))

	// Log the first few calldata elements for debugging
	for i := 0; i < len(calldata) && i < 10; i++ {
		fmt.Printf("     • calldata[%d]: %s\n", i, calldata[i].String())
	}
	if len(calldata) > 10 {
		fmt.Printf("     • ... and %d more calldata elements\n", len(calldata)-10)
	}

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
		ContractAddress: sf.hyperlaneAddr,
		EntryPointSelector: utils.GetSelectorFromNameFelt("quote_gas_payment"),
		Calldata: []*felt.Felt{domainFelt},
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
		ContractAddress: ethFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata: []*felt.Felt{sf.solverAddr, sf.hyperlaneAddr},
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
		FunctionName: "approve",
		CallData: approveCalldata,
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
		utils.Uint64ToFelt(1),           // Array size = 1
		low, high,                        // order_id u256
		gasLow, gasHigh,                  // value u256
	}
	
	fmt.Printf("   🧪 StarknetSettle Debug\n")
	fmt.Printf("     • orderID: %s\n", orderIDHex)
	fmt.Printf("     • orderID.low: %s\n", low.String())
	fmt.Printf("     • orderID.high: %s\n", high.String())
	fmt.Printf("     • gasPayment: %s wei\n", gasPayment.String())
	fmt.Printf("     • gasPayment.low: %s\n", gasLow.String())
	fmt.Printf("     • gasPayment.high: %s\n", gasHigh.String())
	fmt.Printf("     • total calldata felts: %d\n", len(calldata))
	
	invoke := rpc.InvokeFunctionCall{
		ContractAddress: sf.hyperlaneAddr,
		FunctionName: "settle",
		CallData: calldata,
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

// --- Helpers ---
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
