# Database

EVO provides a unified database layer built on [GORM](https://gorm.io/docs). It supports MySQL/MariaDB and PostgreSQL through pluggable drivers, with built-in schema migration, health checks, and connection pool management.

## Supported databases

| Database | Driver package | Notes |
|---|---|---|
| MySQL 5.7+ | `lib/mysql` | Full support, auto-detects MariaDB |
| MariaDB | `lib/mysql` | Detected automatically via `SELECT VERSION()` |
| TiDB | `lib/mysql` | Wire-compatible with MySQL |
| PostgreSQL 12+ | `lib/pgsql` | Full support, schema-aware |

## Quick Start

### MySQL

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/mysql"
)

func main() {
    evo.Setup(mysql.Driver{})
    evo.Run()
}
```

### PostgreSQL

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    evo.Setup(pgsql.Driver{})
    evo.Run()
}
```

## Configuration (`config.yml`)

### MySQL / MariaDB / TiDB

```yaml
Database:
  Enabled: true
  Type: mysql
  Server: "127.0.0.1:3306"
  Username: "root"
  Password: "secret"
  Database: "myapp"
  SSLMode: "false"
  Params: "charset=utf8mb4&parseTime=True&loc=Local"
  Debug: 3
  MaxOpenConns: "100"
  MaxIdleConns: "10"
  ConnMaxLifTime: "1h"
  SlowQueryThreshold: "500ms"
```

### PostgreSQL

```yaml
Database:
  Enabled: true
  Type: postgres
  Server: "localhost:5432"
  Username: "postgres"
  Password: "secret"
  Database: "myapp"
  Schema: "public"       # PostgreSQL schema (default: public)
  SSLMode: "disable"     # disable | require
  Params: ""
  Debug: 3
  MaxOpenConns: "100"
  MaxIdleConns: "10"
  ConnMaxLifTime: "1h"
  SlowQueryThreshold: "500ms"
```

### Configuration reference

| Key | Type | Description |
|---|---|---|
| `Enabled` | bool | Enable/disable the database connection |
| `Type` | string | `mysql` or `postgres` |
| `Server` | string | `host:port` |
| `Username` | string | Database user |
| `Password` | string | Database password |
| `Database` | string | Database name |
| `Schema` | string | PostgreSQL schema name (ignored for MySQL) |
| `SSLMode` | string | SSL mode (`false`/`disable`, `true`/`require`) |
| `Params` | string | Additional DSN parameters appended verbatim |
| `Debug` | int | Log verbosity: 1=silent, 2=warn, 3=error, 4=info |
| `MaxOpenConns` | int | Max concurrent connections |
| `MaxIdleConns` | int | Max idle connections in pool |
| `ConnMaxLifTime` | duration | Max connection lifetime |
| `SlowQueryThreshold` | duration | Log queries slower than this |

## Accessing the database

### Via `db` package (recommended)

```go
import "github.com/getevo/evo/v2/lib/db"

var users []User
db.Where("active = ?", true).Find(&users)
```

### Via `evo.GetDBO()`

```go
import "github.com/getevo/evo/v2"

gormDB := evo.GetDBO()
gormDB.Find(&users)
```

### With context

```go
import (
    "context"
    "github.com/getevo/evo/v2/lib/db"
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

var user User
db.WithContext(ctx).First(&user, id)
```

## CRUD operations

### Create

```go
user := User{Name: "Alice", Email: "alice@example.com"}
result := db.Create(&user)
if result.Error != nil {
    // handle error
}
fmt.Println(user.ID) // auto-set after insert
```

### Read

```go
// By primary key
var user User
db.First(&user, 1)

// With conditions
var users []User
db.Where("email LIKE ?", "%@example.com").
    Order("created_at DESC").
    Limit(10).
    Find(&users)

// Count
var total int64
db.Model(&User{}).Where("active = ?", true).Count(&total)
```

### Update

```go
// Update specific columns
db.Model(&user).Update("name", "Alice Smith")

// Update multiple columns
db.Model(&user).Updates(User{Name: "Alice Smith", Active: true})

// Update with map (zero values are also updated)
db.Model(&user).Updates(map[string]any{
    "name":   "Alice Smith",
    "active": false,
})
```

### Delete

```go
// Delete by value
db.Delete(&user)

// Delete by condition
db.Where("created_at < ?", time.Now().AddDate(-1, 0, 0)).Delete(&User{})
```

