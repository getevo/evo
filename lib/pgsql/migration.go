package pgsql

import (
	"reflect"
	"strings"

	"github.com/getevo/evo/v2/lib/db/schema"
	"gorm.io/gorm"
)

// generateMigration generates migration DDL by comparing local models with remote schema.
func (p *PGDialect) generateMigration(db *gorm.DB, database string, stmts []*gorm.Statement, models []any) schema.MigrationResult {
	var result schema.MigrationResult
	result.TableExists = make(map[string]bool)

	// Introspect remote schema
	is := p.introspectTables(db, database)
	columns := p.introspectColumns(db, database)
	constraints := p.introspectConstraints(db)

	// Assemble columns into tables
	p.assembleColumns(columns, is)

	// Assemble indexes
	p.introspectIndexes(db, database, is)

	// Mark which tables exist
	for _, t := range is {
		result.TableExists[t.Table] = true
	}

	// Generate DDL for each model
	for idx, stmt := range stmts {
		if stmt.Schema == nil {
			continue
		}

		if obj, ok := stmt.Model.(interface{ TableName() string }); ok {
			if strings.HasPrefix(obj.TableName(), "information_schema.") {
				continue
			}
		}

		local := p.fromStatementToTable(stmt)
		tbl := is.GetTable(stmt.Schema.Table)

		var q []string
		if tbl != nil {
			tbl.Model = models[idx]
			tbl.Reflect = reflect.ValueOf(tbl.Model)
			q = p.getDiff(local, *tbl)
		} else {
			q = p.getCreateQuery(local)
		}

		result.Tail = append(result.Tail, p.getConstraintsQuery(local, constraints, is)...)
		if len(q) > 0 {
			result.Queries = append(result.Queries, "\r\n\r\n-- Migrate Model: "+stmt.Schema.ModelType.PkgPath()+"."+stmt.Schema.ModelType.Name()+"("+stmt.Schema.Table+")")
			result.Queries = append(result.Queries, q...)
		}
	}

	return result
}
