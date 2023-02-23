package evaluator

import (
	"fmt"

	"github.com/aadamandersson/lue/internal/binder"
	"github.com/aadamandersson/lue/internal/bir"
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/parser"
	"github.com/aadamandersson/lue/internal/span"
)

type Value interface {
	String() string
	sealed()
}

type (
	Integer int
	Boolean bool
	Unit    struct{}
)

func (Integer) sealed() {}
func (Boolean) sealed() {}
func (Unit) sealed()    {}

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (b Boolean) String() string {
	return fmt.Sprintf("%t", b)
}

func (u Unit) String() string {
	return "()"
}

func Evaluate(filename string, src []byte) (Value, bool) {
	file := span.NewSourceFile(filename, src)
	diags := diagnostic.NewBag(file)
	aExpr := parser.Parse(src, diags)
	expr := binder.Bind(aExpr, diags)
	e := new(diags)
	result, ok := e.eval(expr)

	if !diags.Empty() {
		diags.Dump()
	}

	return result, ok
}

type evaluator struct {
	diags  *diagnostic.Bag
	locals []map[string]Value
}

func new(diags *diagnostic.Bag) evaluator {
	locals := make([]map[string]Value, 1)
	locals = append(locals, make(map[string]Value))
	return evaluator{
		diags:  diags,
		locals: locals,
	}
}

func (e *evaluator) eval(expr bir.Expr) (Value, bool) {
	return e.evalExpr(expr)
}

func (e *evaluator) evalExpr(expr bir.Expr) (Value, bool) {
	switch expr := expr.(type) {
	case *bir.Ident:
		return e.locals[len(e.locals)-1][expr.Name], true
	case *bir.IntegerLiteral:
		return Integer(expr.V), true
	case *bir.BooleanLiteral:
		return Boolean(expr.V), true
	case *bir.BinaryExpr:
		return e.evalBinaryExpr(expr)
	case *bir.LetExpr:
		return e.evalLetExpr(expr)
	case *bir.AssignExpr:
		return e.evalAssignExpr(expr)
	case *bir.IfExpr:
		return e.evalIfExpr(expr)
	case *bir.BlockExpr:
		return e.evalBlockExpr(expr)
	case *bir.ErrExpr:
		return nil, false
	}
	panic("unreachable")
}

func (e *evaluator) evalBinaryExpr(expr *bir.BinaryExpr) (Value, bool) {
	x, ok := e.evalExpr(expr.X)
	if !ok {
		return nil, ok
	}

	y, ok := e.evalExpr(expr.Y)
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
	}
	panic("unreachable")
}

func (e *evaluator) evalLetExpr(expr *bir.LetExpr) (Value, bool) {
	v, ok := e.evalExpr(expr.Init)
	if !ok {
		return nil, ok
	}
	e.locals[len(e.locals)-1][expr.Ident.Name] = v
	return Unit{}, true
}

func (e *evaluator) evalAssignExpr(expr *bir.AssignExpr) (Value, bool) {
	v, ok := e.evalExpr(expr.Y)
	if !ok {
		return nil, ok
	}
	// TODO: we ensure in the binder that this will always be an Ident for now,
	// but eventually we want to support more types.
	ident := expr.X.(*bir.Ident)
	e.locals[len(e.locals)-1][ident.Name] = v
	return Unit{}, true
}

func (e *evaluator) evalIfExpr(expr *bir.IfExpr) (Value, bool) {
	cond, ok := e.evalExpr(expr.Cond)
	if !ok {
		return nil, ok
	}

	if cond.(Boolean) {
		v, ok := e.evalExpr(expr.Then)
		if !ok {
			return nil, ok
		}
		return v, true
	}

	if expr.Else != nil {
		v, ok := e.evalExpr(expr.Else)
		if !ok {
			return nil, ok
		}
		return v, true
	}

	return Unit{}, true
}

func (e *evaluator) evalBlockExpr(block *bir.BlockExpr) (Value, bool) {
	var lastVal Value
	for _, expr := range block.Exprs {
		value, ok := e.evalExpr(expr)
		if !ok {
			return nil, ok
		}
		lastVal = value
	}
	return lastVal, true
}
