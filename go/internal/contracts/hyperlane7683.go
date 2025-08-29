// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// FillInstruction is an auto generated low-level Go binding around an user-defined struct.
type FillInstruction struct {
	DestinationChainId *big.Int
	DestinationSettler [32]byte
	OriginData         []byte
}

// GasRouterGasRouterConfig is an auto generated low-level Go binding around an user-defined struct.
type GasRouterGasRouterConfig struct {
	Domain uint32
	Gas    *big.Int
}

// GaslessCrossChainOrder is an auto generated low-level Go binding around an user-defined struct.
type GaslessCrossChainOrder struct {
	OriginSettler common.Address
	User          common.Address
	Nonce         *big.Int
	OriginChainId *big.Int
	OpenDeadline  uint32
	FillDeadline  uint32
	OrderDataType [32]byte
	OrderData     []byte
}

// OnchainCrossChainOrder is an auto generated low-level Go binding around an user-defined struct.
type OnchainCrossChainOrder struct {
	FillDeadline  uint32
	OrderDataType [32]byte
	OrderData     []byte
}

// Output is an auto generated low-level Go binding around an user-defined struct.
type Output struct {
	Token     [32]byte
	Amount    *big.Int
	Recipient [32]byte
	ChainId   *big.Int
}

// ResolvedCrossChainOrder is an auto generated low-level Go binding around an user-defined struct.
type ResolvedCrossChainOrder struct {
	User             common.Address
	OriginChainId    *big.Int
	OpenDeadline     uint32
	FillDeadline     uint32
	OrderId          [32]byte
	MaxSpent         []Output
	MinReceived      []Output
	FillInstructions []FillInstruction
}

// Hyperlane7683MetaData contains all meta data concerning the Hyperlane7683 contract.
var Hyperlane7683MetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_mailbox\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_permit2\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"FILLED\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OPENED\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PACKAGE_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PERMIT2\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPermit2\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"REFUNDED\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"RESOLVED_CROSS_CHAIN_ORDER_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SETTLED\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"UNKNOWN\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"destinationGas\",\"inputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"domains\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"enrollRemoteRouter\",\"inputs\":[{\"name\":\"_domain\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_router\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"enrollRemoteRouters\",\"inputs\":[{\"name\":\"_domains\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"_addresses\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"fill\",\"inputs\":[{\"name\":\"_orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_originData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_fillerData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"filledOrders\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"originData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"fillerData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"handle\",\"inputs\":[{\"name\":\"_origin\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_sender\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"hook\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPostDispatchHook\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_customHook\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_interchainSecurityModule\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"interchainSecurityModule\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIInterchainSecurityModule\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"invalidateNonces\",\"inputs\":[{\"name\":\"_nonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isValidNonce\",\"inputs\":[{\"name\":\"_from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_nonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"localDomain\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"mailbox\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIMailbox\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"open\",\"inputs\":[{\"name\":\"_order\",\"type\":\"tuple\",\"internalType\":\"structOnchainCrossChainOrder\",\"components\":[{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderDataType\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"orderData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"openFor\",\"inputs\":[{\"name\":\"_order\",\"type\":\"tuple\",\"internalType\":\"structGaslessCrossChainOrder\",\"components\":[{\"name\":\"originSettler\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"originChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"openDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderDataType\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"orderData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"_signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_originFillerData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"openOrders\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"orderData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"orderStatus\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"quoteGasPayment\",\"inputs\":[{\"name\":\"_destinationDomain\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"refund\",\"inputs\":[{\"name\":\"_orders\",\"type\":\"tuple[]\",\"internalType\":\"structOnchainCrossChainOrder[]\",\"components\":[{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderDataType\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"orderData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"refund\",\"inputs\":[{\"name\":\"_orders\",\"type\":\"tuple[]\",\"internalType\":\"structGaslessCrossChainOrder[]\",\"components\":[{\"name\":\"originSettler\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"originChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"openDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderDataType\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"orderData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"resolve\",\"inputs\":[{\"name\":\"_order\",\"type\":\"tuple\",\"internalType\":\"structOnchainCrossChainOrder\",\"components\":[{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderDataType\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"orderData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[{\"name\":\"_resolvedOrder\",\"type\":\"tuple\",\"internalType\":\"structResolvedCrossChainOrder\",\"components\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"originChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"openDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxSpent\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"minReceived\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"fillInstructions\",\"type\":\"tuple[]\",\"internalType\":\"structFillInstruction[]\",\"components\":[{\"name\":\"destinationChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"destinationSettler\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"originData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"resolveFor\",\"inputs\":[{\"name\":\"_order\",\"type\":\"tuple\",\"internalType\":\"structGaslessCrossChainOrder\",\"components\":[{\"name\":\"originSettler\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"originChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"openDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderDataType\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"orderData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"_originFillerData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"_resolvedOrder\",\"type\":\"tuple\",\"internalType\":\"structResolvedCrossChainOrder\",\"components\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"originChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"openDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxSpent\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"minReceived\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"fillInstructions\",\"type\":\"tuple[]\",\"internalType\":\"structFillInstruction[]\",\"components\":[{\"name\":\"destinationChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"destinationSettler\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"originData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"routers\",\"inputs\":[{\"name\":\"_domain\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setDestinationGas\",\"inputs\":[{\"name\":\"domain\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"gas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setDestinationGas\",\"inputs\":[{\"name\":\"gasConfigs\",\"type\":\"tuple[]\",\"internalType\":\"structGasRouter.GasRouterConfig[]\",\"components\":[{\"name\":\"domain\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"gas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setHook\",\"inputs\":[{\"name\":\"_hook\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setInterchainSecurityModule\",\"inputs\":[{\"name\":\"_module\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"settle\",\"inputs\":[{\"name\":\"_orderIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unenrollRemoteRouter\",\"inputs\":[{\"name\":\"_domain\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unenrollRemoteRouters\",\"inputs\":[{\"name\":\"_domains\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"usedNonces\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"witnessHash\",\"inputs\":[{\"name\":\"_resolvedOrder\",\"type\":\"tuple\",\"internalType\":\"structResolvedCrossChainOrder\",\"components\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"originChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"openDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxSpent\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"minReceived\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"fillInstructions\",\"type\":\"tuple[]\",\"internalType\":\"structFillInstruction[]\",\"components\":[{\"name\":\"destinationChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"destinationSettler\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"originData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"witnessTypeString\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"Filled\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"originData\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"},{\"name\":\"fillerData\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"GasSet\",\"inputs\":[{\"name\":\"domain\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"gas\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"HookSet\",\"inputs\":[{\"name\":\"_hook\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"IsmSet\",\"inputs\":[{\"name\":\"_ism\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NonceInvalidation\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Open\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"resolvedOrder\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structResolvedCrossChainOrder\",\"components\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"originChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"openDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"fillDeadline\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"orderId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxSpent\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"minReceived\",\"type\":\"tuple[]\",\"internalType\":\"structOutput[]\",\"components\":[{\"name\":\"token\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"fillInstructions\",\"type\":\"tuple[]\",\"internalType\":\"structFillInstruction[]\",\"components\":[{\"name\":\"destinationChainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"destinationSettler\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"originData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Refund\",\"inputs\":[{\"name\":\"orderIds\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Refunded\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"receiver\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Settle\",\"inputs\":[{\"name\":\"orderIds\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"ordersFillerData\",\"type\":\"bytes[]\",\"indexed\":false,\"internalType\":\"bytes[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Settled\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"receiver\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidDomain\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidGaslessOrderOrigin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidGaslessOrderSettler\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidNativeAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidNonce\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOrderDomain\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOrderId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOrderOrigin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOrderStatus\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOrderType\",\"inputs\":[{\"name\":\"orderType\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidOriginDomain\",\"inputs\":[{\"name\":\"originDomain\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"type\":\"error\",\"name\":\"InvalidSender\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OrderFillExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OrderFillNotExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OrderOpenExpired\",\"inputs\":[]}]",
}

// Hyperlane7683ABI is the input ABI used to generate the binding from.
// Deprecated: Use Hyperlane7683MetaData.ABI instead.
var Hyperlane7683ABI = Hyperlane7683MetaData.ABI

// Hyperlane7683 is an auto generated Go binding around an Ethereum contract.
type Hyperlane7683 struct {
	Hyperlane7683Caller     // Read-only binding to the contract
	Hyperlane7683Transactor // Write-only binding to the contract
	Hyperlane7683Filterer   // Log filterer for contract events
}

// Hyperlane7683Caller is an auto generated read-only Go binding around an Ethereum contract.
type Hyperlane7683Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Hyperlane7683Transactor is an auto generated write-only Go binding around an Ethereum contract.
type Hyperlane7683Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Hyperlane7683Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Hyperlane7683Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Hyperlane7683Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Hyperlane7683Session struct {
	Contract     *Hyperlane7683    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Hyperlane7683CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Hyperlane7683CallerSession struct {
	Contract *Hyperlane7683Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// Hyperlane7683TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Hyperlane7683TransactorSession struct {
	Contract     *Hyperlane7683Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// Hyperlane7683Raw is an auto generated low-level Go binding around an Ethereum contract.
type Hyperlane7683Raw struct {
	Contract *Hyperlane7683 // Generic contract binding to access the raw methods on
}

// Hyperlane7683CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Hyperlane7683CallerRaw struct {
	Contract *Hyperlane7683Caller // Generic read-only contract binding to access the raw methods on
}

// Hyperlane7683TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Hyperlane7683TransactorRaw struct {
	Contract *Hyperlane7683Transactor // Generic write-only contract binding to access the raw methods on
}

// NewHyperlane7683 creates a new instance of Hyperlane7683, bound to a specific deployed contract.
func NewHyperlane7683(address common.Address, backend bind.ContractBackend) (*Hyperlane7683, error) {
	contract, err := bindHyperlane7683(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683{Hyperlane7683Caller: Hyperlane7683Caller{contract: contract}, Hyperlane7683Transactor: Hyperlane7683Transactor{contract: contract}, Hyperlane7683Filterer: Hyperlane7683Filterer{contract: contract}}, nil
}

// NewHyperlane7683Caller creates a new read-only instance of Hyperlane7683, bound to a specific deployed contract.
func NewHyperlane7683Caller(address common.Address, caller bind.ContractCaller) (*Hyperlane7683Caller, error) {
	contract, err := bindHyperlane7683(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683Caller{contract: contract}, nil
}

// NewHyperlane7683Transactor creates a new write-only instance of Hyperlane7683, bound to a specific deployed contract.
func NewHyperlane7683Transactor(address common.Address, transactor bind.ContractTransactor) (*Hyperlane7683Transactor, error) {
	contract, err := bindHyperlane7683(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683Transactor{contract: contract}, nil
}

// NewHyperlane7683Filterer creates a new log filterer instance of Hyperlane7683, bound to a specific deployed contract.
func NewHyperlane7683Filterer(address common.Address, filterer bind.ContractFilterer) (*Hyperlane7683Filterer, error) {
	contract, err := bindHyperlane7683(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683Filterer{contract: contract}, nil
}

// bindHyperlane7683 binds a generic wrapper to an already deployed contract.
func bindHyperlane7683(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := Hyperlane7683MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hyperlane7683 *Hyperlane7683Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hyperlane7683.Contract.Hyperlane7683Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hyperlane7683 *Hyperlane7683Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Hyperlane7683Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hyperlane7683 *Hyperlane7683Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Hyperlane7683Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hyperlane7683 *Hyperlane7683CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hyperlane7683.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hyperlane7683 *Hyperlane7683TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hyperlane7683 *Hyperlane7683TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.contract.Transact(opts, method, params...)
}

// FILLED is a free data retrieval call binding the contract method 0x432314a8.
//
// Solidity: function FILLED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) FILLED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "FILLED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// FILLED is a free data retrieval call binding the contract method 0x432314a8.
//
// Solidity: function FILLED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) FILLED() ([32]byte, error) {
	return _Hyperlane7683.Contract.FILLED(&_Hyperlane7683.CallOpts)
}

// FILLED is a free data retrieval call binding the contract method 0x432314a8.
//
// Solidity: function FILLED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) FILLED() ([32]byte, error) {
	return _Hyperlane7683.Contract.FILLED(&_Hyperlane7683.CallOpts)
}

// OPENED is a free data retrieval call binding the contract method 0xaa23a8f4.
//
// Solidity: function OPENED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) OPENED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "OPENED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OPENED is a free data retrieval call binding the contract method 0xaa23a8f4.
//
// Solidity: function OPENED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) OPENED() ([32]byte, error) {
	return _Hyperlane7683.Contract.OPENED(&_Hyperlane7683.CallOpts)
}

// OPENED is a free data retrieval call binding the contract method 0xaa23a8f4.
//
// Solidity: function OPENED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) OPENED() ([32]byte, error) {
	return _Hyperlane7683.Contract.OPENED(&_Hyperlane7683.CallOpts)
}

// PACKAGEVERSION is a free data retrieval call binding the contract method 0x93c44847.
//
// Solidity: function PACKAGE_VERSION() view returns(string)
func (_Hyperlane7683 *Hyperlane7683Caller) PACKAGEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "PACKAGE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// PACKAGEVERSION is a free data retrieval call binding the contract method 0x93c44847.
//
// Solidity: function PACKAGE_VERSION() view returns(string)
func (_Hyperlane7683 *Hyperlane7683Session) PACKAGEVERSION() (string, error) {
	return _Hyperlane7683.Contract.PACKAGEVERSION(&_Hyperlane7683.CallOpts)
}

// PACKAGEVERSION is a free data retrieval call binding the contract method 0x93c44847.
//
// Solidity: function PACKAGE_VERSION() view returns(string)
func (_Hyperlane7683 *Hyperlane7683CallerSession) PACKAGEVERSION() (string, error) {
	return _Hyperlane7683.Contract.PACKAGEVERSION(&_Hyperlane7683.CallOpts)
}

// PERMIT2 is a free data retrieval call binding the contract method 0x6afdd850.
//
// Solidity: function PERMIT2() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Caller) PERMIT2(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "PERMIT2")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PERMIT2 is a free data retrieval call binding the contract method 0x6afdd850.
//
// Solidity: function PERMIT2() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Session) PERMIT2() (common.Address, error) {
	return _Hyperlane7683.Contract.PERMIT2(&_Hyperlane7683.CallOpts)
}

