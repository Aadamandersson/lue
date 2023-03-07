package machine

import (
	"github.com/aadamandersson/lue/internal/binder"
	"github.com/aadamandersson/lue/internal/ir"
	"github.com/aadamandersson/lue/internal/ir/bir"
	"github.com/aadamandersson/lue/internal/parser"
	"github.com/aadamandersson/lue/internal/session"
)

func Interpret(filename string, src []byte, kernel Kernel) bool {
	sess := session.New(filename, src)
	aItems := parser.Parse(sess)
	classes, fns := binder.Bind(aItems, sess)
	m := newMachine(classes, fns, sess, kernel)
	ok := m.interpret()

	if !sess.Diags.Empty() {
		sess.DumpDiags()
	}

	return ok
}

type stack struct {
	frames []*frame
}

func newStack() *stack {
	return &stack{frames: make([]*frame, 0)}
}

func (s *stack) push(f *frame) {
	s.frames = append(s.frames, f)
}

func (s *stack) pop() *frame {
	n := len(s.frames)
	if n == 0 {
		return nil
	}
	frame := s.frames[n-1]
	s.frames = s.frames[:n-1]
	return frame
}

func (s *stack) peek() *frame {
	n := len(s.frames)
	if n == 0 {
		return nil
	}
	return s.frames[n-1]
}

type frame struct {
	locals map[*bir.VarDecl]Value
}

func newFrame(locals map[*bir.VarDecl]Value) *frame {
	return &frame{locals: locals}
}

func (f *frame) local(decl *bir.VarDecl) Value {
	return f.locals[decl]
}

type machine struct {
	sess    *session.Session
	classes map[string]*bir.Class
	fns     map[string]*bir.Fn
	stack   *stack
	kernel  Kernel
}

func newMachine(
	classes map[string]*bir.Class,
	fns map[string]*bir.Fn,
	sess *session.Session,
	kernel Kernel,
) *machine {
	return &machine{
		sess:    sess,
		classes: classes,
		fns:     fns,
		stack:   newStack(),
		kernel:  kernel,
	}
}

func (m *machine) interpret() bool {
	main, ok := m.fns["main"]
	if !ok {
		m.kernel.Println("no `main` function found")
		return false
	}
	m.stack.push(newFrame(map[*bir.VarDecl]Value{}))
	_, ok = m.evalExpr(main.Body)
	return ok
}

func (m *machine) evalExpr(expr bir.Expr) (Value, bool) {
	switch expr := expr.(type) {
	case *bir.Fn:
		fn := m.fns[expr.Decl.Ident.Name]
		return &Fn{Params: fn.In, Body: fn.Body}, true
	case *bir.VarDecl:
		return m.stack.peek().local(expr), true
	case *bir.IntegerLiteral:
		return Integer(expr.V), true
	case *bir.BooleanLiteral:
		return Boolean(expr.V), true
	case *bir.StringLiteral:
		return String(expr.V), true
	case *bir.BinaryExpr:
		return m.evalBinaryExpr(expr)
	case *bir.LetExpr:
		return m.evalLetExpr(expr)
	case *bir.AssignExpr:
		return m.evalAssignExpr(expr)
	case *bir.IfExpr:
		return m.evalIfExpr(expr)
	case *bir.BlockExpr:
		return m.evalBlockExpr(expr)
	case *bir.CallExpr:
		return m.evalCallExpr(expr)
	case *bir.ArrayExpr:
		return m.evalArrayExpr(expr)
	case *bir.IndexExpr:
		return m.evalIndexExpr(expr)
	case *bir.ForExpr:
		return m.evalForExpr(expr)
	case *bir.BreakExpr:
		return m.evalBreakExpr(expr)
	case *bir.ReturnExpr:
		return m.evalReturnExpr(expr)
	case bir.Intrinsic:
		return Intrinsic(expr), true
	case *bir.ErrExpr:
		return nil, false
	}

	panic("unreachable")
}

func (m *machine) evalBinaryExpr(expr *bir.BinaryExpr) (Value, bool) {
	x, ok := m.evalExpr(expr.X)
	if !ok {
		return nil, ok
	}

	y, ok := m.evalExpr(expr.Y)
	if !ok {
		return nil, ok
	}

	switch x := x.(type) {
	case Integer:
		y := y.(Integer)
		switch expr.Op.Kind {
		case bir.Add:
			return x + y, true
		case bir.Sub:
			return x - y, true
		case bir.Mul:
			return x * y, true
		case bir.Div:
			return x / y, true
		case bir.Gt:
			return Boolean(x > y), true
		case bir.Lt:
			return Boolean(x < y), true
		case bir.Ge:
			return Boolean(x >= y), true
		case bir.Le:
			return Boolean(x <= y), true
		case bir.Eq:
			return Boolean(x == y), true
		case bir.Ne:
			return Boolean(x != y), true
		}
	case Boolean:
		y := y.(Boolean)
		switch expr.Op.Kind {
		case bir.Eq:
			return Boolean(x == y), true
		case bir.Ne:
			return Boolean(x != y), true
		}
	case String:
		y := y.(String)
		switch expr.Op.Kind {
		case bir.Eq:
			return Boolean(x == y), true
		case bir.Ne:
			return Boolean(x != y), true
		}
	}
	panic("unreachable")
}

