package tpl

import "strings"

// parseStmts parses the content of a <? ?> block into a list of Stmts.
// It stops when the token stream is exhausted.
func parseStmts(ts *tokenStream) []Stmt {
	var stmts []Stmt
	for !ts.eof() {
		// Skip bare semicolons / newlines
		for ts.peek().Kind == TkSemicolon {
			ts.next()
		}
		if ts.eof() {
			break
		}
		s := parseOneStmt(ts)
		if s != nil {
			stmts = append(stmts, s)
		}
		// Consume optional trailing semicolon
		for ts.peek().Kind == TkSemicolon {
			ts.next()
		}
	}
	return stmts
}

// parseOneStmt parses a single statement from ts.
func parseOneStmt(ts *tokenStream) Stmt {
	t := ts.peek()

	// echo / print
	if t.Kind == TkIdent && (t.Val == "echo" || t.Val == "print") {
		ts.next()
		return EchoStmt{X: parseExpr(ts)}
	}

	// break
	if t.Kind == TkIdent && t.Val == "break" {
		ts.next()
		return BreakStmt{}
	}

	// continue
	if t.Kind == TkIdent && t.Val == "continue" {
		ts.next()
		return ContinueStmt{}
	}

	// require / include
	if t.Kind == TkIdent && (t.Val == "require" || t.Val == "include") {
		ts.next()
		ts.consume(TkLParen, "")
		_ = parseExpr(ts)
		ts.consume(TkRParen, "")
		return nil // handled as IncludeNode at the engine level; return sentinel
	}

	// $var = expr  or  $var := expr  or  $var++  or  $var--  or  $var += expr  etc.
	if t.Kind == TkVar {
		name := t.Val
		ts.next()

		nxt := ts.peek()

		// $var++  /  $var--
		if nxt.Kind == TkOp && (nxt.Val == "++" || nxt.Val == "--") {
			ts.next()
			return IncDecStmt{Name: name, Op: nxt.Val}
		}

		// $var = expr  or  $var := expr
		if nxt.Kind == TkOp && (nxt.Val == "=" || nxt.Val == ":=") {
			ts.next()
			return AssignStmt{Name: name, X: parseExpr(ts)}
		}

		// Compound assignment: $var += expr  /  $var -= expr  /  $var *= expr  /  $var /= expr
		if nxt.Kind == TkOp && (nxt.Val == "+=" || nxt.Val == "-=" || nxt.Val == "*=" || nxt.Val == "/=") {
			ts.next()
			return CompoundAssignStmt{Name: name, Op: nxt.Val, X: parseExpr(ts)}
		}

		// Bare $var — output its value
		// Re-build the VarExpr including any suffix (.field, [idx])
		path := name + parseSuffix(ts)
		return ExprStmt{X: VarExpr{Path: path}}
	}

	// Bare function call: fn(args)
	if t.Kind == TkIdent {
		ts.next()
		name := t.Val
		if ts.peek().Kind == TkLParen {
			return ExprStmt{X: parseCallTail(ts, name)}
		}
		// Bare ident — unusual; skip
		return nil
	}

	// Skip unrecognised token
	ts.next()
	return nil
}

// parseInclude parses require("path") or include("path") returning an IncludeNode.
func parseInclude(ts *tokenStream) *IncludeNode {
	ts.next() // consume "require"/"include"
	ts.consume(TkLParen, "")
	path := parseExpr(ts)
	ts.consume(TkRParen, "")
	return &IncludeNode{Path: path}
}

// ── Code-block classification ─────────────────────────────────────────────────

type blockKind int

const (
	bkStmts        blockKind = iota // regular statements
	bkForRangeOpen                  // for($k,$v := range $iter){
	bkForCOpen                      // for($i=0; $i<10; $i++){  or while-style for($cond){
	bkIfOpen                        // if(cond){
	bkElseOpen                      // }else{  or  }else if(cond){
	bkClose                         // }
	bkSwitchOpen                    // switch($val){
	bkCaseLabel                     // case X, Y:
	bkDefaultLabel                  // default:
)

