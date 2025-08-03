# Build Library

The Build library provides functionality for embedding build information into Go applications, such as version numbers, build dates, user information, and commit hashes.

## Installation

```go
import "github.com/getevo/evo/v2/lib/build"
```

## Features

- **Version Embedding**: Include version information in your application
- **Build Metadata**: Capture build-time information like date, user, and commit hash
- **Simple API**: Easy-to-use functions for accessing build information
- **JSON Support**: Build information is available in a structured format with JSON tags

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/build"
)

func main() {
    // Register build information (typically called at application startup)
    build.Register()
    
    // Get build information
    info := build.GetInfo()
    
    // Access individual fields
    fmt.Println("Version:", info.Version)
    fmt.Println("Built by:", info.User)
    fmt.Println("Build date:", info.Date)
    fmt.Println("Commit:", info.Commit)
}
```

### Setting Build Information at Compile Time

To set build information when compiling your application, use the `-ldflags` option with the Go compiler:

```bash
go build -ldflags="-X 'github.com/getevo/evo/v2/lib/build.Version=v1.0.0' -X 'github.com/getevo/evo/v2/lib/build.User=developer' -X 'github.com/getevo/evo/v2/lib/build.Date=2025-08-03' -X 'github.com/getevo/evo/v2/lib/build.Commit=e21cf23'" main.go
```

## How It Works

The build library uses Go's ability to set string variables at build time through the `-ldflags` compiler option. This allows you to inject dynamic information into your binary without modifying the source code.

1. The library defines global variables (`Version`, `User`, `Date`, `Commit`) that can be set at build time.

2. The `Register()` function prints the build information to the console and stores it in an internal struct.

3. The `Information` struct provides a structured way to access the build information, with JSON tags for serialization.

4. The `GetInfo()` function returns the build information as an `Information` struct.

This simple but powerful mechanism allows you to track and display important metadata about your application builds, which is useful for debugging, support, and version management.

