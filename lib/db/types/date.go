package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/getevo/evo/v2/lib/date"
	"github.com/getevo/json"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Date struct {
	time.Time
}

// NewDate is a constructor for Date and returns new Date.
func NewDate(year, month, day int) Date {
	return Date{time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)}
}

// GormDataType returns gorm common data type. This type is used for the field's column type.
func (Date) GormDataType() string {
	return "date"
}

// GormDBDataType returns gorm DB data type based on the current using database.
func (Date) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "DATE"
	case "postgres":
		return "DATE"
	case "sqlserver":
		return "DATE"
	case "sqlite":
		return "DATE"
	default:
		return ""
	}
}

// Scan implements sql.Scanner interface and scans value into Time,
func (t *Date) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		t.setFromString(string(v))
	case string:
		t.setFromString(v)
	case time.Time:
		t.setFromTime(v)
	default:
		return errors.New(fmt.Sprintf("failed to scan value: %v", v))
	}

	return nil
}

func (t *Date) setFromString(str string) {
	v, _ := date.Parse(str)
	*t = Date{v.Base}
}

func (t *Date) setFromTime(src time.Time) {
	var d = NewDate(src.Year(), int(src.Month()), src.Day())
	*t = d
}

// Value implements driver.Valuer interface and returns string format of Time.
func (t Date) Value() (driver.Value, error) {
	return t.String(), nil
}

// String implements fmt.Stringer interface.
func (t Date) String() string {
	if t.Year() < 1000 {
		return "0000-00-00"
	}
	return fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

// MarshalJSON implements json.Marshaler to convert Time to json serialization.
func (t Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements json.Unmarshaler to deserialize json data.
func (t *Date) UnmarshalJSON(data []byte) error {
	// ignore null
	if string(data) == "null" {
		return nil
	}
	t.setFromString(strings.Trim(string(data), `"`))
	return nil
}
