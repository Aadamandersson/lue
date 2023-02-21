package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Kind int

const (
	Unknown Kind = iota // An unknown character to the lexer.
	Eof                 // End of file.
	Number              // E.g., `123`
	Plus                // `+`
	Minus               // `-`
	Star                //`*`
	Slash               // `/`
)

var tokens = [...]string{
	Unknown: "unknown",
	Eof:     "eof",
	Number:  "number",
	Plus:    "+",
	Minus:   "-",
	Star:    "*",
	Slash:   "/",
}

func (k Kind) String() string {
	if k < 0 || k >= Kind(len(tokens)) {
		return "TokenKind(" + strconv.FormatInt(int64(k), 10) + ")"
	}
	return tokens[k]
}

type Token struct {
	Kind Kind
	Lit  string // Literal value of token if kind is `Unknown` or `Number`, otherwise empty.
}

func NewToken(kind Kind, lit string) Token {
	return Token{Kind: kind, Lit: lit}
}

// Is returns true if this token kind is equal to k, otherwise false.
func (t Token) Is(k Kind) bool {
	return t.Kind == k
}

// IsOneOf returns true if this token kind is equal to at least one of ks, otherwise false.
func (t Token) IsOneOf(ks ...Kind) bool {
	for _, k := range ks {
		if t.Is(k) {
			return true
		}
	}
	return false
}

func Lex(src []byte) []Token {
	l := lexer{src: src}
	return l.Lex()
}

type lexer struct {
	src []byte
	pos int // Current position in src.
}

func (l *lexer) Lex() []Token {
	var tokens []Token
	for {
		b := l.peek()
		if b == 0 {
			break
		}
		if l.isWhitespace(b) {
			l.eatWhitespace()
			continue
		}

		l.next()
		kind, lit := l.lexToken(b)
		tokens = append(tokens, NewToken(kind, lit))
	}

	tokens = append(tokens, NewToken(Eof, ""))
	return tokens
}

func (l *lexer) lexToken(first byte) (Kind, string) {
	switch first {
	case '+':
		return Plus, ""
	case '-':
		return Minus, ""
	case '*':
		return Star, ""
	case '/':
		return Slash, ""
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return l.lexNumeric(first)
	default:
		return Unknown, string(first)
	}
}

// lexNumeric lexes a number and returns its kind and literal value.
func (l *lexer) lexNumeric(first byte) (Kind, string) {
	s := l.collectString(first, l.isDigit)
	return Number, s
}

// collectString collects bytes into a string while matches returns true and
// the lexer is not at EOF.
func (l *lexer) collectString(first byte, matches func(byte) bool) string {
	var builder strings.Builder
	builder.WriteByte(first)

	for {
		b := l.peek()
		if b == 0 || !matches(b) {
			break
		}
		builder.WriteByte(b)
		l.next()
	}

	return builder.String()
}

// peek returns the next byte without advancing the lexer.
//
// If the lexer is at EOF, 0 is returned.
func (l *lexer) peek() byte {
	if l.pos < len(l.src) {
		return l.src[l.pos]
	}
	return 0
}

// next advances the lexer to the next byte in src.
func (l *lexer) next() {
	if l.pos < len(l.src) {
		l.pos += 1
	}
}

// isDigit returns true if byte b is a digit, otherwise false.
func (l *lexer) isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// isWhitespace returns true if byte b is whitespace, otherwise false.
func (l *lexer) isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// eatWhitespace advances the lexer to the next non-whitespace byte in src.
func (l *lexer) eatWhitespace() {
	for l.isWhitespace(l.peek()) {
		l.next()
	}
}

type Expr interface {
	exprNode()
}

// Expressions
type (
	IntegerLiteral struct {
		V string
	}

	BinaryExpr struct {
		X  Expr
		Op BinOpKind
		Y  Expr
	}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*IntegerLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}

type BinOpKind int

const (
	Add BinOpKind = iota + 1 // `+` (addition)
	Sub                      // `-` (subtraction)
	Mul                      // `*` (multiplication)
	Div                      // `/` (division)
)

// Prec returns the operator precedence for binary operator op.
func (op BinOpKind) Prec() int {
	switch op {
	case Mul, Div:
		return 2
	case Add, Sub:
		return 1
	default:
		return 0
	}
}

// BinOpFromToken returns the binOp for token t and a boolean true, if its a valid binary operator.
// Otherwise, returns the zero value of BinOpKind and a boolean false.
func BinOpFromToken(t Token) (binOp BinOpKind, isBinOp bool) {
	isBinOp = true
	switch t.Kind {
	case Plus:
		binOp = Add
	case Minus:
		binOp = Sub
	case Star:
		binOp = Mul
	case Slash:
		binOp = Div
	default:
		isBinOp = false
	}
	return
}

func Parse(src []byte) Expr {
	tokens := Lex(src)
	p := newParser(tokens)
	return p.parse()
}

type parser struct {
	tokens   []Token
	tok      Token
	prev_tok Token
	pos      int
}

func newParser(tokens []Token) parser {
	p := parser{tokens: tokens}
	p.next()
	return p
}

func (p *parser) parse() Expr {
	return p.parseExpr()
}

func (p *parser) parseExpr() Expr {
	return p.parsePrecExpr(0)
}

func (p *parser) parsePrecExpr(min_prec int) Expr {
	expr := p.parsePrimaryExpr()

	for {
		op, ok := BinOpFromToken(p.tok)
		if !ok || op.Prec() < min_prec {
			break
		}
		p.next()
		rhs := p.parsePrecExpr(op.Prec())
		expr = &BinaryExpr{X: expr, Op: op, Y: rhs}
	}

	return expr
}

func (p *parser) parsePrimaryExpr() Expr {
	if ok := p.eat(Number); ok {
		return &IntegerLiteral{V: p.prev_tok.Lit}
	}

	return nil
}

// eat advances the parser to the next token and returns true, if the current token kind
// is kind k, otherwise returns false.
func (p *parser) eat(k Kind) bool {
	if p.tok.Is(k) {
		p.next()
		return true
	}
	return false
}

// next advances the parser to the next token in tokens.
func (p *parser) next() {
	if p.pos < len(p.tokens) {
		p.prev_tok = p.tok
		p.tok = p.tokens[p.pos]
		p.pos += 1
	}
}

func Evaluate(src []byte) int {
	expr := Parse(src)
	e := evaluator{}
	return e.eval(expr)
}

type evaluator struct{}

func (e *evaluator) eval(expr Expr) int {
	return e.evalExpr(expr)
}

func (e *evaluator) evalExpr(expr Expr) int {
	switch expr := expr.(type) {
	case *IntegerLiteral:
		if v, err := strconv.Atoi(expr.V); err == nil {
			return v
		}

		fmt.Printf("`%s` is not valid integer", expr.V)
		return 0
	case *BinaryExpr:
		return e.evalBinaryExpr(expr)
	}
	panic("unreachable")
}

func (e *evaluator) evalBinaryExpr(expr *BinaryExpr) int {
	x := e.evalExpr(expr.X)
	y := e.evalExpr(expr.Y)

	switch expr.Op {
	case Add:
		return x + y
	case Sub:
		return x - y
	case Mul:
		return x * y
	case Div:
		return x / y
	}

	panic("unreachable")
}

func main() {
	src := "2 + 3 * 4 + 5"
	result := Evaluate([]byte(src))
	fmt.Println(result)
}
