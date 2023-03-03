package token

import (
	"strconv"

	"github.com/aadamandersson/lue/internal/span"
)

type Kind int

const (
	Unknown Kind = iota // An unknown character to the lexer.
	Eof                 // End of file.
	Ident               // E.g., `foo`
	Number              // E.g., `123`
	String              // E.g., `"foo"`
	Plus                // `+`
	Minus               // `-`
	Star                // `*`
	Slash               // `/`
	Eq                  // `=`
	Gt                  // `>`
	Lt                  // `<`
	Ge                  // `>=`
	Le                  // `<=`
	EqEq                // `==`
	Ne                  // `!=`
	Colon               // `:`
	Comma               // `,`
	LParen              // `(`
	LBrack              // `[`
	LBrace              // `{`
	RParen              // `)`
	RBrack              // `]`
	RBrace              // `}`
	Else                // `else`
	False               // `false`
	Fn                  // `fn`
	If                  // `if`
	Let                 // `let`
	Return              // `return`
	True                // `true`
	end
)

var tokens = [...]string{
	Unknown: "unknown",
	Eof:     "eof",
	Ident:   "identifier",
	Number:  "number",
	String:  "string",
	Plus:    "+",
	Minus:   "-",
	Star:    "*",
	Slash:   "/",
	Eq:      "=",
	Gt:      ">",
	Lt:      "<",
	Ge:      ">=",
	Le:      ">=",
	EqEq:    "==",
	Ne:      "!=",
	Colon:   ":",
	Comma:   ",",
	LParen:  "(",
	LBrack:  "[",
	LBrace:  "{",
	RParen:  ")",
	RBrack:  "]",
	RBrace:  "}",
	Else:    "else",
	False:   "false",
	Fn:      "fn",
	If:      "if",
	Let:     "let",
	Return:  "return",
	True:    "true",
}

func (k Kind) String() string {
	if k < 0 || k >= Kind(len(tokens)) {
		return "TokenKind(" + strconv.FormatInt(int64(k), 10) + ")"
	}
	return tokens[k]
}

// Ok returns true if kind k is a valid kind, otherwise false.
func (k Kind) Ok() bool {
	return k >= 0 && k < end
}

type Token struct {
	Kind Kind
	Lit  string // Literal value of token if kind is `Unknown` or `Number`, otherwise empty.
	Sp   span.Span
}

func New(kind Kind, lit string, span span.Span) Token {
	return Token{Kind: kind, Lit: lit, Sp: span}
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

var keywords = map[string]Kind{
	"else":   Else,
	"false":  False,
	"fn":     Fn,
	"if":     If,
	"let":    Let,
	"return": Return,
	"true":   True,
}

// Lookup returns the associated token kind for ident.
func Lookup(ident string) Kind {
	if kw, is_kw := keywords[ident]; is_kw {
		return kw
	}
	return Ident
}
