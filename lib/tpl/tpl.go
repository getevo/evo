package tpl

import (
	"encoding/json"
	"fmt"
	htmlpkg "html"
	"io"
	neturlpkg "net/url"
	"reflect"
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
// Pass 0 to disable caching. If n is smaller than the current cache size, the cache is cleared.
func SetCacheSize(n int) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cacheLimit = n
	if n == 0 || len(cache) > n {
		cache = make(map[string]*Template, n)
	}
}

// registeredFunc holds a reflected function for dynamic dispatch.
type registeredFunc struct {
	fn  reflect.Value
	typ reflect.Type
}

var (
	funcsMu sync.RWMutex
	funcs   = map[string]*registeredFunc{}
)

// RegisterFunc registers a named function for use in templates.
// fn must be a function value of any signature. Arguments are coerced to the
// declared parameter types automatically. The return value is stringified.
// Functions may return one or two values; when two values are returned the last
// must implement error — a non-nil error suppresses any output.
//
// A registered function can be used in three ways:
//
//	$fn(arg1, arg2)   — explicit call; args may be $variables or literals ("str", 42, 3.14)
//	$var|fn           — modifier: receives the variable's resolved value as first argument
//	$fn               — standalone: called with no arguments when name is not in params
func RegisterFunc(name string, fn any) {
	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		panic(fmt.Sprintf("tpl: RegisterFunc %q: fn must be a function, got %T", name, fn))
	}
	funcsMu.Lock()
	funcs[name] = &registeredFunc{fn: rv, typ: rv.Type()}
	funcsMu.Unlock()
}

func lookupFunc(name string) (*registeredFunc, bool) {
	funcsMu.RLock()
	rf, ok := funcs[name]
	funcsMu.RUnlock()
	return rf, ok
}

// callFunc invokes rf with args and returns the stringified result.
// Type mismatches and panics are recovered silently (returns "").
func callFunc(rf *registeredFunc, args []any) (result string) {
	defer func() {
		if r := recover(); r != nil {
			result = ""
		}
	}()

	ft := rf.typ
	numIn := ft.NumIn()
	isVariadic := ft.IsVariadic()

	var in []reflect.Value

	if isVariadic {
		fixedCount := numIn - 1
		varElemType := ft.In(numIn - 1).Elem()
		for i := 0; i < fixedCount; i++ {
			var v any
			if i < len(args) {
				v = args[i]
			}
			in = append(in, coerceArg(v, ft.In(i)))
		}
		for i := fixedCount; i < len(args); i++ {
			in = append(in, coerceArg(args[i], varElemType))
		}
	} else {
		in = make([]reflect.Value, numIn)
		for i := 0; i < numIn; i++ {
			var v any
			if i < len(args) {
				v = args[i]
			}
			in[i] = coerceArg(v, ft.In(i))
		}
	}

	out := rf.fn.Call(in)

	if len(out) == 0 {
		return ""
	}

	// If the last return value implements error, check it.
	errType := reflect.TypeOf((*error)(nil)).Elem()
	if out[len(out)-1].Type().Implements(errType) {
		if !out[len(out)-1].IsNil() {
			return "" // function returned an error — produce no output
		}
		out = out[:len(out)-1]
	}

	if len(out) == 0 {
		return ""
	}
	return stringify(out[0].Interface())
}

// coerceArg converts val to the reflect.Value required for a function parameter of targetType.
func coerceArg(val any, targetType reflect.Type) reflect.Value {
	if val == nil {
		return reflect.Zero(targetType)
	}

	rv := reflect.ValueOf(val)

	// Direct type match.
	if rv.Type() == targetType {
		return rv
	}

	// Interface target (e.g. any / interface{} / fmt.Stringer).
	if targetType.Kind() == reflect.Interface {
		if rv.Type().Implements(targetType) {
			return rv
		}
		return reflect.Zero(targetType)
	}

	// Assignable without conversion.
	if rv.Type().AssignableTo(targetType) {
		return rv
	}

	// Numeric / kind conversion (e.g. int64 literal → int param, float64 → float32).
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType)
	}

	// Stringify to meet a string target.
	if targetType.Kind() == reflect.String {
		return reflect.ValueOf(stringify(val)).Convert(targetType)
	}

	// Parse a string argument into a numeric or bool target.
	if s, ok := val.(string); ok {
		switch targetType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, _ := strconv.ParseInt(s, 10, 64)
			return reflect.ValueOf(n).Convert(targetType)
		case reflect.Float32, reflect.Float64:
			f, _ := strconv.ParseFloat(s, 64)
			return reflect.ValueOf(f).Convert(targetType)
		case reflect.Bool:
			b, _ := strconv.ParseBool(s)
			return reflect.ValueOf(b).Convert(targetType)
		}
	}

	return reflect.Zero(targetType)
}

// argSpec is a parsed argument in a $fn(...) call.
type argSpec struct {
	isVar   bool
	path    string // variable path when isVar == true
	literal any    // pre-parsed value (string, int64, float64) when isVar == false
}

