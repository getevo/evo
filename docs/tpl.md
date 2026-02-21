# Template Library (`lib/tpl`)

A dual-engine template rendering library. The **simple engine** handles lightweight `$variable` substitution with pipe modifiers. The **full engine** supports PHP-style `<? ?>` code blocks with expressions, control flow, loops, and the complete built-in function library.

```go
import "github.com/getevo/evo/v2/lib/tpl"
```

---

## Table of Contents

1. [Simple Engine](#simple-engine)
   - [Variable Substitution](#variable-substitution)
   - [Pipe Modifiers](#pipe-modifiers)
   - [Filter Chaining](#filter-chaining)
   - [Function Calls](#function-calls)
   - [Missing Variable Fallback](#missing-variable-fallback)
   - [Dollar Escape](#dollar-escape)
   - [Multiple Data Sources](#multiple-data-sources)
2. [Full Engine](#full-engine)
   - [Code Blocks](#code-blocks)
   - [Variable Interpolation in Text](#variable-interpolation-in-text)
   - [Statements](#statements)
   - [Expressions](#expressions)
   - [Array and Map Literals](#array-and-map-literals)
   - [Control Flow](#control-flow)
   - [Loops](#loops)
   - [Loop Metadata](#loop-metadata)
   - [Include / Require](#include--require)
3. [Built-in Functions](#built-in-functions)
4. [Custom Functions](#custom-functions)
5. [API Reference](#api-reference)

---

## Simple Engine

The simple engine processes plain text containing `$variable` placeholders. No code blocks are needed — variables are replaced wherever they appear.

### Variable Substitution

Supported path forms:

| Syntax              | Description                          |
|---------------------|--------------------------------------|
| `$name`             | Plain variable                       |
| `$obj.Field`        | Struct field or map key (dot access) |
| `$arr[0]`           | Integer index                        |
| `$arr["key"]`       | String map key                       |
| `$arr[0].Field`     | Chained access                       |
| `$obj.Sub.Field`    | Deeply nested field                  |

```go
// Structs
type User struct {
    Name  string
    Email string
}
u := User{Name: "Alice", Email: "alice@example.com"}
tpl.Render("Hello $Name, your email is $Email", u)
// → "Hello Alice, your email is alice@example.com"

// Maps
data := map[string]any{"city": "Paris", "country": "France"}
tpl.Render("$city, $country", data)
// → "Paris, France"

// Nested struct fields
type Address struct{ Street string }
type Person struct{ Name string; Addr Address }
p := Person{Name: "Bob", Addr: Address{Street: "Main St"}}
tpl.Render("$Name lives at $Addr.Street", p)
// → "Bob lives at Main St"

// Array indexing
items := map[string]any{"tags": []string{"go", "tpl", "fast"}}
tpl.Render("First tag: $tags[0], second: $tags[1]", items)
// → "First tag: go, second: tpl"

// Chained access: array of maps
data := map[string]any{
    "users": []map[string]any{
        {"name": "Alice"},
        {"name": "Bob"},
    },
}
tpl.Render("$users[0][name] and $users[1][name]", data)
// → "Alice and Bob"
```

### Pipe Modifiers

Append `|modifier` to a variable to transform its value on output.

**Built-in transforms:**

| Modifier | Description                              |
|----------|------------------------------------------|
| `upper`  | Convert to UPPERCASE                     |
| `lower`  | Convert to lowercase                     |
| `title`  | Title Case                               |
| `trim`   | Strip leading and trailing whitespace    |
| `html`   | HTML-escape (`<`, `>`, `&`, `"`, `'`)   |
| `url`    | URL query-encode                         |
| `json`   | JSON-encode the value                    |

```go
tpl.Render("$name|upper", tpl.Pairs("name", "alice"))
// → "ALICE"

tpl.Render("$title|title", tpl.Pairs("title", "hello world"))
// → "Hello World"

tpl.Render("$comment|html", tpl.Pairs("comment", "<script>alert(1)</script>"))
// → "&lt;script&gt;alert(1)&lt;/script&gt;"

tpl.Render("$query|url", tpl.Pairs("query", "hello world"))
// → "hello+world"

tpl.Render("$data|json", tpl.Pairs("data", map[string]any{"k": 1}))
// → `{"k":1}`

// Registered functions also work as modifiers (see Custom Functions)
tpl.RegisterFunc("exclaim", func(s string) string { return s + "!" })
tpl.Render("$greeting|exclaim", tpl.Pairs("greeting", "Hello"))
// → "Hello!"
```

### Filter Chaining

Chain multiple modifiers with `|`. They are applied left to right.

```go
tpl.Render("$name|trim|upper", tpl.Pairs("name", "  alice  "))
// → "ALICE"

tpl.Render("$name|trim|lower|title", tpl.Pairs("name", "  JOHN DOE  "))
// → "John Doe"

tpl.Render("$items|count", tpl.Pairs("items", []int{1, 2, 3, 4, 5}))
// → "5"  (using the registered count function)

// Chain a builtin and a custom function
tpl.RegisterFunc("slug", func(s string) string {
    return strings.ReplaceAll(strings.ToLower(s), " ", "-")
})
tpl.Render("$title|trim|slug", tpl.Pairs("title", "  My Blog Post  "))
// → "my-blog-post"
```

### Function Calls

Call a registered function with arguments using `$fn(args...)`. Arguments can be `$variables`, string literals, or numeric literals.

```go
tpl.RegisterFunc("greet", func(name, lang string) string {
    if lang == "es" { return "Hola, " + name + "!" }
    return "Hello, " + name + "!"
})

tpl.Render(`$greet($name, "es")`, tpl.Pairs("name", "Maria"))
// → "Hola, Maria!"

tpl.Render(`$greet("World", "en")`)
// → "Hello, World!"

// String escapes in double-quoted args:  \n \t \" \\
tpl.Render(`$sprintf("%-10s %d", $name, $score)`,
    tpl.Pairs("name", "Alice", "score", 42))
// → "Alice      42"
```

### Missing Variable Fallback

When a variable is not found, the modifier string is used as a literal default value (if it doesn't match any known modifier name or registered function).

```go
tpl.Render("$nickname|Anonymous", map[string]any{})
// → "Anonymous"  (nickname not set → use literal "Anonymous")

tpl.Render("Dear $title|Sir,", map[string]any{})
// → "Dear Sir,"
```

Built-in transforms on missing variables keep the placeholder unchanged:

```go
tpl.Render("$name|upper", map[string]any{})
// → "$name|upper"  (kept as-is because upper is a known transform)
```

### Dollar Escape

Use `$$` to output a literal `$`.

```go
tpl.Render("Price: $$price", tpl.Pairs("price", 9.99))
// → "Price: $price"  (not interpolated — $$ became $, then "price" is literal)

tpl.Render("Cost: $$$amount", tpl.Pairs("amount", "5"))
// → "Cost: $5"  ($$ → $, then $amount → "5")
```

### Multiple Data Sources

Pass multiple structs, maps, or slices. Variables are looked up in each source in order; the first match wins.

```go
type App struct{ Name, Version string }
type User struct{ Name, Role string }

app := App{Name: "MyApp", Version: "2.0"}
user := User{Name: "Alice", Role: "admin"}

tpl.Render("$app.Name $app.Version — logged in as $user.Name ($user.Role)", app, user)
// Works but requires "app" and "user" keys

// OR use tpl.Pairs to mix:
tpl.Render("Welcome $name to $app", tpl.Pairs("name", "Alice", "app", "MyApp"))
// → "Welcome Alice to MyApp"
```

### `tpl.Pairs`

Helper to build a `map[string]any` from flat alternating key/value pairs.

```go
data := tpl.Pairs("name", "Alice", "age", 30, "active", true)
// → map[string]any{"name": "Alice", "age": 30, "active": true}

tpl.Render("$name is $age years old", data)
// → "Alice is 30 years old"
```

---

## Full Engine

The full engine parses templates with `<? ?>` code blocks. Text outside code blocks is output verbatim (with `$var` interpolation). Code blocks contain statements, control flow, loops, and expressions.

### Code Blocks

Everything inside `<? ?>` is executed as code. Text outside is printed as-is with `$var` substitution.

```
Hello <? echo $name ?>! Today is <? echo date("2006-01-02", now()) ?>.
```

Use the `Builder` API (recommended) or the lower-level `CompileEngine`:

```go
// Builder from inline string
result := tpl.Text(`Hello <? echo $name ?>!`).Set(tpl.Pairs("name", "Alice")).Render()
// → "Hello Alice!"

// Builder from file (auto-cached, mtime-invalidated)
result := tpl.File("views/page.html").Set(user).Set(app).Render()

// Write to an io.Writer
tpl.File("layout.html").Set(data).RenderWriter(w)

// Lower-level
engine := tpl.CompileEngine(src)
ctx := tpl.Pairs("x", 42)
result := tpl.RenderText(src, ctx)

// Convenience helpers
tpl.RenderText(`<? echo $x * 2 ?>`, tpl.Pairs("x", 21))  // → "42"
tpl.RenderFile("page.html", tpl.Pairs("title", "Home"))
```

### Variable Interpolation in Text

`$variable` references in plain text (outside `<? ?>`) resolve against the current context, with the same path syntax as the simple engine.

```
<? $user = {name: "Alice", score: 99} ?>
Player: $user.name  Score: $user.score
Greeting: $user["name"]
```

Missing variables are kept as-is: `$unknown` → `$unknown`.

### Statements

#### Output: `echo` / `print`

```
<? echo "Hello, " . $name . "!" ?>
<? print $total * 1.2 ?>
<? $greeting ?>         <!-- bare variable: also outputs -->
<? upper($name) ?>      <!-- bare function call: also outputs -->
```

#### Assignment

```
<? $x = 42 ?>
<? $msg := "Hello" ?>   <!-- := and = are equivalent -->
<? $sum = $a + $b ?>
```

#### Compound Assignment

```
<? $count += 1 ?>
<? $total -= $discount ?>
<? $price *= 1.1 ?>
<? $half /= 2 ?>
```

#### Increment / Decrement

```
<? $i++ ?>
<? $j-- ?>
```

### Expressions

#### Literals

```
<? echo 42 ?>           <!-- integer -->
<? echo 3.14 ?>         <!-- float -->
<? echo "hello" ?>      <!-- double-quoted string (supports \n \t \" \\) -->
<? echo 'world' ?>      <!-- single-quoted string (no escape sequences) -->
<? echo true ?>
<? echo false ?>
<? echo null ?>         <!-- or nil -->
```

#### Arithmetic

```
<? echo 2 + 3 ?>        <!-- 5 -->
<? echo 10 - 4 ?>       <!-- 6 -->
<? echo 3 * 7 ?>        <!-- 21 -->
<? echo 15 / 4 ?>       <!-- 3.75 -->
<? echo 17 % 5 ?>       <!-- 2 -->
<? echo -$n ?>          <!-- unary minus -->
```

#### String Concatenation

Use `.` (dot operator) to concatenate strings:

```
<? echo "Hello, " . $name . "!" ?>
<? $fullName = $first . " " . $last ?>
```

#### Comparison

```
<? echo $a == $b ?>     <!-- true / false -->
<? echo $a != $b ?>
<? echo $a < $b ?>
<? echo $a > $b ?>
<? echo $a <= $b ?>
<? echo $a >= $b ?>
```

#### Logical

```
<? echo $x && $y ?>
<? echo $x || $y ?>
<? echo !$flag ?>
```

#### Ternary

```
<? echo $score >= 60 ? "pass" : "fail" ?>
<? $label = $active ? "enabled" : "disabled" ?>
<? echo isset($name) ? $name : "Guest" ?>
```

#### Null Coalescing

Returns the left side if it is non-nil, otherwise the right side:

```
<? echo $name ?? "Anonymous" ?>
<? $display = $nickname ?? $username ?? "User" ?>
```

#### `isset` — Variable Check

```
<? if(isset($user)){ ?>
    Hello, <? echo $user ?>!
<? } ?>
```

#### Variable Access

```
<? echo $name ?>
<? echo $user.Name ?>          <!-- struct field or map key -->
<? echo $arr[0] ?>             <!-- integer index -->
<? echo $map["key"] ?>         <!-- string map key -->
<? echo $arr[$i] ?>            <!-- dynamic index from variable -->
<? echo $users[0]["name"] ?>   <!-- chained: slice then map key -->
<? echo $data.Items[2].Title ?><!-- chained dot and bracket access -->
```

### Array and Map Literals

Construct arrays and maps inline inside code blocks.

#### Array Literals

```
<? $nums = [1, 2, 3, 4, 5] ?>
<? echo $nums[2] ?>                     <!-- 3 -->

<? $mixed = ["hello", 42, true, null] ?>
<? echo len($mixed) ?>                  <!-- 4 -->

<? $computed = [$a + 1, $b * 2, $c] ?>

<!-- Inline: no variable needed -->
<? echo json([10, 20, 30]) ?>           <!-- [10,20,30] -->
<? echo len([1, 2, 3]) ?>              <!-- 3 -->

<!-- Range over inline array -->
<? for($v := range ["a", "b", "c"]){ ?>$v<? } ?>
<!-- → "abc" -->

<!-- Nested arrays -->
<? $matrix = [[1, 2], [3, 4]] ?>
<? echo $matrix[1][0] ?>               <!-- 3 -->

<!-- Array in ternary -->
<? $list = $active ? [1, 2, 3] : [] ?>
```

#### Map Literals

Keys can be quoted strings or bare identifiers:

```
<? $user = {"name": "Alice", "age": 30} ?>
<? echo $user["name"] ?>                <!-- Alice -->
<? echo $user["age"] ?>                 <!-- 30 -->

<!-- Bare identifier keys (equivalent to string keys) -->
<? $cfg = {host: "localhost", port: 3306} ?>
<? echo $cfg["host"] ?>                 <!-- localhost -->

<!-- Expression values -->
<? $m = {"double": $x * 2, "label": "val=" . $x} ?>

<!-- Inline in function call -->
<? echo json({"status": "ok", "code": 200}) ?>
<!-- → {"code":200,"status":"ok"} -->

<!-- Range over map literal -->
<? for($k, $v := range {"a": 1, "b": 2}){ ?>$k=$v <? } ?>

<!-- Nested maps -->
<? $config = {"db": {"host": "localhost", "port": 5432}} ?>
<? echo $config["db"]["host"] ?>       <!-- localhost -->

<!-- Array of maps -->
<? $users = [{"name": "Alice"}, {"name": "Bob"}] ?>
<? echo $users[0]["name"] ?>           <!-- Alice -->
<? echo $users[1]["name"] ?>           <!-- Bob -->
```

### Control Flow

#### If / Else

```
<? if($score >= 90){ ?>
    Grade: A
<? }else if($score >= 80){ ?>
    Grade: B
<? }else if($score >= 70){ ?>
    Grade: C
<? }else{ ?>
    Grade: F
<? } ?>
```

```
<? if($user.IsAdmin && $user.Active){ ?>
    <span>Admin Panel</span>
<? } ?>
```

```
<? if(!isset($token) || $token == ""){ ?>
    Please log in.
<? } ?>
```

#### Switch

`break` exits the switch. Multiple values per case are comma-separated.

```
<? switch($status){ ?>
<? case "active": ?>
    User is active.
<? case "pending", "invited": ?>
    Awaiting confirmation.
<? case "banned": ?>
    Access denied.
<? default: ?>
    Unknown status.
<? } ?>
```

```
<? switch($role){ ?>
<? case "admin": ?>admin<? break ?>
<? case "editor": ?>editor<? break ?>
<? default: ?>viewer
<? } ?>
```

### Loops

#### For-Range (iterate over collection)

Iterates over slices, arrays, maps, strings (rune by rune), or integers (0 to n-1).

**Value only:**
```
<? for($v := range $items){ ?>
    <li><? echo $v ?></li>
<? } ?>
```

**Key and value:**
```
<? for($k, $v := range $items){ ?>
    <? echo $k ?>: <? echo $v ?>
<? } ?>
```

**Over a map:**
```
<? for($key, $val := range $config){ ?>
    $key = $val
<? } ?>
```

**Over a string (rune by rune):**
```
<? for($ch := range "Hello"){ ?>$ch<? } ?>
<!-- → "Hello" -->
```

**Over an integer (0 to n-1):**
```
<? for($i := range 5){ ?>
    <? echo $i ?>
<? } ?>
<!-- → "01234" -->
```

**Over an inline array literal:**
```
<? for($v := range ["Mon", "Tue", "Wed", "Thu", "Fri"]){ ?>
    $v
<? } ?>
```

#### C-Style For Loop

```
<? for($i = 0; $i < 10; $i++){ ?>
    <? echo $i ?>
<? } ?>
```

```
<? for($i = $start; $i <= $end; $i += $step){ ?>
    <? echo $i ?>
<? } ?>
```

Compound post-step:
```
<? for($i = 100; $i > 0; $i -= 10){ ?>
    <? echo $i ?>
<? } ?>
```

#### While-Style Loop

A `for` with only a condition (no semicolons):

```
<? $n = 1 ?>
<? for($n <= 1000){ ?>
    <? $n *= 2 ?>
<? } ?>
<? echo $n ?>   <!-- 1024 -->
```

#### `break` and `continue`

```
<? for($v := range $items){ ?>
    <? if($v == "stop"){ ?>
        <? break ?>
    <? } ?>
    <? if($v == "skip"){ ?>
        <? continue ?>
    <? } ?>
    $v
<? } ?>
```

`break` inside a `switch` exits the switch only (not any enclosing loop).

### Loop Metadata

Every `for-range` loop automatically sets `$loop` with metadata for the current iteration. Nested loops each get their own `$loop` (inner loop's `$loop` is active inside the inner body).

| Field         | Type    | Description                              |
|---------------|---------|------------------------------------------|
| `$loop.index` | `int64` | Zero-based iteration index               |
| `$loop.first` | `bool`  | `true` on the first iteration            |
| `$loop.last`  | `bool`  | `true` on the last iteration             |
| `$loop.count` | `int64` | Total number of elements in the iterable |

```go
// Example: comma-separated list with no trailing comma
```
```
<? for($v := range $items){ ?>
    <? echo $v ?><? if(!$loop.last){ ?>,<? } ?>
<? } ?>
<!-- items = ["apple", "banana", "cherry"] → "apple,banana,cherry" -->
```

```
<!-- First/last highlighting -->
<? for($v := range $items){ ?>
    <? if($loop.first){ ?>[<? } ?>
    $v
    <? if($loop.last){ ?>]<? } ?>
<? } ?>
```

```
<!-- Index-based output -->
<? for($item := range $products){ ?>
    <? echo $loop.index + 1 ?>. $item.Name
<? } ?>
```

```
<!-- Count in header -->
<? for($u := range $users){ ?>
    <? if($loop.first){ ?>
        Showing <? echo $loop.count ?> users:
    <? } ?>
    - $u.Name
<? } ?>
```

```
<!-- Pagination separator pattern -->
<? for($page := range $pages){ ?>
    <? if(!$loop.first){ ?> | <? } ?>
    $page
<? } ?>
```

```
<!-- Nested loops: each has its own $loop -->
<? for($row := range $rows){ ?>
    <? for($cell := range $row){ ?>
        [<? echo $loop.index ?>]$cell
    <? } ?>
    (row <? echo $loop.index ?>)
<? } ?>
```

### Include / Require

Load and execute another template file. The child template shares the current context. Both keywords behave identically — missing files are skipped silently.

Maximum nesting depth: 32 levels (prevents infinite cycles).

```
<? include("partials/header.html") ?>
<? require("partials/footer.html") ?>

<!-- Dynamic path from a variable -->
<? include($templatePath) ?>

<!-- Path from expression -->
<? include("layouts/" . $layout . ".html") ?>
```

The included file is compiled and executed with a child context. Variables set before the include are visible inside the included file.

```
<!-- main.html -->
<? $title = "My Page" ?>
<? $user = {name: "Alice"} ?>
<? include("partials/nav.html") ?>

<!-- partials/nav.html -->
<nav>$title — Hello $user["name"]</nav>
```

---

## Built-in Functions

All built-in functions are available in both the simple engine (as `$fn(...)` calls) and the full engine (as `fn(...)` expressions and `$var|fn` modifiers).

### String Functions

```
upper(s)                 → UPPERCASE
lower(s)                 → lowercase
title(s)                 → Title Case
trim(s)                  → strip surrounding whitespace
trimLeft(s, cutset)      → strip leading chars in cutset
trimRight(s, cutset)     → strip trailing chars in cutset
trimPrefix(s, prefix)    → remove prefix if present
trimSuffix(s, suffix)    → remove suffix if present
replace(s, old, new)     → replace all occurrences of old with new
contains(s, sub)         → true if s contains sub
hasPrefix(s, prefix)     → true if s starts with prefix
hasSuffix(s, suffix)     → true if s ends with suffix
split(s, sep)            → []string
repeat(s, n)             → s repeated n times
sprintf(format, args...) → formatted string (Go fmt.Sprintf)
join(elems, sep)         → join []string with separator
joinAny(v, sep)          → join any slice (uses stringify per element)
str(v)                   → convert any value to string
```

```
<? echo upper("hello") ?>                         <!-- HELLO -->
<? echo replace("foo bar foo", "foo", "baz") ?>   <!-- baz bar baz -->
<? echo trim("  spaced  ") ?>                     <!-- spaced -->
<? echo sprintf("%.2f", $price) ?>                <!-- 9.99 -->
<? echo join(["a", "b", "c"], ", ") ?>            <!-- a, b, c -->
<? echo joinAny($items, " | ") ?>                 <!-- works on []any -->
<? echo contains($email, "@") ?>                  <!-- true -->
<? echo repeat("*", 5) ?>                         <!-- ***** -->
```

### HTML / URL / JSON

```
html(s)    → HTML-escape: & " ' < >
url(s)     → URL query-encode
json(v)    → JSON-encode any value
```

```
<? echo html("<b>Bold</b>") ?>
<!-- &lt;b&gt;Bold&lt;/b&gt; -->

<? echo url("hello world!") ?>
<!-- hello+world%21 -->

<? echo json({name: "Alice", scores: [95, 87, 91]}) ?>
<!-- {"name":"Alice","scores":[95,87,91]} -->

<? $data = {tags: $tagList, count: len($tagList)} ?>
<? echo json($data) ?>
```

### Type Conversion

```
int(v)     → int64
float(v)   → float64
str(v)     → string
bool(v)    → bool
```

```
<? $n = int("42") ?>
<? echo $n + 8 ?>              <!-- 50 -->

<? $f = float("3.14") ?>
<? echo $f * 2 ?>              <!-- 6.28 -->

<? echo bool(0) ?>             <!-- false -->
<? echo bool("yes") ?>         <!-- true -->
<? echo str(3.14) ?>           <!-- 3.14 -->
```

### Math

```
abs(v)           → absolute value (preserves int64 or float64 type)
floor(v)         → round down → int64
ceil(v)          → round up   → int64
round(v)         → round to nearest → int64
sqrt(v)          → square root → float64
pow(base, exp)   → base^exp → float64
min(a, b)        → smaller of a, b
max(a, b)        → larger of a, b
```

```
<? echo abs(-7) ?>            <!-- 7 -->
<? echo abs(-3.14) ?>         <!-- 3.14 -->
<? echo floor(4.9) ?>         <!-- 4 -->
<? echo ceil(4.1) ?>          <!-- 5 -->
<? echo round(4.5) ?>         <!-- 5 -->
<? echo sqrt(144) ?>          <!-- 12 -->
<? echo pow(2, 10) ?>         <!-- 1024 -->
<? echo min(3, 7) ?>          <!-- 3 -->
<? echo max(3, 7) ?>          <!-- 7 -->
```

### Collections

```
len(v)              → number of elements (slice, array, map, string, chan)
count(v)            → alias for len
keys(map)           → []string of sorted map keys
values(map)         → []any of map values
first(v)            → first element of slice/array, or first rune of string
last(v)             → last element of slice/array, or last rune of string
slice(v, start, end)→ sub-slice or sub-string [start:end]
```

```
<? $tags = ["go", "tpl", "fast"] ?>
<? echo len($tags) ?>              <!-- 3 -->
<? echo count($tags) ?>            <!-- 3 (alias) -->
<? echo first($tags) ?>            <!-- go -->
<? echo last($tags) ?>             <!-- fast -->
<? echo json(slice($tags, 0, 2)) ?> <!-- ["go","tpl"] -->

<? $m = {"b": 2, "a": 1, "c": 3} ?>
<? echo join(keys($m), ",") ?>     <!-- a,b,c (sorted) -->

<? echo len("Hello") ?>            <!-- 5 -->
<? echo first("Hello") ?>          <!-- H -->
<? echo last("Hello") ?>           <!-- o -->
```

### Logical / Conditional

```
default(val, fallback)      → val if truthy, else fallback
not(v)                      → boolean negation
coalesce(a, b, c, ...)      → first non-nil argument
defined(v)                  → true if v is non-nil
ternary(cond, then, else)   → functional ternary
```

```
<? echo default($name, "Anonymous") ?>        <!-- fallback if empty/nil -->
<? echo not(true) ?>                          <!-- false -->
<? echo coalesce(null, null, "found") ?>      <!-- found -->
<? echo defined($user) ?>                     <!-- true if $user is set -->
<? echo ternary($active, "on", "off") ?>      <!-- on or off -->
```

### Date / Time

```
date(format, value)    → formatted string using Go time layout
now()                  → current time as time.Time
```

`date()` accepts: `time.Time`, `*time.Time`, Unix timestamp (`int` or `float64`), or a string in RFC3339, `"2006-01-02 15:04:05"`, or `"2006-01-02"` format.

Go time layout reference: `2006-01-02 15:04:05` (year=2006, month=01, day=02, hour=15, min=04, sec=05).

```
<? echo date("2006-01-02", now()) ?>
<!-- e.g., "2026-02-21" -->

<? echo date("January 2, 2006", now()) ?>
<!-- e.g., "February 21, 2026" -->

<? echo date("2006-01-02 15:04", now()) ?>
<!-- e.g., "2026-02-21 14:30" -->

<!-- From a time.Time variable -->
<? echo date("2006-01-02", $createdAt) ?>

<!-- From a Unix timestamp -->
<? echo date("2006-01-02", 1700000000) ?>
<!-- "2023-11-14" -->

<!-- From a date string -->
<? echo date("Jan 2006", "2024-06-15") ?>
<!-- "Jun 2024" -->

<!-- Store now() and reuse -->
<? $now = now() ?>
<? echo date("Monday", $now) ?>    <!-- day of week -->
<? echo date("15:04", $now) ?>     <!-- time -->
```

### Debugging

```
dump(v)    → pretty-printed Go representation (via kr/pretty)
```

```
<? echo dump($user) ?>
<!-- {Name:"Alice" Age:30 Tags:["go" "tpl"]} -->

<? echo dump([1, 2, 3]) ?>
<!-- [1, 2, 3] -->
```

---

## Custom Functions

Register any Go function for use in templates. Arguments are automatically coerced to the declared parameter types.

```go
// Register once (e.g., at startup)
tpl.RegisterFunc("exclaim", func(s string) string {
    return s + "!"
})

tpl.RegisterFunc("add", func(a, b int) int {
    return a + b
})

tpl.RegisterFunc("formatPrice", func(cents int64, currency string) string {
    return fmt.Sprintf("%s %.2f", currency, float64(cents)/100)
})

tpl.RegisterFunc("truncate", func(s string, n int) string {
    if len(s) <= n {
        return s
    }
    return s[:n] + "…"
})
```

**Use in full engine:**
```
<? echo exclaim("Hello") ?>              <!-- Hello! -->
<? echo add(3, 4) ?>                     <!-- 7 -->
<? echo add($a, $b) ?>
<? echo formatPrice($price, "USD") ?>    <!-- USD 12.99 -->
<? echo truncate($bio, 100) ?>
```

**Use in simple engine (as modifier):**
```go
tpl.Render("$name|exclaim", tpl.Pairs("name", "Hello"))
// → "Hello!"

tpl.Render("$bio|truncate", ...)  // called with (val, nil extra args)
```

**Functions returning error:**
```go
// A non-nil error suppresses output silently
tpl.RegisterFunc("divide", func(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
})
```

**Variadic functions:**
```go
tpl.RegisterFunc("sum", func(nums ...int) int {
    total := 0
    for _, n := range nums {
        return total + n
    }
    return total
})
```
```
<? echo sum(1, 2, 3, 4, 5) ?>    <!-- 15 -->
```

**Standalone (zero-arg) functions:**

When a name in the simple engine isn't found as a variable, it's tried as a no-arg function call:

```go
tpl.RegisterFunc("now", func() time.Time { return time.Now() })
```
```go
tpl.Render("Current year: $date('2006', now())")
// or in text
tpl.Render("Built at: $buildTime")  // calls buildTime() if registered
```

---

## API Reference

### Simple Engine

```go
// Compile and execute in one call
result := tpl.Render(src string, params ...any) string

// Compile and write directly to an io.Writer
tpl.RenderWriter(w io.Writer, src string, params ...any)

// Pre-compile for repeated use
t := tpl.Parse(src string) *Template
result := t.Execute(params ...any) string
t.WriteTo(w io.Writer, params ...any)

// Cache control (default: 1000)
tpl.SetCacheSize(n int)

// Helper: flat kv list → map[string]any
data := tpl.Pairs(args ...any) map[string]any
```

### Full Engine

```go
// Fluent builder API (recommended)
tpl.Text(src string) *Builder          // from inline string (cached)
tpl.File(path string) *Builder         // from file (mtime-invalidated cache)

b.Set(ctx any) *Builder                // add data source (chainable)
b.Render() string                      // execute → string
b.RenderWriter(w io.Writer)            // execute → writer

// Convenience functions
tpl.RenderText(src string, params ...any) string
tpl.RenderFile(path string, params ...any) string

// Clear both simple and full engine caches
tpl.ClearCache()

// Low-level
engine := tpl.CompileEngine(src string) *Engine
```

### Function Registry

```go
// Register a function for use in all templates
tpl.RegisterFunc(name string, fn any)
```

### Cache Behaviour

| API                | Cache type           | Invalidation         |
|--------------------|----------------------|----------------------|
| `tpl.Parse`        | LRU (configurable)   | `tpl.SetCacheSize(0)` |
| `tpl.Text`         | In-memory map (1000) | `tpl.ClearCache()`   |
| `tpl.File`         | mtime-based          | Automatic on change  |

---

## Complete Examples

### Email Notification Template

```go
const emailTpl = `
<? $greeting = $user.FirstName != "" ? "Hi " . $user.FirstName : "Hello" ?>
<? $greeting ?>,

Your order #<? echo $order.ID ?> has been <? echo $order.Status ?>.

Items ordered:
<? for($item := range $order.Items){ ?>
  <? echo $loop.index + 1 ?>. <? echo $item.Name ?> × <? echo $item.Qty ?> — <? echo sprintf("$%.2f", $item.Price) ?>
<? } ?>

Total: <? echo sprintf("$%.2f", $order.Total) ?>

<? if($order.Status == "shipped"){ ?>
Tracking number: <? echo $order.Tracking ?>
<? } ?>

Thanks,
The Team
`

result := tpl.RenderText(emailTpl, map[string]any{
    "user":  user,
    "order": order,
})
```

### Config Report

```go
const reportTpl = `
System Report — <? echo date("2006-01-02 15:04", now()) ?>

<? for($k, $v := range $config){ ?>
  <? echo $k ?>: <? echo $v ?>
<? } ?>

Services (<? echo count($services) ?> total):
<? for($svc := range $services){ ?>
  [<? echo $svc.Status == "up" ? "✓" : "✗" ?>] <? echo $svc.Name ?><? if($loop.last){ ?> (last)<? } ?>
<? } ?>
`
tpl.RenderText(reportTpl, tpl.Pairs("config", cfg, "services", svcs))
```

### HTML Table with Filter Chaining (Simple Engine)

```go
const rowTpl = `<tr><td>$name|trim|title</td><td>$email|lower|html</td><td>$role|upper</td></tr>`
for _, u := range users {
    fmt.Fprintln(w, tpl.Render(rowTpl, u))
}
```

---
#### [< Table of Contents](https://github.com/getevo/evo#table-of-contents)
