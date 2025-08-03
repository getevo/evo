# settings Library

The settings library provides a unified configuration management system that can load settings from multiple sources with a hierarchical override mechanism. It allows you to manage application configuration in a flexible and consistent way.

## Installation

```go
import "github.com/getevo/evo/v2/lib/settings"
```

## Features

- **Multiple Configuration Sources**: Load settings from environment variables, database, YAML files, and command-line arguments
- **Hierarchical Override**: Later sources override earlier ones in a predictable manner
- **Dot Notation**: Access nested settings using dot notation (e.g., "Database.Username")
- **Type Conversion**: Automatically convert settings to the desired type
- **Default Values**: Provide default values for settings that don't exist
- **Database Integration**: Store and retrieve settings from a database
- **Case Insensitive**: Keys are case-insensitive for easier access

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/settings"
)

func main() {
    // Initialize settings from all sources
    settings.Init()
    
    // Get a string setting with a default value
    dbUser := settings.Get("Database.Username", "root").String()
    fmt.Println("Database Username:", dbUser)
    
    // Get an integer setting with a default value
    port := settings.Get("HTTP.Port", 8080).Int()
    fmt.Println("HTTP Port:", port)
    
    // Get a boolean setting with a default value
    debug := settings.Get("App.Debug", false).Bool()
    fmt.Println("Debug Mode:", debug)
    
    // Get a float setting with a default value
    timeout := settings.Get("API.Timeout", 5.5).Float64()
    fmt.Println("API Timeout:", timeout)
}
```

### Setting Values Programmatically

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/settings"
)

func main() {
    // Set individual settings
    settings.Set("App.Name", "My Application")
    settings.Set("App.Version", "1.0.0")
    settings.Set("App.Debug", true)
    
    // Set multiple settings at once
    settings.SetMulti(map[string]any{
        "Database.Host":     "localhost",
        "Database.Port":     3306,
        "Database.Username": "admin",
        "Database.Password": "secret",
    })
    
    // Check if a setting exists
    exists, value := settings.Has("App.Name")
    if exists {
        fmt.Println("App Name:", value.String())
    }
    
    // Get all settings
    allSettings := settings.All()
    fmt.Printf("All Settings: %+v\n", allSettings)
}
```

### Working with YAML Configuration

Create a `config.yml` file:

```yaml
Database:
  Host: localhost
  Port: 3306
  Username: root
  Password: secret
  
HTTP:
  Host: 0.0.0.0
  Port: 8080
  
App:
  Name: My Application
  Debug: true
  LogLevel: info
```

Then in your code:

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/settings"
)

func main() {
    // Initialize settings (loads from config.yml by default)
    settings.Init()
    
    // Access nested settings using dot notation
    dbHost := settings.Get("Database.Host").String()
    dbPort := settings.Get("Database.Port").Int()
    appName := settings.Get("App.Name").String()
    
    fmt.Printf("Connecting to %s:%d for %s\n", dbHost, dbPort, appName)
    
    // Reload settings if configuration changes
    settings.Reload()
}
```

### Using Command-Line Arguments

You can override settings using command-line arguments in two formats:

```
./myapp Database.Username=admin Database.Password=secret
./myapp --HTTP.Port 9090 --App.Debug true
```

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/settings"
)

func main() {
    // Initialize settings (automatically loads command-line arguments)
    settings.Init()
    
    // Access settings that might be overridden by command-line arguments
    dbUser := settings.Get("Database.Username").String()
    httpPort := settings.Get("HTTP.Port").Int()
    
    fmt.Printf("Using database user: %s\n", dbUser)
    fmt.Printf("HTTP server running on port: %d\n", httpPort)
}
```

## How It Works

The settings library loads configuration from multiple sources in the following order:

1. **Environment Variables**: All environment variables are loaded first
2. **Database Settings**: If database is enabled, settings are loaded from the database
3. **YAML Configuration**: Settings are loaded from the YAML configuration file (default: `./config.yml`)
4. **Command-Line Arguments**: Command-line arguments override all previous settings

This order ensures that more specific sources (like command-line arguments) override more general ones (like environment variables).

The library normalizes all keys to uppercase and provides case-insensitive access. Nested settings can be accessed using dot notation, which is automatically handled by the library.

When retrieving a setting with `Get()`, you can provide a default value that will be used if the setting doesn't exist. The library automatically converts the setting to the appropriate type based on the default value or the requested conversion method.

For more detailed information, please refer to the source code and comments within the library.