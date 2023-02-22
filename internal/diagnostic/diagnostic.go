package diagnostic

import "github.com/aadamandersson/lue/internal/span"

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
