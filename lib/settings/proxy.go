package settings

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getevo/evo/v2/lib/generic"
)

var fastAccess = map[string]generic.Value{}
var tracker = map[string]func(){}

type proxy struct {
	instance map[string]Interface
}

func (config *proxy) Name() string {
	return drivers[len(drivers)-1].Name()
}

func (config *proxy) Get(key string) generic.Value {
	if v, ok := fastAccess[key]; ok {
		return v
	}
	for _, instance := range drivers {
		ok, v := instance.Has(key)
		if ok {
			fastAccess[key] = v
			return v
		}
	}
	return generic.Parse("")
}
func (config *proxy) Has(key string) (bool, generic.Value) {
	if v, ok := fastAccess[key]; ok {
		return ok, v
	}
	for _, instance := range drivers {
		ok, v := instance.Has(key)
		if ok {
			fastAccess[key] = v
			return ok, v
		}
	}
	return false, generic.Parse("")
}
func (config *proxy) All() map[string]generic.Value {
	return drivers[len(drivers)-1].All()
}
func (config *proxy) Set(key string, value any) error {
	drivers[len(drivers)-1].Set(key, value)
	delete(fastAccess, key)
	if v, ok := tracker[key]; ok {
		v()
	}
	return nil
}
func (config *proxy) SetMulti(data map[string]any) error {
	drivers[len(drivers)-1].SetMulti(data)
	for key, _ := range data {
		if v, ok := tracker[key]; ok {
			v()
		}
	}
	return nil
}
func (config *proxy) Register(settings ...any) error {
	if len(settings) > 0 {
		if _, ok := settings[0].(Setting); ok {
			return drivers[len(drivers)-1].Register(settings...)
		} else if _, ok := settings[0].(SettingDomain); ok {
			return drivers[len(drivers)-1].Register(settings...)
		} else {
			var pkg = ""
			var set []any
			for _, setting := range settings {
				var s = generic.Parse(setting)
				var ref = s.Indirect()
				if ref.Kind() == reflect.String {
					pkg = fmt.Sprint(ref.Interface())
				} else if ref.Kind() == reflect.Struct {
					if pkg == "" {
						pkg = strings.Split(ref.Type().String(), ".")[0]
					}
					var typ = s.IndirectType()
					for i := 0; i < typ.NumField(); i++ {
						field := typ.Field(i)
						var readonly, _ = field.Tag.Lookup("readonly")
						var visible, _ = field.Tag.Lookup("visible")
						var value = field.Tag.Get("default")
						if value == "" {
							value = s.Prop(field.Name).String()
						}
						var n = Setting{
							Domain:      pkg,
							Name:        field.Name,
							Description: field.Tag.Get("description"),
							Value:       value,
							Params:      field.Tag.Get("params"),
							ReadOnly:    readonly == "false",
							Visible:     visible == "true",
							Type:        guessType(field.Type),
						}
						set = append(set, n)
						v := Get(pkg + "." + field.Name)
						s.SetProp(field.Name, v)
						tracker[pkg+"."+field.Name] = func() {
							s.SetProp(field.Name, Get(pkg+"."+field.Name))
						}
					}
				} else if ref.Kind() == reflect.Map {
					if pkg == "" {
						pkg = "APP"
					}
					for _, prop := range s.Props() {

						var n = Setting{
							Domain:      pkg,
							Name:        prop.Name,
							Description: "",
							Value:       s.Prop(prop.Name).String(),
							Params:      "",
							ReadOnly:    false,
							Visible:     true,
							Type:        "text",
						}
						var value = Get(pkg + "." + prop.Name).String()
						s.SetProp(prop.Name, value)
						tracker[pkg+"."+prop.Name] = func() {
							s.SetProp(prop.Name, Get(pkg+"."+prop.Name))
						}
						set = append(set, n)
					}
				}
			}
			drivers[len(drivers)-1].Register(set...)
		}
	}
	return nil
}

func guessType(t reflect.Type) string {

	switch t.String() {
	case "bool":
		return "switch"
	case "string":
		return "text"
	case "time.Duration":
		return "duration"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "number"

	default:
		return "text"
	}
}
func (config *proxy) Init(params ...string) error {

	return drivers[len(drivers)-1].Init(params...)
}
