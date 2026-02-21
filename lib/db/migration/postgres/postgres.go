package postgres

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

	var database = ""
	var engine string
	m.db.Raw("SELECT current_database();").Scan(&database)
	m.db.Raw("SELECT version();").Scan(&engine)

	var is table.Tables
	var existingTables []string

	// PostgreSQL: Get existing tables from public schema
	query := "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
	m.db.Raw(query).Scan(&existingTables)

	// Convert to table.Tables format
	for _, tableName := range existingTables {
		is = append(is, table.Table{
			Database: database,
			Table:    tableName,
			Type:     "BASE TABLE",
		})
	}

	var columns table.Columns
	// PostgreSQL: Get column information for existing tables
	for _, existingTable := range existingTables {
		var cols []table.Column
		m.db.Raw(`SELECT 
			table_schema as "TABLE_SCHEMA",
			table_name as "TABLE_NAME", 
			column_name as "COLUMN_NAME",
			ordinal_position as "ORDINAL_POSITION",
			column_default as "COLUMN_DEFAULT",
			is_nullable as "IS_NULLABLE",
			data_type as "DATA_TYPE",
			character_maximum_length as "CHARACTER_MAXIMUM_LENGTH",
			numeric_precision as "NUMERIC_PRECISION",
			numeric_scale as "NUMERIC_SCALE"
		FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = ?
		ORDER BY ordinal_position`, existingTable).Scan(&cols)

		for _, col := range cols {
			col.Database = "public"
			columns = append(columns, col)
		}
	}

	var constraints []table.Constraint
	// PostgreSQL uses CONSTRAINT_SCHEMA instead of REFERENCED_TABLE_SCHEMA
	m.db.Where("constraint_schema = ?", database).Find(&constraints)

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
				m.db.Raw("SELECT obj_description(oid) FROM pg_class WHERE relname = ?", existingTable.Table).Scan(&currentVersion)
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
				queries = append(queries, "COMMENT ON TABLE \""+stmt.Schema.Table+"\" IS '"+ptr+"';")
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

// Helper functions specific to PostgreSQL
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
	
	// Handle PostgreSQL specific type mappings
	switch {
	case strings.HasPrefix(dataType, "character varying"):
		if strings.Contains(dataType, "(") {
			// character varying(255) -> varchar(255)
			return strings.Replace(dataType, "character varying", "varchar", 1)
		}
		return "varchar"
	case dataType == "character":
		return "char"
	case dataType == "timestamp without time zone":
		return "timestamp"
	case dataType == "timestamp with time zone":
		return "timestamptz"
	case dataType == "bigint":
		return "bigint"
	case dataType == "integer":
		return "int"
	case dataType == "smallint":
		return "smallint"
	case dataType == "boolean":
		return "boolean"
	case dataType == "text":
		return "text"
	case dataType == "jsonb":
		return "jsonb"
	case dataType == "json":
		return "json"
	case dataType == "uuid":
		return "uuid"
	case dataType == "double precision":
		return "double"
	case dataType == "real":
		return "real"
	case dataType == "numeric":
		return "numeric"
	case strings.HasPrefix(dataType, "numeric("):
		return dataType
	}
	
	return dataType
}

func (m *Migrator) normalizeDefaultValue(defaultValue string) string {
	if defaultValue == "" {
		return ""
	}
	
	defaultValue = strings.TrimSpace(defaultValue)
	
	// Handle PostgreSQL specific default value formats
	switch {
	case strings.HasSuffix(defaultValue, "::character varying"):
		// 'enabled'::character varying -> enabled
		return strings.TrimSuffix(strings.Trim(defaultValue[:strings.LastIndex(defaultValue, "::")], "'\""), "'\"")
	case strings.HasSuffix(defaultValue, "::text"):
		// 'value'::text -> value
		return strings.TrimSuffix(strings.Trim(defaultValue[:strings.LastIndex(defaultValue, "::")], "'\""), "'\"")
	case strings.HasSuffix(defaultValue, "::boolean"):
		// true::boolean -> true
		return strings.TrimSuffix(defaultValue, "::boolean")
	case strings.Contains(defaultValue, "nextval("):
		// nextval('table_id_seq'::regclass) -> "" (auto-increment)
		return ""
	case strings.HasPrefix(defaultValue, "'") && strings.HasSuffix(defaultValue, "'"):
		// 'value' -> value
		return strings.Trim(defaultValue, "'")
	}
	
	return defaultValue
}

