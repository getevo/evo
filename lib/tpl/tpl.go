package tpl

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/getevo/evo/v2/lib/dot"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const defaultCacheLimit = 1000

var (
	cacheLimit = defaultCacheLimit
	cacheMu    sync.RWMutex
	cache      = make(map[string]*Template, defaultCacheLimit)
)

// SetCacheSize sets the maximum number of compiled templates to cache.
// Pass 0 to disable caching. If n is smaller than the current cache size,
// the cache is cleared.
func SetCacheSize(n int) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cacheLimit = n
	if n == 0 || len(cache) > n {
		cache = make(map[string]*Template, n)
	}
}

// segment is one piece of a compiled template: either a literal string or a
// variable reference (with optional modifier).
type segment struct {
	literal  string
	path     string
	modifier string // transform name ("upper","lower","title") or default value
	isVar    bool
}

// Template is a pre-compiled template ready for repeated execution.
type Template struct {
	segments []segment
}

// Parse compiles src into a Template and caches the result up to the configured
// cache size (default 1000). When the cache is full the result is returned but
// not stored. Setting cache size to 0 disables caching entirely.
func Parse(src string) *Template {
	if cacheLimit > 0 {
		cacheMu.RLock()
		if t, ok := cache[src]; ok {
			cacheMu.RUnlock()
			return t
		}
		cacheMu.RUnlock()
	}

	t := compile(src)

	if cacheLimit > 0 {
		cacheMu.Lock()
		if _, ok := cache[src]; !ok && len(cache) < cacheLimit {
			cache[src] = t
		}
		cacheMu.Unlock()
	}

	return t
}

// compile parses src into a Template without touching the cache.
func compile(src string) *Template {
	t := &Template{}
	rem := src
	for len(rem) > 0 {
		i := strings.IndexByte(rem, '$')
		if i < 0 {
			t.segments = append(t.segments, segment{literal: rem})
			break
		}
		if i > 0 {
			t.segments = append(t.segments, segment{literal: rem[:i]})
		}
		rem = rem[i+1:] // skip '$'

		if len(rem) == 0 {
			// trailing bare '$'
			t.segments = append(t.segments, segment{literal: "$"})
			break
		}

		if rem[0] == '$' {
			// '$$' → literal '$'
			t.segments = append(t.segments, segment{literal: "$"})
			rem = rem[1:]
			continue
		}

		r, _ := utf8.DecodeRuneInString(rem)
		if !isIdentStart(r) {
			// '$' followed by non-identifier char → literal '$'
			t.segments = append(t.segments, segment{literal: "$"})
			continue
		}

		// consume identifier path: [a-zA-Z_][a-zA-Z0-9_.\[\]]*
		end := 0
		for end < len(rem) {
			r, sz := utf8.DecodeRuneInString(rem[end:])
			if !isIdentChar(r) {
				break
			}
			end += sz
		}
		path := rem[:end]
		rem = rem[end:]

		// optional modifier: |modifier (ends at whitespace, '$', or end)
		modifier := ""
		if len(rem) > 0 && rem[0] == '|' {
			rem = rem[1:]
			modEnd := 0
			for modEnd < len(rem) {
				r, sz := utf8.DecodeRuneInString(rem[modEnd:])
				if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '$' {
					break
				}
				modEnd += sz
			}
			modifier = rem[:modEnd]
			rem = rem[modEnd:]
		}

		t.segments = append(t.segments, segment{path: path, modifier: modifier, isVar: true})
	}
	return t
}

func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentChar(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9') || r == '.' || r == '[' || r == ']'
}

// Execute renders t with the given params and returns the result as a string.
func (t *Template) Execute(params ...any) string {
	var b strings.Builder
	t.writeTo(&b, params)
	return b.String()
}

// WriteTo renders t into w.
func (t *Template) WriteTo(w io.Writer, params ...any) {
	t.writeTo(w, params)
}

func (t *Template) writeTo(w io.Writer, params []any) {
	for _, seg := range t.segments {
		if !seg.isVar {
			_, _ = io.WriteString(w, seg.literal)
			continue
		}
		val, found := resolve(seg.path, params)
		out := ""
		if found {
			out = stringify(val)
		}
		switch {
		case out != "":
			if seg.modifier != "" && isTransform(seg.modifier) {
				out = applyTransform(out, seg.modifier)
			}
			_, _ = io.WriteString(w, out)
		case seg.modifier != "" && !isTransform(seg.modifier):
			// value missing or empty and modifier is a default — use it
			_, _ = io.WriteString(w, seg.modifier)
		case !found:
			// variable not found, no usable default — keep original placeholder
			_, _ = io.WriteString(w, "$")
			_, _ = io.WriteString(w, seg.path)
			if seg.modifier != "" {
				_, _ = io.WriteString(w, "|")
				_, _ = io.WriteString(w, seg.modifier)
			}
		}
		// found but stringifies to "" with a transform → output nothing
	}
}

// Render compiles src and executes it with the given params.
func Render(src string, params ...any) string {
	return Parse(src).Execute(params...)
}

// RenderWriter compiles src and writes the rendered output to w.
func RenderWriter(w io.Writer, src string, params ...any) {
	Parse(src).WriteTo(w, params...)
}

// Pairs converts a flat alternating key/value list into a map[string]any.
// Example: Pairs("name", "Alice", "age", 30) → map[string]any{"name":"Alice","age":30}
func Pairs(args ...any) map[string]any {
	m := make(map[string]any, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		if k, ok := args[i].(string); ok {
			m[k] = args[i+1]
		}
	}
	return m
}

// resolve looks up path across all params, returning the first non-nil match.
func resolve(path string, params []any) (any, bool) {
	for _, p := range params {
		v, err := dot.Get(p, path)
		if err == nil && v != nil {
			return v, true
		}
	}
	return nil, false
}

var titleCaser = cases.Title(language.English, cases.NoLower)

// applyTransform applies the named string transform to s.
func applyTransform(s, transform string) string {
	switch strings.ToLower(transform) {
	case "upper":
		return strings.ToUpper(s)
	case "lower":
		return strings.ToLower(s)
	case "title":
		return titleCaser.String(s)
	default:
		return s
	}
}

// isTransform reports whether m is a known transform name.
func isTransform(m string) bool {
	switch strings.ToLower(m) {
	case "upper", "lower", "title":
		return true
	}
	return false
}

// stringify converts v to its string representation. It checks for fmt.Stringer
// first, then uses a type switch for all common scalar types, falling back to
// fmt.Sprint for everything else.
func stringify(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(t)
	case int8:
		return strconv.FormatInt(int64(t), 10)
	case int16:
		return strconv.FormatInt(int64(t), 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case uint8:
		return strconv.FormatUint(uint64(t), 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}
