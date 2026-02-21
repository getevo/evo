package tpl

import (
	"strings"
	"unicode/utf8"
)

// Engine is a compiled template ready for repeated execution.
type Engine struct {
	nodes []Node
}

// CompileEngine compiles src into an Engine.
func CompileEngine(src string) *Engine {
	p := &tmplParser{src: src}
	nodes := p.parseNodes(false, false, false)
	return &Engine{nodes: nodes}
}

// ── Template parser ───────────────────────────────────────────────────────────

type tmplParser struct {
	src       string
	pos       int
	lastBlock *parsedBlock // set when parseNodes returns due to a stop condition
}

// parseNodes reads nodes from the current position.
//   - stopOnClose: stop when a bare } is found
//   - stopOnElse:  stop when a }else{ or }else if( is found
//   - stopOnCase:  stop when a case or default label is found (for switch bodies)
//
// In all cases the triggering block is left in p.lastBlock.
func (p *tmplParser) parseNodes(stopOnClose, stopOnElse, stopOnCase bool) []Node {
	var nodes []Node
	p.lastBlock = nil

	for {
		// Find the next <? marker
		rel := strings.Index(p.src[p.pos:], "<?")
		if rel < 0 {
			// Remaining text
			if p.pos < len(p.src) {
				nodes = append(nodes, compileTextNode(p.src[p.pos:]))
				p.pos = len(p.src)
			}
			return nodes
		}
		tagPos := p.pos + rel

		// Text before the tag
		if tagPos > p.pos {
			nodes = append(nodes, compileTextNode(p.src[p.pos:tagPos]))
		}
		p.pos = tagPos + 2 // skip <?

		// Find closing ?>
		rel2 := strings.Index(p.src[p.pos:], "?>")
		var code string
		if rel2 < 0 {
			code = strings.TrimSpace(p.src[p.pos:])
			p.pos = len(p.src)
		} else {
			code = strings.TrimSpace(p.src[p.pos : p.pos+rel2])
			p.pos = p.pos + rel2 + 2
			// Consume a single trailing newline after ?> so that block tags
			// on their own lines don't produce blank lines in output.
			if p.pos < len(p.src) && p.src[p.pos] == '\n' {
				p.pos++
			}
		}

		if code == "" {
			continue
		}

		pb := classifyAndParse(code)

		switch pb.kind {
		case bkClose:
			if stopOnClose {
				p.lastBlock = pb
				return nodes
			}
			// Unmatched close — ignore

		case bkElseOpen:
			if stopOnElse {
				p.lastBlock = pb
				return nodes
			}
			// Unmatched else — ignore

		case bkCaseLabel, bkDefaultLabel:
			if stopOnCase {
				p.lastBlock = pb
				return nodes
			}
			// case/default outside switch — ignore

		case bkForRangeOpen:
			body := p.parseNodes(true, false, false)
			nodes = append(nodes, ForRangeNode{
				Key:  pb.frKey,
				Val:  pb.frVal,
				Iter: pb.frIter,
				Body: body,
			})

		case bkForCOpen:
			body := p.parseNodes(true, false, false)
			nodes = append(nodes, ForCNode{
				Init: pb.fcInit,
				Cond: pb.fcCond,
				Post: pb.fcPost,
				Body: body,
			})

		case bkIfOpen:
			thenBody := p.parseNodes(true, true, false)
			lb := p.lastBlock
			p.lastBlock = nil
			var elseBody []Node
			if lb != nil && lb.kind == bkElseOpen {
				if lb.ifCond != nil {
					// }else if(cond){ — synthesise as nested IfNode
					elseBody = p.buildElseIf(lb)
				} else {
					// plain }else{
					elseBody = p.parseNodes(true, false, false)
				}
			}
			nodes = append(nodes, IfNode{
				Cond: pb.ifCond,
				Then: thenBody,
				Else: elseBody,
			})

		case bkSwitchOpen:
			cases, defaultBody := p.parseSwitchCases()
			nodes = append(nodes, SwitchNode{
				Val:     pb.switchVal,
				Cases:   cases,
				Default: defaultBody,
			})

		case bkStmts:
			if len(pb.stmts) == 0 {
				continue
			}
			var outStmts []Stmt
			for _, s := range pb.stmts {
				if is, ok := s.(includeAsStmt); ok {
					nodes = append(nodes, *is.node)
					continue
				}
				outStmts = append(outStmts, s)
			}
			if len(outStmts) > 0 {
				nodes = append(nodes, StmtsNode{stmts: outStmts})
			}
		}
	}
}

