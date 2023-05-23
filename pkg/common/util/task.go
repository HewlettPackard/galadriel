package util

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
)

// RunTasks concurrently executes multiple tasks and ensures resilience by handling errors and panics.
// It creates a cancelable context for the tasks, launches them as separate goroutines, and captures
// any panics or errors that occur. The function waits for all tasks to complete or until an error occurs.
// If a task is canceled due to the parent context being canceled, it returns the corresponding error.
// If any task returns an error, that error is returned. Otherwise, it returns nil to indicate success.
// This function is useful for concurrently executing independent tasks while maintaining application resilience.
// Errors and panics are logged using the standard log package.
func RunTasks(ctx context.Context, tasks ...func(context.Context) error) error {
	var wg sync.WaitGroup
	wg.Add(len(tasks))

	errCh := make(chan error, len(tasks))

	// Create a cancelable context for all the tasks
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, task := range tasks {
		go func(task func(context.Context) error) {
			defer wg.Done()
			var panicErr error
			err := func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						panicErr = fmt.Errorf("Panic occurred: %v\n%s", r, debug.Stack())
					}
				}()
				err = task(ctx)
				return err
			}()
			if panicErr != nil {
				log.Println(panicErr.Error())
				errCh <- panicErr
			} else if err != nil {
				log.Printf("Error occurred: %v\n", err)
				errCh <- err
			}
		}(task)
	}

	// Wait for all tasks to complete or an error occurs
	wg.Wait()

	// Check if an error occurred during any of the tasks
	select {
	case <-ctx.Done():
		// If the context was canceled, return the corresponding error
		return ctx.Err()
	case err := <-errCh:
		// If an error occurred, return the error
		return err
	default:
		// All tasks completed successfully
		return nil
	}
}
