package schema

import (
	"fmt"
	"strings"
	"gorm.io/gorm"
	
	"github.com/getevo/evo/v2/lib/db/schema/ddl"
	"github.com/getevo/evo/v2/lib/log"
)

// DatabaseType represents the type of database being used
type DatabaseType string

const (
	DatabaseMySQL      DatabaseType = "mysql"
	DatabaseMariaDB    DatabaseType = "mariadb"
	DatabasePostgreSQL DatabaseType = "postgresql"
	DatabaseSQLite     DatabaseType = "sqlite"
	DatabaseUnknown    DatabaseType = "unknown"
)

// DatabaseInfo contains information about the detected database
type DatabaseInfo struct {
	Type     DatabaseType
	Version  string
	Database string
	Schema   string
}

// MigrationEngine handles database migrations with multi-database support
type MigrationEngine struct {
	db          *gorm.DB
	dbInfo      *DatabaseInfo
	models      []any
}

// GetDatabaseType returns the detected database type
func (me *MigrationEngine) GetDatabaseType() string {
	return string(me.dbInfo.Type)
}

// GetDatabaseInfo returns formatted database information
func (me *MigrationEngine) GetDatabaseInfo() string {
	return fmt.Sprintf("%s %s on database '%s'", me.dbInfo.Type, me.dbInfo.Version, me.dbInfo.Database)
}

// NewMigrationEngine creates a new migration engine with database detection
func NewMigrationEngine(database *gorm.DB) (*MigrationEngine, error) {
	dbInfo, err := detectDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("failed to detect database type: %w", err)
	}
	
	return &MigrationEngine{
		db:           database,
		dbInfo:       dbInfo,
		models:       migrations,
	}, nil
}

// detectDatabase determines the database type and version from a GORM DB connection
func detectDatabase(db *gorm.DB) (*DatabaseInfo, error) {
	info := &DatabaseInfo{}
	
	// Get the dialect name first
	dialectName := db.Dialector.Name()
	
	switch dialectName {
	case "mysql":
		return detectMySQL(db)
	case "postgres":
		return detectPostgreSQL(db)
	case "sqlite":
		return detectSQLite(db)
	default:
		info.Type = DatabaseUnknown
		return info, nil
	}
}

// detectMySQL detects MySQL or MariaDB and gets database info
func detectMySQL(db *gorm.DB) (*DatabaseInfo, error) {
	info := &DatabaseInfo{}
	
	// Get version to distinguish MySQL from MariaDB
	var version string
	err := db.Raw("SELECT VERSION()").Scan(&version).Error
	if err != nil {
		return nil, err
	}
	
	info.Version = version
	if strings.Contains(strings.ToLower(version), "mariadb") {
		info.Type = DatabaseMariaDB
	} else {
		info.Type = DatabaseMySQL
	}
	
	// Get current database name
	err = db.Raw("SELECT DATABASE()").Scan(&info.Database).Error
	if err != nil {
		return nil, err
	}
	
	info.Schema = info.Database
	return info, nil
}

// detectPostgreSQL detects PostgreSQL and gets database info
func detectPostgreSQL(db *gorm.DB) (*DatabaseInfo, error) {
	info := &DatabaseInfo{Type: DatabasePostgreSQL}
	
	// Get PostgreSQL version
	var version string
	err := db.Raw("SELECT version()").Scan(&version).Error
	if err != nil {
		return nil, err
	}
	info.Version = version
	
	// Get current database name
	err = db.Raw("SELECT current_database()").Scan(&info.Database).Error
	if err != nil {
		return nil, err
	}
	
	// Get current schema
	err = db.Raw("SELECT current_schema()").Scan(&info.Schema).Error
	if err != nil {
		return nil, err
	}
	
	return info, nil
}

// detectSQLite detects SQLite and gets basic info
func detectSQLite(db *gorm.DB) (*DatabaseInfo, error) {
	info := &DatabaseInfo{Type: DatabaseSQLite}
	
	// Get SQLite version
	var version string
	err := db.Raw("SELECT sqlite_version()").Scan(&version).Error
	if err != nil {
		return nil, err
	}
	info.Version = version
	
	// SQLite doesn't have database/schema concept like MySQL/PostgreSQL
	info.Database = "main"
	info.Schema = "main"
	
	return info, nil
}

// GetMigrationScript generates migration SQL for all registered models
func (me *MigrationEngine) GetMigrationScript() ([]string, error) {
	// Set the DDL engine based on detected database
	me.setDDLEngine()
	
	// Use the original GetMigrationScript with database detection enhancements
	return GetMigrationScript(me.db), nil
}

// DoMigration executes the migration with proper error handling
func (me *MigrationEngine) DoMigration() error {
	queries, err := me.GetMigrationScript()
	if err != nil {
		return fmt.Errorf("failed to generate migration script: %w", err)
	}
	
	if len(queries) == 0 {
		log.Info("No migrations needed")
		return nil
	}
	
	// Execute migrations in a transaction with improved error handling
	return me.db.Transaction(func(tx *gorm.DB) error {
		var errors []string
		
		for _, query := range queries {
			query = strings.TrimSpace(query)
			if query == "" || strings.HasPrefix(query, "--") {
				if strings.HasPrefix(query, "--") {
					log.Info(query)
				}
				continue
			}
			
			log.Debug("Executing migration query:", query)
			if err := tx.Exec(query).Error; err != nil {
				errorMsg := fmt.Sprintf("Migration failed for query '%s': %v", query, err)
				log.Error(errorMsg)
				errors = append(errors, errorMsg)
				
				// For critical errors, stop migration
				if me.isCriticalError(err) {
					return fmt.Errorf("critical migration error: %w", err)
				}
			}
		}
		
		if len(errors) > 0 {
			return fmt.Errorf("migration completed with %d errors: %s", len(errors), strings.Join(errors, "; "))
		}
		
		log.Info("Migration completed successfully")
		return nil
	})
}


// setDDLEngine configures the DDL engine based on detected database
func (me *MigrationEngine) setDDLEngine() {
	switch me.dbInfo.Type {
	case DatabaseMySQL:
		ddl.Engine = "mysql"
	case DatabaseMariaDB:
		ddl.Engine = "mariadb"
	case DatabasePostgreSQL:
		ddl.Engine = "postgresql"
	case DatabaseSQLite:
		ddl.Engine = "sqlite"
	default:
		ddl.Engine = "mysql" // fallback
	}
}

// isCriticalError determines if an error should stop the migration
func (me *MigrationEngine) isCriticalError(err error) bool {
	errStr := strings.ToLower(err.Error())
	
	// Critical errors that should stop migration
	criticalErrors := []string{
		"syntax error",
		"table doesn't exist",
		"column doesn't exist",
		"duplicate column",
		"duplicate key",
		"foreign key constraint fails",
		"data too long",
		"out of range",
	}
	
	for _, critical := range criticalErrors {
		if strings.Contains(errStr, critical) {
			return true
		}
	}
	
	return false
}

// DoMigrationV2 is the new entry point for migrations that replaces DoMigration
func DoMigrationV2(db *gorm.DB) error {
	engine, err := NewMigrationEngine(db)
	if err != nil {
		return err
	}
	
	return engine.DoMigration()
}