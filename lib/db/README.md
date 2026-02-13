# DB Library

The DB library provides a comprehensive wrapper around GORM (Go Object Relational Mapper) for database operations in the EVO Framework. It simplifies database interactions and provides additional functionality for schema management, migrations, health checks, and connection management.

## Installation

```go
import "github.com/getevo/evo/v2/lib/db"
```

## Features

- ✅ **Context-Aware Database Access**: Proper context propagation for timeouts and cancellation
- ✅ **Health Checks**: Comprehensive health monitoring with connection pool statistics
- ✅ **Connection Management**: Functions for managing database connections
- ✅ **CRUD Operations**: Simple functions for creating, reading, updating, and deleting records
- ✅ **Query Building**: Methods for constructing complex database queries
- ✅ **Transaction Support**: Functions for handling database transactions
- ✅ **Schema Management**: Tools for managing database schemas and migrations
- ✅ **Model Registration**: Ability to register and manage database models
- ✅ **Multi-Dialect Support**: MySQL, PostgreSQL with shared schema abstractions

## Subdirectories

The DB library includes several subdirectories for specialized functionality:

- **entity**: Provides base entity structures and functionality
- **schema**: Tools for schema management and migrations (DB-agnostic)
- **types**: Custom data types for database interactions

## Quick Start

### Context-Aware Database Access (Recommended)

```go
import (
    "context"
    "github.com/getevo/evo/v2"
)

func handler(r *evo.Request) any {
    ctx := r.Context()
    db := evo.GetDB(ctx)  // Context-aware database instance

    var users []User
    db.Find(&users)

    return users
}
```

### Health Check Endpoint

```go
import (
    "github.com/getevo/evo/v2/lib/db"
)

func healthHandler(r *evo.Request) any {
    ctx := r.Context()
    database := evo.GetDBO()

    result := db.HealthCheck(ctx, database)
    if !result.Healthy {
        r.Status(503)
        return map[string]any{
            "status": "unhealthy",
            "error": result.Error,
        }
    }

    return map[string]any{
        "status": "healthy",
        "response_time_ms": result.ResponseTime.Milliseconds(),
        "connections_open": result.ConnectionsOpen,
        "connections_idle": result.ConnectionsIdle,
    }
}
```

## Health Check API

### `db.Ping(ctx context.Context, db *gorm.DB) error`

Simple ping to check if database is alive.

```go
if err := db.Ping(ctx, database); err != nil {
    log.Error("Database unreachable:", err)
}
```

### `db.HealthCheck(ctx context.Context, db *gorm.DB) HealthCheckResult`

Comprehensive health check with connection pool statistics.

```go
result := db.HealthCheck(ctx, database)
fmt.Printf("Healthy: %v, Response Time: %v\n", result.Healthy, result.ResponseTime)
fmt.Printf("Connections: %d open, %d in use, %d idle\n",
    result.ConnectionsOpen, result.ConnectionsInUse, result.ConnectionsIdle)
```

**HealthCheckResult Structure:**
```go
type HealthCheckResult struct {
    Healthy          bool          // Overall health status
    ResponseTime     time.Duration // Ping response time
    ConnectionsOpen  int           // Total open connections
    ConnectionsInUse int           // Connections currently in use
    ConnectionsIdle  int           // Idle connections available
    MaxOpenConns     int           // Maximum allowed connections
    Error            string        // Error message if unhealthy
}
```

### `db.WaitForDB(ctx context.Context, db *gorm.DB, maxRetries int, retryInterval time.Duration) error`

Wait for database to become available with retries. Useful during application startup.

```go
ctx := context.Background()
err := db.WaitForDB(ctx, database, 10, 2*time.Second)
if err != nil {
    log.Fatal("Database not available:", err)
}
```

### `db.GetConnectionStats(db *gorm.DB) (sql.DBStats, error)`

Get detailed connection pool statistics.

```go
stats, err := db.GetConnectionStats(database)
if err == nil {
    fmt.Printf("Max Open Connections: %d\n", stats.MaxOpenConnections)
    fmt.Printf("Connections In Use: %d\n", stats.InUse)
}
```

### `db.CloseConnection(db *gorm.DB) error`

Gracefully close database connection.

```go
defer db.CloseConnection(database)
```

## Usage Examples

### Basic CRUD Operations

```go
package main

import (
    "github.com/getevo/evo/v2"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

func main() {
    evo.Setup()

    db := evo.GetDBO()

    // Register models
    schema.UseModel(db, User{})

    // Create a new user
    user := User{Name: "John Doe", Age: 30}
    db.Create(&user)

    // Find a user
    var foundUser User
    db.First(&foundUser, user.ID)

    // Update a user
    db.Model(&foundUser).Update("Age", 31)

    // Delete a user
    db.Delete(&foundUser)
}
```

### Context-Aware Query with Timeout

