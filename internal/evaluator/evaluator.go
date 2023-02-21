package evaluator

import (
	"github.com/aadamandersson/lue/internal/binder"
	"github.com/aadamandersson/lue/internal/bir"
	"github.com/aadamandersson/lue/internal/parser"
)

func Evaluate(src []byte) int {
	aExpr := parser.Parse(src)
	expr := binder.Bind(aExpr)
	e := evaluator{}
	return e.eval(expr)
}

type evaluator struct{}

func (e *evaluator) eval(expr bir.Expr) int {
	return e.evalExpr(expr)
}

func (e *evaluator) evalExpr(expr bir.Expr) int {
	switch expr := expr.(type) {
	case *bir.IntegerLiteral:
		return expr.V
	case *bir.BinaryExpr:
		return e.evalBinaryExpr(expr)
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
