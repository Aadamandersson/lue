package span

type Span struct {
	Start int
	End   int
}

func New(start, end int) Span {
	return Span{Start: start, End: end}
}

func NewEmpty(pos int) Span {
	return New(pos, pos)
}

// To returns a new span that encloses both span s and other.
func (s Span) To(other Span) Span {
	return New(s.Start, other.End)
}
