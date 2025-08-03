# try Library

The try library provides a structured way to handle panics in Go applications using a syntax similar to try-catch-finally blocks found in other programming languages. It builds on top of the `panics` library to offer a more intuitive error handling approach.

## Installation

```go
import "github.com/getevo/evo/v2/lib/try"
```

## Features

- **Structured Error Handling**: Handle panics with a familiar try-catch-finally pattern
- **Panic Recovery**: Safely catch and recover from panics without crashing your application
- **Error Conversion**: Access recovered panics as errors
- **Cleanup Operations**: Define cleanup code that runs regardless of whether a panic occurred
- **Panic Propagation**: Optionally rethrow panics to be caught by outer handlers

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/try"
    "github.com/getevo/evo/v2/lib/panics"
)

func main() {
    // Execute code that might panic
    try.This(func() {
        // This will panic
        panic("something went wrong")
    }).Catch(func(recovered *panics.Recovered) {
        fmt.Printf("Caught panic: %v\n", recovered.Value)
    })
    
    // Program execution continues here
    fmt.Println("Program continues after panic")
}
```

### Handling Nil Pointer Dereference

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/try"
    "github.com/getevo/evo/v2/lib/panics"
    "net"
)

func main() {
    var x *net.IPNet = nil
    
    try.This(func() {
        // This will panic with nil pointer dereference
        x.Contains(net.IP("192.168.1.0"))
    }).Catch(func(recovered *panics.Recovered) {
        fmt.Println("Caught nil pointer dereference")
        fmt.Printf("Error: %v\n", recovered.Value)
    })
    
    // Program execution continues here
    fmt.Println("Program continues after panic")
}
```

### Using Finally for Cleanup

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/try"
    "github.com/getevo/evo/v2/lib/panics"
    "os"
)

func main() {
    var file *os.File
    
    try.This(func() {
        var err error
        file, err = os.Open("file.txt")
        if err != nil {
            panic(err)
        }
        
        // Process file...
        
    }).Finally(func() {
        // This will run regardless of whether a panic occurred
        if file != nil {
            file.Close()
            fmt.Println("File closed in finally block")
        }
    }).Catch(func(recovered *panics.Recovered) {
        fmt.Printf("Error processing file: %v\n", recovered.Value)
    })
}
```

### Rethrowing Panics

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/try"
    "github.com/getevo/evo/v2/lib/panics"
)

func innerFunction() {
    try.This(func() {
        panic("inner panic")
    }).Catch(func(recovered *panics.Recovered) {
        fmt.Println("Inner catch: handling the panic")
        
        // Rethrow the panic to be caught by outer handlers
        try.Throw()
    })
}

func main() {
    try.This(func() {
        innerFunction()
    }).Catch(func(recovered *panics.Recovered) {
        fmt.Println("Outer catch: caught rethrown panic")
    })
    
    fmt.Println("Program continues")
}
```

## How It Works

The try library is built around three main components:

1. **This()**: Executes a function and catches any panics that occur
2. **Catch()**: Handles any caught panics
3. **Finally()**: Defines cleanup code that runs regardless of whether a panic occurred

The library uses Go's built-in panic recovery mechanism with `defer` and `recover()`, but wraps it in a more convenient API that makes it easier to handle panics in a structured way. It leverages the `panics` library to provide detailed information about the panic, including stack traces and caller information.

When you call `This()`, it executes your function inside a protected context that catches any panics. If a panic occurs, it's stored in an `exception` struct. You can then use `Catch()` to handle the panic or `Finally()` to define cleanup code. If you want to rethrow a panic to be caught by an outer handler, you can use `Throw()`.

For more detailed information, please refer to the source code and comments within the library.