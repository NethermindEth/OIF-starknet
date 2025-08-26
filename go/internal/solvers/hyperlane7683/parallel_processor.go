package hyperlane7683

import (
	"context"
	"fmt"
	"sync"

	"github.com/NethermindEth/oif-starknet/go/internal/types"
)

// ParallelProcessor handles concurrent processing of fills and approvals
// Following the TypeScript Promise.all pattern for better performance
type ParallelProcessor struct{}

// ProcessFillsInParallel executes multiple fill instructions concurrently
// This matches the TypeScript implementation that uses Promise.all for fills
func (pp *ParallelProcessor) ProcessFillsInParallel(
	ctx context.Context,
	args types.ParsedArgs,
	data types.IntentData,
	originChainName string,
	fillHandler func(ctx context.Context, instruction types.FillInstruction) error,
) error {
	if len(data.FillInstructions) == 0 {
		return fmt.Errorf("no fill instructions to process")
	}

	// For single instruction, no need for parallelization
	if len(data.FillInstructions) == 1 {
		return fillHandler(ctx, data.FillInstructions[0])
	}

	fmt.Printf("   🔄 Processing %d fill instructions in parallel\n", len(data.FillInstructions))

	var wg sync.WaitGroup
	errChan := make(chan error, len(data.FillInstructions))

	// Process each fill instruction in parallel
	for i, instruction := range data.FillInstructions {
		wg.Add(1)
		go func(idx int, instr types.FillInstruction) {
			defer wg.Done()
			
			fmt.Printf("   📦 Starting fill instruction %d/%d\n", idx+1, len(data.FillInstructions))
			
			if err := fillHandler(ctx, instr); err != nil {
				errChan <- fmt.Errorf("fill instruction %d failed: %w", idx+1, err)
				return
			}
			
			fmt.Printf("   ✅ Fill instruction %d/%d completed\n", idx+1, len(data.FillInstructions))
		}(i, instruction)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		return err // Return first error encountered
	}

	fmt.Printf("   🎉 All %d fill instructions processed successfully in parallel\n", len(data.FillInstructions))
	return nil
}

// ProcessApprovalsInParallel handles token approvals concurrently
// This matches the TypeScript implementation for handling approvals
func (pp *ParallelProcessor) ProcessApprovalsInParallel(
	ctx context.Context,
	maxSpent []types.Output,
	approvalHandler func(ctx context.Context, output types.Output) error,
) error {
	if len(maxSpent) == 0 {
		return nil // No approvals needed
	}

	// For single approval, no need for parallelization
	if len(maxSpent) == 1 {
		return approvalHandler(ctx, maxSpent[0])
	}

	fmt.Printf("   🔄 Processing %d token approvals in parallel\n", len(maxSpent))

	var wg sync.WaitGroup
	errChan := make(chan error, len(maxSpent))

	// Process each approval in parallel
	for i, output := range maxSpent {
		wg.Add(1)
		go func(idx int, out types.Output) {
			defer wg.Done()
			
			fmt.Printf("   💰 Starting approval %d/%d for token %s\n", idx+1, len(maxSpent), out.Token.Hex())
			
			if err := approvalHandler(ctx, out); err != nil {
				errChan <- fmt.Errorf("approval %d failed for token %s: %w", idx+1, out.Token.Hex(), err)
				return
			}
			
			fmt.Printf("   ✅ Approval %d/%d completed for token %s\n", idx+1, len(maxSpent), out.Token.Hex())
		}(i, output)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		return err // Return first error encountered
	}

	fmt.Printf("   🎉 All %d token approvals processed successfully in parallel\n", len(maxSpent))
	return nil
}

// ProcessWithTimeout adds timeout handling to parallel operations
// This provides additional safety for operations that might hang
func (pp *ParallelProcessor) ProcessWithTimeout(
	ctx context.Context,
	operation func(ctx context.Context) error,
	timeoutMsg string,
) error {
	errChan := make(chan error, 1)
	
	go func() {
		errChan <- operation(ctx)
	}()
	
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%s timed out: %w", timeoutMsg, ctx.Err())
	}
}
