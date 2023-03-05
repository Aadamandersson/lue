package bir

import (
	"strconv"

	"github.com/aadamandersson/lue/internal/ir"
	"github.com/aadamandersson/lue/internal/ir/ast"
)

type (
	Expr interface {
		Type() Ty
		exprNode()
	}
)

// Expressions
type (
	// A reference to a function.
	Fn struct {
		Decl *ast.FnDecl
		In   []*VarDecl
		Out  Ty
		Body Expr
	}

	// A variable declaration.
	// `ident: ty`
	VarDecl struct {
		Ident *ir.Ident
		Ty    Ty
	}

	// An integer literal.
	// E.g., `123`
	IntegerLiteral struct {
		V int
	}

	// A boolean literal.
	// `true` or `false`
	BooleanLiteral struct {
		V bool
	}

	// A string literal.
	// E.g., `"foo"`
	StringLiteral struct {
		V string
	}

	// A binary expression.
	// E.g., `x + y`
	BinaryExpr struct {
		X  Expr
		Op BinOp
		Y  Expr
	}

	// A let binding.
	// `let ident [: ty] = init`
	LetExpr struct {
		Decl *VarDecl
		Init Expr
	}

	// An assignment expression.
	// `x = y`
	AssignExpr struct {
		X Expr
		Y Expr
	}

	// An if expression.
	// `if cond { exprs } [else [if cond] { exprs }]`
	IfExpr struct {
		Cond Expr
		Then Expr
		Else Expr // Optional, may be nil.
	}

	// A block expression.
	// `{ exprs }`
	BlockExpr struct {
		Exprs []Expr
	}

	// A function call.
	// `fn(args)`
	CallExpr struct {
		Fn   Expr
		Args []Expr
	}

	// An array expression.
	// `[1, 2, 3]`
	ArrayExpr struct {
		Exprs []Expr
	}

	// An array indexing expression.
	// `arr[i]`
	IndexExpr struct {
		Arr Expr
		I   Expr
	}

	// A for loop.
	// `for { exprs }`
	ForExpr struct {
		Body Expr
	}

	// A break expression.
	// `break [expr]`
	BreakExpr struct {
		X Expr // Optional, may be nil.
	}

	// A return expression.
	// `return [expr]`
	ReturnExpr struct {
		X Expr // Optional, may be nil.
	}

	// An intrinsic.
	// E.g., `println`
	Intrinsic ir.Intrinsic

	// Placeholder when we have some parse or bind error.
	ErrExpr struct{}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*Fn) exprNode()             {}
func (*VarDecl) exprNode()        {}
func (*IntegerLiteral) exprNode() {}
func (*BooleanLiteral) exprNode() {}
func (*StringLiteral) exprNode()  {}
func (*BinaryExpr) exprNode()     {}
func (*LetExpr) exprNode()        {}
func (*AssignExpr) exprNode()     {}
func (*IfExpr) exprNode()         {}
func (*BlockExpr) exprNode()      {}
func (*CallExpr) exprNode()       {}
func (*ArrayExpr) exprNode()      {}
func (*IndexExpr) exprNode()      {}
func (*ForExpr) exprNode()        {}
func (*BreakExpr) exprNode()      {}
func (*ReturnExpr) exprNode()     {}
func (Intrinsic) exprNode()       {}
func (*ErrExpr) exprNode()        {}