// PERMIT2 is a free data retrieval call binding the contract method 0x6afdd850.
//
// Solidity: function PERMIT2() view returns(address)
func (_Hyperlane7683 *Hyperlane7683CallerSession) PERMIT2() (common.Address, error) {
	return _Hyperlane7683.Contract.PERMIT2(&_Hyperlane7683.CallOpts)
}

// REFUNDED is a free data retrieval call binding the contract method 0x94e15c8f.
//
// Solidity: function REFUNDED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) REFUNDED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "REFUNDED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// REFUNDED is a free data retrieval call binding the contract method 0x94e15c8f.
//
// Solidity: function REFUNDED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) REFUNDED() ([32]byte, error) {
	return _Hyperlane7683.Contract.REFUNDED(&_Hyperlane7683.CallOpts)
}

// REFUNDED is a free data retrieval call binding the contract method 0x94e15c8f.
//
// Solidity: function REFUNDED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) REFUNDED() ([32]byte, error) {
	return _Hyperlane7683.Contract.REFUNDED(&_Hyperlane7683.CallOpts)
}

// RESOLVEDCROSSCHAINORDERTYPEHASH is a free data retrieval call binding the contract method 0x74d70750.
//
// Solidity: function RESOLVED_CROSS_CHAIN_ORDER_TYPEHASH() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) RESOLVEDCROSSCHAINORDERTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "RESOLVED_CROSS_CHAIN_ORDER_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RESOLVEDCROSSCHAINORDERTYPEHASH is a free data retrieval call binding the contract method 0x74d70750.
//
// Solidity: function RESOLVED_CROSS_CHAIN_ORDER_TYPEHASH() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) RESOLVEDCROSSCHAINORDERTYPEHASH() ([32]byte, error) {
	return _Hyperlane7683.Contract.RESOLVEDCROSSCHAINORDERTYPEHASH(&_Hyperlane7683.CallOpts)
}

// RESOLVEDCROSSCHAINORDERTYPEHASH is a free data retrieval call binding the contract method 0x74d70750.
//
// Solidity: function RESOLVED_CROSS_CHAIN_ORDER_TYPEHASH() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) RESOLVEDCROSSCHAINORDERTYPEHASH() ([32]byte, error) {
	return _Hyperlane7683.Contract.RESOLVEDCROSSCHAINORDERTYPEHASH(&_Hyperlane7683.CallOpts)
}

// SETTLED is a free data retrieval call binding the contract method 0x80bdaf03.
//
// Solidity: function SETTLED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) SETTLED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "SETTLED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SETTLED is a free data retrieval call binding the contract method 0x80bdaf03.
//
// Solidity: function SETTLED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) SETTLED() ([32]byte, error) {
	return _Hyperlane7683.Contract.SETTLED(&_Hyperlane7683.CallOpts)
}

// SETTLED is a free data retrieval call binding the contract method 0x80bdaf03.
//
// Solidity: function SETTLED() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) SETTLED() ([32]byte, error) {
	return _Hyperlane7683.Contract.SETTLED(&_Hyperlane7683.CallOpts)
}

// UNKNOWN is a free data retrieval call binding the contract method 0x0c78932d.
//
// Solidity: function UNKNOWN() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) UNKNOWN(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "UNKNOWN")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// UNKNOWN is a free data retrieval call binding the contract method 0x0c78932d.
//
// Solidity: function UNKNOWN() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) UNKNOWN() ([32]byte, error) {
	return _Hyperlane7683.Contract.UNKNOWN(&_Hyperlane7683.CallOpts)
}

// UNKNOWN is a free data retrieval call binding the contract method 0x0c78932d.
//
// Solidity: function UNKNOWN() view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) UNKNOWN() ([32]byte, error) {
	return _Hyperlane7683.Contract.UNKNOWN(&_Hyperlane7683.CallOpts)
}

