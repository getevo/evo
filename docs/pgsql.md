# PostgreSQL Driver

EVO provides a first-class PostgreSQL driver via `lib/pgsql`. It integrates with GORM and the EVO schema migration system, with full support for schemas, advisory locking, and constraint introspection.

## Installation

```go
import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/pgsql"
)
```

## Quick Start

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    if err := evo.Setup(pgsql.Driver{}); err != nil {
        panic(err)
    }
    evo.Run()
}
```

## Configuration (`config.yml`)

```yaml
Database:
  Enabled: true
  Type: postgres
  Server: "localhost:5432"    # host:port
  Username: "postgres"
  Password: "secret"
  Database: "myapp"
  Schema: "public"            # PostgreSQL schema (default: public)
  SSLMode: "disable"          # disable | require
  Params: ""                  # extra DSN params appended verbatim
  Debug: 3                    # 1:silent 2:warn 3:error 4:info
  MaxOpenConns: "100"
  MaxIdleConns: "10"
  ConnMaxLifTime: "1h"
  SlowQueryThreshold: "500ms"
```

### SSLMode values

| Value | Meaning |
|-------|---------|
| `disable` (default) | No TLS |
| `require` or `true` | Require TLS |

### Multi-schema / multi-tenant

Run separate instances (processes) each pointing to a different schema:

```yaml
# Tenant A
Database:
  Schema: "tenant_a"

# Tenant B
Database:
  Schema: "tenant_b"
```

The EVO migration system, DDL generation, and all queries automatically use the configured schema.

## Driver API

### `pgsql.Driver`

Implements `db.Driver`. Pass it to `evo.Setup()`:

```go
evo.Setup(pgsql.Driver{})
```

The driver:
1. Builds the DSN from `DriverConfig`.
2. Opens a GORM connection using `gorm.io/driver/postgres`.
3. Registers `PGDialect` for schema-aware DDL and migration.

### `pgsql.RegisterDialect()`

Registers the PostgreSQL dialect without opening a connection. Use this in standalone migration scripts or tests where you manage the connection yourself:

```go
package main

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/getevo/evo/v2/lib/pgsql"
    "github.com/getevo/evo/v2/lib/db/schema"
)

func main() {
    pgsql.RegisterDialect() // register dialect without opening connection

    db, _ := gorm.Open(postgres.Open("host=localhost user=postgres ..."), &gorm.Config{})
    schema.UseModels(&MyModel{})

    // Generate migration SQL
    stmts := pgsql.Driver{}.GetMigrationScript(db)
    for _, s := range stmts {
        println(s)
    }
}
```

## PGDialect internals

`PGDialect` implements `schema.Dialect` and provides PostgreSQL-specific behaviour:

| Method | Description |
|--------|-------------|
| `Name()` | Returns `"postgres"` |
| `Quote(name)` | Wraps identifier in double quotes: `"tablename"` |
| `GetCurrentDatabase(db)` | Runs `SELECT current_database()` |
| `GetTableVersion(db, database, table)` | Reads version from `pg_class` comment |
| `SetTableVersionSQL(table, version)` | `COMMENT ON TABLE ... IS '...'` |
| `GetJoinConstraints(db, database)` | Introspects foreign keys via `pg_constraint` |
| `GenerateMigration(...)` | Generates full CREATE/ALTER DDL |
| `AcquireMigrationLock(db)` | `pg_advisory_lock(hashtext(...))` |
| `ReleaseMigrationLock(db)` | `pg_advisory_unlock(...)` |
| `BootstrapHistoryTable(db)` | Creates `schema_migration` table if absent |

### Advisory locking

The migration system acquires a PostgreSQL advisory lock before running migrations. This prevents concurrent migration runs in multi-replica deployments:

```
SELECT pg_advisory_lock(hashtext('schema_migration_lock'))
-- ... run migrations ...
SELECT pg_advisory_unlock(hashtext('schema_migration_lock'))
```

### Migration history table

Automatically created on first migration:

```sql
CREATE TABLE IF NOT EXISTS "schema_migration" (
  "id"               BIGSERIAL PRIMARY KEY,
  "hash"             CHAR(32)     NOT NULL,
  "status"           VARCHAR(10)  NOT NULL,
  "executed_queries" INT          NOT NULL DEFAULT 0,
  "error_message"    TEXT,
  "created_at"       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## GORM model example

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:255;not null"`
    Email     string    `gorm:"size:255;uniqueIndex"`
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
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    evo.Setup(pgsql.Driver{})
    schema.UseModels(&User{})
    evo.Run() // pass --migration-do to execute migrations
}
```

Run with migration flag:

```shell
./myapp --migration-do          # apply migrations
./myapp --migration-dry-run     # print SQL without executing
./myapp --migration-dump        # dump CREATE TABLE DDL
```

## Database operations

After `evo.Setup(pgsql.Driver{})` you access the database through the `db` package or `evo.GetDBO()`:

```go
import (
    "github.com/getevo/evo/v2/lib/db"
)

// Insert
user := User{Name: "Alice", Email: "alice@example.com"}
db.Create(&user)

// Query
var users []User
db.Where("name = ?", "Alice").Find(&users)

// Update
db.Model(&user).Update("name", "Alice Smith")

// Delete
db.Delete(&user)

// Raw SQL with PostgreSQL syntax
var count int64
db.Raw(`SELECT COUNT(*) FROM "users" WHERE "email" LIKE ?`, "%@example.com").Scan(&count)
```

## PostgreSQL-specific features

### JSON columns

```go
import "github.com/getevo/evo/v2/lib/db/types"

type Product struct {
    ID       uint            `gorm:"primaryKey"`
    Name     string
    Metadata types.JSON      `gorm:"type:jsonb"`
    Tags     types.JSONSlice `gorm:"type:jsonb"`
}
```

### UUID primary key

```go
import "github.com/getevo/evo/v2/lib/db/types"

type Order struct {
    ID        types.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Reference string
}
```

### Array columns

```go
import "github.com/getevo/evo/v2/lib/db/types"

type Post struct {
    ID   uint
    Tags types.Strings `gorm:"type:text[]"`
}
```

## Health checks

Use `db.Ping` and `db.HealthCheck` to integrate with the EVO health check system:

```go
import (
    "context"
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
)

func main() {
    evo.Setup(pgsql.Driver{})

    // Liveness: fast ping
    evo.OnHealthCheck(func() error {
        return db.Ping(context.Background(), db.GetInstance())
    })

    // Readiness: wait for DB on startup
    evo.OnReadyCheck(func() error {
        return db.WaitForDB(context.Background(), db.GetInstance(), 5, time.Second)
    })

    evo.Run()
}
```

See [health-checks.md](health-checks.md) and [database.md](database.md) for full details.

## Troubleshooting

### Cannot connect

- Verify `Server`, `Username`, `Password`, `Database` values.
- Check `SSLMode` — many local setups need `disable`.
- Confirm the PostgreSQL server is reachable from the application host.

### Schema not found

- The `Schema` field defaults to `public`. If you use a custom schema, create it first: `CREATE SCHEMA tenant_a;`

### Migration lock stuck

If the application crashed mid-migration the advisory lock may remain. Release it manually:

```sql
SELECT pg_advisory_unlock(hashtext('schema_migration_lock'));
```

## See Also

- [MySQL Driver](mysql.md)
- [Database](database.md)
- [Health Checks](health-checks.md)
- [Migration](migration.md)
