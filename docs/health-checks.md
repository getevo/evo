# Health Check API

EVO provides built-in health check and readiness check endpoints for monitoring application health and Kubernetes integration.

## Overview

The health check API exposes two endpoints:

- **`GET /health`** - Liveness probe - checks if the application is alive and running
- **`GET /ready`** - Readiness probe - checks if the application is ready to serve traffic

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
)

type App struct{}

func (App) Name() string { return "myapp" }

func (App) Register() error {
    // Register health checks
    evo.OnHealthCheck(func() error {
        // Check database connection
        if db.IsEnabled() {
            sqlDB, _ := db.DB().DB()
            if err := sqlDB.Ping(); err != nil {
                return fmt.Errorf("database unhealthy: %w", err)
            }
        }
        return nil
    })

    // Register readiness checks
    evo.OnReadyCheck(func() error {
        // Check if migrations are complete
        // Check if cache is warmed up
        // Check if external services are reachable
        return nil
    })

    return nil
}

func (App) Router() error { return nil }

func main() {
    evo.Setup()
    evo.Register(&App{})
    evo.Run()
}
```

## Health vs Readiness

### Health Check (Liveness)
- **Purpose**: Determine if the application is alive
- **Use case**: Kubernetes liveness probe
- **Action on failure**: Restart the container
- **Examples**:
  - Basic process check (always returns nil)
  - Database connection ping
  - Memory/resource checks
  - Critical service availability

### Readiness Check
- **Purpose**: Determine if the application is ready to serve traffic
- **Use case**: Kubernetes readiness probe, load balancer health checks
- **Action on failure**: Remove from service temporarily
- **Examples**:
  - Database migrations complete
  - Cache warmed up
  - External dependencies available
  - Configuration loaded

## API Reference

### `evo.OnHealthCheck(fn HealthCheckFunc)`

Registers a health check function.

```go
type HealthCheckFunc func() error

evo.OnHealthCheck(func() error {
    // Return nil if healthy
    // Return error if unhealthy
    if isHealthy() {
        return nil
    }
    return fmt.Errorf("unhealthy: reason")
})
```

**Thread-safe**: Multiple goroutines can register checks concurrently.

### `evo.OnReadyCheck(fn HealthCheckFunc)`

Registers a readiness check function.

```go
evo.OnReadyCheck(func() error {
    // Return nil if ready
    // Return error if not ready
    if isReady() {
        return nil
    }
    return fmt.Errorf("not ready: reason")
})
```

**Thread-safe**: Multiple goroutines can register checks concurrently.

## HTTP Endpoints

### GET /health

Returns the health status of the application.

**Success Response (200 OK):**
```json
{
  "status": "ok"
}
```

**Failure Response (503 Service Unavailable):**
```json
{
  "status": "unhealthy",
  "error": "database ping failed: connection refused"
}
```

### GET /ready

Returns the readiness status of the application.

**Success Response (200 OK):**
```json
{
  "status": "ok"
}
```

**Failure Response (503 Service Unavailable):**
```json
{
  "status": "not ready",
  "error": "cache not initialized"
}
```

## Examples

### Database Health Check

```go
evo.OnHealthCheck(func() error {
    if !db.IsEnabled() {
        return nil
    }

    sqlDB, err := db.DB().DB()
    if err != nil {
        return fmt.Errorf("database unavailable: %w", err)
    }

    if err := sqlDB.Ping(); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }

    return nil
})
```

### External Service Check

```go
evo.OnReadyCheck(func() error {
    client := http.Client{Timeout: 2 * time.Second}
    resp, err := client.Get("https://api.example.com/health")
    if err != nil {
        return fmt.Errorf("external service unreachable: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("external service unhealthy: status %d", resp.StatusCode)
    }

    return nil
})
```

### Memory Usage Check

```go
evo.OnHealthCheck(func() error {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    // Fail if using more than 1GB
    maxMemory := uint64(1024 * 1024 * 1024)
    if m.Alloc > maxMemory {
        return fmt.Errorf("memory usage too high: %d MB", m.Alloc/1024/1024)
    }

    return nil
})
```

### Cache Warmup Check

```go
evo.OnReadyCheck(func() error {
    if !cache.IsInitialized() {
        return fmt.Errorf("cache not initialized")
    }

    if !cache.IsWarmedUp() {
        return fmt.Errorf("cache warmup in progress")
    }

    return nil
})
```

## Kubernetes Integration

### Deployment YAML Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080

        # Liveness probe - restart if unhealthy
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        # Readiness probe - remove from service if not ready
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
```

### Probe Configuration

**Liveness Probe:**
- `initialDelaySeconds`: Wait before first check (app startup time)
- `periodSeconds`: How often to check
- `failureThreshold`: Failures before restart (usually 3)

**Readiness Probe:**
- `initialDelaySeconds`: Wait before first check (shorter than liveness)
- `periodSeconds`: How often to check (more frequent)
- `failureThreshold`: Failures before removing from service

## Best Practices

### 1. Keep Checks Fast
```go
// Good - fast check
evo.OnHealthCheck(func() error {
    return db.Ping()
})

// Bad - slow check
evo.OnHealthCheck(func() error {
    return runComplexQuery() // Avoid!
})
```

### 2. Don't Duplicate Checks
```go
// Register once during app initialization
func (App) Register() error {
    evo.OnHealthCheck(dbHealthCheck)
    return nil
}

// Not in every request!
```

### 3. Fail Fast
```go
evo.OnReadyCheck(func() error {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    return checkServiceWithContext(ctx)
})
```

### 4. Meaningful Error Messages
```go
// Good
return fmt.Errorf("redis connection failed: %w", err)

// Bad
return errors.New("error")
```

### 5. Separate Health and Ready
```go
// Health - critical services only
evo.OnHealthCheck(func() error {
    return db.Ping() // App can't work without this
})

// Ready - nice-to-have services
evo.OnReadyCheck(func() error {
    return cache.Ping() // App can work, but slower
})
```

## Monitoring

### Prometheus Metrics (Future Enhancement)

```go
// TODO: Add metrics integration
// health_check_total{status="success|failure",check="db"}
// health_check_duration_seconds{check="db"}
```

### Logging

Health check failures are automatically logged:
```
[ERROR] Health check failed: database ping failed: connection refused
```

## Testing

```go
func TestHealthChecks(t *testing.T) {
    // Reset hooks
    healthCheckHooks = nil

    // Register test check
    evo.OnHealthCheck(func() error {
        return nil
    })

    // Verify registration
    if len(healthCheckHooks) != 1 {
        t.Fatal("health check not registered")
    }
}
```

## FAQ

**Q: Can I register checks after evo.Run()?**
A: Yes, but it's recommended to register during app initialization for clarity.

**Q: What happens if no checks are registered?**
A: Both endpoints return 200 OK with `{"status":"ok"}`.

**Q: Can I use async checks?**
A: Yes, but ensure they complete quickly (< 1 second recommended).

**Q: How do I disable health checks?**
A: Don't register any checks. The endpoints will still exist but always return OK.

**Q: Can I customize the response format?**
A: Not currently. Future versions may support custom response formats.

## See Also

- [Kubernetes Liveness/Readiness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Health Check API Pattern](https://microservices.io/patterns/observability/health-check-api.html)
