package hyperlane7683

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/config"
	contracts "github.com/NethermindEth/oif-starknet/go/internal/contracts"
	"github.com/NethermindEth/oif-starknet/go/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// HyperlaneEVM contains all EVM-specific logic for the Hyperlane 7683 protocol
type HyperlaneEVM struct {
	client *ethclient.Client
	signer *bind.TransactOpts
}

// NewHyperlaneEVM creates a new EVM handler for Hyperlane operations
func NewHyperlaneEVM(client *ethclient.Client, signer *bind.TransactOpts) *HyperlaneEVM {
	return &HyperlaneEVM{
		client: client,
		signer: signer,
	}
}

// Fill executes a fill operation on an EVM chain
func (h *HyperlaneEVM) Fill(ctx context.Context, args types.ParsedArgs, originChainName string) error {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}
	
	instruction := args.ResolvedOrder.FillInstructions[0]
	fmt.Printf("🔵 EVM Fill: %s on chain %s\n", args.OrderID, instruction.DestinationChainID.String())
	
	// Derive orderId from keccak(origin_data) for EVM legs
	var orderIdArr [32]byte
	orderHash := crypto.Keccak256(instruction.OriginData)
	copy(orderIdArr[:], orderHash)

	// Pre-check: skip if order status != 0 (already processed)
	if processed, err := h.isOrderAlreadyProcessed(ctx, orderIdArr, instruction.DestinationSettler); err == nil && processed {
		fmt.Printf("   ⏭️  Skipping EVM fill: order already processed\n")
		return nil
	}

	// Get the contract instance
	contract, err := contracts.NewHyperlane7683(instruction.DestinationSettler, h.client)
	if err != nil {
		return fmt.Errorf("failed to bind contract at %s: %w", instruction.DestinationSettler.Hex(), err)
	}

	// Handle max spent approvals if needed  
	if len(args.ResolvedOrder.MaxSpent) > 0 {
		maxSpent := args.ResolvedOrder.MaxSpent[0]
		if maxSpent.Token != (common.Address{}) {
			if err := h.ensureERC20Approval(ctx, maxSpent.Token, instruction.DestinationSettler, maxSpent.Amount); err != nil {
				return fmt.Errorf("approval failed for token %s: %w", maxSpent.Token.Hex(), err)
			}
		}
	}

	// Prepare filler data (use empty bytes like original working code)
	var fillerDataBytes []byte

	// Set native token value if needed
	originalValue := h.signer.Value
	if len(args.ResolvedOrder.MaxSpent) > 0 && args.ResolvedOrder.MaxSpent[0].Token == (common.Address{}) {
		h.signer.Value = new(big.Int).Set(args.ResolvedOrder.MaxSpent[0].Amount)
	}
	defer func() { h.signer.Value = originalValue }()

	fmt.Printf("   🔄 Executing fill call to contract %s\n", instruction.DestinationSettler.Hex())
	
	// Execute the fill transaction
	tx, err := contract.Fill(h.signer, orderIdArr, instruction.OriginData, fillerDataBytes)
	if err != nil {
		return fmt.Errorf("fill transaction failed: %w", err)
	}

	fmt.Printf("   🚀 Fill transaction sent: %s\n", tx.Hash().Hex())
	
	// Wait for confirmation
	receipt, err := bind.WaitMined(ctx, h.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for fill confirmation: %w", err)
	}

	if receipt.Status == 1 {
		fmt.Printf("   ✅ EVM Fill successful! Gas used: %d\n", receipt.GasUsed)
		return nil
	} else {
		return fmt.Errorf("fill transaction failed with status: %d", receipt.Status)
	}
}

