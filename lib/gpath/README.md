# gpath Library

The gpath library provides comprehensive file and path handling utilities for Go applications. It offers both high-level file operations with automatic resource management and low-level path manipulation functions.

## Installation

```go
import "github.com/getevo/evo/v2/lib/gpath"
```

## Features

- **File Operations**: Read, write, append, and truncate files with automatic resource management
- **Path Manipulation**: Get parent directories, working directory, path information
- **File/Directory Checking**: Check if files/directories exist or are empty
- **Directory Operations**: Create, copy, and remove directories
- **JSON Handling**: Read and write JSON files easily
- **Resource Management**: Automatic file closing after timeout periods

## Usage Examples

### File Operations

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/gpath"
)

func main() {
    // Open a file (creates if it doesn't exist)
    file, err := gpath.Open("example.txt")
    if err != nil {
        panic(err)
    }
    
    // Write to file
    file.WriteString("Hello, World!")
    
    // Read from file
    content, err := file.ReadAllString()
    if err != nil {
        panic(err)
    }
    fmt.Println(content) // "Hello, World!"
    
    // Append to file
    file.AppendString("\nAppended text")
    
    // Read again
    content, _ = file.ReadAllString()
    fmt.Println(content) // "Hello, World!\nAppended text"
    
    // Truncate file
    file.Truncate()
    
    // Set timeout for automatic closing
    file.SetTimeout(5 * time.Second) // File will close after 5 seconds of inactivity
    
    // Manually close
    file.Close()
}
```

### JSON File Handling

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/gpath"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // Create a file
    file, _ := gpath.Open("person.json")
    
    // Write JSON to file (with pretty printing)
    person := Person{Name: "John", Age: 30}
    file.WriteJson(person, true)
    
    // Read JSON from file
    var readPerson Person
    file.UnmarshalJson(&readPerson)
    
    fmt.Printf("Name: %s, Age: %d\n", readPerson.Name, readPerson.Age)
}
```

### Path Utilities

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/gpath"
)

func main() {
    // Get current working directory
    wd := gpath.WorkingDir()
    fmt.Println("Working directory:", wd)
    
    // Get parent directory
    parent := gpath.Parent("/path/to/file.txt")
    fmt.Println("Parent directory:", parent) // "/path/to"
    
    // Check if directory exists
    if gpath.IsDirExist("/path/to/directory") {
        fmt.Println("Directory exists")
    }
    
    // Check if file exists
    if gpath.IsFileExist("/path/to/file.txt") {
        fmt.Println("File exists")
    }
    
    // Get path information
    info := gpath.PathInfo("/path/to/file.txt")
    fmt.Println("Filename:", info.FileName)   // "file.txt"
    fmt.Println("Directory:", info.Path)      // "/path/to"
    fmt.Println("Extension:", info.Extension) // ".txt"
    
    // Create directory
    gpath.MakePath("/path/to/new/directory")
    
    // Copy file
    gpath.CopyFile("source.txt", "destination.txt")
    
    // Copy directory
    gpath.CopyDir("source_dir", "destination_dir")
    
    // Remove file or directory
    gpath.Remove("file_or_directory")
}
```

### Simple File Functions

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/gpath"
)

func main() {
    // Write to file
    gpath.Write("example.txt", "Hello, World!")
    
    // Append to file
    gpath.Append("example.txt", "\nAppended text")
    
    // Read file
    content, _ := gpath.ReadFile("example.txt")
    fmt.Println(string(content))
    
    // Safe read (doesn't return error)
    safeContent := gpath.SafeFileContent("example.txt")
    fmt.Println(string(safeContent))
}
```

## How It Works

The gpath library provides two main approaches to file handling:

1. **High-level file operations** through the `file` struct, which wraps an `os.File` with additional functionality:
   - Automatic resource management (closes files after a timeout period)
   - Simplified reading and writing operations
   - JSON serialization and deserialization

2. **Utility functions** for common file and path operations:
   - Path manipulation and information
   - File and directory checking
   - File and directory operations

The library is designed to simplify common file operations while providing robust error handling and resource management.

For more detailed information, please refer to the source code and comments within the library.