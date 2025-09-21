# EVO Database Migration System

## Overview

The EVO database migration system provides automatic schema management with support for multiple database engines including MySQL, MariaDB, PostgreSQL, and SQLite. The system automatically detects your database type and generates appropriate SQL for schema creation and migration.

## Supported Databases

- **MySQL** (5.7+)
- **MariaDB** (10.3+) 
- **PostgreSQL** (12+)
- **SQLite** (3.35+)

## Basic Usage

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db/schema"
)

type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Email     string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
    Username  string    `gorm:"size:50;not null" json:"username"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
when
func init() {
    evo.SetupHook(func() {
        schema.UseModel(evo.GetDBO(), User{})
    })
}
```

## GORM Struct Tags Reference

### Primary Key & Auto Increment
```go
type Model struct {
    ID uint `gorm:"primaryKey;autoIncrement"`
}
```

**Database Support:**
- **MySQL/MariaDB**: `AUTO_INCREMENT`
- **PostgreSQL**: `SERIAL` or `BIGSERIAL`
- **SQLite**: `AUTOINCREMENT`

### Column Types

#### String Fields
```go
type Model struct {
    Name        string `gorm:"size:100"`                    // VARCHAR(100)
    Description string `gorm:"type:text"`                   // TEXT
    Status      string `gorm:"type:enum('active','inactive')"` // ENUM (MySQL), VARCHAR (PostgreSQL/SQLite)
}
```

#### Numeric Fields
```go
type Model struct {
    Count    int     `gorm:""`                  // INT/INTEGER
    BigNum   int64   `gorm:""`                  // BIGINT
    Price    float64 `gorm:"type:decimal(10,2)"` // DECIMAL(10,2)
    IsActive bool    `gorm:""`                  // TINYINT(1)/BOOLEAN/INTEGER
}
```

#### Date/Time Fields
```go
type Model struct {
    CreatedAt time.Time  `gorm:"autoCreateTime"`      // Auto-set on create
    UpdatedAt time.Time  `gorm:"autoUpdateTime"`      // Auto-update on save
    DeletedAt *time.Time `gorm:"index"`               // Soft delete support
    PublishedAt *time.Time `gorm:""`                   // Nullable timestamp
}
```

#### JSON Fields
```go
type Model struct {
    Metadata string `gorm:"type:json"` // JSON/JSONB/TEXT depending on database
}
```

**Database Mapping:**
- **MySQL**: `JSON`
- **MariaDB**: `LONGTEXT`
- **PostgreSQL**: `JSONB`
- **SQLite**: `TEXT`

### Constraints

#### NOT NULL
```go
type Model struct {
    Email string `gorm:"not null"`        // NOT NULL
    Phone string `gorm:"NULLABLE"`        // Allow NULL (overrides pointer inference)
}
```

#### Default Values
```go
type Model struct {
    Status   string `gorm:"default:'active'"`           // Default string
    Count    int    `gorm:"default:0"`                  // Default number
    IsActive bool   `gorm:"default:true"`               // Default boolean
    Created  time.Time `gorm:"default:CURRENT_TIMESTAMP"` // Function default
}
```

#### Unique Constraints
```go
type Model struct {
    Email    string `gorm:"uniqueIndex"`                    // Single unique index
    Username string `gorm:"uniqueIndex:idx_user"`          // Named unique index
    Combo1   string `gorm:"uniqueIndex:idx_combo"`         // Composite unique
    Combo2   string `gorm:"uniqueIndex:idx_combo"`         // index (same name)
}
```

### Indexes

#### Single Column Indexes
```go
type Model struct {
    UserID   uint   `gorm:"index"`                  // Simple index
    Category string `gorm:"index:idx_category"`     // Named index
    Status   string `gorm:"index:,priority:1"`      // Index with priority
}
```

#### Composite Indexes
```go
type Model struct {
    PostID uint `gorm:"index:idx_post_user,priority:1"`   // First column
    UserID uint `gorm:"index:idx_post_user,priority:2"`   // Second column
}
```

#### Full-Text Search
```go
type Model struct {
    Title   string `gorm:"FULLTEXT"`                // Single column full-text
    Content string `gorm:"FULLTEXT"`                // Another full-text column
}
```

**Database Support:**
- **MySQL/MariaDB**: Native `FULLTEXT` indexes
- **PostgreSQL**: Converted to GIN indexes on text search vectors
- **SQLite**: Uses FTS5 virtual tables (limited support)

### Foreign Keys

#### Basic Foreign Keys
```go
type Profile struct {
    UserID uint `gorm:"FK:users.id"`              // References users table
    User   User `gorm:"foreignKey:UserID"`        // GORM relationship
}
```

#### Advanced Foreign Key Options
```go
type Model struct {
    CategoryID *uint     `gorm:"FK:categories.id"` // Nullable foreign key
    ParentID   *uint     `gorm:"FK:models.id"`     // Self-referencing
}
```

