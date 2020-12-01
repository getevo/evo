package main

import (
	"github.com/getevo/evo/lib/gpath"
	"testing"
)

func Test_FormatStruct(t *testing.T) {
	b, err := gpath.ReadFile("./test.go")
	if err != nil {
		panic(err)
	}

	code := formatStruct(string(b))
	_ = code
	f, err := gpath.Open("./test.go")
	if err != nil {
		panic(err)
	}

	f.WriteString(code)

}
