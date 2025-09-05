package main

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
	"github.com/joho/godotenv"

	"github.com/NethermindEth/oif-starknet/go/solvercore/config"
)

// Registers EVM routers and sets destination gas configs on Starknet Hyperlane using owner account

func main() {
	_ = godotenv.Load()

	// Initialize networks from centralized config after .env is loaded
	config.InitializeNetworks()

	networkName := "Starknet"
	netCfg, err := config.GetNetworkConfig(networkName)
	if err != nil {
		panic(err)
	}

	// Owner creds
	ownerAddr := mustEnv("STARKNET_DEPLOYER_ADDRESS")
	ownerPub := mustEnv("STARKNET_DEPLOYER_PUBLIC_KEY")
	ownerPriv := mustEnv("STARKNET_DEPLOYER_PRIVATE_KEY")

	// Starknet provider/account
	provider, err := rpc.NewProvider(netCfg.RPCURL)
	if err != nil {
		panic(err)
	}
	ownerAddrF, _ := utils.HexToFelt(ownerAddr)
	ks := account.NewMemKeystore()
	privBI, ok := new(big.Int).SetString(ownerPriv, 0)
	if !ok {
		panic("invalid STARKNET_DEPLOYER_PRIVATE_KEY")
	}
	ks.Put(ownerPub, privBI)
	acct, err := account.NewAccount(provider, ownerAddrF, ownerPub, ks, account.CairoV2)
	if err != nil {
		panic(err)
	}

	// Load Starknet Hyperlane address from .env
	starknetHyperlaneAddr := os.Getenv("STARKNET_HYPERLANE_ADDRESS")
	if starknetHyperlaneAddr == "" {
		panic("STARKNET_HYPERLANE_ADDRESS not found in .env")
	}
	hlAddrF, _ := utils.HexToFelt(starknetHyperlaneAddr)

	// Build arrays of ALL destinations and routers (including Starknet itself)
	type routerEntry struct {
		domain uint32
		b32    [32]byte
		name   string
	}
	var entries []routerEntry

	// Add ALL networks including Starknet itself
	for name, cfg := range config.Networks {
		if name == networkName {
			// Add Starknet itself - it needs to know about itself as a destination
			starknetB32 := hexToBytes32(starknetHyperlaneAddr)
			entries = append(entries, routerEntry{
				domain: uint32(cfg.HyperlaneDomain),
				b32:    starknetB32,
				name:   name,
			})
			fmt.Printf("   🏠 Starknet self-registration: domain %d -> router %s\n", cfg.HyperlaneDomain, starknetHyperlaneAddr)
		} else {
			// Add EVM networks
			entries = append(entries, routerEntry{
				domain: uint32(cfg.HyperlaneDomain),
				b32:    evmAddrToBytes32(cfg.HyperlaneAddress.Bytes()),
				name:   name,
			})
			fmt.Printf("   🔗 EVM %s: domain %d -> router %s\n", name, cfg.HyperlaneDomain, cfg.HyperlaneAddress.Hex())
		}
	}

	// Encode Arrays per Cairo ABI: len + elements
	calldata := make([]*felt.Felt, 0, 1+len(entries)+1+len(entries)*2)
	// destinations: Array<u32>
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(entries))))
	for _, e := range entries {
		calldata = append(calldata, utils.Uint64ToFelt(uint64(e.domain)))
	}
	// routers: Array<u256> (each as low, high felts)
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(entries))))
	for _, e := range entries {
		low, high := bytes32ToU256Felts(e.b32)
		calldata = append(calldata, low, high)
	}

	// enroll_remote_routers(uint32[] destinations, u256[] routers)
	enrollCall := rpc.InvokeFunctionCall{ContractAddress: hlAddrF, FunctionName: "enroll_remote_routers", CallData: calldata}
	tx1, err := acct.BuildAndSendInvokeTxn(context.Background(), []rpc.InvokeFunctionCall{enrollCall}, nil)
	if err != nil {
		panic(fmt.Errorf("enroll_remote_routers failed: %w", err))
	}
	fmt.Printf("   ⛽ enroll_remote_routers tx: %s\n", tx1.Hash.String())

	// Wait for router enrollment to complete before setting gas
	_, err = acct.WaitForTransactionReceipt(context.Background(), tx1.Hash, time.Second)
	if err != nil {
		panic(fmt.Errorf("enroll_remote_routers wait failed: %w", err))
	}
	fmt.Printf("   ✅ Router enrollment confirmed\n")

	// Now set destination gas for all domains in a single batch call
	fmt.Printf("   ⚡ Setting destination gas configs (batch mode)...\n")

	// Gas values for different networks (from real event analysis)
	gasEVM := big.NewInt(64000)       // 64,000 wei for EVM networks
	gasStarknet := big.NewInt(100000) // 100,000 wei for Starknet

	// Build gas_configs array: Option<Array<GasRouterConfig>>
	// Each GasRouterConfig contains: { destination: u32, gas: u256 }

	// Array length
	gasConfigsCalldata := []*felt.Felt{utils.Uint64ToFelt(uint64(len(entries)))} // gas_configs array length

	// Add each GasRouterConfig struct to the array
	for _, entry := range entries {
		// Use different gas amounts based on network type
		var gasAmount *big.Int
		if entry.name == "Starknet" {
			gasAmount = gasStarknet
		} else {
			gasAmount = gasEVM
		}

		fmt.Printf("   ⚡ %s (domain %d): %s units\n", entry.name, entry.domain, gasAmount.String())

		// Convert gas amount to u256 (low, high felts)
		gasLow, gasHigh := bigIntToU256Felts(gasAmount)

		// GasRouterConfig struct: { destination: u32, gas: u256 }
		gasConfigsCalldata = append(gasConfigsCalldata,
			utils.Uint64ToFelt(uint64(entry.domain)), // destination
			gasLow,                                   // gas amount low
			gasHigh,                                  // gas amount high
		)
	}

	// For the function signature:
	// set_destination_gas(gas_configs: Option<Array<GasRouterConfig>>, domain: Option<u32>, gas: Option<u256>)
	// We pass: Some(gas_configs), None, None

	finalCalldata := []*felt.Felt{
		// gas_configs: Some(Array<GasRouterConfig>) - we pass 0 to indicate Some variant
		utils.Uint64ToFelt(0), // Some variant
	}
	finalCalldata = append(finalCalldata, gasConfigsCalldata...) // append the array data

	// domain: None - we pass 1 to indicate None variant
	finalCalldata = append(finalCalldata, utils.Uint64ToFelt(1))

	// gas: None - we pass 1 to indicate None variant
	finalCalldata = append(finalCalldata, utils.Uint64ToFelt(1))

	gasCall := rpc.InvokeFunctionCall{
		ContractAddress: hlAddrF,
		FunctionName:    "set_destination_gas",
		CallData:        finalCalldata,
	}

	tx2, err := acct.BuildAndSendInvokeTxn(context.Background(), []rpc.InvokeFunctionCall{gasCall}, nil)
	if err != nil {
		panic(fmt.Errorf("batch set_destination_gas failed: %w", err))
	}
	fmt.Printf("   ⛽ Batch set_destination_gas tx: %s\n", tx2.Hash.String())

	// Wait for gas config to complete
	_, err = acct.WaitForTransactionReceipt(context.Background(), tx2.Hash, time.Second)
	if err != nil {
		panic(fmt.Errorf("batch set_destination_gas wait failed: %w", err))
	}
	fmt.Printf("   ✅ All %d destination gas configs set successfully in single transaction\n", len(entries))
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}

