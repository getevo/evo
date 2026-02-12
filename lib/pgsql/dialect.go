package pgsql

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/log"
	"gorm.io/gorm"
)

// PGDialect implements schema.Dialect for PostgreSQL.
type PGDialect struct {
	triggerFuncExists map[string]bool // track emitted trigger functions per column
}

func (p *PGDialect) Name() string {
	return "postgres"
}

func (p *PGDialect) Quote(name string) string {
	return `"` + name + `"`
}

func (p *PGDialect) GetCurrentDatabase(db *gorm.DB) string {
	var database string
	db.Raw("SELECT current_database()").Scan(&database)
	return database
}

func (p *PGDialect) GetTableVersion(db *gorm.DB, database, tableName string) string {
	var comment string
	db.Raw(`
		SELECT COALESCE(obj_description(c.oid), '0.0.0')
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = ? AND n.nspname = 'public' AND c.relkind = 'r'
	`, tableName).Scan(&comment)
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
		  AND ns.nspname = 'public'
	`).Scan(&raw)
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
		log.Error("failed to release migration lock: ", err)
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

// --- Private introspection types ---

type pgRemoteTable struct {
	Database      string          `gorm:"column:table_schema"`
	Table         string          `gorm:"column:table_name"`
	Type          string          `gorm:"column:table_type"`
	Engine        string          `gorm:"column:engine"`
	Charset       string          `gorm:"column:table_charset"`
	Collation     string          `gorm:"column:table_collation"`
	Columns       pgRemoteColumns `gorm:"-"`
	Indexes       pgRemoteIndexes `gorm:"-"`
	Model         any             `gorm:"-"`
	Reflect       reflect.Value   `gorm:"-"`
	PrimaryKey    pgRemoteColumns `gorm:"-"`
}

type pgRemoteTables []pgRemoteTable

func (t pgRemoteTables) GetTable(table string) *pgRemoteTable {
	for idx := range t {
		if t[idx].Table == table {
			return &t[idx]
		}
	}
	return nil
}

type pgRemoteColumn struct {
	Database        string  `gorm:"column:table_schema"`
	Table           string  `gorm:"column:table_name"`
	Name            string  `gorm:"column:column_name"`
	OrdinalPosition int     `gorm:"column:ordinal_position"`
	ColumnDefault   *string `gorm:"column:column_default"`
	Nullable        string  `gorm:"column:is_nullable"`
	DataType        string  `gorm:"column:data_type"`
	ColumnType      string  `gorm:"column:column_type"`
	Size            *int    `gorm:"column:character_maximum_length"`
	Precision       *int    `gorm:"column:numeric_precision"`
	Scale           *int    `gorm:"column:numeric_scale"`
	DatePrecision   *int    `gorm:"column:datetime_precision"`
	CharacterSet    string  `gorm:"column:character_set_name"`
	Collation       string  `gorm:"column:collation_name"`
	ColumnKey       string  `gorm:"column:column_key"`
	Extra           string  `gorm:"column:extra"`
	Comment         string  `gorm:"column:column_comment"`
}

type pgRemoteColumns []pgRemoteColumn

func (t pgRemoteColumns) GetColumn(column string) *pgRemoteColumn {
	for idx := range t {
		if t[idx].Name == column {
			return &t[idx]
		}
	}
	return nil
}

func (t pgRemoteColumns) Keys() []string {
	var result []string
	for idx := range t {
		result = append(result, t[idx].Name)
	}
	return result
}

type pgConstraint struct {
	Name             string `gorm:"column:constraint_name"`
	Table            string `gorm:"column:table_name"`
	Column           string `gorm:"column:column_name"`
	ReferencedTable  string `gorm:"column:referenced_table_name"`
	ReferencedColumn string `gorm:"column:referenced_column_name"`
}

type pgRemoteIndexStat struct {
	Database   string `gorm:"column:table_schema"`
	Table      string `gorm:"column:table_name"`
	NonUnique  bool   `gorm:"column:non_unique"`
	Name       string `gorm:"column:index_name"`
	ColumnName string `gorm:"column:column_name"`
}

type pgRemoteIndex struct {
	Name    string
	Table   string
	Unique  bool
	Columns pgRemoteColumns
}

type pgRemoteIndexes []pgRemoteIndex

func (list pgRemoteIndexes) Find(name string) *pgRemoteIndex {
	for idx := range list {
		if list[idx].Name == name {
			return &list[idx]
		}
	}
	return nil
}

// --- DDL Table type (local model representation) ---

type pgDdlTable struct {
	Columns    schema.Columns
	PrimaryKey schema.Columns
	Index      schema.Indexes
	Name       string
}

// --- Migration Generation ---

func (p *PGDialect) generateMigration(db *gorm.DB, database string, stmts []*gorm.Statement, models []any) schema.MigrationResult {
	var result schema.MigrationResult
	result.TableExists = make(map[string]bool)

	// Introspect remote tables
	var is pgRemoteTables
	db.Raw(`
		SELECT table_catalog   AS table_schema,
		       table_name      AS table_name,
		       table_type      AS table_type,
		       ''              AS engine,
		       ''              AS table_collation,
		       ''              AS table_charset
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_catalog = ?
		  AND table_type = 'BASE TABLE'
	`, database).Scan(&is)

	// Introspect remote columns
	var columns pgRemoteColumns
	db.Raw(`
		SELECT c.table_catalog                AS table_schema,
		       c.table_name                   AS table_name,
		       c.column_name                  AS column_name,
		       c.ordinal_position             AS ordinal_position,
		       c.column_default               AS column_default,
		       c.is_nullable                  AS is_nullable,
		       c.udt_name                     AS data_type,
		       CASE
		           WHEN c.data_type = 'USER-DEFINED' THEN c.udt_name
		           WHEN c.udt_name = 'varchar' AND c.character_maximum_length IS NOT NULL
		               THEN 'varchar(' || c.character_maximum_length || ')'
		           WHEN c.udt_name = 'bpchar' AND c.character_maximum_length IS NOT NULL
		               THEN 'char(' || c.character_maximum_length || ')'
		           WHEN c.udt_name = 'numeric' AND c.numeric_precision IS NOT NULL
		               THEN 'decimal(' || c.numeric_precision || ',' || COALESCE(c.numeric_scale, 0) || ')'
		           WHEN c.udt_name = 'int4' THEN 'int'
		           WHEN c.udt_name = 'int8' THEN 'bigint'
		           WHEN c.udt_name = 'int2' THEN 'smallint'
		           WHEN c.udt_name = 'float4' THEN 'real'
		           WHEN c.udt_name = 'float8' THEN 'double precision'
		           WHEN c.udt_name = 'bool' THEN 'boolean'
		           WHEN c.udt_name = 'timestamptz' THEN 'timestamptz'
		           WHEN c.udt_name = 'timestamp' THEN 'timestamp'
		           WHEN c.udt_name = 'text' THEN 'text'
		           WHEN c.udt_name = 'jsonb' THEN 'jsonb'
		           WHEN c.udt_name = 'json' THEN 'json'
		           ELSE c.udt_name
		       END                            AS column_type,
		       c.character_maximum_length      AS character_maximum_length,
		       c.numeric_precision             AS numeric_precision,
		       c.numeric_scale                 AS numeric_scale,
		       c.datetime_precision            AS datetime_precision,
		       ''                              AS character_set_name,
		       ''                              AS collation_name,
		       CASE
		           WHEN pk.column_name IS NOT NULL THEN 'PRI'
		           ELSE ''
		       END                            AS column_key,
		       CASE
		           WHEN c.column_default LIKE 'nextval(%%' THEN 'auto_increment'
		           ELSE ''
		       END                            AS extra,
		       COALESCE(col_description(cls.oid, c.ordinal_position), '') AS column_comment
		FROM information_schema.columns c
		LEFT JOIN pg_class cls
		       ON cls.relname = c.table_name AND cls.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		LEFT JOIN (
		    SELECT ku.table_name, ku.column_name
		    FROM information_schema.key_column_usage ku
		    JOIN information_schema.table_constraints tc
		        ON ku.constraint_name = tc.constraint_name AND ku.table_schema = tc.table_schema
		    WHERE tc.constraint_type = 'PRIMARY KEY'
		      AND ku.table_schema = 'public'
		) pk ON pk.table_name = c.table_name AND pk.column_name = c.column_name
		WHERE c.table_schema = 'public'
		  AND c.table_catalog = ?
		ORDER BY c.table_name, c.ordinal_position
	`, database).Scan(&columns)

	// Introspect constraints
	var constraints []pgConstraint
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
		  AND ns.nspname = 'public'
	`).Scan(&constraints)

	// Assemble columns into tables
	var tb *pgRemoteTable
	for idx := range columns {
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

	// Assemble indexes
	var istats []pgRemoteIndexStat
	db.Raw(`
		SELECT ?                             AS table_schema,
		       t.relname                     AS table_name,
		       NOT ix.indisunique            AS non_unique,
		       i.relname                     AS index_name,
		       a.attname                     AS column_name
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_namespace ns ON t.relnamespace = ns.oid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE ns.nspname = 'public'
		  AND NOT ix.indisprimary
		ORDER BY t.relname, array_position(ix.indkey, a.attnum)
	`, database).Scan(&istats)

	var indexMap = map[string]pgRemoteIndex{}
	for _, item := range istats {
		if item.Name == "PRIMARY" {
			continue
		}
		if _, ok := indexMap[item.Table+item.Name]; !ok {
			indexMap[item.Table+item.Name] = pgRemoteIndex{
				Name:   item.Name,
				Table:  item.Table,
				Unique: !item.NonUnique,
			}
		}
		var m = indexMap[item.Table+item.Name]
		tbl := is.GetTable(item.Table)
		if tbl == nil {
			continue
		}
		var c = tbl.Columns.GetColumn(item.ColumnName)
		if c == nil {
			continue
		}
		m.Columns = append(m.Columns, *c)
		indexMap[item.Table+item.Name] = m
	}
	for key, item := range indexMap {
		tbl := is.GetTable(item.Table)
		if tbl != nil {
			tbl.Indexes = append(tbl.Indexes, indexMap[key])
		}
	}

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

// --- DDL Generation ---

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
		var q string
		if index.FullText {
			// GIN index for fulltext
			var cols []string
			for _, c := range index.Columns {
				cols = append(cols, fmt.Sprintf(`to_tsvector('english', %s)`, p.Quote(c.Name)))
			}
			q = fmt.Sprintf(`CREATE INDEX %s ON %s USING gin(%s);`,
				p.Quote(index.Name), p.Quote(t.Name), strings.Join(cols, " || ' ' || "))
		} else {
			q = "CREATE "
			if index.Unique {
				q += "UNIQUE "
			}
			var keys []string
			for _, c := range index.Columns {
				keys = append(keys, p.Quote(c.Name))
			}
			q += fmt.Sprintf("INDEX %s ON %s (%s);", p.Quote(index.Name), p.Quote(t.Name), strings.Join(keys, ","))
		}
		queries = append(queries, q)
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
			log.Warning("foreign key on ", local.Name, ".", field.Name, " references '", field.ForeignKey, "' but target table not found, skipping constraint")
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

// --- Helper functions ---

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

func (p *PGDialect) pgColumnType(field *schema.Column, t *pgDdlTable) string {
	colType := field.Type

	// Handle ENUM types
	if strings.HasPrefix(strings.ToLower(colType), "enum(") {
		return `"` + t.Name + "_" + field.Name + `_enum"`
	}

	return colType
}

func (p *PGDialect) nullClause(field *schema.Column) string {
	if field.Nullable {
		return " NULL"
	}
	return " NOT NULL"
}

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

// zeroDefault returns the appropriate zero-value default for a PG column type.
// Used when adding a NOT NULL column to an existing table.
func (p *PGDialect) zeroDefault(colType string) string {
	t := strings.ToLower(strings.TrimSpace(colType))
	switch {
	case strings.HasPrefix(t, "varchar"), strings.HasPrefix(t, "character varying"),
		t == "text", strings.HasPrefix(t, "char"):
		return "''"
	case t == "boolean", t == "bool":
		return "false"
	case t == "bigint", t == "int8", t == "integer", t == "int4", t == "int",
		t == "smallint", t == "int2":
		return "0"
	case strings.HasPrefix(t, "numeric"), strings.HasPrefix(t, "decimal"),
		t == "real", t == "float4", t == "double precision", t == "float8":
		return "0"
	case t == "timestamp", t == "timestamptz",
		strings.HasPrefix(t, "timestamp"):
		return "CURRENT_TIMESTAMP"
	case t == "date":
		return "CURRENT_DATE"
	case t == "jsonb", t == "json":
		return "'{}'"
	default:
		// For enum types and unknown types, skip zero default
		return ""
	}
}

func (p *PGDialect) normalizeType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	// Normalize PG aliases
	switch t {
	case "int4", "integer":
		return "int"
	case "int8":
		return "bigint"
	case "int2":
		return "smallint"
	case "float4":
		return "real"
	case "float8":
		return "double precision"
	case "bool":
		return "boolean"
	case "timestamptz":
		return "timestamptz"
	case "tinyint(1)":
		return "boolean"
	case "bigint(20)":
		return "bigint"
	case "decimal":
		return "numeric"
	}
	// Normalize decimal(p,s) -> numeric(p,s)
	if strings.HasPrefix(t, "decimal(") {
		return "numeric" + t[7:]
	}
	return t
}

