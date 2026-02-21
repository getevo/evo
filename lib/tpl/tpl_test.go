package tpl

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type User struct {
	Name   string
	Family string
}

// Named implements fmt.Stringer
type Named struct{ name string }

func (n Named) String() string { return n.name }

// ── Existing tests ────────────────────────────────────────────────────────────

func TestRender(t *testing.T) {
	text := `Hello $title $user.Name $user.Family you have $sender[0] email From $sender[2][from]($sender[2][user].Name $sender[2][user].Family) at $date[0]:$date[1]:$date[2]`
	want := "Hello Mrs Maria Rossy you have 1 email From example@example.com(Marco Pollo) at 10:15:20"
	got := Render(text, map[string]any{
		"title":  "Mrs",
		"user":   User{Name: "Maria", Family: "Rossy"},
		"sender": []any{1, "empty!", map[string]any{"from": "example@example.com", "user": User{Name: "Marco", Family: "Pollo"}}},
		"date":   []int{10, 15, 20},
	})
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestDollarEscape(t *testing.T) {
	if got := Render("price: $$10"); got != "price: $10" {
		t.Errorf("got %q", got)
	}
	if got := Render("a$$b$$c"); got != "a$b$c" {
		t.Errorf("got %q", got)
	}
}

func TestBuiltinTransforms(t *testing.T) {
	cases := []struct {
		src, want string
		params    map[string]any
	}{
		{"$name|upper", "HELLO WORLD", Pairs("name", "hello world")},
		{"$name|lower", "hello", Pairs("name", "HELLO")},
		{"$name|title", "Hello World", Pairs("name", "hello world")},
		{"$name|trim", "hi", Pairs("name", "  hi  ")},
		{"$html|html", "&lt;b&gt;bold&lt;/b&gt;", Pairs("html", "<b>bold</b>")},
		{"$q|url", "hello+world", Pairs("q", "hello world")},
	}
	for _, c := range cases {
		if got := Render(c.src, c.params); got != c.want {
			t.Errorf("Render(%q): got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestJsonTransform(t *testing.T) {
	type Point struct {
		X, Y int
	}
	got := Render("$p|json", Pairs("p", Point{1, 2}))
	if got != `{"X":1,"Y":2}` {
		t.Errorf("json transform: got %q", got)
	}
	// scalar
	got = Render("$n|json", Pairs("n", 42))
	if got != "42" {
		t.Errorf("json scalar: got %q", got)
	}
}

func TestDefault(t *testing.T) {
	// missing variable uses default
	if got := Render("$missing|guest", map[string]any{}); got != "guest" {
		t.Errorf("missing default: got %q", got)
	}
	// present variable ignores default
	if got := Render("$name|guest", Pairs("name", "alice")); got != "alice" {
		t.Errorf("present ignores default: got %q", got)
	}
}

func TestScalarTypes(t *testing.T) {
	m := map[string]any{
		"b":   true,
		"bf":  false,
		"i16": int16(42),
		"u32": uint32(99),
		"f32": float32(1.5),
	}
	if got := Render("$b $bf $i16 $u32 $f32", m); got != "true false 42 99 1.5" {
		t.Errorf("scalars: got %q", got)
	}
}

func TestStringer(t *testing.T) {
	m := map[string]any{"obj": Named{name: "claude"}}
	if got := Render("$obj", m); got != "claude" {
		t.Errorf("stringer: got %q", got)
	}
}

func TestPairs(t *testing.T) {
	m := Pairs("x", 1, "y", "two")
	if m["x"] != 1 || m["y"] != "two" || len(m) != 2 {
		t.Errorf("pairs: got %v", m)
	}
}

func TestRenderWriter(t *testing.T) {
	var sb strings.Builder
	RenderWriter(&sb, "hello $name", Pairs("name", "world"))
	if sb.String() != "hello world" {
		t.Errorf("RenderWriter: got %q", sb.String())
	}
}

func TestTemplateExecute(t *testing.T) {
	tmpl := Parse("$greeting $name!")
	if got := tmpl.Execute(Pairs("greeting", "Hi", "name", "Go")); got != "Hi Go!" {
		t.Errorf("Execute: got %q", got)
	}
}

func TestTemplateWriteTo(t *testing.T) {
	var sb strings.Builder
	tmpl := Parse("$a+$b")
	tmpl.WriteTo(&sb, Pairs("a", "1", "b", "2"))
	if sb.String() != "1+2" {
		t.Errorf("WriteTo: got %q", sb.String())
	}
}

func TestMissingVarKept(t *testing.T) {
	if got := Render("$missing", map[string]any{}); got != "$missing" {
		t.Errorf("missing: got %q", got)
	}
	if got := Render("$x|upper", map[string]any{}); got != "$x|upper" {
		t.Errorf("missing with transform: got %q", got)
	}
}

func TestRegisterFuncAsModifier(t *testing.T) {
	RegisterFunc("currency", func(v any) string {
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("$%.2f", f)
		}
		return fmt.Sprint(v)
	})
	got := Render("Total: $amount|currency", Pairs("amount", 99.99))
	if got != "Total: $99.99" {
		t.Errorf("currency modifier: got %q", got)
	}
}

func TestRegisterFuncStandalone(t *testing.T) {
	RegisterFunc("greeting", func(_ any) string {
		return "Hello, World"
	})
	// $greeting is not in params — should call the registered function
	got := Render("$greeting!", map[string]any{})
	if got != "Hello, World!" {
		t.Errorf("standalone func: got %q", got)
	}
}

func TestRegisterFuncParamShadowsFunc(t *testing.T) {
	RegisterFunc("shadow", func(_ any) string { return "from-func" })
	// param value takes precedence over registered function
	got := Render("$shadow", Pairs("shadow", "from-param"))
	if got != "from-param" {
		t.Errorf("param should shadow func: got %q", got)
	}
}

func TestRegisterFuncModifierOnMissingVar(t *testing.T) {
	RegisterFunc("shout", func(_ any) string { return "HEY" })
	// $x is missing, |shout is a registered function → call with nil
	got := Render("$x|shout", map[string]any{})
	if got != "HEY" {
		t.Errorf("modifier func on missing var: got %q", got)
	}
}

func TestCacheSize(t *testing.T) {
	SetCacheSize(2)
	Parse("tpl-a-$x")
	Parse("tpl-b-$x")
	Parse("tpl-c-$x") // beyond limit, not cached
	cacheMu.RLock()
	n := len(cache)
	cacheMu.RUnlock()
	if n > 2 {
		t.Errorf("cache exceeded limit: %d entries", n)
	}
	SetCacheSize(0) // disable; should clear
	cacheMu.RLock()
	n = len(cache)
	cacheMu.RUnlock()
	if n != 0 {
		t.Errorf("cache not cleared after SetCacheSize(0): %d entries", n)
	}
	SetCacheSize(1000) // restore default
}

func TestCacheHit(t *testing.T) {
	SetCacheSize(1000)
	src := "cached-$val"
	t1 := Parse(src)
	t2 := Parse(src)
	if t1 != t2 {
		t.Error("expected same *Template pointer from cache")
	}
}

func TestMultipleParams(t *testing.T) {
	got := Render("$a $b", Pairs("a", "1"), Pairs("b", "2"))
	if got != "1 2" {
		t.Errorf("multi params: got %q", got)
	}
}

// ── Function call syntax: $fn(arg1, arg2, ...) ───────────────────────────────

func TestCallTypedIntArgs(t *testing.T) {
	RegisterFunc("add", func(a, b int) int { return a + b })
	// int64 literals coerced to int params
	if got := Render("$add(3, 4)", map[string]any{}); got != "7" {
		t.Errorf("add(3,4): got %q", got)
	}
}

func TestCallTypedStringArgs(t *testing.T) {
	RegisterFunc("greet", func(name string) string { return "Hello, " + name })
	if got := Render(`$greet("World")`, map[string]any{}); got != "Hello, World" {
		t.Errorf("greet: got %q", got)
	}
}

func TestCallVarArgs(t *testing.T) {
	RegisterFunc("join2", func(a, b string) string { return a + b })
	got := Render("$join2($first, $last)", Pairs("first", "foo", "last", "bar"))
	if got != "foobar" {
		t.Errorf("join2 var args: got %q", got)
	}
}

func TestCallMixedArgs(t *testing.T) {
	RegisterFunc("prefix", func(sep, a, b string) string { return a + sep + b })
	got := Render(`$prefix("-", $x, $y)`, Pairs("x", "hello", "y", "world"))
	if got != "hello-world" {
		t.Errorf("prefix mixed args: got %q", got)
	}
}

func TestCallZeroArgs(t *testing.T) {
	RegisterFunc("version", func() string { return "v2.0" })
	if got := Render("ver: $version()", map[string]any{}); got != "ver: v2.0" {
		t.Errorf("zero-arg call: got %q", got)
	}
}

func TestCallVariadicFunc(t *testing.T) {
	RegisterFunc("sum", func(nums ...int) int {
		total := 0
		for _, n := range nums {
			total += n
		}
		return total
	})
	// int64 literals coerced to variadic int elem type
	if got := Render("$sum(1, 2, 3, 4)", map[string]any{}); got != "10" {
		t.Errorf("variadic sum: got %q", got)
	}
}

func TestCallVariadicMixed(t *testing.T) {
	RegisterFunc("cat", func(sep string, parts ...string) string {
		return strings.Join(parts, sep)
	})
	got := Render(`$cat(", ", $a, $b, $c)`, Pairs("a", "one", "b", "two", "c", "three"))
	if got != "one, two, three" {
		t.Errorf("variadic cat: got %q", got)
	}
}

func TestCallFloatArgs(t *testing.T) {
	RegisterFunc("mul", func(a, b float64) float64 { return a * b })
	// float64 literal
	if got := Render("$mul(2.5, 4.0)", map[string]any{}); got != "10" {
		t.Errorf("mul float: got %q", got)
	}
	// int64 literal coerced to float64
	if got := Render("$mul(3, 3)", map[string]any{}); got != "9" {
		t.Errorf("mul int coercion: got %q", got)
	}
}

func TestCallSingleQuotedString(t *testing.T) {
	RegisterFunc("echo", func(s string) string { return s })
	if got := Render("$echo('hello world')", map[string]any{}); got != "hello world" {
		t.Errorf("single-quoted arg: got %q", got)
	}
}

func TestCallDoubleQuotedEscapes(t *testing.T) {
	RegisterFunc("echoStr", func(s string) string { return s })
	if got := Render(`$echoStr("line1\nline2")`, map[string]any{}); got != "line1\nline2" {
		t.Errorf("escape in arg: got %q", got)
	}
}

func TestCallDottedPathArg(t *testing.T) {
	RegisterFunc("upper2", func(s string) string { return strings.ToUpper(s) })
	got := Render("$upper2($user.Name)", Pairs("user", User{Name: "alice", Family: "smith"}))
	if got != "ALICE" {
		t.Errorf("dotted path arg: got %q", got)
	}
}

func TestCallNegativeIntLiteral(t *testing.T) {
	RegisterFunc("neg", func(n int) string { return fmt.Sprintf("%d", n) })
	if got := Render("$neg(-5)", map[string]any{}); got != "-5" {
		t.Errorf("negative literal: got %q", got)
	}
}

func TestCallUnknownFuncEmptyOutput(t *testing.T) {
	// unregistered function — must produce no output, not a placeholder
	got := Render("[${noSuchFn}($x)]", Pairs("x", "v"))
	// ${ is not valid identifier start — stays literal; we test the real case:
	got = Render("[$noSuchFn($x)]", Pairs("x", "v"))
	if got != "[]" {
		t.Errorf("unknown func call: got %q", got)
	}
}

func TestCallFuncReturnsError(t *testing.T) {
	RegisterFunc("failFn", func(s string) (string, error) {
		if s == "bad" {
			return "", errors.New("bad input")
		}
		return "ok:" + s, nil
	})
	// error case → empty output
	if got := Render(`$failFn("bad")`, map[string]any{}); got != "" {
		t.Errorf("error return: got %q, want empty", got)
	}
	// success case
	if got := Render(`$failFn("good")`, map[string]any{}); got != "ok:good" {
		t.Errorf("success return: got %q", got)
	}
}

func TestCallAnyParamType(t *testing.T) {
	// func(v any) string — old-style single-arg still works via call syntax
	RegisterFunc("anyFn", func(v any) string {
		return fmt.Sprintf("<%v>", v)
	})
	got := Render("$anyFn($val)", Pairs("val", 42))
	if got != "<42>" {
		t.Errorf("any param: got %q", got)
	}
}

func TestCallMultiReturn(t *testing.T) {
	RegisterFunc("divmod", func(a, b int) string {
		return fmt.Sprintf("%d/%d", a/b, a%b)
	})
	if got := Render("$divmod(10, 3)", map[string]any{}); got != "3/1" {
		t.Errorf("divmod: got %q", got)
	}
}

func TestCallInsideTemplate(t *testing.T) {
	RegisterFunc("bold", func(s string) string { return "**" + s + "**" })
	got := Render("Name: $bold($name), Age: $age", Pairs("name", "Go", "age", 10))
	if got != "Name: **Go**, Age: 10" {
		t.Errorf("call inside template: got %q", got)
	}
}

func TestCallConsecutive(t *testing.T) {
	RegisterFunc("inc", func(n int) int { return n + 1 })
	got := Render("$inc(1) $inc(2) $inc(3)", map[string]any{})
	if got != "2 3 4" {
		t.Errorf("consecutive calls: got %q", got)
	}
}

func TestCallVarShadowsCall(t *testing.T) {
	// When a template param provides "fn", $fn reads the param value,
	// but $fn(...) still calls the function (different parse paths).
	RegisterFunc("stamp", func() string { return "STAMP" })
	// $stamp as plain var — param wins
	got := Render("$stamp", Pairs("stamp", "from-param"))
	if got != "from-param" {
		t.Errorf("param shadows func standalone: got %q", got)
	}
	// $stamp() as call — always calls the function, param irrelevant
	got = Render("$stamp()", Pairs("stamp", "from-param"))
	if got != "STAMP" {
		t.Errorf("call not shadowed by param: got %q", got)
	}
}

func TestCallBracketPathArgNotACall(t *testing.T) {
	// $arr[0](...) must NOT be treated as a function call because the
	// identifier is not a simple name. The '(' becomes a literal character.
	got := Render("$arr[0](hi)", map[string]any{"arr": []string{"X"}})
	// $arr[0] resolves to "X"; "(hi)" is literal text
	if got != "X(hi)" {
		t.Errorf("bracket path not a call: got %q", got)
	}
}

func TestRegisterFuncPanicOnNonFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when registering non-function")
		}
	}()
	RegisterFunc("bad", "not a function")
}
