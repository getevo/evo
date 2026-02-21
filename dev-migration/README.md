# EVO Migration Development & Testing

This directory contains a comprehensive migration testing suite for the EVO database migration system with support for MySQL, MariaDB, PostgreSQL, and SQLite.

## Features

### 📋 Complete Model Coverage
- **Users**: Primary keys, unique indexes, soft deletes
- **UserProfiles**: Foreign keys, nullable fields, custom types
- **Posts**: Full-text search, enums, composite indexes
- **Comments**: Self-referencing FKs, composite indexes
- **Categories**: Custom table options, hierarchical data
- **Tags**: Many-to-many relationships, simple structures
- **PostTags**: Junction tables, composite primary keys
- **Settings**: JSON fields, unique constraints
- **Logs**: Versioned migrations, audit trails
- **SessionData**: String primary keys, nullable FKs

### 🗄️ Database Support Matrix

| Feature | MySQL | MariaDB | PostgreSQL | SQLite |
|---------|-------|---------|------------|--------|
| Auto Increment | ✅ AUTO_INCREMENT | ✅ AUTO_INCREMENT | ✅ SERIAL/BIGSERIAL | ✅ AUTOINCREMENT |
| JSON Fields | ✅ JSON | ⚠️ LONGTEXT | ✅ JSONB | ⚠️ TEXT |
| ENUM Types | ✅ ENUM | ✅ ENUM | ⚠️ VARCHAR(255) | ⚠️ TEXT |
| Full-Text Search | ✅ FULLTEXT | ✅ FULLTEXT | ⚠️ GIN Indexes | ⚠️ FTS5 |
| Foreign Keys | ✅ Full Support | ✅ Full Support | ✅ Full Support | ⚠️ Limited |
| Unique Indexes | ✅ | ✅ | ✅ | ✅ |
| Composite Indexes | ✅ | ✅ | ✅ | ✅ |
| Table Comments | ✅ | ✅ | ✅ | ❌ |
| Column Comments | ✅ | ✅ | ⚠️ Separate | ❌ |

## Quick Start

### 1. Setup Database

#### SQLite (Default)
```bash
cd dev-migration
go run . -test  # Uses config.sqlite.yml
```

#### MySQL
```bash
# Update config.mysql.yml with your credentials
cp config.mysql.yml config.yml  # Or set ENV vars
go run . -test
```

#### PostgreSQL
```bash
# Update config.pgsql.yml with your credentials
cp config.pgsql.yml config.yml  # Or set ENV vars
go run . -test
```

### 2. Run Migration Tests

```bash
# Run comprehensive test suite
go run . -test

# Or start the server for manual testing
go run .
```

### 3. Expected Output

```
🚀 Starting Migration Test Suite
================================
📊 Testing Database Detection...
✅ Database detected: postgresql
📋 Database info: PostgreSQL 14.2 on database 'testdb'

📦 Running Migration...
✅ Migration completed in 245ms

🔍 Verifying Table Creation...
  ✅ Table 'users' exists
  ✅ Table 'user_profiles' exists
  ✅ Table 'posts' exists
  ✅ Table 'comments' exists
  ✅ Table 'categories' exists
  ✅ Table 'tags' exists
  ✅ Table 'post_tags' exists
  ✅ Table 'settings' exists
  ✅ Table 'logs' exists
  ✅ Table 'session_data' exists

💾 Testing Data Insertion...
  ✅ User created with ID: 1
  ✅ User profile created with ID: 1
  ✅ Category created with ID: 1
  ✅ Post created with ID: 1
  ✅ Comment created with ID: 1

🔗 Testing Foreign Key Constraints...
  ✅ Foreign key constraint properly enforced
  ✅ Valid foreign key accepted

📈 Testing Indexes...
  ✅ Unique index properly enforced on email
  ✅ Unique index properly enforced on username

🎉 All tests passed successfully!
Migration system is working correctly.
```

## Model Examples

### Basic Model with Constraints
```go
type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Email     string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
    Username  string    `gorm:"uniqueIndex:idx_username;size:50;not null" json:"username"`
    FirstName string    `gorm:"size:100" json:"first_name"`
    LastName  string    `gorm:"size:100" json:"last_name"`
    Password  string    `gorm:"size:255;not null" json:"-"`
    IsActive  bool      `gorm:"default:true;not null" json:"is_active"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
```

### Foreign Key Relationships
```go
type UserProfile struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID      uint      `gorm:"not null;FK:users.id" json:"user_id"`
    Bio         string    `gorm:"type:text" json:"bio"`
    DateOfBirth *time.Time `json:"date_of_birth"`
    Country     string    `gorm:"size:100;default:'US'" json:"country"`
}
```

### Full-Text Search & Enums
```go
type Post struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID      uint      `gorm:"not null;index;FK:users.id" json:"user_id"`
    Title       string    `gorm:"size:255;not null;FULLTEXT" json:"title"`
    Content     string    `gorm:"type:text;FULLTEXT" json:"content"`
    Status      string    `gorm:"type:enum('draft','published','archived');default:'draft';not null" json:"status"`
    PublishedAt *time.Time `gorm:"index" json:"published_at"`
}
```

### Composite Indexes
```go
type Comment struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    PostID    uint      `gorm:"not null;index:idx_post_user,priority:1;FK:posts.id" json:"post_id"`
    UserID    uint      `gorm:"not null;index:idx_post_user,priority:2;FK:users.id" json:"user_id"`
    ParentID  *uint     `gorm:"index;FK:comments.id" json:"parent_id"`
    Content   string    `gorm:"type:text;not null" json:"content"`
}
```

### Custom Table Configuration
```go
type Category struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string    `gorm:"size:100;not null;uniqueIndex;CHARSET:utf8mb4;COLLATE:utf8mb4_unicode_ci" json:"name"`
    ParentID    *uint     `gorm:"index;FK:categories.id" json:"parent_id"`
}

