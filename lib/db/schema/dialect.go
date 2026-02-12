package schema

import (
	"gorm.io/gorm"
)

// Dialect abstracts all database-specific behavior for migration and DDL generation.
// Each dialect is fully self-contained with its own internal types.
type Dialect interface {
	// Name returns the dialect identifier ("mysql" or "postgres").
	Name() string

	// Quote wraps a SQL identifier in dialect-specific quotes.
	// MySQL: `name`   PostgreSQL: "name"
	Quote(name string) string

	// GetCurrentDatabase returns the current database/schema name.
	GetCurrentDatabase(db *gorm.DB) string

	// GenerateMigration performs full introspection + DDL generation for all given models.
	// Each dialect owns its own internal types for introspection.
	GenerateMigration(db *gorm.DB, database string, stmts []*gorm.Statement, models []any) MigrationResult

	// GetTableVersion retrieves the version string stored in table metadata (comment).
	GetTableVersion(db *gorm.DB, database, tableName string) string

	// SetTableVersionSQL returns the SQL statement to set a table's version string.
	SetTableVersionSQL(tableName, version string) string

	// GetJoinConstraints returns foreign key relationships for building joins.
	GetJoinConstraints(db *gorm.DB, database string) []JoinConstraint

	// AcquireMigrationLock obtains an advisory lock to prevent concurrent migrations.
	AcquireMigrationLock(db *gorm.DB) error

	// ReleaseMigrationLock releases the advisory migration lock.
	ReleaseMigrationLock(db *gorm.DB)

	// BootstrapHistoryTable creates the schema_migration table if it does not exist.
	BootstrapHistoryTable(db *gorm.DB) error
}

// currentDialect holds the active dialect implementation.
var currentDialect Dialect

// InitDialect detects the database dialect from the GORM DB instance and initializes it.
func InitDialect(db *gorm.DB) Dialect {
	if currentDialect != nil {
		return currentDialect
	}
	// Auto-detection from dialector name.
	name := db.Dialector.Name()
	if d, ok := dialectRegistry[name]; ok {
		currentDialect = d
		return currentDialect
	}
	return currentDialect
}

// GetDialect returns the current dialect, or nil if not yet initialized.
func GetDialect() Dialect {
	return currentDialect
}

// SetDialect explicitly sets the current dialect.
func SetDialect(d Dialect) {
	currentDialect = d
}

// dialectRegistry holds registered dialect implementations.
var dialectRegistry = map[string]Dialect{}

// RegisterDialect registers a dialect implementation by name.
// Called by dialect packages in their init() functions.
func RegisterDialect(name string, d Dialect) {
	dialectRegistry[name] = d
}
