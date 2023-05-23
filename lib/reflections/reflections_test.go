package reflections_test

import (
	"github.com/getevo/evo/v2/lib/reflections"
	"reflect"
	"testing"
)

type exampleStruct struct {
	PublicField    string
	privateField   string
	TaggedField    string `customTag:"tagValue"`
	EmbeddedStruct struct {
		EmbeddedField string
	}
}

func TestGetField(t *testing.T) {
	obj := exampleStruct{
		PublicField:  "public",
		privateField: "private",
	}
	expected := "public"

	value, err := reflections.GetField(obj, "PublicField")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if value != expected {
		t.Errorf("Expected: %v, but got: %v", expected, value)
	}
}

func TestGetFieldKind(t *testing.T) {
	obj := exampleStruct{
		PublicField:  "public",
		privateField: "private",
	}
	expected := reflect.String

	kind, err := reflections.GetFieldKind(obj, "PublicField")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if kind != expected {
		t.Errorf("Expected: %v, but got: %v", expected, kind)
	}
}

func TestGetFieldType(t *testing.T) {
	obj := exampleStruct{
		PublicField:  "public",
		privateField: "private",
	}
	expected := "string"

	fieldType, err := reflections.GetFieldType(obj, "PublicField")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if fieldType != expected {
		t.Errorf("Expected: %v, but got: %v", expected, fieldType)
	}
}

func TestGetFieldTag(t *testing.T) {
	obj := exampleStruct{
		TaggedField: "tagged",
	}
	expected := "tagValue"

	tagValue, err := reflections.GetFieldTag(obj, "TaggedField", "customTag")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if tagValue != expected {
		t.Errorf("Expected: %v, but got: %v", expected, tagValue)
	}
}

func TestGetFieldNameByTagValue(t *testing.T) {
	obj := exampleStruct{
		TaggedField: "tagged",
	}
	expected := "TaggedField"

	fieldName, err := reflections.GetFieldNameByTagValue(obj, "customTag", "tagValue")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if fieldName != expected {
		t.Errorf("Expected: %v, but got: %v", expected, fieldName)
	}
}

func TestSetField(t *testing.T) {
	obj := &exampleStruct{}
	expected := "testValue"

	err := reflections.SetField(obj, "PublicField", expected)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if obj.PublicField != expected {
		t.Errorf("Expected: %v, but got: %v", expected, obj.PublicField)
	}
}

func TestHasField(t *testing.T) {
	obj := exampleStruct{}
	existingFieldName := "PublicField"
	nonExistingFieldName := "NonExistingField"

	hasField, err := reflections.HasField(obj, existingFieldName)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if !hasField {
		t.Errorf("Expected field '%s' to exist", existingFieldName)
	}

	hasField, err = reflections.HasField(obj, nonExistingFieldName)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if hasField {
		t.Errorf("Expected field '%s' to not exist", nonExistingFieldName)
	}
}