// DestinationGas is a free data retrieval call binding the contract method 0x775313a1.
//
// Solidity: function destinationGas(uint32 ) view returns(uint256)
func (_Hyperlane7683 *Hyperlane7683Caller) DestinationGas(opts *bind.CallOpts, arg0 uint32) (*big.Int, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "destinationGas", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DestinationGas is a free data retrieval call binding the contract method 0x775313a1.
//
// Solidity: function destinationGas(uint32 ) view returns(uint256)
func (_Hyperlane7683 *Hyperlane7683Session) DestinationGas(arg0 uint32) (*big.Int, error) {
	return _Hyperlane7683.Contract.DestinationGas(&_Hyperlane7683.CallOpts, arg0)
}

// DestinationGas is a free data retrieval call binding the contract method 0x775313a1.
//
// Solidity: function destinationGas(uint32 ) view returns(uint256)
func (_Hyperlane7683 *Hyperlane7683CallerSession) DestinationGas(arg0 uint32) (*big.Int, error) {
	return _Hyperlane7683.Contract.DestinationGas(&_Hyperlane7683.CallOpts, arg0)
}

// Domains is a free data retrieval call binding the contract method 0x440df4f4.
//
// Solidity: function domains() view returns(uint32[])
func (_Hyperlane7683 *Hyperlane7683Caller) Domains(opts *bind.CallOpts) ([]uint32, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "domains")

	if err != nil {
		return *new([]uint32), err
	}

	out0 := *abi.ConvertType(out[0], new([]uint32)).(*[]uint32)

	return out0, err

}

// Domains is a free data retrieval call binding the contract method 0x440df4f4.
//
// Solidity: function domains() view returns(uint32[])
func (_Hyperlane7683 *Hyperlane7683Session) Domains() ([]uint32, error) {
	return _Hyperlane7683.Contract.Domains(&_Hyperlane7683.CallOpts)
}

// Domains is a free data retrieval call binding the contract method 0x440df4f4.
//
// Solidity: function domains() view returns(uint32[])
func (_Hyperlane7683 *Hyperlane7683CallerSession) Domains() ([]uint32, error) {
	return _Hyperlane7683.Contract.Domains(&_Hyperlane7683.CallOpts)
}

// FilledOrders is a free data retrieval call binding the contract method 0xabeaae5e.
//
// Solidity: function filledOrders(bytes32 orderId) view returns(bytes originData, bytes fillerData)
func (_Hyperlane7683 *Hyperlane7683Caller) FilledOrders(opts *bind.CallOpts, orderId [32]byte) (struct {
	OriginData []byte
	FillerData []byte
}, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "filledOrders", orderId)

	outstruct := new(struct {
		OriginData []byte
		FillerData []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.OriginData = *abi.ConvertType(out[0], new([]byte)).(*[]byte)
	outstruct.FillerData = *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return *outstruct, err

}

// FilledOrders is a free data retrieval call binding the contract method 0xabeaae5e.
//
// Solidity: function filledOrders(bytes32 orderId) view returns(bytes originData, bytes fillerData)
func (_Hyperlane7683 *Hyperlane7683Session) FilledOrders(orderId [32]byte) (struct {
	OriginData []byte
	FillerData []byte
}, error) {
	return _Hyperlane7683.Contract.FilledOrders(&_Hyperlane7683.CallOpts, orderId)
}

// FilledOrders is a free data retrieval call binding the contract method 0xabeaae5e.
//
// Solidity: function filledOrders(bytes32 orderId) view returns(bytes originData, bytes fillerData)
func (_Hyperlane7683 *Hyperlane7683CallerSession) FilledOrders(orderId [32]byte) (struct {
	OriginData []byte
	FillerData []byte
}, error) {
	return _Hyperlane7683.Contract.FilledOrders(&_Hyperlane7683.CallOpts, orderId)
}

// Hook is a free data retrieval call binding the contract method 0x7f5a7c7b.
//
// Solidity: function hook() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Caller) Hook(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "hook")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Hook is a free data retrieval call binding the contract method 0x7f5a7c7b.
//
// Solidity: function hook() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Session) Hook() (common.Address, error) {
	return _Hyperlane7683.Contract.Hook(&_Hyperlane7683.CallOpts)
}

// Hook is a free data retrieval call binding the contract method 0x7f5a7c7b.
//
// Solidity: function hook() view returns(address)
func (_Hyperlane7683 *Hyperlane7683CallerSession) Hook() (common.Address, error) {
	return _Hyperlane7683.Contract.Hook(&_Hyperlane7683.CallOpts)
}

// InterchainSecurityModule is a free data retrieval call binding the contract method 0xde523cf3.
//
// Solidity: function interchainSecurityModule() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Caller) InterchainSecurityModule(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "interchainSecurityModule")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// InterchainSecurityModule is a free data retrieval call binding the contract method 0xde523cf3.
//
// Solidity: function interchainSecurityModule() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Session) InterchainSecurityModule() (common.Address, error) {
	return _Hyperlane7683.Contract.InterchainSecurityModule(&_Hyperlane7683.CallOpts)
}

// InterchainSecurityModule is a free data retrieval call binding the contract method 0xde523cf3.
//
// Solidity: function interchainSecurityModule() view returns(address)
func (_Hyperlane7683 *Hyperlane7683CallerSession) InterchainSecurityModule() (common.Address, error) {
	return _Hyperlane7683.Contract.InterchainSecurityModule(&_Hyperlane7683.CallOpts)
}

// IsValidNonce is a free data retrieval call binding the contract method 0x0647ee20.
//
// Solidity: function isValidNonce(address _from, uint256 _nonce) view returns(bool)
func (_Hyperlane7683 *Hyperlane7683Caller) IsValidNonce(opts *bind.CallOpts, _from common.Address, _nonce *big.Int) (bool, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "isValidNonce", _from, _nonce)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidNonce is a free data retrieval call binding the contract method 0x0647ee20.
//
// Solidity: function isValidNonce(address _from, uint256 _nonce) view returns(bool)
func (_Hyperlane7683 *Hyperlane7683Session) IsValidNonce(_from common.Address, _nonce *big.Int) (bool, error) {
	return _Hyperlane7683.Contract.IsValidNonce(&_Hyperlane7683.CallOpts, _from, _nonce)
}

// IsValidNonce is a free data retrieval call binding the contract method 0x0647ee20.
//
// Solidity: function isValidNonce(address _from, uint256 _nonce) view returns(bool)
func (_Hyperlane7683 *Hyperlane7683CallerSession) IsValidNonce(_from common.Address, _nonce *big.Int) (bool, error) {
	return _Hyperlane7683.Contract.IsValidNonce(&_Hyperlane7683.CallOpts, _from, _nonce)
}

// LocalDomain is a free data retrieval call binding the contract method 0x8d3638f4.
//
// Solidity: function localDomain() view returns(uint32)
func (_Hyperlane7683 *Hyperlane7683Caller) LocalDomain(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "localDomain")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// LocalDomain is a free data retrieval call binding the contract method 0x8d3638f4.
//
// Solidity: function localDomain() view returns(uint32)
func (_Hyperlane7683 *Hyperlane7683Session) LocalDomain() (uint32, error) {
	return _Hyperlane7683.Contract.LocalDomain(&_Hyperlane7683.CallOpts)
}

// LocalDomain is a free data retrieval call binding the contract method 0x8d3638f4.
//
// Solidity: function localDomain() view returns(uint32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) LocalDomain() (uint32, error) {
	return _Hyperlane7683.Contract.LocalDomain(&_Hyperlane7683.CallOpts)
}

// Mailbox is a free data retrieval call binding the contract method 0xd5438eae.
//
// Solidity: function mailbox() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Caller) Mailbox(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "mailbox")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Mailbox is a free data retrieval call binding the contract method 0xd5438eae.
//
// Solidity: function mailbox() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Session) Mailbox() (common.Address, error) {
	return _Hyperlane7683.Contract.Mailbox(&_Hyperlane7683.CallOpts)
}

// Mailbox is a free data retrieval call binding the contract method 0xd5438eae.
//
// Solidity: function mailbox() view returns(address)
func (_Hyperlane7683 *Hyperlane7683CallerSession) Mailbox() (common.Address, error) {
	return _Hyperlane7683.Contract.Mailbox(&_Hyperlane7683.CallOpts)
}

// OpenOrders is a free data retrieval call binding the contract method 0x66cb4581.
//
// Solidity: function openOrders(bytes32 orderId) view returns(bytes orderData)
func (_Hyperlane7683 *Hyperlane7683Caller) OpenOrders(opts *bind.CallOpts, orderId [32]byte) ([]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "openOrders", orderId)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// OpenOrders is a free data retrieval call binding the contract method 0x66cb4581.
//
// Solidity: function openOrders(bytes32 orderId) view returns(bytes orderData)
func (_Hyperlane7683 *Hyperlane7683Session) OpenOrders(orderId [32]byte) ([]byte, error) {
	return _Hyperlane7683.Contract.OpenOrders(&_Hyperlane7683.CallOpts, orderId)
}

// OpenOrders is a free data retrieval call binding the contract method 0x66cb4581.
//
// Solidity: function openOrders(bytes32 orderId) view returns(bytes orderData)
func (_Hyperlane7683 *Hyperlane7683CallerSession) OpenOrders(orderId [32]byte) ([]byte, error) {
	return _Hyperlane7683.Contract.OpenOrders(&_Hyperlane7683.CallOpts, orderId)
}

