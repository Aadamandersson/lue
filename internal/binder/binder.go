package binder

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/ast"
	"github.com/aadamandersson/lue/internal/bir"
)

func Bind(expr ast.Expr) bir.Expr {
	b := binder{}
	return b.bind(expr)
}

type binder struct{}

func (b *binder) bind(expr ast.Expr) bir.Expr {
	return b.bindExpr(expr)
}

func (b *binder) bindExpr(expr ast.Expr) bir.Expr {
	switch expr := expr.(type) {
	case *ast.IntegerLiteral:
		if v, err := strconv.Atoi(expr.V); err == nil {
			return &bir.IntegerLiteral{V: v}
		}

		fmt.Printf("`%s` is not valid integer", expr.V)
		return &bir.ErrExpr{}
	case *ast.BinaryExpr:
		return b.bindBinaryExpr(expr)
	}
	panic("unreachable")
}

func (b *binder) bindBinaryExpr(expr *ast.BinaryExpr) bir.Expr {
	x := b.bindExpr(expr.X)
	y := b.bindExpr(expr.Y)
	op, ok := bir.BindBinOp(expr.Op, x.Type(), y.Type())

	if !ok {
		switch expr.Op {
		case ast.Add:
			fmt.Printf("cannot add `%s` to `%s`\n", x.Type(), y.Type())
		case ast.Sub:
			fmt.Printf("cannot subtract `%s` from `%s`\n", y.Type(), x.Type())
		case ast.Mul:
			fmt.Printf("cannot multiply `%s` by `%s`\n", x.Type(), y.Type())
		case ast.Div:
			fmt.Printf("cannot divide `%s` by `%s`\n", x.Type(), y.Type())
		default:
			panic("unreachable")
		}
		return &bir.ErrExpr{}
	}

	return &bir.BinaryExpr{X: x, Op: op, Y: y}
}