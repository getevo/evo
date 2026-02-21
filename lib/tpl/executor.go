package tpl

import (
	"io"
	"os"
	"reflect"
	"strings"
)

// Execute renders the Engine with the given context and returns the output.
func (e *Engine) Execute(ctx *Context) string {
	var b strings.Builder
	executeNodes(e.nodes, ctx, &b)
	return b.String()
}

// ExecuteWriter renders into w.
func (e *Engine) ExecuteWriter(ctx *Context, w io.Writer) {
	executeNodes(e.nodes, ctx, w)
}

// ── Loop control signals ───────────────────────────────────────────────────────

// loopBreak is panicked by BreakStmt and caught by the nearest loop executor.
type loopBreak struct{}

// loopContinue is panicked by ContinueStmt and caught by the nearest loop executor.
type loopContinue struct{}

// runLoopBody executes nodes as a loop iteration body, handling break and continue.
// Returns (shouldBreak): true means the loop should exit, false means continue normally.
func runLoopBody(nodes []Node, ctx *Context, w io.Writer) (shouldBreak bool) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case loopBreak:
				shouldBreak = true
			case loopContinue:
				shouldBreak = false // continue: skip remainder of body, next iteration
			default:
				panic(r) // real panic — re-propagate
			}
		}
	}()
	executeNodes(nodes, ctx, w)
	return false
}

// ── Node execution ────────────────────────────────────────────────────────────

func executeNodes(nodes []Node, ctx *Context, w io.Writer) {
	for _, n := range nodes {
		executeNode(n, ctx, w)
	}
}

func executeNode(n Node, ctx *Context, w io.Writer) {
	switch x := n.(type) {
	case TextNode:
		executeText(x, ctx, w)
	case StmtsNode:
		executeStmts(x.stmts, ctx, w)
	case ForRangeNode:
		executeForRange(x, ctx, w)
	case ForCNode:
		executeForC(x, ctx, w)
	case IfNode:
		executeIf(x, ctx, w)
	case IncludeNode:
		executeInclude(x, ctx, w)
	case SwitchNode:
		executeSwitch(x, ctx, w)
	}
}

// ── Text node ─────────────────────────────────────────────────────────────────

func executeText(n TextNode, ctx *Context, w io.Writer) {
	for _, p := range n.parts {
		if !p.isVar {
			_, _ = io.WriteString(w, p.literal)
			continue
		}
		val := evalVar(p.path, ctx)
		if val != nil {
			_, _ = io.WriteString(w, stringify(val))
		} else {
			// Keep placeholder
			_, _ = io.WriteString(w, "$")
			_, _ = io.WriteString(w, p.path)
		}
	}
}

// ── Statement execution ───────────────────────────────────────────────────────

func executeStmts(stmts []Stmt, ctx *Context, w io.Writer) {
	for _, s := range stmts {
		executeStmt(s, ctx, w)
	}
}

func executeStmt(s Stmt, ctx *Context, w io.Writer) {
	switch x := s.(type) {
	case EchoStmt:
		v := evalExpr(x.X, ctx)
		if v != nil {
			_, _ = io.WriteString(w, stringify(v))
		}
	case ExprStmt:
		v := evalExpr(x.X, ctx)
		if v != nil {
			_, _ = io.WriteString(w, stringify(v))
		}
	case AssignStmt:
		v := evalExpr(x.X, ctx)
		ctx.Set(x.Name, v)
	case CompoundAssignStmt:
		cur, _ := ctx.Get(x.Name)
		if cur == nil {
			cur = int64(0) // treat uninitialised variable as zero
		}
		rhs := evalExpr(x.X, ctx)
		var result any
		switch x.Op {
		case "+=":
			result = numericBin(cur, rhs, func(a, b float64) float64 { return a + b })
		case "-=":
			result = numericBin(cur, rhs, func(a, b float64) float64 { return a - b })
		case "*=":
			result = numericBin(cur, rhs, func(a, b float64) float64 { return a * b })
		case "/=":
			bv, _ := toFloat64(rhs)
			if bv != 0 {
				result = numericBin(cur, rhs, func(a, b float64) float64 { return a / b })
			}
		}
		ctx.Set(x.Name, result)
	case IncDecStmt:
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
	case BreakStmt:
		panic(loopBreak{})
	case ContinueStmt:
		panic(loopContinue{})
	case includeAsStmt:
		// Shouldn't reach here; include is converted to IncludeNode at parse time.
	}
}

// ── For-range loop ────────────────────────────────────────────────────────────

