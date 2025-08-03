# Application Library

The Application library provides a framework for managing and running multiple applications with priority-based execution in Go.

## Installation

```go
import "github.com/getevo/evo/v2/lib/application"
```

## Features

The Application library offers the following key features:

- **Application Management**: Register and manage multiple applications
- **Priority-Based Execution**: Run applications in order of priority
- **Lifecycle Management**: Standardized application lifecycle with Register, Router, and WhenReady phases
- **Debug Mode**: Optional debug logging for application execution
- **Application Reloading**: Support for reloading applications that implement the ReloadInterface

## Core Components

### Interfaces

- **Application**: Core interface that applications must implement
  ```go
  type Application interface {
      Register() error    // For application registration
      Router() error      // For setting up routes
      WhenReady() error   // Called when the application is ready
      Name() string       // Returns the application name
  }
  ```

- **PriorityInterface**: Optional interface for applications to define their priority
  ```go
  type PriorityInterface interface {
      Priority() Priority
  }
  ```

- **ReloadInterface**: Optional interface for applications that can be reloaded
  ```go
  type ReloadInterface interface {
      Reload() error
  }
  ```

### Priority Constants

```go
const (
    HIGHEST Priority = 0
    HIGH    Priority = 1
    DEFAULT Priority = 5
    LOW     Priority = 6
    LOWEST  Priority = 7
)
```

## Usage Examples

### Basic Application Implementation

```go
package myapp

import (
    "github.com/getevo/evo/v2/lib/application"
)

type App struct{}

func (app App) Register() error {
    // Initialize your application
    return nil
}

func (app App) Router() error {
    // Set up routes
    return nil
}

func (app App) WhenReady() error {
    // Execute when all applications are ready
    return nil
}

func (app App) Name() string {
    return "myapp"
}

// Optional: Implement Priority interface
func (app App) Priority() application.Priority {
    return application.DEFAULT
}
```

### Registering and Running Applications

```go
package main

import (
    "github.com/getevo/evo/v2/lib/application"
    "myapp"
    "anotherapp"
)

func main() {
    // Get the application instance
    var app = application.GetInstance()
    
    // Register applications
    app.Register(myapp.App{}, anotherapp.App{})
    
    // Optional: Enable debug mode
    app.Debug(true)
    
    // Run all registered applications
    app.Run()
}
```

### Implementing Reloadable Applications

```go
package myapp

import (
    "github.com/getevo/evo/v2/lib/application"
)

type App struct{
    // Application state
    config Config
}

// Implement Application interface
func (app App) Register() error {
    // Initialize
    return nil
}

func (app App) Router() error {
    return nil
}

func (app App) WhenReady() error {
    return nil
}

func (app App) Name() string {
    return "myapp"
}

// Implement ReloadInterface
func (app App) Reload() error {
    // Reload configuration or state
    app.config = loadConfig()
    return nil
}

// To reload all applications that implement ReloadInterface:
// application.ReloadAll()
```

## How It Works

The Application library works by:

1. **Registration**: Applications are registered with the App instance
2. **Sorting**: Applications are sorted by priority (if they implement PriorityInterface)
3. **Execution**: The Run method executes each application's lifecycle methods in sequence:
   - Register() for all applications
   - Router() for all applications
   - WhenReady() for all applications

Applications that don't implement PriorityInterface are assigned DEFAULT priority. The library uses reflection to detect and call any OnRegister* methods on applications during the registration phase.