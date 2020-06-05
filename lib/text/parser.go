package text

import (
	"encoding/json"
	"regexp"
	"strings"
)

func ParseWildCard(input, expr string) []string {
	i := 0
	for _, item := range expr {
		if item == '*' {
			i++
		}
	}
	expr = strings.Replace(expr, "*", "(.+)", -1)
	r := regexp.MustCompile(expr)
	res := r.FindAllStringSubmatch(input, 1)
	if len(res) == 0 || len(res[0]) != i+1 {
		var empty []string
		for j := 0; j < i; j++ {
			empty = append(empty, "")
		}
		return empty
	}
	return res[0][1:]
}

func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func SplitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}