func (p *PGDialect) normalizeDefault(d string) string {
	d = strings.TrimSpace(d)
	// Strip PG type casts like 'value'::character varying, 'a'::enum_type, etc.
	if idx := strings.Index(d, "::"); idx > 0 {
		d = d[:idx]
	}
	d = schema.TrimQuotes(d)
	d = strings.ToLower(d)
	// Normalize PG-style nextval defaults
	if strings.HasPrefix(d, "nextval(") {
		return ""
	}
	// Normalize boolean defaults
	switch d {
	case "1", "true", "t", "'t'":
		return "true"
	case "0", "false", "f", "'f'":
		return "false"
	}
	// Normalize timestamp variants
	switch d {
	case "current_timestamp", "current_timestamp()", "now()":
		return "current_timestamp"
	case "null", "":
		return ""
	}
	return d
}

func needsUsing(fromType, toType string) bool {
	from := strings.ToLower(fromType)
	to := strings.ToLower(toType)
	if from == to {
		return false
	}
	// Common conversions that need USING
	if (from == "text" || strings.HasPrefix(from, "varchar")) && (to == "int" || to == "bigint" || to == "boolean") {
		return true
	}
	if (from == "int" || from == "bigint") && to == "boolean" {
		return true
	}
	if from == "boolean" && (to == "int" || to == "bigint") {
		return true
	}
	return false
}

// parseEnumValues returns the individual unquoted values from an enum definition.
// Input: "enum('a','b','c')" -> ["a", "b", "c"]
func parseEnumValues(enumType string) []string {
	s := enumType
	idx := strings.Index(strings.ToLower(s), "enum(")
	if idx >= 0 {
		s = s[idx+5:]
	}
	s = strings.TrimSuffix(s, ")")
	s = strings.TrimSpace(s)
	var vals []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "'\"")
		if part != "" {
			vals = append(vals, part)
		}
	}
	return vals
}

func extractEnumValues(enumType string) string {
	// Input: enum('a','b','c') or enum("a","b")
	// Output: 'a','b','c' (with single quotes for PG)
	s := enumType
	// Remove "enum(" prefix and ")" suffix
	idx := strings.Index(strings.ToLower(s), "enum(")
	if idx >= 0 {
		s = s[idx+5:]
	}
	s = strings.TrimSuffix(s, ")")
	s = strings.TrimSpace(s)
	// Normalize quotes to single quotes
	s = strings.ReplaceAll(s, `"`, `'`)
	return s
}

func getStringPtr(v *string) string {
	if v == nil {
		return ""
	}
	return schema.TrimQuotes(*v)
}
