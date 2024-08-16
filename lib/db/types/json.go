package types

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/getevo/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// JSONType give a generic data type for json encoded data.
type JSONType[T any] struct {
	data T
}

func NewJSONType[T any](data T) JSONType[T] {
	return JSONType[T]{
		data: data,
	}
}

// Data return data with generic Type T
func (j JSONType[T]) Data() T {
	return j.data
}

// Value return json value, implement driver.Valuer interface
func (j JSONType[T]) Value() (driver.Value, error) {
	return json.Marshal(j.data)
}

// Scan scan value into JSONType[T], implements sql.Scanner interface
func (j *JSONType[T]) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return json.Unmarshal(bytes, &j.data)
}

// MarshalJSON to output non base64 encoded []byte
func (j JSONType[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.data)
}

// UnmarshalJSON to deserialize []byte
func (j *JSONType[T]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &j.data)
}

// GormDataType gorm common data type
func (JSONType[T]) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (JSONType[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

func (js JSONType[T]) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := js.MarshalJSON()

	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}

	return gorm.Expr("?", string(data))
}

// JSONSlice give a generic data type for json encoded slice data.
type JSONSlice[T any] []T

func NewJSONSlice[T any](s []T) JSONSlice[T] {
	return JSONSlice[T](s)
}

// Value return json value, implement driver.Valuer interface
func (j JSONSlice[T]) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan scan value into JSONType[T], implements sql.Scanner interface
func (j *JSONSlice[T]) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return json.Unmarshal(bytes, &j)
}

// GormDataType gorm common data type
func (JSONSlice[T]) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (JSONSlice[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

func (j JSONSlice[T]) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := json.Marshal(j)

	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}

	return gorm.Expr("?", string(data))
}

// JSON defined JSON data type, need to implements driver.Valuer, sql.Scanner interface
type JSON json.RawMessage

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return string(j), nil
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			bytes = make([]byte, len(v))
			copy(bytes, v)
		}
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage(bytes)
	*j = JSON(result)
	return nil
}

// MarshalJSON to output non base64 encoded []byte
func (j JSON) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON to deserialize []byte
func (j *JSON) UnmarshalJSON(b []byte) error {
	result := json.RawMessage{}
	err := result.UnmarshalJSON(b)
	*j = JSON(result)
	return err
}

func (j JSON) String() string {
	return string(j)
}

// GormDataType gorm common data type
func (JSON) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (JSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

func (js JSON) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if len(js) == 0 {
		return gorm.Expr("NULL")
	}

	data, _ := js.MarshalJSON()

	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}

	return gorm.Expr("?", string(data))
}

// JSONQueryExpression json query expression, implements clause.Expression interface to use as querier
type JSONQueryExpression struct {
	column      string
	keys        []string
	hasKeys     bool
	equals      bool
	equalsValue any
	extract     bool
	path        string
}

// JSONQuery query column as json
func JSONQuery(column string) *JSONQueryExpression {
	return &JSONQueryExpression{column: column}
}

// Extract extract json with path
func (jsonQuery *JSONQueryExpression) Extract(path string) *JSONQueryExpression {
	jsonQuery.extract = true
	jsonQuery.path = path
	return jsonQuery
}

// HasKey returns clause.Expression
func (jsonQuery *JSONQueryExpression) HasKey(keys ...string) *JSONQueryExpression {
	jsonQuery.keys = keys
	jsonQuery.hasKeys = true
	return jsonQuery
}

// Keys returns clause.Expression
func (jsonQuery *JSONQueryExpression) Equals(value any, keys ...string) *JSONQueryExpression {
	jsonQuery.keys = keys
	jsonQuery.equals = true
	jsonQuery.equalsValue = value
	return jsonQuery
}

