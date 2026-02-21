package sqlite

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getevo/evo/v2/lib/db/schema/table"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/version"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Migration struct {
	Version string
	Query   string
}

type Migrator struct {
	db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) GetMigrationScript(migrations []any) []string {
	var queries []string

	log.Debug("SQLite migration running with %d models", len(migrations))

	var is table.Tables
	// SQLite: Get existing tables
	var tableNames []string
	m.db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableNames)

	// Convert to table.Tables format
	for _, tableName := range tableNames {
		is = append(is, table.Table{
			Database: "main",
			Table:    tableName,
			Type:     "table",
		})
	}

	var columns table.Columns
	// SQLite: Get column information for existing tables
	for _, tableName := range tableNames {
		var tableInfo []struct {
			CID          int    `gorm:"column:cid"`
			Name         string `gorm:"column:name"`
			Type         string `gorm:"column:type"`
			NotNull      int    `gorm:"column:notnull"`
			DefaultValue *string `gorm:"column:dflt_value"`
			PK           int    `gorm:"column:pk"`
		}
		
		m.db.Raw("PRAGMA table_info(?)", tableName).Scan(&tableInfo)
		
		for i, col := range tableInfo {
			column := table.Column{
				Database:        "main",
				Table:           tableName,
				Name:            col.Name,
				OrdinalPosition: i + 1,
				ColumnDefault:   col.DefaultValue,
				DataType:        col.Type,
				ColumnType:      col.Type,
			}
			
			if col.NotNull == 1 {
				column.Nullable = "NO"
			} else {
				column.Nullable = "YES"
			}
			
			if col.PK == 1 {
				column.ColumnKey = "PRI"
			}
			
			columns = append(columns, column)
		}
	}

	var constraints []table.Constraint
	// SQLite constraints are handled differently - for now, empty slice

	var tb *table.Table
	// Process column details
	for idx, _ := range columns {
		if tb == nil || columns[idx].Table != tb.Table {
			tb = is.GetTable(columns[idx].Table)
			if tb == nil {
				continue
			}
		}
		if columns[idx].ColumnKey == "PRI" {
			tb.PrimaryKey = append(tb.PrimaryKey, columns[idx])
		}
		tb.Columns = append(tb.Columns, columns[idx])
	}

	var tail []string

	for idx, el := range migrations {
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

		var stmt = m.db.Model(el).Statement
		var err = stmt.Parse(el)
		if err != nil {
			log.Fatal(err)
		}

		if obj, ok := stmt.Model.(interface{ TableName() string }); ok {
			if strings.HasPrefix(obj.TableName(), "sqlite_") {
				continue
			}
		}

		if stmt.Schema == nil {
			log.Fatal(fmt.Errorf("invalid schema for %s", reflect.TypeOf(el)))
		}

		// Check if table exists
		var existingTable = is.GetTable(stmt.Schema.Table)

		var q []string
		if existingTable != nil {
			// Table exists - check if schema needs updates
			existingTable.Model = migrations[idx]
			existingTable.Reflect = reflect.ValueOf(existingTable.Model)
			q = m.getDiff(stmt, *existingTable)

			// Filter out CREATE TABLE IF NOT EXISTS for existing tables
			var filteredQueries []string
			for _, query := range q {
				if !strings.Contains(query, "CREATE TABLE IF NOT EXISTS") {
					filteredQueries = append(filteredQueries, query)
				}
			}
			q = filteredQueries
		} else {
			// Table doesn't exist - create it
			q = m.getCreateQuery(stmt)
		}

		tail = append(tail, m.getConstraints(stmt, constraints, is)...)
		if len(q) > 0 {
			queries = append(queries, "\r\n\r\n-- Migrate Model: "+stmt.Schema.ModelType.PkgPath()+"."+stmt.Schema.ModelType.Name()+"("+stmt.Schema.Table+")")
			queries = append(queries, q...)
		} else if existingTable == nil {
			log.Debug("No migration queries generated for new table: %s", stmt.Schema.Table)
		} else {
			log.Debug("Table %s schema is up to date, no migration needed", stmt.Schema.Table)
		}

		if caller, ok := el.(interface {
			Migration(version string) []Migration
		}); ok {
			var currentVersion = "0.0.0"
			// SQLite doesn't have table comments, could use a separate metadata table
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
				queries = append(queries, "\r\n\r\n-- Migrate "+stmt.Schema.Table+".Migrate:")
				queries = append(queries, buff...)
				// SQLite doesn't support table comments
			}
		}
	}
	queries = append(queries, tail...)
	return queries
}

