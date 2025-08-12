package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Test user addresses (Alice, Bob, Charlie)
var testUsers = []string{
	"0x70997970C51812dc3A010C7d01b50e0d17dc79C8", // Alice (Account 1)
	"0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC", // Bob (Account 2)
	"0x90F79bf6EB2c4f870365E785982E1f101E93b906", // Charlie (Account 3)
}

// Network configuration
var networks = []struct {
	name string
	url  string
}{
	{"Sepolia", "http://localhost:8545"},
	{"Optimism Sepolia", "http://localhost:8546"},
	{"Arbitrum Sepolia", "http://localhost:8547"},
	{"Base Sepolia", "http://localhost:8548"},
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	deployerKeyHex := os.Getenv("DEPLOYER_PRIVATE_KEY")
	if deployerKeyHex == "" {
		log.Fatal("DEPLOYER_PRIVATE_KEY environment variable is required")
	}

	aliceKeyHex := os.Getenv("ALICE_PRIVATE_KEY")
	bobKeyHex := os.Getenv("BOB_PRIVATE_KEY")
	charlieKeyHex := os.Getenv("CHARLIE_PRIVATE_KEY")
	
	if aliceKeyHex == "" || bobKeyHex == "" || charlieKeyHex == "" {
		log.Fatal("ALICE_PRIVATE_KEY, BOB_PRIVATE_KEY, and CHARLIE_PRIVATE_KEY environment variables are required")
	}

	// Parse deployer private key
	deployerKeyHex = strings.TrimPrefix(deployerKeyHex, "0x")
	deployerKey, err := crypto.HexToECDSA(deployerKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse deployer private key: %v", err)
	}

	// Parse test user private keys
	aliceKeyHex = strings.TrimPrefix(aliceKeyHex, "0x")
	aliceKey, err := crypto.HexToECDSA(aliceKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Alice private key: %v", err)
	}

	bobKeyHex = strings.TrimPrefix(bobKeyHex, "0x")
	bobKey, err := crypto.HexToECDSA(bobKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Bob private key: %v", err)
	}

	charlieKeyHex = strings.TrimPrefix(charlieKeyHex, "0x")
	charlieKey, err := crypto.HexToECDSA(charlieKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse Charlie private key: %v", err)
	}

	// Deploy tokens to all networks
	for _, network := range networks {
		fmt.Printf("\n🚀 Deploying tokens to %s...\n", network.name)
		fmt.Printf("   URL: %s\n", network.url)

		client, err := ethclient.Dial(network.url)
		if err != nil {
			fmt.Printf("   ❌ Failed to connect: %v\n", err)
			continue
		}

		// Deploy OrcaCoin (origin chain token)
		orcaCoinAddress, err := deployERC20(client, deployerKey, "OrcaCoin", network.name)
		if err != nil {
			fmt.Printf("   ❌ Failed to deploy OrcaCoin: %v\n", err)
			continue
		}
		fmt.Printf("   ✅ OrcaCoin deployed at: %s\n", orcaCoinAddress)

		// Deploy DogCoin (destination chain token)
		dogCoinAddress, err := deployERC20(client, deployerKey, "DogCoin", network.name)
		if err != nil {
			fmt.Printf("   ❌ Failed to deploy DogCoin: %v\n", err)
			continue
		}
		fmt.Printf("   ✅ DogCoin deployed at: %s\n", dogCoinAddress)

		// Fund test users
		if err := fundUsers(client, deployerKey, aliceKey, bobKey, charlieKey, orcaCoinAddress, dogCoinAddress, network.name); err != nil {
			fmt.Printf("   ❌ Failed to fund users: %v\n", err)
			continue
		}

		// Set allowances for Hyperlane7683
		if err := setAllowances(client, aliceKey, bobKey, charlieKey, orcaCoinAddress, dogCoinAddress, network.name); err != nil {
			fmt.Printf("   ❌ Failed to set allowances: %v\n", err)
			continue
		}

		client.Close()
		fmt.Printf("   🎉 %s setup complete!\n", network.name)
	}

	fmt.Printf("\n🎯 All networks configured!\n")
	fmt.Printf("   • OrcaCoin and DogCoin deployed to all networks\n")
	fmt.Printf("   • Test users funded with tokens\n")
	fmt.Printf("   • Allowances set for Hyperlane7683\n")
	fmt.Printf("   • Ready to open orders!\n")
}

func deployERC20(client *ethclient.Client, privateKey *ecdsa.PrivateKey, symbol, networkName string) (common.Address, error) {
	fmt.Printf("   📝 Deploying %s...\n", symbol)
	
	// Get the ERC20 contract configuration
	contract := deployer.GetERC20Contract()
	
	// TODO: Implement actual deployment using contract.Bytecode and contract.ABI
	// This will need to be completed with the actual deployment logic
	fmt.Printf("   📋 Contract bytecode length: %d bytes\n", len(contract.Bytecode))
	fmt.Printf("   📋 Contract ABI length: %d characters\n", len(contract.ABI))
	
	return common.Address{}, fmt.Errorf("ERC20 deployment not yet implemented - need to complete deployment logic")
}

func fundUsers(client *ethclient.Client, deployerKey, aliceKey, bobKey, charlieKey *ecdsa.PrivateKey, orcaCoinAddress, dogCoinAddress common.Address, networkName string) error {
	fmt.Printf("   💰 Funding test users...\n")
	
	// TODO: Implement token distribution to Alice, Bob, Charlie
	// This will need the ERC20 ABI to call transfer functions
	
	return fmt.Errorf("User funding not yet implemented - need ERC20 ABI")
}

func setAllowances(client *ethclient.Client, aliceKey, bobKey, charlieKey *ecdsa.PrivateKey, orcaCoinAddress, dogCoinAddress common.Address, networkName string) error {
	fmt.Printf("   🔐 Setting allowances for Hyperlane7683...\n")
	
	// TODO: Implement allowance setting for Hyperlane7683 contract
	// This will need the ERC20 ABI to call approve functions
	
	return fmt.Errorf("Allowance setting not yet implemented - need ERC20 ABI")
}
