package diagnostic

import (
	"fmt"
	"strings"

	"github.com/aadamandersson/lue/internal/span"
)

type (
	Bag struct {
		diags []*Diagnostic
	}

	Diagnostic struct {
		Msg    string
		Span   span.Span
		Labels []*Label
	}

	Label struct {
		Msg  string
		Span span.Span
	}

	Builder struct {
		msg    string
		span   span.Span
		labels []*Label
	}
)

func NewBag() *Bag {
	return &Bag{diags: make([]*Diagnostic, 0)}
}

func (b *Bag) Empty() bool {
	return len(b.diags) == 0
}

func (b *Bag) ForEach(f func(*Diagnostic) bool) {
	for _, d := range b.diags {
		if f(d) {
			break
		}
	}
}

func (b *Bag) Dump(filename string, src []byte) {
	var builder strings.Builder
	lines := lines(src)
	indent := func() { builder.WriteString(strings.Repeat(" ", 4)) }
	for _, d := range b.diags {
		builder.WriteByte('\n')
		errStr := fmt.Sprintf("error: %s\n", d.Msg)
		builder.WriteString(errStr)
		lineIdx := lineIdx(d.Span.Start, lines)
		lineStart := lines[lineIdx]
		lineEnd := lines[lineIdx+1]
		col := d.Span.Start - lineStart + 1

		fLoc := fmt.Sprintf("[%s:%d:%d]\n", filename, col, lineIdx+1)
		builder.WriteString(fLoc)

		indent()
		errLine := src[lineStart:lineEnd]
		builder.WriteString(string(errLine))
		if len(d.Labels) != 0 {
			label := d.Labels[0] // FIXME: support multiple labels (secondary ones)

			indent()
			labelStart := label.Span.Start - lineStart
			builder.WriteString(strings.Repeat(" ", labelStart))
			builder.WriteString("^ ")
			builder.WriteString(label.Msg)
			builder.WriteByte('\n')
		}
	}

	fmt.Println(builder.String())
}

func lines(src []byte) []int {
	lines := []int{0}
	cursor := 0
	for _, b := range src {
		if b == '\n' {
			lines = append(lines, cursor+1)
		}
		cursor += 1
	}
	return lines
}

func lineIdx(pos int, lines []int) int {
	lo := 0
	hi := len(lines) - 1
	for lo <= hi {
		mid := (lo + hi) / 2
		curr := lines[mid]

		if curr == pos {
			return mid
		}

		if pos > curr {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return lo - 1
}

func (d *Diagnostic) Emit(bag *Bag) {
	bag.diags = append(bag.diags, d)
}

func NewBuilder(msg string, span span.Span) *Builder {
	return &Builder{
		msg:    msg,
		span:   span,
		labels: make([]*Label, 0),
	}
}

func (b *Builder) WithLabel(msg string) *Builder {
	l := &Label{
		Msg:  msg,
		Span: b.span,
	}
	b.labels = append(b.labels, l)
	return b
}

func (b *Builder) Build() *Diagnostic {
	return &Diagnostic{
		Msg:    b.msg,
		Span:   b.span,
		Labels: b.labels,
	}
}

func (b *Builder) Emit(bag *Bag) {
	b.Build().Emit(bag)
}