func executeForRange(n ForRangeNode, ctx *Context, w io.Writer) {
	iterVal := evalExpr(n.Iter, ctx)
	if iterVal == nil {
		return
	}
	rv := reflect.ValueOf(iterVal)
	// Loop variables and any assignments inside the body are all written into
	// the current context so they persist after the loop (PHP-like semantics).
	// $loop is set each iteration with index/first/last/count metadata.
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		count := rv.Len()
		for i := 0; i < count; i++ {
			if n.Key != "" {
				ctx.setLocal(n.Key, int64(i))
			}
			ctx.setLocal(n.Val, rv.Index(i).Interface())
			ctx.setLocal("loop", map[string]any{
				"index": int64(i),
				"first": i == 0,
				"last":  i == count-1,
				"count": int64(count),
			})
			if runLoopBody(n.Body, ctx, w) {
				return
			}
		}
	case reflect.Map:
		mapKeys := rv.MapKeys()
		count := len(mapKeys)
		for i, k := range mapKeys {
			if n.Key != "" {
				ctx.setLocal(n.Key, k.Interface())
			}
			ctx.setLocal(n.Val, rv.MapIndex(k).Interface())
			ctx.setLocal("loop", map[string]any{
				"index": int64(i),
				"first": i == 0,
				"last":  i == count-1,
				"count": int64(count),
			})
			if runLoopBody(n.Body, ctx, w) {
				return
			}
		}
	case reflect.String:
		runes := []rune(rv.String())
		count := len(runes)
		for i, r := range runes {
			if n.Key != "" {
				ctx.setLocal(n.Key, int64(i))
			}
			ctx.setLocal(n.Val, string(r))
			ctx.setLocal("loop", map[string]any{
				"index": int64(i),
				"first": i == 0,
				"last":  i == count-1,
				"count": int64(count),
			})
			if runLoopBody(n.Body, ctx, w) {
				return
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Range over integer: for($i := range 10) iterates 0, 1, …, 9
		n2 := rv.Int()
		for i := int64(0); i < n2; i++ {
			if n.Key != "" {
				ctx.setLocal(n.Key, i)
			}
			ctx.setLocal(n.Val, i)
			ctx.setLocal("loop", map[string]any{
				"index": i,
				"first": i == 0,
				"last":  i == n2-1,
				"count": n2,
			})
			if runLoopBody(n.Body, ctx, w) {
				return
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n2 := int64(rv.Uint())
		for i := int64(0); i < n2; i++ {
			if n.Key != "" {
				ctx.setLocal(n.Key, i)
			}
			ctx.setLocal(n.Val, i)
			ctx.setLocal("loop", map[string]any{
				"index": i,
				"first": i == 0,
				"last":  i == n2-1,
				"count": n2,
			})
			if runLoopBody(n.Body, ctx, w) {
				return
			}
		}
	}
}

// ── C-style / while loop ──────────────────────────────────────────────────────

func executeForC(n ForCNode, ctx *Context, w io.Writer) {
	child := ctx.child()

	// Init
	if n.Init != nil {
		executeStmt(n.Init, child, w)
	}

	const maxIter = 1_000_000
	for i := 0; i < maxIter; i++ {
		// Condition (nil = infinite, broken by body)
		if n.Cond != nil {
			if !toBool(evalExpr(n.Cond, child)) {
				break
			}
		}
		if runLoopBody(n.Body, child, w) {
			break
		}

		// Post
		if n.Post != nil {
			executeStmt(n.Post, child, w)
		}
	}
}

// ── If statement ──────────────────────────────────────────────────────────────

func executeIf(n IfNode, ctx *Context, w io.Writer) {
	if toBool(evalExpr(n.Cond, ctx)) {
		executeNodes(n.Then, ctx, w)
	} else if len(n.Else) > 0 {
		executeNodes(n.Else, ctx, w)
	}
}

// ── Switch statement ──────────────────────────────────────────────────────────

func executeSwitch(n SwitchNode, ctx *Context, w io.Writer) {
	val := evalExpr(n.Val, ctx)

	// Wrap execution to catch break inside a case body.
	// break exits the switch; continue propagates to any enclosing loop.
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(loopBreak); ok {
					return // break exits the switch cleanly
				}
				panic(r) // re-panic continue or real panics
			}
		}()

		for _, c := range n.Cases {
			for _, cv := range c.Vals {
				if equal(val, evalExpr(cv, ctx)) {
					executeNodes(c.Body, ctx, w)
					return
				}
			}
		}
		if len(n.Default) > 0 {
			executeNodes(n.Default, ctx, w)
		}
	}()
}

// ── Include / require ─────────────────────────────────────────────────────────

// maxIncludeDepth is the maximum number of nested include/require calls allowed.
// It guards against infinite cycles (a.html → b.html → a.html …) and also caps
// legitimately deep but excessive nesting. Both include and require are subject
// to this limit.
const maxIncludeDepth = 32

func executeInclude(n IncludeNode, ctx *Context, w io.Writer) {
	// Depth is stored on the context so it propagates through every code path
	// (if blocks, for loops, switch statements, etc.) without needing to change
	// any other executor signatures.
	if ctx.includeDepth >= maxIncludeDepth {
		return // cycle detected or nesting too deep — abort silently
	}
	pathVal := evalExpr(n.Path, ctx)
	path := stringify(pathVal)
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return // silently skip missing files (require and include behave the same)
	}
	sub := CompileEngine(string(data))
	// Run the included template in a child context whose includeDepth is one
	// greater than the caller's. child() propagates includeDepth automatically,
	// so any further includes or requires encountered anywhere inside the
	// included file — even inside nested if/for/switch blocks — will also be
	// subject to the depth limit.
	child := ctx.child()
	child.includeDepth = ctx.includeDepth + 1
	executeNodes(sub.nodes, child, w)
}
