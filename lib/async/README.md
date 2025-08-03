# Async Library

The Async library provides functionality for concurrent programming in Go with enhanced safety and control. It extends Go's standard concurrency primitives with features like panic handling, worker pools, and structured concurrency patterns.

## Installation

```go
import "github.com/getevo/evo/v2/lib/async"
```

## Features

- **Enhanced WaitGroup**: Extends Go's sync.WaitGroup with panic handling and propagation
- **Goroutine Pools**: Efficient management of goroutines with configurable limits
- **Error Handling**: Specialized pools for tasks that can return errors
- **Context Support**: Context-aware pools for shared cancellation
- **Structured Concurrency**: Tools for safer and more maintainable concurrent code
- **Stream Processing**: Asynchronous stream operations for data processing

## Usage Examples

### Basic Usage with WaitGroup

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/async"
)

func main() {
    // Create a new WaitGroup
    wg := async.NewWaitGroup()
    
    // Execute tasks concurrently
    for i := 0; i < 5; i++ {
        i := i // Capture the variable
        wg.Exec(func() {
            fmt.Println("Task", i)
        })
    }
    
    // Wait for all tasks to complete
    wg.Wait()
}
```

### Using a Pool with Maximum Goroutines

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/async"
)

func main() {
    // Create a new Pool with maximum 3 goroutines
    pool := async.NewPool().WithMaxGoroutines(3)
    
    // Submit 10 tasks to the pool
    for i := 0; i < 10; i++ {
        i := i // Capture the variable
        pool.Exec(func() {
            fmt.Println("Processing task", i)
        })
    }
    
    // Wait for all tasks to complete
    pool.Wait()
}
```

### Error Handling with ErrorPool

```go
package main

import (
    "errors"
    "fmt"
    "github.com/getevo/evo/v2/lib/async"
)

func main() {
    // Create a new ErrorPool
    pool := async.NewPool().WithErrors()
    
    // Submit tasks that can return errors
    pool.Exec(func() error {
        return nil // Success
    })
    
    pool.Exec(func() error {
        return errors.New("something went wrong")
    })
    
    // Wait and collect errors
    err := pool.Wait()
    
    // Handle error
    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

## How It Works

The async library builds on Go's concurrency primitives to provide higher-level abstractions:

1. **WaitGroup** wraps sync.WaitGroup and adds panic catching, allowing panics in goroutines to be propagated to the caller.

2. **Pool** manages a collection of goroutines for executing tasks concurrently. Goroutines are created lazily and reused for multiple tasks, improving efficiency. The pool can be configured with a maximum number of goroutines.

3. **ErrorPool** extends Pool to handle tasks that can return errors, collecting them for processing after all tasks complete.

4. **ContextPool** adds context support to ErrorPool, allowing tasks to respect shared cancellation signals.

These components work together to provide a comprehensive toolkit for concurrent programming that is both powerful and safe to use.