// Build implements clause.Expression
func (jsonQuery *JSONQueryExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql", "sqlite":
			switch {
			case jsonQuery.extract:
				builder.WriteString("JSON_EXTRACT(")
				builder.WriteQuoted(jsonQuery.column)
				builder.WriteByte(',')
				builder.AddVar(stmt, prefix+jsonQuery.path)
				builder.WriteString(")")
			case jsonQuery.hasKeys:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString("JSON_EXTRACT(")
					builder.WriteQuoted(jsonQuery.column)
					builder.WriteByte(',')
					builder.AddVar(stmt, jsonQueryJoin(jsonQuery.keys))
					builder.WriteString(") IS NOT NULL")
				}
			case jsonQuery.equals:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString("JSON_EXTRACT(")
					builder.WriteQuoted(jsonQuery.column)
					builder.WriteByte(',')
					builder.AddVar(stmt, jsonQueryJoin(jsonQuery.keys))
					builder.WriteString(") = ")
					if value, ok := jsonQuery.equalsValue.(bool); ok {
						builder.WriteString(strconv.FormatBool(value))
					} else {
						stmt.AddVar(builder, jsonQuery.equalsValue)
					}
				}
			}
		case "postgres":
			switch {
			case jsonQuery.extract:
				builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(jsonQuery.column)))
				stmt.AddVar(builder, jsonQuery.path)
				builder.WriteByte(')')
			case jsonQuery.hasKeys:
				if len(jsonQuery.keys) > 0 {
					stmt.WriteQuoted(jsonQuery.column)
					stmt.WriteString("::jsonb")
					for _, key := range jsonQuery.keys[0 : len(jsonQuery.keys)-1] {
						stmt.WriteString(" -> ")
						stmt.AddVar(builder, key)
					}

					stmt.WriteString(" ? ")
					stmt.AddVar(builder, jsonQuery.keys[len(jsonQuery.keys)-1])
				}
			case jsonQuery.equals:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(jsonQuery.column)))

					for idx, key := range jsonQuery.keys {
						if idx > 0 {
							builder.WriteByte(',')
						}
						stmt.AddVar(builder, key)
					}
					builder.WriteString(") = ")

					if _, ok := jsonQuery.equalsValue.(string); ok {
						stmt.AddVar(builder, jsonQuery.equalsValue)
					} else {
						stmt.AddVar(builder, fmt.Sprint(jsonQuery.equalsValue))
					}
				}
			}
		}
	}
}

// JSONOverlapsExpression JSON_OVERLAPS expression, implements clause.Expression interface to use as querier
type JSONOverlapsExpression struct {
	column clause.Expression
	val    string
}

// JSONOverlaps query column as json
func JSONOverlaps(column clause.Expression, value string) *JSONOverlapsExpression {
	return &JSONOverlapsExpression{
		column: column,
		val:    value,
	}
}

// Build implements clause.Expression
// only mysql support JSON_OVERLAPS
func (json *JSONOverlapsExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql":
			builder.WriteString("JSON_OVERLAPS(")
			json.column.Build(builder)
			builder.WriteString(",")
			builder.AddVar(stmt, json.val)
			builder.WriteString(")")
		}
	}
}

type columnExpression string

func Column(col string) columnExpression {
	return columnExpression(col)
}

func (col columnExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql", "sqlite", "postgres":
			builder.WriteString(stmt.Quote(string(col)))
		}
	}
}

const prefix = "$."

