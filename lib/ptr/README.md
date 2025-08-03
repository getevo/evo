# ptr Library

The ptr library provides utility functions for creating pointers to primitive types in Go. This is particularly useful when you need to pass a pointer to a literal value or when working with APIs that require pointers.

## Installation

```go
import "github.com/getevo/evo/v2/lib/ptr"
```

## Features

- **Primitive Type Pointers**: Create pointers to all Go primitive types
- **Numeric Types**: Support for all integer and floating-point types
- **String Pointers**: Create pointers to string values
- **Boolean Pointers**: Create pointers to boolean values
- **Time Pointers**: Create pointers to time.Time values
- **Interface Pointers**: Create pointers to interface{} values

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/ptr"
    "time"
)

func main() {
    // Create pointers to primitive types
    intPtr := ptr.Int(42)
    stringPtr := ptr.String("hello")
    boolPtr := ptr.Bool(true)
    floatPtr := ptr.Float64(3.14)
    timePtr := ptr.Time(time.Now())
    
    // Use the pointers
    fmt.Printf("Int value: %d\n", *intPtr)
    fmt.Printf("String value: %s\n", *stringPtr)
    fmt.Printf("Bool value: %t\n", *boolPtr)
    fmt.Printf("Float value: %f\n", *floatPtr)
    fmt.Printf("Time value: %v\n", *timePtr)
}
```

### Using with Struct Fields

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/ptr"
)

type User struct {
    ID        int
    Name      string
    Age       *int       // Optional field
    Email     *string    // Optional field
    IsActive  *bool      // Optional field
}

func main() {
    // Create a user with all fields
    user1 := User{
        ID:       1,
        Name:     "John",
        Age:      ptr.Int(30),
        Email:    ptr.String("john@example.com"),
        IsActive: ptr.Bool(true),
    }
    
    // Create a user with only required fields
    user2 := User{
        ID:   2,
        Name: "Jane",
        // Age, Email, and IsActive are nil
    }
    
    // Check if optional fields are set
    if user1.Age != nil {
        fmt.Printf("User1 age: %d\n", *user1.Age)
    }
    
    if user2.Age != nil {
        fmt.Printf("User2 age: %d\n", *user2.Age)
    } else {
        fmt.Println("User2 age not set")
    }
}
```

### Using with Function Parameters

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/ptr"
)

// Function that takes optional parameters as pointers
func updateUser(id int, name *string, age *int, isActive *bool) {
    fmt.Printf("Updating user %d\n", id)
    
    if name != nil {
        fmt.Printf("New name: %s\n", *name)
    }
    
    if age != nil {
        fmt.Printf("New age: %d\n", *age)
    }
    
    if isActive != nil {
        fmt.Printf("New active status: %t\n", *isActive)
    }
}

func main() {
    // Update only the name
    updateUser(1, ptr.String("John Doe"), nil, nil)
    
    // Update name and age
    updateUser(2, ptr.String("Jane Doe"), ptr.Int(25), nil)
    
    // Update all fields
    updateUser(3, ptr.String("Bob Smith"), ptr.Int(40), ptr.Bool(false))
}
```

## How It Works

The ptr library provides a set of simple functions, each taking a value of a specific type and returning a pointer to that value. For example, `ptr.Int(42)` creates a new int with the value 42 and returns a pointer to it.

This is equivalent to the following Go code:
```go
func Int(v int) *int {
    return &v
}
```

The library includes similar functions for all primitive types in Go, making it easy to create pointers to literal values without having to declare variables first.

For more detailed information, please refer to the source code and comments within the library.