// Settle executes settlement on an EVM chain
func (h *HyperlaneEVM) Settle(ctx context.Context, args types.ParsedArgs) error {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}
	
	instruction := args.ResolvedOrder.FillInstructions[0]
	destinationSettler := instruction.DestinationSettler
	
	fmt.Printf("🔵 EVM Settle: %s on settler %s\n", args.OrderID, destinationSettler.Hex())
	
	// Prepare order ID
	var orderIdArr [32]byte
	orderIDBytes := common.FromHex(args.OrderID)
	copy(orderIdArr[:], orderIDBytes)
	
	// If order ID is zero, fall back to keccak(origin_data)
	isZero := true
	for _, b := range orderIdArr {
		if b != 0 {
			isZero = false
			break
		}
	}
	if isZero {
		fallbackHash := crypto.Keccak256(instruction.OriginData)
		copy(orderIdArr[:], fallbackHash)
		fmt.Printf("   ℹ️  Using fallback orderId (keccak(origin_data)): 0x%s\n", hex.EncodeToString(orderIdArr[:]))
	}

	// Bind to the destination settler contract
	dest, err := contracts.NewHyperlane7683(destinationSettler, h.client)
	if err != nil {
		return fmt.Errorf("failed to bind to destination settler contract: %w", err)
	}

	// Get gas payment (protocol fee) that must be sent with settlement
	// Use the actual origin domain from the resolved order (critical for correct routing!)
	originDomain, err := getOriginDomainFromArgs(args)
	if err != nil {
		return fmt.Errorf("failed to get origin domain: %w", err)
	}
	fmt.Printf("   🔍 Using origin domain for gas quote: %d\n", originDomain)
	gasPayment, err := dest.QuoteGasPayment(&bind.CallOpts{Context: ctx}, originDomain)
	if err != nil {
		return fmt.Errorf("quoteGasPayment failed on %s: %w", destinationSettler.Hex(), err)
	}
	fmt.Printf("   💰 Gas payment quoted: %s wei\n", gasPayment.String())

	// Set value for settle transaction
	originalValue := h.signer.Value
	h.signer.Value = new(big.Int).Set(gasPayment)
	defer func() { h.signer.Value = originalValue }()

	// Pre-settle check: ensure order is FILLED
	if err := h.verifyOrderStatus(ctx, orderIdArr, destinationSettler, "FILLED"); err != nil {
		return fmt.Errorf("pre-settle check failed: %w", err)
	}

	// Prepare order IDs array (contract expects array)
	orderIDs := make([][32]byte, 1)
	orderIDs[0] = orderIdArr

	// Set gas price if not already set
	if h.signer.GasPrice == nil || h.signer.GasPrice.Sign() == 0 {
		if suggested, gerr := h.client.SuggestGasPrice(ctx); gerr == nil {
			h.signer.GasPrice = suggested
		}
	}

	fmt.Printf("   🚀 Sending settle transaction with value %s wei...\n", h.signer.Value.String())
	
	// Execute the settle transaction with protocol fee payment
	tx, err := dest.Settle(h.signer, orderIDs)
	if err != nil {
		return fmt.Errorf("settle tx failed on %s: %w", destinationSettler.Hex(), err)
	}
	fmt.Printf("   📊 Settle transaction sent: %s\n", tx.Hash().Hex())
	
	receipt, err := bind.WaitMined(ctx, h.client, tx)
	if err != nil {
		return fmt.Errorf("waiting settle failed on %s: %w", destinationSettler.Hex(), err)
	}
	
	if receipt.Status == 0 {
		return fmt.Errorf("settle transaction failed on %s at block %d", destinationSettler.Hex(), receipt.BlockNumber)
	}
	
	fmt.Printf("   ✅ Settle transaction confirmed at block %d (gasUsed=%d)\n", receipt.BlockNumber, receipt.GasUsed)
	return nil
}

