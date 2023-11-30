package table

import (
	"reflect"
	"time"
)

type IndexStat struct {
	//model:skip
	Database   string `gorm:"column:TABLE_SCHEMA"`
	Table      string `gorm:"column:TABLE_NAME"`
	NonUnique  bool   `gorm:"column:NON_UNIQUE"`
	Name       string `gorm:"column:INDEX_NAME"`
	ColumnName string `gorm:"column:COLUMN_NAME"`
}

func (IndexStat) TableName() string {
	return "information_schema.statistics"
}

type Index struct {
	Name    string
	Table   string
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

type Table struct {
	//model:skip
	Database      string        `json:"database" gorm:"column:TABLE_SCHEMA"`
	Table         string        `json:"table" gorm:"column:TABLE_NAME"`
	Type          string        `json:"type" gorm:"column:TABLE_TYPE"`
	Engine        string        `json:"engine" gorm:"column:ENGINE"`
	RowFormat     string        `json:"row_format" gorm:"column:ROW_FORMAT"`
	Rows          int           `json:"rows" gorm:"column:TABLE_ROWS"`
	AutoIncrement int           `json:"auto_increment" gorm:"column:AUTO_INCREMENT"`
	Collation     string        `json:"collation" gorm:"column:TABLE_COLLATION"`
	Charset       string        `json:"charset" gorm:"column:TABLE_CHARSET"`
	Columns       Columns       `json:"columns" gorm:"-"`
	Indexes       Indexes       `json:"indexes" gorm:"-"`
	Model         any           `json:"-" gorm:"-"`
	Reflect       reflect.Value `json:"-" gorm:"-"`
}

func (Table) TableName() string {
	return "information_schema.TABLES"
}

type Column struct {
	//model:skip
	Database        string  `json:"database" gorm:"column:TABLE_SCHEMA"`
	Table           string  `json:"table" gorm:"column:TABLE_NAME"`
	Name            string  `json:"name"  gorm:"column:COLUMN_NAME"`
	OrdinalPosition int     `json:"ordinal_position" gorm:"column:ORDINAL_POSITION"`
	ColumnDefault   *string `json:"column_default"  gorm:"column:COLUMN_DEFAULT"`
	Nullable        string  `json:"nullable" gorm:"column:IS_NULLABLE"`
	DataType        string  `json:"data_type" gorm:"column:DATA_TYPE"`
	ColumnType      string  `json:"column_type" gorm:"column:COLUMN_TYPE"`
	Size            *int    `json:"size" gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
	Precision       *int    `json:"precision" gorm:"column:NUMERIC_PRECISION"`
	Scale           *int    `json:"scale" gorm:"column:NUMERIC_SCALE"`
	DatePrecision   *int    `json:"date_precision" gorm:"column:DATETIME_PRECISION"`
	CharacterSet    string  `json:"character_set" gorm:"column:CHARACTER_SET_NAME"`
	Collation       string  `json:"collation" gorm:"column:COLLATION_NAME"`
	ColumnKey       string  `json:"column_key" gorm:"column:COLUMN_KEY"`
	Extra           string  `json:"extra" gorm:"column:EXTRA"`
	Comment         string  `json:"comment" gorm:"column:COLUMN_COMMENT"`
}

func (Column) TableName() string {
	return "information_schema.COLUMNS"
}

type Tables []Table

func (t Tables) GetTable(table string) *Table {
	for idx, _ := range t {
		if t[idx].Table == table {
			return &t[idx]
		}
	}
	return nil
}

type Columns []Column

func (t Columns) GetColumn(column string) *Column {
	for idx, _ := range t {
		if t[idx].Name == column {
			return &t[idx]
		}
	}
	return nil
}

func (t Columns) Keys() []string {
	var result []string
	for idx, _ := range t {
		result = append(result, t[idx].Name)
	}
	return result
}

type TableVersion struct {
	Caller      string `gorm:"column:caller;primaryKey;size:64" json:"caller"`
	Version     string `gorm:"column:version;primaryKey;size:64" json:"version"`
	Query       string `gorm:"column:query" json:"query"`
	Outcome     string `gorm:"column:outcome" json:"outcome"`
	Description string `gorm:"column:description"  json:"description"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (TableVersion) TableName() string {
	return "table_version"
}
