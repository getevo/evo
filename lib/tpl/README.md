# tpl Library

The tpl library provides a simple yet powerful template rendering system for string interpolation. It allows you to replace variables in a template string with values from various data structures, supporting dot notation for accessing nested fields and array indexing.

## Installation

```go
import "github.com/getevo/evo/v2/lib/tpl"
```

## Features

- **Simple Variable Substitution**: Replace `$variable` placeholders with actual values
- **Dot Notation**: Access nested fields with `$object.field` syntax
- **Array Indexing**: Access array elements with `$array[index]` syntax
- **Nested Structures**: Support for complex nested objects and arrays
- **Type Conversion**: Automatic conversion of values to strings
- **Multiple Data Sources**: Pass multiple objects as data sources

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/tpl"
)

func main() {
    // Simple template with variables
    template := "Hello, $name! Welcome to $app."
    
    // Render the template with a map of values
    result := tpl.Render(template, map[string]interface{}{
        "name": "John",
        "app":  "EVO Framework",
    })
    
    fmt.Println(result) // Output: Hello, John! Welcome to EVO Framework.
}
```

### Using Structs and Nested Fields

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/tpl"
)

type User struct {
    FirstName string
    LastName  string
    Email     string
}

type App struct {
    Name    string
    Version string
}

func main() {
    // Template with dot notation for accessing struct fields
    template := "User: $user.FirstName $user.LastName ($user.Email)\nApp: $app.Name v$app.Version"
    
    // Create data structures
    user := User{
        FirstName: "Jane",
        LastName:  "Doe",
        Email:     "jane@example.com",
    }
    
    app := App{
        Name:    "EVO Framework",
        Version: "2.0.0",
    }
    
    // Render the template with multiple data sources
    result := tpl.Render(template, user, app)
    
    fmt.Println(result)
    // Output:
    // User: Jane Doe (jane@example.com)
    // App: EVO Framework v2.0.0
}
```

### Working with Arrays and Complex Structures

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/tpl"
)

type Person struct {
    Name  string
    Email string
}

func main() {
    // Template with array indexing and nested structures
    template := "Message from: $sender[0].Name <$sender[0].Email>\n" +
                "To: $recipients[0], $recipients[1]\n" +
                "Subject: $metadata.subject\n" +
                "Sent at: $metadata.time.hour:$metadata.time.minute"
    
    // Create complex data structure
    data := map[string]interface{}{
        "sender": []Person{
            {Name: "John Smith", Email: "john@example.com"},
        },
        "recipients": []string{"alice@example.com", "bob@example.com"},
        "metadata": map[string]interface{}{
            "subject": "Meeting Reminder",
            "time": map[string]int{
                "hour":   14,
                "minute": 30,
            },
        },
    }
    
    // Render the template
    result := tpl.Render(template, data)
    
    fmt.Println(result)
    // Output:
    // Message from: John Smith <john@example.com>
    // To: alice@example.com, bob@example.com
    // Subject: Meeting Reminder
    // Sent at: 14:30
}
```

### Multiple Data Sources

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/tpl"
)

func main() {
    // Template with variables from different sources
    template := "Hello, $user! Your account ($account) has $balance credits."
    
    // Define multiple data sources
    userData := map[string]string{
        "user": "Alice",
    }
    
    accountData := map[string]interface{}{
        "account": "A12345",
        "balance": 150,
    }
    
    // Render the template with multiple data sources
    // The function will check each source in order until it finds a match
    result := tpl.Render(template, userData, accountData)
    
    fmt.Println(result)
    // Output: Hello, Alice! Your account (A12345) has 150 credits.
}
```

## How It Works

The tpl library uses a regular expression to find variables in the template string and replaces them with values from the provided data sources. The variable syntax is:

- Simple variable: `$variable`
- Nested field: `$object.field`
- Array element: `$array[index]`
- Combination: `$array[index].field` or `$object.array[index]`

When rendering a template, the library:

1. Searches for variables in the template using a regular expression
2. For each variable, it looks for a matching value in the provided data sources
3. If a match is found, it converts the value to a string and replaces the variable
4. If no match is found, it leaves the variable unchanged

The library supports multiple data sources, checking each one in order until it finds a match for a variable. This allows you to combine data from different sources in a single template.

For more detailed information, please refer to the source code and comments within the library.