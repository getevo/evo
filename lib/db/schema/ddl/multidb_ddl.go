package ddl

import (
	"strconv"
	"strings"
)

// DatabaseType represents different database types
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	MariaDB    DatabaseType = "mariadb"
	PostgreSQL DatabaseType = "postgresql"
	SQLite     DatabaseType = "sqlite"
)

// MultiDBDDL provides database-specific DDL generation
type MultiDBDDL struct {
	DatabaseType DatabaseType
}

// NewMultiDBDDL creates a new multi-database DDL generator
func NewMultiDBDDL(dbType DatabaseType) *MultiDBDDL {
	return &MultiDBDDL{DatabaseType: dbType}
}

// GetCreateTableQuery generates database-specific CREATE TABLE query
func (m *MultiDBDDL) GetCreateTableQuery(table Table) []string {
	switch m.DatabaseType {
	case SQLite:
		return m.getSQLiteCreateQuery(table)
	case PostgreSQL:
		return m.getPostgreSQLCreateQuery(table)
	case MySQL, MariaDB:
		return m.getMySQLCreateQuery(table)
	default:
		return m.getMySQLCreateQuery(table) // fallback
	}
}

// getSQLiteCreateQuery generates SQLite-specific CREATE TABLE query
func (m *MultiDBDDL) getSQLiteCreateQuery(table Table) []string {
	var queries []string
	var query = "CREATE TABLE IF NOT EXISTS " + m.quote(table.Name) + " ("
	var compositePrimaryKeys []string
	var hasAutoIncrementPK = false
	
	// Check if we have an auto-increment primary key
	for _, field := range table.Columns {
		if field.PrimaryKey && field.AutoIncrement {
			hasAutoIncrementPK = true
			break
		}
	}
	
	for idx, field := range table.Columns {
		query += "\r\n    "
		
		// For composite primary keys without auto-increment, collect them but don't mark individual fields as PRIMARY KEY
		if field.PrimaryKey && !hasAutoIncrementPK {
			compositePrimaryKeys = append(compositePrimaryKeys, m.quote(field.Name))
			field.Nullable = false
			// Remove PrimaryKey flag to avoid individual PRIMARY KEY declarations
			fieldCopy := field
			fieldCopy.PrimaryKey = false
			query += m.getSQLiteFieldQuery(&fieldCopy)
		} else {
			query += m.getSQLiteFieldQuery(&field)
		}
		
		if idx < len(table.Columns)-1 {
			query += ","
		}
	}
	
	// Add composite primary key if we have multiple primary keys or single non-auto-increment PK
	if len(compositePrimaryKeys) > 0 && !hasAutoIncrementPK {
		query += ","
		query += "\r\n    PRIMARY KEY (" + strings.Join(compositePrimaryKeys, ", ") + ")"
	}
	
	query += "\r\n);"
	queries = append(queries, query)
	
	// SQLite indexes are created separately
	for _, index := range table.Index {
		indexQuery := "CREATE "
		if index.Unique {
			indexQuery += "UNIQUE "
		}
		var keys = index.Columns.Keys()
		for i := range keys {
			keys[i] = m.quote(keys[i])
		}
		indexQuery += "INDEX IF NOT EXISTS " + m.quote(index.Name) + " ON " + m.quote(table.Name) + " (" + strings.Join(keys, ", ") + ");"
		queries = append(queries, indexQuery)
	}
	
	return queries
}

// getPostgreSQLCreateQuery generates PostgreSQL-specific CREATE TABLE query
func (m *MultiDBDDL) getPostgreSQLCreateQuery(table Table) []string {
	var queries []string
	var query = "CREATE TABLE IF NOT EXISTS " + m.quote(table.Name) + " ("
	var compositePrimaryKeys []string
	var hasAutoIncrementPK = false
	
	// Check if we have an auto-increment primary key
	for _, field := range table.Columns {
		if field.PrimaryKey && field.AutoIncrement {
			hasAutoIncrementPK = true
			break
		}
	}
	
	for idx, field := range table.Columns {
		query += "\r\n    "
		
		// For composite primary keys without auto-increment
		if field.PrimaryKey && !hasAutoIncrementPK {
			compositePrimaryKeys = append(compositePrimaryKeys, m.quote(field.Name))
			field.Nullable = false
			// Remove PrimaryKey flag to avoid individual PRIMARY KEY declarations
			fieldCopy := field
			fieldCopy.PrimaryKey = false
			query += m.getPostgreSQLFieldQuery(&fieldCopy)
		} else {
			query += m.getPostgreSQLFieldQuery(&field)
		}
		
		if idx < len(table.Columns)-1 {
			query += ","
		}
	}
	
	// Add composite primary key if we have multiple primary keys or single non-auto-increment PK
	if len(compositePrimaryKeys) > 0 && !hasAutoIncrementPK {
		query += ","
		query += "\r\n    PRIMARY KEY (" + strings.Join(compositePrimaryKeys, ", ") + ")"
	}
	
	query += "\r\n);"
	queries = append(queries, query)
	
	// PostgreSQL indexes
	for _, index := range table.Index {
		indexQuery := "CREATE "
		if index.Unique {
			indexQuery += "UNIQUE "
		}
		var keys = index.Columns.Keys()
		for i := range keys {
			keys[i] = m.quote(keys[i])
		}
		indexQuery += "INDEX IF NOT EXISTS " + m.quote(index.Name) + " ON " + m.quote(table.Name) + " (" + strings.Join(keys, ", ") + ");"
		queries = append(queries, indexQuery)
	}
	
	return queries
}

