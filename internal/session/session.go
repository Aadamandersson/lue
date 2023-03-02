package session

import (
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/span"
)

type Session struct {
	Diags *diagnostic.Bag
	File  *span.SourceFile
}

func New(filename string, src []byte) *Session {
	return &Session{
		Diags: diagnostic.NewBag(),
		File:  span.NewSourceFile(filename, src),
	}
}

func (s *Session) DumpDiags() {
	s.Diags.Dump(s.File)
}
