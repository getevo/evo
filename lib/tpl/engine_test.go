package tpl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func eng(src string, params ...any) string {
	b := Text(src)
	for _, p := range params {
		b.Set(p)
	}
	ClearCache()
	return b.Render()
}

type point struct{ X, Y int }

// ── Variable interpolation in plain text ──────────────────────────────────────

func TestEngTextSimpleVar(t *testing.T) {
	got := eng("Hello $name!", Pairs("name", "World"))
	if got != "Hello World!" {
		t.Errorf("got %q", got)
	}
}

func TestEngTextDottedPath(t *testing.T) {
	got := eng("$p.X,$p.Y", Pairs("p", point{3, 7}))
	if got != "3,7" {
		t.Errorf("got %q", got)
	}
}

func TestEngTextSliceIndex(t *testing.T) {
	got := eng("$x[0] $x[1] $x[2]", Pairs("x", []string{"a", "b", "c"}))
	if got != "a b c" {
		t.Errorf("got %q", got)
	}
}

func TestEngTextMapStringKey(t *testing.T) {
	m := map[string]int{"test": 10, "test2": 20}
	got := eng(`$m["test"] and $m["test2"]`, Pairs("m", m))
	if got != "10 and 20" {
		t.Errorf("got %q", got)
	}
}

func TestEngTextMapSingleQuoteKey(t *testing.T) {
	m := map[string]string{"key": "val"}
	got := eng("$m['key']", Pairs("m", m))
	if got != "val" {
		t.Errorf("got %q", got)
	}
}

func TestEngTextDollarEscape(t *testing.T) {
	got := eng("price: $$10")
	if got != "price: $10" {
		t.Errorf("got %q", got)
	}
}

func TestEngTextMissingVarKept(t *testing.T) {
	got := eng("$missing", Pairs())
	if got != "$missing" {
		t.Errorf("got %q", got)
	}
}

