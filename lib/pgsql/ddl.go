package pgsql

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db/schema"
	"gorm.io/gorm"
)

// fromStatementToTable converts a GORM statement into a pgDdlTable for DDL generation.
func (p *PGDialect) fromStatementToTable(stmt *gorm.Statement) pgDdlTable {
	var t = pgDdlTable{
		Name: stmt.Table,
	}

	for _, field := range stmt.Schema.Fields {
		if field.IgnoreMigration || field.DBName == "" {
			continue
		}

		var datatype = stmt.Dialector.DataTypeOf(field)

		if strings.Contains(datatype, "enum") {
			datatype = schema.CleanEnum(datatype)
		} else {
			datatype = strings.Split(datatype, " ")[0]
		}

		// PG type mapping — do NOT convert to MySQL types
		switch strings.ToLower(datatype) {
		case "datetime(3)":
			datatype = "timestamp"
		case "bigint unsigned":
			datatype = "bigint"
		case "text":
			// GORM defaults string without size to text; use varchar(255) instead
			if _, hasType := field.TagSettings["TYPE"]; !hasType {
				datatype = "varchar"
			}
		}

		var column = schema.Column{
			Name:          field.DBName,
			Type:          datatype,
			Size:          field.Size,
			Scale:         field.Scale,
			Precision:     field.Precision,
			Default:       field.DefaultValue,
			AutoIncrement: field.AutoIncrement,
			Comment:       field.Comment,
			PrimaryKey:    field.PrimaryKey,
			Unique:        field.Unique,
		}

		// Fix: if the Go type is string but GORM gives a numeric type, use varchar
		if (column.Type == "bigint" || column.Type == "int") && field.FieldType.Kind() == reflect.String {
			column.Type = "varchar"
		}
		if column.Type == "varchar" {
			if column.Size == 0 {
				column.Size = 255
			}
			column.Type = "varchar(" + strconv.Itoa(column.Size) + ")"
		}
		if column.Type == "char" {
			if column.Size == 0 {
				column.Size = 255
			}
			column.Type = "char(" + strconv.Itoa(column.Size) + ")"
		}

		if v, ok := field.TagSettings["FK"]; ok {
			column.ForeignKey = v
		}

		if v, ok := field.TagSettings["FK_ON_DELETE"]; ok {
			column.OnDelete = v
		}
		if v, ok := field.TagSettings["FK_ON_UPDATE"]; ok {
			column.FKOnUpdate = v
		}

		if _, ok := field.TagSettings["FULLTEXT"]; ok {
			column.FullText = true
		}

		// Handle ON_UPDATE tag for trigger generation
		if v, ok := field.TagSettings["ON_UPDATE"]; ok {
			column.OnUpdate = v
		}

		var nullable = false
		if _, ok := field.TagSettings["NULLABLE"]; ok {
			nullable = true
		}
		if field.FieldType.Kind() == reflect.Ptr || nullable {
			if _, ok := field.TagSettings["NOT NULL"]; !ok {
				column.Nullable = true
				// PG doesn't support 0000-00-00 — use NULL
				if (strings.ToLower(column.Type) == "timestamp" || strings.ToLower(column.Type) == "timestamptz") && column.Default == "0000-00-00 00:00:00" {
					column.Default = "NULL"
				}
				if column.Default == "" {
					column.Default = "NULL"
				}
			}
		}

		// Normalize MySQL's CURRENT_TIMESTAMP() to PG's CURRENT_TIMESTAMP
		if strings.EqualFold(column.Default, "CURRENT_TIMESTAMP()") {
			column.Default = "CURRENT_TIMESTAMP"
		}

		// For non-nullable timestamps without default, use CURRENT_TIMESTAMP (not 0000-00-00)
		if (strings.ToLower(column.Type) == "timestamp" || strings.ToLower(column.Type) == "timestamptz") && column.Default == "" && !column.Nullable {
			column.Default = "CURRENT_TIMESTAMP"
		}
		// Replace any remaining 0000-00-00 with NULL
		if column.Default == "0000-00-00 00:00:00" {
			if column.Nullable {
				column.Default = "NULL"
			} else {
				column.Default = "CURRENT_TIMESTAMP"
			}
		}

		if column.Unique {
			t.Index = append(t.Index, schema.Index{
				Name:    schema.SafeIndexName("idx_unique_"+t.Name+"_"+column.Name, 63),
				Unique:  true,
				Columns: schema.Columns{column},
			})
		}

		var r = field.IndirectFieldType
		for r.Kind() == reflect.Ptr {
			r = r.Elem()
		}
		var ref = reflect.New(r)
		if obj, ok := ref.Interface().(interface{ ColumnDefinition(column *schema.Column) }); ok {
			obj.ColumnDefinition(&column)
		} else if obj, ok := ref.Elem().Interface().(interface{ ColumnDefinition(column *schema.Column) }); ok {
			obj.ColumnDefinition(&column)
		}

		column.Default = schema.TrimQuotes(column.Default)
		t.Columns = append(t.Columns, column)
		if field.PrimaryKey {
			t.PrimaryKey = append(t.PrimaryKey, column)
		}
	}

	for _, index := range stmt.Schema.ParseIndexes() {
		var idx = schema.Index{
			Name:     schema.SafeIndexName(index.Name, 63),
			Unique:   index.Class == "UNIQUE",
			FullText: index.Class == "FULLTEXT",
		}
		var skip bool
		for _, opt := range index.Fields {
			col := t.Columns.Find(opt.DBName)
			if col == nil {
				skip = true
				break
			}
			idx.Columns = append(idx.Columns, *col)
		}
		if skip {
			continue
		}
		t.Index = append(t.Index, idx)
	}

	return t
}

