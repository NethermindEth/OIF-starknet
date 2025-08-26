package hyperlane7683

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	"github.com/NethermindEth/oif-starknet/go/internal/filler"
	"github.com/NethermindEth/oif-starknet/go/internal/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var ErrIntentAlreadyFilled = fmt.Errorf("intent already filled")

type Hyperlane7683Filler struct {
	*filler.BaseFillerImpl
	client        *ethclient.Client
	clients       map[uint64]*ethclient.Client
	signers       map[uint64]*bind.TransactOpts
	hyperlaneEVM  *HyperlaneEVM
	hyperlaneStarknet *HyperlaneStarknet
	metadata      types.Hyperlane7683Metadata
}

func NewHyperlane7683Filler(client *ethclient.Client) *Hyperlane7683Filler {
	metadata := types.Hyperlane7683Metadata{
		BaseMetadata:  types.BaseMetadata{ProtocolName: "Hyperlane7683"},
		IntentSources: []types.IntentSource{},
		CustomRules:   types.CustomRules{},
	}

	allowBlockLists := types.AllowBlockLists{AllowList: []types.AllowBlockListItem{}, BlockList: []types.AllowBlockListItem{}}

	f := &Hyperlane7683Filler{
		BaseFillerImpl: filler.NewBaseFiller(allowBlockLists, metadata),
		client:         client,
		clients:        make(map[uint64]*ethclient.Client),
		signers:        make(map[uint64]*bind.TransactOpts),
		metadata:       metadata,
	}

	// Initialize protocol handlers with placeholders (will be set per-operation)
	f.hyperlaneEVM = NewHyperlaneEVM(nil, nil)
	f.hyperlaneStarknet = NewHyperlaneStarknet("")

	return f
}





func (f *Hyperlane7683Filler) ProcessIntent(ctx context.Context, args types.ParsedArgs, originChainName string, blockNumber uint64) (bool, error) {
	fmt.Printf("🔵 Processing Intent: %s-%s on chain %s (block %d)\n", f.metadata.ProtocolName, args.OrderID, originChainName, blockNumber)
	intent, err := f.PrepareIntent(ctx, args)
	if err != nil {
		return false, err
	}
	if !intent.Success {
		// Rules rejected the order - check if it's because the order was already filled
		fmt.Printf("   ⏭️  Intent rejected by rules: %s\n", intent.Error)
		
		// If the order was already filled, treat it as "successfully processed"
		// so the listener advances past this block instead of getting stuck
		if intent.Error == ErrIntentAlreadyFilled.Error() {
			fmt.Printf("   ✅ Order already processed by another filler, advancing block\n")
			return true, nil  // Successfully processed (even though we didn't fill it)
		}
		
		// For other rule rejections (insufficient balance, etc.), don't advance
		return false, nil
	}
	if err := f.Fill(ctx, args, intent.Data, originChainName, blockNumber); err != nil {
		return false, fmt.Errorf("fill execution failed: %w", err)
	}
	if err := f.SettleOrder(ctx, args, intent.Data, originChainName); err != nil {
		return false, fmt.Errorf("order settlement failed: %w", err)
	}
	return true, nil
}

