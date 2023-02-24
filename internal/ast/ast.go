package ast

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/span"
	"github.com/aadamandersson/lue/internal/token"
)

type Expr interface {
	Span() span.Span
	exprNode()
}

// Expressions
type (
	// An identifier.
	// E.g., `foo`
	Ident struct {
		Name string
		Sp   span.Span
	}

	// An integer literal.
	// E.g., `123`
	IntegerLiteral struct {
		V  string
		Sp span.Span
	}

	// A boolean literal.
	// `true` or `false`
	BooleanLiteral struct {
		V  bool
		Sp span.Span
	}

	// A binary expression.
	// E.g., `x + y`
	BinaryExpr struct {
		X  Expr
		Op BinOp
		Y  Expr
		Sp span.Span
	}

	// A let binding.
	// `let ident [: ty] = init`
	LetExpr struct {
		Ident *Ident
		Ty    *Ident // Optional, may be nil
		Init  Expr
		Sp    span.Span
	}

	// An assignment expression.
	// `x = y`
	AssignExpr struct {
		X  Expr
		Y  Expr
		Sp span.Span
	}

	// An if expression.
	// `if cond { exprs } [else [if cond] { exprs }]`
	IfExpr struct {
		Cond Expr
		Then Expr
		Else Expr // Optional, may be nil.
		Sp   span.Span
	}

	// A block expression.
	// `{ exprs }`
	BlockExpr struct {
		Exprs []Expr
		Sp    span.Span
	}

	// Placeholder when we have some parse error.
	ErrExpr struct {
		Sp span.Span
	}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*Ident) exprNode()          {}
func (*IntegerLiteral) exprNode() {}
func (*BooleanLiteral) exprNode() {}
func (*BinaryExpr) exprNode()     {}
func (*LetExpr) exprNode()        {}
func (*AssignExpr) exprNode()     {}
func (*IfExpr) exprNode()         {}
func (*BlockExpr) exprNode()      {}
func (*ErrExpr) exprNode()        {}

func (e *Ident) Span() span.Span          { return e.Sp }
func (e *IntegerLiteral) Span() span.Span { return e.Sp }
func (e *BooleanLiteral) Span() span.Span { return e.Sp }
func (e *BinaryExpr) Span() span.Span     { return e.Sp }
func (e *LetExpr) Span() span.Span        { return e.Sp }
func (e *AssignExpr) Span() span.Span     { return e.Sp }
func (e *IfExpr) Span() span.Span         { return e.Sp }
func (e *BlockExpr) Span() span.Span      { return e.Sp }
func (e *ErrExpr) Span() span.Span        { return e.Sp }

type BinOp struct {
	Kind BinOpKind
	Sp   span.Span
}

// BinOpFromToken returns the binOp for token t and a boolean true, if its a valid binary operator.
// Otherwise, returns the zero value of BinOp and a boolean false.
func BinOpFromToken(t token.Token) (BinOp, bool) {
	var kind BinOpKind
	isBinOp := true

	switch t.Kind {
	case token.Plus:
		kind = Add
	case token.Minus:
		kind = Sub
	case token.Star:
		kind = Mul
	case token.Slash:
		kind = Div
	case token.Gt:
		kind = Gt
	case token.Lt:
		kind = Lt
	case token.Ge:
		kind = Ge
	case token.Le:
		kind = Le
	case token.EqEq:
		kind = Eq
	case token.Ne:
		kind = Ne
	case token.Eq:
		kind = Assign
	default:
		isBinOp = false
	}
	return BinOp{Kind: kind, Sp: t.Sp}, isBinOp
}

// Prec returns the operator precedence for binary operator op.
func (op BinOp) Prec() int {
	switch op.Kind {
	case Mul, Div:
		return 4
	case Add, Sub:
		return 3
	case Gt, Lt, Ge, Le, Eq, Ne:
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
func (op BinOp) Assoc() Assoc {
	switch op.Kind {
	case Assign:
		return AssocRight
	case Add, Sub, Mul, Div, Gt, Lt, Ge, Le, Eq, Ne:
		return AssocLeft
	}
	panic(fmt.Sprintf("`%s` is not a valid binary operator\n", op.Kind.String()))
}

type BinOpKind int

const (
	Add    BinOpKind = iota // `+` (addition)
	Sub                     // `-` (subtraction)
	Mul                     // `*` (multiplication)
	Div                     // `/` (division)
	Gt                      // `>` (greater than)
	Lt                      // `<` (less than)
	Ge                      // `>=` (greater than or equal)
	Le                      // `<=` (less than or equal)
	Eq                      // `==` (equality)
	Ne                      // `!=` (not equal)
	Assign                  // `=` (assignment)
)

var binOps = [...]string{
	Add:    "+",
	Sub:    "-",
	Mul:    "*",
	Div:    "/",
	Gt:     ">",
	Lt:     "<",
	Ge:     ">=",
	Le:     "<=",
	Eq:     "==",
	Ne:     "!=",
	Assign: "=",
}

func (op BinOpKind) String() string {
	if op < 0 || op >= BinOpKind(len(binOps)) {
		return "BinOpKind(" + strconv.FormatInt(int64(op), 10) + ")"
	}
	return binOps[op]
}
