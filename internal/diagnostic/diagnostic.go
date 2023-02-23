package diagnostic

import (
	"fmt"
	"strings"

	"github.com/aadamandersson/lue/internal/span"
)

type (
	Bag struct {
		file  *span.SourceFile
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

func NewBag(file *span.SourceFile) *Bag {
	return &Bag{file: file, diags: make([]*Diagnostic, 0)}
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

func (b *Bag) Dump() {
	var builder strings.Builder
	indent := func() { builder.WriteString(strings.Repeat(" ", 4)) }
	for _, d := range b.diags {
		builder.WriteByte('\n')
		errStr := fmt.Sprintf("error: %s\n", d.Msg)
		builder.WriteString(errStr)

		line := b.file.Line(d.Span.Start)
		lineStart := b.file.LinePos(line)
		lineEnd := b.file.LinePos(line + 1)
		col := d.Span.Start - lineStart + 1
		fLoc := fmt.Sprintf("[%s:%d:%d]\n", b.file.Name, col, line+1)
		builder.WriteString(fLoc)

		indent()
		errLine := b.file.LineSlice(lineStart, lineEnd)
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
