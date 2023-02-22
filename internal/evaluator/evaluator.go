package evaluator

import (
	"fmt"

	"github.com/aadamandersson/lue/internal/binder"
	"github.com/aadamandersson/lue/internal/bir"
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/parser"
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

func Evaluate(src []byte) (Value, bool) {
	diags := diagnostic.NewBag()
	aExpr := parser.Parse(src, diags)
	expr := binder.Bind(aExpr, diags)
	e := new(diags)
	result, ok := e.eval(expr)

	if !diags.Empty() {
		diags.ForEach(func(d *diagnostic.Diagnostic) bool {
			fmt.Println(d.Msg)
			return false
		})
	}

	return result, ok
}

type evaluator struct {
	diags  *diagnostic.Bag
	locals map[string]Value
}

func new(diags *diagnostic.Bag) evaluator {
	return evaluator{
		diags:  diags,
		locals: make(map[string]Value),
	}
}

func (e *evaluator) eval(expr bir.Expr) (Value, bool) {
	return e.evalExpr(expr)
}

func (e *evaluator) evalExpr(expr bir.Expr) (Value, bool) {
	switch expr := expr.(type) {
	case *bir.Ident:
		return e.locals[expr.Name], true
	case *bir.IntegerLiteral:
		return Integer(expr.V), true
	case *bir.BooleanLiteral:
		return Boolean(expr.V), true
	case *bir.BinaryExpr:
		return e.evalBinaryExpr(expr)
	case *bir.AssignExpr:
		return e.evalAssignExpr(expr)
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

func (e *evaluator) evalAssignExpr(expr *bir.AssignExpr) (Value, bool) {
	init, ok := e.evalExpr(expr.Init)
	if !ok {
		return nil, ok
	}
	e.locals[expr.Ident.Name] = init
	return Unit{}, true
}
