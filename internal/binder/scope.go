package binder

import (
	"github.com/aadamandersson/lue/internal/bir"
)

type Scope struct {
	outer  *Scope
	values map[string]bir.Expr
}

func NewScope() *Scope {
	return &Scope{values: make(map[string]bir.Expr, 0)}
}

func WithOuter(outer *Scope) *Scope {
	return &Scope{
		outer:  outer,
		values: make(map[string]bir.Expr, 0),
	}
}

// Insert inserts the value in to scope s and returns the shadowed value and a boolean true, if any.
// Otherwise, returns nil and a boolean false.
func (s *Scope) Insert(name string, value bir.Expr) (bir.Expr, bool) {
	if v, ok := s.values[name]; ok {
		s.values[name] = value
		return v, true
	}
	s.values[name] = value
	return nil, false
}

// Get returns the value associated with name and a boolean true.
// If Get cannot find a value associated with name in the current scope,
// it will try to find it in the outer ones, if any.
// Otherwise, returns nil and a boolean false.
func (s *Scope) Get(name string) (bir.Expr, bool) {
	if v, ok := s.values[name]; ok {
		return v, true
	}
	if s.outer != nil {
		return s.outer.Get(name)
	}
	return nil, false
}
