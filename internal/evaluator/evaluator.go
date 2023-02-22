package evaluator

import (
	"fmt"

	"github.com/aadamandersson/lue/internal/binder"
	"github.com/aadamandersson/lue/internal/bir"
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

func Evaluate(src []byte) Value {
	aExpr := parser.Parse(src)
	expr := binder.Bind(aExpr)
	e := new()
	return e.eval(expr)
}

type evaluator struct {
	locals map[string]Value
}

func new() evaluator {
	return evaluator{locals: make(map[string]Value)}
}

func (e *evaluator) eval(expr bir.Expr) Value {
	return e.evalExpr(expr)
}

func (e *evaluator) evalExpr(expr bir.Expr) Value {
	switch expr := expr.(type) {
	case *bir.Ident:
		return e.locals[expr.Name]
	case *bir.IntegerLiteral:
		return Integer(expr.V)
	case *bir.BooleanLiteral:
		return Boolean(expr.V)
	case *bir.BinaryExpr:
		return e.evalBinaryExpr(expr)
	case *bir.AssignExpr:
		return e.evalAssignExpr(expr)
	}
	panic("unreachable")
}

func (e *evaluator) evalBinaryExpr(expr *bir.BinaryExpr) Value {
	x := e.evalExpr(expr.X)
	y := e.evalExpr(expr.Y)

	switch x := x.(type) {
	case Integer:
		y := y.(Integer)
		switch expr.Op.Kind {
		case bir.Add:
			return x + y
		case bir.Sub:
			return x - y
		case bir.Mul:
			return x * y
		case bir.Div:
			return x / y
		default:
			panic("unreachable")
		}
	}
	panic("unreachable")
}

func (e *evaluator) evalAssignExpr(expr *bir.AssignExpr) Value {
	init := e.evalExpr(expr.Init)
	e.locals[expr.Ident.Name] = init
	return Unit{}
}