func TestEngTextNestedPath(t *testing.T) {
	type Sub struct{ P string }
	type Root struct{ Sub Sub }
	got := eng("$st.Sub.P", Pairs("st", Root{Sub: Sub{P: "hello"}}))
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

// ── Echo / print / bare var ───────────────────────────────────────────────────

func TestEngEcho(t *testing.T) {
	got := eng(`<? echo $name ?>`, Pairs("name", "Alice"))
	if got != "Alice" {
		t.Errorf("got %q", got)
	}
}

func TestEngPrint(t *testing.T) {
	got := eng(`<? print $name ?>`, Pairs("name", "Bob"))
	if got != "Bob" {
		t.Errorf("got %q", got)
	}
}

func TestEngBareVar(t *testing.T) {
	got := eng(`<? $name ?>`, Pairs("name", "Carol"))
	if got != "Carol" {
		t.Errorf("got %q", got)
	}
}

func TestEngFuncCallInTag(t *testing.T) {
	RegisterFunc("exclaim", func(s string) string { return s + "!" })
	got := eng(`<? exclaim("hi") ?>`)
	if got != "hi!" {
		t.Errorf("got %q", got)
	}
}

func TestEngFuncCallWithVar(t *testing.T) {
	RegisterFunc("greet2", func(name string) string { return "hello " + name })
	got := eng(`<? greet2($name) ?>`, Pairs("name", "Dave"))
	if got != "hello Dave" {
		t.Errorf("got %q", got)
	}
}

// ── Function call syntax $fn(args) in code tag ────────────────────────────────

func TestEngFuncTagMultiArg(t *testing.T) {
	RegisterFunc("addInts", func(a, b int) int { return a + b })
	got := eng(`<? echo addInts(3, 4) ?>`)
	if got != "7" {
		t.Errorf("got %q", got)
	}
}

// ── Variable assignment and persistence ───────────────────────────────────────

func TestEngAssignPersists(t *testing.T) {
	got := eng(`<? $j = 5 ?>result: $j`)
	if got != "result: 5" {
		t.Errorf("got %q", got)
	}
}

func TestEngAssignArithmetic(t *testing.T) {
	got := eng(`<? $j = $y + $z ?><? echo $j ?>`, Pairs("y", 1, "z", 2))
	if got != "3" {
		t.Errorf("got %q", got)
	}
}

func TestEngMultiStmtSemicolon(t *testing.T) {
	got := eng(`<? $j = $y + $z; echo $j ?>`, Pairs("y", 10, "z", 5))
	if got != "15" {
		t.Errorf("got %q", got)
	}
}

func TestEngMultiStmtNewline(t *testing.T) {
	got := eng("<? \n $j = $y + $z \n echo $j \n?>", Pairs("y", 3, "z", 4))
	if got != "7" {
		t.Errorf("got %q", got)
	}
}

// ── Arithmetic expressions ────────────────────────────────────────────────────

func TestEngArithAdd(t *testing.T) {
	got := eng(`<? echo $a + $b ?>`, Pairs("a", 10, "b", 3))
	if got != "13" {
		t.Errorf("got %q", got)
	}
}

func TestEngArithSub(t *testing.T) {
	got := eng(`<? echo $a - $b ?>`, Pairs("a", 10, "b", 3))
	if got != "7" {
		t.Errorf("got %q", got)
	}
}

func TestEngArithMul(t *testing.T) {
	got := eng(`<? echo $a * $b ?>`, Pairs("a", 4, "b", 5))
	if got != "20" {
		t.Errorf("got %q", got)
	}
}

func TestEngArithDiv(t *testing.T) {
	got := eng(`<? echo $a / $b ?>`, Pairs("a", 10, "b", 4))
	// 10/4 = 2.5 (float)
	if got != "2.5" {
		t.Errorf("got %q", got)
	}
}

func TestEngArithMod(t *testing.T) {
	got := eng(`<? echo $a % $b ?>`, Pairs("a", 10, "b", 3))
	if got != "1" {
		t.Errorf("got %q", got)
	}
}

func TestEngStringConcat(t *testing.T) {
	got := eng(`<? echo $first . " " . $last ?>`, Pairs("first", "John", "last", "Doe"))
	if got != "John Doe" {
		t.Errorf("got %q", got)
	}
}

// ── Comparisons and logic ─────────────────────────────────────────────────────

func TestEngIfTrue(t *testing.T) {
	got := eng(`<? if($x > 5){ ?>yes<? } ?>`, Pairs("x", 10))
	if got != "yes" {
		t.Errorf("got %q", got)
	}
}

func TestEngIfFalse(t *testing.T) {
	got := eng(`<? if($x > 5){ ?>yes<? } ?>`, Pairs("x", 3))
	if got != "" {
		t.Errorf("got %q", got)
	}
}

func TestEngIfElse(t *testing.T) {
	got := eng(`<? if($x > 5){ ?>big<? }else{ ?>small<? } ?>`, Pairs("x", 2))
	if got != "small" {
		t.Errorf("got %q", got)
	}
}

func TestEngIfElseIf(t *testing.T) {
	tpl := `<? if($x == 1){ ?>one<? }else if($x == 2){ ?>two<? }else{ ?>other<? } ?>`
	if got := eng(tpl, Pairs("x", 1)); got != "one" {
		t.Errorf("x=1: got %q", got)
	}
	if got := eng(tpl, Pairs("x", 2)); got != "two" {
		t.Errorf("x=2: got %q", got)
	}
	if got := eng(tpl, Pairs("x", 9)); got != "other" {
		t.Errorf("x=9: got %q", got)
	}
}

func TestEngLogicalAnd(t *testing.T) {
	got := eng(`<? if($a > 0 && $b > 0){ ?>yes<? } ?>`, Pairs("a", 1, "b", 2))
	if got != "yes" {
		t.Errorf("got %q", got)
	}
	got = eng(`<? if($a > 0 && $b > 0){ ?>yes<? } ?>`, Pairs("a", 1, "b", -1))
	if got != "" {
		t.Errorf("got %q", got)
	}
}

func TestEngLogicalOr(t *testing.T) {
	got := eng(`<? if($a > 0 || $b > 0){ ?>yes<? } ?>`, Pairs("a", -1, "b", 1))
	if got != "yes" {
		t.Errorf("got %q", got)
	}
}

func TestEngNot(t *testing.T) {
	got := eng(`<? if(!$flag){ ?>off<? } ?>`, Pairs("flag", false))
	if got != "off" {
		t.Errorf("got %q", got)
	}
}

func TestEngEqualityString(t *testing.T) {
	got := eng(`<? if($v == "z"){ ?>yes<? } ?>`, Pairs("v", "z"))
	if got != "yes" {
		t.Errorf("got %q", got)
	}
}

// ── For-range loops ───────────────────────────────────────────────────────────

func TestEngForRangeSlice(t *testing.T) {
	got := eng(`<? for($i,$v := range $x){ ?>$i:$v <? } ?>`,
		Pairs("x", []string{"a", "b", "c"}))
	// Map iteration order — use Contains for map, but slice is ordered
	if got != "0:a 1:b 2:c " {
		t.Errorf("got %q", got)
	}
}

func TestEngForRangeSingleVar(t *testing.T) {
	got := eng(`<? for($v := range $x){ ?>$v<? } ?>`,
		Pairs("x", []int{1, 2, 3}))
	if got != "123" {
		t.Errorf("got %q", got)
	}
}

func TestEngForRangeMap(t *testing.T) {
	m := map[string]int{"a": 1}
	got := eng(`<? for($k,$v := range $m){ ?>$k=$v<? } ?>`, Pairs("m", m))
	if got != "a=1" {
		t.Errorf("got %q", got)
	}
}

func TestEngForRangeMapMultipleKeys(t *testing.T) {
	m := map[string]string{"hello": "world"}
	got := eng(`<? for($k,$v := range $m){ ?>[$k:$v]<? } ?>`, Pairs("m", m))
	if got != "[hello:world]" {
		t.Errorf("got %q", got)
	}
}

// ── C-style for loop ──────────────────────────────────────────────────────────

func TestEngForC(t *testing.T) {
	got := eng(`<? for($i=0; $i < 3; $i++){ ?>$i<? } ?>`)
	if got != "012" {
		t.Errorf("got %q", got)
	}
}

func TestEngForCDecrement(t *testing.T) {
	got := eng(`<? for($i=2; $i >= 0; $i--){ ?>$i<? } ?>`)
	if got != "210" {
		t.Errorf("got %q", got)
	}
}

func TestEngForCSum(t *testing.T) {
	got := eng(`<? $s=0 ?><? for($i=1; $i <= 5; $i++){ ?><? $s = $s + $i ?><? } ?><? echo $s ?>`)
	if got != "15" {
		t.Errorf("got %q", got)
	}
}

// ── Nested structures ─────────────────────────────────────────────────────────

func TestEngNestedForIf(t *testing.T) {
	tpl := `<? for($i,$v := range $x){ ?><? if($i > 0){ ?>$v<? } ?><? } ?>`
	got := eng(tpl, Pairs("x", []string{"a", "b", "c"}))
	if got != "bc" {
		t.Errorf("got %q", got)
	}
}

func TestEngNestedForInFor(t *testing.T) {
	tpl := `<? for($i,$a := range $rows){ ?><? for($j,$b := range $a){ ?>$b<? } ?>,<? } ?>`
	got := eng(tpl, Pairs("rows", [][]int{{1, 2}, {3, 4}}))
	if got != "12,34," {
		t.Errorf("got %q", got)
	}
}

func TestEngNestedIfInIf(t *testing.T) {
	tpl := `<? if($a){ ?><? if($b){ ?>both<? }else{ ?>only-a<? } ?><? }else{ ?>none<? } ?>`
	if got := eng(tpl, Pairs("a", true, "b", true)); got != "both" {
		t.Errorf("both: got %q", got)
	}
	if got := eng(tpl, Pairs("a", true, "b", false)); got != "only-a" {
		t.Errorf("only-a: got %q", got)
	}
	if got := eng(tpl, Pairs("a", false, "b", false)); got != "none" {
		t.Errorf("none: got %q", got)
	}
}

// ── Comments ──────────────────────────────────────────────────────────────────

func TestEngCommentInBlock(t *testing.T) {
	got := eng("<? // this is a comment\n echo $name ?>", Pairs("name", "X"))
	if got != "X" {
		t.Errorf("got %q", got)
	}
}

func TestEngCommentOnly(t *testing.T) {
	got := eng("<? // entire block is a comment ?>text")
	if got != "text" {
		t.Errorf("got %q", got)
	}
}

// ── Increment / Decrement ─────────────────────────────────────────────────────

func TestEngIncDecStmt(t *testing.T) {
	got := eng(`<? $n = 0 ?><? $n++ ?><? $n++ ?><? echo $n ?>`)
	if got != "2" {
		t.Errorf("got %q", got)
	}
}

// ── Builder API ───────────────────────────────────────────────────────────────

func TestEngBuilderText(t *testing.T) {
	got := Text("Hello $name").Set(Pairs("name", "World")).Render()
	if got != "Hello World" {
		t.Errorf("got %q", got)
	}
}

func TestEngBuilderMultiSet(t *testing.T) {
	got := Text("$a $b").Set(Pairs("a", "1")).Set(Pairs("b", "2")).Render()
	if got != "1 2" {
		t.Errorf("got %q", got)
	}
}

func TestEngBuilderRenderWriter(t *testing.T) {
	var sb strings.Builder
	Text("$x").Set(Pairs("x", "42")).RenderWriter(&sb)
	if sb.String() != "42" {
		t.Errorf("got %q", sb.String())
	}
}

// ── File loading ──────────────────────────────────────────────────────────────

func TestEngFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tmpl.html")
	_ = os.WriteFile(path, []byte("Hello $name"), 0644)

	ClearCache()
	got := File(path).Set(Pairs("name", "File")).Render()
	if got != "Hello File" {
		t.Errorf("got %q", got)
	}
}

