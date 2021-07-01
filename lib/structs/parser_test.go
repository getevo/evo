package structs

import (
	"fmt"
	"testing"
)

var st = struct {
	X string `json:"test"`
	Y int
}{"abc", 20}

func TestFile(t *testing.T) {
	var s = New(st)
	//var s2 = New(&st)

	fmt.Println(s.Copy().FieldByName("X").Field.Tag.Get("json"))
	for _, f := range s.FieldsByTag("json") {
		fmt.Println(f.Value.Interface())
	}

	fmt.Println(s.Copy().Pointer())
}
