package hyperlane7683

// Module: EVM chain handler for Hyperlane7683
// - Executes fill/settle/status calls against EVM Hyperlane7683 contracts
// - Manages ERC20 approvals and gas/value handling for calls
//
// Interface Contract:
// - Fill(): Must acquire mutex, setup approvals, execute fill, return OrderAction
// - Settle(): Must acquire mutex, quote gas, execute settle
// - All methods should use consistent logging patterns and error handling

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/NethermindEth/oif-starknet/solver/pkg/envutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/config"
	contracts "github.com/NethermindEth/oif-starknet/solver/solvercore/contracts"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/logutil"
	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// Order status constants
	orderStatusFilled  = "FILLED"
	orderStatusSettled = "SETTLED"
	orderStatusUnknown = "UNKNOWN"
)

const (
	// Maximum retry attempts for order status checks
	maxRetryAttempts = 5
	// Gas limit for approve transactions
	approveGasLimit = 200000
)

// HyperlaneEVM contains all EVM-specific logic for the Hyperlane7683 protocol
type HyperlaneEVM struct {
	client  *ethclient.Client
	signer  *bind.TransactOpts
	chainID uint64
	mu      sync.Mutex // Serialize operations to prevent nonce conflicts
}

// NewHyperlaneEVM creates a new EVM handler for Hyperlane operations
func NewHyperlaneEVM(client *ethclient.Client, signer *bind.TransactOpts, chainID uint64) *HyperlaneEVM {
	return &HyperlaneEVM{
		client:  client,
		signer:  signer,
		chainID: chainID,
	}
}

// Fill executes a fill operation on an EVM chain
func (h *HyperlaneEVM) Fill(ctx context.Context, args types.ParsedArgs) (OrderAction, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return OrderActionError, fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	// Use the order ID from the event
	var orderID [32]byte
	orderIDBytes := common.FromHex(args.OrderID)
	copy(orderID[:], orderIDBytes)

	// Convert destination settler string to EVM address for contract operations
	destinationSettlerAddr, err := types.ToEVMAddress(instruction.DestinationSettler)
	if err != nil {
		return OrderActionError, fmt.Errorf("failed to convert destination settler to EVM address: %w", err)
	}

	// Pre-check: skip if order is already filled or settled
	status, err := h.GetOrderStatus(ctx, args)
	if err != nil {
		fmt.Printf("   ⚠️  Status check failed: %v\n", err)
		return OrderActionError, err
	}

	// Get network name for logging
	networkName := logutil.NetworkNameByChainID(h.chainID)
	logutil.LogStatusCheck(networkName, 1, 1, status, "UNKNOWN")

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

	// Execute the fill transaction
	// Get chain IDs for cross-chain logging
	originChainID := args.ResolvedOrder.OriginChainID.Uint64()
	destChainID := instruction.DestinationChainID.Uint64()
	logutil.CrossChainOperation(fmt.Sprintf("Executing fill call to contract %s", destinationSettlerAddr.Hex()), originChainID, destChainID, args.OrderID)

	// Set native token value if needed
	originalValue := h.signer.Value
	if len(args.ResolvedOrder.MaxSpent) > 0 && args.ResolvedOrder.MaxSpent[0].Token == "" {
		h.signer.Value = new(big.Int).Set(args.ResolvedOrder.MaxSpent[0].Amount)
	}
	defer func() { h.signer.Value = originalValue }()

	contract, err := contracts.NewHyperlane7683(destinationSettlerAddr, h.client)
	if err != nil {
		return OrderActionError, fmt.Errorf("failed to bind contract at %s: %w", instruction.DestinationSettler, err)
	}

	var fillerDataBytes []byte
	tx, err := contract.Fill(h.signer, orderID, instruction.OriginData, fillerDataBytes)
	if err != nil {
		return OrderActionError, fmt.Errorf("fill transaction failed: %w", err)
	}

	logutil.CrossChainOperation(fmt.Sprintf("Fill transaction sent: %s", tx.Hash().Hex()), originChainID, destChainID, args.OrderID)

	// Wait for confirmation
	receipt, err := bind.WaitMined(ctx, h.client, tx)
	if err != nil {
		return OrderActionError, fmt.Errorf("failed to wait for fill confirmation: %w", err)
	}

	if receipt.Status == 1 {
		logutil.CrossChainOperation(fmt.Sprintf("EVM Fill successful! Gas used: %d", receipt.GasUsed), originChainID, destChainID, args.OrderID)
		return OrderActionSettle, nil // Need to settle this order
	} else {
		return OrderActionError, fmt.Errorf("fill transaction failed with status: %d", receipt.Status)
	}
}