```go
func GetUserByID(ctx context.Context, id int64) (*User, error) {
    // Create a 5-second timeout context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var user User
    db := evo.GetDB(ctx)

    if err := db.Where("id = ?", id).First(&user).Error; err != nil {
        return nil, err
    }

    return &user, nil
}
```

### Using Transactions

```go
package main

import (
    "context"
    "github.com/getevo/evo/v2"
)

func createUsers(ctx context.Context) error {
    db := evo.GetDB(ctx)

    // Start a transaction
    return db.Transaction(func(tx *gorm.DB) error {
        // Perform operations within the transaction
        if err := tx.Create(&User{Name: "User 1"}).Error; err != nil {
            // Return error will rollback the transaction
            return err
        }

        if err := tx.Create(&User{Name: "User 2"}).Error; err != nil {
            return err
        }

        // Return nil will commit the transaction
        return nil
    })
}
```

### Graceful Startup with Database Wait

```go
func main() {
    evo.Setup()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    database := evo.GetDBO()

    // Wait up to 30 seconds for database
    if err := db.WaitForDB(ctx, database, 15, 2*time.Second); err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    log.Info("Database connected successfully")

    evo.Run()
}
```

### Monitor Connection Pool

```go
func monitorConnectionPool() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    database := evo.GetDBO()

    for range ticker.C {
        stats, err := db.GetConnectionStats(database)
        if err != nil {
            log.Error("Failed to get stats:", err)
            continue
        }

        log.Info("Connection Pool Stats",
            "open", stats.OpenConnections,
            "in_use", stats.InUse,
            "idle", stats.Idle,
            "wait_count", stats.WaitCount,
            "wait_duration", stats.WaitDuration,
        )

        // Alert if connection pool is exhausted
        if stats.InUse >= stats.MaxOpenConnections {
            log.Warning("Connection pool exhausted!")
        }
    }
}
```

### Schema Migrations

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db/schema"
)

func main() {
    evo.Setup()

    db := evo.GetDBO()

    // Register models
    schema.UseModel(db, User{})

    // Get migration script (dry run)
    scripts := schema.GetMigrationScript(db)
    for _, script := range scripts {
        fmt.Println(script)
    }

    // Or perform migration directly
    err := evo.DoMigration()
    if err != nil {
        log.Fatal("Migration failed:", err)
    }
}
```

## Advanced Query Building

```go
package main

import (
    "github.com/getevo/evo/v2"
)

func main() {
    var users []User

    ctx := context.Background()
    db := evo.GetDB(ctx)

    // Complex query with conditions, ordering, and limits
    db.Where("age > ?", 18).
       Where("name LIKE ?", "%Doe%").
       Order("age DESC").
       Limit(10).
       Find(&users)

    // Using scopes for reusable query parts
    db.Scopes(ActiveUsers, AgeGreaterThan(18)).Find(&users)
}

// Scope example
func ActiveUsers(d *gorm.DB) *gorm.DB {
    return d.Where("active = ?", true)
}

func AgeGreaterThan(age int) func(*gorm.DB) *gorm.DB {
    return func(d *gorm.DB) *gorm.DB {
        return d.Where("age > ?", age)
    }
}
```

## Best Practices

### 1. Always Use Context

```go
// ✅ Good - with context
db := evo.GetDB(r.Context())

// ❌ Avoid - without context
db := evo.GetDBO()
```

### 2. Set Query Timeouts

```go
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()

db := evo.GetDB(ctx)
```

### 3. Monitor Health in Production

```go
// Register health check endpoint
evo.Get("/health", healthCheckHandler)
evo.Get("/health/db", databaseHealthHandler)
```

### 4. Handle Graceful Shutdown

```go
func shutdown() {
    database := evo.GetDBO()
    if err := db.CloseConnection(database); err != nil {
        log.Error("Error closing database:", err)
    }
}
```

### 5. Use Connection Pool Wisely

```go
// Configure connection pool in config.yml
Database:
  MaxIdleConns: 10
  MaxOpenConns: 100
  ConnMaxLifTime: 1h
```

## Configuration

Database configuration is managed through the settings package:

```yaml
Database:
  Enabled: true
  Type: mysql  # or postgres
  Server: localhost:3306
  Username: root
  Password: password
  Database: mydb
  MaxIdleConns: 10
  MaxOpenConns: 100
  ConnMaxLifTime: 1h
  SlowQueryThreshold: 500ms
  Debug: 3  # 0=Silent, 2=Error, 3=Warn, 4=Info
  SSLMode: false
  Params: "charset=utf8mb4&parseTime=True&loc=Local"
```

## Related Libraries

- **entity**: Base entity structures and functionality
- **schema**: Schema management and migrations (database-agnostic)
- **types**: Custom data types for database interactions
- **settings**: Configuration management
- **log**: Logging system

## See Also

- [Schema Package](./schema/README.md) - Database schema and migrations
- [Settings Package](../settings/README.md) - Configuration management
- [GORM Documentation](https://gorm.io/docs/) - GORM ORM documentation
