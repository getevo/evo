package tpl

import (
	"strconv"
	"strings"
)

// TKind is a code-block token kind.
type TKind uint8

const (
	TkEOF       TKind = iota
	TkIdent           // bare identifier: echo, for, range, translate, …
	TkVar             // $name  (val holds the name without '$')
	TkString          // "…" or '…'  (lit holds the parsed string value)
	TkInt             // integer literal (lit holds int64)
	TkFloat           // float literal   (lit holds float64)
	TkOp              // any operator or punctuation stored as a string
	TkLParen          // (
	TkRParen          // )
	TkLBrace          // {
	TkRBrace          // }
	TkLBracket        // [
	TkRBracket        // ]
	TkComma           // ,
	TkSemicolon       // ;  (or newline acting as statement separator)
)

// Token is one lexed unit from a code block.
type Token struct {
	Kind TKind
	Val  string // raw text
	Lit  any    // parsed value for TkString, TkInt, TkFloat
}

// tokenStream wraps a token slice with a cursor for parser use.
type tokenStream struct {
	toks []Token
	pos  int
}

func (ts *tokenStream) peek() Token {
	if ts.pos < len(ts.toks) {
		return ts.toks[ts.pos]
	}
	return Token{Kind: TkEOF}
}

func (ts *tokenStream) next() Token {
	t := ts.peek()
	if ts.pos < len(ts.toks) {
		ts.pos++
	}
	return t
}

func (ts *tokenStream) consume(kind TKind, val string) bool {
	t := ts.peek()
	if t.Kind == kind && (val == "" || t.Val == val) {
		ts.pos++
		return true
	}
	return false
}

func (ts *tokenStream) eof() bool {
	return ts.pos >= len(ts.toks) || ts.toks[ts.pos].Kind == TkEOF
}

// tokenize lexes src (the content of a <? ?> block) into tokens.
func tokenize(src string) *tokenStream {
	l := &codeLexer{src: src}
	l.scan()
	return &tokenStream{toks: l.toks}
}

// ── Lexer ─────────────────────────────────────────────────────────────────────

type codeLexer struct {
	src  string
	pos  int
	toks []Token
}

func (l *codeLexer) scan() {
	for {
		l.skipSpaceAndComments()
		if l.pos >= len(l.src) {
			l.toks = append(l.toks, Token{Kind: TkEOF})
			return
		}
		b := l.src[l.pos]
		switch {
		case b == '\n' || b == ';':
			// Newline / semicolon = statement separator
			l.toks = append(l.toks, Token{Kind: TkSemicolon, Val: ";"})
			l.pos++
			// Collapse consecutive separators
			for l.pos < len(l.src) && (l.src[l.pos] == '\n' || l.src[l.pos] == ';' || l.src[l.pos] == ' ' || l.src[l.pos] == '\t' || l.src[l.pos] == '\r') {
				l.pos++
			}
		case b == '$':
			l.scanVar()
		case b == '"' || b == '\'':
			l.scanString(b)
		case b >= '0' && b <= '9':
			l.scanNumber(false)
		case isLexAlpha(b):
			l.scanIdent()
		case b == '(':
			l.emit(TkLParen, "(")
		case b == ')':
			l.emit(TkRParen, ")")
		case b == '{':
			l.emit(TkLBrace, "{")
		case b == '}':
			l.emit(TkRBrace, "}")
		case b == '[':
			l.emit(TkLBracket, "[")
		case b == ']':
			l.emit(TkRBracket, "]")
		case b == ',':
			l.emit(TkComma, ",")
		default:
			l.scanOp()
		}
	}
}

func (l *codeLexer) emit(k TKind, v string) {
	l.toks = append(l.toks, Token{Kind: k, Val: v})
	l.pos++
}

func (l *codeLexer) skipSpaceAndComments() {
	for l.pos < len(l.src) {
		b := l.src[l.pos]
		if b == ' ' || b == '\t' || b == '\r' {
			l.pos++
			continue
		}
		// // line comment
		if b == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.pos++
			}
			continue
		}
		break
	}
}

func (l *codeLexer) scanVar() {
	l.pos++ // skip '$'
	start := l.pos
	for l.pos < len(l.src) && isLexAlphaNum(l.src[l.pos]) {
		l.pos++
	}
	name := l.src[start:l.pos]
	if name == "" {
		l.toks = append(l.toks, Token{Kind: TkOp, Val: "$"})
		return
	}
	l.toks = append(l.toks, Token{Kind: TkVar, Val: name})
}

func (l *codeLexer) scanString(quote byte) {
	l.pos++ // skip opening quote
	var sb strings.Builder
	for l.pos < len(l.src) {
		b := l.src[l.pos]
		if b == quote {
			l.pos++
			break
		}
		if b == '\\' && l.pos+1 < len(l.src) {
			l.pos++
			switch l.src[l.pos] {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '"', '\'':
				sb.WriteByte(l.src[l.pos])
			default:
				// Unknown escape — preserve both the backslash and the character
				// so Windows paths like C:\Users\... survive intact.
				sb.WriteByte('\\')
				sb.WriteByte(l.src[l.pos])
			}
			l.pos++
			continue
		}
		sb.WriteByte(b)
		l.pos++
	}
	s := sb.String()
	l.toks = append(l.toks, Token{Kind: TkString, Val: s, Lit: s})
}

func (l *codeLexer) scanNumber(negative bool) {
	start := l.pos
	hasDot := false
	for l.pos < len(l.src) {
		b := l.src[l.pos]
		if b >= '0' && b <= '9' {
			l.pos++
		} else if b == '.' && !hasDot && l.pos+1 < len(l.src) && l.src[l.pos+1] >= '0' && l.src[l.pos+1] <= '9' {
			hasDot = true
			l.pos++
		} else {
			break
		}
	}
	numStr := l.src[start:l.pos]
	if negative {
		numStr = "-" + numStr
	}
	if hasDot {
		f, _ := strconv.ParseFloat(numStr, 64)
		l.toks = append(l.toks, Token{Kind: TkFloat, Val: numStr, Lit: f})
	} else {
		n, _ := strconv.ParseInt(numStr, 10, 64)
		l.toks = append(l.toks, Token{Kind: TkInt, Val: numStr, Lit: n})
	}
}

func (l *codeLexer) scanIdent() {
	start := l.pos
	for l.pos < len(l.src) && isLexAlphaNum(l.src[l.pos]) {
		l.pos++
	}
	l.toks = append(l.toks, Token{Kind: TkIdent, Val: l.src[start:l.pos]})
}

func (l *codeLexer) scanOp() {
	if l.pos >= len(l.src) {
		return
	}
	// Two-character operators (checked first)
	if l.pos+1 < len(l.src) {
		two := l.src[l.pos : l.pos+2]
		switch two {
		case "==", "!=", "<=", ">=", "&&", "||", "++", "--", ":=",
			"+=", "-=", "*=", "/=", "??":
			l.toks = append(l.toks, Token{Kind: TkOp, Val: two})
			l.pos += 2
			return
		}
	}
	// Single-character operators
	b := l.src[l.pos]
	switch b {
	case '+', '-', '*', '/', '%', '.', '<', '>', '=', '!', '?', ':':
		l.toks = append(l.toks, Token{Kind: TkOp, Val: string(b)})
		l.pos++
	default:
		l.pos++ // skip unknown
	}
}

func isLexAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isLexAlphaNum(b byte) bool {
	return isLexAlpha(b) || (b >= '0' && b <= '9')
}
