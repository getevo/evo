package reflections

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrUnsupportedType indicates that the provided type doesn't support the requested reflection operation.
var ErrUnsupportedType = errors.New("unsupported type")

// ErrUnexportedField indicates that an operation failed as a result of
// applying to a non-exported struct field.
var ErrUnexportedField = errors.New("unexported field")

// GetField returns the value of the provided obj field.
// The `obj` can either be a structure or pointer to structure.
func GetField(obj any, name string) (any, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return nil, fmt.Errorf("cannot use GetField on a non-struct object: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	field := objValue.FieldByName(name)
	if !field.IsValid() {
		return nil, fmt.Errorf("no such field: %s in obj", name)
	}

	return field.Interface(), nil
}

// GetFieldKind returns the kind of the provided obj field.
// The `obj` can either be a structure or pointer to structure.
func GetFieldKind(obj any, name string) (reflect.Kind, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return reflect.Invalid, fmt.Errorf("cannot use GetFieldKind on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	field := objValue.FieldByName(name)

	if !field.IsValid() {
		return reflect.Invalid, fmt.Errorf("no such field: %s in obj", name)
	}

	return field.Type().Kind(), nil
}

// GetFieldType returns the kind of the provided obj field.
// The `obj` can either be a structure or pointer to structure.
func GetFieldType(obj any, name string) (string, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return "", fmt.Errorf("cannot use GetFieldType on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	field := objValue.FieldByName(name)

	if !field.IsValid() {
		return "", fmt.Errorf("no such field: %s in obj", name)
	}

	return field.Type().String(), nil
}

// GetFieldTag returns the provided obj field tag value.
// The `obj` parameter can either be a structure or pointer to structure.
func GetFieldTag(obj any, fieldName, tagKey string) (string, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return "", fmt.Errorf("cannot use GetFieldTag on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()

	field, ok := objType.FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("no such field: %s in obj", fieldName)
	}

	if !isExportableField(field) {
		return "", fmt.Errorf("cannot GetFieldTag on a non-exported struct field: %w", ErrUnexportedField)
	}

	return field.Tag.Get(tagKey), nil
}

// GetFieldNameByTagValue looks up a field with a matching `{tagKey}:"{tagValue}"` tag in the provided `obj` item.
// The `obj` parameter must be a `struct`, or a `pointer` to one. If the `obj` parameter doesn't have a field tagged
// with the `tagKey`, and the matching `tagValue`, this function returns an error.
func GetFieldNameByTagValue(obj any, tagKey, tagValue string) (string, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return "", fmt.Errorf("cannot use GetFieldByTag on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	for i := 0; i < fieldsCount; i++ {
		structField := objType.Field(i)
		if structField.Tag.Get(tagKey) == tagValue {
			return structField.Name, nil
		}
	}

	return "", errors.New("tag doesn't exist in the given struct")
}

// SetField sets the provided obj field with provided value.
//
// The `obj` parameter must be a pointer to a struct, otherwise it soundly fails.
// The provided `value` type should match with the struct field being set.
func SetField(obj any, name string, value any) error {
	// Fetch the field reflect.Value
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		invalidTypeError := fmt.Errorf("provided value type didn't match obj field type")
		return invalidTypeError
	}

	structFieldValue.Set(val)
	return nil
}

// HasField checks if the provided `obj` struct has field named `name`.
// The `obj` can either be a structure or pointer to structure.
func HasField(obj any, name string) (bool, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return false, fmt.Errorf("cannot use HasField on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	field, ok := objType.FieldByName(name)
	if !ok || !isExportableField(field) {
		return false, nil
	}

	return true, nil
}

// Fields returns the struct fields names list.
// The `obj` parameter can either be a structure or pointer to structure.
func Fields(obj any) ([]string, error) {
	return fields(obj, false)
}

// FieldsDeep returns "flattened" fields.
//
// Note that FieldsDeept treats fields from anonymous inner structs as normal fields.
func FieldsDeep(obj any) ([]string, error) {
	return fields(obj, true)
}

func fields(obj any, deep bool) ([]string, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return nil, fmt.Errorf("cannot use fields on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	var allFields []string
	for i := 0; i < fieldsCount; i++ {
		field := objType.Field(i)
		if isExportableField(field) {
			if !deep || !field.Anonymous {
				allFields = append(allFields, field.Name)
				continue
			}

			fieldValue := objValue.Field(i)
			subFields, err := fields(fieldValue.Interface(), deep)
			if err != nil {
				return nil, fmt.Errorf("cannot get fields in %s: %w", field.Name, err)
			}
			allFields = append(allFields, subFields...)
		}
	}

	return allFields, nil
}

// Items returns the field:value struct pairs as a map.
// The `obj` parameter can either be a structure or pointer to structure.
func Items(obj any) (map[string]any, error) {
	return items(obj, false)
}

// ItemsDeep returns "flattened" items.
// Note that ItemsDeep will treat fields from anonymous inner structs as normal fields.
func ItemsDeep(obj any) (map[string]any, error) {
	return items(obj, true)
}

func items(obj any, deep bool) (map[string]any, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return nil, fmt.Errorf("cannot use items on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	allItems := make(map[string]any)

	for i := 0; i < fieldsCount; i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i)

		if isExportableField(field) {
			if !deep || !field.Anonymous {
				allItems[field.Name] = fieldValue.Interface()
				continue
			}

			m, err := items(fieldValue.Interface(), deep)
			if err != nil {
				return nil, fmt.Errorf("cannot get items in %s: %w", field.Name, err)
			}

			for k, v := range m {
				allItems[k] = v
			}
		}
	}

	return allItems, nil
}

// Tags lists the struct tag fields.
// The `obj` can whether be a structure or pointer to structure.
func Tags(obj any, key string) (map[string]string, error) {
	return tags(obj, key, false)
}

// TagsDeep returns "flattened" tags.
// Note that TagsDeep treats fields from anonymous
// inner structs as normal fields.
func TagsDeep(obj any, key string) (map[string]string, error) {
	return tags(obj, key, true)
}

func tags(obj any, key string, deep bool) (map[string]string, error) {
	if !isSupportedType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return nil, fmt.Errorf("cannot use tags on a non-struct interface: %w", ErrUnsupportedType)
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()

	allTags := make(map[string]string)

	for i := 0; i < fieldsCount; i++ {
		structField := objType.Field(i)
		if isExportableField(structField) {
			if !deep || !structField.Anonymous {
				allTags[structField.Name] = structField.Tag.Get(key)
				continue
			}

			fieldValue := objValue.Field(i)
			m, err := tags(fieldValue.Interface(), key, deep)
			if err != nil {
				return nil, fmt.Errorf("cannot get items in %s: %w", structField.Name, err)
			}

			for k, v := range m {
				allTags[k] = v
			}
		}
	}

	return allTags, nil
}

func reflectValue(obj any) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	return val
}

func isExportableField(field reflect.StructField) bool {
	// PkgPath is empty for exported fields.
	return field.PkgPath == ""
}

func isSupportedType(obj any, types []reflect.Kind) bool {
	for _, t := range types {
		if reflect.TypeOf(obj).Kind() == t {
			return true
		}
	}

	return false
}