// Settle executes settlement on an EVM chain
func (h *HyperlaneEVM) Settle(ctx context.Context, args types.ParsedArgs) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	// Use the order ID from the event
	var orderID [32]byte
	orderIDBytes := common.FromHex(args.OrderID)
	copy(orderID[:], orderIDBytes)

	// Convert destination settler string to EVM address for contract operations
	destinationSettler, err := types.ToEVMAddress(instruction.DestinationSettler)
	if err != nil {
		return fmt.Errorf("failed to convert destination settler to EVM address: %w", err)
	}

	// Pre-settle check: ensure order is FILLED with retry logic
	status, err := h.waitForOrderStatus(ctx, args, "FILLED", maxRetryAttempts, 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get order status after retries: %w", err)
	}
	if status != "FILLED" {
		return fmt.Errorf("order status must be filled in order to settle, got: %s", status)
	}
	// if err := h.verifyOrderStatus(ctx, orderIDArr, destinationSettler, "FILLED"); err != nil {
	//	return fmt.Errorf("pre-settle check failed: %w", err)
	//}

	// Get the contract instance using the EVM address
	contract, err := contracts.NewHyperlane7683(destinationSettler, h.client)
	if err != nil {
		return fmt.Errorf("failed to bind contract at %s: %w", destinationSettler, err)
	}

	// Get gas payment (protocol fee) that must be sent with settlement
	originDomain, err := h.getOriginDomain(args)
	if err != nil {
		return fmt.Errorf("failed to get origin domain: %w", err)
	}

	// Check if origin is Starknet and we're on live networks (not forking)
	starknetDomain := uint32(config.Networks["Starknet"].HyperlaneDomain)
	if originDomain == starknetDomain {
		if !envutil.IsDevnet() {
			// Live networks: Skip settlement until Starknet domain is registered
			fmt.Printf("   ⚠️  Skipping EVM settlement for Starknet origin (domain %d) on live network\n", originDomain)
			fmt.Printf("   ⏳ Starknet domain not yet registered on EVM contracts - waiting for Hyperlane team\n")
			fmt.Printf("   📝 Order filled successfully, settlement will be available once domain is registered\n")
			return nil // Skip settlement but don't treat as error
		} else {
			// Fork mode: Continue with settlement (domains are mocked/registered)
			fmt.Printf("   🔧 Fork mode detected - proceeding with Starknet settlement (domain %d registered)\n", originDomain)
		}
	}

	// Get chain IDs for cross-chain logging
	originChainID := args.ResolvedOrder.OriginChainID.Uint64()
	destChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID.Uint64()
	logutil.CrossChainOperation(fmt.Sprintf("Quoting gas payment for origin domain: %d", originDomain), originChainID, destChainID, args.OrderID)
	gasPayment, err := contract.QuoteGasPayment(&bind.CallOpts{Context: ctx}, originDomain)
	if err != nil {
		return fmt.Errorf("quoteGasPayment failed on %s: %w", destinationSettler, err)
	}

	// Prepare order IDs array (contract expects array)
	orderIDs := make([][32]byte, 1)
	orderIDs[0] = orderID

	// Execute the settle transaction

	// Set txn value
	originalValue := h.signer.Value
	h.signer.Value = new(big.Int).Set(gasPayment)
	defer func() { h.signer.Value = originalValue }()

	//// Set gas price if not already set
	// if h.signer.GasPrice == nil || h.signer.GasPrice.Sign() == 0 {
	//	if suggested, gerr := h.client.SuggestGasPrice(ctx); gerr == nil {
	//		h.signer.GasPrice = suggested
	//	}
	//}

	tx, err := contract.Settle(h.signer, orderIDs)
	if err != nil {
		return fmt.Errorf("settle tx failed on %s: %w", destinationSettler, err)
	}
	logutil.CrossChainOperation(fmt.Sprintf("Settle transaction sent: %s", tx.Hash().Hex()), originChainID, destChainID, args.OrderID)

	// Wait for confirmation
	receipt, err := bind.WaitMined(ctx, h.client, tx)
	if err != nil {
		return fmt.Errorf("waiting settle failed on %s: %w", destinationSettler, err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("settle transaction failed on %s at block %d", destinationSettler, receipt.BlockNumber)
	}

	logutil.CrossChainOperation(fmt.Sprintf("Settle transaction confirmed at block %d (gasUsed=%d)", receipt.BlockNumber, receipt.GasUsed), originChainID, destChainID, args.OrderID)
	return nil
}