func (e *Fn) Type() Ty             { return e.Out }
func (e *VarDecl) Type() Ty        { return e.Ty }
func (e *IntegerLiteral) Type() Ty { return TyInt }
func (e *BooleanLiteral) Type() Ty { return TyBool }
func (e *StringLiteral) Type() Ty  { return TyString }
func (e *BinaryExpr) Type() Ty     { return e.Op.Ty }
func (e *LetExpr) Type() Ty        { return TyUnit }
func (e *AssignExpr) Type() Ty     { return TyUnit }
func (e *IfExpr) Type() Ty         { return e.Then.Type() }
func (e *BlockExpr) Type() Ty {
	if len(e.Exprs) == 0 {
		return TyUnit
	}
	return e.Exprs[len(e.Exprs)-1].Type()
}
func (e *CallExpr) Type() Ty  { return e.Fn.Type() }
func (e *ArrayExpr) Type() Ty { return TyArray }
func (e *IndexExpr) Type() Ty { return e.Arr.Type() } // FIXME: this should be the element type
func (e *ForExpr) Type() Ty   { return e.Body.Type() }
func (e *BreakExpr) Type() Ty {
	if e.X == nil {
		return TyUnit
	}
	return e.X.Type()
}
func (e *ReturnExpr) Type() Ty {
	if e.X == nil {
		return TyUnit
	}
	return e.X.Type()
}
func (e Intrinsic) Type() Ty { return TyUnit }
func (e *ErrExpr) Type() Ty  { return TyErr }

func (ae *ArrayExpr) ElemTy() Ty {
	if len(ae.Exprs) == 0 {
		return TyErr
	}
	return ae.Exprs[0].Type()
}

type Ty int

const (
	TyErr Ty = iota
	TyInfer
	TyInt
	TyBool
	TyString
	TyArray
	TyUnit
)

var tys = [...]string{
	TyErr:    "?",
	TyInfer:  "?",
	TyInt:    "int",
	TyBool:   "bool",
	TyString: "string",
	TyArray:  "[]", // TODO: we want to show the type of the elements as well
	TyUnit:   "()",
}

func (t Ty) String() string {
	if t < 0 || t >= Ty(len(tys)) {
		return "Ty(" + strconv.FormatInt(int64(t), 10) + ")"
	}
	return tys[t]
}

type BinOp struct {
	Kind BinOpKind
	Ty   Ty
}

var binOps = [...]struct {
	in  ast.BinOpKind
	xTy Ty
	yTy Ty
	out BinOp
}{
	{ast.Add, TyInt, TyInt, BinOp{Kind: Add, Ty: TyInt}},
	{ast.Sub, TyInt, TyInt, BinOp{Kind: Sub, Ty: TyInt}},
	{ast.Mul, TyInt, TyInt, BinOp{Kind: Mul, Ty: TyInt}},
	{ast.Div, TyInt, TyInt, BinOp{Kind: Div, Ty: TyInt}},

	{ast.Gt, TyInt, TyInt, BinOp{Kind: Gt, Ty: TyBool}},
	{ast.Lt, TyInt, TyInt, BinOp{Kind: Lt, Ty: TyBool}},
	{ast.Ge, TyInt, TyInt, BinOp{Kind: Ge, Ty: TyBool}},
	{ast.Le, TyInt, TyInt, BinOp{Kind: Le, Ty: TyBool}},

	{ast.Eq, TyInt, TyInt, BinOp{Kind: Eq, Ty: TyBool}},
	{ast.Eq, TyBool, TyBool, BinOp{Kind: Eq, Ty: TyBool}},
	{ast.Eq, TyString, TyString, BinOp{Kind: Eq, Ty: TyBool}},

	{ast.Ne, TyInt, TyInt, BinOp{Kind: Ne, Ty: TyBool}},
	{ast.Ne, TyBool, TyBool, BinOp{Kind: Ne, Ty: TyBool}},
	{ast.Ne, TyString, TyString, BinOp{Kind: Ne, Ty: TyBool}},
}

func BindBinOp(astOp ast.BinOpKind, xTy, yTy Ty) (BinOp, bool) {
	for _, op := range binOps {
		if op.in == astOp && op.xTy == xTy && op.yTy == yTy {
			return op.out, true
		}
	}
	return *new(BinOp), false
}

type BinOpKind int

const (
	Add BinOpKind = iota // `+` (addition)
	Sub                  // `-` (subtraction)
	Mul                  // `*` (multiplication)
	Div                  // `/` (division)
	Gt                   // `>` (greater than)
	Lt                   // `<` (less than)
	Ge                   // `>=` (greater than or equal)
	Le                   // `<=` (less than or equal)
	Eq                   // `==` (equality)
	Ne                   // `!=` (not equal)
)