// OrderStatus is a free data retrieval call binding the contract method 0x2dff692d.
//
// Solidity: function orderStatus(bytes32 orderId) view returns(bytes32 status)
func (_Hyperlane7683 *Hyperlane7683Caller) OrderStatus(opts *bind.CallOpts, orderId [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "orderStatus", orderId)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OrderStatus is a free data retrieval call binding the contract method 0x2dff692d.
//
// Solidity: function orderStatus(bytes32 orderId) view returns(bytes32 status)
func (_Hyperlane7683 *Hyperlane7683Session) OrderStatus(orderId [32]byte) ([32]byte, error) {
	return _Hyperlane7683.Contract.OrderStatus(&_Hyperlane7683.CallOpts, orderId)
}

// OrderStatus is a free data retrieval call binding the contract method 0x2dff692d.
//
// Solidity: function orderStatus(bytes32 orderId) view returns(bytes32 status)
func (_Hyperlane7683 *Hyperlane7683CallerSession) OrderStatus(orderId [32]byte) ([32]byte, error) {
	return _Hyperlane7683.Contract.OrderStatus(&_Hyperlane7683.CallOpts, orderId)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Caller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Hyperlane7683 *Hyperlane7683Session) Owner() (common.Address, error) {
	return _Hyperlane7683.Contract.Owner(&_Hyperlane7683.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Hyperlane7683 *Hyperlane7683CallerSession) Owner() (common.Address, error) {
	return _Hyperlane7683.Contract.Owner(&_Hyperlane7683.CallOpts)
}

// QuoteGasPayment is a free data retrieval call binding the contract method 0xf2ed8c53.
//
// Solidity: function quoteGasPayment(uint32 _destinationDomain) view returns(uint256)
func (_Hyperlane7683 *Hyperlane7683Caller) QuoteGasPayment(opts *bind.CallOpts, _destinationDomain uint32) (*big.Int, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "quoteGasPayment", _destinationDomain)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// QuoteGasPayment is a free data retrieval call binding the contract method 0xf2ed8c53.
//
// Solidity: function quoteGasPayment(uint32 _destinationDomain) view returns(uint256)
func (_Hyperlane7683 *Hyperlane7683Session) QuoteGasPayment(_destinationDomain uint32) (*big.Int, error) {
	return _Hyperlane7683.Contract.QuoteGasPayment(&_Hyperlane7683.CallOpts, _destinationDomain)
}

// QuoteGasPayment is a free data retrieval call binding the contract method 0xf2ed8c53.
//
// Solidity: function quoteGasPayment(uint32 _destinationDomain) view returns(uint256)
func (_Hyperlane7683 *Hyperlane7683CallerSession) QuoteGasPayment(_destinationDomain uint32) (*big.Int, error) {
	return _Hyperlane7683.Contract.QuoteGasPayment(&_Hyperlane7683.CallOpts, _destinationDomain)
}

// Resolve is a free data retrieval call binding the contract method 0x41b477dd.
//
// Solidity: function resolve((uint32,bytes32,bytes) _order) view returns((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683Caller) Resolve(opts *bind.CallOpts, _order OnchainCrossChainOrder) (ResolvedCrossChainOrder, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "resolve", _order)

	if err != nil {
		return *new(ResolvedCrossChainOrder), err
	}

	out0 := *abi.ConvertType(out[0], new(ResolvedCrossChainOrder)).(*ResolvedCrossChainOrder)

	return out0, err

}

// Resolve is a free data retrieval call binding the contract method 0x41b477dd.
//
// Solidity: function resolve((uint32,bytes32,bytes) _order) view returns((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683Session) Resolve(_order OnchainCrossChainOrder) (ResolvedCrossChainOrder, error) {
	return _Hyperlane7683.Contract.Resolve(&_Hyperlane7683.CallOpts, _order)
}

// Resolve is a free data retrieval call binding the contract method 0x41b477dd.
//
// Solidity: function resolve((uint32,bytes32,bytes) _order) view returns((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683CallerSession) Resolve(_order OnchainCrossChainOrder) (ResolvedCrossChainOrder, error) {
	return _Hyperlane7683.Contract.Resolve(&_Hyperlane7683.CallOpts, _order)
}

// ResolveFor is a free data retrieval call binding the contract method 0x22bcd51a.
//
// Solidity: function resolveFor((address,address,uint256,uint256,uint32,uint32,bytes32,bytes) _order, bytes _originFillerData) view returns((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683Caller) ResolveFor(opts *bind.CallOpts, _order GaslessCrossChainOrder, _originFillerData []byte) (ResolvedCrossChainOrder, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "resolveFor", _order, _originFillerData)

	if err != nil {
		return *new(ResolvedCrossChainOrder), err
	}

	out0 := *abi.ConvertType(out[0], new(ResolvedCrossChainOrder)).(*ResolvedCrossChainOrder)

	return out0, err

}

// ResolveFor is a free data retrieval call binding the contract method 0x22bcd51a.
//
// Solidity: function resolveFor((address,address,uint256,uint256,uint32,uint32,bytes32,bytes) _order, bytes _originFillerData) view returns((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683Session) ResolveFor(_order GaslessCrossChainOrder, _originFillerData []byte) (ResolvedCrossChainOrder, error) {
	return _Hyperlane7683.Contract.ResolveFor(&_Hyperlane7683.CallOpts, _order, _originFillerData)
}

// ResolveFor is a free data retrieval call binding the contract method 0x22bcd51a.
//
// Solidity: function resolveFor((address,address,uint256,uint256,uint32,uint32,bytes32,bytes) _order, bytes _originFillerData) view returns((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683CallerSession) ResolveFor(_order GaslessCrossChainOrder, _originFillerData []byte) (ResolvedCrossChainOrder, error) {
	return _Hyperlane7683.Contract.ResolveFor(&_Hyperlane7683.CallOpts, _order, _originFillerData)
}

// Routers is a free data retrieval call binding the contract method 0x2ead72f6.
//
// Solidity: function routers(uint32 _domain) view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) Routers(opts *bind.CallOpts, _domain uint32) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "routers", _domain)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Routers is a free data retrieval call binding the contract method 0x2ead72f6.
//
// Solidity: function routers(uint32 _domain) view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) Routers(_domain uint32) ([32]byte, error) {
	return _Hyperlane7683.Contract.Routers(&_Hyperlane7683.CallOpts, _domain)
}

// Routers is a free data retrieval call binding the contract method 0x2ead72f6.
//
// Solidity: function routers(uint32 _domain) view returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) Routers(_domain uint32) ([32]byte, error) {
	return _Hyperlane7683.Contract.Routers(&_Hyperlane7683.CallOpts, _domain)
}

// UsedNonces is a free data retrieval call binding the contract method 0x6a8a6894.
//
// Solidity: function usedNonces(address , uint256 ) view returns(bool)
func (_Hyperlane7683 *Hyperlane7683Caller) UsedNonces(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (bool, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "usedNonces", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// UsedNonces is a free data retrieval call binding the contract method 0x6a8a6894.
//
// Solidity: function usedNonces(address , uint256 ) view returns(bool)
func (_Hyperlane7683 *Hyperlane7683Session) UsedNonces(arg0 common.Address, arg1 *big.Int) (bool, error) {
	return _Hyperlane7683.Contract.UsedNonces(&_Hyperlane7683.CallOpts, arg0, arg1)
}

// UsedNonces is a free data retrieval call binding the contract method 0x6a8a6894.
//
// Solidity: function usedNonces(address , uint256 ) view returns(bool)
func (_Hyperlane7683 *Hyperlane7683CallerSession) UsedNonces(arg0 common.Address, arg1 *big.Int) (bool, error) {
	return _Hyperlane7683.Contract.UsedNonces(&_Hyperlane7683.CallOpts, arg0, arg1)
}

// WitnessHash is a free data retrieval call binding the contract method 0xb2aa6daf.
//
// Solidity: function witnessHash((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder) pure returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Caller) WitnessHash(opts *bind.CallOpts, _resolvedOrder ResolvedCrossChainOrder) ([32]byte, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "witnessHash", _resolvedOrder)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// WitnessHash is a free data retrieval call binding the contract method 0xb2aa6daf.
//
// Solidity: function witnessHash((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder) pure returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683Session) WitnessHash(_resolvedOrder ResolvedCrossChainOrder) ([32]byte, error) {
	return _Hyperlane7683.Contract.WitnessHash(&_Hyperlane7683.CallOpts, _resolvedOrder)
}

// WitnessHash is a free data retrieval call binding the contract method 0xb2aa6daf.
//
// Solidity: function witnessHash((address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) _resolvedOrder) pure returns(bytes32)
func (_Hyperlane7683 *Hyperlane7683CallerSession) WitnessHash(_resolvedOrder ResolvedCrossChainOrder) ([32]byte, error) {
	return _Hyperlane7683.Contract.WitnessHash(&_Hyperlane7683.CallOpts, _resolvedOrder)
}

// WitnessTypeString is a free data retrieval call binding the contract method 0x74b9e838.
//
// Solidity: function witnessTypeString() view returns(string)
func (_Hyperlane7683 *Hyperlane7683Caller) WitnessTypeString(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Hyperlane7683.contract.Call(opts, &out, "witnessTypeString")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// WitnessTypeString is a free data retrieval call binding the contract method 0x74b9e838.
//
// Solidity: function witnessTypeString() view returns(string)
func (_Hyperlane7683 *Hyperlane7683Session) WitnessTypeString() (string, error) {
	return _Hyperlane7683.Contract.WitnessTypeString(&_Hyperlane7683.CallOpts)
}

// WitnessTypeString is a free data retrieval call binding the contract method 0x74b9e838.
//
// Solidity: function witnessTypeString() view returns(string)
func (_Hyperlane7683 *Hyperlane7683CallerSession) WitnessTypeString() (string, error) {
	return _Hyperlane7683.Contract.WitnessTypeString(&_Hyperlane7683.CallOpts)
}

// EnrollRemoteRouter is a paid mutator transaction binding the contract method 0xb49c53a7.
//
// Solidity: function enrollRemoteRouter(uint32 _domain, bytes32 _router) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) EnrollRemoteRouter(opts *bind.TransactOpts, _domain uint32, _router [32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "enrollRemoteRouter", _domain, _router)
}

// EnrollRemoteRouter is a paid mutator transaction binding the contract method 0xb49c53a7.
//
// Solidity: function enrollRemoteRouter(uint32 _domain, bytes32 _router) returns()
func (_Hyperlane7683 *Hyperlane7683Session) EnrollRemoteRouter(_domain uint32, _router [32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.EnrollRemoteRouter(&_Hyperlane7683.TransactOpts, _domain, _router)
}

// EnrollRemoteRouter is a paid mutator transaction binding the contract method 0xb49c53a7.
//
// Solidity: function enrollRemoteRouter(uint32 _domain, bytes32 _router) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) EnrollRemoteRouter(_domain uint32, _router [32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.EnrollRemoteRouter(&_Hyperlane7683.TransactOpts, _domain, _router)
}

// EnrollRemoteRouters is a paid mutator transaction binding the contract method 0xe9198bf9.
//
// Solidity: function enrollRemoteRouters(uint32[] _domains, bytes32[] _addresses) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) EnrollRemoteRouters(opts *bind.TransactOpts, _domains []uint32, _addresses [][32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "enrollRemoteRouters", _domains, _addresses)
}

