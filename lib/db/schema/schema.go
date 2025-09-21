package schema

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/db/schema/table"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"path/filepath"
	"reflect"
)

var Models []Model

// ClearModels clears all registered models
func ClearModels() {
	Models = []Model{}
	migrations = []any{}
}

type Model struct {
	Sample      any                 `json:"sample"`
	Value       reflect.Value       `json:"-"`
	Type        reflect.Type        `json:"-"`
	Kind        reflect.Kind        `json:"-"`
	Table       string              `json:"table"`
	Name        string              `json:"name"`
	Package     string              `json:"package"`
	PackagePath string              `json:"package_path"`
	PrimaryKey  []string            `json:"primary_key"`
	Joins       map[string][]string `json:"joins"`
	Schema      *schema.Schema      `json:"-"`
	Statement   *gorm.Statement     `json:"-"`
}

func (m Model) Join(joins ...*Model) ([]string, []string, error) {
	var where []string
	var tables = []string{m.Table}
	for _, join := range joins {
		tables = append(tables, join.Table)
		if v, ok := m.Joins[join.Table]; ok {
			where = append(where, quote(m.Table)+"."+quote(v[0])+" = "+quote(join.Table)+"."+quote(v[1]))
			continue
		}
		if _, ok := join.Schema.FieldsByDBName[m.PrimaryKey[0]]; ok {
			where = append(where, quote(m.Table)+"."+quote(m.PrimaryKey[0])+" = "+quote(join.Table)+"."+quote(m.PrimaryKey[0]))
			continue
		}
		if _, ok := m.Schema.FieldsByDBName[join.PrimaryKey[0]]; ok {
			where = append(where, quote(join.Table)+"."+quote(join.PrimaryKey[0])+" = "+quote(m.Table)+"."+quote(join.PrimaryKey[0]))
			continue
		}

		return nil, nil, fmt.Errorf("unable to find relation between %s and %s", m.Name, join.Name)
	}
	return tables, where, nil
}

func quote(s string) string {
	return "`" + s + "`"
}

var database = ""

func UseModel(db *gorm.DB, values ...any) {
	migrations = append(migrations, values...)
	if database == "" {
		// Use database-specific query for getting current database name
		dialectName := db.Dialector.Name()
		switch dialectName {
		case "mysql":
			db.Raw("SELECT DATABASE();").Scan(&database)
		case "postgres":
			db.Raw("SELECT current_database();").Scan(&database)
		case "sqlite":
			database = "main" // SQLite uses main as default database
		default:
			db.Raw("SELECT DATABASE();").Scan(&database) // fallback
		}
	}
	for index, _ := range values {
		ref := reflect.ValueOf(values[index])
		if ref.Kind() != reflect.Struct {
			return
		}
		var model = Model{
			Sample:      ref.Interface(),
			Value:       ref,
			Type:        reflect.TypeOf(ref.Interface()),
			Kind:        ref.Kind(),
			PackagePath: ref.Type().PkgPath(),
			Package:     filepath.Base(ref.Type().PkgPath()),
		}
		model.Name = model.Package + "." + ref.Type().Name()
		stmt := db.Model(values[index]).Statement
		stmt.Parse(values[index])
		model.Schema = stmt.Schema
		model.PrimaryKey = stmt.Schema.PrimaryFieldDBNames
		model.Statement = stmt
		model.Table = stmt.Table
		model.Joins = make(map[string][]string)

		var constraints []table.Constraint
		db.Where(table.Constraint{Database: database}).Find(&constraints)
		for _, constraint := range constraints {
			model.Joins[constraint.ReferencedTable] = []string{constraint.Column, constraint.ReferencedColumn}
		}

		Models = append(Models, model)
	}
}

func Find(v interface{}) *Model {
	if name, ok := v.(string); ok {
		for idx, _ := range Models {
			if Models[idx].Name == name || Models[idx].Table == name {
				return &Models[idx]
			}
		}
		return nil
	}
	var ref = reflect.ValueOf(v)
	if ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	for idx, _ := range Models {
		if Models[idx].Value.Type().Name() == ref.Type().Name() {
			return &Models[idx]
		}
	}
	return nil
}
