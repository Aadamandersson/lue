package binder

import (
	"github.com/aadamandersson/lue/internal/bir"
)

type Scope struct {
	outer *Scope
	defs  map[string]bir.Expr
}

func NewScope() *Scope {
	return &Scope{defs: make(map[string]bir.Expr, 0)}
}

func WithOuter(outer *Scope) *Scope {
	return &Scope{
		outer: outer,
		defs:  make(map[string]bir.Expr, 0),
	}
}

// Insert inserts the definition in to scope s and returns the shadowed definition and a boolean true, if any.
// Otherwise, returns nil and a boolean false.
func (s *Scope) Insert(name string, def bir.Expr) (bir.Expr, bool) {
	if d, ok := s.defs[name]; ok {
		s.defs[name] = def
		return d, true
	}
	s.defs[name] = def
	return nil, false
}

// Get returns the definition associated with name and a boolean true.
// If Get cannot find a definition associated with name in the current scope,
// it will try to find it in the outer ones, if any.
// Otherwise, returns nil and a boolean false.
func (s *Scope) Get(name string) (bir.Expr, bool) {
	if d, ok := s.defs[name]; ok {
		return d, true
	}
	if s.outer != nil {
		return s.outer.Get(name)
	}
	return nil, false
}

func (s *Scope) Functions() map[string]*bir.Fn {
	fns := make(map[string]*bir.Fn, 1)
	for _, f := range s.defs {
		switch f := f.(type) {
		case *bir.Fn:
			fns[f.Decl.Ident.Name] = f
		}
	}
	return fns
}
