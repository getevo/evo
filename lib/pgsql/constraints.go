package pgsql

import (
	"fmt"
	"strings"

	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/log"
)

// getConstraintsQuery generates foreign key constraint DDL.
func (p *PGDialect) getConstraintsQuery(local pgDdlTable, constraints []pgConstraint, is pgRemoteTables) []string {
	var queries []string
	for idx := range local.Columns {
		var field = local.Columns[idx]
		if field.ForeignKey == "" {
			continue
		}
		var referencedTable string
		var referencedCol string
		chunks := strings.Split(field.ForeignKey, ".")
		if len(chunks) == 1 {
			tb := is.GetTable(chunks[0])
			if tb != nil && len(tb.PrimaryKey) > 0 {
				referencedTable = tb.Table
				referencedCol = tb.PrimaryKey[0].Name
			}
		} else if len(chunks) == 2 {
			referencedTable = chunks[0]
			referencedCol = chunks[1]
		}

		if referencedTable == "" || referencedCol == "" {
			log.Warning("foreign key target table not found, skipping constraint", "table", local.Name, "field", field.Name, "references", field.ForeignKey)
			continue
		}

		name := "fk_" + schema.Generate32CharHash(local.Name+"."+field.Name+"_"+referencedTable+"."+referencedCol)

		// Check if constraint already exists
		var skip = false
		for _, constraint := range constraints {
			if constraint.Table == local.Name && constraint.Column == field.Name &&
				constraint.ReferencedTable == referencedTable && constraint.ReferencedColumn == referencedCol {
				skip = true
			}
		}

		if !skip {
			queries = append(queries, "-- create foreign key")
			onDelete := "CASCADE"
			if field.OnDelete != "" {
				onDelete = field.OnDelete
			}
			onUpdate := "CASCADE"
			if field.FKOnUpdate != "" {
				onUpdate = field.FKOnUpdate
			}
			queries = append(queries, fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s;",
				p.Quote(local.Name), p.Quote(name), p.Quote(field.Name),
				p.Quote(referencedTable), p.Quote(referencedCol), onDelete, onUpdate))
		}
	}
	return queries
}
