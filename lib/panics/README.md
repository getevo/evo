# panics Library

The panics library provides utilities for safely handling and recovering from panics in Go applications. It allows you to catch panics, inspect their details, and convert them to errors for more graceful error handling.

## Installation

```go
import "github.com/getevo/evo/v2/lib/panics"
```

## Features

- **Panic Recovery**: Safely catch and recover from panics
- **Stack Traces**: Access detailed stack traces for debugging
- **Error Conversion**: Convert panics to standard Go errors
- **Caller Information**: Get information about where the panic occurred
- **Safe Execution**: Run functions with automatic panic recovery

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/panics"
)

func main() {
    // Create a new panic catcher
    pc := &panics.Catcher{}
    
    // Execute a function that might panic
    pc.Try(func() {
        // This will panic
        panic("something went wrong")
    })
    
    // Check if a panic occurred
    if pc.Recovered() != nil {
        fmt.Printf("Caught panic: %v\n", pc.Recovered().Value)
        
        // Get the stack trace
        fmt.Printf("Stack trace: %s\n", pc.Recovered().Stack)
    }
}
```

### Converting Panics to Errors

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/panics"
)

func riskyFunction() error {
    // Create a new panic catcher
    pc := &panics.Catcher{}
    
    // Execute a function that might panic
    pc.Try(func() {
        // This will panic
        panic("something went wrong")
    })
    
    // Convert any panic to an error
    if pc.Recovered() != nil {
        return pc.Recovered().AsError()
    }
    
    return nil
}

func main() {
    if err := riskyFunction(); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Using the Try Helper Function

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/panics"
)

func main() {
    // Execute a function and catch any panics
    recovered := panics.Try(func() {
        // This will panic
        panic("something went wrong")
    })
    
    if recovered != nil {
        fmt.Printf("Caught panic: %v\n", recovered.Value)
    }
}
```

### Accessing Caller Information

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/panics"
    "runtime"
)

func main() {
    pc := &panics.Catcher{}
    
    pc.Try(func() {
        panic("something went wrong")
    })
    
    if pc.Recovered() != nil {
        // Get information about the callers
        frames := runtime.CallersFrames(pc.Recovered().Callers)
        for {
            frame, more := frames.Next()
            fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
            if !more {
                break
            }
        }
    }
}
```

## How It Works

The panics library is built around the `Catcher` struct, which provides methods to execute functions and catch any panics that occur. When a panic is caught, it's stored in a `Recovered` struct that contains the panic value, stack trace, and caller information.

The library uses Go's built-in panic recovery mechanism with `defer` and `recover()`, but wraps it in a more convenient API that makes it easier to handle panics in a structured way.

For more detailed information, please refer to the source code and comments within the library.