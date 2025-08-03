# Log Library

The Log library provides a flexible and extensible logging system for Go applications in the EVO Framework. It supports multiple logging levels, customizable output formats, and various output destinations including console and files.

## Installation

```go
import "github.com/getevo/evo/v2/lib/log"
```

For file-based logging:

```go
import "github.com/getevo/evo/v2/lib/log/file"
```

## Features

### Main Logging Features

- **Multiple Logging Levels**: Critical, Error, Warning, Notice, Info, Debug
- **Customizable Writers**: Add your own output destinations
- **Structured Log Entries**: Each entry includes timestamp, file, line number, and severity level
- **Formatted Logging**: Support for printf-style formatting
- **Fatal and Panic Handling**: Special functions for critical errors

### File Logging Features

- **Log Rotation**: Automatic daily log rotation
- **Configurable File Naming**: Use date/time placeholders in file names
- **Log Expiration**: Automatic cleanup of old log files
- **Custom Formatting**: Define your own log entry format
- **Thread-Safe**: Mutex protection for concurrent logging

## Usage Examples

### Basic Logging

```go
package main

import (
    "github.com/getevo/evo/v2/lib/log"
)

func main() {
    // Set the global log level
    log.SetLevel(log.DebugLevel)
    
    // Log messages at different levels
    log.Debug("This is a debug message")
    log.Info("This is an info message")
    log.Notice("This is a notice message")
    log.Warning("This is a warning message")
    log.Error("This is an error message")
    log.Critical("This is a critical message")
    
    // Formatted logging
    log.Infof("User %s logged in from %s", "john", "192.168.1.1")
    
    // Alternative format (uppercase 'F')
    log.InfoF("Processing item %d of %d", 1, 10)
}
```

### Custom Log Writers

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/log"
)

func main() {
    // Create a custom writer
    customWriter := func(entry *log.Entry) {
        fmt.Printf("[%s] %s: %s\n", 
            entry.Level, 
            entry.Date.Format("2006-01-02 15:04:05"),
            entry.Message)
    }
    
    // Add the custom writer
    log.AddWriter(customWriter)
    
    // Or replace all writers
    log.SetWriters(customWriter)
    
    log.Info("This message will be formatted by the custom writer")
}
```

### File Logging

```go
package main

import (
    "github.com/getevo/evo/v2/lib/log"
    "github.com/getevo/evo/v2/lib/log/file"
    "time"
)

func main() {
    // Configure file logger
    fileLogger := file.NewFileLogger(file.Config{
        Path:       "/var/log/myapp",
        FileName:   "app_%y-%m-%d.log",  // Will create files like app_2025-08-03.log
        Expiration: 7 * 24 * time.Hour,  // Keep logs for 7 days
    })
    
    // Add file logger to writers
    log.AddWriter(fileLogger)
    
    log.Info("This message will be logged to both console and file")
}
```

### Custom File Log Format

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/log"
    "github.com/getevo/evo/v2/lib/log/file"
    "time"
)

func main() {
    // Custom log format
    customFormat := func(entry *log.Entry) string {
        return fmt.Sprintf("[%s] %s - %s",
            entry.Level,
            entry.Date.Format("2006-01-02 15:04:05"),
            entry.Message)
    }
    
    // Configure file logger with custom format
    fileLogger := file.NewFileLogger(file.Config{
        Path:       "logs",
        FileName:   "app.log",
        LogFormat:  customFormat,
    })
    
    // Add file logger to writers
    log.AddWriter(fileLogger)
    
    log.Info("This message will use the custom format in the log file")
}
```

## Log Levels

The library supports the following log levels, in order of increasing verbosity:

1. **CriticalLevel**: Severe errors that cause the application to abort
2. **ErrorLevel**: Error conditions that should be addressed
3. **WarningLevel**: Warning conditions that indicate potential issues
4. **NoticeLevel**: Normal but significant conditions
5. **InfoLevel**: Informational messages about normal operation
6. **DebugLevel**: Detailed information for debugging purposes

You can set the global log level using `log.SetLevel()`. Only messages at or above the set level will be logged.

## Subdirectories

- **file**: Provides file-based logging capabilities with rotation and cleanup