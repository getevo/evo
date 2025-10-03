package schema

import (
	"github.com/getevo/evo/v2/lib/db/migration/mysql"
	"github.com/getevo/evo/v2/lib/db/migration/postgres"
	"github.com/getevo/evo/v2/lib/db/migration/sqlite"
	"github.com/getevo/evo/v2/lib/log"
	"gorm.io/gorm"
)

var migrations []any

const null = "NULL"

func GetMigrationScript(db *gorm.DB) []string {
	dialectName := db.Dialector.Name()

	switch dialectName {
	case "postgres":
		migrator := postgres.NewMigrator(db)
		return migrator.GetMigrationScript(migrations)
	case "mysql":
		migrator := mysql.NewMigrator(db)
		return migrator.GetMigrationScript(migrations)
	case "sqlite":
		migrator := sqlite.NewMigrator(db)
		return migrator.GetMigrationScript(migrations)
	default:
		// Default to MySQL for unknown dialects
		log.Warning("Unknown database dialect, defaulting to MySQL migrator", "dialect", dialectName)
		migrator := mysql.NewMigrator(db)
		return migrator.GetMigrationScript(migrations)
	}
}

func DoMigration(db *gorm.DB) error {
	dialectName := db.Dialector.Name()

	log.Info("Starting migration dialect: %s total models: %d", dialectName, len(migrations))
	switch dialectName {
	case "postgres":
		migrator := postgres.NewMigrator(db)
		return migrator.DoMigration(migrations)
	case "mysql":
		migrator := mysql.NewMigrator(db)
		return migrator.DoMigration(migrations)
	case "sqlite":
		migrator := sqlite.NewMigrator(db)
		return migrator.DoMigration(migrations)
	default:
		// Default to MySQL for unknown dialects
		log.Warning("Unknown database dialect, defaulting to MySQL migrator", "dialect", dialectName)
		migrator := mysql.NewMigrator(db)
		return migrator.DoMigration(migrations)
	}
}

type Migration struct {
	Version string
	Query   string
}
