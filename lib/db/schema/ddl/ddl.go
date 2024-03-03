package ddl

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db/schema/table"
	"gorm.io/gorm"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var Engine = "mysql"

var EngineDataTypes = map[string]map[string]string{
	"mariadb": {"json": "longtext"},
}

var InternalFunctions = [][]string{
	{"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP()", "current_timestamp()", "current_timestamp", "NOW()", "now()", "CURRENT_DATE", "CURRENT_DATE()", "current_date", "current_date()"},
	{"NULL", "null"},
}

var (
	DefaultEngine    = "INNODB"
	DefaultCharset   = "utf8mb4"
	DefaultCollation = "utf8mb4_unicode_ci"
)

type Table struct {
	Columns     Columns
	PrimaryKey  Columns
	Index       Indexes
	Constraints []string
	Engine      string
	Name        string
	Charset     string
	Collate     string
}

type Column struct {
	Name       string
	Nullable   bool
	PrimaryKey bool
	//Size          int
	Scale         int
	Precision     int
	Type          string
	Default       string
	AutoIncrement bool
	Unique        bool
	Comment       string
	Charset       string
	OnUpdate      string
	Collate       string
	ForeignKey    string
	After         string
}

type Columns []Column

func (list Columns) Find(name string) *Column {
	for idx, _ := range list {
		if list[idx].Name == name {
			return &list[idx]
		}
	}
	return nil
}

func (list Columns) Keys() []string {
	var result []string
	for _, item := range list {
		result = append(result, item.Name)
	}
	return result
}

type Index struct {
	Name    string
	Unique  bool
	Columns Columns
}

type Indexes []Index

func (list Indexes) Find(name string) *Index {
	for idx, _ := range list {
		if list[idx].Name == name {
			return &list[idx]
		}
	}
	return nil
}

func FromStatement(stmt *gorm.Statement) Table {
	var table = Table{
		Name:    stmt.Table,
		Engine:  GetEngine(stmt),
		Charset: GetCharset(stmt),
		Collate: GetCollate(stmt),
	}

	for _, field := range stmt.Schema.Fields {
		if field.IgnoreMigration || field.DBName == "" {
			continue
		}

		var datatype = stmt.Dialector.DataTypeOf(field)
		if strings.Contains(datatype, "enum") {
			datatype = cleanEnum(datatype)
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
		}

		var column = Column{
			Name:          field.DBName,
			Type:          datatype,
			Scale:         field.Scale,
			Precision:     field.Precision,
			Default:       trimQuotes(field.DefaultValue),
			AutoIncrement: field.AutoIncrement,
			Comment:       field.Comment,
			PrimaryKey:    field.PrimaryKey,
			Unique:        field.Unique,
		}

		if v, ok := field.TagSettings["CHARSET"]; ok {
			column.Charset = v
		}

		if v, ok := field.TagSettings["FK"]; ok {
			column.ForeignKey = v
		}

		if v, ok := field.TagSettings["COLLATE"]; ok {
			column.Collate = v
		}

		if strings.ToLower(column.Type) == "datetime(3)" {
			column.Type = "TIMESTAMP"
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

		if column.Type == "TIMESTAMP" && column.Default == "" && !column.Nullable {
			column.Default = "0000-00-00 00:00:00"
		}

		if column.Name == "deleted_at" {
			column.Nullable = true
			column.Default = "NULL"
		}
		if column.Name == "created_at" {
			column.Nullable = false
			column.Default = "CURRENT_TIMESTAMP()"
		}
		if column.Name == "updated_at" {
			column.Nullable = false
			column.Default = "CURRENT_TIMESTAMP()"
			column.OnUpdate = "CURRENT_TIMESTAMP()"
		}

		if column.Unique {
			table.Index = append(table.Index, Index{
				Name:    "idx_unique_" + column.Name,
				Unique:  true,
				Columns: Columns{column},
			})
		}
		table.Columns = append(table.Columns, column)
		if field.PrimaryKey {
			table.PrimaryKey = append(table.PrimaryKey, column)
		}
	}

	for _, index := range stmt.Schema.ParseIndexes() {
		var idx = Index{
			Name:   index.Name,
			Unique: index.Class == "UNIQUE",
		}
		for _, opt := range index.Fields {
			idx.Columns = append(idx.Columns, *table.Columns.Find(opt.DBName))
		}

		table.Index = append(table.Index, idx)
	}

	return table
}

func cleanEnum(str string) string {
	var result = ""
	var inside = false
	for _, char := range str {
		if char == '\'' || char == '"' || char == '`' {
			inside = !inside
		}
		if !inside && char == ' ' {
			continue
		}
		result += string(char)
	}
	return result
}

func (table Table) GetCreateQuery() []string {
	var queries []string
	var query = "CREATE TABLE IF NOT EXISTS " + quote(table.Name) + "("
	var primaryKeys []string
	for idx, _ := range table.Columns {
		var field = table.Columns[idx]
		query += "\r\n\t"
		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, quote(field.Name))
			field.Nullable = false
		}
		query += getFieldQuery(&field)
		if idx < len(table.Columns)-1 {
			query += ","
		}
	}
	if len(primaryKeys) > 0 {
		query += ","
		query += "\r\n\t" + "PRIMARY KEY (" + strings.Join(primaryKeys, ",") + ")"
	}
	query += "\r\n) DEFAULT CHARSET=" + table.Charset + " COLLATE=" + table.Collate + " ENGINE=" + table.Engine + " COMMENT '0.0.0';"
	queries = append(queries, query)

	for _, index := range table.Index {
		query = "CREATE "
		if index.Unique {
			query += "UNIQUE "
		}
		var keys = index.Columns.Keys()
		for idx, _ := range keys {
			keys[idx] = quote(keys[idx])
		}
		query += "INDEX `" + index.Name + "` ON `" + table.Name + "` (" + strings.Join(keys, ",") + ");"
		queries = append(queries, query)
	}
	return queries
}

