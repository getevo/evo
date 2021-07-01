package jparser

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/getevo/evo"
	"github.com/getevo/evo/lib/log"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"reflect"
)

func RegisterDBPlugin() {
	rowCallback := evo.GetDBO().Callback().Query()
	rowCallback.Register("gorm:test", func(db *gorm.DB) {
		if db.Statement.Dest != nil {
			var t = reflect.TypeOf(db.Statement.Dest)
			if t.Kind() != reflect.Ptr {
				log.Error("parser only accept pointer")
				return
			}
			var v = reflect.ValueOf(db.Statement.Dest).Elem()
			if v.Kind() == reflect.Slice {
				ParseSlice(db.Statement.Dest)
			} else if v.Kind() == reflect.Struct {
				ParseStruct(db.Statement.Dest)
			}
		}
	})
}
func Parse(input interface{}) error {
	var t = reflect.TypeOf(input)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("parser only accept pointer")
	}
	var v = reflect.ValueOf(input).Elem()
	if v.Kind() == reflect.Slice {
		return ParseSlice(input)
	} else if v.Kind() == reflect.Struct {
		return ParseStruct(input)
	}
	return nil
}
func ParseSlice(input interface{}) error {
	var v = reflect.ValueOf(input).Elem()
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			var item = v.Index(i)
			if item.Kind() == reflect.Struct {
				return ParseStruct(item.Addr().Interface())
			}
		}
	}

	return nil
}

var gjsont = reflect.ValueOf(gjson.Result{}).Type()

func ParseStruct(v interface{}) error {
	st := structs.New(v)
	for _, field := range st.Fields() {
		if field.Tag("src") != "" {
			src := st.Field(field.Tag("src")).Value()
			if str, ok := src.(string); ok {
				var fi = reflect.ValueOf(field.Value())
				if fi.Type() == gjsont {
					field.Set(gjson.Parse(str))
				} else {
					var ptr = reflect.Indirect(reflect.New(fi.Type())).Addr().Interface()
					var err = json.Unmarshal([]byte(str), ptr)
					if err != nil {
						return err
					}
					field.Set(reflect.ValueOf(ptr).Elem().Interface())
				}

			}
		}
	}
	return nil
}
