package evaluator

import (
	"github.com/aadamandersson/lue/internal/binder"
	"github.com/aadamandersson/lue/internal/bir"
	"github.com/aadamandersson/lue/internal/parser"
)

func Evaluate(src []byte) int {
	aExpr := parser.Parse(src)
	expr := binder.Bind(aExpr)
	e := new()
	return e.eval(expr)
}

type evaluator struct {
	locals map[string]int
}

func new() evaluator {
	return evaluator{locals: make(map[string]int)}
}

func (e *evaluator) eval(expr bir.Expr) int {
	return e.evalExpr(expr)
}

func (e *evaluator) evalExpr(expr bir.Expr) int {
	switch expr := expr.(type) {
	case *bir.Ident:
		return e.locals[expr.Name]
	case *bir.IntegerLiteral:
		return expr.V
	case *bir.BinaryExpr:
		return e.evalBinaryExpr(expr)
	case *bir.AssignExpr:
		return e.evalAssignExpr(expr)
	}
	panic("unreachable")
}

func (e *evaluator) evalBinaryExpr(expr *bir.BinaryExpr) int {
	x := e.evalExpr(expr.X)
	y := e.evalExpr(expr.Y)

	switch expr.Op.Kind {
	case bir.Add:
		return x + y
	case bir.Sub:
		return x - y
	case bir.Mul:
		return x * y
	case bir.Div:
		return x / y
	}

	panic("unreachable")
}

func (e *evaluator) evalAssignExpr(expr *bir.AssignExpr) int {
	init := e.evalExpr(expr.Init)
	e.locals[expr.Ident.Name] = init
	return init
}
