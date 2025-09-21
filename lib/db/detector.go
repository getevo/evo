package db

import (
	"strings"
	"gorm.io/gorm"
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

// DetectDatabase determines the database type and version from a GORM DB connection
func DetectDatabase(db *gorm.DB) (*DatabaseInfo, error) {
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

// IsMySQL returns true if the database is MySQL or MariaDB
func (info *DatabaseInfo) IsMySQL() bool {
	return info.Type == DatabaseMySQL || info.Type == DatabaseMariaDB
}

// IsPostgreSQL returns true if the database is PostgreSQL
func (info *DatabaseInfo) IsPostgreSQL() bool {
	return info.Type == DatabasePostgreSQL
}

// IsSQLite returns true if the database is SQLite
func (info *DatabaseInfo) IsSQLite() bool {
	return info.Type == DatabaseSQLite
}

// SupportsEnums returns true if the database supports ENUM types
func (info *DatabaseInfo) SupportsEnums() bool {
	return info.IsMySQL() || info.IsPostgreSQL()
}

// SupportsFullText returns true if the database supports full-text search
func (info *DatabaseInfo) SupportsFullText() bool {
	return info.IsMySQL() || info.IsPostgreSQL()
}

// SupportsJSON returns true if the database supports JSON column types
func (info *DatabaseInfo) SupportsJSON() bool {
	return info.IsMySQL() || info.IsPostgreSQL()
}

// GetQuoteChar returns the character used for quoting identifiers
func (info *DatabaseInfo) GetQuoteChar() string {
	switch info.Type {
	case DatabaseMySQL, DatabaseMariaDB:
		return "`"
	case DatabasePostgreSQL:
		return "\""
	case DatabaseSQLite:
		return "`"
	default:
		return "`"
	}
}