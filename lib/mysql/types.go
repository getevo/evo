package mysql

import "reflect"

// --- Remote introspection types (MySQL information_schema) ---

type remoteTable struct {
	Database      string        `json:"database" gorm:"column:TABLE_SCHEMA"`
	Table         string        `json:"table" gorm:"column:TABLE_NAME"`
	Type          string        `json:"type" gorm:"column:TABLE_TYPE"`
	Engine        string        `json:"engine" gorm:"column:ENGINE"`
	RowFormat     string        `json:"row_format" gorm:"column:ROW_FORMAT"`
	Rows          int           `json:"rows" gorm:"column:TABLE_ROWS"`
	AutoIncrement uint64        `json:"auto_increment" gorm:"column:AUTO_INCREMENT"`
	Collation     string        `json:"collation" gorm:"column:TABLE_COLLATION"`
	Charset       string        `json:"charset" gorm:"column:TABLE_CHARSET"`
	Columns       remoteColumns `json:"columns" gorm:"-"`
	Indexes       remoteIndexes `json:"indexes" gorm:"-"`
	Constraints   []remoteConstraint `json:"constraints" gorm:"-"`
	Model         any           `json:"-" gorm:"-"`
	Reflect       reflect.Value `json:"-" gorm:"-"`
	PrimaryKey    []remoteColumn `json:"primary_key" gorm:"-"`
}

type remoteTables []remoteTable

func (t remoteTables) GetTable(table string) *remoteTable {
	for idx := range t {
		if t[idx].Table == table {
			return &t[idx]
		}
	}
	return nil
}

type remoteColumn struct {
	Database        string  `json:"database" gorm:"column:TABLE_SCHEMA"`
	Table           string  `json:"table" gorm:"column:TABLE_NAME"`
	Name            string  `json:"name" gorm:"column:COLUMN_NAME"`
	OrdinalPosition int     `json:"ordinal_position" gorm:"column:ORDINAL_POSITION"`
	ColumnDefault   *string `json:"column_default" gorm:"column:COLUMN_DEFAULT"`
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

type remoteColumns []remoteColumn

func (t remoteColumns) GetColumn(column string) *remoteColumn {
	for idx := range t {
		if t[idx].Name == column {
			return &t[idx]
		}
	}
	return nil
}

func (t remoteColumns) Keys() []string {
	var result []string
	for idx := range t {
		result = append(result, t[idx].Name)
	}
	return result
}

type remoteConstraint struct {
	Name             string `gorm:"column:CONSTRAINT_NAME" json:"name"`
	Table            string `gorm:"column:TABLE_NAME" json:"table"`
	Column           string `gorm:"column:COLUMN_NAME" json:"column"`
	ReferencedTable  string `gorm:"column:REFERENCED_TABLE_NAME" json:"referenced_table"`
	ReferencedColumn string `gorm:"column:REFERENCED_COLUMN_NAME" json:"referenced_column"`
	Database         string `gorm:"column:REFERENCED_TABLE_SCHEMA" json:"database"`
}

type remoteIndexStat struct {
	Database   string `gorm:"column:TABLE_SCHEMA"`
	Table      string `gorm:"column:TABLE_NAME"`
	NonUnique  bool   `gorm:"column:NON_UNIQUE"`
	Name       string `gorm:"column:INDEX_NAME"`
	ColumnName string `gorm:"column:COLUMN_NAME"`
}

type remoteIndex struct {
	Name    string
	Table   string
	Unique  bool
	Columns remoteColumns
}

type remoteIndexes []remoteIndex

func (list remoteIndexes) Find(name string) *remoteIndex {
	for idx := range list {
		if list[idx].Name == name {
			return &list[idx]
		}
	}
	return nil
}
