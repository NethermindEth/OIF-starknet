package hyperlane7683

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// HyperlaneStarknet contains all Starknet-specific logic for the Hyperlane 7683 protocol
type HyperlaneStarknet struct {
	rpcURL string
}

// NewHyperlaneStarknet creates a new Starknet handler for Hyperlane operations
func NewHyperlaneStarknet(rpcURL string) *HyperlaneStarknet {
	return &HyperlaneStarknet{
		rpcURL: rpcURL,
	}
}

// Fill executes a fill operation on Starknet
func (h *HyperlaneStarknet) Fill(ctx context.Context, args types.ParsedArgs, originChainName string) error {
	orderID := args.OrderID
	
	// Extract origin data from the first fill instruction
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}
	
	instruction := args.ResolvedOrder.FillInstructions[0]
	originData := instruction.OriginData
	
	fmt.Printf("🔵 Starknet Fill: %s\n", orderID)
	
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
	
	// Set up ERC20 approvals before filling
	if err := h.setupApprovals(ctx, sf, args); err != nil {
		return fmt.Errorf("failed to setup approvals: %w", err)
	}
	
	// Execute the fill
	fmt.Printf("   🚀 Proceeding with Starknet fill after approvals\n")
	return sf.Fill(ctx, orderID, originData)
}

// Settle executes settlement on Starknet
func (h *HyperlaneStarknet) Settle(ctx context.Context, args types.ParsedArgs) error {
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
	ethAddress := "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	settlerOps := NewStarknetOps(sf)
	if err := settlerOps.EnsureApproval(ctx, ethAddress, gasPayment); err != nil {
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
	
	hyperlaneAddr, err := utils.HexToFelt(hyperlaneAddressHex)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("invalid Hyperlane address: %w", err)
	}
	
	// Create RPC provider
	provider, err := rpc.NewProvider(h.rpcURL)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to create provider: %w", err)
	}
	
	processed, status, err := h.isOrderProcessed(ctx, provider, hyperlaneAddr, orderID)
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
	// Use the original 32-byte address from the event if available
	if len(args.ResolvedOrder.FillInstructions) > 0 {
		instruction := args.ResolvedOrder.FillInstructions[0]
		if h.isStarknetChain(instruction.DestinationChainID) {
			// Use the original 32-byte address from the event
			hyperlaneAddressHex := fmt.Sprintf("0x%x", instruction.OriginalDestinationSettler)
			fmt.Printf("   🎯 Using original 32-byte Hyperlane address from event: %s\n", hyperlaneAddressHex)
			return hyperlaneAddressHex, nil
		}
	}
	
	// Fallback to deployment state
	ds, err := deployer.GetDeploymentState()
	if err != nil {
		return "", fmt.Errorf("failed to load deployment state: %w", err)
	}
	
	if networkState, exists := ds.Networks["Starknet Sepolia"]; exists && networkState.HyperlaneAddress != "" {
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
	
	ops := NewStarknetOps(sf)
	
	for i, maxSpent := range args.ResolvedOrder.MaxSpent {
		// Skip native ETH (zero address)
		if maxSpent.Token == (common.Address{}) {
			fmt.Printf("   ⏭️  Skipping approval for native ETH (index %d)\n", i)
			continue
		}
		
		fmt.Printf("   📊 MaxSpent[%d] Token: %s, Amount: %s, Recipient: %s, ChainID: %s\n", 
			i, maxSpent.Token.Hex(), maxSpent.Amount.String(), maxSpent.Recipient.Hex(), maxSpent.ChainID.String())
		
		// Convert token address to Starknet format
		tokenAddressHex := h.getTokenAddress(maxSpent)
		
		fmt.Printf("   🎯 TOKEN[%d] APPROVAL CALL:\n", i)
		fmt.Printf("     • Token address: %s\n", tokenAddressHex)
		fmt.Printf("     • Amount to approve: %s\n", maxSpent.Amount.String())
		
		if err := ops.EnsureApproval(ctx, tokenAddressHex, maxSpent.Amount); err != nil {
			return fmt.Errorf("starknet approval failed for token %s: %w", tokenAddressHex, err)
		}
		
		fmt.Printf("   ✅ TOKEN[%d] approval completed\n", i)
	}
	
	return nil
}

func (h *HyperlaneStarknet) getTokenAddress(maxSpent types.Output) string {
	// For Starknet destinations, use the original 32-byte address if available
	if h.isStarknetChain(maxSpent.ChainID) {
		// Check if OriginalToken is not zero array
		hasOriginalToken := false
		for _, b := range maxSpent.OriginalToken {
			if b != 0 {
				hasOriginalToken = true
				break
			}
		}
		
		if hasOriginalToken {
			tokenAddressHex := fmt.Sprintf("0x%x", maxSpent.OriginalToken)
			fmt.Printf("   🎯 Using original Starknet token address: %s\n", tokenAddressHex)
			return tokenAddressHex
		}
	}
	
	// Fallback to converted address  
	tokenAddressHex := maxSpent.Token.Hex()
	fmt.Printf("   ⚠️  Using converted token address (fallback): %s\n", tokenAddressHex)
	return tokenAddressHex
}

func (h *HyperlaneStarknet) isOrderProcessed(ctx context.Context, provider *rpc.Provider, hyperlaneAddr *felt.Felt, orderID string) (bool, string, error) {
	// Convert orderID to u256 (low, high) for Starknet
	idBytes := utils.HexToBN(orderID).Bytes()
	if len(idBytes) < 32 {
		idBytes = append(make([]byte, 32-len(idBytes)), idBytes...)
	}
	
	// Apply u256_reverse_endian as expected by Cairo OrderEncoder
	rev := func(in []byte) []byte {
		out := make([]byte, len(in))
		for i := 0; i < len(in); i++ {
			out[i] = in[len(in)-1-i]
		}
		return out
	}
	
	low := utils.BigIntToFelt(new(big.Int).SetBytes(rev(idBytes[0:16])))
	high := utils.BigIntToFelt(new(big.Int).SetBytes(rev(idBytes[16:32])))
	
	call := rpc.FunctionCall{
		ContractAddress:    hyperlaneAddr,
		EntryPointSelector: utils.GetSelectorFromNameFelt("order_status"),
		Calldata:           []*felt.Felt{low, high},
	}
	
	resp, err := provider.Call(ctx, call, rpc.WithBlockTag("latest"))
	if err != nil || len(resp) == 0 {
		return false, "", fmt.Errorf("failed to call order_status: %w", err)
	}
	
	status := resp[0].String()
	return status != "0x0" && status != "0", status, nil
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
	orderHash := crypto.Keccak256(originData)
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
