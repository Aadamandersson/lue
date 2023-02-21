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
	// An integer literal.
	// E.g., `123`
	IntegerLiteral struct {
		V int
	}

	// A binary expression.
	// E.g., `x + y`
	BinaryExpr struct {
		X  Expr
		Op BinOp
		Y  Expr
	}

	// Placeholder when we have some bind error.
	ErrExpr struct{}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*IntegerLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}
func (*ErrExpr) exprNode()        {}

func (il *IntegerLiteral) Type() Ty {
	return TInt
}

func (be *BinaryExpr) Type() Ty {
	return be.Op.Ty
}

func (ee *ErrExpr) Type() Ty {
	return TErr
}

type Ty int

const (
	TErr Ty = iota
	TInt
)

var tys = [...]string{
	TErr: "?",
	TInt: "int",
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
)
