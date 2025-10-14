package base

import (
	"context"
	"fmt"

	"github.com/NethermindEth/oif-starknet/solver/solvercore/types"
)

// Rule represents a rule that can be evaluated during intent processing
type Rule func(args *types.ParsedArgs, context *SolverContext) error

// SolverContext contains context information for rule evaluation
type SolverContext struct {
	OriginInfo []string
	TargetInfo []string
	Metadata   interface{}
}

// Solver defines the interface for intent solvers
type Solver interface {
	// ProcessIntent processes an intent through the complete lifecycle
	// Returns (success, error) where success=true means the order was fully settled
	ProcessIntent(ctx context.Context, args *types.ParsedArgs, originChainName string, blockNumber uint64) (bool, error)

	// PrepareIntent evaluates rules and determines if intent should be filled
	PrepareIntent(ctx context.Context, args *types.ParsedArgs) (*types.Result[types.IntentData], error)

	// Fill executes the actual intent filling
	Fill(ctx context.Context, args *types.ParsedArgs, data types.IntentData, originChainName string, blockNumber uint64) error

	// SettleOrder handles post-fill settlement
	SettleOrder(ctx context.Context, args *types.ParsedArgs, data types.IntentData, originChainName string) error

	// AddRule adds a custom rule to the solver
	AddRule(rule Rule)

	// GetRules returns all rules
	GetRules() []Rule
}

// solverImpl provides a base implementation of Solver
type solverImpl struct {
	rules           []Rule
	allowBlockLists types.AllowBlockLists
	metadata        interface{}
}

// NewSolver creates a new base solver
func NewSolver(allowBlockLists types.AllowBlockLists, metadata interface{}) *solverImpl {
	return &solverImpl{
		rules:           make([]Rule, 0),
		allowBlockLists: allowBlockLists,
		metadata:        metadata,
	}
}

// AddRule adds a rule to the solver
func (f *solverImpl) AddRule(rule Rule) {
	f.rules = append(f.rules, rule)
}

// GetRules returns all rules
func (f *solverImpl) GetRules() []Rule {
	return f.rules
}

// ProcessIntent implements the complete intent processing lifecycle
func (f *solverImpl) ProcessIntent(ctx context.Context, args *types.ParsedArgs, originChainName string, blockNumber uint64) (bool, error) {
	// Step 1: Prepare intent (evaluate rules)
	intent, err := f.PrepareIntent(ctx, args)
	if err != nil {
		return false, err
	}

	if !intent.Success {
		return false, nil // Intent was rejected by rules
	}

	// Step 2: Fill the intent
	if err := f.Fill(ctx, args, intent.Data, originChainName, blockNumber); err != nil {
		return false, err
	}

	// Step 3: Settle the order
	if err := f.SettleOrder(ctx, args, intent.Data, originChainName); err != nil {
		return false, err
	}

	return true, nil // Successfully filled and settled
}

// PrepareIntent evaluates rules to determine if intent should be filled
func (f *solverImpl) PrepareIntent(ctx context.Context, args *types.ParsedArgs) (*types.Result[types.IntentData], error) {
	// Check allow/block lists first
	if !f.isAllowedIntent(args) {
		result := types.NewErrorResult[types.IntentData](fmt.Errorf("intent blocked by allow/block lists"))
		return &result, nil
	}

	// Evaluate all rules
	for _, rule := range f.rules {
		context := &SolverContext{
			Metadata:   f.metadata,
			OriginInfo: nil, // Not used in this context
			TargetInfo: nil, // Not used in this context
		}

		if err := rule(args, context); err != nil {
			result := types.NewErrorResult[types.IntentData](err)
			return &result, nil
		}
	}

	// If all rules pass, create intent data
	intentData := types.IntentData{
		FillInstructions: args.ResolvedOrder.FillInstructions,
		MaxSpent:         args.ResolvedOrder.MaxSpent,
	}

	result := types.NewSuccessResult(intentData)
	return &result, nil
}

// Fill executes the actual intent filling (to be implemented by concrete solvers)
func (f *solverImpl) Fill(ctx context.Context, args *types.ParsedArgs, data types.IntentData, originChainName string, blockNumber uint64) error {
	// This is a placeholder - concrete implementations should override this
	return nil
}

// SettleOrder handles post-fill settlement (to be implemented by concrete solvers)
func (f *solverImpl) SettleOrder(ctx context.Context, args *types.ParsedArgs, data types.IntentData, originChainName string) error {
	// This is a placeholder - concrete implementations should override this
	return nil
}

// isAllowedIntent checks if an intent is allowed based on allow/block lists
func (f *solverImpl) isAllowedIntent(args *types.ParsedArgs) bool {
	// Check block list first
	for _, blockItem := range f.allowBlockLists.BlockList {
		if f.matchesAllowBlockItem(blockItem, args) {
			return false
		}
	}

	// If no allow list is specified, allow everything
	if len(f.allowBlockLists.AllowList) == 0 {
		return true
	}

	// Check allow list
	for _, allowItem := range f.allowBlockLists.AllowList {
		if f.matchesAllowBlockItem(allowItem, args) {
			return true
		}
	}

	return false
}

// matchesAllowBlockItem checks if args match an allow/block list item
func (f *solverImpl) matchesAllowBlockItem(item types.AllowBlockListItem, args *types.ParsedArgs) bool {
	// Check sender address
	if item.SenderAddress != "*" && item.SenderAddress != args.SenderAddress {
		return false
	}

	// Check recipients
	for _, recipient := range args.Recipients {
		// Check destination domain
		if item.DestinationDomain != "*" && item.DestinationDomain != recipient.DestinationChainName {
			continue
		}

		// Check recipient address
		if item.RecipientAddress != "*" && item.RecipientAddress != recipient.RecipientAddress {
			continue
		}

		return true
	}

	return false
}
