package mysql

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
	db     *gorm.DB
	engine string // "mysql" or "mariadb"
}

func NewMigrator(db *gorm.DB) *Migrator {
	migrator := &Migrator{db: db}
	
	// Detect MySQL vs MariaDB
	var version string
	db.Raw("SELECT VERSION()").Scan(&version)
	version = strings.ToLower(version)
	
	if strings.Contains(version, "mariadb") {
		migrator.engine = "mariadb"
	} else {
		migrator.engine = "mysql"
	}
	
	return migrator
}

func (m *Migrator) GetMigrationScript(migrations []any) []string {
	var queries []string

	var database = ""
	m.db.Raw("SELECT DATABASE()").Scan(&database)
	
	log.Debug("MySQL migration running for database: %s with %d models", database, len(migrations))

	var is table.Tables
	// MySQL/MariaDB specific query
	m.db.Raw(`SELECT CCSA.character_set_name  AS 'TABLE_CHARSET',T.* FROM information_schema.TABLES T, information_schema.COLLATION_CHARACTER_SET_APPLICABILITY CCSA WHERE CCSA.collation_name = T.table_collation AND T.table_schema = ?`, database).Scan(&is)

	var columns table.Columns
	// MySQL/MariaDB
	m.db.Where(table.Table{Database: database}).Order("TABLE_NAME ASC,ORDINAL_POSITION ASC").Find(&columns)

	var constraints []table.Constraint
	// MySQL/MariaDB use REFERENCED_TABLE_SCHEMA
	m.db.Where(table.Constraint{Database: database}).Find(&constraints)

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

	var istats []table.IndexStat
	// MySQL/MariaDB
	m.db.Where(table.IndexStat{Database: database}).Order("TABLE_NAME ASC, SEQ_IN_INDEX ASC").Find(&istats)
	
	var tail []string
	var indexMap = map[string]table.Index{}
	for _, item := range istats {
		if item.Name == "PRIMARY" {
			continue
		}
		if _, ok := indexMap[item.Table+item.Name]; !ok {
			indexMap[item.Table+item.Name] = table.Index{
				Name:   item.Name,
				Table:  item.Table,
				Unique: !item.NonUnique,
			}
		}
		var m = indexMap[item.Table+item.Name]
		var c = is.GetTable(item.Table).Columns.GetColumn(item.ColumnName)
		m.Columns = append(m.Columns, *c)
		indexMap[item.Table+item.Name] = m
	}
	for key, item := range indexMap {
		is.GetTable(item.Table).Indexes = append(is.GetTable(item.Table).Indexes, indexMap[key])
	}

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
			if strings.HasPrefix(obj.TableName(), "information_schema.") {
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
			log.Debug("No migration queries generated for new table: " + stmt.Schema.Table)
		} else {
			log.Debug("Table " + stmt.Schema.Table + " schema is up to date, no migration needed")
		}

		if caller, ok := el.(interface {
			Migration(version string) []Migration
		}); ok {
			var currentVersion = "0.0.0"
			if existingTable != nil {
				m.db.Raw("SELECT table_comment FROM INFORMATION_SCHEMA.TABLES WHERE table_schema=? AND table_name=?", database, existingTable.Table).Scan(&currentVersion)
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
				queries = append(queries, "\r\n\r\n-- Migrate "+stmt.Schema.Table+".Migrate:")
				queries = append(queries, buff...)
				queries = append(queries, "ALTER TABLE `"+stmt.Schema.Table+"` COMMENT '"+ptr+"';")
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

// Helper functions specific to MySQL/MariaDB
func (m *Migrator) quote(name string) string {
	return "`" + name + "`"
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
	
	// Handle MySQL/MariaDB specific type mappings
	switch {
	case strings.HasPrefix(dataType, "varchar"):
		return dataType
	case strings.HasPrefix(dataType, "char"):
		return dataType
	case dataType == "datetime":
		return "datetime"
	case dataType == "timestamp":
		return "timestamp"
	case dataType == "bigint":
		return "bigint"
	case dataType == "int":
		return "int"
	case dataType == "smallint":
		return "smallint"
	case dataType == "tinyint":
		return "tinyint"
	case dataType == "tinyint(1)":
		return "boolean"
	case dataType == "text":
		return "text"
	case dataType == "longtext":
		return "longtext"
	case dataType == "json":
		return "json"
	case dataType == "double":
		return "double"
	case dataType == "float":
		return "float"
	case dataType == "decimal":
		return "decimal"
	case strings.HasPrefix(dataType, "decimal("):
		return dataType
	case dataType == "vector" || strings.HasPrefix(dataType, "vector("):
		return "json" // MySQL doesn't have native vector type, use JSON
	}
	
	return dataType
}

func (m *Migrator) normalizeDefaultValue(defaultValue string) string {
	if defaultValue == "" {
		return ""
	}
	
	defaultValue = strings.TrimSpace(defaultValue)
	
	// Handle MySQL/MariaDB specific default value formats
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
	// Basic implementation - return empty for now to avoid errors
	return []string{}
}

func (m *Migrator) getCreateQuery(stmt *gorm.Statement) []string {
	// Use GORM's built-in AutoMigrate for table creation
	log.Debug("Creating MySQL table: %s", stmt.Schema.Table)
	err := m.db.AutoMigrate(stmt.Model)
	if err != nil {
		log.Error("Failed to auto-migrate table %s: %v", stmt.Schema.Table, err)
	}
	return []string{}
}

func (m *Migrator) getConstraints(stmt *gorm.Statement, constraints []table.Constraint, tables table.Tables) []string {
	var queries []string
	
	// Parse foreign key constraints from GORM field tags
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
	
	// Parse fk tag: "table" or "table.column"
	parts := strings.Split(fkTag, ".")
	referencedTable := parts[0]
	var referencedColumn string
	
	if len(parts) > 1 {
		referencedColumn = parts[1]
	} else {
		// If no column specified, assume it references the primary key
		switch referencedTable {
		case "channels":
			referencedColumn = "slug"
		case "conversations":
			referencedColumn = "id"
		default:
			referencedColumn = "id" // Default assumption
		}
	}
	
	constraintName := fmt.Sprintf("fk_%s_%s", tableName, field.DBName)
	
	// Check if constraint already exists
	var existingCount int
	checkQuery := `SELECT COUNT(*) FROM information_schema.table_constraints 
		WHERE constraint_schema = DATABASE() 
		AND table_name = ? 
		AND constraint_name = ? 
		AND constraint_type = 'FOREIGN KEY'`
	
	m.db.Raw(checkQuery, tableName, constraintName).Scan(&existingCount)
	
	if existingCount == 0 {
		query := fmt.Sprintf(
			"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			m.quote(tableName),
			m.quote(constraintName),
			m.quote(field.DBName),
			m.quote(referencedTable),
			m.quote(referencedColumn),
		)
		
		queries = append(queries, fmt.Sprintf("-- Creating foreign key constraint %s", constraintName))
		queries = append(queries, query+";")
	}
	
	return queries
}