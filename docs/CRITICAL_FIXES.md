# Critical Fixes - February 2026

This document details the critical fixes applied to the EVO framework to improve error handling, code maintainability, and configurability.

## 🚨 Critical Fixes Applied

### 1. Graceful Shutdown - Removed log.Fatal() Calls

**Problem**: The framework used `log.Fatal()` in 8+ locations during initialization, which prevented graceful shutdown and caused resource leaks.

**Files Changed**:
- `evo.go`
- `evo.database.go`
- `lib/mysql/dialect.go`
- `lib/pgsql/dialect.go`

**Changes**:

#### evo.go
- `Setup()` now returns `error` instead of calling `log.Fatal()`
- `Run()` now returns `error` instead of calling `log.Fatal()`

**Before**:
```go
func Setup(params ...any) {
    var err = settings.Init()
    if err != nil {
        log.Fatal(err)  // Terminates immediately!
    }
    // ...
}
```

**After**:
```go
func Setup(params ...any) error {
    var err = settings.Init()
    if err != nil {
        return fmt.Errorf("failed to initialize settings: %w", err)
    }
    // ...
    return nil
}
```

#### evo.database.go
- `setupDatabase()` now returns `error`
- Errors are propagated up instead of terminating

**Migration Required**: Applications must update their main.go to handle errors:

**Before**:
```go
func main() {
    evo.Setup(pgsql.Driver{})
    evo.Run()
}
```

**After**:
```go
func main() {
    if err := evo.Setup(pgsql.Driver{}); err != nil {
        log.Fatal("failed to setup: ", err)
    }
    if err := evo.Run(); err != nil {
        log.Fatal("failed to run server: ", err)
    }
}
```

**Benefits**:
- ✅ Graceful shutdown with proper cleanup
- ✅ Resource leaks prevented (DB connections, file handles closed)
- ✅ Better error context and stack traces
- ✅ Testable initialization code

---

### 2. Proper Error Handling in Database Dialects

**Problem**: Database introspection queries silently ignored errors, leading to incorrect migration behavior.

**Files Changed**:
- `lib/mysql/dialect.go`
- `lib/pgsql/dialect.go`

**Changes**:

All database query operations now check for errors:

**Before**:
```go
func (p *PGDialect) GetCurrentDatabase(db *gorm.DB) string {
    var database string
    db.Raw("SELECT current_database()").Scan(&database)  // Error ignored!
    return database
}
```

**After**:
```go
func (p *PGDialect) GetCurrentDatabase(db *gorm.DB) string {
    var database string
    if err := db.Raw("SELECT current_database()").Scan(&database).Error; err != nil {
        log.Error("failed to get current database", "error", err)
        return ""
    }
    return database
}
```

**Benefits**:
- ✅ Errors are logged with context
- ✅ Migration failures are visible
- ✅ Debugging is easier
- ✅ Silent failures eliminated

---

### 3. PostgreSQL Dialect Refactored for Maintainability

**Problem**: Single `dialect.go` file had 1,236 lines, violating the 300 lines/file guideline and making it unmaintainable.

**Files Created**:
```
lib/pgsql/
├── dialect.go          (107 lines) - Main interface, locks, bootstrap
├── types.go            (118 lines) - Type definitions
├── introspection.go    (177 lines) - Database introspection
├── migration.go        ( 64 lines) - Migration orchestration
├── ddl.go              (600 lines) - DDL generation
├── constraints.go      ( 66 lines) - Constraint handling
└── helpers.go          (165 lines) - Helper utilities
```

**Benefits**:
- ✅ Clear separation of concerns
- ✅ Easier to navigate and understand
- ✅ Simpler code reviews
- ✅ Reduced merge conflicts
- ✅ Follows Go best practices

**File Responsibilities**:

| File | Purpose |
|------|---------|
| `dialect.go` | PGDialect struct, interface methods, locks |
| `types.go` | All struct definitions (pgRemoteTable, etc.) |
| `introspection.go` | Database schema discovery |
| `migration.go` | Main generateMigration function |
| `ddl.go` | CREATE/ALTER TABLE generation |
| `constraints.go` | Foreign key constraints |
| `helpers.go` | Utility functions (type normalization, etc.) |

---

### 4. Configurable Schema Names (PostgreSQL)

**Problem**: PostgreSQL schema was hardcoded to 'public', breaking multi-schema setups.

**Files Changed**:
- `settings.go` - Added `Schema` field
- `lib/db/driver.go` - Added `Schema` to DriverConfig
- `evo.database.go` - Pass schema to driver
- `lib/pgsql/dialect.go` - Use configurable schema
- `lib/pgsql/driver.go` - Initialize dialect with schema
- `lib/pgsql/introspection.go` - Parameterized all schema references

**Configuration**:

Add to your `config.yml`:

```yaml
DATABASE:
  TYPE: postgres
  SERVER: localhost:5432
  USERNAME: myuser
  PASSWORD: mypass
  DATABASE: mydb
  SCHEMA: public  # NEW: Configure PostgreSQL schema (defaults to 'public')
```

**Code Changes**:

The `PGDialect` now stores and uses the configured schema:

```go
// Create dialect with schema
dialect := NewPGDialect(config.Schema)

// Use in queries
func (p *PGDialect) Schema() string {
    if p.schema == "" {
        return "public"
    }
    return p.schema
}
```

All hardcoded `'public'` references replaced with `p.Schema()`:
- 8 SQL queries updated in introspection.go
- 2 SQL queries updated in dialect.go

**Benefits**:
- ✅ Multi-schema PostgreSQL support
- ✅ Backwards compatible (defaults to 'public')
- ✅ Explicit configuration
- ✅ No code changes required for existing setups

---

## Impact Summary

| Fix | Impact | Breaking Change |
|-----|--------|-----------------|
| log.Fatal() removal | **High** - Enables graceful shutdown | ⚠️ Yes - main.go requires updates |
| Error handling in dialects | **Medium** - Better debugging | ✅ No |
| PostgreSQL dialect split | **Low** - Internal refactoring | ✅ No |
| Configurable schema | **Low** - New feature | ✅ No |

## Testing Checklist

After applying these fixes:

- [ ] Update main.go to handle Setup() and Run() errors
- [ ] Test database connection failures (should not panic)
- [ ] Test migration failures (should not panic)
- [ ] Verify graceful shutdown (Ctrl+C)
- [ ] Test with non-default PostgreSQL schema (if using PostgreSQL)
- [ ] Run existing integration tests
- [ ] Check logs for previously silent errors

## Rollback Instructions

If you need to rollback:

1. Check out the commit before these changes:
   ```bash
   git log --oneline --graph  # Find the commit hash
   git checkout <commit-before-fixes>
   ```

2. Note: Rollback is **not recommended** as it reintroduces critical bugs.

## Questions & Support

For issues or questions:
- Open an issue: https://github.com/getevo/evo/issues
- Read the migration guide: `docs/MIGRATION_GUIDE.md`
- Check configuration docs: `docs/CONFIGURATION.md`