// buildElseIf handles }else if(cond){ by creating a nested IfNode.
func (p *tmplParser) buildElseIf(lb *parsedBlock) []Node {
	thenBody := p.parseNodes(true, true, false)
	next := p.lastBlock
	p.lastBlock = nil
	var elseBody []Node
	if next != nil && next.kind == bkElseOpen {
		if next.ifCond != nil {
			elseBody = p.buildElseIf(next)
		} else {
			elseBody = p.parseNodes(true, false, false)
		}
	}
	return []Node{IfNode{Cond: lb.ifCond, Then: thenBody, Else: elseBody}}
}

// parseSwitchCases parses the body of a switch block, collecting case clauses.
// It consumes all content up to and including the closing }.
func (p *tmplParser) parseSwitchCases() ([]CaseClause, []Node) {
	var cases []CaseClause
	var defaultBody []Node
	var curVals []Expr
	inCase := false
	inDefault := false

	for {
		// Parse nodes until we hit a case label, default label, or close brace.
		body := p.parseNodes(true, false, true)
		lb := p.lastBlock
		p.lastBlock = nil

		// Attach the collected body to the currently open case/default.
		if inDefault {
			defaultBody = body
		} else if inCase {
			cases = append(cases, CaseClause{Vals: curVals, Body: body})
		}
		// Body before the first case label is discarded.

		if lb == nil || lb.kind == bkClose {
			return cases, defaultBody
		}

		if lb.kind == bkCaseLabel {
			curVals = lb.caseVals
			inCase = true
			inDefault = false
		} else if lb.kind == bkDefaultLabel {
			curVals = nil
			inCase = false
			inDefault = true
		}
	}
}

// ── Text-node compiler ────────────────────────────────────────────────────────

// compileTextNode pre-scans text for $var references and returns a TextNode.
func compileTextNode(src string) TextNode {
	var parts []textPart
	rem := src
	for len(rem) > 0 {
		i := strings.IndexByte(rem, '$')
		if i < 0 {
			parts = append(parts, textPart{literal: rem})
			break
		}
		if i > 0 {
			parts = append(parts, textPart{literal: rem[:i]})
		}
		rem = rem[i+1:]

		if len(rem) == 0 {
			parts = append(parts, textPart{literal: "$"})
			break
		}
		if rem[0] == '$' {
			parts = append(parts, textPart{literal: "$"})
			rem = rem[1:]
			continue
		}

		r, _ := utf8.DecodeRuneInString(rem)
		if !isIdentStart(r) {
			parts = append(parts, textPart{literal: "$"})
			continue
		}

		path, consumed := scanTextPath(rem)
		rem = rem[consumed:]
		parts = append(parts, textPart{path: path, isVar: true})
	}
	return TextNode{parts: parts}
}

// scanTextPath reads a variable path from rem right after the '$'.
// Handles  name, name.Field, name[0], name["key"], name['key'].
func scanTextPath(rem string) (string, int) {
	var sb strings.Builder
	pos := 0

	// Root identifier
	for pos < len(rem) {
		r, sz := utf8.DecodeRuneInString(rem[pos:])
		if !isIdentStart(r) && !(r >= '0' && r <= '9') && r != '_' {
			break
		}
		sb.WriteRune(r)
		pos += sz
	}

	for pos < len(rem) {
		// Dotted field
		if rem[pos] == '.' && pos+1 < len(rem) {
			r, _ := utf8.DecodeRuneInString(rem[pos+1:])
			if isIdentStart(r) {
				sb.WriteByte('.')
				pos++
				for pos < len(rem) {
					r2, sz := utf8.DecodeRuneInString(rem[pos:])
					if !isIdentStart(r2) && !(r2 >= '0' && r2 <= '9') && r2 != '_' {
						break
					}
					sb.WriteRune(r2)
					pos += sz
				}
				continue
			}
		}
		// Bracket access
		if rem[pos] == '[' {
			pos++
			if pos >= len(rem) {
				sb.WriteByte('[')
				break
			}
			if rem[pos] == '"' || rem[pos] == '\'' {
				q := rem[pos]
				pos++
				keyStart := pos
				for pos < len(rem) && rem[pos] != q {
					pos++
				}
				key := rem[keyStart:pos]
				if pos < len(rem) {
					pos++ // closing quote
				}
				if pos < len(rem) && rem[pos] == ']' {
					pos++
				}
				sb.WriteByte('[')
				sb.WriteString(key)
				sb.WriteByte(']')
			} else {
				keyStart := pos
				for pos < len(rem) && rem[pos] != ']' {
					pos++
				}
				key := strings.TrimSpace(rem[keyStart:pos])
				if pos < len(rem) {
					pos++
				}
				sb.WriteByte('[')
				sb.WriteString(key)
				sb.WriteByte(']')
			}
			continue
		}
		break
	}
	return sb.String(), pos
}