// GetOrderStatus returns the current status of an order
func (h *HyperlaneEVM) GetOrderStatus(ctx context.Context, args types.ParsedArgs) (string, error) {
	if len(args.ResolvedOrder.FillInstructions) == 0 {
		return orderStatusUnknown, fmt.Errorf("no fill instructions found")
	}

	instruction := args.ResolvedOrder.FillInstructions[0]

	var orderIDArr [32]byte
	orderIDBytes := common.FromHex(args.OrderID)
	copy(orderIDArr[:], orderIDBytes)

	//	// Derive orderId from keccak(origin_data)
	//	var orderIDArr [32]byte
	//	orderHash := crypto.Keccak256(instruction.OriginData)
	//	copy(orderIDArr[:], orderHash)

	// Check order status
	orderStatusABI := `[{"type":"function","name":"orderStatus","inputs":[{"type":"bytes32","name":"orderId"}],"outputs":[{"type":"bytes32","name":""}],"stateMutability":"view"}]`
	parsedABI, err := abi.JSON(strings.NewReader(orderStatusABI))
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to parse orderStatus ABI: %w", err)
	}

	callData, err := parsedABI.Pack("orderStatus", orderIDArr)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to pack orderStatus: %w", err)
	}

	// Convert destination settler string to EVM address for contract call
	destinationSettlerAddr, err := types.ToEVMAddress(instruction.DestinationSettler)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("failed to convert destination settler to EVM address: %w", err)
	}

	dummyFrom := common.HexToAddress("0x1000000000000000000000000000000000000000")
	res, err := h.client.CallContract(ctx, ethereum.CallMsg{From: dummyFrom, To: &destinationSettlerAddr, Data: callData}, nil)
	if err != nil {
		return "UNKNOWN", fmt.Errorf("orderStatus call failed: %w", err)
	}

	if len(res) < 32 {
		return "UNKNOWN", fmt.Errorf("invalid orderStatus result length: %d", len(res))
	}

	statusHash := common.BytesToHash(res[:32])
	status := h.interpretStatusHash(ctx, statusHash)
	return status, nil
}

// getOriginDomainFromArgs extracts the origin domain using the config system
func (h *HyperlaneEVM) getOriginDomain(args types.ParsedArgs) (uint32, error) {
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

// setupApprovals handles all ERC20 approvals needed for the fill operation
func (h *HyperlaneEVM) setupApprovals(ctx context.Context, args types.ParsedArgs, destinationSettlerAddr common.Address) error {
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
	// logutil.CrossChainOperation("Setting up token approvals", originChainID, destinationChainID, args.OrderID)

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

		// Convert token address string to EVM address for approval
		tokenAddr, err := types.ToEVMAddress(maxSpent.Token)
		if err != nil {
			return fmt.Errorf("failed to convert token address for approval: %w", err)
		}

		if err := h.ensureTokenApproval(ctx, tokenAddr, destinationSettlerAddr, maxSpent.Amount); err != nil {
			return fmt.Errorf("approval failed for token %s: %w", maxSpent.Token, err)
		}
	}
	logutil.CrossChainOperation("EVM token approvals set", originChainID, destinationChainID, args.OrderID)

	// Add a small delay to ensure blockchain state is updated after approvals
	time.Sleep(1 * time.Second)

	return nil
}

func (h *HyperlaneEVM) interpretStatusHash(_ context.Context, statusHash common.Hash) string {
	if statusHash == (common.Hash{}) {
		return "UNKNOWN"
	}

	// Try to read constants from contract for comparison
	filledHash := common.HexToHash("0x46494c4c45440000000000000000000000000000000000000000000000000000")
	if statusHash == filledHash {
		return "FILLED"
	}
	//	}
	//}

	// Check hardcoded SETTLED constant
	settledHash := common.HexToHash("0x534554544c454400000000000000000000000000000000000000000000000000")
	if statusHash == settledHash {
		return "SETTLED"
	}

	return statusHash.Hex()
}

// ensureTokenApproval ensures the solver has approved an arbitrary ERC20 token for the Hyperlane contract
func (h *HyperlaneEVM) ensureTokenApproval(ctx context.Context, tokenAddr, spender common.Address, amount *big.Int) error {
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

	if len(result) == 0 {
		// Token doesn't exist on this chain (likely cross-chain order) - skip approval
		fmt.Printf("   ⚠️  Token %s not found on this chain, skipping approval (cross-chain order)\n", tokenAddr.Hex())
		chainID, err := h.client.ChainID(ctx)
		if err == nil {
			fmt.Printf("   ⚠️  This chain ID: %s\n", chainID.String())
		}
		return nil
	}

	if len(result) < 32 {
		return fmt.Errorf("invalid allowance result length: %d", len(result))
	}

	currentAllowance := new(big.Int).SetBytes(result)

	// If allowance is sufficient, no approval needed
	if currentAllowance.Cmp(amount) >= 0 {
		return nil
	}

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
		approveGasLimit, // Gas limit for approve
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

// waitForOrderStatus waits for the order status to become the expected value with retry logic
func (h *HyperlaneEVM) waitForOrderStatus(ctx context.Context, args types.ParsedArgs, expectedStatus string, maxRetries int, initialDelay time.Duration) (string, error) {
	delay := initialDelay

	networkName := logutil.NetworkNameByChainID(h.chainID)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		status, err := h.GetOrderStatus(ctx, args)
		if err != nil {
			fmt.Printf("   ⚠️  Status check attempt %d failed: %v\n", attempt, err)
		} else {
			logutil.LogStatusCheck(networkName, attempt, maxRetries, status, expectedStatus)
			if status == expectedStatus {
				return status, nil
			}
		}

		// Don't wait after the last attempt
		if attempt < maxRetries {
			logutil.LogRetryWait(networkName, attempt, maxRetries, delay.String())
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
		return "UNKNOWN", fmt.Errorf("final status check failed after %d attempts: %w", maxRetries, err)
	}

	return finalStatus, nil
}
