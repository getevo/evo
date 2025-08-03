# Logging Library

This logging library provides a structured and customizable logging system for Go applications. You can easily log messages with different severity levels, customize output methods, and define new logging behaviors.

## Table of Contents

1. [How to Call Log Functions](#how-to-call-log-functions)
2. [How to Change Log Level](#how-to-change-log-level)
3. [How to Implement New Logging Methods and Attach Them](#how-to-implement-new-logging-methods-and-attach-them)
4. [How to Define a New StdLog Function and Set It as Default Logger](#how-to-define-a-new-stdlog-function-and-set-it-as-default-logger)
5. [File Logging](#file-logging)

---

## 1. How to Call Log Functions

To log messages, use any of the provided log functions based on the severity level:

- `Fatal`, `FatalF`, `Fatalf`
- `Panic`, `PanicF`, `Panicf`
- `Critical`, `CriticalF`, `Criticalf`
- `Error`, `ErrorF`, `Errorf`
- `Warning`, `WarningF`, `Warningf`
- `Notice`, `NoticeF`, `Noticef`
- `Info`, `InfoF`, `Infof`
- `Debug`, `DebugF`, `Debugf`

Each function accepts a message and optional parameters for formatting.

### Example:
```go
package main
import "github.com/getevo/evo/v2/lib/log"

func main() {
    log.Info("Application started")
    log.WarningF("Config file %s not found, using default values", "config.yaml")
    log.Error("Failed to connect to database")
}
```

---

## 2. How to Change Log Level

You can control the minimum severity level to log using `SetLevel`.

### Available Levels:
- `log.CriticalLevel`
- `log.ErrorLevel`
- `log.WarningLevel`
- `log.NoticeLevel`
- `log.InfoLevel`
- `log.DebugLevel`

### Example:
```go
package main
import "github.com/getevo/evo/v2/lib/log"

func main() {
    log.SetLevel(log.DebugLevel) // Set to debug level
    log.Info("This is an info message")
    log.Debug("This is a debug message")
}
```

By default, the log level is set to `WarningLevel`.

---

## 3. How to Implement New Logging Methods and Attach Them

You can implement a custom log output method and attach it to the library using `AddWriter`. A writer function receives a `log.Entry` and processes it.

### Example: Log to a File (Single Instance, Concurrent Safe)
```go
package main
import (
    "github.com/getevo/evo/v2/lib/log"
    "os"
    "sync"
    "fmt"
)

var (
    file     *os.File
    fileOnce sync.Once
    fileLock sync.Mutex
)

func FileWriter(logEntry *log.Entry) {
    // Ensure file is opened only once
    fileOnce.Do(func() {
        var err error
        file, err = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            panic("Failed to open log file")
        }
    })

    fileLock.Lock() // Ensure concurrent writes are safe
    defer fileLock.Unlock()

    fmt.Fprintf(file, "%s [%s] %s:%d %s\n", logEntry.Date.Format("15:04:05"), logEntry.Level, logEntry.File, logEntry.Line, logEntry.Message)
}

func main() {
    log.AddWriter(FileWriter) // Attach custom writer
    log.Info("This will be written to the file")
    log.Warning("This is a warning message")
}
```

This version ensures:
1. The log file is opened only once per application lifecycle.
2. Writes to the file are synchronized, making it thread-safe for concurrent logging.

- Official EVO [File Logger](docs/file_logger.md) documentation
---

## 4. How to Define a New StdLog Function and Set It as Default Logger

The library allows replacing the default `StdWriter` with a custom function using `SetWriters`. This function processes log entries before outputting them.

### Example: Define a New StdLog
```go
package main
import (
    "github.com/getevo/evo/v2/lib/log"
    "fmt"
)

func MyCustomStdWriter(entry *log.Entry) {
    fmt.Printf("[%s] %s (%s:%d) -> %s\n", entry.Date.Format("2006-01-02 15:04:05"), entry.Level, entry.File, entry.Line, entry.Message)
}

func main() {
    log.SetWriters(MyCustomStdWriter) // Replace default writer

    log.Info("Using custom standard logger")
    log.Error("This is an error message")
}
```

With `SetWriters`, you can completely replace the behavior of the logger.

---

## 5. File Logging

The log library provides a file-based logging system with rotation and cleanup capabilities through the `file` subpackage.

### Features:
- Automatic daily log rotation
- Configurable file naming with date/time placeholders
- Automatic cleanup of old log files
- Custom log entry formatting
- Thread-safe for concurrent logging

### Example: Basic File Logger
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

### Example: Custom File Log Format
```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/log"
    "github.com/getevo/evo/v2/lib/log/file"
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

For more details, see the [File Logger](file_logger.md) documentation.

---
#### [< Table of Contents](https://github.com/getevo/evo#table-of-contents)
