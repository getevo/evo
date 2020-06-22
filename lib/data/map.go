package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

// TODO: Merge with concurrent map

// Map gorm compatible dynamic map
type Map map[string]interface{}

func (p Map) Get(v string) interface{} {
	return p[v]
}

func (p Map) Has(v string) (interface{}, bool) {
	if v, ok := p[v]; ok {
		return v, ok
	}
	return nil, false
}

func (p Map) ToStruct(v interface{}) {
	mapstructure.Decode(p, v)
}

// Value return json value to store by gorm
func (p Map) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan parse json from gorm to map
func (p *Map) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Type assertion .([]byte) failed.")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Type assertion .(map[string]interface{}) failed.")
	}

	return nil
}
