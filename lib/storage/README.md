# storage Library

The storage library provides a unified interface for working with different storage backends, including local filesystem, FTP, SFTP, and Amazon S3. It allows you to perform file and directory operations using a consistent API regardless of the underlying storage system.

## Installation

```go
import "github.com/getevo/evo/v2/lib/storage"
```

## Features

- **Unified Interface**: Common API for different storage backends
- **Multiple Backends**: Support for filesystem, FTP, SFTP, and Amazon S3
- **File Operations**: Read, write, append, delete files
- **Directory Operations**: Create, list, and delete directories
- **Metadata Support**: Get and set file metadata
- **Search Capability**: Find files matching patterns
- **Multiple Instances**: Manage multiple storage instances with different configurations

## Usage Examples

### Creating Storage Instances

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/storage"
)

func main() {
    // Create a filesystem storage instance
    fsStorage, err := storage.NewStorageInstance("local", "fs:///path/to/directory")
    if err != nil {
        fmt.Printf("Error creating filesystem storage: %v\n", err)
        return
    }
    
    // Create an S3 storage instance
    s3Storage, err := storage.NewStorageInstance("s3backup", "s3://access_key:secret_key@bucket_name/prefix")
    if err != nil {
        fmt.Printf("Error creating S3 storage: %v\n", err)
        return
    }
    
    // Create an FTP storage instance
    ftpStorage, err := storage.NewStorageInstance("ftpserver", "ftp://username:password@hostname:port/path")
    if err != nil {
        fmt.Printf("Error creating FTP storage: %v\n", err)
        return
    }
    
    // Create an SFTP storage instance
    sftpStorage, err := storage.NewStorageInstance("sftpserver", "sftp://username:password@hostname:port/path")
    if err != nil {
        fmt.Printf("Error creating SFTP storage: %v\n", err)
        return
    }
    
    // Get a storage instance by name
    myStorage := storage.GetStorage("local")
    if myStorage == nil {
        fmt.Println("Storage not found")
        return
    }
    
    // List all storage instances
    instances := storage.Instances()
    for name, driver := range instances {
        fmt.Printf("Storage: %s, Type: %s\n", name, driver.Type())
    }
}
```

### File Operations

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/storage"
)

func main() {
    // Create a filesystem storage instance
    fsStorage, err := storage.NewStorageInstance("local", "fs:///path/to/directory")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    driver := storage.GetStorage("local")
    
    // Write a file
    err = driver.Write("example.txt", "Hello, World!")
    if err != nil {
        fmt.Printf("Error writing file: %v\n", err)
        return
    }
    
    // Append to a file
    err = driver.Append("example.txt", "\nAppended content")
    if err != nil {
        fmt.Printf("Error appending to file: %v\n", err)
        return
    }
    
    // Read a file
    content, err := driver.ReadAllString("example.txt")
    if err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        return
    }
    fmt.Println("File content:", content)
    
    // Write JSON to a file
    data := map[string]interface{}{
        "name": "John",
        "age":  30,
        "city": "New York",
    }
    err = driver.WriteJson("data.json", data)
    if err != nil {
        fmt.Printf("Error writing JSON: %v\n", err)
        return
    }
    
    // Check if a file exists
    if driver.IsFileExists("example.txt") {
        fmt.Println("File exists")
    }
    
    // Get file information
    fileInfo, err := driver.Stat("example.txt")
    if err != nil {
        fmt.Printf("Error getting file info: %v\n", err)
        return
    }
    fmt.Printf("File: %s, Size: %d bytes, Modified: %s\n", 
        fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime())
    
    // Delete a file
    err = driver.Remove("example.txt")
    if err != nil {
        fmt.Printf("Error deleting file: %v\n", err)
        return
    }
}
```

### Directory Operations

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/storage"
)