// EnrollRemoteRouters is a paid mutator transaction binding the contract method 0xe9198bf9.
//
// Solidity: function enrollRemoteRouters(uint32[] _domains, bytes32[] _addresses) returns()
func (_Hyperlane7683 *Hyperlane7683Session) EnrollRemoteRouters(_domains []uint32, _addresses [][32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.EnrollRemoteRouters(&_Hyperlane7683.TransactOpts, _domains, _addresses)
}

// EnrollRemoteRouters is a paid mutator transaction binding the contract method 0xe9198bf9.
//
// Solidity: function enrollRemoteRouters(uint32[] _domains, bytes32[] _addresses) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) EnrollRemoteRouters(_domains []uint32, _addresses [][32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.EnrollRemoteRouters(&_Hyperlane7683.TransactOpts, _domains, _addresses)
}

// Fill is a paid mutator transaction binding the contract method 0x82e2c43f.
//
// Solidity: function fill(bytes32 _orderId, bytes _originData, bytes _fillerData) payable returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) Fill(opts *bind.TransactOpts, _orderId [32]byte, _originData []byte, _fillerData []byte) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "fill", _orderId, _originData, _fillerData)
}

// Fill is a paid mutator transaction binding the contract method 0x82e2c43f.
//
// Solidity: function fill(bytes32 _orderId, bytes _originData, bytes _fillerData) payable returns()
func (_Hyperlane7683 *Hyperlane7683Session) Fill(_orderId [32]byte, _originData []byte, _fillerData []byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Fill(&_Hyperlane7683.TransactOpts, _orderId, _originData, _fillerData)
}

// Fill is a paid mutator transaction binding the contract method 0x82e2c43f.
//
// Solidity: function fill(bytes32 _orderId, bytes _originData, bytes _fillerData) payable returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) Fill(_orderId [32]byte, _originData []byte, _fillerData []byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Fill(&_Hyperlane7683.TransactOpts, _orderId, _originData, _fillerData)
}

// Handle is a paid mutator transaction binding the contract method 0x56d5d475.
//
// Solidity: function handle(uint32 _origin, bytes32 _sender, bytes _message) payable returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) Handle(opts *bind.TransactOpts, _origin uint32, _sender [32]byte, _message []byte) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "handle", _origin, _sender, _message)
}

// Handle is a paid mutator transaction binding the contract method 0x56d5d475.
//
// Solidity: function handle(uint32 _origin, bytes32 _sender, bytes _message) payable returns()
func (_Hyperlane7683 *Hyperlane7683Session) Handle(_origin uint32, _sender [32]byte, _message []byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Handle(&_Hyperlane7683.TransactOpts, _origin, _sender, _message)
}

// Handle is a paid mutator transaction binding the contract method 0x56d5d475.
//
// Solidity: function handle(uint32 _origin, bytes32 _sender, bytes _message) payable returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) Handle(_origin uint32, _sender [32]byte, _message []byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Handle(&_Hyperlane7683.TransactOpts, _origin, _sender, _message)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _customHook, address _interchainSecurityModule, address _owner) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) Initialize(opts *bind.TransactOpts, _customHook common.Address, _interchainSecurityModule common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "initialize", _customHook, _interchainSecurityModule, _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _customHook, address _interchainSecurityModule, address _owner) returns()
func (_Hyperlane7683 *Hyperlane7683Session) Initialize(_customHook common.Address, _interchainSecurityModule common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Initialize(&_Hyperlane7683.TransactOpts, _customHook, _interchainSecurityModule, _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _customHook, address _interchainSecurityModule, address _owner) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) Initialize(_customHook common.Address, _interchainSecurityModule common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Initialize(&_Hyperlane7683.TransactOpts, _customHook, _interchainSecurityModule, _owner)
}

// InvalidateNonces is a paid mutator transaction binding the contract method 0x22f888e7.
//
// Solidity: function invalidateNonces(uint256 _nonce) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) InvalidateNonces(opts *bind.TransactOpts, _nonce *big.Int) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "invalidateNonces", _nonce)
}

// InvalidateNonces is a paid mutator transaction binding the contract method 0x22f888e7.
//
// Solidity: function invalidateNonces(uint256 _nonce) returns()
func (_Hyperlane7683 *Hyperlane7683Session) InvalidateNonces(_nonce *big.Int) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.InvalidateNonces(&_Hyperlane7683.TransactOpts, _nonce)
}

// InvalidateNonces is a paid mutator transaction binding the contract method 0x22f888e7.
//
// Solidity: function invalidateNonces(uint256 _nonce) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) InvalidateNonces(_nonce *big.Int) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.InvalidateNonces(&_Hyperlane7683.TransactOpts, _nonce)
}

// Open is a paid mutator transaction binding the contract method 0xe917a962.
//
// Solidity: function open((uint32,bytes32,bytes) _order) payable returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) Open(opts *bind.TransactOpts, _order OnchainCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "open", _order)
}

// Open is a paid mutator transaction binding the contract method 0xe917a962.
//
// Solidity: function open((uint32,bytes32,bytes) _order) payable returns()
func (_Hyperlane7683 *Hyperlane7683Session) Open(_order OnchainCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Open(&_Hyperlane7683.TransactOpts, _order)
}

// Open is a paid mutator transaction binding the contract method 0xe917a962.
//
// Solidity: function open((uint32,bytes32,bytes) _order) payable returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) Open(_order OnchainCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Open(&_Hyperlane7683.TransactOpts, _order)
}

// OpenFor is a paid mutator transaction binding the contract method 0x844fac8e.
//
// Solidity: function openFor((address,address,uint256,uint256,uint32,uint32,bytes32,bytes) _order, bytes _signature, bytes _originFillerData) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) OpenFor(opts *bind.TransactOpts, _order GaslessCrossChainOrder, _signature []byte, _originFillerData []byte) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "openFor", _order, _signature, _originFillerData)
}

// OpenFor is a paid mutator transaction binding the contract method 0x844fac8e.
//
// Solidity: function openFor((address,address,uint256,uint256,uint32,uint32,bytes32,bytes) _order, bytes _signature, bytes _originFillerData) returns()
func (_Hyperlane7683 *Hyperlane7683Session) OpenFor(_order GaslessCrossChainOrder, _signature []byte, _originFillerData []byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.OpenFor(&_Hyperlane7683.TransactOpts, _order, _signature, _originFillerData)
}

// OpenFor is a paid mutator transaction binding the contract method 0x844fac8e.
//
// Solidity: function openFor((address,address,uint256,uint256,uint32,uint32,bytes32,bytes) _order, bytes _signature, bytes _originFillerData) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) OpenFor(_order GaslessCrossChainOrder, _signature []byte, _originFillerData []byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.OpenFor(&_Hyperlane7683.TransactOpts, _order, _signature, _originFillerData)
}

// Refund is a paid mutator transaction binding the contract method 0x0cbd66e3.
//
// Solidity: function refund((uint32,bytes32,bytes)[] _orders) payable returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) Refund(opts *bind.TransactOpts, _orders []OnchainCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "refund", _orders)
}

// Refund is a paid mutator transaction binding the contract method 0x0cbd66e3.
//
// Solidity: function refund((uint32,bytes32,bytes)[] _orders) payable returns()
func (_Hyperlane7683 *Hyperlane7683Session) Refund(_orders []OnchainCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Refund(&_Hyperlane7683.TransactOpts, _orders)
}

// Refund is a paid mutator transaction binding the contract method 0x0cbd66e3.
//
// Solidity: function refund((uint32,bytes32,bytes)[] _orders) payable returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) Refund(_orders []OnchainCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Refund(&_Hyperlane7683.TransactOpts, _orders)
}

// Refund0 is a paid mutator transaction binding the contract method 0xe92971f5.
//
// Solidity: function refund((address,address,uint256,uint256,uint32,uint32,bytes32,bytes)[] _orders) payable returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) Refund0(opts *bind.TransactOpts, _orders []GaslessCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "refund0", _orders)
}

