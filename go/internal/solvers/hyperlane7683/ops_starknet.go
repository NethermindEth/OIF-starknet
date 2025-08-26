package hyperlane7683

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

// StarknetOps provides JIT approvals and status checks for Starknet fills
type StarknetOps struct {
	sf *StarknetFiller
}

func NewStarknetOps(sf *StarknetFiller) *StarknetOps { return &StarknetOps{sf: sf} }

// EnsureApproval checks allowance(owner=solverAddr, spender=hyperlaneAddr) and approves exact amount if insufficient
func (ops *StarknetOps) EnsureApproval(ctx context.Context, tokenHex string, amount *big.Int) error {
	// 🔍 COMPREHENSIVE DEBUG: Log all allowance call parameters
	fmt.Printf("   🔍 EnsureApproval DEBUG:\n")
	fmt.Printf("     • Token: %s\n", tokenHex)
	fmt.Printf("     • Amount needed: %s\n", amount.String())
	fmt.Printf("     • Owner (solver): 0x%s\n", ops.sf.solverAddr.String())
	fmt.Printf("     • Spender (hyperlane): 0x%s\n", ops.sf.hyperlaneAddr.String())
	
	tokenFelt, err := utils.HexToFelt(tokenHex)
	if err != nil {
		return fmt.Errorf("invalid Starknet token address: %w", err)
	}
	owner := ops.sf.solverAddr
	spender := ops.sf.hyperlaneAddr

	// allowance(owner, spender): returns (low, high) u256
	call := rpc.FunctionCall{ContractAddress: tokenFelt, EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"), Calldata: []*felt.Felt{owner, spender}}
	
	fmt.Printf("     • Calling allowance(owner=0x%s, spender=0x%s) on token %s\n", owner.String(), spender.String(), tokenHex)
	
	resp, err := ops.sf.provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil {
		fmt.Printf("     ❌ Allowance call FAILED: %v\n", err)
		return fmt.Errorf("starknet allowance call failed: %w", err)
	}
	if len(resp) < 2 {
		fmt.Printf("     ❌ Allowance response too short: %d felts\n", len(resp))
		return fmt.Errorf("starknet allowance response too short: %d", len(resp))
	}
	
	low := utils.FeltToBigInt(resp[0])
	high := utils.FeltToBigInt(resp[1])
	current := new(big.Int).Add(low, new(big.Int).Lsh(high, 128))
	
	fmt.Printf("     • Current allowance: %s (low=%s, high=%s)\n", current.String(), low.String(), high.String())
	fmt.Printf("     • Need vs Have: %s vs %s\n", amount.String(), current.String())
	
	if current.Cmp(amount) >= 0 {
		fmt.Printf("     ✅ Starknet allowance sufficient: %s >= %s\n", current.String(), amount.String())
		return nil
	}

	fmt.Printf("     🚨 INSUFFICIENT! Setting approval for %s\n", amount.String())
	
	// approve(spender, amount) where amount is u256 split into low/high 128-bit felts
	low128 := new(big.Int).And(amount, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
	high128 := new(big.Int).Rsh(amount, 128)
	lowF := utils.BigIntToFelt(low128)
	highF := utils.BigIntToFelt(high128)

	fmt.Printf("     • Approve calldata: spender=0x%s, amount_low=%s, amount_high=%s\n", spender.String(), lowF.String(), highF.String())
	
	invoke := rpc.InvokeFunctionCall{ContractAddress: tokenFelt, FunctionName: "approve", CallData: []*felt.Felt{spender, lowF, highF}}
	tx, err := ops.sf.account.BuildAndSendInvokeTxn(ctx, []rpc.InvokeFunctionCall{invoke}, nil)
	if err != nil {
		fmt.Printf("     ❌ Approve send FAILED: %v\n", err)
		return fmt.Errorf("starknet approve send failed: %w", err)
	}
	
	fmt.Printf("     🚀 Approve tx sent: %s\n", tx.Hash.String())
	
	_, err = ops.sf.account.WaitForTransactionReceipt(ctx, tx.Hash, 2*time.Second)
	if err != nil {
		fmt.Printf("     ❌ Approve wait FAILED: %v\n", err)
		return fmt.Errorf("starknet approve wait failed: %w", err)
	}
	fmt.Printf("     ✅ Starknet approved token %s for spender 0x%s amount %s\n", tokenHex, ops.sf.hyperlaneAddr.String(), amount.String())
	return nil
}

// IsFilled returns whether the order status is non-zero
func (ops *StarknetOps) IsFilled(ctx context.Context, orderIDHex string) (bool, string, error) {
	return ops.sf.isOrderProcessed(ctx, orderIDHex)
}