## Transactions

```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err // auto-rollback
    }
    if err := tx.Model(&product).Update("stock", gorm.Expr("stock - ?", 1)).Error; err != nil {
        return err
    }
    return nil // auto-commit
})
```

Manual transaction:

```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&record).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

## Raw SQL

```go
// Execute
db.Exec("UPDATE users SET active = ? WHERE last_login < ?", false, cutoff)

// Query with scan
type Result struct {
    Name  string
    Count int64
}
var results []Result
db.Raw("SELECT name, COUNT(*) AS count FROM orders GROUP BY name").Scan(&results)
```

## Driver interface

You can implement custom drivers by satisfying the `db.Driver` interface:

```go
// lib/db/driver.go
type Driver interface {
    Name() string
    Open(config DriverConfig, gormConfig *gorm.Config) (*gorm.DB, error)
    GetMigrationScript(db *gorm.DB) []string
}

type DriverConfig struct {
    Server   string // host:port
    Username string
    Password string
    Database string
    Schema   string  // PostgreSQL schema name
    SSLMode  string
    Params   string  // extra DSN params
}
```

Register and retrieve the driver:

```go
db.RegisterDriver(myDriver)
driver := db.GetDriver()
```

## Connection management

### Check if enabled

```go
if db.IsEnabled() {
    // safe to use database
}
```

### Get raw `*sql.DB` for pool management

```go
sqlDB, err := db.GetInstance().DB()
if err != nil {
    return err
}

// Set pool settings at runtime
sqlDB.SetMaxOpenConns(50)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(30 * time.Minute)
```

### Close connection gracefully

```go
import "github.com/getevo/evo/v2/lib/db"

if err := db.CloseConnection(db.GetInstance()); err != nil {
    log.Error("failed to close db", "error", err)
}
```

## Health checks

The `db` package provides functions for Kubernetes probes and startup checks.

### `db.Ping` — liveness check

```go
err := db.Ping(context.Background(), db.GetInstance())
```

### `db.HealthCheck` — detailed stats

```go
result := db.HealthCheck(context.Background(), db.GetInstance())
// result.Healthy          bool
// result.ResponseTime     time.Duration
// result.ConnectionsOpen  int
// result.ConnectionsInUse int
// result.ConnectionsIdle  int
// result.MaxOpenConns     int
// result.Error            string
```

### `db.WaitForDB` — startup retry loop

```go
// Retry up to 10 times with 2s interval
err := db.WaitForDB(context.Background(), db.GetInstance(), 10, 2*time.Second)
```

### `db.GetConnectionStats`

```go
stats, err := db.GetConnectionStats(db.GetInstance())
// stats is sql.DBStats with full pool info
```

### Integrating with EVO health system

```go
evo.OnHealthCheck(func() error {
    return db.Ping(context.Background(), db.GetInstance())
})

evo.OnReadyCheck(func() error {
    return db.WaitForDB(context.Background(), db.GetInstance(), 5, time.Second)
})
```

## Migration hooks

Register callbacks that run before and after schema migrations:

```go
db.OnBeforeMigration(func(gormDB *gorm.DB) {
    log.Info("migration starting")
})

db.OnAfterMigration(func(gormDB *gorm.DB) {
    log.Info("migration complete")
    seedDatabase(gormDB)
})
```

## Dialect helpers

```go
// Get dialect name: "mysql" or "postgres"
name := db.DialectName()

// Quote an identifier using the correct dialect syntax
// MySQL: `tablename`  PostgreSQL: "tablename"
quoted := db.QuoteIdent("tablename")

// Get query SQL without executing (useful for debugging)
sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
    return tx.Where("active = ?", true).Find(&User{})
})
```

## Model registration

Register GORM models for migration:

```go
import "github.com/getevo/evo/v2/lib/db/schema"

schema.UseModels(&User{}, &Order{}, &Product{})
```

Retrieve registered models:

```go
models := evo.Models()          // []schema.Model
model  := evo.GetModel("users") // *schema.Model by table name
```

## See Also

- [MySQL Driver](mysql.md)
- [PostgreSQL Driver](pgsql.md)
- [Database Migration](migration.md)
- [Health Checks](health-checks.md)
- [GORM Documentation](https://gorm.io/docs)