// getCreateQuery generates CREATE TABLE DDL for a new table.
func (p *PGDialect) getCreateQuery(t pgDdlTable) []string {
	var queries []string

	// Pre-CREATE: ENUM types
	for _, col := range t.Columns {
		if strings.HasPrefix(strings.ToLower(col.Type), "enum(") {
			typeName := t.Name + "_" + col.Name + "_enum"
			values := extractEnumValues(col.Type)
			queries = append(queries, fmt.Sprintf(
				`DO $$ BEGIN CREATE TYPE "%s" AS ENUM (%s); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
				typeName, values,
			))
		}
	}

	// CREATE TABLE
	var query = `CREATE TABLE IF NOT EXISTS ` + p.Quote(t.Name) + `(`
	var primaryKeys []string
	var fullTextColumns []string
	var onUpdateColumns []schema.Column

	for idx := range t.Columns {
		var field = t.Columns[idx]
		query += "\r\n\t"
		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, p.Quote(field.Name))
			field.Nullable = false
		}
		query += p.getFieldQuery(&field, &t)
		if idx < len(t.Columns)-1 {
			query += ","
		}
		if field.FullText {
			fullTextColumns = append(fullTextColumns, p.Quote(field.Name))
		}
		if field.OnUpdate != "" {
			onUpdateColumns = append(onUpdateColumns, field)
		}
	}
	if len(primaryKeys) > 0 {
		query += ","
		query += "\r\n\t" + "PRIMARY KEY (" + strings.Join(primaryKeys, ",") + ")"
	}
	query += "\r\n);"
	queries = append(queries, query)

	// Post-CREATE: table comment
	queries = append(queries, fmt.Sprintf(`COMMENT ON TABLE %s IS '0.0.0';`, p.Quote(t.Name)))

	// Post-CREATE: column comments
	for _, col := range t.Columns {
		if col.Comment != "" {
			queries = append(queries, fmt.Sprintf(`COMMENT ON COLUMN %s.%s IS '%s';`,
				p.Quote(t.Name), p.Quote(col.Name), strings.ReplaceAll(col.Comment, "'", "''")))
		}
	}

	// Post-CREATE: indexes
	for _, index := range t.Index {
		queries = append(queries, p.createIndexSQL(index, t.Name))
	}

	// Post-CREATE: fulltext index on inline columns
	if len(fullTextColumns) > 0 {
		var cols []string
		for _, c := range fullTextColumns {
			cols = append(cols, fmt.Sprintf(`to_tsvector('english', %s)`, c))
		}
		idxName := "ft_" + t.Name
		queries = append(queries, fmt.Sprintf(`CREATE INDEX %s ON %s USING gin(%s);`,
			p.Quote(idxName), p.Quote(t.Name), strings.Join(cols, " || ' ' || ")))
	}

	// Post-CREATE: ON UPDATE triggers
	for _, col := range onUpdateColumns {
		queries = append(queries, p.getUpdateTriggerStatements(t.Name, col.Name)...)
	}

	return queries
}

// getDiff generates ALTER TABLE statements by comparing local and remote tables.
func (p *PGDialect) getDiff(local pgDdlTable, remote pgRemoteTable) []string {
	var queries []string
	var afterPK []string
	var primaryKeys []string

	// Check for new ENUM types needed
	for _, col := range local.Columns {
		if strings.HasPrefix(strings.ToLower(col.Type), "enum(") {
			typeName := local.Name + "_" + col.Name + "_enum"
			values := extractEnumValues(col.Type)
			queries = append(queries, fmt.Sprintf(
				`DO $$ BEGIN CREATE TYPE "%s" AS ENUM (%s); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
				typeName, values,
			))
			// Also add any new values to the enum type if it already exists
			for _, v := range parseEnumValues(col.Type) {
				queries = append(queries, fmt.Sprintf(
					`DO $$ BEGIN ALTER TYPE "%s" ADD VALUE IF NOT EXISTS '%s'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
					typeName, v,
				))
			}
		}
	}

	for idx := range local.Columns {
		var field = local.Columns[idx]
		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, p.Quote(field.Name))
			field.Nullable = false
		}

		r := remote.Columns.GetColumn(field.Name)
		if r == nil {
			// Column does not exist — ADD
			queries = append(queries, fmt.Sprintf("-- column %s does not exist", field.Name))
			addQuery := p.getFieldQuery(&field, &local)
			// PG: when adding NOT NULL column to existing table, add DEFAULT for the zero value
			if !field.Nullable && field.Default == "" && !field.AutoIncrement {
				zeroDefault := p.zeroDefault(field.Type)
				if zeroDefault != "" {
					addQuery += " DEFAULT " + zeroDefault
				}
			}
			queries = append(queries, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;",
				p.Quote(local.Name), addQuery))
		} else {
			// Column exists — check for differences
			var alterStatements []string

			// Skip type comparison for auto-increment columns (bigserial is a pseudo-type, not alterable)
			// Skip type comparison for enum columns (enum changes handled by ALTER TYPE ADD VALUE above)
			isEnum := strings.HasPrefix(strings.ToLower(field.Type), "enum(")
			if !field.AutoIncrement && !isEnum {
				// Type mismatch
				localType := p.normalizeType(field.Type)
				remoteType := p.normalizeType(r.ColumnType)
				if localType != remoteType {
					queries = append(queries, fmt.Sprintf("-- column %s type does not match. new:%s old:%s", field.Name, localType, remoteType))
					stmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
						p.Quote(local.Name), p.Quote(field.Name), p.pgColumnType(&field, &local))
					// Add USING clause for type conversions that need it
					if needsUsing(remoteType, localType) {
						stmt += fmt.Sprintf(" USING %s::%s", p.Quote(field.Name), p.pgColumnType(&field, &local))
					}
					stmt += ";"
					alterStatements = append(alterStatements, stmt)
				}
			}

			// Default mismatch
			localDefault := p.normalizeDefault(field.Default)
			remoteDefault := p.normalizeDefault(getStringPtr(r.ColumnDefault))
			if localDefault != remoteDefault {
				var skip = false
				for _, fns := range schema.InternalFunctions {
					if slices.Contains(fns, field.Default) && slices.Contains(fns, getStringPtr(r.ColumnDefault)) {
						skip = true
					}
				}
				if field.Default == "" && (getStringPtr(r.ColumnDefault) == "0000-00-00 00:00:00" || getStringPtr(r.ColumnDefault) == "") {
					skip = true
				}
				if !skip && !(field.Default == "NULL" && r.ColumnDefault == nil) {
					queries = append(queries, fmt.Sprintf("-- column %s default does not match. new:%s old:%s", field.Name, field.Default, getStringPtr(r.ColumnDefault)))
					if field.Default == "" || field.Default == "NULL" {
						alterStatements = append(alterStatements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;",
							p.Quote(local.Name), p.Quote(field.Name)))
					} else {
						var v = field.Default
						needQuote := true
						for _, fns := range schema.InternalFunctions {
							if slices.Contains(fns, field.Default) {
								needQuote = false
								break
							}
						}
						if needQuote {
							v = "'" + v + "'"
						}
						alterStatements = append(alterStatements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;",
							p.Quote(local.Name), p.Quote(field.Name), v))
					}
				}
			}

			// Nullable mismatch
			if field.Nullable != (r.Nullable == "YES") {
				queries = append(queries, fmt.Sprintf("-- column %s nullable does not match. new:%t old:%t", field.Name, field.Nullable, r.Nullable == "YES"))
				if field.Nullable {
					alterStatements = append(alterStatements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;",
						p.Quote(local.Name), p.Quote(field.Name)))
				} else {
					alterStatements = append(alterStatements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;",
						p.Quote(local.Name), p.Quote(field.Name)))
				}
			}

			// Comment mismatch
			if field.Comment != r.Comment {
				queries = append(queries, fmt.Sprintf("-- column %s comment does not match. new:%s old:%s", field.Name, field.Comment, r.Comment))
				alterStatements = append(alterStatements, fmt.Sprintf(`COMMENT ON COLUMN %s.%s IS '%s';`,
					p.Quote(local.Name), p.Quote(field.Name), field.Comment))
			}

			// Auto-increment mismatch
			if field.AutoIncrement && strings.ToLower(r.Extra) != "auto_increment" {
				afterPK = append(afterPK, fmt.Sprintf("-- column %s auto_increment does not match", field.Name))
				seqName := local.Name + "_" + field.Name + "_seq"
				afterPK = append(afterPK, fmt.Sprintf(`CREATE SEQUENCE IF NOT EXISTS "%s";`, seqName))
				afterPK = append(afterPK, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT nextval('%s');",
					p.Quote(local.Name), p.Quote(field.Name), seqName))
				afterPK = append(afterPK, fmt.Sprintf(`ALTER SEQUENCE "%s" OWNED BY %s.%s;`,
					seqName, p.Quote(local.Name), p.Quote(field.Name)))
			}

			queries = append(queries, alterStatements...)
		}
	}

	// Primary key check
	var pks []string
	var pksMatch = true
	for _, column := range remote.Columns {
		if column.ColumnKey == "PRI" {
			pks = append(pks, column.Name)
			var col = local.PrimaryKey.Find(column.Name)
			if col == nil || !col.PrimaryKey {
				pksMatch = false
			}
		}

		if args.Exists("--strict") {
			var found = false
			for _, lc := range local.Columns {
				if lc.Name == column.Name {
					found = true
					break
				}
			}
			if !found {
				queries = append(queries, fmt.Sprintf("-- column %s does not exist on schema", column.Name))
				queries = append(queries, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;",
					p.Quote(local.Name), p.Quote(column.Name)))
			}
		}
	}

	if pksMatch {
		pksMatch = len(pks) == len(local.PrimaryKey.Keys())
	}
	if !pksMatch {
		queries = append(queries, fmt.Sprintf("-- primary key does not match. new:%s old:%s",
			strings.Join(local.PrimaryKey.Keys(), ","), strings.Join(pks, ",")))
		if len(pks) > 0 {
			// Find the PK constraint name
			pkConstraint := local.Name + "_pkey"
			queries = append(queries, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;",
				p.Quote(local.Name), p.Quote(pkConstraint)))
		}
		var pkCols []string
		for _, k := range local.PrimaryKey.Keys() {
			pkCols = append(pkCols, p.Quote(k))
		}
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY(%s);",
			p.Quote(local.Name), strings.Join(pkCols, ",")))
	}
	queries = append(queries, afterPK...)

	// Index handling
	for _, index := range local.Index {
		r := remote.Indexes.Find(index.Name)
		if r == nil {
			// Index doesn't exist — create
			queries = append(queries, "-- append not existing index")
			queries = append(queries, p.createIndexSQL(index, local.Name))
		} else {
			var changed = false
			if r.Unique != index.Unique {
				changed = true
				queries = append(queries, fmt.Sprintf("-- index unique flag not match. new:%t old:%t", index.Unique, r.Unique))
			}
			if len(r.Columns) != len(index.Columns) {
				queries = append(queries, fmt.Sprintf("-- index columns does not match"))
				changed = true
			} else {
				for i, n := range index.Columns {
					if n.Name != r.Columns[i].Name {
						queries = append(queries, fmt.Sprintf("-- index columns does not match"))
						changed = true
						break
					}
				}
			}
			if changed {
				// PG: DROP INDEX without ON table
				queries = append(queries, fmt.Sprintf("DROP INDEX IF EXISTS %s;", p.Quote(index.Name)))
				queries = append(queries, p.createIndexSQL(index, local.Name))
			}
		}
	}

	// Drop removed indexes
	for _, index := range remote.Indexes {
		if local.Index.Find(index.Name) == nil {
			if !strings.HasPrefix(index.Name, "fk_") {
				queries = append(queries, "-- drop unnecessary index")
				queries = append(queries, fmt.Sprintf("DROP INDEX IF EXISTS %s;", p.Quote(index.Name)))
			}
		}
	}

	// ON UPDATE trigger check
	for _, col := range local.Columns {
		if col.OnUpdate != "" {
			queries = append(queries, p.getUpdateTriggerStatements(local.Name, col.Name)...)
		}
	}

	return queries
}

// getFieldQuery generates the column definition SQL for a field.
func (p *PGDialect) getFieldQuery(field *schema.Column, t *pgDdlTable) string {
	colType := p.pgColumnType(field, t)

	// Auto-increment PK uses BIGSERIAL/SERIAL
	if field.AutoIncrement {
		lower := strings.ToLower(colType)
		if strings.Contains(lower, "big") || strings.Contains(lower, "int8") {
			return p.Quote(field.Name) + " BIGSERIAL" + p.nullClause(field)
		}
		return p.Quote(field.Name) + " SERIAL" + p.nullClause(field)
	}

	var query = p.Quote(field.Name) + " " + colType

	if field.Default != "" {
		var v = field.Default
		needQuote := true
		for _, fns := range schema.InternalFunctions {
			if slices.Contains(fns, field.Default) {
				needQuote = false
				break
			}
		}
		if needQuote && v != "NULL" {
			v = "'" + v + "'"
		}
		query += " DEFAULT " + v
	}

	query += p.nullClause(field)
	return query
}

// pgColumnType returns the PostgreSQL column type for a field.
func (p *PGDialect) pgColumnType(field *schema.Column, t *pgDdlTable) string {
	colType := field.Type

	// Handle ENUM types
	if strings.HasPrefix(strings.ToLower(colType), "enum(") {
		return `"` + t.Name + "_" + field.Name + `_enum"`
	}

	return colType
}

// nullClause returns the NULL/NOT NULL clause for a column.
func (p *PGDialect) nullClause(field *schema.Column) string {
	if field.Nullable {
		return " NULL"
	}
	return " NOT NULL"
}

// createIndexSQL generates CREATE INDEX SQL for an index.
func (p *PGDialect) createIndexSQL(index schema.Index, tableName string) string {
	if index.FullText {
		var cols []string
		for _, c := range index.Columns {
			cols = append(cols, fmt.Sprintf(`to_tsvector('english', %s)`, p.Quote(c.Name)))
		}
		return fmt.Sprintf(`CREATE INDEX %s ON %s USING gin(%s);`,
			p.Quote(index.Name), p.Quote(tableName), strings.Join(cols, " || ' ' || "))
	}
	q := "CREATE "
	if index.Unique {
		q += "UNIQUE "
	}
	var keys []string
	for _, c := range index.Columns {
		keys = append(keys, p.Quote(c.Name))
	}
	q += fmt.Sprintf("INDEX %s ON %s (%s);", p.Quote(index.Name), p.Quote(tableName), strings.Join(keys, ","))
	return q
}

// getUpdateTriggerStatements generates trigger function and trigger DDL for ON UPDATE columns.
func (p *PGDialect) getUpdateTriggerStatements(tableName, columnName string) []string {
	var stmts []string
	if p.triggerFuncExists == nil {
		p.triggerFuncExists = make(map[string]bool)
	}
	funcName := fmt.Sprintf("update_%s_%s_column", tableName, columnName)
	if !p.triggerFuncExists[funcName] {
		stmts = append(stmts, fmt.Sprintf(
			`CREATE OR REPLACE FUNCTION %s() RETURNS TRIGGER AS $$ BEGIN NEW.%s = CURRENT_TIMESTAMP; RETURN NEW; END; $$ LANGUAGE plpgsql;`,
			funcName, columnName))
		p.triggerFuncExists[funcName] = true
	}
	triggerName := fmt.Sprintf("set_%s_%s", columnName, tableName)
	stmts = append(stmts, fmt.Sprintf(
		`DO $$ BEGIN CREATE TRIGGER "%s" BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION %s(); EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		triggerName, p.Quote(tableName), funcName))
	return stmts
}
