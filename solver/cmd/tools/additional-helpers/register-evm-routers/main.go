package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
)

const (
	// Wait time for transaction receipt
	receiptWaitMs = 500
)

// Minimal tool to impersonate owner on each EVM fork and call enrollRemoteRouters and setDestinationGas

func main() {
	// Load .env from likely locations
	_ = godotenv.Load(".env")
	_ = godotenv.Overload("../.env")
	_ = godotenv.Overload("../../.env")

	ownerHex := os.Getenv("EVM_HYPERLANE_OWNER")
	if ownerHex == "" {
		log.Fatal("EVM_HYPERLANE_OWNER env var (owner/admin to impersonate) is required")
	}

	// Initialize networks from config after .env is loaded
	config.InitializeNetworks()

	// Get Starknet Hyperlane address from config (.env)
	starknetHyperlaneAddr := os.Getenv("STARKNET_HYPERLANE_ADDRESS")
	if starknetHyperlaneAddr == "" {
		log.Fatalf("STARKNET_HYPERLANE_ADDRESS not found in .env file")
	}

	// Precompute destinations and routers per network
	networkNames := config.GetNetworkNames()
	for _, networkName := range networkNames {
		if networkName == "Starknet" {
			continue
		}

		netCfg := config.Networks[networkName]
		fmt.Printf("\n🔧 Registering routers on %s (%s)\n", netCfg.Name, netCfg.RPCURL)

		// Build arrays: destinations (domains) and routers (bytes32)
		var destDomains []uint32
		var routerBytes [][32]byte

		for _, otherName := range networkNames {
			if otherName == networkName {
				continue
			}
			dom, err := config.GetHyperlaneDomain(otherName)
			if err != nil {
				log.Fatalf("failed to get domain for %s: %v", otherName, err)
			}
			destDomains = append(destDomains, uint32(dom))

			if otherName == "Starknet" {
				// Starknet router as raw 32-byte felt
				rb := hexToBytes32(starknetHyperlaneAddr)
				routerBytes = append(routerBytes, rb)
				fmt.Printf("   🌉 Starknet domain %d -> router %s (0x%s)\n", dom, starknetHyperlaneAddr, hex.EncodeToString(rb[:]))
			} else {
				// EVM router is 20-byte address left-padded to 32
				evmAddr := config.Networks[otherName].HyperlaneAddress
				var b32 [32]byte
				copy(b32[12:], evmAddr.Bytes())
				routerBytes = append(routerBytes, b32)
				fmt.Printf("   🔗 EVM domain %d -> router %s (0x%s)\n", dom, evmAddr.Hex(), hex.EncodeToString(b32[:]))
			}
		}

		fmt.Printf("   📊 Total destinations: %d, Total routers: %d\n", len(destDomains), len(routerBytes))

		// Gas configs: much higher gas for cross-chain operations
		// Previous: 0xfa00 = 64,000 wei (too low!)
		// New: 0x186a0 = 100,000 wei (still conservative but much better)
		gasDefault := new(big.Int)
		gasDefault.SetString("0x186a0", 0) // 100,000 wei

		// Special handling for Starknet domain - it needs more gas due to complex operations
		starknetDomain := uint32(config.Networks["Starknet"].HyperlaneDomain) // Starknet domain ID from config
		starknetGas := new(big.Int)
		starknetGas.SetString("0x3d090", 0) // 250,000 wei for Starknet operations

		var gasDomains []uint32
		var gasValues []*big.Int
		for _, other := range destDomains {
			gasDomains = append(gasDomains, other)

			// Use higher gas for Starknet domain
			if other == starknetDomain {
				gasValues = append(gasValues, new(big.Int).Set(starknetGas))
				fmt.Printf("   ⚡ Starknet domain %d: gas = %s wei (0x%s)\n", other, starknetGas.String(), starknetGas.Text(16))
			} else {
				gasValues = append(gasValues, new(big.Int).Set(gasDefault))
				fmt.Printf("   ⚡ Domain %d: gas = %s wei (0x%s)\n", other, gasDefault.String(), gasDefault.Text(16))
			}
		}

		// Connect RPC
		rpcClient, err := rpc.Dial(netCfg.RPCURL)
		if err != nil {
			log.Fatalf("failed to dial RPC %s: %v", netCfg.RPCURL, err)
		}

		owner := common.HexToAddress(ownerHex)
		// Impersonate owner (anvil returns null on success)
		var dummy any
		if err := rpcClient.Call(&dummy, "anvil_impersonateAccount", owner.Hex()); err != nil {
			rpcClient.Close()
			log.Fatalf("failed to impersonate %s on %s: %v", owner.Hex(), networkName, err)
		}
		fmt.Printf("   👤 Impersonating owner %s\n", owner.Hex())

		// Ensure owner has large ETH balance and verify
		// Set ~1e27 wei (~1e9 ETH)
		rich := "0x33B2E3C9FD0803CE8000000"
		if err := rpcClient.Call(&dummy, "anvil_setBalance", owner.Hex(), rich); err != nil {
			log.Fatalf("failed to set balance for %s on %s: %v", owner.Hex(), networkName, err)
		}
		if !hasBalance(rpcClient, owner) {
			// Retry once
			_ = rpcClient.Call(&dummy, "anvil_setBalance", owner.Hex(), rich)
			if !hasBalance(rpcClient, owner) {
				log.Fatalf("owner %s still unfunded on %s", owner.Hex(), networkName)
			}
		}

		// Prepare ABI encodings
		evmAbiJSON := `[
            {"type":"function","name":"enrollRemoteRouters","inputs":[{"type":"uint32[]","name":"_destinations"},{"type":"bytes32[]","name":"_routers"}],"outputs":[],"stateMutability":"nonpayable"},
            {"type":"function","name":"setDestinationGas","inputs":[{"type":"tuple[]","name":"gasConfigs","components":[{"type":"uint32","name":"destination"},{"type":"uint256","name":"gas"}]}],"outputs":[],"stateMutability":"nonpayable"}
        ]`
		evmAbi, err := abi.JSON(strings.NewReader(evmAbiJSON))
		if err != nil {
			log.Fatalf("failed to parse ABI: %v", err)
		}

		// Encode enrollRemoteRouters (bytes32[] must be [][32]byte)
		enrollData, err := evmAbi.Pack("enrollRemoteRouters", destDomains, routerBytes)
		if err != nil {
			log.Fatalf("pack enrollRemoteRouters failed: %v", err)
		}

		// Build gas tuple array
		type gasTuple struct {
			Destination uint32
			Gas         *big.Int
		}
		gasArr := make([]gasTuple, 0, len(gasDomains))
		for i := range gasDomains {
			gasArr = append(gasArr, gasTuple{Destination: gasDomains[i], Gas: gasValues[i]})
		}
		gasData, err := evmAbi.Pack("setDestinationGas", gasArr)
		if err != nil {
			log.Fatalf("pack setDestinationGas failed: %v", err)
		}

		// Send transactions via eth_sendTransaction
		hlAddr := netCfg.HyperlaneAddress
		if err := sendImpersonatedTx(rpcClient, owner, hlAddr, enrollData); err != nil {
			log.Fatalf("enrollRemoteRouters failed: %v", err)
		}
		if err := sendImpersonatedTx(rpcClient, owner, hlAddr, gasData); err != nil {
			log.Fatalf("setDestinationGas failed: %v", err)
		}

		// Stop impersonation
		_ = rpcClient.Call(&dummy, "anvil_stopImpersonatingAccount", owner.Hex())
		fmt.Printf("   ✅ Routers/gas registered on %s\n", networkName)
		
		// Close the RPC connection
		rpcClient.Close()
	}

	fmt.Printf("\n✅ EVM router registration complete\n")
}

