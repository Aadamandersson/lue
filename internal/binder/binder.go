package binder

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/ast"
	"github.com/aadamandersson/lue/internal/bir"
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/span"
)

func Bind(items []ast.Item, diags *diagnostic.Bag) []bir.Item {
	b := new(diags)
	return b.bind(items)
}

type binder struct {
	diags *diagnostic.Bag
	scope *Scope
}

func new(diags *diagnostic.Bag) binder {
	return binder{diags: diags}
}

func (b *binder) bind(aItems []ast.Item) []bir.Item {
	var items []bir.Item
	for _, aItem := range aItems {
		items = append(items, b.bindItem(aItem))
	}
	return items
}

func (b *binder) bindItem(item ast.Item) bir.Item {
	switch item := item.(type) {
	case *ast.FnDecl:
		b.scope = WithOuter(b.scope)
		return b.bindFnDecl(item)
	case *ast.ErrItem:
		return &bir.ErrItem{}
	}
	panic("unreachable")
}

func (b *binder) bindFnDecl(decl *ast.FnDecl) bir.Item {
	params := b.bindParams(decl.In)
	var ty bir.Ty
	if decl.Out == nil {
		ty = bir.TUnit
	} else {
		ty = lookupTy(decl.Out.Name)
		if ty == bir.TErr {
			b.error(decl.Out.Sp, "cannot find type `%s` in this scope", decl.Out.Name)
			return &bir.ErrItem{}
		}
	}

	ident := &bir.Ident{Name: decl.Ident.Name, Ty: ty}
	body := b.bindExpr(decl.Body)
	if ty != body.Type() {
		if decl.Out != nil {
			b.error(
				decl.Out.Sp,
				"expected this function to return `%s`, but got `%s`",
				ty,
				body.Type(),
			)
		} else {
			// FIXME: change fn decl signature to always have a block expr
			block := decl.Body.(*ast.BlockExpr)
			sp := block.Exprs[len(block.Exprs)-1].Span()
			b.error(sp, "expected this function to return `%s`, but got `%s`", ty, body.Type())
		}
		return &bir.ErrItem{}
	}

	return &bir.FnDecl{Ident: ident, In: params, Out: ty, Body: body}
}

func (b *binder) bindParams(aParams []*ast.Param) []*bir.Param {
	var params []*bir.Param
	seen := make(map[string]bool, len(aParams))
	for _, aParam := range aParams {
		if ok := seen[aParam.Ident.Name]; ok {
			b.error(aParam.Ident.Sp, "parameter `%s` already exists", aParam.Ident.Name)
		} else {
			seen[aParam.Ident.Name] = true
			param := b.bindParam(aParam)
			if param != nil {
				params = append(params, param)
				b.scope.Insert(param.Ident.Name, param)
			}
		}
	}
	return params
}

func (b *binder) bindParam(aParam *ast.Param) *bir.Param {
	ty := lookupTy(aParam.Ty.Name)
	if ty == bir.TErr {
		b.error(aParam.Ty.Sp, "cannot find type `%s` in this scope", aParam.Ty.Name)
		return nil
	}
	ident := &bir.Ident{Name: aParam.Ident.Name}
	return &bir.Param{Ident: ident, Ty: ty}
}

func (b *binder) bindExpr(expr ast.Expr) bir.Expr {
	switch expr := expr.(type) {
	case *ast.Ident:
		if d, ok := b.scope.Get(expr.Name); ok {
			return &bir.Ident{Name: expr.Name, Ty: d.Type()}
		}
		b.error(expr.Sp, "could not find anything named `%s` in this scope", expr.Name)
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
	var ty bir.Ty
	init := b.bindExpr(expr.Init)
	if expr.Ty != nil {
		ty = lookupTy(expr.Ty.Name)
		if ty == bir.TErr {
			b.error(expr.Ty.Sp, "cannot find type `%s` in this scope", expr.Ty.Name)
			return &bir.ErrExpr{}
		}

		if ty != init.Type() {
			b.error(expr.Init.Span(), "expected `%s`, but got `%s`", ty, init.Type())
			return &bir.ErrExpr{}
		}
	} else {
		ty = init.Type()
	}

	ident := &bir.Ident{Name: expr.Ident.Name, Ty: ty}
	le := &bir.LetExpr{Ident: ident, Ty: ty, Init: init}
	b.scope.Insert(expr.Ident.Name, le)
	return le
}

func (b *binder) bindAssignExpr(expr *ast.AssignExpr) bir.Expr {
	x := b.bindExpr(expr.X)
	y := b.bindExpr(expr.Y)
	switch x := x.(type) {
	case *bir.Ident:
		if d, ok := b.scope.Get(x.Name); ok {
			switch d := d.(type) {
			case *bir.LetExpr:
				if d.Type() != y.Type() {
					b.error(expr.Y.Span(), "expected `%s`, but got `%s`", d.Type(), y.Type())
					return &bir.ErrExpr{}
				}
			default:
				b.error(expr.X.Span(), "can only assign to identifiers for now")
				return &bir.ErrExpr{}
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
	prev := b.scope
	b.scope = WithOuter(b.scope)
	for _, e := range expr.Exprs {
		exprs = append(exprs, b.bindExpr(e))
	}
	b.scope = prev
	return &bir.BlockExpr{Exprs: exprs}
}

func (b *binder) error(span span.Span, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	diagnostic.NewBuilder(msg, span).WithLabel("here").Emit(b.diags)
}

func lookupTy(name string) bir.Ty {
	switch name {
	case "int":
		return bir.TInt
	case "bool":
		return bir.TBool
	default:
		return bir.TErr
	}
}
