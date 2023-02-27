package bir

import (
	"strconv"

	"github.com/aadamandersson/lue/internal/ast"
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
		Ident *Ident
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

	// Placeholder when we have some parse or bind error.
	ErrExpr struct{}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*Fn) exprNode()             {}
func (*VarDecl) exprNode()        {}
func (*IntegerLiteral) exprNode() {}
func (*BooleanLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}
func (*LetExpr) exprNode()        {}
func (*AssignExpr) exprNode()     {}
func (*IfExpr) exprNode()         {}
func (*BlockExpr) exprNode()      {}
func (*CallExpr) exprNode()       {}
func (*ErrExpr) exprNode()        {}

func (e *Fn) Type() Ty             { return e.Out }
func (e *VarDecl) Type() Ty        { return e.Ty }
func (e *IntegerLiteral) Type() Ty { return TInt }
func (e *BooleanLiteral) Type() Ty { return TBool }
func (e *BinaryExpr) Type() Ty     { return e.Op.Ty }
func (e *LetExpr) Type() Ty        { return TUnit }
func (e *AssignExpr) Type() Ty     { return TUnit }
func (e *IfExpr) Type() Ty         { return e.Then.Type() }
func (e *BlockExpr) Type() Ty {
	if len(e.Exprs) == 0 {
		return TUnit
	}
	return e.Exprs[len(e.Exprs)-1].Type()
}
func (e *CallExpr) Type() Ty { return e.Fn.Type() }
func (e *ErrExpr) Type() Ty  { return TErr }

type Ident struct {
	Name string
}

type Ty int

const (
	TErr Ty = iota
	TInt
	TBool
	TUnit
)

var tys = [...]string{
	TErr:  "?",
	TInt:  "int",
	TBool: "bool",
	TUnit: "()",
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
	{ast.Add, TInt, TInt, BinOp{Kind: Add, Ty: TInt}},
	{ast.Sub, TInt, TInt, BinOp{Kind: Sub, Ty: TInt}},
	{ast.Mul, TInt, TInt, BinOp{Kind: Mul, Ty: TInt}},
	{ast.Div, TInt, TInt, BinOp{Kind: Div, Ty: TInt}},

	{ast.Gt, TInt, TInt, BinOp{Kind: Gt, Ty: TBool}},
	{ast.Lt, TInt, TInt, BinOp{Kind: Lt, Ty: TBool}},
	{ast.Ge, TInt, TInt, BinOp{Kind: Ge, Ty: TBool}},
	{ast.Le, TInt, TInt, BinOp{Kind: Le, Ty: TBool}},

	{ast.Eq, TInt, TInt, BinOp{Kind: Eq, Ty: TBool}},
	{ast.Eq, TBool, TBool, BinOp{Kind: Eq, Ty: TBool}},

	{ast.Ne, TInt, TInt, BinOp{Kind: Ne, Ty: TBool}},
	{ast.Ne, TBool, TBool, BinOp{Kind: Ne, Ty: TBool}},
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
