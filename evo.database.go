package evo

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	dbpkg "github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/db/schema"
	evolog "github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/settings"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func setupDatabase() error {
	var err error
	var config = DatabaseConfig{}

	settings.Register("Database", &config)
	settings.Get("Database").Cast(&config)

	if !config.Enabled {
		return nil
	}
	var logLevel logger.LogLevel

	switch config.Debug {
	case 4:
		logLevel = logger.Info
	case 3:
		logLevel = logger.Warn
	case 2:
		logLevel = logger.Error
	default:
		logLevel = logger.Silent
	}

	var newLog = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: config.SlowQueryThreshold, // Slow SQL threshold
			LogLevel:      logLevel,                  // Log level
			Colorful:      true,                      // Disable color
		},
	)
	cfg := &gorm.Config{
		Logger: newLog,
	}

	driver := dbpkg.GetDriver()
	if driver == nil {
		return fmt.Errorf("no database driver registered")
	}
	driverCfg := dbpkg.DriverConfig{
		Server:   config.Server,
		Username: config.Username,
		Password: config.Password,
		Database: config.Database,
		Schema:   config.Schema,
		SSLMode:  config.SSLMode,
		Params:   config.Params,
	}
	db, err = driver.Open(driverCfg, cfg)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	return nil
}

// GetDBO return database object instance
func GetDBO(ctx ...context.Context) *gorm.DB {
	if db == nil {
		if err := setupDatabase(); err != nil {
			evolog.Fatal("failed to setup database", "error", err)
		}
	}
	if len(ctx) > 0 {
		return db.WithContext(ctx[0])
	}
	return db
}

// GetDB returns a database instance with context for proper context propagation.
// This is the preferred method for obtaining database connections.
func GetDB(ctx ...context.Context) *gorm.DB {
	return GetDBO(ctx...)
}

type Model struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

func DoMigration() error {
	driver := dbpkg.GetDriver()
	if driver == nil {
		return fmt.Errorf("no database driver registered")
	}
	for _, fn := range schema.OnBeforeMigration {
		fn(db)
	}
	queries := driver.GetMigrationScript(db)
	var err error
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" || strings.HasPrefix(query, "--") {
			if query != "" {
				fmt.Println(query)
			}
			continue
		}
		if e := db.Exec(query).Error; e != nil {
			evolog.Error(e)
			err = e
		}
	}
	for _, fn := range schema.OnAfterMigration {
		fn(db)
	}
	return err
}

func Models() []schema.Model {
	return schema.Models
}

func GetModel(name string) *schema.Model {
	return schema.Find(name)
}
