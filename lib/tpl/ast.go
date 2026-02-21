package tpl

// ── Nodes ─────────────────────────────────────────────────────────────────────
// A Node is a compiled template piece executed in order to produce output.

type Node interface{ isNode() }

// TextNode is literal template text with $var placeholders pre-scanned.
type TextNode struct{ parts []textPart }

// StmtsNode is a list of statements from one <? ... ?> code block.
type StmtsNode struct{ stmts []Stmt }

// ForRangeNode is  <? for($key, $val := range $iter){ ?> body <? } ?>
type ForRangeNode struct {
	Key  string // "" when only one variable
	Val  string
	Iter Expr
	Body []Node
}

// ForCNode is  <? for($i=0; $i<10; $i++){ ?> body <? } ?>
// Also serves as a while loop when Init and Post are nil:  for($cond){
type ForCNode struct {
	Init Stmt // may be nil
	Cond Expr // nil → infinite loop
	Post Stmt // may be nil
	Body []Node
}

// IfNode is  <? if(cond){ ?> then <? }else{ ?> else <? } ?>
type IfNode struct {
	Cond Expr
	Then []Node
	Else []Node // nil if no else
}

// IncludeNode is  <? require("path") ?> or <? include("path") ?>
type IncludeNode struct {
	Path Expr
}

// SwitchNode is  <? switch($val){ ?> cases <? } ?>
type SwitchNode struct {
	Val     Expr
	Cases   []CaseClause
	Default []Node
}

// CaseClause is one arm of a switch statement.
// A single case may match multiple values:  case "a", "b":
type CaseClause struct {
	Vals []Expr
	Body []Node
}

func (TextNode) isNode()    {}
func (StmtsNode) isNode()   {}
func (ForRangeNode) isNode() {}
func (ForCNode) isNode()    {}
func (IfNode) isNode()      {}
func (IncludeNode) isNode() {}
func (SwitchNode) isNode()  {}

// ── Statements ────────────────────────────────────────────────────────────────
// Stmts appear inside <? ?> blocks and do NOT produce output on their own
// (except EchoStmt and ExprStmt).

type Stmt interface{ isStmt() }

// EchoStmt outputs the value of X.  echo $expr  /  print $expr
type EchoStmt struct{ X Expr }

// ExprStmt evaluates X and outputs a non-nil result.
// Used for bare  <? $name ?>  and  <? fn(args) ?>
type ExprStmt struct{ X Expr }

// AssignStmt is  $var = expr
type AssignStmt struct {
	Name string
	X    Expr
}

// CompoundAssignStmt is  $var += expr / $var -= expr / $var *= expr / $var /= expr
type CompoundAssignStmt struct {
	Name string
	Op   string // "+=", "-=", "*=", "/="
	X    Expr
}

// IncDecStmt is  $var++  or  $var--
type IncDecStmt struct {
	Name string
	Op   string // "++" or "--"
}

// BreakStmt exits the nearest enclosing for loop.
type BreakStmt struct{}

// ContinueStmt skips to the next iteration of the nearest enclosing for loop.
type ContinueStmt struct{}

func (EchoStmt) isStmt()           {}
func (ExprStmt) isStmt()           {}
func (AssignStmt) isStmt()         {}
func (CompoundAssignStmt) isStmt() {}
func (IncDecStmt) isStmt()         {}
func (BreakStmt) isStmt()          {}
func (ContinueStmt) isStmt()       {}

// ── Expressions ───────────────────────────────────────────────────────────────

type Expr interface{ isExpr() }

// LitExpr is a literal value: string, int64, float64, bool, or nil.
type LitExpr struct{ Val any }

// VarExpr is a variable path: plain name, dotted, or indexed.
// Examples: "name", "user.Name", "arr[0]", "m[key]"
type VarExpr struct{ Path string }

// CallExpr is  fn(arg1, arg2, ...)
type CallExpr struct {
	Name string
	Args []Expr
}

// BinExpr is  L op R
type BinExpr struct {
	Op   string
	L, R Expr
}

// UnExpr is  op X   (e.g. !$flag, -$n)
type UnExpr struct {
	Op string
	X  Expr
}

// IncDecExpr is  $var++  or  $var--  used inside a larger expression.
type IncDecExpr struct {
	Name string
	Op   string // "++" or "--"
}

// TernaryExpr is  cond ? then : else
type TernaryExpr struct {
	Cond Expr
	Then Expr
	Else Expr
}

// IsSetExpr reports whether a variable path is defined in the context.
// Produced by  isset($path)  in code blocks.
type IsSetExpr struct {
	Path string
}

// ArrayLitExpr is a literal array constructed in a code block: [expr, expr, ...]
type ArrayLitExpr struct{ Elems []Expr }

// MapLitExpr is a literal map constructed in a code block: {"key": val, "key2": val2}
type MapLitExpr struct {
	Keys []string
	Vals []Expr
}

func (LitExpr) isExpr()      {}
func (VarExpr) isExpr()      {}
func (CallExpr) isExpr()     {}
func (BinExpr) isExpr()      {}
func (UnExpr) isExpr()       {}
func (IncDecExpr) isExpr()   {}
func (TernaryExpr) isExpr()  {}
func (IsSetExpr) isExpr()    {}
func (ArrayLitExpr) isExpr() {}
func (MapLitExpr) isExpr()   {}

// ── Text segment ──────────────────────────────────────────────────────────────

// textPart is one piece of a pre-scanned TextNode.
type textPart struct {
	literal string
	path    string // variable path when isVar==true
	isVar   bool
}
