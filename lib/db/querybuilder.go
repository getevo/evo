package db

import (
	"fmt"
	"strings"
)

// QueryBuilder interface for database-specific operations
type QueryBuilder interface {
	GetTablesQuery(database, schema string) string
	GetColumnsQuery(database, schema string) string
	GetIndexesQuery(database, schema string) string
	GetConstraintsQuery(database, schema string) string
	GetTableExistsQuery(tableName string) string
	QuoteIdentifier(name string) string
	FormatDataType(dataType string) string
	BuildCreateTableQuery(tableName string, columns []string, options map[string]string) string
	BuildAlterTableQuery(tableName, operation string, details string) string
}

// GetQueryBuilder returns the appropriate query builder for the database type
func GetQueryBuilder(info *DatabaseInfo) QueryBuilder {
	switch info.Type {
	case DatabaseMySQL, DatabaseMariaDB:
		return &MySQLQueryBuilder{info: info}
	case DatabasePostgreSQL:
		return &PostgreSQLQueryBuilder{info: info}
	case DatabaseSQLite:
		return &SQLiteQueryBuilder{info: info}
	default:
		return &MySQLQueryBuilder{info: info} // fallback
	}
}

// MySQLQueryBuilder implements QueryBuilder for MySQL/MariaDB
type MySQLQueryBuilder struct {
	info *DatabaseInfo
}

func (m *MySQLQueryBuilder) GetTablesQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT 
			CCSA.character_set_name AS 'TABLE_CHARSET',
			T.* 
		FROM information_schema.TABLES T
		LEFT JOIN information_schema.COLLATION_CHARACTER_SET_APPLICABILITY CCSA 
			ON CCSA.collation_name = T.table_collation 
		WHERE T.table_schema = '%s'`, database)
}

func (m *MySQLQueryBuilder) GetColumnsQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT * FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = '%s' 
		ORDER BY TABLE_NAME ASC, ORDINAL_POSITION ASC`, database)
}

func (m *MySQLQueryBuilder) GetIndexesQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT 
			TABLE_SCHEMA,
			TABLE_NAME,
			NON_UNIQUE,
			INDEX_NAME,
			COLUMN_NAME,
			SEQ_IN_INDEX
		FROM information_schema.statistics 
		WHERE TABLE_SCHEMA = '%s' 
		ORDER BY TABLE_NAME ASC, SEQ_IN_INDEX ASC`, database)
}

func (m *MySQLQueryBuilder) GetConstraintsQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT 
			CONSTRAINT_NAME,
			TABLE_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME,
			REFERENCED_TABLE_SCHEMA
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
		WHERE REFERENCED_TABLE_SCHEMA = '%s'`, database)
}

func (m *MySQLQueryBuilder) GetTableExistsQuery(tableName string) string {
	return fmt.Sprintf(`
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = DATABASE() AND table_name = '%s'`, tableName)
}

func (m *MySQLQueryBuilder) QuoteIdentifier(name string) string {
	return "`" + name + "`"
}

func (m *MySQLQueryBuilder) FormatDataType(dataType string) string {
	// MySQL-specific data type formatting
	switch strings.ToLower(dataType) {
	case "boolean":
		return "TINYINT(1)"
	case "json":
		if m.info.Type == DatabaseMariaDB {
			return "LONGTEXT"
		}
		return "JSON"
	default:
		return dataType
	}
}

func (m *MySQLQueryBuilder) BuildCreateTableQuery(tableName string, columns []string, options map[string]string) string {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n)",
		m.QuoteIdentifier(tableName),
		strings.Join(columns, ",\n\t"))
	
	// Add MySQL-specific options
	if engine, ok := options["engine"]; ok {
		query += fmt.Sprintf(" ENGINE=%s", engine)
	}
	if charset, ok := options["charset"]; ok {
		query += fmt.Sprintf(" DEFAULT CHARSET=%s", charset)
	}
	if collation, ok := options["collation"]; ok {
		query += fmt.Sprintf(" COLLATE=%s", collation)
	}
	
	return query + ";"
}

func (m *MySQLQueryBuilder) BuildAlterTableQuery(tableName, operation, details string) string {
	return fmt.Sprintf("ALTER TABLE %s %s %s;",
		m.QuoteIdentifier(tableName), operation, details)
}

// PostgreSQLQueryBuilder implements QueryBuilder for PostgreSQL
type PostgreSQLQueryBuilder struct {
	info *DatabaseInfo
}

func (p *PostgreSQLQueryBuilder) GetTablesQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT 
			'%s' as table_schema,
			tablename as table_name,
			'BASE TABLE' as table_type,
			'' as engine,
			'' as row_format,
			0 as table_rows,
			0 as auto_increment,
			'' as table_collation,
			'UTF8' as table_charset
		FROM pg_tables 
		WHERE schemaname = '%s'`, database, schema)
}

func (p *PostgreSQLQueryBuilder) GetColumnsQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT 
			'%s' as table_schema,
			table_name,
			column_name,
			ordinal_position,
			column_default,
			CASE WHEN is_nullable = 'YES' THEN 'YES' ELSE 'NO' END as is_nullable,
			data_type,
			CASE 
				WHEN data_type = 'character varying' THEN 'varchar(' || character_maximum_length || ')'
				WHEN data_type = 'character' THEN 'char(' || character_maximum_length || ')'
				WHEN data_type = 'text' THEN 'text'
				WHEN data_type = 'integer' THEN 'int'
				WHEN data_type = 'bigint' THEN 'bigint'
				WHEN data_type = 'boolean' THEN 'boolean'
				WHEN data_type = 'timestamp without time zone' THEN 'timestamp'
				WHEN data_type = 'timestamp with time zone' THEN 'timestamptz'
				ELSE data_type 
			END as column_type,
			character_maximum_length,
			numeric_precision,
			numeric_scale,
			datetime_precision,
			'' as character_set_name,
			'' as collation_name,
			CASE WHEN column_name = ANY(
				SELECT unnest(array_agg(a.attname))
				FROM pg_index i
				JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
				WHERE i.indrelid = (quote_ident(table_schema)||'.'||quote_ident(table_name))::regclass
				AND i.indisprimary
			) THEN 'PRI' ELSE '' END as column_key,
			CASE WHEN column_default LIKE 'nextval%%' THEN 'auto_increment' ELSE '' END as extra,
			'' as column_comment
		FROM information_schema.columns 
		WHERE table_schema = '%s'
		ORDER BY table_name ASC, ordinal_position ASC`, database, schema)
}