func (Category) TableEngine() string {
    return "InnoDB"
}

func (Category) TableCharset() string {
    return "utf8mb4"
}

func (Category) TableCollation() string {
    return "utf8mb4_unicode_ci"
}
```

### Versioned Migrations
```go
type Log struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    *uint     `gorm:"index;FK:users.id" json:"user_id"`
    Action    string    `gorm:"size:50;not null;index" json:"action"`
    Resource  string    `gorm:"size:50;not null;index" json:"resource"`
    Details   string    `gorm:"type:text" json:"details"`
    CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (Log) Migration(currentVersion string) []schema.Migration {
    return []schema.Migration{
        {
            Version: "1.0.0",
            Query:   "ALTER TABLE logs ADD COLUMN session_id VARCHAR(255) AFTER user_id",
        },
        {
            Version: "1.1.0",
            Query:   "ALTER TABLE logs ADD INDEX idx_session_id (session_id)",
        },
        {
            Version: "1.2.0",
            Query:   "ALTER TABLE logs ADD COLUMN metadata JSON AFTER details",
        },
    }
}
```

### Junction Tables (Many-to-Many)
```go
type PostTag struct {
    PostID uint `gorm:"primaryKey;FK:posts.id" json:"post_id"`
    TagID  uint `gorm:"primaryKey;FK:tags.id" json:"tag_id"`
}
```

## Configuration Files

### SQLite (config.sqlite.yml)
```yaml
Database:
  Enabled: true
  Type: sqlite
  Server: "./database.sqlite"
  Debug: 3
```

### MySQL (config.mysql.yml)
```yaml
Database:
  Enabled: true
  Type: mysql
  Server: "localhost:3306"
  Database: "evo_test"
  Username: "root"
  Password: "password"
  Debug: 3
```

### PostgreSQL (config.pgsql.yml)
```yaml
Database:
  Enabled: true
  Type: postgres
  Server: "localhost:5432"
  Database: "evo_test"
  Username: "postgres"
  Password: "password"
  Debug: 3
```

## Testing Different Scenarios

### 1. Fresh Database Migration
```bash
# Delete existing database/tables
rm database.sqlite  # For SQLite
# Or DROP DATABASE for MySQL/PostgreSQL

# Run migration
go run . -test
```

### 2. Schema Updates
```bash
# Modify models in models.go
# Add new fields, indexes, or constraints
# Run migration again
go run . -test
```

### 3. Data Type Changes
```bash
# Change field types in models
# Test type conversion and compatibility
go run . -test
```

### 4. Foreign Key Testing
```bash
# Test constraint enforcement
# Verify cascade operations
go run . -test
```

## Debugging

### Enable SQL Logging
```yaml
Database:
  Debug: 4  # Shows all SQL queries
```

### Manual Database Inspection

#### SQLite
```bash
sqlite3 database.sqlite
.tables
.schema users
```

#### MySQL
```sql
USE evo_test;
SHOW TABLES;
DESCRIBE users;
SHOW CREATE TABLE posts;
```

#### PostgreSQL
```sql
\c evo_test
\dt
\d users
\d+ posts
```

## Known Limitations

### SQLite
- No ENUM support (converted to TEXT)
- No JSON support (stored as TEXT)
- Limited ALTER TABLE support
- No column comments
- Foreign keys need to be enabled with `PRAGMA foreign_keys=ON`

### PostgreSQL
- ENUM types converted to VARCHAR(255)
- Different syntax for indexes and constraints
- No table-level charset/collation
- Different auto-increment mechanism (SERIAL)

### MariaDB
- JSON stored as LONGTEXT
- Some MySQL features may not be available

## Troubleshooting

### Common Issues

1. **Connection Errors**
   - Verify database credentials
   - Ensure database server is running
   - Check network connectivity

2. **Migration Fails**
   - Check database permissions
   - Verify foreign key references exist
   - Review constraint conflicts

3. **Type Conversion Errors**
   - Some types may need manual conversion
   - Check database-specific limitations

4. **Performance Issues**
   - Large tables may slow migration
   - Consider adding indexes after data migration

## Contributing

When adding new models or features:

1. Add the model to `models.go`
2. Register it in the `init()` function
3. Add corresponding tests in `test_migration.go`
4. Update this documentation
5. Test with all supported databases

## Performance Benchmarks

Migration times for the complete test suite:

| Database | Tables | Indexes | Time |
|----------|--------|---------|------|
| SQLite | 10 | 15 | ~100ms |
| MySQL | 10 | 15 | ~200ms |
| PostgreSQL | 10 | 15 | ~250ms |
| MariaDB | 10 | 15 | ~220ms |

*Times may vary based on system performance and database configuration.*