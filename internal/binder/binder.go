package binder

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/ir"
	"github.com/aadamandersson/lue/internal/ir/ast"
	"github.com/aadamandersson/lue/internal/ir/bir"
	"github.com/aadamandersson/lue/internal/session"
	"github.com/aadamandersson/lue/internal/span"
)

func Bind(items []ast.Item, sess *session.Session) (map[string]*bir.Class, map[string]*bir.Fn) {
	scope := bindGlobalScope(items, sess)
	fns := scope.Functions()
	classes := scope.Classes()
	b := new(sess, scope)
	for _, fn := range fns {
		b.bindFnDecl(fn, sess, scope)
	}

	return classes, fns
}

func bindGlobalScope(aItems []ast.Item, sess *session.Session) *Scope {
	scope := NewScope()
	b := new(sess, scope)
	for _, aItem := range aItems {
		switch aItem := aItem.(type) {
		case *ast.FnDecl:
			fn := &bir.Fn{Decl: aItem, Out: b.lookupTy(aItem.Out)}
			if _, exists := scope.Insert(aItem.Ident.Name, fn); exists {
				b.error(aItem.Ident.Sp, "function `%s` already exists", aItem.Ident.Name)
			}
		case *ast.ClassDecl:
			class := &bir.Class{Decl: aItem, Fields: b.bindFields(aItem.Fields)}
			if _, exists := scope.Insert(aItem.Ident.Name, class); exists {
				b.error(aItem.Ident.Sp, "class `%s` already exists", aItem.Ident.Name)
			}
		case *ast.ErrItem:
			continue
		default:
			panic("unreachable")
		}
	}

	for _, intr := range ir.Intrinsics() {
		scope.Insert(intr.String(), (bir.Intrinsic)(intr))
	}

	return scope
}

func (b *binder) bindFnDecl(fn *bir.Fn, sess *session.Session, scope *Scope) {
	prev := b.scope
	b.fn = fn
	b.scope = WithOuter(b.scope)
	fn.In = b.bindParams(fn.Decl.In)
	var ty *bir.Ty
	if fn.Decl.Out == nil {
		ty = bir.BasicTys[bir.TyUnit]
	} else {
		ty = b.lookupTy(fn.Decl.Out)
		if ty.IsErr() {
			b.error(fn.Decl.Out.Sp, "cannot find type `%s` in this scope", fn.Decl.Out)
		}
	}

	body := b.bindExpr(fn.Decl.Body)
	if blk, ok := body.(*bir.BlockExpr); ok {
		if !ty.IsUnit() && len(blk.Exprs) == 0 {
			b.error(
				fn.Decl.Out.Sp,
				"expected this function to return `%s`, but the body is empty",
				ty,
			)
			body = &bir.ErrExpr{}
		}
	} else if ty != body.Type() {
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
		body = &bir.ErrExpr{}
	}
	fn.Body = body
	b.scope = prev
}

type binder struct {
	sess      *session.Session
	loopLevel int
	fn        *bir.Fn
	scope     *Scope
}

func new(sess *session.Session, scope *Scope) binder {
	return binder{
		sess:  sess,
		scope: scope,
	}
}

func (b *binder) bindParams(aParams []*ast.VarDecl) []*bir.VarDecl {
	var params []*bir.VarDecl
	seen := make(map[string]bool, len(aParams))
	for _, aParam := range aParams {
		if ok := seen[aParam.Ident.Name]; ok {
			b.error(aParam.Ident.Sp, "parameter `%s` already exists", aParam.Ident.Name)
		} else {
			seen[aParam.Ident.Name] = true
			param := b.bindVarDecl(aParam)
			if param != nil {
				params = append(params, param)
				b.scope.Insert(aParam.Ident.Name, param)
			}
		}
	}
	return params
}

func (b *binder) bindFields(aFields []*ast.VarDecl) []*bir.VarDecl {
	var fields []*bir.VarDecl
	seen := make(map[string]bool, len(aFields))
	for _, aField := range aFields {
		if ok := seen[aField.Ident.Name]; ok {
			b.error(aField.Ident.Sp, "field `%s` already exists", aField.Ident.Name)
		} else {
			seen[aField.Ident.Name] = true
			field := b.bindVarDecl(aField)
			fields = append(fields, field)
		}
	}
	return fields
}

