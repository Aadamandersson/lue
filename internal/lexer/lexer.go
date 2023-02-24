package lexer

import (
	"strings"

	"github.com/aadamandersson/lue/internal/span"
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

		if isWhitespace(b) {
			l.eatWhile(isWhitespace)
			continue
		}

		start := l.pos
		l.next()
		if b == '/' && l.peek() == '/' {
			l.eatWhile(func(b byte) bool { return b != '\n' })
			continue
		}
		kind, lit := l.lexToken(b)
		tokens = append(tokens, token.New(kind, lit, span.New(start, l.pos)))
	}

	tokens = append(tokens, token.New(token.Eof, "", span.NewEmpty(l.pos)))
	return tokens
}

func (l *lexer) lexToken(first byte) (token.Kind, string) {
	peek := l.peek()
	switch first {
	case '+':
		return token.Plus, ""
	case '-':
		return token.Minus, ""
	case '*':
		return token.Star, ""
	case '/':
		return token.Slash, ""
	case '>':
		if peek == '=' {
			l.next()
			return token.Ge, ""
		}
		return token.Gt, ""
	case '<':
		if peek == '=' {
			l.next()
			return token.Le, ""
		}
		return token.Lt, ""
	case '=':
		if peek == '=' {
			l.next()
			return token.EqEq, ""
		}
		return token.Eq, ""
	case '!':
		if peek == '=' {
			l.next()
			return token.Ne, ""
		}
		return token.Unknown, string(first)
	case ':':
		return token.Colon, ""
	case ',':
		return token.Comma, ""
	case '(':
		return token.LParen, ""
	case '{':
		return token.LBrace, ""
	case ')':
		return token.RParen, ""
	case '}':
		return token.RBrace, ""
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return l.lexNumeric(first)
	default:
		if isIdentStart(first) {
			return l.lexIdent(first)
		}
		return token.Unknown, string(first)
	}
}

// lexNumeric lexes a number and returns its kind and literal value.
func (l *lexer) lexNumeric(first byte) (token.Kind, string) {
	s := l.collectString(first, isDigit)
	return token.Number, s
}

func (l *lexer) lexIdent(first byte) (token.Kind, string) {
	s := l.collectString(first, isIdentCont)
	return token.Lookup(s), s
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

// eatWhile eats bytes while matches returns true and the lexer is not at EOF.
func (l *lexer) eatWhile(matches func(byte) bool) {
	for matches(l.peek()) && !l.isEof() {
		l.next()
	}
}

// isEof returns true if the lexer is at EOF, otherwise false.
func (l *lexer) isEof() bool {
	return l.pos == len(l.src)
}

// isDigit returns true if byte b is a digit, otherwise false.
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// isIdentStart returns true if byte b is valid as a first character
// of an identifier, otherwise false.
func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

// isIdentCont returns true if byte b is valid as a non-first character
// of an identifier, otherwise false.
func isIdentCont(b byte) bool {
	return isIdentStart(b) || isDigit(b)
}

// isWhitespace returns true if byte b is whitespace, otherwise false.
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}
