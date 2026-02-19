package mysql

import (
	"fmt"
	"strings"

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
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return db, err
	}
	// Detect MariaDB at connection time so isMariaDB() works without requiring migration.
	// The GORM MySQL driver already queries SELECT VERSION() internally; we replicate it
	// here to set the shared config used by types.JSON.GormValue and other helpers.
	var ver string
	db.Raw("SELECT VERSION()").Scan(&ver)
	if strings.Contains(strings.ToLower(ver), "mariadb") {
		schema.SetConfig("mysql_engine", "mariadb")
	} else {
		schema.SetConfig("mysql_engine", "mysql")
	}
	return db, nil
}

func (Driver) GetMigrationScript(db *gorm.DB) []string {
	return schema.GetMigrationScript(db)
}

// RegisterDialect registers the MySQL dialect without opening a connection.
// Use this in standalone test scripts that connect directly via gorm.Open.
func RegisterDialect() {
	schema.RegisterDialect("mysql", &MySQLDialect{})
}
