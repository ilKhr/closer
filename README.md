# Closer
[![Test][github-actions-ci-image]][github-actions-ci-url]
[![Tag Version][tag-version-image]][tag-version-url]

**Closer** is a Go package that provides a mechanism for managing the closing of multiple functions in a controlled and concurrency-safe manner. The package allows you to add functions that should be executed upon closing and then close them one by one or all at once.

### Key Features

- **Add Functions**: You can add functions that need to be executed upon closing using the `Add` method.
- **Close All Functions**: The `Close` method allows you to close all added functions simultaneously, executing them in separate goroutines. **Note**: The Close method does not guarantee the order of execution.
- **Step-by-Step Closing**: The `CloseOne` method allows you to close functions one by one in a `FIFO` (First-In-First-Out) order, which can be useful in scenarios where sequential resource closing is required.
- **Concurrency Safety**: All operations with functions are synchronized using a mutex, ensuring safety in a multi-threaded environment.
- **Error Handling**: If errors occur while closing functions, they are collected and returned as a single error message.

### Example Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ilKhr/closer"
)

func main() {
	ctx := context.Background()

	// Create an instance of Closer
	var cl closer.Closer

	// Add functions to be closed
	cl.Add(func(ctx context.Context) error {
		fmt.Println("Closing service 1")
		time.Sleep(5 * time.Second)
		return nil
	})

	cl.Add(func(ctx context.Context) error {
		fmt.Println("Closing service 2")
		time.Sleep(3 * time.Second)
		return fmt.Errorf("error closing service 2")
	})

	// Close all functions
	err := cl.Close(ctx)
	if err != nil {
		fmt.Printf("Error closing services: %v\n", err)
	}
}
```

### Methods

#### `Add(f Func)`
Adds the function `f` to the list of functions that should be closed.

#### `Close(ctx context.Context) error`
Closes all added functions simultaneously. If errors occur while closing, they are collected and returned as a single error message.

#### `CloseOne(ctx context.Context) error`
Closes one function and updates the index for the next operation. If all functions have already been closed, it returns the `ErrAllServicesClosed` error.

#### `Size() int`
Returns the number of added functions to be closed.

### Types

#### `Func func(ctx context.Context) error`
The type of function that takes a context and returns an error. This type is used for adding functions to the closing list.

### Errors

- **`ErrAllServicesClosed`**: Returned if all functions have already been closed, and attempting to close them again is meaningless.

### Dependencies

The package uses standard Go libraries such as `context`, `fmt`, `strings`, and `sync`.

### Installation

```bash
go get github.com/ilKhr/closer
```

### License

This project is licensed under the [MIT License](LICENSE).

### Author

Khorishko Ilya

### Contributing

Contributions are welcome! Please create an issue or pull request on GitHub.


[github-actions-ci-image]: https://badgen.net/github/checks/ilKhr/closer/main/test
[github-actions-ci-url]: https://github.com/ilKhr/closer/actions/workflows/test.yml
[tag-version-image]: https://badgen.net/github/tag/ilKhr/closer
[tag-version-url]: https://badgen.net/github/tag/ilKhr/closer