func (b *binder) bindVarDecl(decl *ast.VarDecl) *bir.VarDecl {
	ty := b.lookupTy(decl.Ty)
	if ty.IsErr() {
		b.error(decl.Ty.Sp, "cannot find type `%s` in this scope", decl.Ty)
		return nil
	}
	return &bir.VarDecl{Ident: (*ir.Ident)(decl.Ident), Ty: ty}
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
	case *ast.StringLiteral:
		return &bir.StringLiteral{V: expr.V}
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
	case *ast.ClassExpr:
		return b.bindClassExpr(expr)
	case *ast.FieldExpr:
		return b.bindFieldExpr(expr)
	case *ast.ArrayExpr:
		return b.bindArrayExpr(expr)
	case *ast.IndexExpr:
		return b.bindIndexExpr(expr)
	case *ast.ForExpr:
		return b.bindForExpr(expr)
	case *ast.BreakExpr:
		return b.bindBreakExpr(expr)
	case *ast.ReturnExpr:
		return b.bindReturnExpr(expr)
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
	op, ok := bir.BindBinOp(expr.Op.Kind, x.Type().Kind, y.Type().Kind)

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
	var ty *bir.Ty
	init := b.bindExpr(expr.Init)
	ty = b.lookupTy(expr.Decl.Ty)
	if !ty.IsInfer() {
		if ty.IsErr() {
			b.error(expr.Decl.Ty.Sp, "cannot find type `%s` in this scope", expr.Decl.Ty)
			return &bir.ErrExpr{}
		}

		if !ty.Equal(init.Type()) {
			b.error(expr.Init.Span(), "expected `%s`, but got `%s`", ty, init.Type())
			return &bir.ErrExpr{}
		}
	} else {
		ty = init.Type()
	}

	decl := &bir.VarDecl{Ident: (*ir.Ident)(expr.Decl.Ident), Ty: ty}
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

		return &bir.AssignExpr{X: x, Y: y}
	default:
		b.error(expr.X.Span(), "can only assign to identifiers for now")
		return &bir.ErrExpr{}
	}
}

