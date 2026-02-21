# Configuration and Settings

EVO provides a thread-safe, multi-source configuration system via `lib/settings`. Settings are loaded from environment variables, YAML files, databases, and command-line arguments, with a clear priority order.

## Loading priority (highest wins)

1. **Command-line arguments** (`--KEY=value`)
2. **YAML configuration file** (`./config.yml` or path from `-c` flag)
3. **Database settings** (if database is enabled)
4. **Environment variables**

## Quick Start

```go
import "github.com/getevo/evo/v2/lib/settings"

// Read a value (with optional default)
host := settings.Get("DATABASE.HOST", "localhost").String()
port := settings.Get("DATABASE.PORT", 5432).Int()
debug := settings.Get("APP.DEBUG", false).Bool()

// Set a value at runtime
settings.Set("APP.NAME", "MyService")

// Check existence
exists, value := settings.Has("FEATURE.ENABLED")
```

EVO calls `settings.Init()` automatically during `evo.Setup()`.

## YAML configuration

The default config file is `./config.yml`. Override with `-c /path/to/config.yml`.

```yaml
# config.yml
HTTP:
  Host: "0.0.0.0"
  Port: "8080"

Database:
  Enabled: true
  Type: postgres
  Server: "localhost:5432"
  Username: "postgres"
  Password: "secret"
  Database: "myapp"
  Schema: "public"
  SSLMode: "disable"
  Debug: 3
  MaxOpenConns: "100"
  MaxIdleConns: "10"
  ConnMaxLifTime: "1h"
  SlowQueryThreshold: "500ms"

App:
  Name: "MyApp"
  Version: "1.0.0"
  Debug: "false"
```

Keys are **case-insensitive** and **dot-notation** is supported. `DATABASE.HOST`, `database.host`, and `Database.Host` all refer to the same key internally.

## Environment variables

Any environment variable is available as a setting. Dots and hyphens are treated as separators:

```bash
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export APP_NAME=MyService
```

```go
settings.Get("DATABASE.HOST").String() // "localhost"
settings.Get("APP.NAME").String()      // "MyService"
```

## Command-line arguments

Override any setting at startup:

```bash
./myapp --DATABASE.HOST=prod-db --DATABASE.PORT=5432 --APP.DEBUG=true
```

Select a custom config file:

```bash
./myapp -c /etc/myapp/config.yml
```

## API Reference

### `settings.Get(key, default...) generic.Value`

Retrieves a setting. Returns `generic.Value` which converts to any type.

```go
// String
name := settings.Get("APP.NAME", "default").String()

// Int
port := settings.Get("HTTP.PORT", 8080).Int()

// Bool
debug := settings.Get("APP.DEBUG", false).Bool()

// Duration
timeout := settings.Get("HTTP.TIMEOUT", "30s").Duration()

// Float
ratio := settings.Get("CACHE.RATIO", 0.75).Float64()
```

### `settings.Has(key) (bool, generic.Value)`

Checks if a key exists.

```go
exists, value := settings.Has("FEATURE.ENABLED")
if exists {
    fmt.Println("Feature flag:", value.Bool())
}
```

### `settings.Set(key, value)`

Updates a setting at runtime. Persists to database if database settings are enabled. Triggers registered change callbacks.

```go
settings.Set("APP.MAINTENANCE", true)
settings.Set("CACHE.TTL", "5m")
settings.Set("MAX.RETRIES", 3)
```

### `settings.SetMulti(map[string]any)`

Updates multiple settings efficiently in a single lock cycle.

```go
settings.SetMulti(map[string]any{
    "DATABASE.HOST": "new-host",
    "DATABASE.PORT": 5433,
    "APP.DEBUG":     false,
})
```

### `settings.All() map[string]any`

Returns a snapshot of all current settings.

```go
all := settings.All()
for key, value := range all {
    fmt.Printf("%s = %v\n", key, value)
}
```

### `settings.Delete(key) bool`

Removes a setting. Returns true if it existed.

```go
removed := settings.Delete("TEMP.VALUE")
```

### `settings.SaveToYAML(filename) error`

Saves all current settings to a YAML file.

```go
settings.Set("DATABASE.HOST", "localhost")
settings.Set("DATABASE.PORT", 5432)
settings.SaveToYAML("./config.yml")
// Result:
// database:
//   host: localhost
//   port: 5432
```

### `settings.SaveToDB() error`

Saves all settings to the database (requires database to be enabled).

```go
if err := settings.SaveToDB(); err != nil {
    log.Error("failed to save settings", "error", err)
}
```

### `settings.Reload() error`

Hot-reloads all settings from all sources and triggers `OnReload` callbacks. Useful for configuration changes without restart.

```go
if err := settings.Reload(); err != nil {
    log.Error("reload failed", "error", err)
}
```

