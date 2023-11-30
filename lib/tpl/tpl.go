package tpl

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/dot"
	"github.com/getevo/evo/v2/lib/generic"
	"reflect"
	"regexp"
)

var varRegex = regexp.MustCompile(`(?mi)\$([a-z\.\_\[\]0-9]+)*`)

func Render(src string, params ...any) string {
	return varRegex.ReplaceAllStringFunc(src, func(s string) string {
		var key = s[1:]
		for _, item := range params {
			var v, err = dot.Get(item, key)
			if err == nil && v != nil {
				obj := generic.Parse(v)

				if obj.IsAny(reflect.String, reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Int8, reflect.Float32, reflect.Float64) {
					return fmt.Sprint(v)
				}
				return obj.IndirectType().String()
			}
		}
		return s
	})
}
