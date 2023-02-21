package ast

import "github.com/aadamandersson/lue/internal/token"

type Expr interface {
	exprNode()
}

// Expressions
type (
	IntegerLiteral struct {
		V string
	}

	BinaryExpr struct {
		X  Expr
		Op BinOpKind
		Y  Expr
	}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*IntegerLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}

type BinOpKind int

const (
	Add BinOpKind = iota + 1 // `+` (addition)
	Sub                      // `-` (subtraction)
	Mul                      // `*` (multiplication)
	Div                      // `/` (division)
)

// Prec returns the operator precedence for binary operator op.
func (op BinOpKind) Prec() int {
	switch op {
	case Mul, Div:
		return 2
	case Add, Sub:
		return 1
	default:
		return 0
	}
}

// BinOpFromToken returns the binOp for token t and a boolean true, if its a valid binary operator.
// Otherwise, returns the zero value of BinOpKind and a boolean false.
func BinOpFromToken(t token.Token) (binOp BinOpKind, isBinOp bool) {
	isBinOp = true
	switch t.Kind {
	case token.Plus:
		binOp = Add
	case token.Minus:
		binOp = Sub
	case token.Star:
		binOp = Mul
	case token.Slash:
		binOp = Div
	default:
		isBinOp = false
	}
	return
}
