# Dot Library

The Dot library provides a powerful utility for accessing and modifying nested properties in objects using dot notation. It simplifies working with complex nested structures in Go by allowing you to use string paths to navigate through maps, structs, and arrays.

## Installation

```go
import "github.com/getevo/evo/v2/lib/dot"
```

## Features

- **Dot Notation Access**: Access nested properties using simple string paths (e.g., "user.address.city")
- **Type Support**: Works with maps, structs, arrays, and slices
- **Get & Set Operations**: Both retrieve and modify values in complex objects
- **Array Indexing**: Support for array/slice access using index notation (e.g., "users[0].name")
- **Error Handling**: Proper error reporting for invalid paths or operations
- **Reflection-Based**: Uses Go's reflection capabilities for dynamic property access
- **Integration**: Seamless integration with the EVO Framework

## Usage Examples

### Basic Usage with Maps

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/dot"
)

func main() {
    // Create a map with nested properties
    data := map[string]any{
        "user": map[string]any{
            "name": "John Doe",
            "address": map[string]any{
                "city": "New York",
                "zip": "10001",
            },
        },
    }
    
    // Get a nested property
    city, err := dot.Get(data, "user.address.city")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("City:", city) // Output: City: New York
    
    // Set a nested property
    err = dot.Set(&data, "user.address.city", "San Francisco")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    // Verify the change
    city, _ = dot.Get(data, "user.address.city")
    fmt.Println("New City:", city) // Output: New City: San Francisco
}
```

### Working with Structs

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/dot"
)

func main() {
    // Define nested structs
    type Address struct {
        City string
        Zip  string
    }
    
    type User struct {
        Name    string
        Address Address
    }
    
    // Create a struct with nested properties
    user := User{
        Name: "Jane Smith",
        Address: Address{
            City: "Boston",
            Zip:  "02108",
        },
    }
    
    // Get a nested property
    city, err := dot.Get(user, "Address.City")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("City:", city) // Output: City: Boston
    
    // Set a nested property (note: struct must be passed as pointer for Set)
    err = dot.Set(&user, "Address.City", "Chicago")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    // Verify the change
    fmt.Println("New City:", user.Address.City) // Output: New City: Chicago
}
```

### Working with Arrays and Slices

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/dot"
)

func main() {
    // Create a map with an array
    data := map[string]any{
        "users": []map[string]any{
            {
                "name": "Alice",
                "age":  30,
            },
            {
                "name": "Bob",
                "age":  25,
            },
        },
    }
    
    // Get a value from an array using index
    name, err := dot.Get(data, "users[0].name")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("First user:", name) // Output: First user: Alice
    
    // Set a value in an array
    err = dot.Set(&data, "users[1].age", 26)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    // Verify the change
    age, _ := dot.Get(data, "users[1].age")
    fmt.Println("Bob's new age:", age) // Output: Bob's new age: 26
}
```

## API Reference

### Get

```go
func Get(obj any, prop string) (any, error)
```

Retrieves a value from an object using dot notation.

Parameters:
- `obj`: The object to retrieve the value from (can be a map, struct, array, or slice)
- `prop`: The property path using dot notation (e.g., "user.address.city" or "users[0].name")

Returns:
- The value at the specified path
- An error if the path is invalid or the property doesn't exist

### Set

```go
func Set(input any, prop string, value any) error
```

Sets a value in an object using dot notation.

Parameters:
- `input`: The object to modify (must be a pointer for structs)
- `prop`: The property path using dot notation (e.g., "user.address.city" or "users[0].name")
- `value`: The new value to set

Returns:
- An error if the path is invalid, the property doesn't exist, or the object is not modifiable

## How It Works

The Dot library uses Go's reflection capabilities to dynamically access and modify properties in objects. It works by:

1. Parsing the dot notation path into individual segments
2. Traversing the object structure one segment at a time
3. Using reflection to access properties in maps, structs, arrays, and slices
4. Handling special cases like array indexing with a regular expression

For maps, it directly accesses the map keys. For structs, it uses reflection to access the fields. For arrays and slices, it uses indexing to access elements.

When setting values, it ensures that the object is properly modifiable (e.g., structs must be passed as pointers) and creates intermediate objects as needed (e.g., creating nested maps if they don't exist).

## Related Libraries

- **reflections**: Used internally for struct field access
- **generic**: Used for type conversions
- **text**: Used for string manipulation

For more detailed information, please refer to the source code and comments within the library.