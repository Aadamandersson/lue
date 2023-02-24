package binder

import (
	"github.com/aadamandersson/lue/internal/bir"
)

type Scope struct {
	outer       *Scope
	definitions map[string]bir.Definition
}

func NewScope() *Scope {
	return &Scope{definitions: make(map[string]bir.Definition, 0)}
}

func WithOuter(outer *Scope) *Scope {
	return &Scope{
		outer:       outer,
		definitions: make(map[string]bir.Definition, 0),
	}
}

// Insert inserts the definition in to scope s and returns the shadowed definition and a boolean true, if any.
// Otherwise, returns nil and a boolean false.
func (s *Scope) Insert(name string, definition bir.Definition) (bir.Definition, bool) {
	if d, ok := s.definitions[name]; ok {
		s.definitions[name] = definition
		return d, true
	}
	s.definitions[name] = definition
	return nil, false
}

// Get returns the definition associated with name and a boolean true.
// If Get cannot find a definition associated with name in the current scope,
// it will try to find it in the outer ones, if any.
// Otherwise, returns nil and a boolean false.
func (s *Scope) Get(name string) (bir.Definition, bool) {
	if d, ok := s.definitions[name]; ok {
		return d, true
	}
	if s.outer != nil {
		return s.outer.Get(name)
	}
	return nil, false
}