func (m *machine) evalLetExpr(expr *bir.LetExpr) (Value, bool) {
	v, ok := m.evalExpr(expr.Init)
	if !ok {
		return nil, ok
	}
	m.stack.peek().locals[expr.Decl] = v
	return Unit{}, true
}

func (m *machine) evalAssignExpr(expr *bir.AssignExpr) (Value, bool) {
	v, ok := m.evalExpr(expr.Y)
	if !ok {
		return nil, ok
	}
	// TODO: we ensure in the binder that this will always be an Ident for now,
	// but eventually we want to support more types.
	decl := expr.X.(*bir.VarDecl)
	m.stack.peek().locals[decl] = v
	return Unit{}, true
}

func (m *machine) evalIfExpr(expr *bir.IfExpr) (Value, bool) {
	cond, ok := m.evalExpr(expr.Cond)
	if !ok {
		return nil, ok
	}

	if cond.(Boolean) {
		v, ok := m.evalExpr(expr.Then)
		if !ok {
			return nil, ok
		}
		return v, true
	}

	if expr.Else != nil {
		v, ok := m.evalExpr(expr.Else)
		if !ok {
			return nil, ok
		}
		return v, true
	}

	return Unit{}, true
}

func (m *machine) evalBlockExpr(block *bir.BlockExpr) (Value, bool) {
	var lastVal Value

	for _, expr := range block.Exprs {
		value, ok := m.evalExpr(expr)
		if !ok {
			return nil, ok
		}

		if rv, ok := value.(*RetVal); ok {
			return rv.V, ok
		}

		if bv, ok := value.(*BreakVal); ok {
			return bv, ok
		}

		lastVal = value
	}
	return lastVal, true
}

func (m *machine) evalCallExpr(expr *bir.CallExpr) (Value, bool) {
	fnVal, ok := m.evalExpr(expr.Fn)
	if !ok {
		return nil, ok
	}

	switch fn := fnVal.(type) {
	case Intrinsic:
		switch fn {
		case Intrinsic(ir.IntrPrintln):
			arg, ok := m.evalExpr(expr.Args[0])
			if !ok {
				return nil, ok
			}
			m.kernel.Println(arg.String())
			return Unit{}, true
		}
	case *Fn:
		if expr.Args == nil {
			return m.evalExpr(fn.Body)
		}

		locals := make(map[*bir.VarDecl]Value, len(expr.Args))
		for i, arg := range expr.Args {
			argVal, ok := m.evalExpr(arg)
			if !ok {
				return nil, ok
			}
			param := fn.Params[i]
			locals[param] = argVal
		}
		m.stack.push(newFrame(locals))
		v, ok := m.evalExpr(fn.Body)
		m.stack.pop()
		if !ok {
			return nil, ok
		}

		return v, true
	}

	panic("unreachable")
}

func (m *machine) evalArrayExpr(expr *bir.ArrayExpr) (Value, bool) {
	var elems []Value

	for _, expr := range expr.Exprs {
		v, ok := m.evalExpr(expr)
		if !ok {
			return nil, false
		}
		elems = append(elems, v)
	}

	return &Array{Elems: elems}, true
}

func (m *machine) evalIndexExpr(expr *bir.IndexExpr) (Value, bool) {
	arrExpr, ok := m.evalExpr(expr.Arr)
	if !ok {
		return nil, ok
	}

	idxExpr, ok := m.evalExpr(expr.I)
	if !ok {
		return nil, ok
	}
	// TODO: out of bounds handling
	arr := arrExpr.(*Array)
	i := idxExpr.(Integer)
	return arr.Elems[i], true
}

func (m *machine) evalForExpr(expr *bir.ForExpr) (Value, bool) {
	for {
		v, ok := m.evalExpr(expr.Body)
		if !ok {
			return nil, ok
		}

		if bv, ok := v.(*BreakVal); ok {
			return bv.V, true
		}
	}
}

func (m *machine) evalBreakExpr(expr *bir.BreakExpr) (Value, bool) {
	if expr.X != nil {
		v, ok := m.evalExpr(expr.X)
		if !ok {
			return nil, ok
		}
		return &BreakVal{V: v}, true
	}
	return &BreakVal{V: Unit{}}, true
}

func (m *machine) evalReturnExpr(expr *bir.ReturnExpr) (Value, bool) {
	if expr.X != nil {
		v, ok := m.evalExpr(expr.X)
		if !ok {
			return nil, ok
		}
		return &RetVal{V: v}, true
	}
	return &RetVal{V: Unit{}}, true
}