// Refund0 is a paid mutator transaction binding the contract method 0xe92971f5.
//
// Solidity: function refund((address,address,uint256,uint256,uint32,uint32,bytes32,bytes)[] _orders) payable returns()
func (_Hyperlane7683 *Hyperlane7683Session) Refund0(_orders []GaslessCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Refund0(&_Hyperlane7683.TransactOpts, _orders)
}

// Refund0 is a paid mutator transaction binding the contract method 0xe92971f5.
//
// Solidity: function refund((address,address,uint256,uint256,uint32,uint32,bytes32,bytes)[] _orders) payable returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) Refund0(_orders []GaslessCrossChainOrder) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Refund0(&_Hyperlane7683.TransactOpts, _orders)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Hyperlane7683 *Hyperlane7683Session) RenounceOwnership() (*types.Transaction, error) {
	return _Hyperlane7683.Contract.RenounceOwnership(&_Hyperlane7683.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Hyperlane7683.Contract.RenounceOwnership(&_Hyperlane7683.TransactOpts)
}

// SetDestinationGas is a paid mutator transaction binding the contract method 0x49d462ef.
//
// Solidity: function setDestinationGas(uint32 domain, uint256 gas) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) SetDestinationGas(opts *bind.TransactOpts, domain uint32, gas *big.Int) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "setDestinationGas", domain, gas)
}

// SetDestinationGas is a paid mutator transaction binding the contract method 0x49d462ef.
//
// Solidity: function setDestinationGas(uint32 domain, uint256 gas) returns()
func (_Hyperlane7683 *Hyperlane7683Session) SetDestinationGas(domain uint32, gas *big.Int) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetDestinationGas(&_Hyperlane7683.TransactOpts, domain, gas)
}

// SetDestinationGas is a paid mutator transaction binding the contract method 0x49d462ef.
//
// Solidity: function setDestinationGas(uint32 domain, uint256 gas) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) SetDestinationGas(domain uint32, gas *big.Int) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetDestinationGas(&_Hyperlane7683.TransactOpts, domain, gas)
}

// SetDestinationGas0 is a paid mutator transaction binding the contract method 0xb1bd6436.
//
// Solidity: function setDestinationGas((uint32,uint256)[] gasConfigs) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) SetDestinationGas0(opts *bind.TransactOpts, gasConfigs []GasRouterGasRouterConfig) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "setDestinationGas0", gasConfigs)
}

// SetDestinationGas0 is a paid mutator transaction binding the contract method 0xb1bd6436.
//
// Solidity: function setDestinationGas((uint32,uint256)[] gasConfigs) returns()
func (_Hyperlane7683 *Hyperlane7683Session) SetDestinationGas0(gasConfigs []GasRouterGasRouterConfig) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetDestinationGas0(&_Hyperlane7683.TransactOpts, gasConfigs)
}

// SetDestinationGas0 is a paid mutator transaction binding the contract method 0xb1bd6436.
//
// Solidity: function setDestinationGas((uint32,uint256)[] gasConfigs) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) SetDestinationGas0(gasConfigs []GasRouterGasRouterConfig) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetDestinationGas0(&_Hyperlane7683.TransactOpts, gasConfigs)
}

// SetHook is a paid mutator transaction binding the contract method 0x3dfd3873.
//
// Solidity: function setHook(address _hook) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) SetHook(opts *bind.TransactOpts, _hook common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "setHook", _hook)
}

// SetHook is a paid mutator transaction binding the contract method 0x3dfd3873.
//
// Solidity: function setHook(address _hook) returns()
func (_Hyperlane7683 *Hyperlane7683Session) SetHook(_hook common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetHook(&_Hyperlane7683.TransactOpts, _hook)
}

// SetHook is a paid mutator transaction binding the contract method 0x3dfd3873.
//
// Solidity: function setHook(address _hook) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) SetHook(_hook common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetHook(&_Hyperlane7683.TransactOpts, _hook)
}

// SetInterchainSecurityModule is a paid mutator transaction binding the contract method 0x0e72cc06.
//
// Solidity: function setInterchainSecurityModule(address _module) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) SetInterchainSecurityModule(opts *bind.TransactOpts, _module common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "setInterchainSecurityModule", _module)
}

// SetInterchainSecurityModule is a paid mutator transaction binding the contract method 0x0e72cc06.
//
// Solidity: function setInterchainSecurityModule(address _module) returns()
func (_Hyperlane7683 *Hyperlane7683Session) SetInterchainSecurityModule(_module common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetInterchainSecurityModule(&_Hyperlane7683.TransactOpts, _module)
}

// SetInterchainSecurityModule is a paid mutator transaction binding the contract method 0x0e72cc06.
//
// Solidity: function setInterchainSecurityModule(address _module) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) SetInterchainSecurityModule(_module common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.SetInterchainSecurityModule(&_Hyperlane7683.TransactOpts, _module)
}

// Settle is a paid mutator transaction binding the contract method 0xe7f921a2.
//
// Solidity: function settle(bytes32[] _orderIds) payable returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) Settle(opts *bind.TransactOpts, _orderIds [][32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "settle", _orderIds)
}

// Settle is a paid mutator transaction binding the contract method 0xe7f921a2.
//
// Solidity: function settle(bytes32[] _orderIds) payable returns()
func (_Hyperlane7683 *Hyperlane7683Session) Settle(_orderIds [][32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Settle(&_Hyperlane7683.TransactOpts, _orderIds)
}

// Settle is a paid mutator transaction binding the contract method 0xe7f921a2.
//
// Solidity: function settle(bytes32[] _orderIds) payable returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) Settle(_orderIds [][32]byte) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.Settle(&_Hyperlane7683.TransactOpts, _orderIds)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Hyperlane7683 *Hyperlane7683Session) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.TransferOwnership(&_Hyperlane7683.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.TransferOwnership(&_Hyperlane7683.TransactOpts, newOwner)
}

// UnenrollRemoteRouter is a paid mutator transaction binding the contract method 0xefae508a.
//
// Solidity: function unenrollRemoteRouter(uint32 _domain) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) UnenrollRemoteRouter(opts *bind.TransactOpts, _domain uint32) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "unenrollRemoteRouter", _domain)
}

// UnenrollRemoteRouter is a paid mutator transaction binding the contract method 0xefae508a.
//
// Solidity: function unenrollRemoteRouter(uint32 _domain) returns()
func (_Hyperlane7683 *Hyperlane7683Session) UnenrollRemoteRouter(_domain uint32) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.UnenrollRemoteRouter(&_Hyperlane7683.TransactOpts, _domain)
}

// UnenrollRemoteRouter is a paid mutator transaction binding the contract method 0xefae508a.
//
// Solidity: function unenrollRemoteRouter(uint32 _domain) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) UnenrollRemoteRouter(_domain uint32) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.UnenrollRemoteRouter(&_Hyperlane7683.TransactOpts, _domain)
}

// UnenrollRemoteRouters is a paid mutator transaction binding the contract method 0x71a15b38.
//
// Solidity: function unenrollRemoteRouters(uint32[] _domains) returns()
func (_Hyperlane7683 *Hyperlane7683Transactor) UnenrollRemoteRouters(opts *bind.TransactOpts, _domains []uint32) (*types.Transaction, error) {
	return _Hyperlane7683.contract.Transact(opts, "unenrollRemoteRouters", _domains)
}

// UnenrollRemoteRouters is a paid mutator transaction binding the contract method 0x71a15b38.
//
// Solidity: function unenrollRemoteRouters(uint32[] _domains) returns()
func (_Hyperlane7683 *Hyperlane7683Session) UnenrollRemoteRouters(_domains []uint32) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.UnenrollRemoteRouters(&_Hyperlane7683.TransactOpts, _domains)
}

// UnenrollRemoteRouters is a paid mutator transaction binding the contract method 0x71a15b38.
//
// Solidity: function unenrollRemoteRouters(uint32[] _domains) returns()
func (_Hyperlane7683 *Hyperlane7683TransactorSession) UnenrollRemoteRouters(_domains []uint32) (*types.Transaction, error) {
	return _Hyperlane7683.Contract.UnenrollRemoteRouters(&_Hyperlane7683.TransactOpts, _domains)
}

