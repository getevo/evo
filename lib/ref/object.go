package ref

import (
	"fmt"
	"reflect"
	"strings"
)

type obj struct {
	Name      string
	Package   string
	Path      string
	IsPointer bool
	Fields    []reflect.StructField
	Methods   []reflect.Method
	internal  wrapedStruct
	ref       reflect.Type
}

func Parse(v interface{}) obj {

	item := obj{}
	item.ref = reflect.TypeOf(v)

	pkg := strings.Split(item.ref.String(), ".")[0]
	if pkg[0] == '*' {
		item.IsPointer = true
		v = reflect.ValueOf(v).Elem().Interface()
		item.ref = reflect.TypeOf(v)
		item.internal = wrapedStruct{v}
	}
	item.Name = item.ref.Name()
	item.Package = strings.TrimLeft(pkg, "*")
	item.Path = string(item.ref.PkgPath())

	indirect := reflect.Indirect(reflect.ValueOf(v))
	t := indirect.Type()
	item.Fields = make([]reflect.StructField, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		item.Fields[i] = t.Field(i)
	}

	item.Methods = make([]reflect.Method, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		item.Methods[i] = t.Method(i)
	}
	return item
}

type wrapedStruct struct {
	obj interface{}
}

func Wrap(v interface{}) wrapedStruct {
	return wrapedStruct{v}
}

// Invoke - firstResult, err := Invoke(AnyStructInterface, MethodName, Params...)
func (o wrapedStruct) Invoke(name string, args ...interface{}) (reflect.Value, error) {
	method := reflect.ValueOf(o.obj).MethodByName(name)
	methodType := method.Type()
	numIn := methodType.NumIn()
	if numIn > len(args) {
		return reflect.ValueOf(nil), fmt.Errorf("Method %s must have minimum %d params. Have %d", name, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return reflect.ValueOf(nil), fmt.Errorf("Method %s must have %d params. Have %d", name, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return reflect.ValueOf(nil), fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return reflect.ValueOf(nil), fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argType)
		}
	}
	return method.Call(in)[0], nil
}

func (o obj) Invoke(name string, args ...interface{}) (reflect.Value, error) {
	if !o.IsPointer {
		return reflect.Value{}, fmt.Errorf("object is not an pointer")
	}
	return o.internal.Invoke(name, args...)
}

func (o obj) Get(name string) (reflect.Value, error) {
	if !o.IsPointer {
		return reflect.Value{}, fmt.Errorf("object is not an pointer")
	}
	return o.internal.Get(name)
}

func (o obj) Set(key string, value interface{}) error {
	if !o.IsPointer {
		return fmt.Errorf("object is not an pointer")
	}
	return o.Set(key, value)
}

func (o wrapedStruct) Set(key string, value interface{}) error {
	ps := reflect.ValueOf(o.obj)
	s := ps.Elem()
	if s.Kind() == reflect.Struct {
		f := s.FieldByName(key)
		if f.IsValid() {
			if f.CanSet() {
				f.Set(reflect.ValueOf(value))
			} else {
				return fmt.Errorf("key is not writable")
			}
		} else {
			return fmt.Errorf("key is not exist")
		}
	} else {
		return fmt.Errorf("element is not struct")
	}
	return nil
}

func (o wrapedStruct) Get(key string) (reflect.Value, error) {
	ps := reflect.ValueOf(o.obj)
	v := reflect.Indirect(ps).FieldByName(key)
	if !v.IsValid() {
		return v, fmt.Errorf("%s is not member of %s", key, ps.Type())
	}
	return v, nil
}

// Invoke - firstResult, err := Invoke(AnyStructInterface, MethodName, Params...)
func Invoke(any interface{}, name string, args ...interface{}) (reflect.Value, error) {
	method := reflect.ValueOf(any).MethodByName(name)
	methodType := method.Type()
	numIn := methodType.NumIn()
	if numIn > len(args) {
		return reflect.ValueOf(nil), fmt.Errorf("Method %s must have minimum %d params. Have %d", name, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return reflect.ValueOf(nil), fmt.Errorf("Method %s must have %d params. Have %d", name, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return reflect.ValueOf(nil), fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return reflect.ValueOf(nil), fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argType)
		}
	}
	return method.Call(in)[0], nil
}
