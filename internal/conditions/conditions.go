package conditions

import (
	"fmt"
	"strings"
	"unicode"
)

// tokenType enumerates the token kinds produced by the lexer.
type tokenType int

const (
	tokenIdent  tokenType = iota // dotted identifier like os.macos or file.exists
	tokenAnd                     // "and"
	tokenOr                      // "or"
	tokenNot                     // "not"
	tokenLParen                  // "("
	tokenRParen                  // ")"
	tokenString                  // quoted string literal, e.g. ".env"
	tokenEOF
)

type token struct {
	typ tokenType
	val string
	pos int // byte offset in the source string
}

// --- Lexer ---

type lexer struct {
	src    string
	pos    int
	tokens []token
}

func lex(src string) ([]token, error) {
	l := &lexer{src: src}
	if err := l.run(); err != nil {
		return nil, err
	}
	return l.tokens, nil
}

func (l *lexer) run() error {
	for l.pos < len(l.src) {
		ch := l.src[l.pos]

		if unicode.IsSpace(rune(ch)) {
			l.pos++
			continue
		}

		switch ch {
		case '(':
			l.tokens = append(l.tokens, token{typ: tokenLParen, val: "(", pos: l.pos})
			l.pos++
		case ')':
			l.tokens = append(l.tokens, token{typ: tokenRParen, val: ")", pos: l.pos})
			l.pos++
		case '"':
			tok, err := l.readString()
			if err != nil {
				return err
			}
			l.tokens = append(l.tokens, tok)
		default:
			if isIdentStart(ch) {
				tok := l.readIdent()
				l.tokens = append(l.tokens, tok)
			} else {
				return fmt.Errorf("unexpected character %q at position %d", string(ch), l.pos)
			}
		}
	}
	l.tokens = append(l.tokens, token{typ: tokenEOF, pos: l.pos})
	return nil
}

func (l *lexer) readString() (token, error) {
	start := l.pos
	l.pos++ // skip opening quote
	for l.pos < len(l.src) {
		if l.src[l.pos] == '"' {
			val := l.src[start+1 : l.pos]
			l.pos++ // skip closing quote
			return token{typ: tokenString, val: val, pos: start}, nil
		}
		l.pos++
	}
	return token{}, fmt.Errorf("unterminated string starting at position %d", start)
}

func (l *lexer) readIdent() token {
	start := l.pos
	for l.pos < len(l.src) && isIdentChar(l.src[l.pos]) {
		l.pos++
	}
	val := l.src[start:l.pos]

	var typ tokenType
	switch val {
	case "and":
		typ = tokenAnd
	case "or":
		typ = tokenOr
	case "not":
		typ = tokenNot
	default:
		typ = tokenIdent
	}
	return token{typ: typ, val: val, pos: start}
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9') || ch == '.'
}

// --- AST ---

type node interface {
	nodeType() string
}

type identNode struct {
	name string
}

func (n *identNode) nodeType() string { return "ident" }

type funcCallNode struct {
	name string
	arg  string
}

func (n *funcCallNode) nodeType() string { return "funcCall" }

type notNode struct {
	operand node
}

func (n *notNode) nodeType() string { return "not" }

type binaryNode struct {
	op    tokenType // tokenAnd or tokenOr
	left  node
	right node
}

func (n *binaryNode) nodeType() string { return "binary" }

// --- Parser ---

type parser struct {
	tokens []token
	pos    int
	src    string
}

func parse(src string) (node, error) {
	tokens, err := lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens, src: src}
	n, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.peek().typ != tokenEOF {
		t := p.peek()
		return nil, fmt.Errorf("unexpected token %q at position %d", t.val, t.pos)
	}
	return n, nil
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokenEOF, pos: len(p.src)}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() token {
	t := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return t
}

// parseExpr → orExpr
func (p *parser) parseExpr() (node, error) {
	return p.parseOr()
}

// parseOr → andExpr ("or" andExpr)*
func (p *parser) parseOr() (node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: tokenOr, left: left, right: right}
	}
	return left, nil
}

