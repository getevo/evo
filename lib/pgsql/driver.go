package pgsql

import (
	"fmt"
	"strings"

	dbpkg "github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/db/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Driver implements db.Driver for PostgreSQL.
type Driver struct{}

func (Driver) Name() string { return "postgres" }

func (Driver) Open(config dbpkg.DriverConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	schema.RegisterDialect("postgres", NewPGDialect(config.Schema))
	pgHost := config.Server
	pgPort := "5432"
	if idx := strings.LastIndex(config.Server, ":"); idx > 0 {
		pgHost = config.Server[:idx]
		pgPort = config.Server[idx+1:]
	}
	sslMode := "disable"
	if config.SSLMode == "true" || config.SSLMode == "require" || config.SSLMode == "1" {
		sslMode = "require"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s %s",
		pgHost, pgPort, config.Username, config.Password, config.Database, sslMode, config.Params)
	return gorm.Open(postgres.Open(dsn), gormConfig)
}

func (Driver) GetMigrationScript(db *gorm.DB) []string {
	return schema.GetMigrationScript(db)
}

// RegisterDialect registers the PostgreSQL dialect without opening a connection.
// Use this in standalone test scripts that connect directly via gorm.Open.
func RegisterDialect() {
	schema.RegisterDialect("postgres", &PGDialect{})
}
