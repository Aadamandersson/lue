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

	// A binary expression.
	// E.g., `x + y`
	BinaryExpr struct {
		X  Expr
		Op BinOp
		Y  Expr
	}

	// An assignment expression.
	// `expr = init`
	AssignExpr struct {
		Ident *Ident // FIXME: change to expr when we support let bindings.
		Init  Expr
	}

	// Placeholder when we have some bind error.
	ErrExpr struct{}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*Ident) exprNode()          {}
func (*IntegerLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}
func (*AssignExpr) exprNode()     {}
func (*ErrExpr) exprNode()        {}

func (i *Ident) Type() Ty {
	return i.Ty
}

func (il *IntegerLiteral) Type() Ty {
	return TInt
}

func (be *BinaryExpr) Type() Ty {
	return be.Op.Ty
}

func (ae *AssignExpr) Type() Ty {
	return ae.Init.Type()
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
