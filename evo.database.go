package evo

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

var Database *gorm.DB

func setupDatabase() {
	Events.Go("database.starts")
	config := config.Database
	var err error
	if config.Enabled == false {
		return
	}
	var logLevel = logger.Silent
	config.Debug = strings.ToLower(config.Debug)

	switch config.Debug {
	case "true", "all", "*", "any":
		logLevel = logger.Info
	case "warn", "warning":
		logLevel = logger.Warn
	case "err", "error":
		logLevel = logger.Error
	default:
		logLevel = logger.Silent
	}

	if config.Debug == "true" || config.Debug == "all" {
		logLevel = logger.Info
	}
	if config.Debug == "warn" || config.Debug == "warning" {
		logLevel = logger.Warn
	}
	if config.Debug == "err" || config.Debug == "error" {
		logLevel = logger.Error
	}
	var newLog = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logLevel,    // Log level
			Colorful:      true,        // Disable color
		},
	)
	cfg := &gorm.Config{
		Logger: newLog,
	}
	switch strings.ToLower(config.Type) {
	case "mysql":
		connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", config.Username, config.Password, config.Server, config.Database, config.Params)
		Database, err = gorm.Open(mysql.Open(connectionString), cfg)

		/*	case "postgres":
			connectionString := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s "+config.Params, config.Username, config.Password, config.Server, config.Database, config.SSLMode)
				Database, err = gorm.Open(postgres.Open(connectionString), cfg)*/
	case "mssql":
		connectionString := fmt.Sprintf("user id=%s;password=%s;server=%s;database:%s;"+config.Params, config.Username, config.Password, config.Server, config.Database)
		Database, err = gorm.Open(sqlserver.Open(connectionString), cfg)
	default:
		Database, err = gorm.Open(sqlite.Open(config.Database+config.Params), cfg)
	}

	if err != nil {
		panic(err)
		return
	}

	Events.Go("database.started")

}

// GetDBO return database object instance
func GetDBO() *gorm.DB {
	if Database == nil {
		setupDatabase()
	}
	return Database
}

type Model struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}
