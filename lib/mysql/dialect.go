package mysql

import (
	"fmt"

	"github.com/getevo/evo/v2/lib/db/schema"
	"gorm.io/gorm"
)

// MySQLDialect implements schema.Dialect for MySQL/MariaDB.
type MySQLDialect struct {
	eng string
}

func (m *MySQLDialect) Name() string {
	return "mysql"
}

func (m *MySQLDialect) Quote(name string) string {
	return quote(name)
}

func (m *MySQLDialect) GetCurrentDatabase(db *gorm.DB) string {
	var database string
	db.Raw("SELECT DATABASE();").Scan(&database)
	return database
}

func (m *MySQLDialect) GenerateMigration(db *gorm.DB, database string, stmts []*gorm.Statement, models []any) schema.MigrationResult {
	return generateMigration(db, database, stmts, models)
}

func (m *MySQLDialect) GetTableVersion(db *gorm.DB, database, tableName string) string {
	var comment string
	db.Raw("SELECT table_comment FROM INFORMATION_SCHEMA.TABLES WHERE table_schema=? AND table_name=?", database, tableName).Scan(&comment)
	return comment
}

func (m *MySQLDialect) SetTableVersionSQL(tableName, version string) string {
	return fmt.Sprintf("ALTER TABLE `%s` COMMENT '%s';", tableName, version)
}

func (m *MySQLDialect) GetJoinConstraints(db *gorm.DB, database string) []schema.JoinConstraint {
	var raw []remoteConstraint
	db.Raw(`SELECT CONSTRAINT_NAME, TABLE_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME, REFERENCED_TABLE_SCHEMA
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE REFERENCED_TABLE_SCHEMA = ?`, database).Scan(&raw)
	var result []schema.JoinConstraint
	for _, c := range raw {
		result = append(result, schema.JoinConstraint{
			Table:            c.Table,
			Column:           c.Column,
			ReferencedTable:  c.ReferencedTable,
			ReferencedColumn: c.ReferencedColumn,
		})
	}
	return result
}
