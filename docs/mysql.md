# MySQL Driver

EVO provides a MySQL/MariaDB driver via `lib/mysql`. It integrates with GORM and the EVO schema migration system, with automatic MariaDB detection.

## Installation

```go
import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/mysql"
)
```

## Quick Start

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/mysql"
)

func main() {
    if err := evo.Setup(mysql.Driver{}); err != nil {
        panic(err)
    }
    evo.Run()
}
```

## Configuration (`config.yml`)

```yaml
Database:
  Enabled: true
  Type: mysql
  Server: "127.0.0.1:3306"    # host:port
  Username: "root"
  Password: "secret"
  Database: "myapp"
  SSLMode: "false"             # false | true
  Params: "charset=utf8mb4&parseTime=True&loc=Local"
  Debug: 3                     # 1:silent 2:warn 3:error 4:info
  MaxOpenConns: "100"
  MaxIdleConns: "10"
  ConnMaxLifTime: "1h"
  SlowQueryThreshold: "500ms"
```

### Recommended `Params` for MySQL

```
charset=utf8mb4&parseTime=True&loc=Local
```

- `charset=utf8mb4` — full Unicode support including emoji.
- `parseTime=True` — scan `DATE`/`DATETIME` into `time.Time`.
- `loc=Local` — use the server's local timezone.

### TiDB

TiDB is wire-compatible with MySQL. Use the same `mysql` driver and point `Server` at your TiDB cluster:

```yaml
Database:
  Type: mysql
  Server: "tidb-host:4000"
```

## Driver API

### `mysql.Driver`

Implements `db.Driver`. Pass it to `evo.Setup()`:

```go
evo.Setup(mysql.Driver{})
```

The driver:
1. Builds the DSN: `user:password@tcp(host:port)/database?params`
2. Opens a GORM connection.
3. Runs `SELECT VERSION()` to detect MariaDB vs MySQL and stores the result for use by the JSON type system.
4. Registers `MySQLDialect` for schema-aware DDL and migration.

### `mysql.RegisterDialect()`

Registers the MySQL dialect without opening a connection. Use in tests or standalone scripts:

```go
package main

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "github.com/getevo/evo/v2/lib/mysql"
    "github.com/getevo/evo/v2/lib/db/schema"
)

func main() {
    mysql.RegisterDialect()

    dsn := "root:password@tcp(localhost:3306)/myapp?parseTime=True"
    db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    schema.UseModels(&MyModel{})

    stmts := mysql.Driver{}.GetMigrationScript(db)
    for _, s := range stmts {
        println(s)
    }
}
```

## GORM model example

```go
type Article struct {
    ID        uint      `gorm:"primaryKey;autoIncrement"`
    Title     string    `gorm:"size:255;not null"`
    Body      string    `gorm:"type:longtext"`
    Published bool      `gorm:"default:false;index"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Register the model and run migrations:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db/schema"
    "github.com/getevo/evo/v2/lib/mysql"
)

func main() {
    evo.Setup(mysql.Driver{})
    schema.UseModels(&Article{})
    evo.Run()
}
```

Run with migration flag:

```shell
./myapp --migration-do          # apply migrations
./myapp --migration-dry-run     # print SQL without executing
./myapp --migration-dump        # dump CREATE TABLE DDL
```

## Database operations

```go
import "github.com/getevo/evo/v2/lib/db"

// Insert
article := Article{Title: "Hello World", Body: "Content here"}
db.Create(&article)

// Query
var articles []Article
db.Where("published = ?", true).Order("created_at DESC").Find(&articles)

// Update
db.Model(&article).Updates(Article{Title: "Updated Title", Published: true})

// Delete (soft-delete if model has DeletedAt)
db.Delete(&article)

// Raw SQL
var count int64
db.Raw("SELECT COUNT(*) FROM articles WHERE published = ?", true).Scan(&count)
```

## MySQL-specific features

### JSON columns

```go
import "github.com/getevo/evo/v2/lib/db/types"

type Config struct {
    ID       uint
    Name     string
    Settings types.JSON `gorm:"type:json"`
}
```

The `types.JSON` type is automatically serialized/deserialized and works with both MySQL 5.7+ (`json`) and MariaDB (`longtext`-based JSON simulation).

### Full-text index

```go
type Post struct {
    ID      uint
    Title   string `gorm:"size:255"`
    Content string `gorm:"type:text"`
}

// In migration: CREATE FULLTEXT INDEX idx_posts_content ON posts(content);
```

### Enum column

```go
type User struct {
    ID   uint
    Role string `gorm:"type:enum('admin','user','guest');default:'user'"`
}
```

With validation:

```go
type User struct {
    ID   uint
    Role string `gorm:"type:enum('admin','user','guest');default:'user'" validation:"enum"`
}
```

### UUID primary key

```go
import "github.com/getevo/evo/v2/lib/db/types"

type Order struct {
    ID   types.UUID `gorm:"type:varchar(36);primaryKey"`
    Ref  string
}
```

## MariaDB detection

The driver automatically detects MariaDB by checking `SELECT VERSION()`:

```go
// internal (happens automatically in Driver.Open)
if strings.Contains(strings.ToLower(ver), "mariadb") {
    schema.SetConfig("mysql_engine", "mariadb")
} else {
    schema.SetConfig("mysql_engine", "mysql")
}
```

This affects JSON type handling — MariaDB uses `longtext` while MySQL 5.7+ uses the native `json` type.

## Health checks

```go
import (
    "context"
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/getevo/evo/v2/lib/mysql"
)

func main() {
    evo.Setup(mysql.Driver{})

    evo.OnHealthCheck(func() error {
        return db.Ping(context.Background(), db.GetInstance())
    })

    evo.OnReadyCheck(func() error {
        result := db.HealthCheck(context.Background(), db.GetInstance())
        if !result.Healthy {
            return fmt.Errorf("database unhealthy: %s", result.Error)
        }
        return nil
    })

    evo.Run()
}
```

## Troubleshooting

### `Error 1049: Unknown database`

Create the database first, or use `CREATE DATABASE IF NOT EXISTS myapp;`.

### `Invalid utf8mb4 character`

Add `charset=utf8mb4` to `Params`.

### Slow queries logged

Adjust `SlowQueryThreshold` in config — queries exceeding this duration are logged as warnings.

### MariaDB JSON stored as text

This is expected. `types.JSON` handles serialization transparently for both engines.

## See Also

- [PostgreSQL Driver](pgsql.md)
- [Database](database.md)
- [Health Checks](health-checks.md)
- [Migration](migration.md)