func TestEngFileMissing(t *testing.T) {
	ClearCache()
	got := File("/nonexistent/path/tmpl.html").Render()
	if got != "" {
		t.Errorf("missing file should produce empty, got %q", got)
	}
}

// ── Include ───────────────────────────────────────────────────────────────────

func TestEngInclude(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub.html")
	_ = os.WriteFile(sub, []byte("included:$val"), 0644)

	tpl := `before <? include("` + sub + `") ?> after`
	got := eng(tpl, Pairs("val", "X"))
	if got != "before included:X after" {
		t.Errorf("got %q", got)
	}
}

func TestEngRequire(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "req.html")
	_ = os.WriteFile(sub, []byte("[required]"), 0644)

	tpl := `<? require("` + sub + `") ?>`
	got := eng(tpl)
	if got != "[required]" {
		t.Errorf("got %q", got)
	}
}

func TestEngIncludeMissing(t *testing.T) {
	// Missing file: silently skip (both require and include behave the same)
	got := eng(`before<? include("/no/such/file") ?>after`)
	if got != "beforeafter" {
		t.Errorf("got %q", got)
	}
}

// ── Edge cases ────────────────────────────────────────────────────────────────

func TestEngEmptyBlock(t *testing.T) {
	got := eng("a<? ?>b")
	if got != "ab" {
		t.Errorf("got %q", got)
	}
}

