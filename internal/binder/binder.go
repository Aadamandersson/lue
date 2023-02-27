package binder

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/ast"
	"github.com/aadamandersson/lue/internal/bir"
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/span"
)

func Bind(items []ast.Item, diags *diagnostic.Bag) map[string]*bir.Fn {
	scope := bindGlobalScope(items, diags)
	fns := scope.Functions()

	b := new(diags, scope)

	for _, fn := range fns {
		b.bindFnDecl(fn, diags, scope)
	}

	return fns
}

func bindGlobalScope(aItems []ast.Item, diags *diagnostic.Bag) *Scope {
	scope := NewScope()
	b := new(diags, scope)
	for _, aItem := range aItems {
		switch aItem := aItem.(type) {
		case *ast.FnDecl:
			fn := &bir.Fn{Decl: aItem, Out: lookupTy(aItem.Out)}
			if _, exists := scope.Insert(aItem.Ident.Name, fn); exists {
				b.error(aItem.Ident.Sp, "function `%s` already exists", aItem.Ident.Name)
			}
		case *ast.ErrItem:
			continue
		default:
			panic("unreachable")
		}
	}

	return scope
}

func (b *binder) bindFnDecl(fn *bir.Fn, diags *diagnostic.Bag, scope *Scope) {
	prev := b.scope
	b.scope = WithOuter(b.scope)
	fn.In = b.bindParams(fn.Decl.In)
	var ty bir.Ty
	if fn.Decl.Out == nil {
		ty = bir.TUnit
	} else {
		ty = lookupTy(fn.Decl.Out)
		if ty == bir.TErr {
			b.error(fn.Decl.Out.Sp, "cannot find type `%s` in this scope", fn.Decl.Out.Name)
		}
	}
	body := b.bindExpr(fn.Decl.Body)
	if ty != body.Type() {
		if fn.Decl.Out != nil {
			b.error(
				fn.Decl.Out.Sp,
				"expected this function to return `%s`, but got `%s`",
				ty,
				body.Type(),
			)
		} else {
			// FIXME: change fn decl signature to always have a block expr
			block := fn.Decl.Body.(*ast.BlockExpr)
			sp := block.Exprs[len(block.Exprs)-1].Span()
			b.error(sp, "expected this function to return `%s`, but got `%s`", ty, body.Type())
		}
	}
	fn.Body = body
	b.scope = prev
}

type binder struct {
	diags *diagnostic.Bag
	scope *Scope
}

func new(diags *diagnostic.Bag, scope *Scope) binder {
	return binder{diags: diags, scope: scope}
}

func (b *binder) bindParams(aParams []*ast.VarDecl) []*bir.VarDecl {
	var params []*bir.VarDecl
	seen := make(map[string]bool, len(aParams))
	for _, aParam := range aParams {
		if ok := seen[aParam.Ident.Name]; ok {
			b.error(aParam.Ident.Sp, "parameter `%s` already exists", aParam.Ident.Name)
		} else {
			seen[aParam.Ident.Name] = true
			param := b.bindParam(aParam)
			if param != nil {
				params = append(params, param)
				b.scope.Insert(aParam.Ident.Name, param)
			}
		}
	}
	return params
}

func (b *binder) bindParam(aParam *ast.VarDecl) *bir.VarDecl {
	ty := lookupTy(aParam.Ty)
	if ty == bir.TErr {
		b.error(aParam.Ty.Sp, "cannot find type `%s` in this scope", aParam.Ty.Name)
		return nil
	}
	ident := &bir.Ident{Name: aParam.Ident.Name}
	return &bir.VarDecl{Ident: ident, Ty: ty}
}

func (b *binder) bindExpr(expr ast.Expr) bir.Expr {
	switch expr := expr.(type) {
	case *ast.Ident:
		if d, ok := b.scope.Get(expr.Name); ok {
			return d
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
	case *ast.CallExpr:
		return b.bindCallExpr(expr)
	case *ast.ErrExpr:
		return &bir.ErrExpr{}
	}
	panic("unreachable")
}

func (b *binder) bindBinaryExpr(expr *ast.BinaryExpr) bir.Expr {
	x := b.bindExpr(expr.X)
	y := b.bindExpr(expr.Y)
	if isErr(x) || isErr(y) {
		return &bir.ErrExpr{}
	}
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
	if expr.Decl.Ty != nil {
		ty = lookupTy(expr.Decl.Ty)
		if ty == bir.TErr {
			b.error(expr.Decl.Ty.Sp, "cannot find type `%s` in this scope", expr.Decl.Ty.Name)
			return &bir.ErrExpr{}
		}

		if ty != init.Type() {
			b.error(expr.Init.Span(), "expected `%s`, but got `%s`", ty, init.Type())
			return &bir.ErrExpr{}
		}
	} else {
		ty = init.Type()
	}

	ident := &bir.Ident{Name: expr.Decl.Ident.Name}
	decl := &bir.VarDecl{Ident: ident, Ty: ty}
	le := &bir.LetExpr{Decl: decl, Init: init}
	b.scope.Insert(expr.Decl.Ident.Name, decl)
	return le
}

func (b *binder) bindAssignExpr(expr *ast.AssignExpr) bir.Expr {
	x := b.bindExpr(expr.X)
	y := b.bindExpr(expr.Y)
	switch x := x.(type) {
	case *bir.VarDecl:
		if d, ok := b.scope.Get(x.Ident.Name); ok {
			switch d := d.(type) {
			case *bir.VarDecl:
				if d.Type() != y.Type() {
					b.error(expr.Y.Span(), "expected `%s`, but got `%s`", d.Type(), y.Type())
					return &bir.ErrExpr{}
				}
			default:
				b.error(expr.X.Span(), "can only assign to identifiers for now")
				return &bir.ErrExpr{}
			}
		}
		ident := &bir.Ident{Name: x.Ident.Name}
		return &bir.AssignExpr{X: &bir.VarDecl{Ident: ident, Ty: y.Type()}, Y: y}
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

func (b *binder) bindCallExpr(expr *ast.CallExpr) bir.Expr {
	fn := b.bindExpr(expr.Fn)
	switch fn := fn.(type) {
	case *bir.Fn:
		if len(fn.In) != len(expr.Args) {
			b.error(expr.Fn.Span(), "this functions expects %d argument(s) but %d arguments were supplied", len(fn.In), len(expr.Args))
			return &bir.ErrExpr{}
		}
	default:
		b.error(expr.Fn.Span(), "expected a function")
		return &bir.ErrExpr{}
	}
	var args []bir.Expr
	if expr.Args != nil {
		for _, arg := range expr.Args {
			args = append(args, b.bindExpr(arg))
		}
	}
	return &bir.CallExpr{Fn: fn, Args: args}
}

func isErr(expr bir.Expr) bool {
	switch expr.(type) {
	case *bir.ErrExpr:
		return true
	default:
		return false
	}
}

func (b *binder) error(span span.Span, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	diagnostic.NewBuilder(msg, span).WithLabel("here").Emit(b.diags)
}

func lookupTy(out *ast.Ident) bir.Ty {
	if out == nil {
		return bir.TUnit
	}

	switch out.Name {
	case "int":
		return bir.TInt
	case "bool":
		return bir.TBool
	default:
		return bir.TErr
	}
}
