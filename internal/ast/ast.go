package ast

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/token"
)

type Expr interface {
	exprNode()
}

// Expressions
type (
	// An identifier.
	// E.g., `foo`
	Ident struct {
		Name string
	}

	// An integer literal.
	// E.g., `123`
	IntegerLiteral struct {
		V string
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
		Op BinOpKind
		Y  Expr
	}

	// An assignment expression.
	// `expr = init`
	AssignExpr struct {
		Ident *Ident // FIXME: change to expr when we support let bindings.
		Init  Expr
	}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*Ident) exprNode()          {}
func (*IntegerLiteral) exprNode() {}
func (*BooleanLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}
func (*AssignExpr) exprNode()     {}

type BinOpKind int

const (
	Add    BinOpKind = iota // `+` (addition)
	Sub                     // `-` (subtraction)
	Mul                     // `*` (multiplication)
	Div                     // `/` (division)
	Assign                  // `=` (assignment)
)

var binOps = [...]string{
	Add:    "+",
	Sub:    "-",
	Mul:    "*",
	Div:    "/",
	Assign: "=",
}

func (op BinOpKind) String() string {
	if op < 0 || op >= BinOpKind(len(binOps)) {
		return "BinOpKind(" + strconv.FormatInt(int64(op), 10) + ")"
	}
	return binOps[op]
}

// Prec returns the operator precedence for binary operator op.
func (op BinOpKind) Prec() int {
	switch op {
	case Mul, Div:
		return 3
	case Add, Sub:
		return 2
	case Assign:
		return 1
	default:
		return 0
	}
}

type Assoc int

const (
	AssocRight Assoc = iota // Right-associative.
	AssocLeft               // Left-associative.
)

// Assoc returns the associativity of the binary operator op.
func (op BinOpKind) Assoc() Assoc {
	switch op {
	case Assign:
		return AssocRight
	case Add, Sub, Mul, Div:
		return AssocLeft
	}
	panic(fmt.Sprintf("`%s` is not a valid binary operator\n", op.String()))
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
	case token.Eq:
		binOp = Assign
	default:
		isBinOp = false
	}
	return
}
