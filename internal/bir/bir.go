package bir

import (
	"strconv"

	"github.com/aadamandersson/lue/internal/ast"
)

type Expr interface {
	Type() Ty
	exprNode()
}

// Expressions
type (
	// An identifier.
	// E.g., `foo`
	Ident struct {
		Name string
		Ty   Ty
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
	// `let ident = init`
	LetExpr struct {
		Ident *Ident
		Init  Expr
	}

	// An assignment expression.
	// `expr = init`
	AssignExpr struct {
		Ident *Ident // FIXME: change to expr when we support let bindings.
		Init  Expr
	}

	// A block expression.
	// `{ exprs }`
	BlockExpr struct {
		Exprs []Expr
	}

	// Placeholder when we have some bind error.
	ErrExpr struct{}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*Ident) exprNode()          {}
func (*IntegerLiteral) exprNode() {}
func (*BooleanLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}
func (*LetExpr) exprNode()        {}
func (*AssignExpr) exprNode()     {}
func (*BlockExpr) exprNode()      {}
func (*ErrExpr) exprNode()        {}

func (i *Ident) Type() Ty {
	return i.Ty
}

func (il *IntegerLiteral) Type() Ty {
	return TInt
}

func (il *BooleanLiteral) Type() Ty {
	return TBool
}

func (be *BinaryExpr) Type() Ty {
	return be.Op.Ty
}

func (le *LetExpr) Type() Ty {
	return le.Init.Type()
}

func (ae *AssignExpr) Type() Ty {
	return ae.Init.Type()
}

func (be *BlockExpr) Type() Ty {
	if len(be.Exprs) == 0 {
		return TUnit
	}
	return be.Exprs[len(be.Exprs)-1].Type()
}

func (ee *ErrExpr) Type() Ty {
	return TErr
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
