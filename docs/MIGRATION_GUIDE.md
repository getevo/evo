# Migration Guide - Error Handling Changes

This guide helps you migrate your application from the old EVO framework API (with `log.Fatal()`) to the new error-handling API.

## Overview

The EVO framework has been updated to return errors instead of calling `log.Fatal()`, enabling graceful shutdown and proper error handling.

## Who Needs to Migrate?

**All applications using EVO v2** need to update their `main.go` or initialization code.

## Breaking Changes

### 1. Setup() Function Signature Changed

**Before**:
```go
func Setup(params ...any)
```

**After**:
```go
func Setup(params ...any) error
```

### 2. Run() Function Signature Changed

**Before**:
```go
func Run()
```

**After**:
```go
func Run() error
```

## Step-by-Step Migration

### Step 1: Update main.go

#### Option A: Simple Error Handling (Recommended for most apps)

**Before**:
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

**After**:
```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/log"
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    if err := evo.Setup(pgsql.Driver{}); err != nil {
        log.Fatal("failed to setup application", "error", err)
    }

    if err := evo.Run(); err != nil {
        log.Fatal("failed to run server", "error", err)
    }
}
```

#### Option B: Graceful Shutdown with Cleanup

For production applications that need to clean up resources:

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/log"
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    // Setup application
    if err := evo.Setup(pgsql.Driver{}); err != nil {
        log.Fatal("failed to setup application", "error", err)
    }

    // Setup graceful shutdown
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

    // Run server in goroutine
    go func() {
        if err := evo.Run(); err != nil {
            log.Error("server error", "error", err)
            done <- syscall.SIGTERM
        }
    }()

    // Wait for shutdown signal
    <-done
    log.Info("shutting down gracefully...")

    // Cleanup resources
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Add your cleanup code here
    // - Close database connections
    // - Flush logs
    // - Save state
    // - etc.

    log.Info("shutdown complete")
}
```

### Step 2: Update Custom Initialization Functions

If you have custom initialization that calls EVO functions:

**Before**:
```go
func initializeApp() {
    // Configure settings
    settings.Set("HTTP.Port", "8080")

    // Setup EVO
    evo.Setup(pgsql.Driver{})
}
```

**After**:
```go
func initializeApp() error {
    // Configure settings
    settings.Set("HTTP.Port", "8080")

    // Setup EVO
    if err := evo.Setup(pgsql.Driver{}); err != nil {
        return fmt.Errorf("failed to setup EVO: %w", err)
    }

    return nil
}
```

### Step 3: Update Tests

Test files that use EVO Setup/Run also need updates:

**Before**:
```go
func TestMyApp(t *testing.T) {
    evo.Setup(mysql.Driver{})
    // ... test code
}
```

**After**:
```go
func TestMyApp(t *testing.T) {
    if err := evo.Setup(mysql.Driver{}); err != nil {
        t.Fatalf("setup failed: %v", err)
    }
    // ... test code
}
```

## Common Errors After Migration

### Error: "undefined: error in function signature"

**Cause**: You're still using the old function signature.

**Fix**: Update function calls to handle the returned error:
```go
// Wrong
evo.Setup(driver)

// Correct
if err := evo.Setup(driver); err != nil {
    log.Fatal("setup failed", "error", err)
}
```

### Error: "not enough return values"

**Cause**: Custom functions that wrap Setup/Run need to return errors.

**Fix**: Add error return type:
```go
// Wrong
func mySetup() {
    evo.Setup(driver)
}

// Correct
func mySetup() error {
    return evo.Setup(driver)
}
```

## Benefits of the New API

### 1. Graceful Shutdown

**Before**: Application terminates immediately on errors, potentially corrupting data.

**After**: Application can clean up resources before exiting:
- Close database connections properly
- Flush pending logs
- Save application state
- Notify other services

### 2. Better Error Context

**Before**:
```
FATAL: unable to connect to database
```

**After**:
```
ERROR: failed to setup application: failed to initialize settings: config file not found: /etc/app/config.yml
```

Error messages now include the full error chain.

### 3. Testable Initialization

**Before**: Tests would call `os.Exit()`, making them impossible to run.

**After**: Tests can verify error conditions:
```go
func TestInvalidConfig(t *testing.T) {
    err := evo.Setup(driver)
    if err == nil {
        t.Fatal("expected error with invalid config")
    }
    if !strings.Contains(err.Error(), "config file not found") {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### 4. Production-Ready

The new API follows Go best practices:
- Errors are values
- Panics are reserved for truly exceptional cases
- Graceful degradation is possible

## Deprecation Timeline

| Version | Status | Notes |
|---------|--------|-------|
| v2.0.x | Old API | log.Fatal() in Setup/Run |
| v2.1.0 | **Current** | New error-returning API |
| v2.2.0 | Future | Old API removed (planned) |

**Recommendation**: Migrate as soon as possible. The old behavior will be removed in v2.2.0.

## Getting Help

If you encounter issues during migration:

1. **Check the error message**: The new API provides detailed error context
2. **Review examples**: See `examples/` directory for updated examples
3. **Open an issue**: https://github.com/getevo/evo/issues
4. **Join discussions**: https://github.com/getevo/evo/discussions

## Example Applications

Full example applications with the new API:

- `examples/basic-app/` - Simple REST API
- `examples/graceful-shutdown/` - Production-ready shutdown
- `examples/testing/` - How to test with new API

## Checklist

Use this checklist to verify your migration:

- [ ] Updated `main.go` to handle `Setup()` error
- [ ] Updated `main.go` to handle `Run()` error
- [ ] Updated all test files
- [ ] Updated custom initialization functions
- [ ] Tested application startup
- [ ] Tested graceful shutdown (send SIGTERM)
- [ ] Tested error conditions (invalid config, DB connection failure)
- [ ] Updated CI/CD scripts if needed
- [ ] Updated documentation
- [ ] Informed team members of changes

## Questions?

Contact: https://github.com/getevo/evo/issues