func (f *Hyperlane7683Filler) Fill(ctx context.Context, args types.ParsedArgs, data types.IntentData, originChainName string, blockNumber uint64) error {
	fmt.Printf("🔵 Filling Intent: %s-%s on chain %s (block %d)\n", f.metadata.ProtocolName, args.OrderID, originChainName, blockNumber)
	fmt.Printf("   Fill Instructions: %d instructions\n", len(data.FillInstructions))
	fmt.Printf("   Max Spent: %d outputs\n", len(data.MaxSpent))

	for i, instruction := range data.FillInstructions {
		fmt.Printf("   📦 Instruction %d: Chain %s, Settler %s\n", i+1, instruction.DestinationChainID.String(), instruction.DestinationSettler.Hex())

		// Simple chain router - clean and extensible
		switch {
		case f.isStarknetChain(instruction.DestinationChainID):
			// Get Starknet RPC URL from config by finding the network with matching chain ID
			chainConfig, err := f.getNetworkConfigByChainID(instruction.DestinationChainID)
			if err != nil {
				return fmt.Errorf("Starknet network not found for chain ID %s: %w", instruction.DestinationChainID.String(), err)
			}
			f.hyperlaneStarknet = NewHyperlaneStarknet(chainConfig.RPCURL)
			
			if err := f.hyperlaneStarknet.Fill(ctx, args, originChainName); err != nil {
				return fmt.Errorf("Starknet fill failed for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			
		case f.isEVMChain(instruction.DestinationChainID):
			// Get EVM client and signer for this chain
			client, err := f.getClientForChain(instruction.DestinationChainID)
			if err != nil {
				return fmt.Errorf("failed to get client for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			signer, err := f.getSignerForChain(instruction.DestinationChainID)
			if err != nil {
				return fmt.Errorf("failed to get signer for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			f.hyperlaneEVM = NewHyperlaneEVM(client, signer)
			
			if err := f.hyperlaneEVM.Fill(ctx, args, originChainName); err != nil {
				return fmt.Errorf("EVM fill failed for chain %s: %w", instruction.DestinationChainID.String(), err)
			}
			
		default:
			return fmt.Errorf("unsupported destination chain: %s", instruction.DestinationChainID.String())
		}
	}
	
	fmt.Printf("   🎉 All fill instructions processed\n")
	return nil
}



func (f *Hyperlane7683Filler) SettleOrder(ctx context.Context, args types.ParsedArgs, data types.IntentData, originChainName string) error {
	fmt.Printf("🔵 Settling Order: %s on destination chain\n", args.OrderID)
	
	// Settlement happens on the destination chain - same as fill
	if len(data.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found for settlement")
	}
	
	instruction := data.FillInstructions[0]
	
	// Simple chain router for settlement
	switch {
	case f.isStarknetChain(instruction.DestinationChainID):
		// Get Starknet RPC URL from config by finding the network with matching chain ID
		chainConfig, err := f.getNetworkConfigByChainID(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("Starknet network not found for chain ID %s: %w", instruction.DestinationChainID.String(), err)
		}
		f.hyperlaneStarknet = NewHyperlaneStarknet(chainConfig.RPCURL)
		
		if err := f.hyperlaneStarknet.Settle(ctx, args); err != nil {
			return fmt.Errorf("Starknet settlement failed for chain %s: %w", instruction.DestinationChainID.String(), err)
		}
		
	case f.isEVMChain(instruction.DestinationChainID):
		// Get EVM client and signer for this chain
		client, err := f.getClientForChain(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get client for chain %s: %w", instruction.DestinationChainID.String(), err)
		}
		signer, err := f.getSignerForChain(instruction.DestinationChainID)
		if err != nil {
			return fmt.Errorf("failed to get signer for chain %s: %w", instruction.DestinationChainID.String(), err)
		}
		f.hyperlaneEVM = NewHyperlaneEVM(client, signer)
		
		if err := f.hyperlaneEVM.Settle(ctx, args); err != nil {
			return fmt.Errorf("EVM settlement failed for chain %s: %w", instruction.DestinationChainID.String(), err)
		}
		
	default:
		return fmt.Errorf("unsupported destination chain: %s", instruction.DestinationChainID.String())
	}
	
	fmt.Printf("✅ Settlement successful for order %s\n", args.OrderID)
	return nil
}

func (f *Hyperlane7683Filler) AddDefaultRules() {
	f.AddRule(f.enoughBalanceOnDestination)  // Pre-validate filler has enough tokens
	f.AddRule(f.filterByTokenAndAmount)      // Validate profitability and limits  
	f.AddRule(f.intentNotFilled)             // Check order hasn't been filled yet
}

// Note: Enhanced rules implementations are now in rules.go following TypeScript modular structure

func (f *Hyperlane7683Filler) intentNotFilled(args types.ParsedArgs, _ *filler.FillerContext) error {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}
	
	first := args.ResolvedOrder.FillInstructions[0]
	fmt.Printf("   🔍 intentNotFilled rule: checking destination chain %s\n", first.DestinationChainID.String())
	
	// Simple chain router for order status checking
	switch {
	case f.isStarknetChain(first.DestinationChainID):
		// For Starknet, we'll let the actual fill handle the status check
		// to keep this rule simple
		fmt.Printf("   ✅ intentNotFilled: Starknet destination, skipping status check\n")
		return nil
		
	case f.isEVMChain(first.DestinationChainID):
		// Use the EVM order status checking
		fmt.Printf("   🔍 intentNotFilled: checking EVM destination chain %s\n", first.DestinationChainID.String())
		
		client, err := f.getClientForChain(first.DestinationChainID)
		if err != nil {
			return fmt.Errorf("intentNotFilled: failed to get client for chain %s: %w", first.DestinationChainID.String(), err)
		}
		signer, err := f.getSignerForChain(first.DestinationChainID)
		if err != nil {
			return fmt.Errorf("intentNotFilled: failed to get signer for chain %s: %w", first.DestinationChainID.String(), err)
		}
		
		evmHandler := NewHyperlaneEVM(client, signer)
		status, err := evmHandler.GetOrderStatus(context.Background(), args)
		if err != nil {
			fmt.Printf("   ❌ intentNotFilled: failed to get order status: %v\n", err)
			return fmt.Errorf("intentNotFilled: failed to get order status: %w", err)
		}
		
		fmt.Printf("   🔍 intentNotFilled: order status = %s\n", status)
		if status != "UNKNOWN" {
			fmt.Printf("   ⏩ Skipping EVM fill: order status=%s (already processed)\n", status)
			return ErrIntentAlreadyFilled
		}
		
		fmt.Printf("   ✅ intentNotFilled: order not yet filled, proceeding\n")
		return nil
		
	default:
		return fmt.Errorf("intentNotFilled: unsupported chain %s", first.DestinationChainID.String())
	}
}



func (f *Hyperlane7683Filler) getClientForChain(chainID *big.Int) (*ethclient.Client, error) {
	chainIDUint := chainID.Uint64()
	if client, ok := f.clients[chainIDUint]; ok {
		return client, nil
	}
	rpcURL, err := config.GetRPCURLByChainID(chainIDUint)
	if err != nil {
		return nil, fmt.Errorf("failed to get RPC URL for chain %d: %w", chainIDUint, err)
	}
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chain %d at %s: %w", chainIDUint, rpcURL, err)
	}
	f.clients[chainIDUint] = client
	return client, nil
}

func (f *Hyperlane7683Filler) getSignerForChain(chainID *big.Int) (*bind.TransactOpts, error) {
	chainIDUint := chainID.Uint64()
	if signer, ok := f.signers[chainIDUint]; ok {
		return signer, nil
	}
	solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
	if solverPrivateKey == "" {
		return nil, fmt.Errorf("SOLVER_PRIVATE_KEY environment variable not set")
	}
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(solverPrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse solver private key: %w", err)
	}
	from := crypto.PubkeyToAddress(pk.PublicKey)
	signer, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(int64(chainIDUint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create signer with chain ID %d: %w", chainIDUint, err)
	}
	signer.From = from
	f.signers[chainIDUint] = signer
	return signer, nil
}

// Simple chain identification helpers - works with any Starknet/EVM network names
func (f *Hyperlane7683Filler) isStarknetChain(chainID *big.Int) bool {
	// Find any network with "Starknet" in the name that matches this chain ID
	for networkName, network := range config.Networks {
		if network.ChainID == chainID.Uint64() {
			// Check if network name contains "Starknet" (case insensitive)
			return strings.Contains(strings.ToLower(networkName), "starknet")
		}
	}
	return false
}

func (f *Hyperlane7683Filler) isEVMChain(chainID *big.Int) bool {
	// Find any network that matches this chain ID and is NOT a Starknet chain
	for networkName, network := range config.Networks {
		if network.ChainID == chainID.Uint64() {
			// If it's not Starknet, it's EVM
			return !strings.Contains(strings.ToLower(networkName), "starknet")
		}
	}
	return false
}

// getNetworkConfigByChainID finds the network config for a given chain ID
func (f *Hyperlane7683Filler) getNetworkConfigByChainID(chainID *big.Int) (config.NetworkConfig, error) {
	chainIDUint := chainID.Uint64()
	for _, network := range config.Networks {
		if network.ChainID == chainIDUint {
			return network, nil
		}
	}
	return config.NetworkConfig{}, fmt.Errorf("network config not found for chain ID %d", chainIDUint)
}