// segment is one compiled piece of a template: a literal string, a variable reference,
// or a function call.
type segment struct {
	literal string

	// variable reference: $path or $path|mod1|mod2|...
	path      string
	modifiers []string // zero or more pipe-chained modifiers/transforms
	isVar     bool

	// function call: $funcName(callArgs...)
	isCall   bool
	funcName string
	callArgs []argSpec
}

// Template is a pre-compiled template ready for repeated execution.
type Template struct {
	segments []segment
}

// Parse compiles src into a Template and caches the result up to the configured cache size.
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
			// '$' followed by a non-identifier character → literal '$'
			t.segments = append(t.segments, segment{literal: "$"})
			continue
		}

		// Consume the identifier. Track whether it contains dots or brackets so
		// that only plain names (e.g. "fn", not "arr[0]") trigger call syntax.
		end := 0
		isSimpleName := true
		for end < len(rem) {
			r, sz := utf8.DecodeRuneInString(rem[end:])
			if !isIdentChar(r) {
				break
			}
			if r == '.' || r == '[' || r == ']' {
				isSimpleName = false
			}
			end += sz
		}
		path := rem[:end]
		rem = rem[end:]

		// Function call: $name(args...) — only for plain, unqualified names.
		if isSimpleName && len(rem) > 0 && rem[0] == '(' {
			rem = rem[1:] // skip '('
			args := parseCallArgs(&rem)
			t.segments = append(t.segments, segment{
				isCall:   true,
				funcName: path,
				callArgs: args,
			})
			continue
		}

		// Optional pipe-chained modifiers: |mod1|mod2|...
		var modifiers []string
		for len(rem) > 0 && rem[0] == '|' {
			rem = rem[1:]
			modEnd := 0
			for modEnd < len(rem) {
				r, sz := utf8.DecodeRuneInString(rem[modEnd:])
				if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '$' || r == '|' {
					break
				}
				modEnd += sz
			}
			if modEnd == 0 {
				break
			}
			modifiers = append(modifiers, rem[:modEnd])
			rem = rem[modEnd:]
		}

		t.segments = append(t.segments, segment{path: path, modifiers: modifiers, isVar: true})
	}
	return t
}

// parseCallArgs reads a comma-separated argument list until ')' and advances *rem past ')'.
func parseCallArgs(rem *string) []argSpec {
	var args []argSpec
	for {
		*rem = strings.TrimLeft(*rem, " \t\n\r")
		if len(*rem) == 0 {
			break
		}
		if (*rem)[0] == ')' {
			*rem = (*rem)[1:] // consume ')'
			break
		}
		if (*rem)[0] == ',' {
			*rem = (*rem)[1:]
			continue
		}
		arg, ok := parseCallArg(rem)
		if !ok {
			// Unrecognised token — skip to the next ',' or ')' to avoid an infinite loop.
			i := strings.IndexAny(*rem, ",)")
			if i < 0 {
				*rem = ""
				break
			}
			*rem = (*rem)[i:]
			continue
		}
		args = append(args, arg)
	}
	return args
}