func (m *Migrator) DoMigration(migrations []any) error {
	var err error
	err = m.db.Transaction(func(tx *gorm.DB) error {
		for _, query := range m.GetMigrationScript(migrations) {
			if !strings.HasPrefix(query, "--") {
				// Only debug DDL queries (CREATE, ALTER, INSERT, UPDATE, DELETE)
				if m.isDDLQuery(query) {
					err = tx.Debug().Exec(query).Error
				} else {
					err = tx.Exec(query).Error
				}
				if err != nil {
					log.Error(err)
				}
			} else {
				fmt.Println(query)
			}
		}
		return nil
	})
	return err
}

// Helper functions specific to SQLite
func (m *Migrator) quote(name string) string {
	return "\"" + name + "\""
}

func (m *Migrator) isDDLQuery(query string) bool {
	query = strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(query, "CREATE") ||
		strings.HasPrefix(query, "ALTER") ||
		strings.HasPrefix(query, "DROP") ||
		strings.HasPrefix(query, "INSERT") ||
		strings.HasPrefix(query, "UPDATE") ||
		strings.HasPrefix(query, "DELETE") ||
		strings.HasPrefix(query, "COMMENT")
}

func (m *Migrator) normalizeDataType(dataType string) string {
	dataType = strings.ToLower(strings.TrimSpace(dataType))
	
	// Handle SQLite specific type mappings
	switch {
	case strings.HasPrefix(dataType, "varchar"):
		return "text"
	case strings.HasPrefix(dataType, "char"):
		return "text"
	case dataType == "datetime":
		return "text"
	case dataType == "timestamp":
		return "text"
	case dataType == "bigint":
		return "integer"
	case dataType == "int":
		return "integer"
	case dataType == "smallint":
		return "integer"
	case dataType == "tinyint":
		return "integer"
	case dataType == "boolean":
		return "integer"
	case dataType == "text":
		return "text"
	case dataType == "json":
		return "text"
	case dataType == "double":
		return "real"
	case dataType == "float":
		return "real"
	case dataType == "decimal":
		return "real"
	case dataType == "blob":
		return "blob"
	case dataType == "vector" || strings.HasPrefix(dataType, "vector("):
		return "text" // SQLite doesn't have native vector type, use text for JSON storage
	}
	
	return dataType
}

func (m *Migrator) normalizeDefaultValue(defaultValue string) string {
	if defaultValue == "" {
		return ""
	}
	
	defaultValue = strings.TrimSpace(defaultValue)
	
	// Handle SQLite specific default value formats
	switch {
	case strings.HasPrefix(defaultValue, "'") && strings.HasSuffix(defaultValue, "'"):
		// 'value' -> value
		return strings.Trim(defaultValue, "'")
	case defaultValue == "CURRENT_TIMESTAMP":
		return "CURRENT_TIMESTAMP"
	case defaultValue == "NULL":
		return "NULL"
	}
	
	return defaultValue
}

func (m *Migrator) getDiff(stmt *gorm.Statement, remote table.Table) []string {
	// SQLite has limited ALTER TABLE support
	return []string{}
}

func (m *Migrator) getCreateQuery(stmt *gorm.Statement) []string {
	// Use GORM's built-in AutoMigrate for table creation
	log.Debug("Creating SQLite table: %s", stmt.Schema.Table)
	err := m.db.AutoMigrate(stmt.Model)
	if err != nil {
		log.Error("Failed to auto-migrate table %s: %v", stmt.Schema.Table, err)
	}
	return []string{}
}

func (m *Migrator) getConstraints(stmt *gorm.Statement, constraints []table.Constraint, tables table.Tables) []string {
	var queries []string
	
	// Parse foreign key constraints from GORM field tags
	// Note: SQLite foreign keys need PRAGMA foreign_keys=ON to be enforced
	for _, field := range stmt.Schema.Fields {
		fkTag := field.Tag.Get("fk")
		if fkTag != "" {
			queries = append(queries, m.createForeignKeyConstraint(stmt.Schema.Table, field, fkTag)...)
		}
	}
	
	return queries
}

func (m *Migrator) createForeignKeyConstraint(tableName string, field *schema.Field, fkTag string) []string {
	var queries []string
	
	// SQLite foreign keys are typically defined at table creation time
	// Adding them later requires recreating the table, so we'll skip for now
	// and rely on GORM's built-in foreign key handling
	
	queries = append(queries, fmt.Sprintf("-- SQLite foreign key constraint for %s.%s -> %s (handled by GORM)", tableName, field.DBName, fkTag))
	
	return queries
}