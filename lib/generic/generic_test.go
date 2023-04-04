package generic

import (
	"fmt"
	"testing"
)

func TestIs(t *testing.T) {
	type Struct struct {
		Name string `json:"name"`
	}
	var x = Struct{
		Name: "test name",
	}

	var y = map[string]interface{}{
		"a": "b",
	}
	Parse(&x).SetProp("Name", "Reza")
	Parse(&y).SetProp("a", "c")
	fmt.Println(Parse(x).PropByTag("name"))
	fmt.Println(Parse(y).Prop("a"))
}
