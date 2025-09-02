package starknetutil

// Module: Starknet utilities for common operations
// - Provides shared utilities for Starknet token operations
// - Mirrors the functionality of ethutil for EVM chains
// - Reduces code duplication across the codebase

import (
	"context"
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/holiman/uint256"
)

// Constants for Starknet operations
const (
	U128BitShift = 128
	Bytes32Length = 32
	Bytes16Length = 16
	TokenDecimals = 18
)

// Helper functions for uint256 conversion
// ToUint256 converts *big.Int to uint256.Int
func ToUint256(bi *big.Int) *uint256.Int {
	if bi == nil {
		return uint256.NewInt(0)
	}
	u, _ := uint256.FromBig(bi)
	return u
}

// ToBigInt converts uint256.Int to *big.Int
func ToBigInt(u *uint256.Int) *big.Int {
	return u.ToBig()
}

// StarknetERC20ABI contains the minimal ABI for Starknet ERC20 operations
// Note: Starknet uses different function selectors than EVM
var StarknetERC20ABI = map[string]string{
	"balanceOf": "0x2e4263afad30923c891518314c3c95dbe830a16874e8abc5777a9a20b54c76e", // balanceOf(owner: felt) -> (low: felt, high: felt)
	"allowance": "0x219209519083abdd73264e9d09587b6ac54c8e5965d30f081f327dc0d3ab5d2", // allowance(owner: felt, spender: felt) -> (low: felt, high: felt)
	"approve":   "0x219209519083abdd73264e9d09587b6ac54c8e5965d30f081f327dc0d3ab5d2", // approve(spender: felt, amount: u256) -> ()
}

// ERC20Balance gets the ERC20 token balance for a given address on Starknet
func ERC20Balance(provider rpc.RpcProvider, tokenAddress string, ownerAddress string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	ownerAddrFelt, err := utils.HexToFelt(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	// Build the balanceOf function call
	balanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
		Calldata:           []*felt.Felt{ownerAddrFelt},
	}

	// Call the contract to get balance
	resp, err := provider.Call(context.Background(), balanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no response from balanceOf call")
	}

	// Convert felt response to big.Int
	balanceFelt := resp[0]
	balanceBigInt := utils.FeltToBigInt(balanceFelt)

	return balanceBigInt, nil
}

// ERC20Allowance gets the ERC20 token allowance for a given owner and spender on Starknet
func ERC20Allowance(provider rpc.RpcProvider, tokenAddress string, ownerAddress string, spenderAddress string) (*big.Int, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	ownerAddrFelt, err := utils.HexToFelt(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	spenderAddrFelt, err := utils.HexToFelt(spenderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid spender address: %w", err)
	}

	// Build the allowance function call
	allowanceCall := rpc.FunctionCall{
		ContractAddress:    tokenAddrFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("allowance"),
		Calldata:           []*felt.Felt{ownerAddrFelt, spenderAddrFelt},
	}

	// Call the contract to get allowance
	resp, err := provider.Call(context.Background(), allowanceCall, rpc.WithBlockTag("latest"))
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance: %w", err)
	}

	if len(resp) < 2 {
		return nil, fmt.Errorf("invalid allowance result length: expected 2 felts, got %d", len(resp))
	}

	// Convert two felts (low, high) back to u256
	low := utils.FeltToBigInt(resp[0])
	high := utils.FeltToBigInt(resp[1])

	// Use uint256 for better performance
	lowU := ToUint256(low)
	highU := ToUint256(high)
	
	// Combine low and high into u256: (high << 128) | low
	result := new(uint256.Int)
	result.Lsh(highU, U128BitShift)
	result.Or(result, lowU)

	return result.ToBig(), nil
}

// ERC20Approve creates an approve transaction for ERC20 tokens on Starknet
func ERC20Approve(tokenAddress string, spenderAddress string, amount *big.Int) (*rpc.InvokeFunctionCall, error) {
	// Convert addresses to felt
	tokenAddrFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token address: %w", err)
	}

	spenderAddrFelt, err := utils.HexToFelt(spenderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid spender address: %w", err)
	}

	// Convert amount to two felts (low, high) for u256 using uint256 for better performance
	amountU := ToUint256(amount)
	
	// Create mask for lower 128 bits
	lowerMask := uint256.NewInt(1)
	lowerMask.Lsh(lowerMask, U128BitShift)
	lowerMask.SubUint64(lowerMask, 1)
	
	// Extract low and high parts
	low128 := new(uint256.Int)
	low128.And(amountU, lowerMask)
	high128 := new(uint256.Int)
	high128.Rsh(amountU, U128BitShift)

	lowFelt := utils.BigIntToFelt(low128.ToBig())
	highFelt := utils.BigIntToFelt(high128.ToBig())

	// Build approve calldata: approve(spender: felt, amount: u256)
	approveCalldata := []*felt.Felt{spenderAddrFelt, lowFelt, highFelt}

	invoke := rpc.InvokeFunctionCall{
		ContractAddress: tokenAddrFelt,
		FunctionName:    "approve",
		CallData:        approveCalldata,
	}

	return &invoke, nil
}

// FormatTokenAmount formats a token amount for display (converts from wei to tokens)
func FormatTokenAmount(amount *big.Int, decimals int) string {
	if amount == nil {
		return "0"
	}

	// Convert from wei to tokens
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	tokenAmount := new(big.Float).Quo(
		new(big.Float).SetInt(amount),
		new(big.Float).SetInt(divisor))

	return tokenAmount.Text('f', 2) + " tokens"
}

// ConvertBigIntToU256Felts converts a big.Int to two felts, one for the low 128 bits and one for the high 128 bits
func ConvertBigIntToU256Felts(value *big.Int) (low *felt.Felt, high *felt.Felt) {
	// Convert to uint256 for better performance
	u := ToUint256(value)
	
	// Create mask for lower 128 bits
	lowerMask := uint256.NewInt(1)
	lowerMask.Lsh(lowerMask, U128BitShift)
	lowerMask.SubUint64(lowerMask, 1)
	
	// Extract low and high parts
	lowPart := new(uint256.Int)
	lowPart.And(u, lowerMask)
	highPart := new(uint256.Int)
	highPart.Rsh(u, U128BitShift)
	
	// Convert back to felts
	low = utils.BigIntToFelt(lowPart.ToBig())
	high = utils.BigIntToFelt(highPart.ToBig())
	return low, high
}

// ConvertSolidityOrderIDForStarknet converts a Solidity-style orderID (bytes32) into the low and high felts of a Starknet u256 orderID
// Note: Assigns the left 16 bytes to the high felt and the right 16 bytes to the low felt
func ConvertSolidityOrderIDForStarknet(orderID string) (low *felt.Felt, high *felt.Felt, err error) {
	orderBytes := utils.HexToBN(orderID).Bytes()
	if len(orderBytes) < Bytes32Length {
		pad := make([]byte, Bytes32Length-len(orderBytes))
		orderBytes = append(pad, orderBytes...)
	}

	left16 := utils.BigIntToFelt(new(big.Int).SetBytes(orderBytes[0:16]))
	right16 := utils.BigIntToFelt(new(big.Int).SetBytes(orderBytes[16:32]))

	low = right16
	high = left16

	return low, high, nil
}

// BytesToU128Felts converts bytes to u128 felts for Cairo
func BytesToU128Felts(b []byte) []*felt.Felt {
	words := make([]*felt.Felt, 0, (len(b)+Bytes16Length-1)/Bytes16Length)
	for i := 0; i < len(b); i += Bytes16Length {
		end := i + Bytes16Length
		chunk := make([]byte, Bytes16Length)
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
