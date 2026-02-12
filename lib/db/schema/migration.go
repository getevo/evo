package schema

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/version"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

var OnBeforeMigration []func(db *gorm.DB)
var OnAfterMigration []func(db *gorm.DB)

var migrations []any

// ResetMigrations clears the registered migrations and models (used for testing).
func ResetMigrations() {
	migrations = nil
	Models = nil
	database = ""
}

const null = "NULL"

func GetMigrationScript(db *gorm.DB) []string {
	var queries []string

	// Initialize dialect if needed
	d := GetDialect()
	if d == nil {
		d = InitDialect(db)
	}

	// Get current database
	var database = d.GetCurrentDatabase(db)

	// Parse all registered models into GORM statements
	var stmts []*gorm.Statement
	var modelSlice []any
	for _, el := range migrations {
		var ref = reflect.ValueOf(el)
		for {
			if ref.Kind() == reflect.Ptr {
				ref = ref.Elem()
			} else {
				break
			}
		}
		if ref.Kind() != reflect.Struct {
			continue
		}

		var stmt = db.Model(el).Statement
		var err = stmt.Parse(el)
		if err != nil {
			log.Fatal(err)
		}

		if obj, ok := stmt.Model.(interface{ TableName() string }); ok {
			if strings.HasPrefix(obj.TableName(), "information_schema.") {
				continue
			}
		}

		if stmt.Schema == nil {
			log.Fatal(fmt.Errorf("invalid schema for %s", reflect.TypeOf(el)))
		}

		stmts = append(stmts, stmt)
		modelSlice = append(modelSlice, el)
	}

	// Delegate all introspection + DDL generation to the dialect
	result := d.GenerateMigration(db, database, stmts, modelSlice)

	queries = append(queries, result.Queries...)

	// Version-based migrations (shared across dialects)
	for idx, el := range modelSlice {
		if caller, ok := el.(interface {
			Migration(version string) []Migration
		}); ok {
			tableName := stmts[idx].Schema.Table
			var currentVersion = "0.0.0"
			if result.TableExists[tableName] {
				currentVersion = d.GetTableVersion(db, database, tableName)
			}
			var buff []string
			var ptr = "0.0.0"
			for _, item := range caller.Migration(currentVersion) {
				if item.Version == "*" || version.Compare(currentVersion, item.Version, "<") {
					if item.Version == "*" {
						ptr = currentVersion
					} else if version.Compare(ptr, item.Version, "<=") {
						ptr = item.Version
					}
					item.Query = strings.TrimSpace(item.Query)
					if !strings.HasSuffix(item.Query, ";") {
						item.Query += ";"
					}
					buff = append(buff, item.Query)
				}
			}

			if len(buff) > 0 {
				queries = append(queries, "\r\n\r\n-- Migrate "+tableName+".Migrate:")
				queries = append(queries, buff...)
				queries = append(queries, d.SetTableVersionSQL(tableName, ptr))
			}
		}
	}

	queries = append(queries, result.Tail...)
	return queries
}

func DoMigration(db *gorm.DB) error {
	var err error
	var migrations = GetMigrationScript(db)
	if len(migrations) == 0 {
		return nil
	}
	for _, fn := range OnBeforeMigration {
		fn(db)
	}
	for _, query := range migrations {
		query = strings.TrimSpace(query)
		if !strings.HasPrefix(strings.TrimSpace(query), "--") && strings.TrimSpace(query) != "" {
			err = db.Debug().Exec(query).Error
			if err != nil {
				log.Error(err)
			}
		} else {
			fmt.Println(query)
		}
	}
	for _, fn := range OnAfterMigration {
		fn(db)
	}

	return err
}

type Migration struct {
	Version string
	Query   string
}
