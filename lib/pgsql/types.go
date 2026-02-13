package pgsql

import (
	"reflect"

	"github.com/getevo/evo/v2/lib/db/schema"
)

// pgRemoteTable represents a table retrieved from the database.
type pgRemoteTable struct {
	Database   string          `gorm:"column:table_schema"`
	Table      string          `gorm:"column:table_name"`
	Type       string          `gorm:"column:table_type"`
	Engine     string          `gorm:"column:engine"`
	Charset    string          `gorm:"column:table_charset"`
	Collation  string          `gorm:"column:table_collation"`
	Columns    pgRemoteColumns `gorm:"-"`
	Indexes    pgRemoteIndexes `gorm:"-"`
	Model      any             `gorm:"-"`
	Reflect    reflect.Value   `gorm:"-"`
	PrimaryKey pgRemoteColumns `gorm:"-"`
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

// pgRemoteColumn represents a column retrieved from the database.
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

// pgConstraint represents a foreign key constraint retrieved from the database.
type pgConstraint struct {
	Name             string `gorm:"column:constraint_name"`
	Table            string `gorm:"column:table_name"`
	Column           string `gorm:"column:column_name"`
	ReferencedTable  string `gorm:"column:referenced_table_name"`
	ReferencedColumn string `gorm:"column:referenced_column_name"`
}

// pgRemoteIndexStat represents an index statistic retrieved from the database.
type pgRemoteIndexStat struct {
	Database   string `gorm:"column:table_schema"`
	Table      string `gorm:"column:table_name"`
	NonUnique  bool   `gorm:"column:non_unique"`
	Name       string `gorm:"column:index_name"`
	ColumnName string `gorm:"column:column_name"`
}

// pgRemoteIndex represents an index retrieved from the database.
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

// pgDdlTable represents a local model definition used for DDL generation.
type pgDdlTable struct {
	Columns    schema.Columns
	PrimaryKey schema.Columns
	Index      schema.Indexes
	Name       string
}
