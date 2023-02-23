package binder

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/ast"
	"github.com/aadamandersson/lue/internal/bir"
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/span"
)

func Bind(expr ast.Expr, diags *diagnostic.Bag) bir.Expr {
	b := new(diags)
	return b.bind(expr)
}

type binder struct {
	diags  *diagnostic.Bag
	values map[string]bir.Expr
}

func new(diags *diagnostic.Bag) binder {
	return binder{
		diags:  diags,
		values: make(map[string]bir.Expr),
	}
}

func (b *binder) bind(expr ast.Expr) bir.Expr {
	return b.bindExpr(expr)
}

func (b *binder) bindExpr(expr ast.Expr) bir.Expr {
	switch expr := expr.(type) {
	case *ast.Ident:
		if e, ok := b.values[expr.Name]; ok {
			return &bir.Ident{Name: expr.Name, Ty: e.Type()}
		}
		b.error(expr.Sp, "could not find anything named `%s`", expr.Name)
		return &bir.ErrExpr{}
	case *ast.IntegerLiteral:
		if v, err := strconv.Atoi(expr.V); err == nil {
			return &bir.IntegerLiteral{V: v}
		}

		b.error(expr.Sp, "`%s` is not valid integer", expr.V)
		return &bir.ErrExpr{}
	case *ast.BooleanLiteral:
		return &bir.BooleanLiteral{V: expr.V}
	case *ast.BinaryExpr:
		return b.bindBinaryExpr(expr)
	case *ast.LetExpr:
		return b.bindLetExpr(expr)
	case *ast.AssignExpr:
		return b.bindAssignExpr(expr)
	case *ast.IfExpr:
		return b.bindIfExpr(expr)
	case *ast.BlockExpr:
		return b.bindBlockExpr(expr)
	case *ast.ErrExpr:
		return &bir.ErrExpr{}
	}
	panic("unreachable")
}

func (b *binder) bindBinaryExpr(expr *ast.BinaryExpr) bir.Expr {
	x := b.bindExpr(expr.X)
	y := b.bindExpr(expr.Y)
	op, ok := bir.BindBinOp(expr.Op.Kind, x.Type(), y.Type())

	if !ok {
		sp := expr.Op.Sp
		switch expr.Op.Kind {
		case ast.Add:
			b.error(sp, "cannot add `%s` to `%s`", x.Type(), y.Type())
		case ast.Sub:
			b.error(sp, "cannot subtract `%s` from `%s`", y.Type(), x.Type())
		case ast.Mul:
			b.error(sp, "cannot multiply `%s` by `%s`", x.Type(), y.Type())
		case ast.Div:
			b.error(sp, "cannot divide `%s` by `%s`", x.Type(), y.Type())
		case ast.Gt, ast.Lt, ast.Ge, ast.Le, ast.Eq, ast.Ne:
			b.error(sp, "cannot compare `%s` with `%s`", x.Type(), y.Type())
		default:
			panic("unreachable")
		}
		return &bir.ErrExpr{}
	}

	return &bir.BinaryExpr{X: x, Op: op, Y: y}
}

func (b *binder) bindLetExpr(expr *ast.LetExpr) bir.Expr {
	init := b.bindExpr(expr.Init)
	b.values[expr.Ident.Name] = init
	ident := &bir.Ident{Name: expr.Ident.Name, Ty: init.Type()}
	return &bir.LetExpr{Ident: ident, Init: init}
}

func (b *binder) bindAssignExpr(expr *ast.AssignExpr) bir.Expr {
	x := b.bindExpr(expr.X)
	y := b.bindExpr(expr.Y)
	switch x := x.(type) {
	case *bir.Ident:
		if v, ok := b.values[x.Name]; ok {
			if v.Type() != y.Type() {
				b.error(expr.Y.Span(), "expected `%s`, but got `%s`", v.Type(), y.Type())
				return &bir.ErrExpr{}
			} else {
				b.values[x.Name] = y
			}
		}
		return &bir.AssignExpr{X: &bir.Ident{Name: x.Name, Ty: y.Type()}, Y: y}
	default:
		b.error(expr.X.Span(), "can only assign to identifiers for now")
		return &bir.ErrExpr{}
	}
}

func (b *binder) bindIfExpr(expr *ast.IfExpr) bir.Expr {
	cond := b.bindExpr(expr.Cond)
	if cond.Type() != bir.TBool {
		b.error(expr.Cond.Span(), "expected `bool`, but got %s", cond.Type())
		return &bir.ErrExpr{}
	}

	then := b.bindExpr(expr.Then)
	var els bir.Expr
	if expr.Else != nil {
		els = b.bindExpr(expr.Else)
		if then.Type() != els.Type() {
			b.error(
				expr.Span(),
				"`if` and else have incompatible types, expected `%s`, but got `%s`",
				then.Type(),
				els.Type(),
			)
		}
	}
	return &bir.IfExpr{Cond: cond, Then: then, Else: els}
}

func (b *binder) bindBlockExpr(expr *ast.BlockExpr) bir.Expr {
	var exprs []bir.Expr
	for _, e := range expr.Exprs {
		exprs = append(exprs, b.bindExpr(e))
	}
	return &bir.BlockExpr{Exprs: exprs}
}

func (b *binder) error(span span.Span, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	diagnostic.NewBuilder(msg, span).WithLabel("here").Emit(b.diags)
}
