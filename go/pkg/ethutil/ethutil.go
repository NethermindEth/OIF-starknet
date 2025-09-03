package ethutil

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/solvercore/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC20ABI contains the minimal ABI for ERC20 operations
var ERC20ABI = `[
	{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
	{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
	{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}
]`

// NewTransactor creates a new transactor for transaction signing
func NewTransactor(chainID *big.Int, privateKey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(privateKey, chainID)
}

// SuggestGas suggests a gas price for the current network
func SuggestGas(client *ethclient.Client) (*big.Int, error) {
	return client.SuggestGasPrice(context.Background())
}

// GetChainID gets the chain ID for a given RPC URL
func GetChainID(rpcURL string) (*big.Int, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}
	defer client.Close()

	return client.ChainID(context.Background())
}

// GetBlockNumber gets the current block number for a given RPC URL
func GetBlockNumber(rpcURL string) (uint64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, fmt.Errorf("failed to dial RPC: %w", err)
	}
	defer client.Close()

	return client.BlockNumber(context.Background())
}

// ERC20Balance gets the ERC20 token balance for a given address
func ERC20Balance(client *ethclient.Client, tokenAddress, ownerAddress common.Address) (*big.Int, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	data, err := parsedABI.Pack("balanceOf", ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf call: %w", err)
	}

	msg := ethereum.CallMsg{To: &tokenAddress, Data: data}
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	// Check if result is empty (contract might not exist)
	if len(result) == 0 {
		return nil, fmt.Errorf("empty result from balanceOf call - contract may not exist at address %s", tokenAddress.Hex())
	}

	var balance *big.Int
	if err := parsedABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, fmt.Errorf("failed to unpack balanceOf result: %w", err)
	}

	return balance, nil
}

// ERC20Allowance gets the ERC20 token allowance for a given owner and spender
func ERC20Allowance(client *ethclient.Client, tokenAddress, ownerAddress, spenderAddress common.Address) (*big.Int, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	data, err := parsedABI.Pack("allowance", ownerAddress, spenderAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to pack allowance call: %w", err)
	}

	msg := ethereum.CallMsg{To: &tokenAddress, Data: data}
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance: %w", err)
	}

	var allowance *big.Int
	if err := parsedABI.UnpackIntoInterface(&allowance, "allowance", result); err != nil {
		return nil, fmt.Errorf("failed to unpack allowance result: %w", err)
	}

	return allowance, nil
}

// ERC20Transfer creates a transfer transaction for ERC20 tokens
func ERC20Transfer(client *ethclient.Client, auth *bind.TransactOpts, tokenAddress, recipientAddress common.Address, amount *big.Int) (*gethtypes.Transaction, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	data, err := parsedABI.Pack("transfer", recipientAddress, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transfer call: %w", err)
	}

	// Get current nonce
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get current gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction
	tx := gethtypes.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0), // No ETH value
		100000,        // Gas limit
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transfer transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transfer transaction: %w", err)
	}

	return signedTx, nil
}

// ERC20Approve creates an approve transaction for ERC20 tokens
func ERC20Approve(client *ethclient.Client, auth *bind.TransactOpts, tokenAddress, spenderAddress common.Address, amount *big.Int) (*gethtypes.Transaction, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	data, err := parsedABI.Pack("approve", spenderAddress, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack approve call: %w", err)
	}

	// Get current nonce
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get current gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction
	tx := gethtypes.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0), // No ETH value
		200000,        // Gas limit
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign approve transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send approve transaction: %w", err)
	}

	return signedTx, nil
}

// WaitForTransaction waits for a transaction to be mined and returns the receipt
func WaitForTransaction(client *ethclient.Client, tx *gethtypes.Transaction) (*gethtypes.Receipt, error) {
	return bind.WaitMined(context.Background(), client, tx)
}

// FormatTokenAmount formats a token amount from wei to tokens with specified decimals
// Uses the shared utility function from types package
func FormatTokenAmount(amount *big.Int, decimals int) string {
	return types.FormatTokenAmount(amount, decimals)
}

// ParsePrivateKey parses a hex private key string
func ParsePrivateKey(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	return crypto.HexToECDSA(privateKeyHex)
}
