package lexer

import (
	"testing"

	"github.com/aadamandersson/lue/internal/token"
)

var cases = []struct {
	in   string
	want token.Token
}{
	{"", token.New(token.Eof, "")},
	{" ", token.New(token.Eof, "")},
	{"\t", token.New(token.Eof, "")},
	{"\r", token.New(token.Eof, "")},
	{"\r\n", token.New(token.Eof, "")},
	{"123", token.New(token.Number, "123")},
	{"+", token.New(token.Plus, "")},
	{"-", token.New(token.Minus, "")},
	{"*", token.New(token.Star, "")},
	{"/", token.New(token.Slash, "")},
}

func TestLex(t *testing.T) {
	for _, c := range cases {
		got := Lex([]byte(c.in))[0]
		if got != c.want {
			t.Errorf("Lex(\"%s\") = %#v, want %#v\n", c.in, got, c.want)
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
