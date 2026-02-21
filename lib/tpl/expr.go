package tpl

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/getevo/evo/v2/lib/dot"
)

// ── Expression Parser ─────────────────────────────────────────────────────────
// Precedence (low → high):
//   ternary ? : → ?? → || → && → == != → < > <= >= → + - . → * / % → unary ! - → primary

func parseExpr(ts *tokenStream) Expr {
	e := parseNullCoalescing(ts)
	// Ternary: cond ? then : else
	if ts.peek().Kind == TkOp && ts.peek().Val == "?" {
		ts.next()
		then := parseExpr(ts)
		ts.consume(TkOp, ":")
		els := parseExpr(ts)
		return TernaryExpr{Cond: e, Then: then, Else: els}
	}
	return e
}

func parseNullCoalescing(ts *tokenStream) Expr { return parseBin(ts, parseOr, "??") }
func parseOr(ts *tokenStream) Expr             { return parseBin(ts, parseAnd, "||") }
func parseAnd(ts *tokenStream) Expr            { return parseBin(ts, parseEquality, "&&") }
func parseEquality(ts *tokenStream) Expr       { return parseBin(ts, parseRelational, "==", "!=") }
func parseRelational(ts *tokenStream) Expr {
	return parseBin(ts, parseAdditive, "<", ">", "<=", ">=")
}
func parseAdditive(ts *tokenStream) Expr {
	return parseBin(ts, parseMultiplicative, "+", "-", ".")
}
func parseMultiplicative(ts *tokenStream) Expr {
	return parseBin(ts, parseUnary, "*", "/", "%")
}

func parseBin(ts *tokenStream, next func(*tokenStream) Expr, ops ...string) Expr {
	l := next(ts)
	for {
		t := ts.peek()
		if t.Kind != TkOp {
			break
		}
		matched := false
		for _, op := range ops {
			if t.Val == op {
				matched = true
				break
			}
		}
		if !matched {
			break
		}
		ts.next()
		r := next(ts)
		l = BinExpr{Op: t.Val, L: l, R: r}
	}
	return l
}

func parseUnary(ts *tokenStream) Expr {
	t := ts.peek()
	if t.Kind == TkOp && (t.Val == "!" || t.Val == "-") {
		ts.next()
		return UnExpr{Op: t.Val, X: parseUnary(ts)}
	}
	return parsePrimary(ts)
}

func parsePrimary(ts *tokenStream) Expr {
	t := ts.peek()

	// Parenthesised expression
	if t.Kind == TkLParen {
		ts.next()
		e := parseExpr(ts)
		ts.consume(TkRParen, "")
		return e
	}

	// Literals
	if t.Kind == TkString {
		ts.next()
		return LitExpr{Val: t.Lit}
	}
	if t.Kind == TkInt {
		ts.next()
		return LitExpr{Val: t.Lit}
	}
	if t.Kind == TkFloat {
		ts.next()
		return LitExpr{Val: t.Lit}
	}

	// $variable  (may be followed by ++ / --)
	if t.Kind == TkVar {
		ts.next()
		name := t.Val
		// Check for $var++ / $var--
		nxt := ts.peek()
		if nxt.Kind == TkOp && (nxt.Val == "++" || nxt.Val == "--") {
			ts.next()
			return IncDecExpr{Name: name, Op: nxt.Val}
		}
		// Build full path: handle .field and [index] access
		path := name + parseSuffix(ts)
		return VarExpr{Path: path}
	}

	// Bare identifier → keyword, function call, or bare ident
	if t.Kind == TkIdent {
		switch t.Val {
		case "true":
			ts.next()
			return LitExpr{Val: true}
		case "false":
			ts.next()
			return LitExpr{Val: false}
		case "null", "nil":
			ts.next()
			return LitExpr{Val: nil}
		case "isset":
			// isset($path) — special form: reports whether variable is defined
			ts.next()
			ts.consume(TkLParen, "")
			inner := ts.peek()
			path := ""
			if inner.Kind == TkVar {
				ts.next()
				path = inner.Val + parseSuffix(ts)
			}
			ts.consume(TkRParen, "")
			return IsSetExpr{Path: path}
		}

		ts.next()
		name := t.Val
		if ts.peek().Kind == TkLParen {
			return parseCallTail(ts, name)
		}
		// bare ident used as expr — treat as string literal (e.g. bare keyword fallback)
		return LitExpr{Val: name}
	}

	// Array literal: [expr, expr, ...]
	if t.Kind == TkLBracket {
		ts.next()
		var elems []Expr
		for !ts.eof() && ts.peek().Kind != TkRBracket {
			if ts.peek().Kind == TkComma {
				ts.next()
				continue
			}
			elems = append(elems, parseExpr(ts))
		}
		ts.consume(TkRBracket, "")
		return ArrayLitExpr{Elems: elems}
	}

	// Map literal: {"key": val, "key2": val2}
	if t.Kind == TkLBrace {
		ts.next()
		var keys []string
		var vals []Expr
		for !ts.eof() && ts.peek().Kind != TkRBrace {
			if ts.peek().Kind == TkComma {
				ts.next()
				continue
			}
			keyTok := ts.next()
			var key string
			switch keyTok.Kind {
			case TkString:
				key = fmt.Sprint(keyTok.Lit)
			default:
				key = keyTok.Val
			}
			ts.consume(TkOp, ":")
			val := parseExpr(ts)
			keys = append(keys, key)
			vals = append(vals, val)
		}
		ts.consume(TkRBrace, "")
		return MapLitExpr{Keys: keys, Vals: vals}
	}

	// Fallback: consume and return nil literal
	ts.next()
	return LitExpr{Val: nil}
}

