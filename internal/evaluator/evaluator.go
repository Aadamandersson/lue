package evaluator

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/ast"
	"github.com/aadamandersson/lue/internal/parser"
)

func Evaluate(src []byte) int {
	expr := parser.Parse(src)
	e := evaluator{}
	return e.eval(expr)
}

type evaluator struct{}

func (e *evaluator) eval(expr ast.Expr) int {
	return e.evalExpr(expr)
}

func (e *evaluator) evalExpr(expr ast.Expr) int {
	switch expr := expr.(type) {
	case *ast.IntegerLiteral:
		if v, err := strconv.Atoi(expr.V); err == nil {
			return v
		}

		fmt.Printf("`%s` is not valid integer", expr.V)
		return 0
	case *ast.BinaryExpr:
		return e.evalBinaryExpr(expr)
	}
	panic("unreachable")
}

func (e *evaluator) evalBinaryExpr(expr *ast.BinaryExpr) int {
	x := e.evalExpr(expr.X)
	y := e.evalExpr(expr.Y)

	switch expr.Op {
	case ast.Add:
		return x + y
	case ast.Sub:
		return x - y
	case ast.Mul:
		return x * y
	case ast.Div:
		return x / y
	}

	panic("unreachable")
}
