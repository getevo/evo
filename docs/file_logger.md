# File Logger

This library extends your logging capabilities by adding a file-based logger to your existing log package. It supports features like log rotation at midnight, configurable file paths, customizable log formatting, and automatic cleanup of old logs based on expiration settings.

## Features
- **Automatic Log Rotation**: Creates a new log file every midnight.
- **Customizable Log Format**: Allows custom formatting of log entries.
- **Optional Expiration**: Automatically deletes old log files after a defined number of days.
- **Thread-Safe**: Ensures safe concurrent writes to the log file.
- **Default Configuration**: Works out-of-the-box without requiring any configuration.

---

## Installation
Import the logger package into your project:

```go
import "github.com/getevo/evo/v2/lib/log/file" // Replace with the correct import path
import "github.com/getevo/evo/v2/lib/log"
```

---

## Usage
### Adding the File Logger to Log Writers
You can integrate the file logger into your log package's writers using the `NewFileLogger` function. Pass an optional `Config` struct to customize its behavior.

Here is an example:

```go
package main

import (
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/log/file" // Replace with the correct import path
)

func main() {
	// Add the file logger with custom configuration
	log.AddWriter(file.NewFileLogger(file.Config{
		Path:       "./logs",              // Directory to store logs
		FileName:   "app_%y-%m-%d.log",    // Filename template with wildcards
		Expiration: 7,                      // Keep logs for 7 days
		LogFormat:  nil,                    // Use default log format
	}))

	// Example logs
	log.Info("Application started")
	log.Warning("This is a warning message")
	log.Error("An error occurred")
}
```

---

## Configuration
The `Config` struct allows you to customize the behavior of the logger. Here's a breakdown of the available fields:

| Field       | Type                                | Default                          | Description                                                                 |
|-------------|-------------------------------------|----------------------------------|-----------------------------------------------------------------------------|
| `Path`      | `string`                            | Current working directory        | Directory where the log files will be saved.                                |
| `FileName`  | `string`                            | `executable_name.log`            | Filename template. Supports `%y`, `%m`, `%d` for year, month, and day.      |
| `Expiration`| `int`                               | `0`                              | Number of days to keep log files. `<= 0` means no cleanup of old logs.      |
| `LogFormat` | `func(entry *log.Entry) string`     | Default format                   | Custom function to format log entries. Defaults to a standard log format.   |

### Filename Wildcards
The `FileName` field supports the following wildcards:
- `%y` → Year (e.g., `2024`)
- `%m` → Month (e.g., `04`)
- `%d` → Day (e.g., `25`)

Example: `app_%y-%m-%d.log` → `app_2024-04-25.log`

---

## Default Behavior
If no configuration is provided, the logger uses the following defaults:
1. **Path**: The current working directory.
2. **FileName**: The executable name with a `.log` extension.
3. **Expiration**: No automatic cleanup of old logs.
4. **LogFormat**: A standard log format with the following structure:
   ```
   YYYY-MM-DD HH:MM:SS [LEVEL] file:line message
   ```

**Example**:
```
2024-04-25 14:30:01 [INFO] main.go:45 Application started
```

---

## Log Rotation
- At midnight, the logger automatically rotates the log file based on the current date.
- A new log file is created using the `FileName` template.

---

## Expiration
If the `Expiration` field is set (e.g., 7 days), log files older than the specified number of days are automatically deleted during log rotation.

**Example**:
- Set `Expiration: 7` → Log files older than 7 days will be removed.

---

## Thread Safety
All writes to the log file are protected using a `sync.Mutex`, ensuring thread-safe operations in concurrent environments.

---

## Example File Structure
With the configuration `Path: "./logs"` and `FileName: "app_%y-%m-%d.log"`, the log files will be structured as follows:

```
./logs/
   ├── app_2024-04-25.log
   ├── app_2024-04-26.log
   ├── app_2024-04-27.log
   └── ...
```