// parseSuffix reads zero or more  .field  or  [expr]  suffixes after a variable name.
// Returns the additional path string to append.
func parseSuffix(ts *tokenStream) string {
	var sb strings.Builder
	for {
		t := ts.peek()
		if t.Kind == TkOp && t.Val == "." {
			// Only treat as field accessor when the NEXT token (after the dot) is
			// an identifier. Otherwise "." is the binary string-concat operator.
			if ts.pos+1 < len(ts.toks) {
				field := ts.toks[ts.pos+1]
				if field.Kind == TkIdent || field.Kind == TkVar {
					ts.next() // consume "."
					ts.next() // consume field name
					sb.WriteByte('.')
					sb.WriteString(field.Val)
					continue
				}
			}
			break // leave "." for the binary operator parser
		}
		if t.Kind == TkLBracket {
			ts.next()
			inner := ts.peek()
			// String key: ["key"] → [key]
			if inner.Kind == TkString {
				ts.next()
				sb.WriteByte('[')
				sb.WriteString(fmt.Sprint(inner.Lit))
				sb.WriteByte(']')
			} else if inner.Kind == TkInt {
				ts.next()
				sb.WriteByte('[')
				sb.WriteString(inner.Val)
				sb.WriteByte(']')
			} else if inner.Kind == TkVar {
				// Dynamic variable key — not expressible in a static path.
				// We store a special marker and the executor handles it.
				ts.next()
				sb.WriteString("[$")
				sb.WriteString(inner.Val)
				sb.WriteByte(']')
			}
			ts.consume(TkRBracket, "")
			continue
		}
		break
	}
	return sb.String()
}

// parseCallTail parses  (arg, arg, ...)  after the function name.
func parseCallTail(ts *tokenStream, name string) Expr {
	ts.next() // skip '('
	var args []Expr
	for !ts.eof() {
		if ts.peek().Kind == TkRParen {
			break
		}
		if ts.peek().Kind == TkComma {
			ts.next()
			continue
		}
		args = append(args, parseExpr(ts))
	}
	ts.consume(TkRParen, "")
	return CallExpr{Name: name, Args: args}
}

// ── Expression Evaluator ──────────────────────────────────────────────────────

// evalExpr evaluates expr against ctx, returning an untyped Go value.
func evalExpr(e Expr, ctx *Context) any {
	if e == nil {
		return nil
	}
	switch x := e.(type) {
	case LitExpr:
		return x.Val
	case VarExpr:
		return evalVar(x.Path, ctx)
	case CallExpr:
		return evalCall(x, ctx)
	case BinExpr:
		return evalBin(x, ctx)
	case UnExpr:
		return evalUn(x, ctx)
	case IncDecExpr:
		return evalIncDec(x, ctx)
	case TernaryExpr:
		if toBool(evalExpr(x.Cond, ctx)) {
			return evalExpr(x.Then, ctx)
		}
		return evalExpr(x.Else, ctx)
	case IsSetExpr:
		_, ok := ctx.GetPath(x.Path)
		return ok
	case ArrayLitExpr:
		result := make([]any, len(x.Elems))
		for i, e := range x.Elems {
			result[i] = evalExpr(e, ctx)
		}
		return result
	case MapLitExpr:
		result := make(map[string]any, len(x.Keys))
		for i, k := range x.Keys {
			result[k] = evalExpr(x.Vals[i], ctx)
		}
		return result
	}
	return nil
}