// getMySQLCreateQuery generates MySQL/MariaDB-specific CREATE TABLE query
func (m *MultiDBDDL) getMySQLCreateQuery(table Table) []string {
	// Use existing MySQL implementation
	return table.GetCreateQuery()
}

// getSQLiteFieldQuery generates SQLite-specific field definition
func (m *MultiDBDDL) getSQLiteFieldQuery(field *Column) string {
	query := m.quote(field.Name)
	
	// SQLite type mapping
	sqliteType := m.mapToSQLiteType(field.Type)
	query += " " + sqliteType
	
	// In SQLite, PRIMARY KEY AUTOINCREMENT must be together
	if field.PrimaryKey && field.AutoIncrement {
		query += " PRIMARY KEY AUTOINCREMENT"
	} else if field.PrimaryKey {
		query += " PRIMARY KEY"
	}
	
	if field.Default != "" && field.Default != "NULL" {
		defaultValue := m.convertSQLiteDefault(field.Default, field.Type)
		if m.isInternalFunction(field.Default) {
			query += " DEFAULT " + defaultValue
		} else {
			query += " DEFAULT " + defaultValue
		}
	}
	
	if !field.Nullable {
		query += " NOT NULL"
	}
	
	if field.Unique && !field.PrimaryKey {
		query += " UNIQUE"
	}
	
	return query
}

// getPostgreSQLFieldQuery generates PostgreSQL-specific field definition
func (m *MultiDBDDL) getPostgreSQLFieldQuery(field *Column) string {
	query := m.quote(field.Name)
	
	// PostgreSQL type mapping with special handling for auto-increment
	pgType := m.mapToPostgreSQLType(field.Type)
	
	// Special case: if it's INTEGER with boolean default, make it BOOLEAN
	if strings.ToLower(pgType) == "integer" && (field.Default == "true" || field.Default == "false") {
		pgType = "BOOLEAN"
	}
	
	// Handle auto-increment fields
	if field.AutoIncrement && field.PrimaryKey {
		if strings.Contains(strings.ToLower(field.Type), "bigint") {
			query += " BIGSERIAL PRIMARY KEY"
		} else {
			query += " SERIAL PRIMARY KEY"
		}
	} else {
		query += " " + pgType
		
		if field.PrimaryKey {
			query += " PRIMARY KEY"
		}
	}
	
	if field.Default != "" && field.Default != "NULL" && !field.AutoIncrement {
		defaultValue := m.convertPostgreSQLDefault(field.Default, pgType)
		if m.isInternalFunction(field.Default) {
			query += " DEFAULT " + defaultValue
		} else {
			query += " DEFAULT " + defaultValue
		}
	}
	
	if !field.Nullable && !field.PrimaryKey {
		query += " NOT NULL"
	}
	
	if field.Unique && !field.PrimaryKey {
		query += " UNIQUE"
	}
	
	return query
}

// mapToSQLiteType maps MySQL/GORM types to SQLite types
func (m *MultiDBDDL) mapToSQLiteType(mysqlType string) string {
	mysqlType = strings.ToLower(mysqlType)
	
	// Handle enum types
	if strings.Contains(mysqlType, "enum") {
		return "TEXT" // SQLite doesn't have ENUM, use TEXT with CHECK constraint
	}
	
	// Handle sized types
	if strings.Contains(mysqlType, "varchar") || strings.Contains(mysqlType, "char") {
		return "TEXT"
	}
	
	switch {
	case strings.Contains(mysqlType, "int"):
		return "INTEGER"
	case strings.Contains(mysqlType, "float") || strings.Contains(mysqlType, "double") || strings.Contains(mysqlType, "decimal"):
		return "REAL"
	case strings.Contains(mysqlType, "text") || strings.Contains(mysqlType, "longtext"):
		return "TEXT"
	case strings.Contains(mysqlType, "timestamp") || strings.Contains(mysqlType, "datetime"):
		return "DATETIME"
	case strings.Contains(mysqlType, "date"):
		return "DATE"
	case strings.Contains(mysqlType, "time"):
		return "TIME"
	case strings.Contains(mysqlType, "bool") || strings.Contains(mysqlType, "tinyint(1)"):
		return "INTEGER"
	case strings.Contains(mysqlType, "json"):
		return "TEXT"
	case strings.Contains(mysqlType, "vector(") || strings.Contains(mysqlType, "[]float32"):
		return "TEXT" // SQLite fallback for vectors
	default:
		return "TEXT"
	}
}

