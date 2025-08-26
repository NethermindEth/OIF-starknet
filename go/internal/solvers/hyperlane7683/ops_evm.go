package hyperlane7683

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/NethermindEth/oif-starknet/go/internal/deployer"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EVMOps provides approval and status helpers for EVM fills
type EVMOps struct {
	client *ethclient.Client
}

func NewEVMOps(client *ethclient.Client) *EVMOps { return &EVMOps{client: client} }

// EnsureApproval checks allowance(owner, spender) and approves max if insufficient
func (ops *EVMOps) EnsureApproval(ctx context.Context, owner common.Address, token, spender common.Address, amount *big.Int) error {
	parsedABI, err := abi.JSON(strings.NewReader(deployer.GetERC20Contract().ABI))
	if err != nil {
		return fmt.Errorf("erc20 abi parse failed: %w", err)
	}
	// allowance(owner, spender)
	callData, err := parsedABI.Pack("allowance", owner, spender)
	if err != nil {
		return fmt.Errorf("pack allowance failed: %w", err)
	}
	resp, err := ops.client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: callData}, nil)
	if err != nil {
		return fmt.Errorf("allowance call failed: %w", err)
	}
	if len(resp) < 32 {
		return fmt.Errorf("invalid allowance resp: %d", len(resp))
	}
	current := new(big.Int).SetBytes(resp)
	if current.Cmp(amount) >= 0 {
		return nil
	}
	// caller must handle signing/sending approve in filler where signer is available
	return fmt.Errorf("insufficient allowance: have %s need %s", current.String(), amount.String())
}