// parsedBlock is the result of parsing one <? ?> content string.
type parsedBlock struct {
	kind   blockKind
	stmts  []Stmt // for bkStmts
	elseIf *parsedBlock

	// ForRange fields
	frKey  string
	frVal  string
	frIter Expr

	// ForC fields
	fcInit Stmt
	fcCond Expr
	fcPost Stmt

	// If fields
	ifCond Expr

	// Switch fields
	switchVal Expr

	// Case fields
	caseVals []Expr
}

// classifyAndParse inspects the trimmed code content of a <? ?> block and
// returns a parsedBlock describing what it found.
func classifyAndParse(code string) *parsedBlock {
	code = strings.TrimSpace(code)
	if code == "" {
		return &parsedBlock{kind: bkStmts}
	}

	// Bare  }
	if code == "}" {
		return &parsedBlock{kind: bkClose}
	}

	// }else{ or }else if(cond){
	if strings.HasPrefix(code, "}") {
		rest := strings.TrimSpace(code[1:])
		if strings.HasPrefix(rest, "else") {
			rest2 := strings.TrimSpace(rest[4:])
			if strings.HasPrefix(rest2, "if") {
				// }else if(cond){
				inner := strings.TrimSpace(rest2[2:])
				inner = strings.TrimSuffix(strings.TrimSpace(inner), "{")
				inner = strings.TrimPrefix(strings.TrimSpace(inner), "(")
				inner = strings.TrimSuffix(strings.TrimSpace(inner), ")")
				ts := tokenize(inner)
				cond := parseExpr(ts)
				return &parsedBlock{kind: bkElseOpen, ifCond: cond}
			}
			// }else{
			return &parsedBlock{kind: bkElseOpen}
		}
	}

	ts := tokenize(code)

	// for(...)
	if ts.peek().Kind == TkIdent && ts.peek().Val == "for" {
		ts.next() // consume "for"
		ts.consume(TkLParen, "")
		return parseForHeader(ts)
	}

	// if(cond){
	if ts.peek().Kind == TkIdent && ts.peek().Val == "if" {
		ts.next()
		ts.consume(TkLParen, "")
		cond := parseExpr(ts)
		ts.consume(TkRParen, "")
		ts.consume(TkLBrace, "")
		return &parsedBlock{kind: bkIfOpen, ifCond: cond}
	}

	// switch($val){
	if ts.peek().Kind == TkIdent && ts.peek().Val == "switch" {
		ts.next()
		ts.consume(TkLParen, "")
		val := parseExpr(ts)
		ts.consume(TkRParen, "")
		ts.consume(TkLBrace, "")
		return &parsedBlock{kind: bkSwitchOpen, switchVal: val}
	}

	// case X, Y:
	if ts.peek().Kind == TkIdent && ts.peek().Val == "case" {
		ts.next()
		var vals []Expr
		for !ts.eof() {
			// Stop at the trailing ":"
			if ts.peek().Kind == TkOp && ts.peek().Val == ":" {
				ts.next()
				break
			}
			if ts.peek().Kind == TkComma {
				ts.next()
				continue
			}
			vals = append(vals, parseExpr(ts))
		}
		return &parsedBlock{kind: bkCaseLabel, caseVals: vals}
	}

	// default:
	if ts.peek().Kind == TkIdent && ts.peek().Val == "default" {
		return &parsedBlock{kind: bkDefaultLabel}
	}

	// require / include
	if ts.peek().Kind == TkIdent && (ts.peek().Val == "require" || ts.peek().Val == "include") {
		node := parseInclude(ts)
		return &parsedBlock{kind: bkStmts, stmts: []Stmt{includeAsStmt{node}}}
	}

	// Regular statements
	stmts := parseStmts(ts)
	return &parsedBlock{kind: bkStmts, stmts: stmts}
}

// includeAsStmt wraps IncludeNode so it satisfies the Stmt interface for the
// regular-statement path (include/require can appear mid-block).
type includeAsStmt struct{ node *IncludeNode }

func (includeAsStmt) isStmt() {}