func main() {
    // Create a filesystem storage instance
    fsStorage, err := storage.NewStorageInstance("local", "fs:///path/to/directory")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    driver := storage.GetStorage("local")
    
    // Create a directory
    err = driver.Mkdir("new_directory")
    if err != nil {
        fmt.Printf("Error creating directory: %v\n", err)
        return
    }
    
    // Create nested directories
    err = driver.MkdirAll("path/to/nested/directory")
    if err != nil {
        fmt.Printf("Error creating nested directories: %v\n", err)
        return
    }
    
    // List files in a directory
    files, err := driver.List("new_directory")
    if err != nil {
        fmt.Printf("Error listing directory: %v\n", err)
        return
    }
    
    fmt.Println("Files in directory:")
    for _, file := range files {
        fileType := "File"
        if file.IsDir() {
            fileType = "Directory"
        }
        fmt.Printf("- %s (%s, %d bytes)\n", file.Name(), fileType, file.Size())
    }
    
    // List files recursively
    allFiles, err := driver.List(".", true)
    if err != nil {
        fmt.Printf("Error listing directory recursively: %v\n", err)
        return
    }
    fmt.Printf("Found %d files and directories\n", len(allFiles))
    
    // Check if a directory exists
    if driver.IsDirExists("new_directory") {
        fmt.Println("Directory exists")
    }
    
    // Delete a directory
    err = driver.Remove("new_directory")
    if err != nil {
        fmt.Printf("Error deleting directory: %v\n", err)
        return
    }
    
    // Delete a directory and all its contents
    err = driver.RemoveAll("path/to/nested")
    if err != nil {
        fmt.Printf("Error deleting directory recursively: %v\n", err)
        return
    }
}
```

### Searching for Files

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/storage"
)

func main() {
    // Create a filesystem storage instance
    fsStorage, err := storage.NewStorageInstance("local", "fs:///path/to/directory")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    driver := storage.GetStorage("local")
    
    // Search for files matching a pattern
    txtFiles, err := driver.Search("*.txt")
    if err != nil {
        fmt.Printf("Error searching for files: %v\n", err)
        return
    }
    
    fmt.Printf("Found %d text files:\n", len(txtFiles))
    for _, file := range txtFiles {
        fmt.Printf("- %s (%d bytes)\n", file.Path(), file.Size())
    }
    
    // Search for files in a specific directory
    configFiles, err := driver.Search("config/*.json")
    if err != nil {
        fmt.Printf("Error searching for config files: %v\n", err)
        return
    }
    
    fmt.Printf("Found %d config files:\n", len(configFiles))
    for _, file := range configFiles {
        fmt.Printf("- %s (Last modified: %s)\n", file.Name(), file.ModTime())
    }
}
```

### Working with FileInfo

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/storage"
)

func main() {
    // Create a filesystem storage instance
    fsStorage, err := storage.NewStorageInstance("local", "fs:///path/to/directory")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    driver := storage.GetStorage("local")
    
    // Write a test file
    driver.Write("test.txt", "Test content")
    
    // Get file information
    fileInfo, err := driver.Stat("test.txt")
    if err != nil {
        fmt.Printf("Error getting file info: %v\n", err)
        return
    }
    
    // Use FileInfo methods
    fmt.Printf("File: %s\n", fileInfo.Name())
    fmt.Printf("Path: %s\n", fileInfo.Path())
    fmt.Printf("Directory: %s\n", fileInfo.Dir())
    fmt.Printf("Extension: %s\n", fileInfo.Extension())
    fmt.Printf("Size: %d bytes\n", fileInfo.Size())
    fmt.Printf("Mode: %s\n", fileInfo.Mode())
    fmt.Printf("Modified: %s\n", fileInfo.ModTime())
    fmt.Printf("Is Directory: %t\n", fileInfo.IsDir())
    
    // Perform operations directly on FileInfo
    err = fileInfo.Append("\nAppended from FileInfo")
    if err != nil {
        fmt.Printf("Error appending to file: %v\n", err)
        return
    }
    
    // Read the updated content
    content, err := driver.ReadAllString("test.txt")
    if err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        return
    }
    fmt.Println("Updated content:", content)
    
    // Delete the file using FileInfo
    err = fileInfo.Remove()
    if err != nil {
        fmt.Printf("Error removing file: %v\n", err)
        return
    }
}
```

## How It Works

The storage library is built around the `Driver` interface, which defines a common set of methods for interacting with different storage backends. Each storage backend (filesystem, FTP, SFTP, S3) implements this interface, providing a consistent API regardless of the underlying storage system.

The library uses a URL-like format for configuration strings, where the protocol determines which driver to use:

- Filesystem: `fs:///path/to/directory`
- Amazon S3: `s3://access_key:secret_key@bucket_name/prefix`
- FTP: `ftp://username:password@hostname:port/path`
- SFTP: `sftp://username:password@hostname:port/path`

Storage instances are created using the `NewStorageInstance` function, which takes a tag (name) and a configuration string. The tag is used to identify the storage instance later using the `GetStorage` function.

The `FileInfo` struct provides information about files and directories, including name, size, modification time, and permissions. It also includes methods for performing operations directly on the file, such as appending, writing, and removing.

For more detailed information, please refer to the source code and comments within the library.