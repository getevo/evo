package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type StringArray []string

func (o *StringArray) Scan(src any) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("src value cannot cast to []byte")
	}
	json.Unmarshal(bytes, o)
	return nil
}

func (o StringArray) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(o)
	return string(b), err
}

func (StringArray) GormDataType() string {
	return "text"
}

func (StringArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return "text"
}

type IntArray []string

func (o *IntArray) Scan(src any) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("src value cannot cast to []byte")
	}
	json.Unmarshal(bytes, o)
	return nil
}

func (o IntArray) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(o)
	return string(b), err
}

func (IntArray) GormDataType() string {
	return "text"
}

func (IntArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return "text"
}

type Int64Array []string

func (o *Int64Array) Scan(src any) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("src value cannot cast to []byte")
	}
	json.Unmarshal(bytes, o)
	return nil
}

func (o Int64Array) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(o)
	return string(b), err
}

func (Int64Array) GormDataType() string {
	return "text"
}

func (Int64Array) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return "text"
}