// evalVar resolves a path that may contain bracket indices like [0], ["key"], or [$var].
func evalVar(path string, ctx *Context) any {
	// Any bracket access (static or dynamic) is handled by evalDynamicPath which
	// correctly handles both integer slice indices and string map keys.
	if strings.Contains(path, "[") {
		return evalDynamicPath(path, ctx)
	}
	v, _ := ctx.GetPath(path)
	return v
}

// evalDynamicPath resolves a path that may contain one or more bracket or dot
// segments, e.g. "m[0][name]", "m[$k1][$k2]", or "m[$k].Field".
// It processes segments iteratively so chained indices all work.
func evalDynamicPath(path string, ctx *Context) any {
	// Resolve the static base prefix (everything before the first '[').
	i := strings.IndexByte(path, '[')
	base := path[:i]
	var current any
	if base != "" {
		current, _ = ctx.GetPath(base)
	}

	rest := path[i:]
	for len(rest) > 0 {
		switch {
		case len(rest) > 1 && rest[0] == '[' && rest[1] == '$':
			// Dynamic variable index: [$varname]
			j := strings.IndexByte(rest, ']')
			if j < 0 {
				return current
			}
			varName := rest[2:j]
			rest = rest[j+1:]
			keyVal, _ := ctx.Get(varName)
			current = applyDynIndex(current, keyVal)

		case rest[0] == '[':
			// Static index: [key] or [0]
			j := strings.IndexByte(rest, ']')
			if j < 0 {
				return current
			}
			keyStr := rest[1:j]
			rest = rest[j+1:]
			current = applyDynIndex(current, keyStr)

		case rest[0] == '.':
			// Field access: .FieldName — consume up to the next . or [
			rest = rest[1:]
			end := 0
			for end < len(rest) && rest[end] != '.' && rest[end] != '[' {
				end++
			}
			if end == 0 {
				continue
			}
			field := rest[:end]
			rest = rest[end:]
			if current == nil {
				return nil
			}
			v, err := dot.Get(current, field)
			if err != nil {
				return nil
			}
			current = v

		default:
			// Remaining is a plain path segment with no dynamic indices.
			result, _ := (&Context{params: []any{current}}).GetPath(rest)
			return result
		}
	}
	return current
}

// applyDynIndex indexes into a map or slice using key.
// For maps with integer key types the key is coerced to the map's key type.
func applyDynIndex(base any, key any) any {
	if base == nil {
		return nil
	}
	rv := reflect.ValueOf(base)
	if !rv.IsValid() {
		return nil
	}
	switch rv.Kind() {
	case reflect.Map:
		keyStr := stringify(key)
		kv := reflect.ValueOf(keyStr)
		if k := rv.Type().Key().Kind(); k == reflect.Int || k == reflect.Int64 {
			n, _ := toInt64(key)
			kv = reflect.ValueOf(n).Convert(rv.Type().Key())
		}
		mv := rv.MapIndex(kv)
		if mv.IsValid() {
			return mv.Interface()
		}
		return nil
	case reflect.Slice, reflect.Array:
		n, _ := toInt64(key)
		if n >= 0 && n < int64(rv.Len()) {
			return rv.Index(int(n)).Interface()
		}
		return nil
	}
	return nil
}

func evalCall(x CallExpr, ctx *Context) any {
	rf, ok := lookupFunc(x.Name)
	if !ok {
		return nil
	}
	args := make([]any, len(x.Args))
	for i, a := range x.Args {
		args[i] = evalExpr(a, ctx)
	}
	return callFuncRaw(rf, args)
}

func evalBin(x BinExpr, ctx *Context) any {
	// Short-circuit logical operators
	switch x.Op {
	case "&&":
		if !toBool(evalExpr(x.L, ctx)) {
			return false
		}
		return toBool(evalExpr(x.R, ctx))
	case "||":
		lv := evalExpr(x.L, ctx)
		if toBool(lv) {
			return true
		}
		return toBool(evalExpr(x.R, ctx))
	case "??":
		// Null-coalescing: return left if non-nil, else right
		lv := evalExpr(x.L, ctx)
		if lv != nil {
			return lv
		}
		return evalExpr(x.R, ctx)
	}

	l := evalExpr(x.L, ctx)
	r := evalExpr(x.R, ctx)

	switch x.Op {
	case "+":
		return numericBin(l, r, func(a, b float64) float64 { return a + b })
	case "-":
		return numericBin(l, r, func(a, b float64) float64 { return a - b })
	case "*":
		return numericBin(l, r, func(a, b float64) float64 { return a * b })
	case "/":
		bv, _ := toFloat64(r)
		if bv == 0 {
			return nil
		}
		return numericBin(l, r, func(a, b float64) float64 { return a / b })
	case "%":
		ai, _ := toInt64(l)
		bi, _ := toInt64(r)
		if bi == 0 {
			return nil
		}
		return ai % bi
	case ".":
		// String concatenation
		return stringify(l) + stringify(r)
	case "==":
		return equal(l, r)
	case "!=":
		return !equal(l, r)
	case "<":
		return compare(l, r) < 0
	case ">":
		return compare(l, r) > 0
	case "<=":
		return compare(l, r) <= 0
	case ">=":
		return compare(l, r) >= 0
	}
	return nil
}