// Hyperlane7683FilledIterator is returned from FilterFilled and is used to iterate over the raw logs and unpacked data for Filled events raised by the Hyperlane7683 contract.
type Hyperlane7683FilledIterator struct {
	Event *Hyperlane7683Filled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683FilledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683Filled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683Filled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683FilledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683FilledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683Filled represents a Filled event raised by the Hyperlane7683 contract.
type Hyperlane7683Filled struct {
	OrderId    [32]byte
	OriginData []byte
	FillerData []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterFilled is a free log retrieval operation binding the contract event 0x57f1f65270c1c2c1771948825ee86f8d23d11ab44b16eb9c213056e042d06e59.
//
// Solidity: event Filled(bytes32 orderId, bytes originData, bytes fillerData)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterFilled(opts *bind.FilterOpts) (*Hyperlane7683FilledIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "Filled")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683FilledIterator{contract: _Hyperlane7683.contract, event: "Filled", logs: logs, sub: sub}, nil
}

// WatchFilled is a free log subscription operation binding the contract event 0x57f1f65270c1c2c1771948825ee86f8d23d11ab44b16eb9c213056e042d06e59.
//
// Solidity: event Filled(bytes32 orderId, bytes originData, bytes fillerData)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchFilled(opts *bind.WatchOpts, sink chan<- *Hyperlane7683Filled) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "Filled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683Filled)
				if err := _Hyperlane7683.contract.UnpackLog(event, "Filled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFilled is a log parse operation binding the contract event 0x57f1f65270c1c2c1771948825ee86f8d23d11ab44b16eb9c213056e042d06e59.
//
// Solidity: event Filled(bytes32 orderId, bytes originData, bytes fillerData)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseFilled(log types.Log) (*Hyperlane7683Filled, error) {
	event := new(Hyperlane7683Filled)
	if err := _Hyperlane7683.contract.UnpackLog(event, "Filled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683GasSetIterator is returned from FilterGasSet and is used to iterate over the raw logs and unpacked data for GasSet events raised by the Hyperlane7683 contract.
type Hyperlane7683GasSetIterator struct {
	Event *Hyperlane7683GasSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683GasSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683GasSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683GasSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683GasSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683GasSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683GasSet represents a GasSet event raised by the Hyperlane7683 contract.
type Hyperlane7683GasSet struct {
	Domain uint32
	Gas    *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterGasSet is a free log retrieval operation binding the contract event 0xc3de732a98b24a2b5c6f67e8a7fb057ffc14046b83968a2c73e4148d2fba978b.
//
// Solidity: event GasSet(uint32 domain, uint256 gas)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterGasSet(opts *bind.FilterOpts) (*Hyperlane7683GasSetIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "GasSet")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683GasSetIterator{contract: _Hyperlane7683.contract, event: "GasSet", logs: logs, sub: sub}, nil
}

// WatchGasSet is a free log subscription operation binding the contract event 0xc3de732a98b24a2b5c6f67e8a7fb057ffc14046b83968a2c73e4148d2fba978b.
//
// Solidity: event GasSet(uint32 domain, uint256 gas)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchGasSet(opts *bind.WatchOpts, sink chan<- *Hyperlane7683GasSet) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "GasSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683GasSet)
				if err := _Hyperlane7683.contract.UnpackLog(event, "GasSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseGasSet is a log parse operation binding the contract event 0xc3de732a98b24a2b5c6f67e8a7fb057ffc14046b83968a2c73e4148d2fba978b.
//
// Solidity: event GasSet(uint32 domain, uint256 gas)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseGasSet(log types.Log) (*Hyperlane7683GasSet, error) {
	event := new(Hyperlane7683GasSet)
	if err := _Hyperlane7683.contract.UnpackLog(event, "GasSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683HookSetIterator is returned from FilterHookSet and is used to iterate over the raw logs and unpacked data for HookSet events raised by the Hyperlane7683 contract.
type Hyperlane7683HookSetIterator struct {
	Event *Hyperlane7683HookSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683HookSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683HookSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683HookSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683HookSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683HookSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683HookSet represents a HookSet event raised by the Hyperlane7683 contract.
type Hyperlane7683HookSet struct {
	Hook common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterHookSet is a free log retrieval operation binding the contract event 0x4eab7b127c764308788622363ad3e9532de3dfba7845bd4f84c125a22544255a.
//
// Solidity: event HookSet(address _hook)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterHookSet(opts *bind.FilterOpts) (*Hyperlane7683HookSetIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "HookSet")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683HookSetIterator{contract: _Hyperlane7683.contract, event: "HookSet", logs: logs, sub: sub}, nil
}

// WatchHookSet is a free log subscription operation binding the contract event 0x4eab7b127c764308788622363ad3e9532de3dfba7845bd4f84c125a22544255a.
//
// Solidity: event HookSet(address _hook)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchHookSet(opts *bind.WatchOpts, sink chan<- *Hyperlane7683HookSet) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "HookSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683HookSet)
				if err := _Hyperlane7683.contract.UnpackLog(event, "HookSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseHookSet is a log parse operation binding the contract event 0x4eab7b127c764308788622363ad3e9532de3dfba7845bd4f84c125a22544255a.
//
// Solidity: event HookSet(address _hook)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseHookSet(log types.Log) (*Hyperlane7683HookSet, error) {
	event := new(Hyperlane7683HookSet)
	if err := _Hyperlane7683.contract.UnpackLog(event, "HookSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683InitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Hyperlane7683 contract.
type Hyperlane7683InitializedIterator struct {
	Event *Hyperlane7683Initialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683InitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683Initialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683Initialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683InitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683InitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683Initialized represents a Initialized event raised by the Hyperlane7683 contract.
type Hyperlane7683Initialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterInitialized(opts *bind.FilterOpts) (*Hyperlane7683InitializedIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683InitializedIterator{contract: _Hyperlane7683.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *Hyperlane7683Initialized) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683Initialized)
				if err := _Hyperlane7683.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseInitialized(log types.Log) (*Hyperlane7683Initialized, error) {
	event := new(Hyperlane7683Initialized)
	if err := _Hyperlane7683.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683IsmSetIterator is returned from FilterIsmSet and is used to iterate over the raw logs and unpacked data for IsmSet events raised by the Hyperlane7683 contract.
type Hyperlane7683IsmSetIterator struct {
	Event *Hyperlane7683IsmSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683IsmSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683IsmSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683IsmSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683IsmSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683IsmSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683IsmSet represents a IsmSet event raised by the Hyperlane7683 contract.
type Hyperlane7683IsmSet struct {
	Ism common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterIsmSet is a free log retrieval operation binding the contract event 0xc47cbcc588c67679e52261c45cc315e56562f8d0ccaba16facb9093ff9498799.
//
// Solidity: event IsmSet(address _ism)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterIsmSet(opts *bind.FilterOpts) (*Hyperlane7683IsmSetIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "IsmSet")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683IsmSetIterator{contract: _Hyperlane7683.contract, event: "IsmSet", logs: logs, sub: sub}, nil
}

// WatchIsmSet is a free log subscription operation binding the contract event 0xc47cbcc588c67679e52261c45cc315e56562f8d0ccaba16facb9093ff9498799.
//
// Solidity: event IsmSet(address _ism)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchIsmSet(opts *bind.WatchOpts, sink chan<- *Hyperlane7683IsmSet) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "IsmSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683IsmSet)
				if err := _Hyperlane7683.contract.UnpackLog(event, "IsmSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseIsmSet is a log parse operation binding the contract event 0xc47cbcc588c67679e52261c45cc315e56562f8d0ccaba16facb9093ff9498799.
//
// Solidity: event IsmSet(address _ism)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseIsmSet(log types.Log) (*Hyperlane7683IsmSet, error) {
	event := new(Hyperlane7683IsmSet)
	if err := _Hyperlane7683.contract.UnpackLog(event, "IsmSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683NonceInvalidationIterator is returned from FilterNonceInvalidation and is used to iterate over the raw logs and unpacked data for NonceInvalidation events raised by the Hyperlane7683 contract.
type Hyperlane7683NonceInvalidationIterator struct {
	Event *Hyperlane7683NonceInvalidation // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683NonceInvalidationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683NonceInvalidation)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683NonceInvalidation)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683NonceInvalidationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683NonceInvalidationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683NonceInvalidation represents a NonceInvalidation event raised by the Hyperlane7683 contract.
type Hyperlane7683NonceInvalidation struct {
	Owner common.Address
	Nonce *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNonceInvalidation is a free log retrieval operation binding the contract event 0x239b0d63832ec06e8082928dc583392b57a30d16e4bb425b49d82c7808e308b0.
//
// Solidity: event NonceInvalidation(address indexed owner, uint256 nonce)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterNonceInvalidation(opts *bind.FilterOpts, owner []common.Address) (*Hyperlane7683NonceInvalidationIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "NonceInvalidation", ownerRule)
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683NonceInvalidationIterator{contract: _Hyperlane7683.contract, event: "NonceInvalidation", logs: logs, sub: sub}, nil
}

// WatchNonceInvalidation is a free log subscription operation binding the contract event 0x239b0d63832ec06e8082928dc583392b57a30d16e4bb425b49d82c7808e308b0.
//
// Solidity: event NonceInvalidation(address indexed owner, uint256 nonce)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchNonceInvalidation(opts *bind.WatchOpts, sink chan<- *Hyperlane7683NonceInvalidation, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "NonceInvalidation", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683NonceInvalidation)
				if err := _Hyperlane7683.contract.UnpackLog(event, "NonceInvalidation", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNonceInvalidation is a log parse operation binding the contract event 0x239b0d63832ec06e8082928dc583392b57a30d16e4bb425b49d82c7808e308b0.
//
// Solidity: event NonceInvalidation(address indexed owner, uint256 nonce)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseNonceInvalidation(log types.Log) (*Hyperlane7683NonceInvalidation, error) {
	event := new(Hyperlane7683NonceInvalidation)
	if err := _Hyperlane7683.contract.UnpackLog(event, "NonceInvalidation", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683OpenIterator is returned from FilterOpen and is used to iterate over the raw logs and unpacked data for Open events raised by the Hyperlane7683 contract.
type Hyperlane7683OpenIterator struct {
	Event *Hyperlane7683Open // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683OpenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683Open)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683Open)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683OpenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683OpenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683Open represents a Open event raised by the Hyperlane7683 contract.
type Hyperlane7683Open struct {
	OrderId       [32]byte
	ResolvedOrder ResolvedCrossChainOrder
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOpen is a free log retrieval operation binding the contract event 0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d.
//
// Solidity: event Open(bytes32 indexed orderId, (address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterOpen(opts *bind.FilterOpts, orderId [][32]byte) (*Hyperlane7683OpenIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "Open", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683OpenIterator{contract: _Hyperlane7683.contract, event: "Open", logs: logs, sub: sub}, nil
}

// WatchOpen is a free log subscription operation binding the contract event 0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d.
//
// Solidity: event Open(bytes32 indexed orderId, (address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchOpen(opts *bind.WatchOpts, sink chan<- *Hyperlane7683Open, orderId [][32]byte) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "Open", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683Open)
				if err := _Hyperlane7683.contract.UnpackLog(event, "Open", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOpen is a log parse operation binding the contract event 0x3448bbc2203c608599ad448eeb1007cea04b788ac631f9f558e8dd01a3c27b3d.
//
// Solidity: event Open(bytes32 indexed orderId, (address,uint256,uint32,uint32,bytes32,(bytes32,uint256,bytes32,uint256)[],(bytes32,uint256,bytes32,uint256)[],(uint256,bytes32,bytes)[]) resolvedOrder)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseOpen(log types.Log) (*Hyperlane7683Open, error) {
	event := new(Hyperlane7683Open)
	if err := _Hyperlane7683.contract.UnpackLog(event, "Open", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683OwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Hyperlane7683 contract.
type Hyperlane7683OwnershipTransferredIterator struct {
	Event *Hyperlane7683OwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683OwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683OwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683OwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683OwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683OwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683OwnershipTransferred represents a OwnershipTransferred event raised by the Hyperlane7683 contract.
type Hyperlane7683OwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*Hyperlane7683OwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683OwnershipTransferredIterator{contract: _Hyperlane7683.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *Hyperlane7683OwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683OwnershipTransferred)
				if err := _Hyperlane7683.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseOwnershipTransferred(log types.Log) (*Hyperlane7683OwnershipTransferred, error) {
	event := new(Hyperlane7683OwnershipTransferred)
	if err := _Hyperlane7683.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683RefundIterator is returned from FilterRefund and is used to iterate over the raw logs and unpacked data for Refund events raised by the Hyperlane7683 contract.
type Hyperlane7683RefundIterator struct {
	Event *Hyperlane7683Refund // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683RefundIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683Refund)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683Refund)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683RefundIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683RefundIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683Refund represents a Refund event raised by the Hyperlane7683 contract.
type Hyperlane7683Refund struct {
	OrderIds [][32]byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRefund is a free log retrieval operation binding the contract event 0x536286146d4af271695884c3088546e593d6165b08a2e29d706a87b7db74b201.
//
// Solidity: event Refund(bytes32[] orderIds)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterRefund(opts *bind.FilterOpts) (*Hyperlane7683RefundIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "Refund")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683RefundIterator{contract: _Hyperlane7683.contract, event: "Refund", logs: logs, sub: sub}, nil
}

// WatchRefund is a free log subscription operation binding the contract event 0x536286146d4af271695884c3088546e593d6165b08a2e29d706a87b7db74b201.
//
// Solidity: event Refund(bytes32[] orderIds)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchRefund(opts *bind.WatchOpts, sink chan<- *Hyperlane7683Refund) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "Refund")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683Refund)
				if err := _Hyperlane7683.contract.UnpackLog(event, "Refund", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRefund is a log parse operation binding the contract event 0x536286146d4af271695884c3088546e593d6165b08a2e29d706a87b7db74b201.
//
// Solidity: event Refund(bytes32[] orderIds)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseRefund(log types.Log) (*Hyperlane7683Refund, error) {
	event := new(Hyperlane7683Refund)
	if err := _Hyperlane7683.contract.UnpackLog(event, "Refund", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683RefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the Hyperlane7683 contract.
type Hyperlane7683RefundedIterator struct {
	Event *Hyperlane7683Refunded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683RefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683Refunded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683Refunded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683RefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683RefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683Refunded represents a Refunded event raised by the Hyperlane7683 contract.
type Hyperlane7683Refunded struct {
	OrderId  [32]byte
	Receiver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x5e9f0820fcfb53b644becb775b651bae68c337106f21433e526551d1e02c1c0e.
//
// Solidity: event Refunded(bytes32 orderId, address receiver)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterRefunded(opts *bind.FilterOpts) (*Hyperlane7683RefundedIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683RefundedIterator{contract: _Hyperlane7683.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x5e9f0820fcfb53b644becb775b651bae68c337106f21433e526551d1e02c1c0e.
//
// Solidity: event Refunded(bytes32 orderId, address receiver)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *Hyperlane7683Refunded) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "Refunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683Refunded)
				if err := _Hyperlane7683.contract.UnpackLog(event, "Refunded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRefunded is a log parse operation binding the contract event 0x5e9f0820fcfb53b644becb775b651bae68c337106f21433e526551d1e02c1c0e.
//
// Solidity: event Refunded(bytes32 orderId, address receiver)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseRefunded(log types.Log) (*Hyperlane7683Refunded, error) {
	event := new(Hyperlane7683Refunded)
	if err := _Hyperlane7683.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683SettleIterator is returned from FilterSettle and is used to iterate over the raw logs and unpacked data for Settle events raised by the Hyperlane7683 contract.
type Hyperlane7683SettleIterator struct {
	Event *Hyperlane7683Settle // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683SettleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683Settle)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683Settle)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683SettleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683SettleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683Settle represents a Settle event raised by the Hyperlane7683 contract.
type Hyperlane7683Settle struct {
	OrderIds         [][32]byte
	OrdersFillerData [][]byte
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterSettle is a free log retrieval operation binding the contract event 0xc993982486b74c2846b97472fb81c2ffe7b9b135062b8da6663d73886422fbad.
//
// Solidity: event Settle(bytes32[] orderIds, bytes[] ordersFillerData)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterSettle(opts *bind.FilterOpts) (*Hyperlane7683SettleIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "Settle")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683SettleIterator{contract: _Hyperlane7683.contract, event: "Settle", logs: logs, sub: sub}, nil
}

// WatchSettle is a free log subscription operation binding the contract event 0xc993982486b74c2846b97472fb81c2ffe7b9b135062b8da6663d73886422fbad.
//
// Solidity: event Settle(bytes32[] orderIds, bytes[] ordersFillerData)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchSettle(opts *bind.WatchOpts, sink chan<- *Hyperlane7683Settle) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "Settle")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683Settle)
				if err := _Hyperlane7683.contract.UnpackLog(event, "Settle", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSettle is a log parse operation binding the contract event 0xc993982486b74c2846b97472fb81c2ffe7b9b135062b8da6663d73886422fbad.
//
// Solidity: event Settle(bytes32[] orderIds, bytes[] ordersFillerData)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseSettle(log types.Log) (*Hyperlane7683Settle, error) {
	event := new(Hyperlane7683Settle)
	if err := _Hyperlane7683.contract.UnpackLog(event, "Settle", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hyperlane7683SettledIterator is returned from FilterSettled and is used to iterate over the raw logs and unpacked data for Settled events raised by the Hyperlane7683 contract.
type Hyperlane7683SettledIterator struct {
	Event *Hyperlane7683Settled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hyperlane7683SettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hyperlane7683Settled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hyperlane7683Settled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hyperlane7683SettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hyperlane7683SettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hyperlane7683Settled represents a Settled event raised by the Hyperlane7683 contract.
type Hyperlane7683Settled struct {
	OrderId  [32]byte
	Receiver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSettled is a free log retrieval operation binding the contract event 0xa569bfd2e3bd9bd14cfdabad61aef5f3d5b18b0fcdf78805e65349dda2210fbc.
//
// Solidity: event Settled(bytes32 orderId, address receiver)
func (_Hyperlane7683 *Hyperlane7683Filterer) FilterSettled(opts *bind.FilterOpts) (*Hyperlane7683SettledIterator, error) {

	logs, sub, err := _Hyperlane7683.contract.FilterLogs(opts, "Settled")
	if err != nil {
		return nil, err
	}
	return &Hyperlane7683SettledIterator{contract: _Hyperlane7683.contract, event: "Settled", logs: logs, sub: sub}, nil
}

// WatchSettled is a free log subscription operation binding the contract event 0xa569bfd2e3bd9bd14cfdabad61aef5f3d5b18b0fcdf78805e65349dda2210fbc.
//
// Solidity: event Settled(bytes32 orderId, address receiver)
func (_Hyperlane7683 *Hyperlane7683Filterer) WatchSettled(opts *bind.WatchOpts, sink chan<- *Hyperlane7683Settled) (event.Subscription, error) {

	logs, sub, err := _Hyperlane7683.contract.WatchLogs(opts, "Settled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hyperlane7683Settled)
				if err := _Hyperlane7683.contract.UnpackLog(event, "Settled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSettled is a log parse operation binding the contract event 0xa569bfd2e3bd9bd14cfdabad61aef5f3d5b18b0fcdf78805e65349dda2210fbc.
//
// Solidity: event Settled(bytes32 orderId, address receiver)
func (_Hyperlane7683 *Hyperlane7683Filterer) ParseSettled(log types.Log) (*Hyperlane7683Settled, error) {
	event := new(Hyperlane7683Settled)
	if err := _Hyperlane7683.contract.UnpackLog(event, "Settled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
