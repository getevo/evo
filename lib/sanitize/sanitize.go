package sanitize

import (
	"reflect"
)

func Generic(v interface{}, fn ...func(str string) string) {
	if len(fn) == 0 {
		fn = []func(str string) string{String}
	}

	s := reflect.ValueOf(v)
	if s.Type().String() == "reflect.Value" {
		s = v.(reflect.Value)
	}
	if s.Kind() != reflect.Ptr {
		return
	}
	s = s.Elem()
	switch s.Kind() {
	case reflect.Struct:
		_struct(s, fn[0])
		break
	case reflect.String:
		s.SetString(fn[0](s.String()))
		break
	case reflect.Map:
		_map(s, fn[0])
		break
	case reflect.Slice, reflect.Array:
		_slice(s, fn[0])
		break
	case reflect.Interface:
		_interface(s, v, fn[0])

	}
}

func _interface(v reflect.Value, _interface interface{}, fn func(str string) string) {

	switch v.Elem().Kind() {
	case reflect.String:
		var new interface{}
		new = fn(v.Interface().(string))
		v.Set(reflect.ValueOf(new))
		break
	case reflect.Ptr:
		Generic(v.Interface(), fn)
		break
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Interface:
		T := v.Elem().Type()
		cpy := reflect.New(T)
		cpy.Elem().Set(v.Elem())
		Generic(cpy.Interface(), fn)
		v.Set(cpy.Elem())
		break
	}
}

func _struct(v reflect.Value, fn func(str string) string, params ...interface{}) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			field.SetString(fn(field.String()))
			break
		case reflect.Ptr:
			Generic(field.Interface(), fn)
			break
		case reflect.Struct:
			_struct(field, fn)
			break
		case reflect.Map:
			_map(field, fn)
			break
		case reflect.Slice, reflect.Array:
			_slice(field, fn)
			break
		case reflect.Interface:
			_interface(field, field.Interface(), fn)
			break

		}
	}
}

func _map(v reflect.Value, fn func(str string) string) {

	for _, key := range v.MapKeys() {
		item := v.MapIndex(key)
		switch item.Kind() {
		case reflect.String:
			v.SetMapIndex(key, reflect.ValueOf(fn(item.String())))
			break
		case reflect.Ptr:
			Generic(item.Interface(), fn)
			break
		case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Interface:
			cpy := reflect.New(item.Type())
			cpy.Elem().Set(item)
			Generic(cpy.Interface(), fn)
			v.SetMapIndex(key, cpy.Elem())
			break
		}
	}
}

func _slice(v reflect.Value, fn func(str string) string) {
	if v.Len() == 0 {
		return
	}
	slice := reflect.MakeSlice(v.Type(), 0, v.Len())
	ptr := reflect.New(slice.Type()).Elem()
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		cpy := reflect.New(item.Type())
		cpy.Elem().Set(item)
		Generic(cpy, fn)
		ptr.Set(reflect.Append(ptr, cpy.Elem()))
	}
	v.Set(ptr)
}

func String(str string) string {
	var tmpRune []rune
	strRune := []rune(str)
	for _, ch := range strRune {
		switch ch {
		case []rune{'\\'}[0], []rune{'"'}[0], []rune{'\''}[0]:
			tmpRune = append(tmpRune, []rune{'\\'}[0])
			tmpRune = append(tmpRune, ch)
		default:
			tmpRune = append(tmpRune, ch)
		}
	}
	return string(tmpRune)
}
