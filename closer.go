package closer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Closer manages a list of functions
// to be closed in a controlled manner with concurrency support.
type Closer struct {
	mu    sync.Mutex // Mutex for synchronizing access to the function
	funcs []Func     // List of functions to close
	size  int        // Total number of added functions
	i     int        // Index of the current function to close
}

const (
	ErrAllServicesClosed = "all services closed"
)

// Add adds a function to the list for closing.
func (c *Closer) Add(f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f)
	c.size++
}

// Close closes all the functions in the list, starting from the current function.
func (c *Closer) Close(ctx context.Context) error {
	op := "closer.Close"

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if all functions have already been closed
	if c.i >= c.size {
		return fmt.Errorf("%s: %v", op, ErrAllServicesClosed)
	}

	var (
		fErrorChanels = make([]<-chan error, 0, c.size) // Error channels for each function
		fErrors       = make([]string, 0, c.size)       // List of errors
		wg            sync.WaitGroup                    // Wait group for concurrent operations
	)

	// Run each function to close it in a separate goroutine
	for _, f := range c.funcs[c.i:] {
		wg.Add(1)

		fErrChan := execF(ctx, f, &wg)

		// Append the error channel to the list
		fErrorChanels = append(fErrorChanels, fErrChan)
	}

	wg.Wait()

	// Collect all errors from the channels
	for _, el := range fErrorChanels {
		for err := range el {
			if err != nil {
				fErrors = append(fErrors, err.Error())
			}
		}
	}

	// Disable further calls to CloseOne by setting the index to the size
	c.i = c.size

	if len(fErrors) > 0 {
		return fmt.Errorf("%s: %v", op, strings.Join(fErrors, ";\x20"))
	}

	return nil
}

// CloseOne closes one function and updates the index for the next operation.
func (c *Closer) CloseOne(ctx context.Context) error {
	op := "closer.CloseOne"

	c.mu.Lock()

	// Save the current index for calling the function
	prev := c.i

	err := func() error {
		defer c.mu.Unlock()

		// Check if all functions have already been closed
		if c.i >= c.size {
			return fmt.Errorf("%s: %v", op, ErrAllServicesClosed)
		}

		// Increment the index for the next function
		c.i++

		return nil
	}()

	if err != nil {
		return err
	}

	return c.funcs[prev](ctx)
}

// Size returns the number of added functions to close.
func (c *Closer) Size() int {
	return c.size
}

// execF runs a function in a goroutine and returns a channel to receive any error.
func execF(ctx context.Context, f Func, wg *sync.WaitGroup) <-chan error {
	errCh := make(chan error, 1)

	go func(ctx context.Context, f Func, wg *sync.WaitGroup, errCh chan<- error) {
		defer wg.Done()
		defer close(errCh)

		// Execute the function and send any error to the channel
		err := f(ctx)

		if err != nil {
			errCh <- err
		}

	}(ctx, f, wg, errCh)

	return errCh
}

type Func func(ctx context.Context) error