func (b *binder) bindIfExpr(expr *ast.IfExpr) bir.Expr {
	cond := b.bindExpr(expr.Cond)
	if !cond.Type().IsBool() {
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
		if len(fn.Decl.In) != len(expr.Args) {
			b.error(
				expr.Fn.Span(),
				"this functions expects %d argument(s), but %d argument(s) were supplied",
				len(fn.In),
				len(expr.Args),
			)
			return &bir.ErrExpr{}
		}
	case bir.Intrinsic:
		switch (ir.Intrinsic)(fn) {
		case ir.IntrPrintln:
			if len(expr.Args) != 1 {
				b.error(
					expr.Fn.Span(),
					"`println` expects 1 argument, but %d argument(s) were supplied",
					len(expr.Args),
				)
				return &bir.ErrExpr{}
			}
		default:
			panic("unreachable")
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

func (b *binder) bindClassExpr(expr *ast.ClassExpr) bir.Expr {
	e, ok := b.scope.Get(expr.Ident.Name)
	if !ok {
		b.error(expr.Ident.Sp, "could not find a class named `%s` in this scope", expr.Ident.Name)
		return &bir.ErrExpr{}
	}

	c, ok := e.(*bir.Class)
	if !ok {
		b.error(expr.Ident.Sp, "`%s` is not a class", expr.Ident.Name)
		return &bir.ErrExpr{}
	}

	exprFields := b.bindExprFields(expr.Fields)
	hasError := false
	for _, field := range c.Fields {
		found := false
		for _, exprField := range exprFields {
			if field.Ident.Name != exprField.Ident.Name {
				continue
			}

			if !field.Ty.Equal(exprField.Expr.Type()) {
				b.error(
					field.Ident.Sp,
					"expected `%s`, but got `%s`",
					field.Ty,
					exprField.Expr.Type(),
				)
				hasError = true
			}
			found = true
		}

		if !found {
			b.error(field.Ident.Sp, "missing field `%s` in initializer", field.Ident.Name)
			hasError = true
		}
	}

	if hasError {
		return &bir.ErrExpr{}
	}

	return &bir.ClassExpr{Ident: (*ir.Ident)(expr.Ident), Fields: exprFields}
}

func (b *binder) bindExprFields(aFields []*ast.ExprField) []*bir.ExprField {
	var fields []*bir.ExprField
	for _, f := range aFields {
		fields = append(fields, b.bindExprField(f))
	}
	return fields
}

func (b *binder) bindExprField(aField *ast.ExprField) *bir.ExprField {
	expr := b.bindExpr(aField.Expr)
	return &bir.ExprField{Ident: (*ir.Ident)(aField.Ident), Expr: expr}
}

func (b *binder) bindFieldExpr(aExpr *ast.FieldExpr) bir.Expr {
	expr := b.bindExpr(aExpr.Expr)
	decl := expr.(*bir.VarDecl)
	var ty *bir.Ty
	switch decl.Ty.Kind {
	case bir.TyClass:
		classExpr, _ := b.scope.Get(decl.Ty.Class.Name)
		class := classExpr.(*bir.Class)
		found := false
		for _, f := range class.Fields {
			if f.Ident.Name == aExpr.Ident.Name {
				ty = f.Ty
				found = true
			}
		}
		if !found {
			b.error(
				aExpr.Ident.Sp,
				"could not find field `%s` in class `%s`",
				aExpr.Ident.Name,
				class.Decl.Ident.Name,
			)
		}
	default:
		b.error(aExpr.Expr.Span(), "expected a class, but got `%s`", expr.Type())
		return &bir.ErrExpr{}
	}
	return &bir.FieldExpr{Ident: (*ir.Ident)(aExpr.Ident), Expr: expr, Ty: ty}
}

func (b *binder) bindArrayExpr(expr *ast.ArrayExpr) bir.Expr {
	if len(expr.Exprs) == 0 {
		return &bir.ArrayExpr{Exprs: []bir.Expr{}}
	}

	var exprs []bir.Expr
	for _, expr := range expr.Exprs {
		exprs = append(exprs, b.bindExpr(expr))
	}

	expectedTy := exprs[0].Type()
	for i := 1; i < len(exprs); i++ {
		actualTy := exprs[i].Type()
		if expectedTy != actualTy {
			b.error(expr.Exprs[i].Span(), "expected `%s`, but got `%s`", expectedTy, actualTy)
			return &bir.ErrExpr{}
		}
	}

	return &bir.ArrayExpr{Exprs: exprs}
}

func (b *binder) bindIndexExpr(expr *ast.IndexExpr) bir.Expr {
	arr := b.bindExpr(expr.Arr)
	if !arr.Type().IsArray() {
		b.error(expr.Arr.Span(), "expected an array, but got `%s`", arr.Type())
		return &bir.ErrExpr{}
	}

	i := b.bindExpr(expr.I)
	if !i.Type().IsInt() {
		b.error(expr.I.Span(), "expected an integer, but got `%s`", i.Type())
		return &bir.ErrExpr{}
	}

	return &bir.IndexExpr{Arr: arr, I: i}
}

func (b *binder) bindForExpr(expr *ast.ForExpr) bir.Expr {
	var body bir.Expr = &bir.ErrExpr{}
	b.loopLevel += 1
	body = b.bindExpr(expr.Body)
	return &bir.ForExpr{Body: body}
}

func (b *binder) bindBreakExpr(expr *ast.BreakExpr) bir.Expr {
	if b.loopLevel == 0 {
		b.error(expr.Sp, "cannot `break` outside a `for` loop")
		return &bir.ErrExpr{}
	}

	var x bir.Expr
	if expr.X != nil {
		x = b.bindExpr(expr.X)
	}
	b.loopLevel -= 1
	return &bir.BreakExpr{X: x}
}

func (b *binder) bindReturnExpr(expr *ast.ReturnExpr) bir.Expr {
	var x bir.Expr
	if expr.X != nil {
		x = b.bindExpr(expr.X)
	}

	if b.fn.Out.IsUnit() && x != nil {
		b.error(
			expr.X.Span(),
			"expected this function to return `%s`, but got `%s`",
			b.fn.Out,
			x.Type(),
		)
		return &bir.ErrExpr{}
	}

	return &bir.ReturnExpr{X: x}
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
	diagnostic.NewBuilder(msg, span).WithLabel("here").Emit(b.sess.Diags)
}

func (b *binder) lookupTy(ty *ast.Ty) *bir.Ty {
	switch ty.Kind {
	case ast.TyInfer:
		return bir.BasicTys[bir.TyInfer]
	case ast.TyArray:
		return bir.NewArray(lookUpBasicTy(ty.Ident))
	case ast.TyIdent:
		if _, ok := b.scope.Get(ty.Ident.Name); ok {
			return bir.NewClass((*ir.Ident)(ty.Ident))
		}
		return lookUpBasicTy(ty.Ident)
	case ast.TyUnit:
		return bir.BasicTys[bir.TyUnit]
	default:
		return bir.BasicTys[bir.TyErr]
	}
}

func lookUpBasicTy(ty *ast.Ident) *bir.Ty {
	switch ty.Name {
	case "int":
		return bir.BasicTys[bir.TyInt]
	case "bool":
		return bir.BasicTys[bir.TyBool]
	case "string":
		return bir.BasicTys[bir.TyString]
	default:
		return bir.BasicTys[bir.TyErr]
	}
}
