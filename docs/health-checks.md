# Health Checks

EVO provides built-in health check and readiness endpoints for monitoring and Kubernetes integration, plus a database health library for detailed connection diagnostics.

## Overview

Two HTTP endpoints are registered automatically when `evo.Run()` is called:

| Endpoint | Purpose | Kubernetes use |
|---|---|---|
| `GET /health` | Liveness probe — is the app alive? | Restart container on failure |
| `GET /ready` | Readiness probe — is the app ready for traffic? | Remove from load balancer on failure |

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    evo.Setup(pgsql.Driver{})

    // Liveness: check database ping
    evo.OnHealthCheck(func() error {
        return db.Ping(context.Background(), db.GetInstance())
    })

    // Readiness: wait for DB, check migrations done
    evo.OnReadyCheck(func() error {
        return db.WaitForDB(context.Background(), db.GetInstance(), 3, time.Second)
    })

    evo.Run()
}
```

## Application health check API

### `evo.OnHealthCheck(fn HealthCheckFunc)`

Registers a liveness check. Return `nil` for healthy, a non-nil error if unhealthy.

```go
type HealthCheckFunc func() error

evo.OnHealthCheck(func() error {
    if isHealthy() {
        return nil
    }
    return fmt.Errorf("unhealthy: reason")
})
```

Multiple checks can be registered — all must pass for the endpoint to return 200.

**Thread-safe**: safe to call from multiple goroutines.

### `evo.OnReadyCheck(fn HealthCheckFunc)`

Registers a readiness check. Return `nil` if ready, error if not.

```go
evo.OnReadyCheck(func() error {
    if migrationsComplete && cacheWarmedUp {
        return nil
    }
    return fmt.Errorf("not ready yet")
})
```

## HTTP endpoints

### `GET /health`

**Success (200 OK):**
```json
{"status": "ok"}
```

**Failure (503 Service Unavailable):**
```json
{"status": "unhealthy", "error": "database ping failed: connection refused"}
```

### `GET /ready`

**Success (200 OK):**
```json
{"status": "ok"}
```

**Failure (503 Service Unavailable):**
```json
{"status": "not ready", "error": "cache not initialized"}
```

## Database health library

`lib/db` provides functions for detailed database health checking.

### `db.Ping(ctx, db) error`

Fast liveness check — sends a ping to the database.

```go
import (
    "context"
    "github.com/getevo/evo/v2/lib/db"
)

err := db.Ping(context.Background(), db.GetInstance())
if err != nil {
    log.Error("database unreachable", "error", err)
}
```

### `db.HealthCheck(ctx, db) HealthCheckResult`

Comprehensive health check with connection pool statistics.

```go
type HealthCheckResult struct {
    Healthy          bool          `json:"healthy"`
    ResponseTime     time.Duration `json:"response_time"`
    ConnectionsOpen  int           `json:"connections_open"`
    ConnectionsInUse int           `json:"connections_in_use"`
    ConnectionsIdle  int           `json:"connections_idle"`
    MaxOpenConns     int           `json:"max_open_conns"`
    Error            string        `json:"error,omitempty"`
}
```

```go
result := db.HealthCheck(context.Background(), db.GetInstance())
if !result.Healthy {
    log.Error("db unhealthy", "error", result.Error)
    return
}
log.Info("db healthy",
    "response_time", result.ResponseTime,
    "open_conns",    result.ConnectionsOpen,
    "in_use",        result.ConnectionsInUse,
)
```

### `db.WaitForDB(ctx, db, maxRetries, retryInterval) error`

Blocks until the database is reachable or retries are exhausted. Useful during application startup when the database may not be ready immediately (e.g., Docker Compose startup order).

```go
// Retry 10 times with 2-second intervals
err := db.WaitForDB(
    context.Background(),
    db.GetInstance(),
    10,
    2*time.Second,
)
if err != nil {
    log.Fatal("database never became available", "error", err)
}
```

With context timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := db.WaitForDB(ctx, db.GetInstance(), 15, 2*time.Second); err != nil {
    log.Fatal("database not ready after 30s")
}
```

