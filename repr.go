package evo

import "github.com/alecthomas/repr"

func Repr(v any) string {
	return repr.String(v)
}

func Dump(v any) {
	repr.Println(v)
}
