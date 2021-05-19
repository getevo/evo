package evo

import "github.com/alecthomas/repr"

func Repr(v interface{}) string {
	return repr.String(v)
}

func Dump(v interface{}) {
	repr.Println(v)
}