### `db.GetConnectionStats(db) (sql.DBStats, error)`

Returns raw Go `sql.DBStats` for the connection pool.

```go
stats, err := db.GetConnectionStats(db.GetInstance())
if err != nil {
    return err
}
fmt.Printf("open=%d in_use=%d idle=%d wait=%d\n",
    stats.OpenConnections,
    stats.InUse,
    stats.Idle,
    stats.WaitCount,
)
```

### `db.CloseConnection(db) error`

Gracefully closes the database connection pool.

```go
if err := db.CloseConnection(db.GetInstance()); err != nil {
    log.Error("error closing db", "error", err)
}
```

## Complete example with all checks

```go
package main

import (
    "context"
    "fmt"
    "runtime"
    "time"

    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    evo.Setup(pgsql.Driver{})

    // --- Liveness checks ---

    // 1. Database ping (fast)
    evo.OnHealthCheck(func() error {
        ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()
        return db.Ping(ctx, db.GetInstance())
    })

    // 2. Memory check
    evo.OnHealthCheck(func() error {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        maxBytes := uint64(2 * 1024 * 1024 * 1024) // 2 GB
        if m.Alloc > maxBytes {
            return fmt.Errorf("memory usage critical: %d MB", m.Alloc/1024/1024)
        }
        return nil
    })

    // --- Readiness checks ---

    // 1. Wait for database on startup
    evo.OnReadyCheck(func() error {
        return db.WaitForDB(context.Background(), db.GetInstance(), 5, time.Second)
    })

    // 2. Check external API
    evo.OnReadyCheck(func() error {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        return checkExternalService(ctx)
    })

    evo.Run()
}

func checkExternalService(ctx context.Context) error {
    // implement your check
    return nil
}
```

## Expose health result as JSON endpoint

```go
evo.Get("/health/detail", func(r *evo.Request) any {
    result := db.HealthCheck(context.Background(), db.GetInstance())
    return outcome.OK(result)
})
```

Response:
```json
{
    "healthy": true,
    "response_time": 1200000,
    "connections_open": 5,
    "connections_in_use": 2,
    "connections_idle": 3,
    "max_open_conns": 100
}
```

## Health vs Readiness

| | Liveness (`/health`) | Readiness (`/ready`) |
|---|---|---|
| **Purpose** | App is alive | App can serve traffic |
| **Failure action** | Restart container | Remove from load balancer |
| **What to check** | DB ping, memory | Migrations done, cache ready, external APIs |
| **Check speed** | Very fast (< 100ms) | Can be slightly slower |

## Kubernetes integration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080

        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
```

## Best practices

### Keep checks fast

```go
// Good: fast ping
evo.OnHealthCheck(func() error {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()
    return db.Ping(ctx, db.GetInstance())
})

// Bad: slow query in health check
evo.OnHealthCheck(func() error {
    return runComplexQuery() // avoid!
})
```

### Use context timeouts

```go
evo.OnReadyCheck(func() error {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    return checkExternalService(ctx)
})
```

### Write meaningful error messages

```go
// Good
return fmt.Errorf("redis connection failed: host=%s err=%w", redisHost, err)

// Bad
return errors.New("error")
```

### Separate concerns

```go
// Health — must work for app to function at all
evo.OnHealthCheck(func() error {
    return db.Ping(ctx, db.GetInstance()) // app can't work without this
})

// Ready — nice-to-have, app might degrade without it
evo.OnReadyCheck(func() error {
    return redis.Ping(ctx).Err() // app works, but slower
})
```

## FAQ

**Q: What if no checks are registered?**
Both endpoints return `200 OK` with `{"status":"ok"}`.

**Q: Can I register checks after `evo.Run()`?**
Yes, the hooks slice is protected by a mutex and can be appended to at any time.

**Q: Can checks be async?**
The check functions are called synchronously. Use a context with timeout to prevent hanging.

**Q: How do I disable the endpoints?**
You can't disable them, but if no checks are registered they always return OK.

## See Also

- [Database](database.md)
- [PostgreSQL Driver](pgsql.md)
- [MySQL Driver](mysql.md)
- [Kubernetes Liveness/Readiness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