// parseAnd → notExpr ("and" notExpr)*
func (p *parser) parseAnd() (node, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokenAnd {
		p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: tokenAnd, left: left, right: right}
	}
	return left, nil
}

// parseNot → "not" notExpr | primary
func (p *parser) parseNot() (node, error) {
	if p.peek().typ == tokenNot {
		p.advance()
		operand, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &notNode{operand: operand}, nil
	}
	return p.parsePrimary()
}

// parsePrimary → IDENT "(" STRING ")" | IDENT | "(" expr ")"
func (p *parser) parsePrimary() (node, error) {
	t := p.peek()

	switch t.typ {
	case tokenLParen:
		p.advance()
		n, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().typ != tokenRParen {
			return nil, fmt.Errorf("expected ')' at position %d, got %q", p.peek().pos, p.peek().val)
		}
		p.advance()
		return n, nil

	case tokenIdent:
		p.advance()
		// Check for function-call syntax: ident "(" string ")"
		if p.peek().typ == tokenLParen {
			p.advance()
			if p.peek().typ != tokenString {
				return nil, fmt.Errorf("expected string argument at position %d, got %q", p.peek().pos, p.peek().val)
			}
			arg := p.advance()
			if p.peek().typ != tokenRParen {
				return nil, fmt.Errorf("expected ')' at position %d, got %q", p.peek().pos, p.peek().val)
			}
			p.advance()
			return &funcCallNode{name: t.val, arg: arg.val}, nil
		}
		return &identNode{name: t.val}, nil

	case tokenEOF:
		return nil, fmt.Errorf("unexpected end of expression at position %d", t.pos)

	default:
		return nil, fmt.Errorf("unexpected token %q at position %d", t.val, t.pos)
	}
}

// --- Evaluator ---

// Eval evaluates a boolean condition expression against a set of resolved predicates.
// An empty expression evaluates to true (no condition = always run).
func Eval(expr string, predicates map[string]bool) (bool, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return true, nil
	}

	ast, err := parse(trimmed)
	if err != nil {
		return false, err
	}

	return eval(ast, predicates)
}

func eval(n node, predicates map[string]bool) (bool, error) {
	switch v := n.(type) {
	case *identNode:
		val, ok := predicates[v.name]
		if !ok {
			return false, fmt.Errorf("unknown predicate %q", v.name)
		}
		return val, nil

	case *funcCallNode:
		key := v.name + "(\"" + v.arg + "\")"
		val, ok := predicates[key]
		if !ok {
			return false, fmt.Errorf("unknown predicate %q", key)
		}
		return val, nil

	case *notNode:
		val, err := eval(v.operand, predicates)
		if err != nil {
			return false, err
		}
		return !val, nil

	case *binaryNode:
		left, err := eval(v.left, predicates)
		if err != nil {
			return false, err
		}
		right, err := eval(v.right, predicates)
		if err != nil {
			return false, err
		}
		if v.op == tokenAnd {
			return left && right, nil
		}
		return left || right, nil

	default:
		return false, fmt.Errorf("unexpected node type %T", n)
	}
}

// --- Predicate Extraction ---

// Predicates extracts all predicate names from an expression string.
// These names can be used to pre-resolve predicates before calling Eval.
// An empty expression returns an empty slice (no predicates needed).
func Predicates(expr string) ([]string, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil, nil
	}

	ast, err := parse(trimmed)
	if err != nil {
		return nil, err
	}

	var preds []string
	seen := map[string]bool{}
	collectPredicates(ast, &preds, seen)
	return preds, nil
}

func collectPredicates(n node, preds *[]string, seen map[string]bool) {
	switch v := n.(type) {
	case *identNode:
		if !seen[v.name] {
			*preds = append(*preds, v.name)
			seen[v.name] = true
		}
	case *funcCallNode:
		key := v.name + "(\"" + v.arg + "\")"
		if !seen[key] {
			*preds = append(*preds, key)
			seen[key] = true
		}
	case *notNode:
		collectPredicates(v.operand, preds, seen)
	case *binaryNode:
		collectPredicates(v.left, preds, seen)
		collectPredicates(v.right, preds, seen)
	}
}
