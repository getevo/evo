# stract Library

The stract library provides a parser for structured text files with variable substitution and hierarchical contexts. It's designed for parsing configuration files, templates, or any text with a specific structure and variable references.

## Installation

```go
import "github.com/getevo/evo/v2/lib/stract"
```

## Features

- **Structured Text Parsing**: Parse text files with a specific syntax
- **Variable Substitution**: Replace variables using `${variable}` syntax
- **Hierarchical Contexts**: Support for nested contexts with scoped variables
- **File Imports**: Import other files with `@import` directive
- **Rich Query API**: Methods for querying and manipulating the parsed context
- **Pattern Matching**: Check variables against patterns, prefixes, suffixes, etc.

## Usage Examples

### Basic Parsing

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/stract"
)

func main() {
    // Parse a file
    context, err := stract.OpenAndParse("config.txt")
    if err != nil {
        fmt.Printf("Error parsing file: %v\n", err)
        return
    }
    
    // Access variables
    exists, values := context.Get("database_host")
    if exists {
        fmt.Printf("Database host: %s\n", values[0])
    }
    
    // Get a single value
    port := context.GetSingleValue("database_port")
    fmt.Printf("Database port: %s\n", port)
    
    // Check if a variable has a specific value
    if context.VaryDictHas("environment", "production") {
        fmt.Println("Running in production mode")
    }
    
    // Print the entire context structure
    fmt.Println(stract.PrettyStruct(context))
}
```

### Working with Hierarchical Contexts

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/stract"
)

func main() {
    // Parse a file with nested contexts
    context, err := stract.OpenAndParse("config.txt")
    if err != nil {
        fmt.Printf("Error parsing file: %v\n", err)
        return
    }
    
    // Access a child context
    exists, dbContext := context.GetChild("database")
    if exists {
        // Access variables in the child context
        host := dbContext.GetSingleValue("host")
        port := dbContext.GetSingleValue("port")
        user := dbContext.GetSingleValue("username")
        pass := dbContext.GetSingleValue("password")
        
        fmt.Printf("Database connection: %s:%s@%s:%s\n", user, pass, host, port)
    }
    
    // Iterate through all children
    for _, child := range context.GetChildren() {
        fmt.Printf("Child context: %s\n", child.Name)
        
        // Print all variables in this context
        for _, vary := range child.GetVaryDicts() {
            fmt.Printf("  %s = %v\n", vary.Key, vary.Values)
        }
    }
}
```

### Pattern Matching with Variables

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/stract"
    "regexp"
)

func main() {
    // Parse a file
    context, err := stract.OpenAndParse("config.txt")
    if err != nil {
        fmt.Printf("Error parsing file: %v\n", err)
        return
    }
    
    // Check if a variable contains a substring
    hasLog, logPath := context.VaryDictContains("paths", "logs")
    if hasLog {
        fmt.Printf("Log path found: %s\n", logPath)
    }
    
    // Check if a variable starts with a prefix
    hasTemp, tempPath := context.VaryDictStartsWith("paths", "/tmp")
    if hasTemp {
        fmt.Printf("Temporary path found: %s\n", tempPath)
    }
    
    // Check if a variable ends with a suffix
    hasConfig, configPath := context.VaryDictEndsWith("paths", ".conf")
    if hasConfig {
        fmt.Printf("Config path found: %s\n", configPath)
    }
    
    // Match a variable against a regular expression
    ipRegex := regexp.MustCompile(`^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`)
    hasIP, matches := context.VaryDictMatch("allowed_ips", ipRegex)
    if hasIP {
        for _, match := range matches {
            fmt.Printf("IP address found: %s\n", match[0])
        }
    }
}
```

### Parsing Variables from Strings

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/stract"
)

func main() {
    // Parse a variable from a string
    value := stract.ParseVar("timeout:30", "timeout")
    fmt.Printf("Timeout value: %s\n", value) // Output: 30
    
    value = stract.ParseVar("retry(3)", "retry")
    fmt.Printf("Retry value: %s\n", value) // Output: 3
    
    // No match
    value = stract.ParseVar("unknown:value", "timeout")
    fmt.Printf("Value: %s\n", value) // Output: empty string
}
```

## Example Input File Format

The stract library can parse files with the following format:

```
# This is a comment

# Simple key-value pairs
database_host localhost
database_port 5432
database_user admin
database_pass secret

# Multiple values for a key
allowed_ips 192.168.1.1 192.168.1.2 10.0.0.1

# Variable substitution
server_url http://${database_host}:8080

# Nested context
database {
    host localhost
    port 5432
    username admin
    password secret
}

# Importing another file
@import common.txt
```

## How It Works

The stract library parses text files into a hierarchical context structure. The parsing process involves:

1. Tokenizing the input text, handling special characters and quoted strings
2. Building a context tree with variables and nested contexts
3. Resolving variable references using the `${variable}` syntax
4. Processing import directives to include other files

The resulting context structure can be queried using various methods:

- `Get()`: Get all values for a variable
- `GetSingleValue()`: Get the first value of a variable
- `GetChild()`: Get a child context by name
- `VaryDictHas()`: Check if a variable has a specific value
- `VaryDictContains()`: Check if a variable contains a substring
- `VaryDictStartsWith()`: Check if a variable starts with a prefix
- `VaryDictEndsWith()`: Check if a variable ends with a suffix
- `VaryDictMatch()`: Match a variable against a regular expression

The library is particularly useful for parsing configuration files, templates, or any text with a specific structure and variable references.

For more detailed information, please refer to the source code and comments within the library.