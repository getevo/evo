package structs

import (
	"reflect"
)

type Ref struct {
	Input interface{}
	Type  reflect.Type
	Value reflect.Value
}

type Field struct {
	Value reflect.Value
	Field reflect.StructField
}

func (r Ref) Fields() []*Field {
	var fields = make([]*Field, r.Type.NumField(), r.Type.NumField())
	for i := 0; i < r.Type.NumField(); i++ {
		fields[i] = r.Field(i)
	}
	return fields
}

func (r Ref) Field(i int) *Field {
	var f = Field{}
	if i > r.Type.NumField() {
		return nil
	}
	f.Field = r.Type.Field(i)
	f.Value = r.Value.Field(i)
	return &f
}

func (r Ref) FieldByName(name string) *Field {
	var f = Field{}
	var exists = false
	f.Field, exists = r.Type.FieldByName(name)
	if !exists {
		return nil
	}
	f.Value = r.Value.FieldByName(name)
	return &f
}

func (r Ref) FieldsByTag(tag string) []*Field {
	var fields []*Field
	for i := 0; i < r.Type.NumField(); i++ {
		f := r.Field(i)
		if f != nil && f.Field.Tag.Get(tag) != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

func (r Ref) Copy() Ref {
	var instance = r.Instance()
	instance.Value.Set(r.Value)
	instance.Input = instance.Value.Interface()
	return instance
}

func (r Ref) Instance() Ref {
	var new = Ref{
		Value: reflect.Indirect(reflect.New(r.Type)),
		Type:  r.Type,
	}
	new.Input = new.Value.Interface()
	return new
}

func (r Ref) Pointer() interface{} {
	return reflect.Indirect(r.Value).Interface()
}

func New(input interface{}) *Ref {
	var obj Ref
	var t = reflect.TypeOf(input)
	if t.Kind() == reflect.Ptr {
		obj = Ref{
			Input: input,
			Type:  reflect.TypeOf(input).Elem(),
			Value: reflect.ValueOf(input).Elem(),
		}
	} else {
		obj = Ref{
			Input: input,
			Type:  reflect.TypeOf(input),
			Value: reflect.ValueOf(input),
		}
	}

	return &obj
}
