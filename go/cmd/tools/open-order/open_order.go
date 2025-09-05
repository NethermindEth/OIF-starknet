package openorder

// OpenOrder package - contains the order creation logic
// This allows the order creation tools to be imported and run from the main CLI

import (
	"fmt"
	"os"
	"strings"
)

// RunOpenOrder runs Alice's order creation tool
func RunOpenOrder(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: open-order <chain> [command]")
		fmt.Println("Available chains: starknet, evm")
		os.Exit(1)
	}

	chain := strings.ToLower(args[0])
	command := "default"
	if len(args) > 1 {
		command = args[1]
	}

	switch chain {
	case "starknet":
		fmt.Println("🎯 Running Alice's Starknet order creation...")
		RunStarknetOrder(command)
	case "evm":
		fmt.Println("🎯 Running Alice's EVM order creation...")
		RunEVMOrder(command)
	default:
		fmt.Printf("Unknown chain: %s\n", chain)
		fmt.Println("Available chains: starknet, evm")
		os.Exit(1)
	}
}
