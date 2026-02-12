package mysql

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

// engine tracks the detected MySQL/MariaDB engine type.
var engine = "mysql"

var engineDataTypes = map[string]map[string]string{
	"mariadb": {
		"json":                 "longtext",
		"current_timestamp(3)": "CURRENT_TIMESTAMP()",
		"datetime(3)":          "timestamp",
	},
}

func fromStatementToTable(stmt *gorm.Statement) ddlTable {
	var t = ddlTable{
		Name:    stmt.Table,
		Engine:  getEngine(stmt),
		Charset: getCharset(stmt),
		Collate: getCollate(stmt),
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

		switch datatype {
		case "boolean":
			datatype = "tinyint(1)"
		case "bigint":
			datatype = "bigint(20)"
		case "datetime(3)":
			datatype = "timestamp"
		case "longtext":
			// GORM defaults string without size to longtext; use varchar(255) instead
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

		// normalize bool defaults for MySQL tinyint(1)
		if strings.ToLower(datatype) == "tinyint(1)" || field.FieldType.Kind() == reflect.Bool {
			if column.Default == "false" {
				column.Default = "0"
			} else if column.Default == "true" {
				column.Default = "1"
			}
		}

		// fix gorm problem with primaryKey
		if (column.Type == "bigint(20)" || column.Type == "bigint" || column.Type == "int") && field.FieldType.Kind() == reflect.String {
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

		if v, ok := field.TagSettings["CHARSET"]; ok {
			column.Charset = v
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

		if v, ok := field.TagSettings["COLLATE"]; ok {
			column.Collate = v
		}

		if _, ok := field.TagSettings["FULLTEXT"]; ok {
			column.FullText = true
		}

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
				if column.Type == "TIMESTAMP" && column.Default == "0000-00-00 00:00:00" {
					column.Default = "NULL"
				}
				if column.Default == "" {
					column.Default = "NULL"
				}
			}
		}

		if (column.Type == "TIMESTAMP" || column.Type == "timestamp") && column.Default == "" && !column.Nullable {
			column.Default = "CURRENT_TIMESTAMP"
		}

		if column.Unique {
			t.Index = append(t.Index, schema.Index{
				Name:    schema.SafeIndexName("idx_unique_"+column.Name, 64),
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
			Name:     schema.SafeIndexName(index.Name, 64),
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

// ddlTable is the MySQL-private DDL table representation.
type ddlTable struct {
	Columns    schema.Columns
	PrimaryKey schema.Columns
	Index      schema.Indexes
	Engine     string
	Name       string
	Charset    string
	Collate    string
}

func getCreateQuery(t ddlTable) []string {
	var queries []string
	var query = "CREATE TABLE IF NOT EXISTS " + quote(t.Name) + "("
	var primaryKeys []string
	var fullTextIndexes []string
	for idx := range t.Columns {
		var field = t.Columns[idx]
		query += "\r\n\t"
		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, quote(field.Name))
			field.Nullable = false
		}
		query += getFieldQuery(&field)
		if idx < len(t.Columns)-1 {
			query += ","
		}
		if t.Columns[idx].FullText {
			fullTextIndexes = append(fullTextIndexes, quote(field.Name))
		}
	}
	if len(primaryKeys) > 0 {
		query += ","
		query += "\r\n\t" + "PRIMARY KEY (" + strings.Join(primaryKeys, ",") + ")"
	}
	if len(fullTextIndexes) > 0 {
		query += ","
		query += "\r\n\t" + "FULLTEXT (" + strings.Join(fullTextIndexes, ",") + ")"
	}

	query += "\r\n) DEFAULT CHARSET=" + t.Charset + " COLLATE=" + t.Collate + " ENGINE=" + t.Engine + " COMMENT '0.0.0';"
	queries = append(queries, query)

	for _, index := range t.Index {
		query = "CREATE "
		if index.Unique {
			query += "UNIQUE "
		}
		if index.FullText {
			query += "FULLTEXT "
		}
		var keys = index.Columns.Keys()
		for idx := range keys {
			keys[idx] = quote(keys[idx])
		}
		query += "INDEX `" + index.Name + "` ON `" + t.Name + "` (" + strings.Join(keys, ",") + ");"
		queries = append(queries, query)
	}
	return queries
}

func getFieldQuery(field *schema.Column) string {
	var query = quote(field.Name)
	query += " " + fieldType(field.Type)
	if field.AutoIncrement {
		query += " AUTO_INCREMENT"
	}

	if field.Default != "" {
		var v = field.Default
		var needQuote = true
		for _, fns := range schema.InternalFunctions {
			if slices.Contains(fns, field.Default) {
				needQuote = false
				break
			}
		}
		if needQuote {
			v = strconv.Quote(v)
		}
		query += " DEFAULT " + v
	}

	if field.OnUpdate != "" {
		query += " ON UPDATE " + field.OnUpdate
	}
	if len(field.Comment) > 0 {
		query += " COMMENT " + strconv.Quote(field.Comment)
	}

	if len(field.Charset) > 0 {
		query += " CHARACTER SET " + field.Charset
	}

	if len(field.Collate) > 0 {
		query += " COLLATE " + field.Collate
	}

	if field.Nullable {
		query += " NULL"
	} else {
		query += " NOT NULL"
	}
	return query
}

func getDiff(local ddlTable, remote remoteTable) []string {
	var queries []string
	var afterPK []string
	var primaryKeys []string
	if local.Charset != remote.Charset || local.Collate != remote.Collation {
		queries = append(queries, fmt.Sprintf("--  table charset does not match. new:%s old:%s", local.Charset, remote.Charset))
		queries = append(queries, fmt.Sprintf("--  table collate does not match. new:%s old:%s", local.Collate, remote.Collation))
		queries = append(queries, fmt.Sprintf("ALTER TABLE `%s` DEFAULT CHARACTER SET %s COLLATE %s", local.Name, local.Charset, local.Collate))
	}

	if strings.ToLower(local.Engine) != strings.ToLower(remote.Engine) {
		queries = append(queries, fmt.Sprintf("--  table engine does not match. new:%s old:%s", local.Engine, remote.Engine))
		queries = append(queries, fmt.Sprintf("ALTER TABLE `%s` ENGINE=%s;", quote(local.Name), local.Engine))
	}

	for idx := range local.Columns {
		var field = local.Columns[idx]
		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, quote(field.Name))
			field.Nullable = false
		}

		if r := remote.Columns.GetColumn(field.Name); r == nil {
			var position = ""
			if idx > 0 {
				position = " AFTER " + quote(local.Columns[idx-1].Name)
			}
			queries = append(queries, fmt.Sprintf("--  column %s does not exists", field.Name))
			queries = append(queries, "ALTER TABLE "+quote(local.Name)+" ADD "+getFieldQuery(&field)+position+";")
		} else {
			var diff = false

			if idx > 0 && idx < len(remote.Columns) && remote.Columns[idx].Name != field.Name {
				queries = append(queries, fmt.Sprintf("-- column %s position does not match", field.Name))
				diff = true
			}

			if fieldType(strings.ToLower(field.Type)) != fieldType(strings.ToLower(r.ColumnType)) {
				queries = append(queries, fmt.Sprintf("-- column %s type does not match. new:%s old:%s", field.Name, fieldType(field.Type), strings.ToLower(r.ColumnType)))
				diff = true
			}
			if len(field.Collate) > 0 && strings.ToLower(field.Collate) != strings.ToLower(r.Collation) {
				queries = append(queries, fmt.Sprintf("-- column %s collation does not match. new:%s old:%s", field.Name, field.Collate, r.Collation))
				diff = true
			}
			if len(field.Charset) > 0 && strings.ToLower(field.Charset) != strings.ToLower(r.CharacterSet) {
				queries = append(queries, fmt.Sprintf("-- column %s charset does not match. new:%s old:%s", field.Name, field.Charset, r.CharacterSet))
				diff = true
			}
			if field.Comment != r.Comment {
				queries = append(queries, fmt.Sprintf("-- column %s comment does not match. new:%s old:%s", field.Name, field.Comment, r.Comment))
				diff = true
			}
			if field.Default != getString(r.ColumnDefault) {
				var skip = false
				for _, row := range schema.InternalFunctions {
					if slices.Contains(row, field.Default) && slices.Contains(row, getString(r.ColumnDefault)) {
						skip = true
					}
				}
				if field.Default == "" && getString(r.ColumnDefault) == "0000-00-00 00:00:00" {
					skip = true
				}

				if !skip && !(field.Default == "NULL" && r.ColumnDefault == nil) {
					queries = append(queries, fmt.Sprintf("-- field %s default value does not match. new:%s old:%s", field.Name, field.Default, getString(r.ColumnDefault)))
					diff = true
				}
			}
			if field.Nullable != (r.Nullable == "YES") {
				queries = append(queries, fmt.Sprintf("-- column %s nullable does not match. new:%t old:%t", field.Name, field.Nullable, r.Nullable == "YES"))
				diff = true
			}
			var needPK = false
			if field.AutoIncrement && strings.ToLower(r.Extra) != "auto_increment" {
				afterPK = append(afterPK, fmt.Sprintf("--  field %s auto_increment does not match. new:%t old:%t", field.Name, field.AutoIncrement, !field.AutoIncrement))
				diff = true
				needPK = true
			}
			if diff {
				var position = ""
				if idx > 0 {
					position = " AFTER " + quote(local.Columns[idx-1].Name)
				}
				var q = "ALTER TABLE " + quote(local.Name) + " MODIFY COLUMN " + getFieldQuery(&field) + position + ";"
				if needPK {
					afterPK = append(afterPK, q)
				} else {
					queries = append(queries, q)
				}

			}
		}

	}
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
				queries = append(queries, fmt.Sprintf("--  column %s does not exists on schema", column.Name))
				queries = append(queries, "ALTER TABLE "+quote(local.Name)+" DROP COLUMN "+quote(column.Name)+";")
			}
		}
	}

	if pksMatch {
		pksMatch = len(pks) == len(local.PrimaryKey.Keys())
	}
	if !pksMatch {
		queries = append(queries, fmt.Sprintf("-- primary key does not match. new:%s old:%s", strings.Join(local.PrimaryKey.Keys(), ","), strings.Join(pks, ",")))
		if len(pks) > 0 {
			queries = append(queries, "ALTER TABLE "+quote(local.Name)+" DROP PRIMARY KEY;")
		}
		queries = append(queries, "ALTER TABLE "+quote(local.Name)+" ADD PRIMARY KEY("+strings.Join(local.PrimaryKey.Keys(), ",")+");")
	}
	queries = append(queries, afterPK...)
	for _, index := range local.Index {
		var r = remote.Indexes.Find(index.Name)
		if r == nil {
			var query = "CREATE "
			if index.Unique {
				query += "UNIQUE "
			}
			var keys = index.Columns.Keys()
			for idx := range keys {
				keys[idx] = quote(keys[idx])
			}
			query += "INDEX `" + index.Name + "` ON `" + local.Name + "` (" + strings.Join(keys, ",") + ");"
			queries = append(queries, "-- append not existing index")
			queries = append(queries, query)
		} else {
			var changed = false
			if r.Unique != index.Unique {
				changed = true
				queries = append(queries, fmt.Sprintf("-- index unique flag not match. new:%t old:%t", index.Unique, r.Unique))
			}
			if len(r.Columns) != len(index.Columns) {
				queries = append(queries, fmt.Sprintf("-- index columns does not match. new:%s old:%s", strings.Join(index.Columns.Keys(), ","), strings.Join(r.Columns.Keys(), ",")))
				changed = true
			} else {
				for idx, n := range index.Columns {
					if n.Name != r.Columns[idx].Name {
						queries = append(queries, fmt.Sprintf("-- index columns does not match. new:%s old:%s", strings.Join(index.Columns.Keys(), ","), strings.Join(r.Columns.Keys(), ",")))
						changed = true
						break
					}
				}
			}
			if changed {
				queries = append(queries, "DROP INDEX "+quote(index.Name)+" ON "+quote(local.Name)+";")
				var query = "CREATE "
				if index.Unique {
					query += "UNIQUE "
				}
				var keys = index.Columns.Keys()
				for idx := range keys {
					keys[idx] = quote(keys[idx])
				}
				query += "INDEX `" + index.Name + "` ON `" + local.Name + "` (" + strings.Join(keys, ",") + ");"
				queries = append(queries, query)
			}
		}

	}
	for _, index := range remote.Indexes {
		if findIndexCaseInsensitive(local.Index, index.Name) == nil {
			if !strings.HasPrefix(strings.ToLower(index.Name), "fk_") {
				queries = append(queries, "-- drop unnecessary index")
				queries = append(queries, "DROP INDEX "+quote(index.Name)+" ON "+quote(local.Name)+";")
			}
		}
	}
	return queries
}

