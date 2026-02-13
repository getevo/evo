package pgsql

import (
	"fmt"
	"strings"

	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/log"
	"gorm.io/gorm"
)

// PGDialect implements schema.Dialect for PostgreSQL.
type PGDialect struct {
	triggerFuncExists map[string]bool // track emitted trigger functions per column
	schema            string          // PostgreSQL schema name (defaults to 'public')
}

// NewPGDialect creates a new PostgreSQL dialect with the specified schema name.
// If schema is empty, defaults to 'public'.
func NewPGDialect(schema string) *PGDialect {
	if schema == "" {
		schema = "public"
	}
	return &PGDialect{
		triggerFuncExists: make(map[string]bool),
		schema:            schema,
	}
}

// Schema returns the configured schema name.
func (p *PGDialect) Schema() string {
	if p.schema == "" {
		return "public"
	}
	return p.schema
}

func (p *PGDialect) Name() string {
	return "postgres"
}

func (p *PGDialect) Quote(name string) string {
	return `"` + name + `"`
}

func (p *PGDialect) GetCurrentDatabase(db *gorm.DB) string {
	var database string
	if err := db.Raw("SELECT current_database()").Scan(&database).Error; err != nil {
		log.Error("failed to get current database", "error", err)
		return ""
	}
	return database
}

func (p *PGDialect) GetTableVersion(db *gorm.DB, database, tableName string) string {
	var comment string
	if err := db.Raw(`
		SELECT COALESCE(obj_description(c.oid), '0.0.0')
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = ? AND n.nspname = ? AND c.relkind = 'r'
	`, tableName, p.Schema()).Scan(&comment).Error; err != nil {
		log.Error("failed to get table version", "error", err, "table", tableName)
		return "0.0.0"
	}
	if comment == "" {
		comment = "0.0.0"
	}
	return comment
}

func (p *PGDialect) SetTableVersionSQL(tableName, version string) string {
	return fmt.Sprintf(`COMMENT ON TABLE "%s" IS '%s';`, tableName, strings.ReplaceAll(version, "'", "''"))
}

func (p *PGDialect) GetJoinConstraints(db *gorm.DB, database string) []schema.JoinConstraint {
	var raw []pgConstraint
	db.Raw(`
		SELECT con.conname                              AS constraint_name,
		       cl.relname                               AS table_name,
		       att.attname                              AS column_name,
		       ref_cl.relname                           AS referenced_table_name,
		       ref_att.attname                          AS referenced_column_name
		FROM pg_constraint con
		JOIN pg_class cl ON con.conrelid = cl.oid
		JOIN pg_namespace ns ON cl.relnamespace = ns.oid
		JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey)
		JOIN pg_class ref_cl ON con.confrelid = ref_cl.oid
		JOIN pg_attribute ref_att ON ref_att.attrelid = con.confrelid AND ref_att.attnum = ANY(con.confkey)
		WHERE con.contype = 'f'
		  AND ns.nspname = ?
	`, p.Schema()).Scan(&raw)
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

func (p *PGDialect) GenerateMigration(db *gorm.DB, database string, stmts []*gorm.Statement, models []any) schema.MigrationResult {
	return p.generateMigration(db, database, stmts, models)
}

func (p *PGDialect) AcquireMigrationLock(db *gorm.DB) error {
	return db.Exec("SELECT pg_advisory_lock(hashtext('schema_migration_lock'))").Error
}

func (p *PGDialect) ReleaseMigrationLock(db *gorm.DB) {
	if err := db.Exec("SELECT pg_advisory_unlock(hashtext('schema_migration_lock'))").Error; err != nil {
		log.Error("failed to release migration lock", "error", err)
	}
}

func (p *PGDialect) BootstrapHistoryTable(db *gorm.DB) error {
	return db.Exec(`CREATE TABLE IF NOT EXISTS "schema_migration" (
  "id" BIGSERIAL PRIMARY KEY,
  "hash" CHAR(32) NOT NULL,
  "status" VARCHAR(10) NOT NULL,
  "executed_queries" INT NOT NULL DEFAULT 0,
  "error_message" TEXT,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`).Error
}
