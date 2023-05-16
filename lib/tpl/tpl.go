package tpl

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/dot"
	"regexp"
)

var varRegex = regexp.MustCompile(`(?mi)\$([a-z\.\_\[\]0-9]+)*`)

func Render(src string, params ...interface{}) string {
	return varRegex.ReplaceAllStringFunc(src, func(s string) string {
		var key = s[1:]
		for _, item := range params {
			var v, err = dot.Get(item, key)
			if err == nil && v != nil {
				return fmt.Sprint(v)
			}
		}
		return s
	})
}