func getConstraintsQuery(local ddlTable, constraints []remoteConstraint, is remoteTables) []string {
	var queries []string
	for idx := range local.Columns {
		var field = local.Columns[idx]
		//Foreign Keys
		if field.ForeignKey != "" {
			var referencedTable = ""
			var referencedCol = ""
			var chunks = strings.Split(field.ForeignKey, ".")

			if len(chunks) == 1 {
				var tb = is.GetTable(chunks[0])
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

			var name = local.Name + "." + field.Name + "_" + referencedTable + "." + referencedCol
			name = "fk_" + schema.Generate32CharHash(name)
			var skip = false
			for _, constraint := range constraints {
				if constraint.Table == local.Name && constraint.Column == field.Name && constraint.ReferencedTable == referencedTable && constraint.ReferencedColumn == referencedCol {
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
				queries = append(queries, "ALTER TABLE "+quote(local.Name)+" ADD CONSTRAINT "+quote(name)+" FOREIGN KEY ("+quote(field.Name)+") REFERENCES "+quote(referencedTable)+"("+quote(referencedCol)+") ON DELETE "+onDelete+" ON UPDATE "+onUpdate)
			}
		}
	}
	return queries
}

// generateMigration performs the full MySQL migration generation:
// introspection + diff/create for all models.
func generateMigration(db *gorm.DB, database string, stmts []*gorm.Statement, models []any) schema.MigrationResult {
	var result schema.MigrationResult
	result.TableExists = make(map[string]bool)

	// Detect engine
	var ver string
	db.Raw("SELECT VERSION();").Scan(&ver)
	ver = strings.ToLower(ver)
	switch {
	case strings.Contains(ver, "mariadb"):
		engine = "mariadb"
	default:
		engine = "mysql"
	}
	schema.SetConfig("mysql_engine", engine)

	// Introspect remote schema
	var is remoteTables
	db.Raw(`SELECT CCSA.character_set_name AS 'TABLE_CHARSET', T.*
		FROM information_schema.TABLES T,
		     information_schema.COLLATION_CHARACTER_SET_APPLICABILITY CCSA
		WHERE CCSA.collation_name = T.table_collation
		  AND T.table_schema = ?`, database).Scan(&is)

	var columns remoteColumns
	db.Raw(`SELECT * FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME ASC, ORDINAL_POSITION ASC`, database).Scan(&columns)

	var constraints []remoteConstraint
	db.Raw(`SELECT * FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE WHERE REFERENCED_TABLE_SCHEMA = ?`, database).Scan(&constraints)

	// Assemble columns into tables
	var tb *remoteTable
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
	var istats []remoteIndexStat
	db.Raw(`SELECT * FROM information_schema.statistics WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME ASC, SEQ_IN_INDEX ASC`, database).Scan(&istats)

	var indexMap = map[string]remoteIndex{}
	for _, item := range istats {
		if item.Name == "PRIMARY" {
			continue
		}
		if _, ok := indexMap[item.Table+item.Name]; !ok {
			indexMap[item.Table+item.Name] = remoteIndex{
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

		local := fromStatementToTable(stmt)
		tbl := is.GetTable(stmt.Schema.Table)

		var q []string
		if tbl != nil {
			tbl.Model = models[idx]
			tbl.Reflect = reflect.ValueOf(tbl.Model)
			q = getDiff(local, *tbl)
		} else {
			q = getCreateQuery(local)
		}

		result.Tail = append(result.Tail, getConstraintsQuery(local, constraints, is)...)
		if len(q) > 0 {
			result.Queries = append(result.Queries, "\r\n\r\n-- Migrate Model: "+stmt.Schema.ModelType.PkgPath()+"."+stmt.Schema.ModelType.Name()+"("+stmt.Schema.Table+")")
			result.Queries = append(result.Queries, q...)
		}
	}

	return result
}

// --- MySQL-specific helpers ---

func quote(name string) string {
	return "`" + name + "`"
}

func fieldType(t string) string {
	if _, ok := engineDataTypes[engine]; ok {
		if vi, ok := engineDataTypes[engine][t]; ok {
			return vi
		}
	}
	// Normalize: decimal and numeric are synonyms
	if strings.HasPrefix(t, "numeric") {
		return "decimal" + t[len("numeric"):]
	}
	return t
}

func getString(v *string) string {
	if v == nil {
		return ""
	}
	return schema.TrimQuotes(*v)
}

func getEngine(statement *gorm.Statement) string {
	if v, ok := statement.Model.(interface{ TableEngine() string }); ok {
		return v.TableEngine()
	}
	return schema.GetConfigDefault("default_engine", "INNODB")
}

func getCharset(statement *gorm.Statement) string {
	if v, ok := statement.Model.(interface{ TableCharset() string }); ok {
		return v.TableCharset()
	}
	return schema.GetConfigDefault("default_charset", "utf8mb4")
}

func getCollate(statement *gorm.Statement) string {
	if v, ok := statement.Model.(interface{ TableCollation() string }); ok {
		return v.TableCollation()
	}
	return schema.GetConfigDefault("default_collation", "utf8mb4_unicode_ci")
}

// findIndexCaseInsensitive does a case-insensitive search in schema.Indexes.
// MySQL index names are case-insensitive.
func findIndexCaseInsensitive(indexes schema.Indexes, name string) *schema.Index {
	lower := strings.ToLower(name)
	for idx := range indexes {
		if strings.ToLower(indexes[idx].Name) == lower {
			return &indexes[idx]
		}
	}
	return nil
}
