package lexer

import (
	"strings"

	"github.com/aadamandersson/lue/internal/token"
)

func Lex(src []byte) []token.Token {
	l := lexer{src: src}
	return l.Lex()
}

type lexer struct {
	src []byte
	pos int // Current position in src.
}

func (l *lexer) Lex() []token.Token {
	var tokens []token.Token
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
		tokens = append(tokens, token.New(kind, lit))
	}

	tokens = append(tokens, token.New(token.Eof, ""))
	return tokens
}

func (l *lexer) lexToken(first byte) (token.Kind, string) {
	switch first {
	case '+':
		return token.Plus, ""
	case '-':
		return token.Minus, ""
	case '*':
		return token.Star, ""
	case '/':
		return token.Slash, ""
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return l.lexNumeric(first)
	default:
		return token.Unknown, string(first)
	}
}

// lexNumeric lexes a number and returns its kind and literal value.
func (l *lexer) lexNumeric(first byte) (token.Kind, string) {
	s := l.collectString(first, l.isDigit)
	return token.Number, s
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
