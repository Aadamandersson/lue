package ast

import (
	"fmt"
	"strconv"

	"github.com/aadamandersson/lue/internal/ir"
	"github.com/aadamandersson/lue/internal/span"
	"github.com/aadamandersson/lue/internal/token"
)

type (
	Item interface {
		isItem()
	}

	Expr interface {
		Span() span.Span
		isExpr()
	}
)

// Items
type (
	// A function declaration.
	// `fn ident([params]) [: ty] { exprs }`
	FnDecl struct {
		Ident *Ident
		In    []*VarDecl
		Out   *Ty
		Body  Expr
		Sp    span.Span
	}

	// A class declaration.
	// `class ident { fields }`
	ClassDecl struct {
		Ident  *Ident
		Fields []*VarDecl
		Sp     span.Span
	}

	// Placeholder when we have some parse error.
	ErrItem struct{}
)

type VarDecl struct {
	Ident *Ident
	Ty    *Ty
}

type TyKind int

const (
	TyInfer TyKind = iota
	TyArray
	TyIdent
	TyUnit
)

type Ty struct {
	Kind  TyKind
	Ident *Ident // Nil if kind is `TyInfer` or `TyUnit`
	Sp    span.Span
}

func (t *Ty) String() string {
	switch t.Kind {
	case TyInfer:
		return "?"
	case TyArray:
		return "[" + t.Ident.Name + "]"
	case TyIdent:
		return t.Ident.Name
	case TyUnit:
		return "()"
	default:
		panic("unreachable")
	}
}

// Ensure that we can only assign item nodes to an Item.
func (*FnDecl) isItem()    {}
func (*ClassDecl) isItem() {}
func (*ErrItem) isItem()   {}

// Expressions
type (
	// An identifier.
	// E.g., `foo`
	Ident ir.Ident

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

	// A string literal.
	// E.g., `"foo"`
	StringLiteral struct {
		V  string
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
		Decl *VarDecl
		Init Expr
		Sp   span.Span
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

	// A function call.
	// `fn(args)`
	CallExpr struct {
		Fn   Expr
		Args []Expr
		Sp   span.Span
	}

	// An array expression.
	// `[1, 2, 3]`
	ArrayExpr struct {
		Exprs []Expr
		Sp    span.Span
	}

	// An array indexing expression.
	// `arr[i]`
	IndexExpr struct {
		Arr Expr
		I   Expr
		Sp  span.Span
	}

	// A for loop.
	// `for { exprs }`
	ForExpr struct {
		Body Expr
		Sp   span.Span
	}

	// A break expression.
	// `break [expr]`
	BreakExpr struct {
		X  Expr // Optional, may be nil.
		Sp span.Span
	}

	// A return expression.
	// `return [expr]`
	ReturnExpr struct {
		X  Expr // Optional, may be nil.
		Sp span.Span
	}

	// Placeholder when we have some parse error.
	ErrExpr struct {
		Sp span.Span
	}
)

// Ensure that we can only assign expression nodes to an Expr.
func (*Ident) isExpr()          {}
func (*IntegerLiteral) isExpr() {}
func (*BooleanLiteral) isExpr() {}
func (*StringLiteral) isExpr()  {}
func (*BinaryExpr) isExpr()     {}
func (*LetExpr) isExpr()        {}
func (*AssignExpr) isExpr()     {}
func (*IfExpr) isExpr()         {}
func (*BlockExpr) isExpr()      {}
func (*CallExpr) isExpr()       {}
func (*ArrayExpr) isExpr()      {}
func (*IndexExpr) isExpr()      {}
func (*ForExpr) isExpr()        {}
func (*BreakExpr) isExpr()      {}
func (*ReturnExpr) isExpr()     {}
func (*ErrExpr) isExpr()        {}

func (e *Ident) Span() span.Span          { return e.Sp }
func (e *IntegerLiteral) Span() span.Span { return e.Sp }
func (e *BooleanLiteral) Span() span.Span { return e.Sp }
func (e *StringLiteral) Span() span.Span  { return e.Sp }
func (e *BinaryExpr) Span() span.Span     { return e.Sp }
func (e *LetExpr) Span() span.Span        { return e.Sp }
func (e *AssignExpr) Span() span.Span     { return e.Sp }
func (e *IfExpr) Span() span.Span         { return e.Sp }
func (e *BlockExpr) Span() span.Span      { return e.Sp }
func (e *CallExpr) Span() span.Span       { return e.Sp }
func (e *ArrayExpr) Span() span.Span      { return e.Sp }
func (e *IndexExpr) Span() span.Span      { return e.Sp }
func (e *ForExpr) Span() span.Span        { return e.Sp }
func (e *BreakExpr) Span() span.Span      { return e.Sp }
func (e *ReturnExpr) Span() span.Span     { return e.Sp }
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