// mapToPostgreSQLType maps MySQL/GORM types to PostgreSQL types
func (m *MultiDBDDL) mapToPostgreSQLType(mysqlType string) string {
	mysqlType = strings.ToLower(mysqlType)
	
	// Handle enum types - PostgreSQL supports ENUM but needs to be created first
	if strings.Contains(mysqlType, "enum") {
		return "VARCHAR(50)" // Simplified for now
	}
	
	// Handle sized types
	if strings.Contains(mysqlType, "varchar") {
		return mysqlType // PostgreSQL supports VARCHAR with size
	}
	
	// Handle vector types with dimensions
	if strings.Contains(mysqlType, "vector(") {
		return mysqlType // Return as-is: vector(1536), vector(384), etc.
	}
	
	switch {
	case strings.Contains(mysqlType, "bigint"):
		return "BIGINT"
	case strings.Contains(mysqlType, "int"):
		return "INTEGER"
	case strings.Contains(mysqlType, "float"):
		return "REAL"
	case strings.Contains(mysqlType, "double"):
		return "DOUBLE PRECISION"
	case strings.Contains(mysqlType, "decimal"):
		return "DECIMAL"
	case strings.Contains(mysqlType, "text") || strings.Contains(mysqlType, "longtext"):
		return "TEXT"
	case strings.Contains(mysqlType, "timestamp"):
		return "TIMESTAMP"
	case strings.Contains(mysqlType, "datetime"):
		return "TIMESTAMP"
	case strings.Contains(mysqlType, "date"):
		return "DATE"
	case strings.Contains(mysqlType, "time"):
		return "TIME"
	case strings.Contains(mysqlType, "bool") || strings.Contains(mysqlType, "tinyint(1)"):
		return "BOOLEAN"
	case strings.Contains(mysqlType, "json"):
		return "JSONB"
	case strings.Contains(mysqlType, "[]float32") || strings.Contains(mysqlType, "vector"):
		return "vector" // PostgreSQL vector extension
	default:
		return "TEXT"
	}
}

// convertPostgreSQLDefault converts default values to PostgreSQL format
func (m *MultiDBDDL) convertPostgreSQLDefault(value, fieldType string) string {
	// Handle boolean values for PostgreSQL
	if strings.Contains(strings.ToLower(fieldType), "bool") {
		if strings.ToLower(value) == "true" {
			return "true"
		} else if strings.ToLower(value) == "false" {
			return "false"
		}
	}
	
	// Handle CURRENT_TIMESTAMP variants
	if m.isInternalFunction(value) {
		switch strings.ToUpper(value) {
		case "CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP()", "NOW()":
			return "CURRENT_TIMESTAMP"
		case "CURRENT_DATE", "CURRENT_DATE()":
			return "CURRENT_DATE"
		case "NULL":
			return "NULL"
		}
	}
	
	// Quote string values
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// convertSQLiteDefault converts default values to SQLite format
func (m *MultiDBDDL) convertSQLiteDefault(value, fieldType string) string {
	// Handle boolean values for SQLite
	if strings.Contains(strings.ToLower(fieldType), "bool") || strings.Contains(strings.ToLower(fieldType), "tinyint(1)") {
		if strings.ToLower(value) == "true" {
			return "1"
		} else if strings.ToLower(value) == "false" {
			return "0"
		}
	}
	
	// Handle CURRENT_TIMESTAMP variants
	if m.isInternalFunction(value) {
		switch strings.ToUpper(value) {
		case "CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP()", "NOW()":
			return "CURRENT_TIMESTAMP"
		case "CURRENT_DATE", "CURRENT_DATE()":
			return "CURRENT_DATE"
		case "NULL":
			return "NULL"
		}
	}
	
	// Quote string values
	return strconv.Quote(value)
}

// isInternalFunction checks if a value is a database function
func (m *MultiDBDDL) isInternalFunction(value string) bool {
	internalFunctions := []string{
		"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP()", "current_timestamp()", "current_timestamp",
		"NOW()", "now()", "CURRENT_DATE", "CURRENT_DATE()", "current_date", "current_date()",
		"NULL", "null",
	}
	
	for _, fn := range internalFunctions {
		if strings.EqualFold(value, fn) {
			return true
		}
	}
	return false
}

// quote wraps identifiers in database-specific quotes
func (m *MultiDBDDL) quote(name string) string {
	switch m.DatabaseType {
	case PostgreSQL:
		return `"` + name + `"`
	case SQLite:
		return `"` + name + `"`
	default: // MySQL, MariaDB
		return "`" + name + "`"
	}
}

// GetDatabaseTypeFromDialect maps GORM dialect names to DatabaseType
func GetDatabaseTypeFromDialect(dialectName string) DatabaseType {
	switch strings.ToLower(dialectName) {
	case "mysql":
		return MySQL
	case "postgres", "postgresql":
		return PostgreSQL
	case "sqlite", "sqlite3":
		return SQLite
	default:
		return MySQL // fallback
	}
}