func evmAddrToBytes32(addr20 []byte) (out [32]byte) { copy(out[12:], addr20); return }

func hexToBytes32(hexStr string) (out [32]byte) {
	// Remove 0x prefix if present
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// Decode hex string to bytes
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return out
	}

	// Copy to 32-byte array (right-aligned for addresses)
	if len(bytes) <= 32 {
		copy(out[32-len(bytes):], bytes)
	} else {
		copy(out[:], bytes[len(bytes)-32:])
	}
	return out
}

func bytes32ToU256Felts(b32 [32]byte) (*felt.Felt, *felt.Felt) {
	// Split into high(16 bytes) and low(16 bytes)
	high := new(big.Int).SetBytes(b32[0:16])
	low := new(big.Int).SetBytes(b32[16:32])
	return utils.BigIntToFelt(low), utils.BigIntToFelt(high)
}

func bigIntToU256Felts(num *big.Int) (*felt.Felt, *felt.Felt) {
	// Convert a big.Int to u256 representation (low and high felts)
	// u256 is represented as two felt.Felt values where:
	// - low contains the first 128 bits
	// - high contains the remaining 128 bits

	// Create a mask for 128 bits (2^128 - 1)
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

	// Extract low part (first 128 bits)
	lowBigInt := new(big.Int).And(num, mask)
	low := utils.BigIntToFelt(lowBigInt)

	// Extract high part (remaining bits)
	highBigInt := new(big.Int).Rsh(num, 128)
	high := utils.BigIntToFelt(highBigInt)

	return low, high
}
