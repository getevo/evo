package dot

import (
	"errors"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/reflections"
	"github.com/getevo/evo/v2/lib/text"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"reflect"
	"regexp"
	"strings"
)

var arrayRegex = regexp.MustCompile(`\[[\"\'\x60]{0,1}(.*?)[\"\'\x60]{0,1}\]`)

func Get(obj any, prop string) (any, error) {
	// fmt.Println("getting property")
	// fmt.Println(args)

	// Get the array access
	chunks := text.SplitAny(prop, ".")
	var arr []string
	for _, item := range chunks {
		var matches = arrayRegex.FindAllStringSubmatch(item, -1)

		if len(matches) > 0 {
			arr = append(arr, strings.Split(item, "[")[0])

			for _, m := range matches {
				arr = append(arr, m[1])
			}
		} else {

			arr = append(arr, item)
		}
	}

	var err error
	// last, arr := arr[len(arr)-1], arr[:len(arr)-1]
	for _, key := range arr {

		obj, err = getProperty(obj, key)
		if err != nil {
			return nil, err
		}
		if obj == nil {
			return nil, nil
		}
	}
	return obj, nil
}

// Loop through this to get properties via dot notation
func getProperty(obj any, prop string) (any, error) {
	var _type = reflect.TypeOf(obj)
	if _type.Kind() == reflect.Array || _type.Kind() == reflect.Slice {
		val := reflect.ValueOf(obj)
		idx := val.Index(generic.Parse(prop).Int())
		if !idx.IsValid() {
			return nil, nil
		}
		return idx.Interface(), nil
	}
	if _type.Kind() == reflect.Map {
		val := reflect.ValueOf(obj)

		valueOf := val.MapIndex(reflect.ValueOf(prop))
		if !valueOf.IsValid() {
			return nil, nil
		}
		return valueOf.Interface(), nil
	}

	prop = cases.Title(language.English, cases.NoLower).String(prop)
	return reflections.GetField(obj, prop)
}

func Set(input any, prop string, value any) error {
	// Get the array access
	arr := strings.Split(prop, ".")
	var val = reflect.ValueOf(input)
	var obj reflect.Value
	if val.Kind() == reflect.Ptr {
		obj = val.Elem()
	} else {
		obj = val
	}
	last, arr := arr[len(arr)-1], arr[:len(arr)-1]

	for _, key := range arr {
		if obj.Kind() == reflect.Map {
			v := obj.MapIndex(reflect.ValueOf(key))
			if v.IsValid() {
				obj = v
			} else {
				var m = map[string]any{}
				obj.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(m))
				obj = obj.MapIndex(reflect.ValueOf(key))
			}

		} else {
			var ref, err = getProperty(obj.Interface(), key)
			if err != nil {
				return err
			}
			obj = reflect.ValueOf(ref)

		}

	}

	return setProperty(obj.Interface(), last, value)

}

func setProperty(obj any, prop string, val any) error {
	var ref = reflect.ValueOf(obj)
	if ref.Kind() == reflect.Map {
		ref.SetMapIndex(reflect.ValueOf(prop), reflect.ValueOf(val))
		return nil
	}

	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("object must be a pointer to a struct")
	}
	prop = cases.Title(language.English, cases.NoLower).String(prop)

	return reflections.SetField(obj, prop, val)
}
