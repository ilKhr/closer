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

	length := c.size - c.i

	var (
		fErrChan = make(chan error, length)  // Error channels for each function
		fErrors  = make([]string, 0, length) // List of errors
		wg       sync.WaitGroup              // Wait group for concurrent operations
	)

	// Run each function to close it in a separate goroutine
	for _, f := range c.funcs[c.i:] {
		wg.Add(1)

		execF(ctx, f, &wg, fErrChan)
	}

	wg.Wait()

	// Collect all errors from the channels

	for range length {
		select {
		case err := <-fErrChan:
			if err != nil {
				fErrors = append(fErrors, err.Error())
			}
		default:
			break
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
func execF(ctx context.Context, f Func, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	// Execute the function and send any error to the channel
	err := f(ctx)

	if err != nil {
		errCh <- err
	}
}

func (c *Closer) reset() {
	c.mu.Lock()
	c.i = 0
	c.mu.Unlock()
}

type Func func(ctx context.Context) error