func getFieldQuery(field *Column) string {
	var query = quote(field.Name)
	query += " " + fieldType(field.Type)
	if field.AutoIncrement {
		query += " AUTO_INCREMENT"
	}

	if field.Default != "" {
		var v = field.Default
		if field.Default != "" {
			if (strings.ToLower(field.Type) == "timestamp" || strings.ToLower(field.Type) == "datetime") && !field.Nullable && field.Default == "" {
				v = strconv.Quote("0000-00-00 00:00:00")
			} else {
				var needQuote = true
				for _, fns := range InternalFunctions {
					if slices.Contains(fns, field.Default) {
						needQuote = false
						break
					}
				}
				if needQuote {
					v = strconv.Quote(v)
				}
			}
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

func (local Table) GetDiff(remote table.Table) []string {
	var queries []string
	var afterPK []string
	var primaryKeys []string
	if local.Collate != remote.Collation {
		queries = append(queries, fmt.Sprintf("--  table collation does not match. new:%s old:%s", local.Collate, remote.Collation))
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s CONVERT TO COLLATE %s;", quote(local.Name), local.Collate))
	}
	if local.Charset != remote.Charset {
		queries = append(queries, fmt.Sprintf("--  table charset does not match. new:%s old:%s", local.Charset, remote.Charset))
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s CONVERT TO CHARACTER SET %s;", local.Name, local.Charset))
	}

	if strings.ToLower(local.Engine) != strings.ToLower(remote.Engine) {
		queries = append(queries, fmt.Sprintf("--  table engine does not match. new:%s old:%s", local.Engine, remote.Engine))
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s ENGINE=%s;", quote(local.Name), local.Engine))
	}

	for idx, _ := range local.Columns {
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
				diff = true
			}
			if fieldType(strings.ToLower(field.Type)) != fieldType(strings.ToLower(r.ColumnType)) {
				queries = append(queries, fmt.Sprintf("--  type does not match. new:%s old:%s", fieldType(field.Type), strings.ToLower(r.ColumnType)))
				diff = true
			}
			if len(field.Collate) > 0 && strings.ToLower(field.Collate) != strings.ToLower(r.Collation) {
				queries = append(queries, fmt.Sprintf("--  collation does not match. new:%s old:%s", field.Collate, r.Collation))
				diff = true
			}
			if len(field.Charset) > 0 && strings.ToLower(field.Charset) != strings.ToLower(r.CharacterSet) {
				queries = append(queries, fmt.Sprintf("--  charset does not match. new:%s old:%s", field.Charset, r.CharacterSet))
				diff = true
			}
			if field.Comment != r.Comment {
				queries = append(queries, fmt.Sprintf("--  comment does not match. new:%s old:%s", field.Comment, r.Comment))
				diff = true
			}
			if field.Default != getString(r.ColumnDefault) {
				var skip = false
				for _, row := range InternalFunctions {
					if slices.Contains(row, field.Default) && slices.Contains(row, getString(r.ColumnDefault)) {
						skip = true
					}
				}
				if field.Default == "" && getString(r.ColumnDefault) == "0000-00-00 00:00:00" {
					skip = true
				}

				if !skip && !(field.Default == "NULL" && r.ColumnDefault == nil) {
					queries = append(queries, fmt.Sprintf("--  default value does not match. new:%s old:%s", field.Default, getString(r.ColumnDefault)))
					diff = true
				}
			}
			if field.Nullable != (r.Nullable == "YES") {
				queries = append(queries, fmt.Sprintf("--  nullable does not match. new:%t old:%t", field.Nullable, r.Nullable == "YES"))
				diff = true
			}
			var needPK = false
			if field.AutoIncrement && strings.ToLower(r.Extra) != "auto_increment" {
				afterPK = append(afterPK, fmt.Sprintf("--  auto_increment does not match. new:%t old:%t", field.AutoIncrement, !field.AutoIncrement))
				diff = true
				needPK = true
			}
			if diff {
				var position = ""
				if idx > 0 {
					position = " AFTER " + quote(local.Columns[idx-1].Name)
				}
				if needPK {
					afterPK = append(afterPK, "ALTER TABLE "+quote(local.Name)+" MODIFY COLUMN "+getFieldQuery(&field)+position+";")
				} else {
					queries = append(queries, "ALTER TABLE "+quote(local.Name)+" MODIFY COLUMN "+getFieldQuery(&field)+position+";")
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
			for idx, _ := range keys {
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
				queries = append(queries, "DROP INDEX "+quote(index.Name)+" ON "+local.Name+";")
				var query = "CREATE "
				if index.Unique {
					query += "UNIQUE "
				}
				var keys = index.Columns.Keys()
				for idx, _ := range keys {
					keys[idx] = quote(keys[idx])
				}
				query += "INDEX `" + index.Name + "` ON `" + local.Name + "` (" + strings.Join(keys, ",") + ");"
				queries = append(queries, query)
			}
		}

	}
	for _, index := range remote.Indexes {
		if local.Index.Find(index.Name) == nil {
			if !strings.HasPrefix(index.Name, "fk_") {
				queries = append(queries, "-- drop unnecessary index")
				queries = append(queries, "DROP INDEX "+quote(index.Name)+" ON "+quote(local.Name)+";")
			}
		}
	}
	return queries
}

func fieldType(t string) string {
	if v, ok := EngineDataTypes[Engine]; ok {
		if v, ok := v[t]; ok {
			return v
		}
	}
	return t
}

func getInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func getString(v *string) string {
	if v == nil {
		return ""
	} else {
		return trimQuotes(*v)
	}
}

func quote(name string) string {
	return "`" + name + "`"
}

func GetEngine(statement *gorm.Statement) string {
	if v, ok := statement.Model.(interface{ TableEngine() string }); ok {
		return v.TableEngine()
	}
	return DefaultEngine
}

func GetCharset(statement *gorm.Statement) string {
	if v, ok := statement.Model.(interface{ TableCharset() string }); ok {
		return v.TableCharset()
	}
	return DefaultCharset
}

func GetCollate(statement *gorm.Statement) string {
	if v, ok := statement.Model.(interface{ TableCollation() string }); ok {
		return v.TableCollation()
	}
	return DefaultCollation
}

func (local Table) Constrains(constraints []table.Constraint, is table.Tables) []string {
	var queries []string
	for idx, _ := range local.Columns {
		var field = local.Columns[idx]
		//Foreign Keys
		if field.ForeignKey != "" {
			var referencedTable = ""
			var referencedCol = ""
			var chunks = strings.Split(field.ForeignKey, ".")
			/*			if len(chunks) == 1 {
						if tb := schema.Find(chunks[0]); tb != nil {
							referencedTable = tb.Table
							referencedCol = tb.PrimaryKey[0]
						}
					}*/
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

			if referencedTable != "" && referencedCol != "" {

				var name = "fk_" + local.Name + "." + field.Name + "_" + referencedTable + "." + referencedCol
				if len(name) > 64 {
					name = "fk_" + field.Name + "_" + referencedTable + "." + referencedCol
				}
				var skip = false
				for _, constraint := range constraints {
					//fmt.Println(constraint.Table, "==", local.Name, constraint.Column, "==", field.Name, constraint.ReferencedTable, "==", dstTable, constraint.ReferencedColumn, "==", dstCol)
					if constraint.Table == local.Name && constraint.Column == field.Name && constraint.ReferencedTable == referencedTable && constraint.ReferencedColumn == referencedCol {
						skip = true
					}
				}

				if !skip {
					queries = append(queries, "-- create foreign key")
					var onDelete = "RESTRICT"
					queries = append(queries, "ALTER TABLE "+quote(local.Name)+" ADD CONSTRAINT "+quote(name)+" FOREIGN KEY ("+quote(field.Name)+") REFERENCES  "+quote(referencedTable)+"("+quote(referencedCol)+") ON DELETE "+onDelete+" ON UPDATE RESTRICT")
				}
				/*		ALTER TABLE `rabbits`
						ADD CONSTRAINT `fk_rabbits_main_page` FOREIGN KEY IF NOT EXISTS
						(`main_page_id`) REFERENCES `rabbit_pages` (`id`);*/
			}
		}
	}
	return queries
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
