# Project Guidelines for Junie

## Project Overview

EVO Framework is a backend development solution designed to facilitate efficient development using the Go programming language. It is built with a focus on modularity and follows the MVC (Model-View-Controller) architectural pattern. The core of EVO Framework is highly extensible, allowing for seamless extension or replacement of its main modules.

### Key Features

- **Modularity**: EVO Framework promotes modularity, enabling developers to structure their codebase in a modular manner.
- **MVC Structure**: Following the widely adopted MVC pattern, EVO Framework separates concerns and improves code organization.
- **Comprehensive Toolset**: EVO Framework provides a rich set of tools, eliminating the need for developers to deal with low-level libraries and technologies.
- **Enhanced Readability**: By leveraging the EVO Framework, code becomes more readable and clear, enhancing collaboration and maintainability.

## Project Structure

The EVO Framework project is organized as follows:

```
evo/
├── assets/             # Static assets
├── dev/                # Development utilities and examples
│   ├── app/            # Example applications
│   ├── example/        # Code examples
│   └── logs/           # Development logs
├── docs/               # Documentation
├── example/            # Additional examples
├── lib/                # Core libraries
│   ├── application/    # Application management
│   ├── args/           # Command-line argument handling
│   ├── async/          # Asynchronous operations
│   ├── build/          # Build utilities
│   ├── connectors/     # External system connectors
│   ├── curl/           # HTTP client functionality
│   ├── date/           # Date and time utilities
│   ├── db/             # Database operations
│   ├── dot/            # Dot notation for accessing nested data
│   ├── errors/         # Error handling
│   ├── frm/            # Form handling
│   ├── generic/        # Generic data structures
│   ├── gpath/          # Path manipulation
│   ├── is/             # Type checking utilities
│   ├── json/           # JSON handling
│   ├── log/            # Logging system
│   ├── memo/           # Memoization and caching
│   ├── model/          # Data modeling
│   ├── outcome/        # Function result handling
│   ├── panics/         # Panic handling
│   ├── ptr/            # Pointer utilities
│   ├── pubsub/         # Publish-subscribe pattern
│   ├── reflections/    # Reflection utilities
│   ├── scheduler/      # Task scheduling
│   ├── serializer/     # Data serialization
│   ├── settings/       # Configuration management
│   ├── storage/        # Storage interfaces
│   ├── stract/         # Structured data handling
│   ├── text/           # Text manipulation
│   ├── tpl/            # Template system
│   ├── try/            # Error handling with try/catch pattern
│   ├── validation/     # Data validation
│   └── version/        # Version management
├── *.go                # Core framework files
├── go.mod              # Go module definition
└── go.sum              # Go module checksums
```

## Architecture

EVO Framework follows a modular architecture with the following key components:

1. **Core Framework**: The main package (`evo`) provides the core functionality for setting up and running the framework.
2. **Web Server**: The framework uses [Fiber](https://github.com/gofiber/fiber) as its web server, providing high-performance HTTP handling.
3. **Application Management**: The `application` library manages multiple applications with priority-based execution.
4. **Database Layer**: The framework includes a database layer with ORM capabilities, migrations, and other database utilities.
5. **Settings Management**: The `settings` library provides configuration management with support for different environments.
6. **Validation**: The `validation` library offers comprehensive data validation capabilities.
7. **Error Handling**: The `try` and `panics` libraries provide structured error handling with try-catch-finally patterns.
8. **Utilities**: The framework includes numerous utility libraries for common tasks like logging, date handling, and file operations.

## Development Guidelines

### Testing Approach

When working with the EVO Framework, follow these testing guidelines:

1. **Unit Tests**: Write unit tests for individual functions and methods using Go's standard testing package.
2. **Integration Tests**: Write integration tests for components that interact with external systems like databases or APIs.
3. **Test Coverage**: Aim for high test coverage, especially for critical components.
4. **Running Tests**: Use the standard Go test command: `go test ./...`

### Build Process

To build an EVO Framework project:

1. **Development Build**: Use `go run main.go` for development.
2. **Production Build**: Use `go build -o app main.go` for production builds.
3. **Docker Build**: Use the provided Dockerfile for containerized deployments.
4. **Configuration**: Ensure proper configuration files are in place before building.

### Code Style Guidelines

Follow these code style guidelines when working with the EVO Framework:

1. **Go Standard**: Follow the Go standard code style and conventions.
2. **Error Handling**: Use the `try` library for structured error handling.
3. **Validation**: Use the `validation` library for input validation.
4. **Logging**: Use the `log` library for consistent logging.
5. **Configuration**: Use the `settings` library for configuration management.
6. **Database Operations**: Use the provided database utilities for database operations.
7. **HTTP Requests**: Use the `curl` library for making HTTP requests.

## Working with the Framework

### Application Setup

To set up an EVO Framework application:

```go
package main

import (
    "github.com/getevo/evo/v2"
)

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Your code goes here...
    
    // Run the server
    evo.Run()
}
```

### Routing

To define routes in an EVO Framework application:

```go
package main

import (
    "github.com/getevo/evo/v2"
)

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Define routes
    evo.Get("/api/users", func(request *evo.Request) interface{} {
        // Handle GET request
        return "List of users"
    })
    
    evo.Post("/api/users", func(request *evo.Request) interface{} {
        // Handle POST request
        return "User created"
    })
    
    // Run the server
    evo.Run()
}
```

### Database Operations

To perform database operations:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
)

type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name" validation:"required"`
    Email string `json:"email" validation:"required,email"`
}

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Create a user
    user := User{Name: "John Doe", Email: "john@example.com"}
    db.Create(&user)
    
    // Find a user
    var foundUser User
    db.First(&foundUser, "email = ?", "john@example.com")
    
    // Run the server
    evo.Run()
}
```

### Validation

To validate data:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/validation"
)

type User struct {
    Name     string `json:"name" validation:"required,name"`
    Email    string `json:"email" validation:"required,email"`
    Password string `json:"password" validation:"required,password(medium)"`
    Age      int    `json:"age" validation:">=18"`
}

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Define routes
    evo.Post("/api/users", func(request *evo.Request) interface{} {
        var user User
        request.BodyParser(&user)
        
        errors := validation.Struct(user)
        if len(errors) > 0 {
            return errors
        }
        
        // Create user...
        return "User created"
    })
    
    // Run the server
    evo.Run()
}
```

### Error Handling

To handle errors:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/try"
    "github.com/getevo/evo/v2/lib/panics"
)

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Define routes
    evo.Get("/api/risky", func(request *evo.Request) interface{} {
        var result string
        
        try.This(func() {
            // Risky operation
            result = riskyOperation()
        }).Catch(func(recovered *panics.Recovered) {
            // Handle error
            request.Error("Operation failed: " + recovered.Value.(string))
        })
        
        return result
    })
    
    // Run the server
    evo.Run()
}
```

## Additional Resources

For more information about the EVO Framework, refer to the following resources:

- [EVO Framework Documentation](https://github.com/getevo/evo/tree/master/docs)
- [AI Guidelines for EVO Framework](https://github.com/getevo/evo/blob/master/docs/ai_guideline.md)
- [Library Documentation](https://github.com/getevo/evo/tree/master/lib)

## Junie-Specific Guidelines

When using Junie to work with the EVO Framework, follow these guidelines:

1. **Use the Correct Import Paths**: Always use `github.com/getevo/evo/v2` for imports.
2. **Follow the MVC Pattern**: Organize code according to the MVC pattern.
3. **Use the Provided Libraries**: Leverage the extensive library ecosystem provided by the framework.
4. **Handle Errors Properly**: Use the structured error handling provided by the framework.
5. **Validate Input**: Always validate user input using the validation library.
6. **Document Code**: Provide clear documentation for code changes.
7. **Test Changes**: Ensure changes are tested before submission.
8. **Follow Go Best Practices**: Adhere to Go best practices and conventions.