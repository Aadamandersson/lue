package span

type SourceFile struct {
	Name  string
	src   []byte
	lines []int // Line beginnings in src.
}

func NewSourceFile(name string, src []byte) *SourceFile {
	lines := lines(src)
	return &SourceFile{
		Name:  name,
		src:   src,
		lines: lines,
	}
}

// Line returns the 0-based line number for pos.
func (f *SourceFile) Line(pos int) int {
	lo := 0
	hi := len(f.lines) - 1
	for lo <= hi {
		mid := (lo + hi) / 2
		curr := f.lines[mid]

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

// LinePos returns the byte pos of line in src.
// If there is no line at line, -1 is returned.
// Note that line is 0-based.
func (f *SourceFile) LinePos(line int) int {
	if line < len(f.lines) {
		return f.lines[line]
	}
	return -1

}

// LineSlice returns the bytes between start and end in src.
// If the given bounds are outside src, the zero value is returned.
func (f *SourceFile) LineSlice(lo, hi int) []byte {
	len := len(f.src)
	if lo < len && len >= hi {
		return f.src[lo:hi]
	}
	return *new([]byte)
}

// lines returns all line beginnings in src.
func lines(src []byte) []int {
	lines := []int{0}
	for i, b := range src {
		if b == '\n' {
			lines = append(lines, i+1)
		}
	}
	return lines
}
