package mysql

import (
	"fmt"

	dbpkg "github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/db/schema"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Driver implements db.Driver for MySQL/MariaDB.
type Driver struct{}

func (Driver) Name() string { return "mysql" }

func (Driver) Open(config dbpkg.DriverConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	schema.RegisterDialect("mysql", &MySQLDialect{})
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		config.Username, config.Password, config.Server, config.Database, config.Params)
	return gorm.Open(mysql.Open(dsn), gormConfig)
}

func (Driver) GetMigrationScript(db *gorm.DB) []string {
	return schema.GetMigrationScript(db)
}

// RegisterDialect registers the MySQL dialect without opening a connection.
// Use this in standalone test scripts that connect directly via gorm.Open.
func RegisterDialect() {
	schema.RegisterDialect("mysql", &MySQLDialect{})
}
