# Settings Package

Thread-safe, multi-source configuration management system for Go applications.

## Features

- ✅ **Thread-Safe**: All operations protected by `sync.RWMutex`
- ✅ **Multi-Source Loading**: Environment variables, YAML, Database, CLI arguments
- ✅ **Type Conversion**: Automatic conversion to string, int, bool, duration, size, etc.
- ✅ **Dot Notation**: Hierarchical keys (e.g., `DATABASE.HOST`)
- ✅ **Change Notifications**: Callbacks for config changes and reloads
- ✅ **Persistence**: Save settings back to YAML or Database
- ✅ **Case-Insensitive**: Keys normalized to uppercase

## Loading Priority

Settings are loaded in this order (highest priority last):

1. **Environment Variables** (lowest)
2. **Database Settings** (if enabled)
3. **YAML Configuration File**
4. **Command-Line Arguments** (highest)

Later sources override earlier ones.

## Quick Start

### Basic Usage

```go
import "github.com/getevo/evo/v2/lib/settings"

// Initialize settings (call once at startup)
settings.Init()

// Get settings with type conversion
host := settings.Get("DATABASE.HOST", "localhost").String()
port := settings.Get("DATABASE.PORT", 3306).Int()
timeout := settings.Get("HTTP.TIMEOUT", "30s").Duration()
maxSize := settings.Get("UPLOAD.MAX_SIZE", "10MB").SizeInBytes()
enabled := settings.Get("FEATURE.ENABLED", false).Bool()

// Set settings (automatically persists to database if enabled)
settings.Set("APP.NAME", "MyApp")
settings.Set("APP.VERSION", "1.0.0")

// Set multiple at once (automatically persists to database if enabled)
settings.SetMulti(map[string]any{
    "DATABASE.HOST": "localhost",
    "DATABASE.PORT": 3306,
    "DATABASE.NAME": "mydb",
})
```

See full documentation in the package godoc.

## Advanced Features

### Change Notifications

```go
// Watch for reload events
settings.OnReload(func() {
    log.Info("Configuration reloaded!")
})

// Watch for specific setting changes
settings.Track("DATABASE.HOST", func(key string, oldValue, newValue any) {
    log.Info("Database host changed:", oldValue, "->", newValue)
})

// Watch all database settings (wildcard)
settings.Track("DATABASE.*", func(key string, oldValue, newValue any) {
    log.Info("Database setting changed:", key, "from", oldValue, "to", newValue)
    db.Reconnect() // Reconnect on any DB config change
})

// Watch all settings
settings.Track("*", func(key string, oldValue, newValue any) {
    log.Info("Config changed:", key, "=", newValue)
})
```

### Persistence

Settings are automatically persisted to the database when you call `Set()` or `SetMulti()` if database settings are enabled. You can also manually save all settings to YAML or database:

```go
// Automatic database persistence (if db enabled)
settings.Set("APP.NAME", "MyApp")  // Saved to DB immediately

// Manual YAML persistence
settings.SaveToYAML("./config.yml")

// Manual database persistence (saves all settings)
settings.SaveToDB()
```

**Note:** When database is enabled:
- `Set()` and `SetMulti()` automatically update the database
- `SaveToDB()` saves ALL current in-memory settings to database
- `SaveToYAML()` is always manual (not automatic)

## Type Conversions

```go
// String, Int, Bool, Float
name := settings.Get("APP.NAME").String()
port := settings.Get("PORT").Int()
enabled := settings.Get("ENABLED").Bool()

// Duration and Time
timeout := settings.Get("TIMEOUT").Duration() // "30s" -> 30 * time.Second

// Size in Bytes
maxSize := settings.Get("MAX_SIZE").SizeInBytes() // "10MB" -> 10485760
```

## Thread Safety

All operations are thread-safe and can be called from multiple goroutines.

## Key Normalization

Keys are case-insensitive and normalized:
- `database.host`, `DATABASE.HOST`, `DATABASE_HOST` all access the same setting
- Non-alphanumeric characters converted to underscores