// parseCallArg parses one argument — a $variable, a "string", a 'string', or a numeric literal.
func parseCallArg(rem *string) (argSpec, bool) {
	*rem = strings.TrimLeft(*rem, " \t\n\r")
	if len(*rem) == 0 {
		return argSpec{}, false
	}

	// $variable — supports dotted paths and bracket access: $user.Name, $arr[0]
	if (*rem)[0] == '$' {
		*rem = (*rem)[1:]
		end := 0
		for end < len(*rem) {
			r, sz := utf8.DecodeRuneInString((*rem)[end:])
			if !isIdentChar(r) {
				break
			}
			end += sz
		}
		if end == 0 {
			return argSpec{}, false
		}
		path := (*rem)[:end]
		*rem = (*rem)[end:]
		return argSpec{isVar: true, path: path}, true
	}

	// "double-quoted string literal" with escape sequences
	if (*rem)[0] == '"' {
		*rem = (*rem)[1:]
		var sb strings.Builder
		for len(*rem) > 0 {
			if (*rem)[0] == '"' {
				*rem = (*rem)[1:]
				break
			}
			if (*rem)[0] == '\\' && len(*rem) > 1 {
				switch (*rem)[1] {
				case '"':
					sb.WriteByte('"')
				case '\\':
					sb.WriteByte('\\')
				case 'n':
					sb.WriteByte('\n')
				case 't':
					sb.WriteByte('\t')
				default:
					sb.WriteByte((*rem)[1])
				}
				*rem = (*rem)[2:]
				continue
			}
			sb.WriteByte((*rem)[0])
			*rem = (*rem)[1:]
		}
		return argSpec{literal: sb.String()}, true
	}

	// 'single-quoted string literal'
	if (*rem)[0] == '\'' {
		*rem = (*rem)[1:]
		var sb strings.Builder
		for len(*rem) > 0 {
			if (*rem)[0] == '\'' {
				*rem = (*rem)[1:]
				break
			}
			sb.WriteByte((*rem)[0])
			*rem = (*rem)[1:]
		}
		return argSpec{literal: sb.String()}, true
	}

	// Numeric literal: optional sign, integer digits, optional fractional part.
	end := 0
	if end < len(*rem) && ((*rem)[end] == '-' || (*rem)[end] == '+') {
		end++
	}
	digitStart := end
	hasDot := false
	for end < len(*rem) {
		c := (*rem)[end]
		if c >= '0' && c <= '9' {
			end++
		} else if c == '.' && !hasDot {
			hasDot = true
			end++
		} else {
			break
		}
	}
	if end > digitStart { // at least one digit
		numStr := (*rem)[:end]
		*rem = (*rem)[end:]
		if hasDot {
			f, _ := strconv.ParseFloat(numStr, 64)
			return argSpec{literal: f}, true
		}
		n, _ := strconv.ParseInt(numStr, 10, 64)
		return argSpec{literal: n}, true
	}

	return argSpec{}, false
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
		switch {

		case seg.isCall:
			// Resolve each argument, then call the registered function.
			resolvedArgs := make([]any, len(seg.callArgs))
			for i, arg := range seg.callArgs {
				if arg.isVar {
					v, _ := resolve(arg.path, params)
					resolvedArgs[i] = v
				} else {
					resolvedArgs[i] = arg.literal
				}
			}
			if rf, ok := lookupFunc(seg.funcName); ok {
				_, _ = io.WriteString(w, callFunc(rf, resolvedArgs))
			}
			// Unknown function: produce no output.

		case !seg.isVar:
			_, _ = io.WriteString(w, seg.literal)

		default:
			// Variable reference.
			val, found := resolve(seg.path, params)
			out := ""
			if found {
				out = stringify(val)
			}

			// Multi-modifier (chained): apply each in sequence.
			if len(seg.modifiers) > 1 {
				for _, mod := range seg.modifiers {
					if result, ok := applyModifier(out, val, mod); ok {
						out = result
						val = out // updated value feeds subsequent modifiers
					}
					// Unknown modifier in a chain: silently skip.
				}
				_, _ = io.WriteString(w, out)
				continue
			}

			// Zero or one modifier: original behaviour preserved exactly.
			mod := ""
			if len(seg.modifiers) == 1 {
				mod = seg.modifiers[0]
			}

			switch {
			case out != "":
				if mod != "" {
					if result, ok := applyModifier(out, val, mod); ok {
						out = result
					}
					// else: modifier is a default-value string — keep the current value.
				}
				_, _ = io.WriteString(w, out)

			case mod != "":
				if isBuiltinTransform(mod) {
					// Built-in transform on a missing variable — keep the placeholder.
					_, _ = io.WriteString(w, "$")
					_, _ = io.WriteString(w, seg.path)
					_, _ = io.WriteString(w, "|")
					_, _ = io.WriteString(w, mod)
				} else if rf, ok := lookupFunc(mod); ok {
					// User-registered function used as modifier on a missing variable — call with nil.
					_, _ = io.WriteString(w, callFunc(rf, []any{nil}))
				} else {
					// Unknown modifier — treat as a default value.
					_, _ = io.WriteString(w, mod)
				}

			case !found:
				// Variable not found and no modifier — try the path as a standalone function call.
				if rf, ok := lookupFunc(seg.path); ok {
					_, _ = io.WriteString(w, callFunc(rf, nil))
				} else {
					// Keep the original placeholder unchanged.
					_, _ = io.WriteString(w, "$")
					_, _ = io.WriteString(w, seg.path)
					if mod != "" {
						_, _ = io.WriteString(w, "|")
						_, _ = io.WriteString(w, mod)
					}
				}
			}
		}
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

// applyModifier applies a built-in transform or registered function as a modifier.
// Returns (result, true) when the modifier is recognised; ("", false) for default-value strings.
func applyModifier(str string, val any, mod string) (string, bool) {
	switch strings.ToLower(mod) {
	case "upper":
		return strings.ToUpper(str), true
	case "lower":
		return strings.ToLower(str), true
	case "title":
		return titleCaser.String(str), true
	case "html":
		return htmlpkg.EscapeString(str), true
	case "url":
		return neturlpkg.QueryEscape(str), true
	case "trim":
		return strings.TrimSpace(str), true
	case "json":
		b, err := json.Marshal(val)
		if err != nil {
			return str, true
		}
		return string(b), true
	}
	if rf, ok := lookupFunc(mod); ok {
		return callFunc(rf, []any{val}), true
	}
	return "", false
}

// isBuiltinTransform reports whether mod is a reserved built-in transform name.
func isBuiltinTransform(mod string) bool {
	switch strings.ToLower(mod) {
	case "upper", "lower", "title", "html", "url", "trim", "json":
		return true
	}
	return false
}

// stringify converts v to its string representation. It checks for fmt.Stringer
// first, then uses a type switch for common scalar types, falling back to fmt.Sprint.
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
