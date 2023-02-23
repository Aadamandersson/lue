package span

import (
	"testing"
)

func TestLine(t *testing.T) {
	src := "12\n34\n56\n789"
	f := NewSourceFile("filename", []byte(src))
	cases := []struct {
		in   int
		want int
	}{
		{0, 0}, {1, 0}, {2, 0},
		{3, 1}, {4, 1}, {5, 1},
		{6, 2}, {7, 2}, {8, 2},
		{9, 3}, {10, 3}, {11, 3}, {12, 3},
	}

	for _, c := range cases {
		got := f.Line(c.in)
		if got != c.want {
			t.Errorf("Line(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestLinePos(t *testing.T) {
	src := "12\n34\n56\n789"
	f := NewSourceFile("filename", []byte(src))
	cases := []struct {
		in   int
		want int
	}{
		{0, 0}, {1, 3}, {2, 6}, {3, 9}, {4, -1},
	}

	for _, c := range cases {
		got := f.LinePos(c.in)
		if got != c.want {
			t.Errorf("LinePos(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestLineSlice(t *testing.T) {
	src := "12\n34\n56\n789"
	f := NewSourceFile("filename", []byte(src))
	cases := []struct {
		inLo int
		inHi int
		want string
	}{
		{0, 3, "12\n"}, {6, 9, "56\n"},
		{9, 12, "789"}, {12, 15, ""},
	}

	for _, c := range cases {
		got := string(f.LineSlice(c.inLo, c.inHi))
		if got != c.want {
			t.Errorf("LineSlice(%d, %d) = %s, want %s", c.inLo, c.inHi, got, c.want)
		}
	}
}
