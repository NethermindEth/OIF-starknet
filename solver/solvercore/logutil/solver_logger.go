package logutil

import (
	"fmt"
	"strings"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
)

// SolverLogger provides enhanced logging for solver operations
type SolverLogger struct {
	networkName string
}

// NewSolverLogger creates a new solver logger for a specific network
func NewSolverLogger(networkName string) *SolverLogger {
	return &SolverLogger{networkName: networkName}
}

// GetNetworkTag returns the network tag for the logger's network
func (sl *SolverLogger) GetNetworkTag() string {
	return Prefix(sl.networkName)
}

// GetNetworkTagByChainID returns the network tag for a given chain ID
func GetNetworkTagByChainID(chainID uint64) string {
	if networkName := NetworkNameByChainID(chainID); networkName != "" {
		return Prefix(networkName)
	}
	return "[UNK]"
}

// CrossChainOperation logs a cross-chain operation with origin → destination format
func CrossChainOperation(operation string, originChainID, destChainID uint64, orderID string) {
	originTag := GetNetworkTagByChainID(originChainID)
	destTag := GetNetworkTagByChainID(destChainID)

	// Keep the color codes for better readability
	originClean := strings.TrimSpace(originTag)
	destClean := strings.TrimSpace(destTag)

	fmt.Printf("%s → %s 🔄 %s (Order: %s)\n", originClean, destClean, operation, orderID[:8]+"...")
}

// removeColorCodes removes ANSI color codes from a string
func removeColorCodes(s string) string {
	// Remove common ANSI escape sequences
	s = strings.ReplaceAll(s, "\033[32m", "")       // green
	s = strings.ReplaceAll(s, "\033[91m", "")       // pastelRed
	s = strings.ReplaceAll(s, "\033[35m", "")       // purple
	s = strings.ReplaceAll(s, "\033[38;5;27m", "")  // royalBlue
	s = strings.ReplaceAll(s, "\033[38;5;208m", "") // orange
	s = strings.ReplaceAll(s, "\033[36m", "")       // cyan
	s = strings.ReplaceAll(s, "\033[0m", "")        // reset
	return s
}

// LogOrderProcessing logs order processing with cross-chain context
func LogOrderProcessing(args *types.ParsedArgs, operation string) {
	if args.ResolvedOrder.OriginChainID != nil && len(args.ResolvedOrder.FillInstructions) > 0 {
		destChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID
		if destChainID != nil {
			CrossChainOperation(operation,
				args.ResolvedOrder.OriginChainID.Uint64(),
				destChainID.Uint64(),
				args.OrderID)
		} else {
			fmt.Printf("🔄 %s (Order: %s)\n", operation, args.OrderID[:8]+"...")
		}
	} else {
		fmt.Printf("🔄 %s (Order: %s)\n", operation, args.OrderID[:8]+"...")
	}
}

// LogFillOperation logs a fill operation with network context
func LogFillOperation(networkName, orderID string, success bool) {
	tag := Prefix(networkName)
	if success {
		fmt.Printf("%s✅ Fill completed (Order: %s)\n", tag, orderID[:8]+"...")
	} else {
		fmt.Printf("%s❌ Fill failed (Order: %s)\n", tag, orderID[:8]+"...")
	}
}

// LogSettleOperation logs a settlement operation with network context
func LogSettleOperation(networkName, orderID string, success bool) {
	tag := Prefix(networkName)
	if success {
		fmt.Printf("%s✅ Settlement completed (Order: %s)\n", tag, orderID[:8]+"...")
	} else {
		fmt.Printf("%s❌ Settlement failed (Order: %s)\n", tag, orderID[:8]+"...")
	}
}

// LogBlockProcessing logs block processing with reduced verbosity
func LogBlockProcessing(networkName string, fromBlock, toBlock uint64, eventCount int) {
	tag := Prefix(networkName)
	if eventCount > 0 {
		fmt.Printf("%s📦 Processed blocks %d-%d: %d events\n", tag, fromBlock, toBlock, eventCount)
	} else if toBlock-fromBlock > 0 {
		// Only log if processing multiple blocks to reduce noise
		fmt.Printf("%s📦 Processed blocks %d-%d\n", tag, fromBlock, toBlock)
	}
}

// LogStatusCheck logs order status checks with retry information
func LogStatusCheck(networkName string, attempt, maxAttempts int, status, expected string) {
	tag := Prefix(networkName)
	if attempt == 1 {
		fmt.Printf("%s📊 Status: %s (expected: %s)\n", tag, status, expected)
	} else {
		fmt.Printf("%s📊 Retry %d/%d: %s (expected: %s)\n", tag, attempt, maxAttempts, status, expected)
	}
}

// LogRetryWait logs retry wait information
func LogRetryWait(networkName string, attempt, maxAttempts int, delay string) {
	tag := Prefix(networkName)
	fmt.Printf("%s⏳ Waiting %s before retry %d/%d...\n", tag, delay, attempt+1, maxAttempts)
}

// LogOperationComplete logs the completion of an operation with cross-chain context
func LogOperationComplete(args *types.ParsedArgs, operation string, success bool) {
	if args.ResolvedOrder.OriginChainID != nil && len(args.ResolvedOrder.FillInstructions) > 0 {
		destChainID := args.ResolvedOrder.FillInstructions[0].DestinationChainID
		if destChainID != nil {
			originTag := GetNetworkTagByChainID(args.ResolvedOrder.OriginChainID.Uint64())
			destTag := GetNetworkTagByChainID(destChainID.Uint64())

			originClean := strings.TrimSpace(originTag)
			destClean := strings.TrimSpace(destTag)

			if success {
				fmt.Printf("%s → %s ✅ %s completed (Order: %s)\n", originClean, destClean, operation, args.OrderID[:8]+"...")
			} else {
				fmt.Printf("%s → %s ❌ %s failed (Order: %s)\n", originClean, destClean, operation, args.OrderID[:8]+"...")
			}
		} else {
			if success {
				fmt.Printf("✅ %s completed (Order: %s)\n", operation, args.OrderID[:8]+"...")
			} else {
				fmt.Printf("❌ %s failed (Order: %s)\n", operation, args.OrderID[:8]+"...")
			}
		}
	} else {
		if success {
			fmt.Printf("✅ %s completed (Order: %s)\n", operation, args.OrderID[:8]+"...")
		} else {
			fmt.Printf("❌ %s failed (Order: %s)\n", operation, args.OrderID[:8]+"...")
		}
	}
}

// LogWithNetworkTagf adds a network tag to any log message
func LogWithNetworkTagf(networkName, format string, args ...interface{}) {
	tag := Prefix(networkName)
	fmt.Printf(tag+format, args...)
}

// LogPersistence logs persistence operations with reduced frequency
var persistenceCounters = make(map[string]int)

func LogPersistence(networkName string, blockNumber uint64) {
	counter := persistenceCounters[networkName]
	counter++
	persistenceCounters[networkName] = counter

	// Only log every 30 blocks or if it's the first block
	if counter%30 == 1 || counter == 1 {
		tag := Prefix(networkName)
		fmt.Printf("%s💾 Persisted LastIndexedBlock=%d\n", tag, blockNumber)
	}
}
