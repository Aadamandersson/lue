package token

import (
	"fmt"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	for k := Kind(0); k.Ok(); k++ {
		got := fmt.Sprint(k)
		if strings.HasPrefix(got, "TokenKind") {
			t.Errorf("String() does not know how to represent `%s`.\n", got)
		}
	}
}

func TestIsOneOf(t *testing.T) {
	ks := []Kind{
		Number,
		Plus,
		Minus,
	}
	cases := []struct {
		in   Token
		want bool
	}{
		{New(Number, "dummy"), true},
		{New(Plus, ""), true},
		{New(Minus, ""), true},
		{New(Star, ""), false},
	}

	for _, c := range cases {
		got := c.in.IsOneOf(ks...)
		if c.want != got {
			t.Errorf("%s IsOneOf(%v) = %t, want %t\n", c.in.Kind.String(), ks, got, c.want)
		}
	}
}