// GetOrderStatus returns the current status of an order
func (h *HyperlaneEVM) GetOrderStatus(ctx context.Context, args types.ParsedArgs) (string, error) {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return "UNKNOWN", fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]
	
	// Derive orderId from keccak(origin_data)
	var orderIdArr [32]byte
	orderHash := crypto.Keccak256(instruction.OriginData)
	copy(orderIdArr[:], orderHash)

	// Check order status 
	orderStatusABI := `[{"type":"function","name":"orderStatus","inputs":[{"type":"bytes32","name":"orderId"}],"outputs":[{"type":"bytes32","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(orderStatusABI))
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to parse orderStatus ABI: %w", err)
	}

	callData, err := parsedABI.Pack("orderStatus", orderIdArr)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to pack orderStatus: %w", err)
	}

	dummyFrom := common.HexToAddress("0x1000000000000000000000000000000000000000")
	res, err := h.client.CallContract(ctx, ethereum.CallMsg{From: dummyFrom, To: &instruction.DestinationSettler, Data: callData}, nil)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("orderStatus call failed: %w", err)
	}

	if len(res) < 32 {
		return "UNKNOWN", fmt.Errorf("invalid orderStatus result length: %d", len(res))
	}

	statusHash := common.BytesToHash(res[:32])
	return h.interpretStatusHash(ctx, statusHash, instruction.DestinationSettler), nil
}

// Helper methods
func (h *HyperlaneEVM) isOrderAlreadyProcessed(ctx context.Context, orderIdArr [32]byte, settlerAddr common.Address) (bool, error) {
	orderStatusABI := `[{"type":"function","name":"orderStatus","inputs":[{"type":"bytes32","name":"orderId"}],"outputs":[{"type":"bytes32","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(orderStatusABI))
	if err != nil {
		return false, fmt.Errorf("failed to parse orderStatus ABI: %w", err)
	}
	
	callData, err := parsedABI.Pack("orderStatus", orderIdArr)
	if err != nil {
		return false, fmt.Errorf("failed to pack orderStatus: %w", err)
	}
	
	res, err := h.client.CallContract(ctx, ethereum.CallMsg{From: h.signer.From, To: &settlerAddr, Data: callData}, nil)
	if err != nil {
		return false, fmt.Errorf("orderStatus call failed: %w", err)
	}
	
	if len(res) >= 32 {
		status := common.BytesToHash(res[:32])
		return status != (common.Hash{}), nil
	}
	
	return false, nil
}

func (h *HyperlaneEVM) verifyOrderStatus(ctx context.Context, orderIdArr [32]byte, settlerAddr common.Address, expectedStatus string) error {
	orderStatusABI := `[{"type":"function","name":"orderStatus","inputs":[{"type":"bytes32","name":"orderId"}],"outputs":[{"type":"bytes32","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(orderStatusABI))
	if err != nil {
		return fmt.Errorf("failed to parse orderStatus ABI: %w", err)
	}
	
	callData, err := parsedABI.Pack("orderStatus", orderIdArr)
	if err != nil {
		return fmt.Errorf("failed to pack orderStatus: %w", err)
	}
	
	res, err := h.client.CallContract(ctx, ethereum.CallMsg{From: h.signer.From, To: &settlerAddr, Data: callData}, nil)
	if err != nil {
		return fmt.Errorf("orderStatus call failed: %w", err)
	}
	
	if len(res) < 32 {
		return fmt.Errorf("invalid orderStatus result length: %d", len(res))
	}
	
	statusHash := common.BytesToHash(res[:32])
	actualStatus := h.interpretStatusHash(ctx, statusHash, settlerAddr)
	
	if actualStatus != expectedStatus {
		return fmt.Errorf("order status mismatch: expected %s, got %s", expectedStatus, actualStatus)
	}
	
	return nil
}

func (h *HyperlaneEVM) interpretStatusHash(ctx context.Context, statusHash common.Hash, contractAddr common.Address) string {
	if statusHash == (common.Hash{}) {
		return "UNKNOWN"
	}

	// Try to read constants from contract for comparison
	if dest, err := contracts.NewHyperlane7683(contractAddr, h.client); err == nil {
		if filledConst, err := dest.FILLED(&bind.CallOpts{Context: ctx}); err == nil {
			filledHash := common.BytesToHash(filledConst[:])
			if statusHash == filledHash {
				return "FILLED"
			}
		}
	}

	// Check hardcoded SETTLED constant
	settledHash := common.HexToHash("0x534554544c454400000000000000000000000000000000000000000000000000")
	if statusHash == settledHash {
		return "SETTLED"
	}

	return statusHash.Hex()
}

func (h *HyperlaneEVM) ensureERC20Approval(ctx context.Context, tokenAddr, spender common.Address, amount *big.Int) error {
	// Check current allowance
	allowanceABI := `[{"type":"function","name":"allowance","inputs":[{"type":"address","name":"owner"},{"type":"address","name":"spender"}],"outputs":[{"type":"uint256","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(allowanceABI))
	if err != nil {
		return fmt.Errorf("failed to parse allowance ABI: %w", err)
	}
	
	callData, err := parsedABI.Pack("allowance", h.signer.From, spender)
	if err != nil {
		return fmt.Errorf("failed to pack allowance call: %w", err)
	}
	
	result, err := h.client.CallContract(ctx, ethereum.CallMsg{To: &tokenAddr, Data: callData}, nil)
	if err != nil {
		return fmt.Errorf("allowance call failed: %w", err)
	}
	
	if len(result) < 32 {
		return fmt.Errorf("invalid allowance result length: %d", len(result))
	}
	
	currentAllowance := new(big.Int).SetBytes(result)
	
	// If allowance is sufficient, no approval needed
	if currentAllowance.Cmp(amount) >= 0 {
		fmt.Printf("   ✅ Sufficient allowance: have %s, need %s\n", currentAllowance.String(), amount.String())
		return nil
	}
	
	fmt.Printf("   📝 Approving %s tokens for %s\n", amount.String(), spender.Hex())
	
	// Approve exact amount needed
	approveABI := `[{"type":"function","name":"approve","inputs":[{"type":"address","name":"spender"},{"type":"uint256","name":"amount"}],"outputs":[{"type":"bool","name":""}],"stateMutability":"nonpayable"}]`
	parsedApproveABI, err := abi.JSON(strings.NewReader(approveABI))
	if err != nil {
		return fmt.Errorf("failed to parse approve ABI: %w", err)
	}
	
	approveData, err := parsedApproveABI.Pack("approve", spender, amount)
	if err != nil {
		return fmt.Errorf("failed to pack approve call: %w", err)
	}
	
	// Get current nonce
	nonce, err := h.client.PendingNonceAt(ctx, h.signer.From)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}
	
	// Get gas price
	gasPrice, err := h.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}
	
	// Create approve transaction
	approveTx := gethtypes.NewTransaction(
		nonce,
		tokenAddr,
		big.NewInt(0),
		200000, // Gas limit for approve
		gasPrice,
		approveData,
	)
	
	// Sign transaction
	signedTx, err := h.signer.Signer(h.signer.From, approveTx)
	if err != nil {
		return fmt.Errorf("failed to sign approve transaction: %w", err)
	}
	
	// Send transaction
	if err := h.client.SendTransaction(ctx, signedTx); err != nil {
		return fmt.Errorf("failed to send approve transaction: %w", err)
	}
	
	fmt.Printf("   🚀 Approve transaction sent: %s\n", signedTx.Hash().Hex())
	
	// Wait for confirmation
	receipt, err := bind.WaitMined(ctx, h.client, signedTx)
	if err != nil {
		return fmt.Errorf("failed to wait for approve confirmation: %w", err)
	}
	
	if receipt.Status != 1 {
		return fmt.Errorf("approve transaction failed with status: %d", receipt.Status)
	}
	
	fmt.Printf("   ✅ Approval confirmed! Gas used: %d\n", receipt.GasUsed)
	return nil
}

// getOriginDomainFromArgs extracts the origin domain using the config system
func getOriginDomainFromArgs(args types.ParsedArgs) (uint32, error) {
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