func jsonQueryJoin(keys []string) string {
	if len(keys) == 1 {
		return prefix + keys[0]
	}

	n := len(prefix)
	n += len(keys) - 1
	for i := 0; i < len(keys); i++ {
		n += len(keys[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(prefix)
	b.WriteString(keys[0])
	for _, key := range keys[1:] {
		b.WriteString(".")
		b.WriteString(key)
	}
	return b.String()
}

// JSONSetExpression json set expression, implements clause.Expression interface to use as updater
type JSONSetExpression struct {
	column     string
	path2value map[string]any
	mutex      sync.RWMutex
}

// JSONSet update fields of json column
func JSONSet(column string) *JSONSetExpression {
	return &JSONSetExpression{column: column, path2value: make(map[string]any)}
}

// Set return clause.Expression.
//
//	{
//		"age": 20,
//		"name": "json-1",
//		"orgs": {"orga": "orgv"},
//		"tags": ["tag1", "tag2"]
//	}
//
//	// In MySQL/SQLite, path is `age`, `name`, `orgs.orga`, `tags[0]`, `tags[1]`.
//	DB.UpdateColumn("attr", JSONSet("attr").Set("orgs.orga", 42))
//
//	// In PostgreSQL, path is `{age}`, `{name}`, `{orgs,orga}`, `{tags, 0}`, `{tags, 1}`.
//	DB.UpdateColumn("attr", JSONSet("attr").Set("{orgs, orga}", "bar"))
func (jsonSet *JSONSetExpression) Set(path string, value any) *JSONSetExpression {
	jsonSet.mutex.Lock()
	jsonSet.path2value[path] = value
	jsonSet.mutex.Unlock()
	return jsonSet
}

// Build implements clause.Expression
// support mysql, sqlite and postgres
func (jsonSet *JSONSetExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql":

			var isMariaDB bool
			if v, ok := stmt.Dialector.(*mysql.Dialector); ok {
				isMariaDB = strings.Contains(v.ServerVersion, "MariaDB")
			}

			builder.WriteString("JSON_SET(")
			builder.WriteQuoted(jsonSet.column)
			for path, value := range jsonSet.path2value {
				builder.WriteByte(',')
				builder.AddVar(stmt, prefix+path)
				builder.WriteByte(',')

				if _, ok := value.(clause.Expression); ok {
					stmt.AddVar(builder, value)
					continue
				}

				rv := reflect.ValueOf(value)
				if rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}
				switch rv.Kind() {
				case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
					b, _ := json.Marshal(value)
					if isMariaDB {
						stmt.AddVar(builder, string(b))
						break
					}
					stmt.AddVar(builder, gorm.Expr("CAST(? AS JSON)", string(b)))
				default:
					stmt.AddVar(builder, value)
				}
			}
			builder.WriteString(")")

		case "sqlite":
			builder.WriteString("JSON_SET(")
			builder.WriteQuoted(jsonSet.column)
			for path, value := range jsonSet.path2value {
				builder.WriteByte(',')
				builder.AddVar(stmt, prefix+path)
				builder.WriteByte(',')

				if _, ok := value.(clause.Expression); ok {
					stmt.AddVar(builder, value)
					continue
				}

				rv := reflect.ValueOf(value)
				if rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}
				switch rv.Kind() {
				case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
					b, _ := json.Marshal(value)
					stmt.AddVar(builder, gorm.Expr("JSON(?)", string(b)))
				default:
					stmt.AddVar(builder, value)
				}
			}
			builder.WriteString(")")

		case "postgres":
			var expr clause.Expression = columnExpression(jsonSet.column)
			for path, value := range jsonSet.path2value {
				if _, ok = value.(clause.Expression); ok {
					expr = gorm.Expr("JSONB_SET(?,?,?)", expr, path, value)
					continue
				} else {
					b, _ := json.Marshal(value)
					expr = gorm.Expr("JSONB_SET(?,?,?)", expr, path, string(b))
				}
			}
			stmt.AddVar(builder, expr)
		}
	}
}

func JSONArrayQuery(column string) *JSONArrayExpression {
	return &JSONArrayExpression{
		column: column,
	}
}

type JSONArrayExpression struct {
	column      string
	equalsValue any
}

func (json *JSONArrayExpression) Contains(value any) *JSONArrayExpression {
	json.equalsValue = value
	return json
}

// Build implements clause.Expression
func (json *JSONArrayExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql":
			builder.WriteString("JSON_CONTAINS (" + stmt.Quote(json.column) + ", JSON_ARRAY(")
			builder.AddVar(stmt, json.equalsValue)
			builder.WriteString("))")
		}
	}
}