func sendImpersonatedTx(c *rpc.Client, from, to common.Address, data []byte) error {
	params := map[string]interface{}{
		"from": from.Hex(),
		"to":   to.Hex(),
		"data": "0x" + hex.EncodeToString(data),
	}
	var txHash common.Hash
	if err := c.Call(&txHash, "eth_sendTransaction", params); err != nil {
		return err
	}
	for i := 0; i < 60; i++ {
		var raw json.RawMessage
		if err := c.Call(&raw, "eth_getTransactionReceipt", txHash.Hex()); err == nil && len(raw) > 0 {
			fmt.Printf("   ⛽ Tx mined: %s\n", txHash.Hex())
			var rec map[string]any
			if err := json.Unmarshal(raw, &rec); err == nil {
				if status, ok := rec["status"].(string); ok {
					if status == "0x1" || status == "0x01" {
						return nil
					}
					return fmt.Errorf("transaction reverted (status=%s)", status)
				}
			}
			return nil
		}
		time.Sleep(receiptWaitMs * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting receipt for %s", to.Hex())
}

func hexToBytes32(hexStr string) (out [32]byte) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	b, _ := hex.DecodeString(hexStr)
	if len(b) >= 32 {
		copy(out[:], b[len(b)-32:])
	} else {
		copy(out[32-len(b):], b)
	}
	return
}

func hasBalance(c *rpc.Client, addr common.Address) bool {
	var balHex string
	if err := c.Call(&balHex, "eth_getBalance", addr.Hex(), "latest"); err != nil {
		return false
	}
	if len(balHex) <= 2 {
		return false
	}
	// consider any non-zero balance sufficient
	return balHex != "0x0" && balHex != "0x"
}
