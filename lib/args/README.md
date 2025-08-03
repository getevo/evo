# Args Library

The Args library provides simple utilities for handling command-line arguments in Go applications.

## Installation

```go
import "github.com/getevo/evo/v2/lib/args"
```

## Features

The Args library offers two main functions:

- `Get(sw string)`: Retrieves the value of a command-line argument
- `Exists(sw string)`: Checks if a command-line argument has been passed to the application

## Usage Examples

### Checking if an argument exists

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/args"
)

func main() {
    // Check if "--debug" flag is provided
    if args.Exists("--debug") {
        fmt.Println("Debug mode is enabled")
    }
}
```

### Getting the value of an argument

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/args"
)

func main() {
    // Get the value of "--config" argument
    configPath := args.Get("--config")
    if configPath != "" {
        fmt.Printf("Using config file: %s\n", configPath)
    } else {
        fmt.Println("Using default config file")
    }
}
```

## How It Works

The Args library works by parsing the `os.Args` slice, which contains the command-line arguments passed to the program. It provides a simple interface to check for the existence of arguments and retrieve their values without having to manually parse the arguments slice.