package entity

import (
	"github.com/getevo/evo/v2/lib/application"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/db/schema/ddl"
	"github.com/getevo/evo/v2/lib/db/types"
	"github.com/getevo/restify"
	"gorm.io/gorm/clause"
	"strings"
)

type Type string

const (
	TypeNotAssigned Type = "NA"
	TypeAPI         Type = "API"
	TypeDictionary  Type = "dictionary"
)

type Entity struct {
	EntitySlug string                  `gorm:"column:entity_slug;size:32;primaryKey" json:"entity_slug"`
	Table      string                  `gorm:"column:table;size:32;unique"  json:"table"`
	Name       string                  `gorm:"column:name;size:64"  json:"name"`
	Package    string                  `gorm:"column:package;size:64" json:"package"`
	PrimaryKey types.JSONSlice[string] `gorm:"column:primary_key;type:varchar(512)" json:"primary_key"`
	Schema     schema.Model            `gorm:"-" json:"-"`
	Instance   interface{}             `gorm:"-" json:"-"`
}

func (Entity) TableName() string {
	return "entity"
}

type Field struct {
	FieldSlug  string                     `gorm:"column:field_slug;size:128;primaryKey" json:"field_slug"`
	PrimaryKey bool                       `gorm:"column:primary_key" json:"primary_key"`
	EntitySlug string                     `gorm:"column:entity_slug;size:32;fk:entity" json:"entity_slug"`
	Entity     Entity                     `gorm:"foreignKey:EntitySlug;references:EntitySlug" json:"entity"`
	FieldName  string                     `gorm:"column:field_name;size:64" json:"field_name"`
	JSONTag    string                     `gorm:"column:json_tag;size:64"  json:"json_tag"`
	DBField    string                     `gorm:"column:db_field;size:64" json:"db_field"`
	Type       string                     `gorm:"column:type;size:255" json:"type"`
	DBType     string                     `gorm:"column:db_type;size:64" json:"db_type"`
	DataSource types.JSONType[DataSource] `gorm:"column:data_source;nullable;type:text" json:"data_source"`
}

func (Field) TableName() string {
	return "entity_field"
}

type FieldOption struct {
	Key         string `gorm:"column:key;primaryKey;size:32" json:"key"`
	Label       string `gorm:"column:label;size:128" json:"label"`
	Description string `gorm:"column:description;size:512" json:"description"`
	Icon        string `gorm:"column:icon;size:64" json:"icon"`
	Image       string `gorm:"column:image;size:255" json:"image"`
	FieldSlug   string `gorm:"column:field_slug;size:128;fk:entity_field" json:"field_slug"`
	EntityField *Field `gorm:"foreignKey:FieldSlug" json:"entity_field,omitempty"`
	VisualOrder int    `gorm:"column:visual_order" json:"visual_order"`
	restify.API
}

func (FieldOption) TableName() string {
	return "entity_field_option"
}

type DataSource struct {
	Type       Type                        `json:"type"`
	URL        *string                     `json:"url,omitempty"`
	Dictionary *types.Dictionary[any, any] `json:"dictionary,omitempty"`
	Mapper     *Mapper                     `json:"mapper,omitempty"`
}

type Mapper struct {
	Iterator string            `json:"iterator"`
	Map      map[string]string `json:"map"`
}

type Option string

func (o Option) ColumnDefinition(column *ddl.Column) {
	column.ForeignKey = "entity_field_option"
	column.Type = "varchar(32)"
}

type App struct {
}

func (a App) Register() error {
	db.UseModel(Entity{}, Field{}, FieldOption{})
	return nil
}

func (a App) Router() error {
	return nil
}

func (a App) WhenReady() error {

	var entities []Entity
	var fields []Field
	for _, model := range schema.Models {
		var entity = Entity{
			EntitySlug: model.Name,
			Table:      model.Table,
			Package:    model.Package,
			Name:       strings.Split(model.Name, ".")[1],
		}

		var pk []string
		for _, field := range model.Schema.Fields {
			if field.DBName == "" {
				continue
			}
			var entityField = Field{
				FieldSlug:  model.Name + "." + field.Name,
				PrimaryKey: field.PrimaryKey,
				EntitySlug: model.Name,
				FieldName:  field.Name,
				DBField:    field.DBName,
				JSONTag:    strings.Split(field.Tag.Get("json"), ",")[0],
				Type:       field.FieldType.String(),
				DBType:     string(field.GORMDataType),
			}

			if field.PrimaryKey {
				pk = append(pk, field.DBName)
			}
			fields = append(fields, entityField)

		}
		entity.PrimaryKey = pk
		entities = append(entities, entity)
	}
	db.Save(&entities)
	db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"field_name", "primary_key", "db_field", "json_tag", "type", "db_type"}),
	}).Create(&fields)

	return nil
}

func (a App) Name() string {
	return "entity"
}

func (a App) Priority() application.Priority {
	return application.LOWEST
}
