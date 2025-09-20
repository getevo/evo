package evo

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/settings"
	"log"
	"os"
	"strings"
	"time"

	"github.com/getevo/evo/v2/lib/db/schema"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func setupDatabase() {
	var err error
	var config = DatabaseConfig{}

	settings.Register("Database", &config)
	settings.Get("Database").Cast(&config)
	if !config.Enabled {
		return
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
	switch strings.ToLower(config.Type) {
	case "mysql":
		connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", config.Username, config.Password, config.Server, config.Database, config.Params)
		db, err = gorm.Open(mysql.Open(connectionString), cfg)
	case "postgres", "postgresql", "pgsql":
		connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s %s", config.Server, config.Username, config.Password, config.Database, config.Params)
		db, err = gorm.Open(postgres.Open(connectionString), cfg)
	default:
		db, err = gorm.Open(sqlite.Open(config.Database+config.Params), cfg)
	}
	if err != nil {
		log.Fatal("unable to connect to database", "error", err)
		return
	}

	//switch settings to database
	/*	var driver = database.Database{}
		driver.Init()
		settings.SetInterface(&driver)*/

}

// GetDBO return database object instance
func GetDBO() *gorm.DB {
	if db == nil {
		setupDatabase()
	}
	return db
}

type Model struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

func DoMigration() error {
	return schema.DoMigration(db)
}

func Models() []schema.Model {
	return schema.Models
}

func GetModel(name string) *schema.Model {
	return schema.Find(name)
}
