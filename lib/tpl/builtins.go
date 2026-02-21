package tpl

import (
	"encoding/json"
	"fmt"
	"github.com/kr/pretty"
	htmlpkg "html"
	"math"
	neturlpkg "net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var builtinTitleCaser = cases.Title(language.English, cases.NoLower)

func init() {
	registerBuiltins()
}

func registerBuiltins() {
	// ── String transforms ───────────────────────────────────────────────────────

	RegisterFunc("upper", strings.ToUpper)
	RegisterFunc("lower", strings.ToLower)
	RegisterFunc("title", builtinTitleCaser.String)
	RegisterFunc("trim", strings.TrimSpace)
	RegisterFunc("trimLeft", strings.TrimLeft)
	RegisterFunc("trimRight", strings.TrimRight)
	RegisterFunc("trimPrefix", strings.TrimPrefix)
	RegisterFunc("trimSuffix", strings.TrimSuffix)
	RegisterFunc("replace", strings.ReplaceAll)
	RegisterFunc("contains", strings.Contains)
	RegisterFunc("hasPrefix", strings.HasPrefix)
	RegisterFunc("hasSuffix", strings.HasSuffix)
	RegisterFunc("split", strings.Split)
	RegisterFunc("repeat", strings.Repeat)
	RegisterFunc("sprintf", fmt.Sprintf)

	RegisterFunc("join", func(elems []string, sep string) string {
		return strings.Join(elems, sep)
	})

	// joinAny joins any slice using reflection — useful when the slice is []any
	RegisterFunc("joinAny", func(v any, sep string) string {
		if v == nil {
			return ""
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.String {
			return rv.String()
		}
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return stringify(v)
		}
		parts := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts[i] = stringify(rv.Index(i).Interface())
		}
		return strings.Join(parts, sep)
	})

	// ── HTML / URL / JSON ───────────────────────────────────────────────────────

	RegisterFunc("html", htmlpkg.EscapeString)
	RegisterFunc("url", neturlpkg.QueryEscape)
	RegisterFunc("json", func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	})

	// ── Type conversion ─────────────────────────────────────────────────────────

	RegisterFunc("int", func(v any) int64 {
		n, _ := toInt64(v)
		return n
	})
	RegisterFunc("float", func(v any) float64 {
		f, _ := toFloat64(v)
		return f
	})
	RegisterFunc("str", func(v any) string {
		return stringify(v)
	})
	RegisterFunc("bool", func(v any) bool {
		return toBool(v)
	})

	// ── Math ────────────────────────────────────────────────────────────────────

	RegisterFunc("abs", func(v any) any {
		// Check float types first to preserve precision (toInt64 would truncate them).
		switch tv := v.(type) {
		case float64:
			return math.Abs(tv)
		case float32:
			return math.Abs(float64(tv))
		}
		if n, ok := toInt64(v); ok {
			if n < 0 {
				return -n
			}
			return n
		}
		if f, ok := toFloat64(v); ok {
			return math.Abs(f)
		}
		return v
	})
	RegisterFunc("floor", func(v any) int64 {
		f, _ := toFloat64(v)
		return int64(math.Floor(f))
	})
	RegisterFunc("ceil", func(v any) int64 {
		f, _ := toFloat64(v)
		return int64(math.Ceil(f))
	})
	RegisterFunc("round", func(v any) int64 {
		f, _ := toFloat64(v)
		return int64(math.Round(f))
	})
	RegisterFunc("sqrt", func(v any) float64 {
		f, _ := toFloat64(v)
		return math.Sqrt(f)
	})
	RegisterFunc("pow", func(base, exp any) float64 {
		b, _ := toFloat64(base)
		e, _ := toFloat64(exp)
		return math.Pow(b, e)
	})
	RegisterFunc("min", func(a, b any) any {
		af, aok := toFloat64(a)
		bf, bok := toFloat64(b)
		if !aok || !bok {
			return nil
		}
		if af <= bf {
			return a
		}
		return b
	})
	RegisterFunc("max", func(a, b any) any {
		af, aok := toFloat64(a)
		bf, bok := toFloat64(b)
		if !aok || !bok {
			return nil
		}
		if af >= bf {
			return a
		}
		return b
	})

	// ── Collections ─────────────────────────────────────────────────────────────

	RegisterFunc("len", func(v any) int {
		if v == nil {
			return 0
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
			return rv.Len()
		}
		return 0
	})

	// count is an alias for len — returns the number of elements/characters.
	RegisterFunc("count", func(v any) int {
		if v == nil {
			return 0
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
			return rv.Len()
		}
		return 0
	})

	RegisterFunc("keys", func(v any) []string {
		if v == nil {
			return nil
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Map {
			return nil
		}
		result := make([]string, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			result = append(result, fmt.Sprint(k.Interface()))
		}
		sort.Strings(result)
		return result
	})

	RegisterFunc("values", func(v any) []any {
		if v == nil {
			return nil
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Map {
			return nil
		}
		keys := rv.MapKeys()
		result := make([]any, len(keys))
		for i, k := range keys {
			result[i] = rv.MapIndex(k).Interface()
		}
		return result
	})

	RegisterFunc("first", func(v any) any {
		if v == nil {
			return nil
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			if rv.Len() == 0 {
				return nil
			}
			return rv.Index(0).Interface()
		case reflect.String:
			s := rv.String()
			if s == "" {
				return nil
			}
			return string([]rune(s)[0])
		}
		return nil
	})

	RegisterFunc("last", func(v any) any {
		if v == nil {
			return nil
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			if rv.Len() == 0 {
				return nil
			}
			return rv.Index(rv.Len() - 1).Interface()
		case reflect.String:
			runes := []rune(rv.String())
			if len(runes) == 0 {
				return nil
			}
			return string(runes[len(runes)-1])
		}
		return nil
	})

	RegisterFunc("slice", func(v any, start, end int) any {
		if v == nil {
			return nil
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			l := rv.Len()
			if start < 0 {
				start = 0
			}
			if end > l {
				end = l
			}
			if start >= end {
				return rv.Slice(0, 0).Interface()
			}
			return rv.Slice(start, end).Interface()
		case reflect.String:
			runes := []rune(rv.String())
			l := len(runes)
			if start < 0 {
				start = 0
			}
			if end > l {
				end = l
			}
			if start >= end {
				return ""
			}
			return string(runes[start:end])
		}
		return nil
	})

	// ── Logical / misc ──────────────────────────────────────────────────────────

	// default(val, fallback) — returns val if truthy, else fallback
	RegisterFunc("default", func(v, fallback any) any {
		if toBool(v) {
			return v
		}
		return fallback
	})

	// not(v) — boolean negation
	RegisterFunc("not", func(v any) bool {
		return !toBool(v)
	})

	// coalesce(a, b, c, ...) — returns first non-nil value
	RegisterFunc("coalesce", func(args ...any) any {
		for _, a := range args {
			if a != nil {
				return a
			}
		}
		return nil
	})

	// defined(v) — returns true if v is non-nil (alias for isset as a function)
	RegisterFunc("defined", func(v any) bool {
		return v != nil
	})

	// ternary(cond, then, else) — functional ternary for contexts where ?: is unavailable
	RegisterFunc("ternary", func(cond, then, els any) any {
		if toBool(cond) {
			return then
		}
		return els
	})

	// ── Date / time ─────────────────────────────────────────────────────────────

	// date(format, value) — formats a time value using a Go time layout string.
	// value may be a time.Time, a Unix timestamp (int/float), or a string in
	// RFC3339 / "2006-01-02 15:04:05" / "2006-01-02" format.
	// Example: date("2006-01-02", $createdAt)
	RegisterFunc("date", func(format string, v any) string {
		var t time.Time
		switch tv := v.(type) {
		case time.Time:
			t = tv
		case *time.Time:
			if tv != nil {
				t = *tv
			}
		case int64:
			t = time.Unix(tv, 0)
		case int:
			t = time.Unix(int64(tv), 0)
		case float64:
			t = time.Unix(int64(tv), 0)
		case string:
			for _, layout := range []string{
				time.RFC3339, time.RFC3339Nano,
				"2006-01-02 15:04:05", "2006-01-02",
			} {
				if parsed, err := time.Parse(layout, tv); err == nil {
					t = parsed
					break
				}
			}
		}
		if t.IsZero() {
			return ""
		}
		return t.Format(format)
	})

	// now() — returns current time as time.Time (useful with date())
	RegisterFunc("now", func() time.Time {
		return time.Now()
	})

	// ── Debugging ────────────────────────────────────────────────────────────────

	// dump(v) — returns a Go-style type+value representation for debugging.
	RegisterFunc("dump", func(v any) string {
		return pretty.Sprint(v)
	})
}