## Change tracking

### `settings.OnReload(func())`

Called whenever `settings.Reload()` runs (also called at startup).

```go
settings.OnReload(func() {
    log.Info("configuration reloaded")
    cache.Clear()
})
```

### `settings.Track(pattern, func())`

Called when a matching setting changes. Supports wildcards.

```go
// Exact key
settings.Track("DATABASE.HOST", func() {
    log.Info("database host changed — reconnecting...")
    reconnectDB()
})

// Wildcard: any DATABASE.* key
settings.Track("DATABASE.*", func() {
    log.Info("database config changed")
    reconnectDB()
})

// Global wildcard: any setting
settings.Track("*", func() {
    log.Info("configuration changed")
    app.Reload()
})
```

**Note**: The callback receives no parameters. Read the new value inside the callback:

```go
settings.Track("HTTP.PORT", func() {
    newPort := settings.Get("HTTP.PORT").Int()
    log.Info("port changed", "port", newPort)
})
```

The callback is also invoked once immediately at registration time.

## Key normalization

All keys are normalized before storage:
- Converted to **uppercase**
- Non-alphanumeric characters replaced with **underscore**

```
"database.host"   → "DATABASE_HOST"
"HTTP-Port"       → "HTTP_PORT"
"app/name"        → "APP_NAME"
```

Use any separator in your code — it all maps to the same key:

```go
settings.Get("database.host")
settings.Get("DATABASE_HOST")
settings.Get("database-host")
// All return the same value
```

## Database settings

When the database is enabled, settings are loaded from and saved to a `settings` table. The table is created automatically.

```go
// After db is enabled and evo.Setup() is called:

// Persist a setting to the database
settings.Set("FEATURE.DARK_MODE", true)  // auto-saved to DB

// Read a setting that was changed in the database
settings.Reload() // reloads all sources including DB
value := settings.Get("FEATURE.DARK_MODE").Bool()
```

## `generic.Value` type conversions

`settings.Get` returns a `generic.Value`. Available conversion methods:

```go
v := settings.Get("MY.KEY", "default")

v.String()    // string
v.Int()       // int
v.Int64()     // int64
v.Float64()   // float64
v.Bool()      // bool
v.Duration()  // time.Duration  (parses "30s", "5m", "1h")
v.Time()      // time.Time
v.Bytes()     // []byte
v.IsNil()     // bool
```

## Practical examples

### Database connection with settings

```go
import (
    "github.com/getevo/evo/v2/lib/settings"
    "github.com/getevo/evo/v2/lib/pgsql"
    "github.com/getevo/evo/v2"
)

func main() {
    evo.Setup(pgsql.Driver{})

    // Settings are loaded — DATABASE.* keys available
    host := settings.Get("DATABASE.SERVER").String()
    db   := settings.Get("DATABASE.DATABASE").String()
    log.Info("connected", "host", host, "db", db)

    evo.Run()
}
```

### Feature flags

```go
// In config.yml:
// Feature:
//   NewUI: "true"
//   Beta:  "false"

enabled := settings.Get("FEATURE.NEWUI", false).Bool()
if enabled {
    evo.Get("/dashboard", newDashboardHandler)
}
```

### Dynamic reconnection

```go
settings.Track("DATABASE.*", func() {
    log.Info("database config changed — reconnecting")
    // trigger reconnection logic
})
```

### Reactive app variables (AI client, API keys)

`settings.Track` is the recommended pattern for keeping app-level variables in sync with configuration. The callback fires once at registration (to initialize the value) and again whenever the setting changes:

```go
import (
    "github.com/getevo/evo/v2/lib/settings"
    openai "github.com/sashabaranov/go-openai"
)

var (
    APIKey string
    client *openai.Client
)

func (a App) Register() error {
    settings.Track("OPENAI.API_KEY", func() {
        APIKey = settings.Get("OPENAI.API_KEY").String()
        client = openai.NewClient(APIKey)
    })
    return nil
}
```

When `OPENAI.API_KEY` changes at runtime (e.g. via `settings.Set` or a config reload), the callback re-runs and `client` is replaced automatically — no restart required.

### Startup validation

```go
required := []string{"DATABASE.SERVER", "DATABASE.USERNAME", "DATABASE.PASSWORD", "APP.SECRET"}
for _, key := range required {
    if exists, _ := settings.Has(key); !exists {
        log.Fatal("missing required config: " + key)
    }
}
```

### Saving config back to YAML

```go
// Load, override, save
settings.Set("HTTP.PORT", "9090")
settings.Set("APP.DEBUG", "true")
settings.SaveToYAML("./config.yml")
```

## See Also

- [Database](database.md)
- [Args](args.md)
- [Logging](log.md)