// parseForHeader parses the part after  for(  up to and including the opening {.
func parseForHeader(ts *tokenStream) *parsedBlock {
	// Peek: is the first var followed by , or :=  → range loop
	// Or is it  $i = 0; cond; post  → C-style loop
	// Or is there no semicolon at all → while-style loop

	// Check for range loop: $key, $val := range $iter
	//                    or $val := range $iter
	saved := ts.pos

	if ts.peek().Kind == TkVar {
		firstVar := ts.peek().Val
		ts.next()

		// $val := range $iter){
		if ts.peek().Kind == TkOp && ts.peek().Val == ":=" {
			ts.next()
			if ts.peek().Kind == TkIdent && ts.peek().Val == "range" {
				ts.next()
				iter := parseExpr(ts)
				ts.consume(TkRParen, "")
				ts.consume(TkLBrace, "")
				return &parsedBlock{kind: bkForRangeOpen, frKey: "", frVal: firstVar, frIter: iter}
			}
		}

		// $key, $val := range $iter){
		if ts.peek().Kind == TkComma {
			ts.next()
			if ts.peek().Kind == TkVar {
				secondVar := ts.peek().Val
				ts.next()
				if ts.peek().Kind == TkOp && ts.peek().Val == ":=" {
					ts.next()
					if ts.peek().Kind == TkIdent && ts.peek().Val == "range" {
						ts.next()
						iter := parseExpr(ts)
						ts.consume(TkRParen, "")
						ts.consume(TkLBrace, "")
						return &parsedBlock{kind: bkForRangeOpen, frKey: firstVar, frVal: secondVar, frIter: iter}
					}
				}
			}
		}
	}

	// Rewind and check if this is a while-style loop (no semicolons before the matching ')')
	ts.pos = saved
	hasSemi := false
	for _, tok := range ts.toks[ts.pos:] {
		if tok.Kind == TkSemicolon {
			hasSemi = true
			break
		}
	}
	if !hasSemi {
		// while-style: for($cond){ body }
		var cond Expr
		if ts.peek().Kind != TkRParen {
			cond = parseExpr(ts)
		}
		ts.consume(TkRParen, "")
		ts.consume(TkLBrace, "")
		return &parsedBlock{kind: bkForCOpen, fcInit: nil, fcCond: cond, fcPost: nil}
	}

	// Rewind and try C-style: $i=0; cond; post
	ts.pos = saved

	// Init: may be  $i=0  or empty
	var initStmt Stmt
	if ts.peek().Kind == TkVar {
		varName := ts.peek().Val
		ts.next()
		nxt := ts.peek()
		if nxt.Kind == TkOp && nxt.Val == "=" {
			ts.next()
			initStmt = AssignStmt{Name: varName, X: parseExpr(ts)}
		} else if nxt.Kind == TkOp && (nxt.Val == "+=" || nxt.Val == "-=" || nxt.Val == "*=" || nxt.Val == "/=") {
			ts.next()
			initStmt = CompoundAssignStmt{Name: varName, Op: nxt.Val, X: parseExpr(ts)}
		}
	}
	ts.consume(TkSemicolon, "")

	// Condition
	var cond Expr
	if ts.peek().Kind != TkSemicolon {
		cond = parseExpr(ts)
	}
	ts.consume(TkSemicolon, "")

	// Post: $i++ / $i-- / $i=expr / $i+=expr
	var postStmt Stmt
	if ts.peek().Kind == TkVar {
		varName := ts.peek().Val
		ts.next()
		nxt := ts.peek()
		if nxt.Kind == TkOp && (nxt.Val == "++" || nxt.Val == "--") {
			ts.next()
			postStmt = IncDecStmt{Name: varName, Op: nxt.Val}
		} else if nxt.Kind == TkOp && nxt.Val == "=" {
			ts.next()
			postStmt = AssignStmt{Name: varName, X: parseExpr(ts)}
		} else if nxt.Kind == TkOp && (nxt.Val == "+=" || nxt.Val == "-=" || nxt.Val == "*=" || nxt.Val == "/=") {
			ts.next()
			postStmt = CompoundAssignStmt{Name: varName, Op: nxt.Val, X: parseExpr(ts)}
		}
	}
	ts.consume(TkRParen, "")
	ts.consume(TkLBrace, "")

	return &parsedBlock{kind: bkForCOpen, fcInit: initStmt, fcCond: cond, fcPost: postStmt}
}
