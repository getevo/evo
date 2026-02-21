package evo

import (
	"github.com/kr/pretty"
)

func Repr(v any) string {
	return pretty.Sprint(v)
}

func Dump(v any) {
	_, _ = pretty.Println(v)
}
