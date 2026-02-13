package db

import (
	"gorm.io/gorm"
)

// Driver abstracts database connection and migration logic so that each driver
// (mysql, pgsql) is a self-contained package. Users pass the desired driver to evo.Setup().
type Driver interface {
	Name() string
	Open(config DriverConfig, gormConfig *gorm.Config) (*gorm.DB, error)
	GetMigrationScript(db *gorm.DB) []string
}

// DriverConfig holds the connection parameters extracted from DatabaseConfig.
type DriverConfig struct {
	Server   string
	Username string
	Password string
	Database string
	Schema   string // PostgreSQL schema name (defaults to 'public')
	SSLMode  string
	Params   string
}

var registeredDriver Driver

// RegisterDriver sets the active database driver.
func RegisterDriver(d Driver) { registeredDriver = d }

// GetDriver returns the currently registered driver, or nil.
func GetDriver() Driver { return registeredDriver }
