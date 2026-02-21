package tpl

import (
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
