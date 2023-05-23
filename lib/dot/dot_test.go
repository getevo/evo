package dot_test

import (
	"github.com/getevo/evo/v2/lib/dot"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	// Test case 1: Accessing a valid property in a map
	obj := map[string]interface{}{
		"foo": "bar",
	}
	result, err := dot.Get(obj, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", result)

	// Test case 2: Accessing a nested property in a struct
	type InnerStruct struct {
		InnerProp string
	}
	type MyStruct struct {
		Nested InnerStruct
	}
	obj2 := MyStruct{
		Nested: InnerStruct{
			InnerProp: "value",
		},
	}
	result, err = dot.Get(obj2, "Nested.InnerProp")
	assert.NoError(t, err)
	assert.Equal(t, "value", result)

	// Add more test cases for different scenarios
}

func TestSet(t *testing.T) {
	// Test case 1: Setting a property in a map
	/*	obj := map[string]interface{}{
			"foo": "bar",
		}
		err := dot.Set(&obj, "foo", "new value")
		assert.NoError(t, err)
		assert.Equal(t, "new value", obj["foo"])*/

	// Test case 2: Setting a nested property in a struct
	type InnerStruct struct {
		InnerProp string
	}
	type MyStruct struct {
		Nested InnerStruct
	}
	obj2 := MyStruct{
		Nested: InnerStruct{
			InnerProp: "old value",
		},
	}
	err := dot.Set(&obj2, "Nested.InnerProp", "new value")
	assert.NoError(t, err)
	assert.Equal(t, "new value", obj2.Nested.InnerProp)

	// Add more test cases for different scenarios
}