func TestEngNoTags(t *testing.T) {
	got := eng("plain text $name", Pairs("name", "Go"))
	if got != "plain text Go" {
		t.Errorf("got %q", got)
	}
}

func TestEngOnlyTag(t *testing.T) {
	got := eng(`<? echo "hello" ?>`)
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestEngExprInTagWithVar(t *testing.T) {
	got := eng(`<? echo $x[1] ?>`, Pairs("x", []string{"a", "b", "c"}))
	if got != "b" {
		t.Errorf("got %q", got)
	}
}

func TestEngMapAccessInTag(t *testing.T) {
	m := map[string]int{"score": 99}
	got := eng(`<? echo $m["score"] ?>`, Pairs("m", m))
	if got != "99" {
		t.Errorf("got %q", got)
	}
}

func TestEngComplexTemplate(t *testing.T) {
	RegisterFunc("upper3", func(s string) string {
		result := ""
		for _, r := range s {
			if r >= 'a' && r <= 'z' {
				result += string(r - 32)
			} else {
				result += string(r)
			}
		}
		return result
	})

	tpl := strings.TrimSpace(`
Hello $name, welcome to <? upper3("my site") ?>.
Items:
<? for($i,$v := range $items){ ?>
  - $i: $v
<? } ?>
Total: $total`)

	want := strings.TrimSpace(`
Hello Alice, welcome to MY SITE.
Items:
  - 0: apple
  - 1: banana
  - 2: cherry
Total: 3`)

	got := strings.TrimSpace(eng(tpl, Pairs(
		"name", "Alice",
		"items", []string{"apple", "banana", "cherry"},
		"total", 3,
	)))

	if got != want {
		t.Errorf("complex:\ngot  %q\nwant %q", got, want)
	}
}

func TestEngVarAssignedInLoop(t *testing.T) {
	// Variables assigned inside a loop body should be visible in subsequent blocks.
	tpl := `<? for($i,$v := range $x){ ?><? $last = $v ?><? } ?>last: $last`
	got := eng(tpl, Pairs("x", []string{"a", "b", "c"}))
	if got != "last: c" {
		t.Errorf("got %q", got)
	}
}

func TestEngBoolExpr(t *testing.T) {
	got := eng(`<? if($x == true){ ?>yes<? } ?>`, Pairs("x", true))
	if got != "yes" {
		t.Errorf("got %q", got)
	}
}

func TestEngUnaryMinus(t *testing.T) {
	got := eng(`<? echo -$n ?>`, Pairs("n", 5))
	if got != "-5" {
		t.Errorf("got %q", got)
	}
}

func TestEngStringLiteralInEcho(t *testing.T) {
	got := eng(`<? echo "hello world" ?>`)
	if got != "hello world" {
		t.Errorf("got %q", got)
	}
}

func TestEngIntLiteralInEcho(t *testing.T) {
	got := eng(`<? echo 42 ?>`)
	if got != "42" {
		t.Errorf("got %q", got)
	}
}

func TestEngFloatLiteralInEcho(t *testing.T) {
	got := eng(`<? echo 3.14 ?>`)
	if got != "3.14" {
		t.Errorf("got %q", got)
	}
}

// ── Keywords: true / false / null ─────────────────────────────────────────────

func TestEngTrueFalseKeyword(t *testing.T) {
	got := eng(`<? $x = true ?><? if($x){ ?>yes<? } ?>`)
	if got != "yes" {
		t.Errorf("got %q", got)
	}
	got = eng(`<? $x = false ?><? if($x){ ?>yes<? }else{ ?>no<? } ?>`)
	if got != "no" {
		t.Errorf("got %q", got)
	}
}

func TestEngNullKeyword(t *testing.T) {
	got := eng(`<? $x = null ?><? if($x == null){ ?>nil<? } ?>`)
	if got != "nil" {
		t.Errorf("got %q", got)
	}
}

func TestEngNilKeyword(t *testing.T) {
	got := eng(`<? if($missing == nil){ ?>empty<? } ?>`)
	if got != "empty" {
		t.Errorf("got %q", got)
	}
}

// ── isset ─────────────────────────────────────────────────────────────────────

func TestEngIsset(t *testing.T) {
	got := eng(`<? if(isset($name)){ ?>yes<? }else{ ?>no<? } ?>`, Pairs("name", "Alice"))
	if got != "yes" {
		t.Errorf("isset present: got %q", got)
	}
	got = eng(`<? if(isset($missing)){ ?>yes<? }else{ ?>no<? } ?>`)
	if got != "no" {
		t.Errorf("isset absent: got %q", got)
	}
}

// ── Null-coalescing ?? ────────────────────────────────────────────────────────

func TestEngNullCoalescing(t *testing.T) {
	got := eng(`<? echo $name ?? "guest" ?>`, Pairs("name", "Alice"))
	if got != "Alice" {
		t.Errorf("?? with value: got %q", got)
	}
	got = eng(`<? echo $missing ?? "guest" ?>`)
	if got != "guest" {
		t.Errorf("?? fallback: got %q", got)
	}
}

// ── Ternary ?: ────────────────────────────────────────────────────────────────

func TestEngTernary(t *testing.T) {
	got := eng(`<? echo $x > 5 ? "big" : "small" ?>`, Pairs("x", 10))
	if got != "big" {
		t.Errorf("ternary true: got %q", got)
	}
	got = eng(`<? echo $x > 5 ? "big" : "small" ?>`, Pairs("x", 3))
	if got != "small" {
		t.Errorf("ternary false: got %q", got)
	}
}

// ── Compound assignment +=, -=, *=, /= ───────────────────────────────────────

func TestEngCompoundAssign(t *testing.T) {
	got := eng(`<? $n = 10 ?><? $n += 5 ?><? echo $n ?>`)
	if got != "15" {
		t.Errorf("+= : got %q", got)
	}
	got = eng(`<? $n = 10 ?><? $n -= 3 ?><? echo $n ?>`)
	if got != "7" {
		t.Errorf("-= : got %q", got)
	}
	got = eng(`<? $n = 4 ?><? $n *= 3 ?><? echo $n ?>`)
	if got != "12" {
		t.Errorf("*= : got %q", got)
	}
	got = eng(`<? $n = 10 ?><? $n /= 4 ?><? echo $n ?>`)
	if got != "2.5" {
		t.Errorf("/= : got %q", got)
	}
}

func TestEngCompoundAssignInFor(t *testing.T) {
	got := eng(`<? for($i=0; $i < 10; $i+=3){ ?>$i<? } ?>`)
	if got != "0369" {
		t.Errorf("for += step: got %q", got)
	}
}

// ── break and continue ────────────────────────────────────────────────────────

func TestEngBreak(t *testing.T) {
	// break must be in its own tag; block bodies span separate <? ?> pairs
	got := eng(`<? for($i := range 5){ ?><? if($i == 3){ ?><? break ?><? } ?>$i<? } ?>`)
	if got != "012" {
		t.Errorf("break: got %q", got)
	}
}

func TestEngContinue(t *testing.T) {
	got := eng(`<? for($i := range 5){ ?><? if($i == 2){ ?><? continue ?><? } ?>$i<? } ?>`)
	if got != "0134" {
		t.Errorf("continue: got %q", got)
	}
}

func TestEngBreakNested(t *testing.T) {
	// break exits only the inner loop
	tpl := `<? for($i := range 3){ ?><? for($j := range 3){ ?><? if($j == 1){ ?><? break ?><? } ?>$j<? } ?>|<? } ?>`
	got := eng(tpl)
	if got != "0|0|0|" {
		t.Errorf("nested break: got %q", got)
	}
}

// ── switch / case / default ───────────────────────────────────────────────────

func TestEngSwitch(t *testing.T) {
	tpl := `<? switch($x){ ?><? case "a": ?>A<? case "b": ?>B<? default: ?>other<? } ?>`
	if got := eng(tpl, Pairs("x", "a")); got != "A" {
		t.Errorf("switch a: got %q", got)
	}
	if got := eng(tpl, Pairs("x", "b")); got != "B" {
		t.Errorf("switch b: got %q", got)
	}
	if got := eng(tpl, Pairs("x", "z")); got != "other" {
		t.Errorf("switch default: got %q", got)
	}
}

func TestEngSwitchMultiValue(t *testing.T) {
	tpl := `<? switch($x){ ?><? case "a", "b": ?>AB<? case "c": ?>C<? } ?>`
	if got := eng(tpl, Pairs("x", "a")); got != "AB" {
		t.Errorf("multi-val a: got %q", got)
	}
	if got := eng(tpl, Pairs("x", "b")); got != "AB" {
		t.Errorf("multi-val b: got %q", got)
	}
	if got := eng(tpl, Pairs("x", "c")); got != "C" {
		t.Errorf("multi-val c: got %q", got)
	}
}

func TestEngSwitchNoDefault(t *testing.T) {
	tpl := `<? switch($x){ ?><? case 1: ?>one<? } ?>`
	if got := eng(tpl, Pairs("x", 2)); got != "" {
		t.Errorf("switch no match no default: got %q", got)
	}
}

func TestEngSwitchBreak(t *testing.T) {
	// break inside a switch exits the switch, not the enclosing loop
	tpl := `<? for($i := range 3){ ?><? switch($i){ ?><? case 1: ?><? break ?><? default: ?>$i<? } ?><? } ?>`
	got := eng(tpl)
	// i=0 → default → "0"; i=1 → case 1 → break exits switch; i=2 → default → "2"
	if got != "02" {
		t.Errorf("switch break: got %q", got)
	}
}

// ── while-style loop ──────────────────────────────────────────────────────────

func TestEngWhileLoop(t *testing.T) {
	got := eng(`<? $i = 0 ?><? for($i < 3){ ?>$i<? $i++ ?><? } ?>`)
	if got != "012" {
		t.Errorf("while loop: got %q", got)
	}
}

func TestEngWhileBreak(t *testing.T) {
	got := eng(`<? $i = 0 ?><? for($i < 10){ ?><? if($i == 3){ ?><? break ?><? } ?>$i<? $i++ ?><? } ?>`)
	if got != "012" {
		t.Errorf("while break: got %q", got)
	}
}

// ── for range over integer ────────────────────────────────────────────────────

func TestEngForRangeInt(t *testing.T) {
	got := eng(`<? for($v := range 4){ ?>$v<? } ?>`)
	if got != "0123" {
		t.Errorf("range int: got %q", got)
	}
}

func TestEngForRangeIntKeyVal(t *testing.T) {
	got := eng(`<? for($k,$v := range 3){ ?>$k:$v <? } ?>`)
	if got != "0:0 1:1 2:2 " {
		t.Errorf("range int key,val: got %q", got)
	}
}

// ── Built-in functions ────────────────────────────────────────────────────────

func TestEngBuiltinLen(t *testing.T) {
	got := eng(`<? echo len($x) ?>`, Pairs("x", []int{1, 2, 3, 4}))
	if got != "4" {
		t.Errorf("len slice: got %q", got)
	}
	got = eng(`<? echo len($s) ?>`, Pairs("s", "hello"))
	if got != "5" {
		t.Errorf("len string: got %q", got)
	}
}

func TestEngBuiltinUpper(t *testing.T) {
	got := eng(`<? echo upper($s) ?>`, Pairs("s", "hello"))
	if got != "HELLO" {
		t.Errorf("upper: got %q", got)
	}
}

func TestEngBuiltinLower(t *testing.T) {
	got := eng(`<? echo lower($s) ?>`, Pairs("s", "WORLD"))
	if got != "world" {
		t.Errorf("lower: got %q", got)
	}
}

func TestEngBuiltinTrim(t *testing.T) {
	got := eng(`<? echo trim($s) ?>`, Pairs("s", "  hi  "))
	if got != "hi" {
		t.Errorf("trim: got %q", got)
	}
}

func TestEngBuiltinReplace(t *testing.T) {
	got := eng(`<? echo replace($s, "world", "Go") ?>`, Pairs("s", "hello world"))
	if got != "hello Go" {
		t.Errorf("replace: got %q", got)
	}
}

func TestEngBuiltinContains(t *testing.T) {
	got := eng(`<? if(contains($s, "lo")){ ?>yes<? } ?>`, Pairs("s", "hello"))
	if got != "yes" {
		t.Errorf("contains: got %q", got)
	}
}

func TestEngBuiltinSplit(t *testing.T) {
	got := eng(`<? $parts = split($s, ",") ?><? for($v := range $parts){ ?>$v|<? } ?>`, Pairs("s", "a,b,c"))
	if got != "a|b|c|" {
		t.Errorf("split: got %q", got)
	}
}

func TestEngBuiltinJoinAny(t *testing.T) {
	got := eng(`<? echo joinAny($x, "-") ?>`, Pairs("x", []string{"a", "b", "c"}))
	if got != "a-b-c" {
		t.Errorf("joinAny: got %q", got)
	}
}

func TestEngBuiltinSprintf(t *testing.T) {
	got := eng(`<? echo sprintf("%.2f", $n) ?>`, Pairs("n", 3.14159))
	if got != "3.14" {
		t.Errorf("sprintf: got %q", got)
	}
}

func TestEngBuiltinMath(t *testing.T) {
	if got := eng(`<? echo abs(-5) ?>`); got != "5" {
		t.Errorf("abs: got %q", got)
	}
	if got := eng(`<? echo floor(3.7) ?>`); got != "3" {
		t.Errorf("floor: got %q", got)
	}
	if got := eng(`<? echo ceil(3.2) ?>`); got != "4" {
		t.Errorf("ceil: got %q", got)
	}
	if got := eng(`<? echo round(3.5) ?>`); got != "4" {
		t.Errorf("round: got %q", got)
	}
	if got := eng(`<? echo min(3, 7) ?>`); got != "3" {
		t.Errorf("min: got %q", got)
	}
	if got := eng(`<? echo max(3, 7) ?>`); got != "7" {
		t.Errorf("max: got %q", got)
	}
	if got := eng(`<? echo pow(2, 10) ?>`); got != "1024" {
		t.Errorf("pow: got %q", got)
	}
}

func TestEngBuiltinKeys(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	got := eng(`<? for($k := range keys($m)){ ?>$k<? } ?>`, Pairs("m", m))
	if got != "abc" { // keys returns sorted
		t.Errorf("keys: got %q", got)
	}
}

func TestEngBuiltinFirstLast(t *testing.T) {
	got := eng(`<? echo first($x) ?>`, Pairs("x", []string{"a", "b", "c"}))
	if got != "a" {
		t.Errorf("first: got %q", got)
	}
	got = eng(`<? echo last($x) ?>`, Pairs("x", []string{"a", "b", "c"}))
	if got != "c" {
		t.Errorf("last: got %q", got)
	}
}

func TestEngBuiltinDefault(t *testing.T) {
	got := eng(`<? echo default($name, "guest") ?>`, Pairs("name", "Alice"))
	if got != "Alice" {
		t.Errorf("default with value: got %q", got)
	}
	got = eng(`<? echo default($missing, "guest") ?>`)
	if got != "guest" {
		t.Errorf("default missing: got %q", got)
	}
}

func TestEngBuiltinInt(t *testing.T) {
	got := eng(`<? echo int("42") ?>`)
	if got != "42" {
		t.Errorf("int: got %q", got)
	}
}

func TestEngBuiltinStr(t *testing.T) {
	got := eng(`<? echo str(99) ?>`)
	if got != "99" {
		t.Errorf("str: got %q", got)
	}
}

func TestEngBuiltinCoalesce(t *testing.T) {
	got := eng(`<? echo coalesce($a, $b, "fallback") ?>`, Pairs("b", "found"))
	if got != "found" {
		t.Errorf("coalesce: got %q", got)
	}
}

// ── File cache mtime invalidation ────────────────────────────────────────────

func TestEngFileCacheMtime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tmpl.html")
	_ = os.WriteFile(path, []byte("v1"), 0644)

	ClearCache()
	got := File(path).Render()
	if got != "v1" {
		t.Errorf("initial: got %q", got)
	}

	// Modify file (bump mtime by writing new content)
	_ = os.WriteFile(path, []byte("v2"), 0644)
	// Touch the file to ensure a different mtime even on fast filesystems
	future := time.Now().Add(2 * time.Second)
	_ = os.Chtimes(path, future, future)

	got = File(path).Render()
	if got != "v2" {
		t.Errorf("after update: got %q", got)
	}
}
