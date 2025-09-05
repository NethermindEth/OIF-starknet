package main

// Main entry point for the OIF Starknet solver
// Provides a single binary with CLI routing to different tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/cmd/solver"
	openorder "github.com/NethermindEth/oif-starknet/go/cmd/tools/open-order"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "solver":
		// Run the main solver
		runSolver()
	case "tools":
		// Route to development tools
		runTools()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("OIF Starknet Solver - Single Binary for All Operations")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  solver <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  solver                    Run the main solver")
	fmt.Println("  tools <tool> [options]    Run development tools")
	fmt.Println("  help                      Show this help message")
	fmt.Println()
	fmt.Println("Development Tools:")
	fmt.Println("  tools open-order <chain>  Create test orders (starknet|evm)")
	fmt.Println("  tools setup-forks <cmd>   Setup forked networks")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  solver solver                    # Run main solver")
	fmt.Println("  solver tools open-order starknet # Create Starknet order")
	fmt.Println("  solver tools open-order evm      # Create EVM order")
	fmt.Println("  solver tools setup-forks deploy  # Deploy to forks")
}

func runSolver() {
	// Run the main solver
	solver.RunSolver()
}

func runTools() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: solver tools <tool> [options]")
		fmt.Println("Available tools: open-order, setup-forks")
		os.Exit(1)
	}

	tool := os.Args[2]

	switch tool {
	case "open-order":
		runOpenOrder()
	case "setup-forks":
		runSetupForks()
	default:
		fmt.Printf("Unknown tool: %s\n", tool)
		fmt.Println("Available tools: open-order, setup-forks")
		os.Exit(1)
	}
}

func runOpenOrder() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: solver tools open-order <chain> [command]")
		fmt.Println("Available chains: starknet, evm")
		fmt.Println("Available EVM commands: random-to-evm, random-to-sn, default-evm-evm, default-evm-sn")
		fmt.Println("Available Starknet commands: random, default")
		os.Exit(1)
	}

	chain := os.Args[3]

	switch strings.ToLower(chain) {
	case "starknet":
		// Get the command (default to random if not provided)
		command := "random"
		if len(os.Args) > 4 {
			command = os.Args[4]
		}
		// Run the real Starknet order creation logic
		openorder.RunStarknetOrder(command)
	case "evm":
		// Get the command (default to random-to-evm if not provided)
		command := "random-to-evm"
		if len(os.Args) > 4 {
			command = os.Args[4]
		}
		// Run the real EVM order creation logic
		openorder.RunEVMOrder(command)
	default:
		fmt.Printf("Unknown chain: %s\n", chain)
		fmt.Println("Available chains: starknet, evm")
		os.Exit(1)
	}
}

func runSetupForks() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: solver tools setup-forks <command>")
		fmt.Println("Available commands: deploy, declare, verify")
		os.Exit(1)
	}

	cmd := os.Args[3]

	switch strings.ToLower(cmd) {
	case "deploy":
		fmt.Println("🔧 Running fork deployment...")
		// This will call the deployment logic
	case "declare":
		fmt.Println("🔧 Running contract declaration...")
		// This will call the declaration logic
	case "verify":
		fmt.Println("🔧 Running deployment verification...")
		// This will call the verification logic
	default:
		fmt.Printf("Unknown setup command: %s\n", cmd)
		fmt.Println("Available commands: deploy, declare, verify")
		os.Exit(1)
	}
}