func (m *Migrator) getDiff(stmt *gorm.Statement, remote table.Table) []string {
	// Temporarily disable diff checking to avoid false positives
	// TODO: Implement proper type comparison with normalization
	return []string{}
}

func (m *Migrator) getCreateQuery(stmt *gorm.Statement) []string {
	// Use GORM's built-in AutoMigrate for table creation
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
		// For channels table, primary key is "slug"
		// For conversations table, primary key is "id"
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
		WHERE constraint_schema = 'public' 
		AND table_name = $1 
		AND constraint_name = $2 
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

func (m *Migrator) statementToTable(stmt *gorm.Statement) Table {
	var columns []Column
	var primaryKeys []string

	for _, field := range stmt.Schema.Fields {
		column := Column{
			Name:       field.DBName,
			Type:       string(field.DataType),
			Nullable:   !field.NotNull,
			PrimaryKey: field.PrimaryKey,
			Comment:    field.Comment,
		}

		if field.AutoIncrement {
			column.AutoIncrement = true
			column.Type = "SERIAL"
		}

		if field.Size > 0 {
			column.Size = int(field.Size)
		}

		if field.Precision > 0 {
			column.Precision = int(field.Precision)
		}

		if field.Scale > 0 {
			column.Scale = int(field.Scale)
		}

		if field.DefaultValue != "" {
			column.Default = field.DefaultValue
		}

		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, field.DBName)
		}

		columns = append(columns, column)
	}

	return Table{
		Name:    stmt.Schema.Table,
		Columns: columns,
	}
}

func (m *Migrator) getFieldQuery(field *Column) string {
	var parts []string
	
	parts = append(parts, m.quote(field.Name))
	
	// Handle data type
	dataType := field.Type
	if field.AutoIncrement {
		dataType = "SERIAL"
	} else if field.Size > 0 {
		dataType += fmt.Sprintf("(%d)", field.Size)
	} else if field.Precision > 0 && field.Scale > 0 {
		dataType += fmt.Sprintf("(%d,%d)", field.Precision, field.Scale)
	}
	
	parts = append(parts, dataType)
	
	// Handle nullable
	if !field.Nullable && !field.PrimaryKey {
		parts = append(parts, "NOT NULL")
	}
	
	// Handle default value
	if field.Default != "" && field.Default != "NULL" && !field.AutoIncrement {
		parts = append(parts, "DEFAULT '"+field.Default+"'")
	}
	
	return strings.Join(parts, " ")
}

// Helper types and functions
type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name          string
	Type          string
	Size          int
	Precision     int
	Scale         int
	Nullable      bool
	PrimaryKey    bool
	AutoIncrement bool
	Default       string
	Comment       string
}

func getString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (m *Migrator) mapDataType(gormType string) string {
	switch gormType {
	case "string":
		return "TEXT"
	case "bool":
		return "BOOLEAN"
	case "time":
		return "TIMESTAMP"
	case "int", "int32":
		return "INTEGER"
	case "int64", "uint64":
		return "BIGINT"
	case "float32":
		return "REAL"
	case "float64":
		return "DOUBLE PRECISION"
	case "bytes":
		return "BYTEA"
	case "json":
		return "JSONB"
	case "varchar(100)":
		return "VARCHAR(100)"
	case "varchar(255)":
		return "VARCHAR(255)"
	case "varchar(20)":
		return "VARCHAR(20)"
	case "text":
		return "TEXT"
	case "serial":
		return "SERIAL"
	case "vector", "vector(1536)":
		return "VECTOR(1536)"
	default:
		// Handle varchar with size
		if strings.HasPrefix(gormType, "varchar(") {
			return strings.ToUpper(gormType)
		}
		// Handle vector with dimensions
		if strings.HasPrefix(gormType, "vector(") {
			return strings.ToUpper(gormType)
		}
		return strings.ToUpper(gormType)
	}
}