func (p *PostgreSQLQueryBuilder) GetIndexesQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT 
			'%s' as table_schema,
			t.relname as table_name,
			NOT i.indisunique as non_unique,
			i.relname as index_name,
			a.attname as column_name,
			a.attnum as seq_in_index
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE n.nspname = '%s'
		AND t.relkind = 'r'
		ORDER BY t.relname ASC, a.attnum ASC`, database, schema)
}

func (p *PostgreSQLQueryBuilder) GetConstraintsQuery(database, schema string) string {
	return fmt.Sprintf(`
		SELECT 
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS referenced_table_name,
			ccu.column_name AS referenced_column_name,
			ccu.table_schema AS referenced_table_schema
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu 
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_schema = '%s'`, schema)
}

func (p *PostgreSQLQueryBuilder) GetTableExistsQuery(tableName string) string {
	return fmt.Sprintf(`
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = current_schema() AND table_name = '%s'`, tableName)
}

func (p *PostgreSQLQueryBuilder) QuoteIdentifier(name string) string {
	return "\"" + name + "\""
}

func (p *PostgreSQLQueryBuilder) FormatDataType(dataType string) string {
	// PostgreSQL-specific data type formatting
	switch strings.ToLower(dataType) {
	case "tinyint(1)", "boolean":
		return "BOOLEAN"
	case "varchar":
		return "VARCHAR"
	case "text":
		return "TEXT"
	case "json":
		return "JSONB"
	case "bigint(20)", "bigint":
		return "BIGINT"
	case "int":
		return "INTEGER"
	case "timestamp":
		return "TIMESTAMP"
	case "datetime(3)", "datetime":
		return "TIMESTAMP"
	default:
		return dataType
	}
}

func (p *PostgreSQLQueryBuilder) BuildCreateTableQuery(tableName string, columns []string, options map[string]string) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n);",
		p.QuoteIdentifier(tableName),
		strings.Join(columns, ",\n\t"))
}

func (p *PostgreSQLQueryBuilder) BuildAlterTableQuery(tableName, operation, details string) string {
	return fmt.Sprintf("ALTER TABLE %s %s %s;",
		p.QuoteIdentifier(tableName), operation, details)
}

// SQLiteQueryBuilder implements QueryBuilder for SQLite
type SQLiteQueryBuilder struct {
	info *DatabaseInfo
}

func (s *SQLiteQueryBuilder) GetTablesQuery(database, schema string) string {
	return `
		SELECT 
			'main' as table_schema,
			name as table_name,
			'BASE TABLE' as table_type,
			'' as engine,
			'' as row_format,
			0 as table_rows,
			0 as auto_increment,
			'' as table_collation,
			'UTF8' as table_charset
		FROM sqlite_master 
		WHERE type = 'table' 
		AND name NOT LIKE 'sqlite_%'`
}

func (s *SQLiteQueryBuilder) GetColumnsQuery(database, schema string) string {
	// SQLite requires dynamic query building per table
	return "PRAGMA table_info(?)"
}

func (s *SQLiteQueryBuilder) GetIndexesQuery(database, schema string) string {
	return `
		SELECT 
			'main' as table_schema,
			tbl_name as table_name,
			1 as non_unique,
			name as index_name,
			'' as column_name,
			0 as seq_in_index
		FROM sqlite_master 
		WHERE type = 'index' 
		AND tbl_name NOT LIKE 'sqlite_%'`
}

func (s *SQLiteQueryBuilder) GetConstraintsQuery(database, schema string) string {
	// SQLite foreign keys require parsing CREATE TABLE statements
	return "SELECT sql FROM sqlite_master WHERE type = 'table'"
}

func (s *SQLiteQueryBuilder) GetTableExistsQuery(tableName string) string {
	return fmt.Sprintf(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type = 'table' AND name = '%s'`, tableName)
}

func (s *SQLiteQueryBuilder) QuoteIdentifier(name string) string {
	return "`" + name + "`"
}

func (s *SQLiteQueryBuilder) FormatDataType(dataType string) string {
	// SQLite-specific data type formatting
	switch strings.ToLower(dataType) {
	case "boolean", "tinyint(1)":
		return "INTEGER"
	case "varchar", "char":
		return "TEXT"
	case "text":
		return "TEXT"
	case "json":
		return "TEXT"
	case "bigint(20)", "bigint", "int":
		return "INTEGER"
	case "timestamp", "datetime(3)", "datetime":
		return "DATETIME"
	default:
		return "TEXT"
	}
}

func (s *SQLiteQueryBuilder) BuildCreateTableQuery(tableName string, columns []string, options map[string]string) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n);",
		s.QuoteIdentifier(tableName),
		strings.Join(columns, ",\n\t"))
}

func (s *SQLiteQueryBuilder) BuildAlterTableQuery(tableName, operation, details string) string {
	return fmt.Sprintf("ALTER TABLE %s %s %s;",
		s.QuoteIdentifier(tableName), operation, details)
}