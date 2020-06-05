package evo

import (
	"fmt"
	"github.com/getevo/evo/lib/log"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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
	switch strings.ToLower(config.Type) {
	case "mysql":
		connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?"+config.Params, config.Username, config.Password, config.Server, config.Database)
		Database, err = gorm.Open("mysql", connectionString)
	case "postgres":
		connectionString := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s "+config.Params, config.Username, config.Password, config.Server, config.Database, config.SSLMode)
		Database, err = gorm.Open("postgres", connectionString)
	case "mssql":
		connectionString := fmt.Sprintf("user id=%s;password=%s;server=%s;database:%s;"+config.Params, config.Username, config.Password, config.Server, config.Database)
		Database, err = gorm.Open("mssql", connectionString)
	default:
		Database, err = gorm.Open("sqlite3", config.Database+config.Params)
	}
	Database.LogMode(config.Debug == "true")
	if err != nil {
		log.Critical(err)
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