func evalUn(x UnExpr, ctx *Context) any {
	v := evalExpr(x.X, ctx)
	switch x.Op {
	case "!":
		return !toBool(v)
	case "-":
		// Check float types first to preserve precision (toInt64 would truncate them).
		switch tv := v.(type) {
		case float64:
			return -tv
		case float32:
			return -float64(tv)
		}
		if n, ok := toInt64(v); ok {
			return -n
		}
		if f, ok := toFloat64(v); ok {
			return -f
		}
	}
	return nil
}

func evalIncDec(x IncDecExpr, ctx *Context) any {
	cur, _ := ctx.Get(x.Name)
	var newVal any
	if n, ok := toInt64(cur); ok {
		if x.Op == "++" {
			newVal = n + 1
		} else {
			newVal = n - 1
		}
	} else if f, ok := toFloat64(cur); ok {
		if x.Op == "++" {
			newVal = f + 1
		} else {
			newVal = f - 1
		}
	} else {
		newVal = int64(0)
	}
	ctx.Set(x.Name, newVal)
	return newVal
}

// ── callFuncRaw ───────────────────────────────────────────────────────────────

// callFuncRaw is like callFunc but returns the raw (any) return value instead
// of stringifying it. Panics are recovered → nil.
func callFuncRaw(rf *registeredFunc, args []any) (result any) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
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
		return nil
	}
	errType := reflect.TypeOf((*error)(nil)).Elem()
	if out[len(out)-1].Type().Implements(errType) {
		if !out[len(out)-1].IsNil() {
			return nil
		}
		out = out[:len(out)-1]
	}
	if len(out) == 0 {
		return nil
	}
	return out[0].Interface()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toBool(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case string:
		return t != "" && t != "0" && t != "false"
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !rv.IsNil()
	}
	return true
}

func toFloat64(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint, uint8, uint16, uint32, uint64:
		rv := reflect.ValueOf(v)
		return float64(rv.Uint()), true
	case string:
		f, err := strconvParseFloat(t)
		return f, err == nil
	}
	return 0, false
}

func toInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int:
		return int64(t), true
	case int8:
		return int64(t), true
	case int16:
		return int64(t), true
	case int32:
		return int64(t), true
	case int64:
		return t, true
	case uint:
		return int64(t), true
	case uint8:
		return int64(t), true
	case uint16:
		return int64(t), true
	case uint32:
		return int64(t), true
	case uint64:
		return int64(t), true
	case float64:
		return int64(t), true
	case float32:
		return int64(t), true
	case string:
		n, err := strconvParseInt(t)
		return n, err == nil
	case bool:
		if t {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func numericBin(l, r any, fn func(float64, float64) float64) any {
	lf, lok := toFloat64(l)
	rf, rok := toFloat64(r)
	if !lok || !rok {
		return nil
	}
	result := fn(lf, rf)
	// Return int64 if both operands were integer-like and result has no fraction
	_, lIsInt := toInt64(l)
	_, rIsInt := toInt64(r)
	if lIsInt && rIsInt && result == math.Trunc(result) {
		return int64(result)
	}
	return result
}

func equal(l, r any) bool {
	if l == nil && r == nil {
		return true
	}
	if l == nil || r == nil {
		return false
	}
	// Try numeric comparison first
	lf, lok := toFloat64(l)
	rf, rok := toFloat64(r)
	if lok && rok {
		return lf == rf
	}
	// Bool comparison
	lb, lok := l.(bool)
	rb, rok := r.(bool)
	if lok && rok {
		return lb == rb
	}
	return stringify(l) == stringify(r)
}

func compare(l, r any) int {
	lf, lok := toFloat64(l)
	rf, rok := toFloat64(r)
	if lok && rok {
		if lf < rf {
			return -1
		}
		if lf > rf {
			return 1
		}
		return 0
	}
	ls, rs := stringify(l), stringify(r)
	if ls < rs {
		return -1
	}
	if ls > rs {
		return 1
	}
	return 0
}

func strconvParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func strconvParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
