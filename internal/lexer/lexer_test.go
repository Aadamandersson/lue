package lexer

import (
	"testing"

	"github.com/aadamandersson/lue/internal/session"
	"github.com/aadamandersson/lue/internal/span"
	"github.com/aadamandersson/lue/internal/token"
)

var cases = []struct {
	in   string
	want token.Token
}{
	{"", token.New(token.Eof, "", span.NewEmpty(0))},
	{" ", token.New(token.Eof, "", span.NewEmpty(1))},
	{"\t", token.New(token.Eof, "", span.NewEmpty(1))},
	{"\r", token.New(token.Eof, "", span.NewEmpty(1))},
	{"\n", token.New(token.Eof, "", span.NewEmpty(1))},
	{"\r\n", token.New(token.Eof, "", span.NewEmpty(2))},
	{"// some comment", token.New(token.Eof, "", span.New(15, 15))},
	{"foo", token.New(token.Ident, "foo", span.New(0, 3))},
	{"_foo", token.New(token.Ident, "_foo", span.New(0, 4))},
	{"foo123", token.New(token.Ident, "foo123", span.New(0, 6))},
	{"123", token.New(token.Number, "123", span.New(0, 3))},
	{`"foo"`, token.New(token.String, "foo", span.New(0, 5))},
	{`"foo\"bar\""`, token.New(token.String, `foo"bar"`, span.New(0, 12))},
	{"+", token.New(token.Plus, "", span.New(0, 1))},
	{"-", token.New(token.Minus, "", span.New(0, 1))},
	{"*", token.New(token.Star, "", span.New(0, 1))},
	{"/", token.New(token.Slash, "", span.New(0, 1))},
	{"=", token.New(token.Eq, "", span.New(0, 1))},
	{">", token.New(token.Gt, "", span.New(0, 1))},
	{"<", token.New(token.Lt, "", span.New(0, 1))},
	{">=", token.New(token.Ge, "", span.New(0, 2))},
	{"<=", token.New(token.Le, "", span.New(0, 2))},
	{"==", token.New(token.EqEq, "", span.New(0, 2))},
	{"!=", token.New(token.Ne, "", span.New(0, 2))},
	{":", token.New(token.Colon, "", span.New(0, 1))},
	{",", token.New(token.Comma, "", span.New(0, 1))},
	{"(", token.New(token.LParen, "", span.New(0, 1))},
	{"[", token.New(token.LBrack, "", span.New(0, 1))},
	{"{", token.New(token.LBrace, "", span.New(0, 1))},
	{")", token.New(token.RParen, "", span.New(0, 1))},
	{"]", token.New(token.RBrack, "", span.New(0, 1))},
	{"}", token.New(token.RBrace, "", span.New(0, 1))},
	{"else", token.New(token.Else, "else", span.New(0, 4))},
	{"false", token.New(token.False, "false", span.New(0, 5))},
	{"fn", token.New(token.Fn, "fn", span.New(0, 2))},
	{"if", token.New(token.If, "if", span.New(0, 2))},
	{"let", token.New(token.Let, "let", span.New(0, 3))},
	{"return", token.New(token.Return, "return", span.New(0, 6))},
	{"true", token.New(token.True, "true", span.New(0, 4))},
}

func TestLex(t *testing.T) {
	for _, c := range cases {
		got := lex(c.in)[0]
		if got != c.want {
			t.Errorf("Lex(\"%s\") = %+v, want %+v\n", c.in, got, c.want)
		}
	}
}

func TestLexerLexesAllTokens(t *testing.T) {
	for k := token.Kind(1); k.Ok(); k++ {
		found := false
		for _, c := range cases {
			if k == c.want.Kind {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Token `%s` is not handled by the lexer.\n", k.String())
		}
	}
}

func lex(src string) []token.Token {
	sess := session.New("test", []byte(src))
	return Lex(sess)
}