**Cascade Options** (automatically applied):
- **ON DELETE CASCADE**: When referenced record is deleted
- **ON UPDATE CASCADE**: When referenced key is updated

### Custom Table Configuration

#### Table-Level Settings
```go
type Category struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:100"`
}

// Custom table engine (MySQL/MariaDB only)
func (Category) TableEngine() string {
    return "InnoDB"
}

// Custom charset (MySQL/MariaDB only)
func (Category) TableCharset() string {
    return "utf8mb4"
}

// Custom collation (MySQL/MariaDB only)
func (Category) TableCollation() string {
    return "utf8mb4_unicode_ci"
}
```

#### Custom Table Name
```go
func (Category) TableName() string {
    return "product_categories"
}
```

### Column-Level Customization

#### Custom Column Definition
```go
type CustomField struct {
    Value string
}

func (cf *CustomField) ColumnDefinition(column *ddl.Column) {
    // Custom column definition logic
    if column.Name == "value" {
        column.Type = "VARCHAR(500)"
        column.Default = "'default_value'"
    }
}
```

#### Character Set and Collation (MySQL/MariaDB)
```go
type Model struct {
    Name string `gorm:"CHARSET:utf8mb4;COLLATE:utf8mb4_unicode_ci"`
}
```

### Versioned Migrations

Implement versioned migrations for complex schema changes:

```go
type LogEntry struct {
    ID     uint   `gorm:"primaryKey"`
    Action string `gorm:"size:50"`
    // ... other fields
}

func (LogEntry) Migration(currentVersion string) []schema.Migration {
    return []schema.Migration{
        {
            Version: "1.0.0",
            Query:   "ALTER TABLE log_entries ADD COLUMN session_id VARCHAR(255)",
        },
        {
            Version: "1.1.0", 
            Query:   "CREATE INDEX idx_session_id ON log_entries(session_id)",
        },
        {
            Version: "1.2.0",
            Query:   "ALTER TABLE log_entries ADD COLUMN metadata JSON",
        },
    }
}
```

**Version Comparison:**
- Uses semantic versioning (semver)
- Only runs migrations newer than current table version
- Version `"*"` runs always (use carefully)

## Database-Specific Considerations

### MySQL/MariaDB
- Full support for all features
- InnoDB engine recommended
- UTF8MB4 charset recommended for full Unicode support

### PostgreSQL
- ENUM types converted to VARCHAR(255)
- FULLTEXT converted to GIN indexes
- Auto-increment uses SERIAL/BIGSERIAL
- No table-level charset/collation

### SQLite
- Limited constraint support
- All text types become TEXT
- All numeric types become INTEGER or REAL
- No ENUM or JSON native support
- Foreign keys require PRAGMA foreign_keys=ON

## Migration Process

1. **Detection**: Automatically detects database type and version
2. **Schema Analysis**: Compares model definitions with existing schema
3. **Query Generation**: Creates database-specific SQL
4. **Transaction Execution**: Runs all changes in a transaction
5. **Error Handling**: Rolls back on critical errors

### Migration Commands

```go
// Run migrations
err := evo.DoMigration()

// Or use the new v2 engine directly
err := schema.DoMigrationV2(evo.GetDBO())
```

## Advanced Features

### Skip Migration
```go
type InternalModel struct {
    //model:skip
    ID uint `gorm:"primaryKey"`
}
```

### Junction Tables (Many-to-Many)
```go
type PostTag struct {
    PostID uint `gorm:"primaryKey;FK:posts.id"`
    TagID  uint `gorm:"primaryKey;FK:tags.id"`
}
```

### Soft Deletes
```go
type Model struct {
    DeletedAt *time.Time `gorm:"index"` // Enables soft delete
}
```

## Best Practices

1. **Always use transactions** for migrations
2. **Test migrations** on development data first
3. **Backup production** before running migrations
4. **Use semantic versioning** for versioned migrations
5. **Keep models simple** - complex logic belongs in services
6. **Index foreign keys** for performance
7. **Use appropriate field sizes** to optimize storage

## Troubleshooting

### Common Issues

1. **Foreign key constraints fail**
   - Ensure referenced table exists first
   - Check that foreign key column types match

2. **Migration hangs**
   - Large tables may need manual migration
   - Consider adding indexes after data migration

3. **Charset issues** 
   - Use UTF8MB4 for full Unicode support
   - Ensure database, table, and column charsets match

4. **Type conversion errors**
   - Check database-specific type mappings
   - Some types may need manual conversion

### Debug Mode

Enable debug mode to see generated SQL:

```yaml
Database:
  Debug: 4  # Shows all SQL queries
```

## Examples

See the `dev-migration` directory for comprehensive examples of all supported features and migration patterns.