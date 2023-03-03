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
	fns := binder.Bind(aItems, sess)
	m := new(fns, sess, kernel)
	ok := m.interpret()

	if !sess.Diags.Empty() {
		sess.DumpDiags()
	}

	return ok
}

type machine struct {
	sess   *session.Session
	fns    map[string]*bir.Fn
	locals []map[string]Value
	kernel Kernel
}

func new(fns map[string]*bir.Fn, sess *session.Session, kernel Kernel) *machine {
	locals := make([]map[string]Value, 1)
	locals = append(locals, make(map[string]Value))
	return &machine{
		sess:   sess,
		fns:    fns,
		locals: locals,
		kernel: kernel,
	}
}

func (m *machine) interpret() bool {
	main, ok := m.fns["main"]
	if !ok {
		m.kernel.Println("no `main` function found")
		return false
	}
	_, ok = m.evalExpr(main.Body)
	return ok
}

func (m *machine) evalExpr(expr bir.Expr) (Value, bool) {
	switch expr := expr.(type) {
	case *bir.Fn:
		fn := m.fns[expr.Decl.Ident.Name]
		return &Fn{Params: fn.In, Body: fn.Body}, true
	case *bir.VarDecl:
		for i := len(m.locals) - 1; i >= 0; i-- {
			if local, ok := m.locals[i][expr.Ident.Name]; ok {
				return local, true
			}
		}
	case *bir.IntegerLiteral:
		return Integer(expr.V), true
	case *bir.BooleanLiteral:
		return Boolean(expr.V), true
	case *bir.StringLiteral:
		return String(expr.V), true
	case *bir.ArrayExpr:
		return m.evalArrayExpr(expr)
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
	case *bir.ReturnExpr:
		return m.evalReturnExpr(expr)
	case bir.Intrinsic:
		return Intrinsic(expr), true
	case *bir.ErrExpr:
		return nil, false
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
	m.locals[len(m.locals)-1][expr.Decl.Ident.Name] = v
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
	m.locals[len(m.locals)-1][decl.Ident.Name] = v
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
	m.locals = append(m.locals, make(map[string]Value))
	for _, expr := range block.Exprs {
		value, ok := m.evalExpr(expr)
		if !ok {
			return nil, ok
		}
		if retVal, ok := value.(*RetVal); ok {
			return retVal.V, true
		}
		lastVal = value
	}
	m.locals = m.locals[:len(m.locals)-1]
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

		frame := make(map[string]Value, len(expr.Args))
		for i, arg := range expr.Args {
			argVal, ok := m.evalExpr(arg)
			if !ok {
				return nil, ok
			}
			param := fn.Params[i]
			frame[param.Ident.Name] = argVal
		}
		m.locals = append(m.locals, frame)
		v, ok := m.evalExpr(fn.Body)
		m.locals = m.locals[:len(m.locals)-1]
		if !ok {
			return nil, ok
		}

		return v, true
	}

	panic("unreachable")
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
