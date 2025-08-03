# reflections Library

The reflections library provides a set of utilities for working with Go's reflection capabilities. It simplifies common reflection tasks such as getting and setting struct fields, inspecting field types and tags, and working with nested structs.

## Installation

```go
import "github.com/getevo/evo/v2/lib/reflections"
```

## Features

- **Field Access**: Get and set struct fields by name
- **Field Inspection**: Get field kind, type, and tag information
- **Field Discovery**: List all fields in a struct
- **Tag Handling**: Access and search for struct tags
- **Deep Reflection**: Work with nested struct fields
- **Type Safety**: Proper error handling for type mismatches

## Usage Examples

### Getting and Setting Fields

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/reflections"
)

type Person struct {
    Name    string
    Age     int
    Address string
}

func main() {
    person := Person{
        Name:    "John Doe",
        Age:     30,
        Address: "123 Main St",
    }
    
    // Get a field value
    name, err := reflections.GetField(person, "Name")
    if err == nil {
        fmt.Printf("Name: %s\n", name)
    }
    
    // Set a field value
    err = reflections.SetField(&person, "Age", 31)
    if err == nil {
        fmt.Printf("Updated age: %d\n", person.Age)
    }
    
    // Check if a field exists
    hasField, _ := reflections.HasField(person, "Email")
    fmt.Printf("Has Email field: %t\n", hasField)
}
```

### Working with Field Types and Tags

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/reflections"
)

type User struct {
    ID        int    `json:"id" db:"user_id"`
    Username  string `json:"username" db:"username"`
    Email     string `json:"email" db:"email"`
    CreatedAt string `json:"created_at" db:"created_at"`
}

func main() {
    user := User{
        ID:        1,
        Username:  "johndoe",
        Email:     "john@example.com",
        CreatedAt: "2023-01-01",
    }
    
    // Get field kind
    kind, _ := reflections.GetFieldKind(user, "Username")
    fmt.Printf("Username field kind: %s\n", kind)
    
    // Get field type
    fieldType, _ := reflections.GetFieldType(user, "ID")
    fmt.Printf("ID field type: %s\n", fieldType)
    
    // Get field tag
    jsonTag, _ := reflections.GetFieldTag(user, "Email", "json")
    fmt.Printf("Email JSON tag: %s\n", jsonTag)
    
    // Find field by tag value
    fieldName, _ := reflections.GetFieldNameByTagValue(user, "db", "user_id")
    fmt.Printf("Field with db tag 'user_id': %s\n", fieldName)
}
```

### Working with All Fields and Tags

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/reflections"
)

type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
    InStock     bool    `json:"in_stock"`
}

func main() {
    product := Product{
        ID:          101,
        Name:        "Laptop",
        Price:       999.99,
        Description: "High-performance laptop",
        InStock:     true,
    }
    
    // Get all field names
    fields, _ := reflections.Fields(product)
    fmt.Println("Fields:", fields)
    
    // Get all field values as a map
    items, _ := reflections.Items(product)
    for field, value := range items {
        fmt.Printf("%s: %v\n", field, value)
    }
    
    // Get all JSON tags
    tags, _ := reflections.Tags(product, "json")
    fmt.Println("JSON tags:", tags)
}
```

### Working with Nested Structs

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/reflections"
)

type Address struct {
    Street  string
    City    string
    Country string
}

type Customer struct {
    Name    string
    Age     int
    Address Address
}

func main() {
    customer := Customer{
        Name: "Jane Doe",
        Age:  28,
        Address: Address{
            Street:  "456 Oak Ave",
            City:    "Metropolis",
            Country: "USA",
        },
    }
    
    // Get all fields including nested ones
    fieldsDeep, _ := reflections.FieldsDeep(customer)
    fmt.Println("All fields (deep):", fieldsDeep)
    
    // Get all field values including nested ones
    itemsDeep, _ := reflections.ItemsDeep(customer)
    for field, value := range itemsDeep {
        fmt.Printf("%s: %v\n", field, value)
    }
}
```

## How It Works

The reflections library uses Go's built-in `reflect` package to inspect and manipulate struct values at runtime. It provides a higher-level API that makes common reflection tasks easier and safer.

The library handles various edge cases, such as unexported fields, pointer values, and nested structs. It also provides both shallow and deep operations for working with nested structures.

For more detailed information, please refer to the source code and comments within the library.