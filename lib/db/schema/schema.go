package schema

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"path/filepath"
	"reflect"
)

var Models []Model

type Model struct {
	Sample      interface{}     `json:"sample"`
	Value       reflect.Value   `json:"-"`
	Type        reflect.Type    `json:"-"`
	Kind        reflect.Kind    `json:"-"`
	Table       string          `json:"table"`
	Name        string          `json:"name"`
	Package     string          `json:"package"`
	PackagePath string          `json:"package_path"`
	PrimaryKey  []string        `json:"primary_key"`
	Schema      *schema.Schema  `json:"-"`
	Statement   *gorm.Statement `json:"-"`
}

func (m Model) Join(joins ...*Model) ([]string, []string, error) {
	var where []string
	var tables = []string{m.Table}
	for _, join := range joins {
		tables = append(tables, join.Table)
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

func UseModel(db *gorm.DB, values ...interface{}) {
	migrations = append(migrations, values...)
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

		Models = append(Models, model)
	}
}

func Find(name string) *Model {
	for idx, _ := range Models {
		if Models[idx].Name == name || Models[idx].Table == name {
			return &Models[idx]
		}
	}
	return nil
}
