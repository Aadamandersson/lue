package bir

import (
	"github.com/aadamandersson/lue/internal/ir"
	"github.com/aadamandersson/lue/internal/ir/ast"
)

type (
	Expr interface {
		Type() *Ty
		isExpr()
	}
)

// Expressions
type (
	// A reference to a function.
	Fn struct {
		Decl *ast.FnDecl
		In   []*VarDecl
		Out  *Ty
		Body Expr
	}

	// A reference to a class.
	Class struct {
		Decl   *ast.ClassDecl
		Fields []*VarDecl
	}

	// A variable declaration.
	// `ident: ty`
	VarDecl struct {
		Ident *ir.Ident
		Ty    *Ty
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
func (*Fn) isExpr()             {}
func (*Class) isExpr()          {}
func (*VarDecl) isExpr()        {}
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
func (Intrinsic) isExpr()       {}
func (*ErrExpr) isExpr()        {}

func (e *Fn) Type() *Ty             { return e.Out }
func (e *Class) Type() *Ty          { return NewClass((*ir.Ident)(e.Decl.Ident)) }
func (e *VarDecl) Type() *Ty        { return e.Ty }
func (e *IntegerLiteral) Type() *Ty { return BasicTys[TyInt] }
func (e *BooleanLiteral) Type() *Ty { return BasicTys[TyBool] }
func (e *StringLiteral) Type() *Ty  { return BasicTys[TyString] }
func (e *BinaryExpr) Type() *Ty     { return e.Op.Ty }
func (e *LetExpr) Type() *Ty        { return BasicTys[TyUnit] }
func (e *AssignExpr) Type() *Ty     { return BasicTys[TyUnit] }
func (e *IfExpr) Type() *Ty         { return e.Then.Type() }
func (e *BlockExpr) Type() *Ty {
	if len(e.Exprs) == 0 {
		return BasicTys[TyUnit]
	}
	return e.Exprs[len(e.Exprs)-1].Type()
}
func (e *CallExpr) Type() *Ty { return e.Fn.Type() }
func (e *ArrayExpr) Type() *Ty {
	if len(e.Exprs) == 0 {
		return NewArray(BasicTys[TyInfer])
	}
	return NewArray(e.Exprs[0].Type())
}
func (e *IndexExpr) Type() *Ty { return e.Arr.(*VarDecl).Ty.Elem }
func (e *ForExpr) Type() *Ty   { return e.Body.Type() }
func (e *BreakExpr) Type() *Ty {
	if e.X == nil {
		return BasicTys[TyUnit]
	}
	return e.X.Type()
}
func (e *ReturnExpr) Type() *Ty {
	if e.X == nil {
		return BasicTys[TyUnit]
	}
	return e.X.Type()
}
func (e Intrinsic) Type() *Ty { return BasicTys[TyUnit] }
func (e *ErrExpr) Type() *Ty  { return BasicTys[TyErr] }

type TyKind int

const (
	TyErr TyKind = iota
	TyInfer
	TyInt
	TyBool
	TyString
	TyArray
	TyClass
	TyUnit
)

var BasicTys = [...]*Ty{
	TyErr:    {Kind: TyErr},
	TyInfer:  {Kind: TyInfer},
	TyInt:    {Kind: TyInt},
	TyBool:   {Kind: TyBool},
	TyString: {Kind: TyString},
	TyUnit:   {Kind: TyUnit},
}

type Ty struct {
	Kind  TyKind
	Elem  *Ty
	Class *ir.Ident
}

func (t *Ty) IsErr() bool {
	return t.Kind == TyErr
}

func (t *Ty) IsInfer() bool {
	return t.Kind == TyInfer
}

func (t *Ty) IsInt() bool {
	return t.Kind == TyInt
}

func (t *Ty) IsBool() bool {
	return t.Kind == TyBool
}

func (t *Ty) IsString() bool {
	return t.Kind == TyString
}

func (t *Ty) IsUnit() bool {
	return t.Kind == TyUnit
}

func (t *Ty) IsArray() bool {
	return t.Kind == TyArray
}

func (t *Ty) Equal(other *Ty) bool {
	if t.Kind == other.Kind {
		if t.Kind == TyArray {
			return t.Elem.Kind == other.Elem.Kind
		}
		return true
	}
	return false
}

func NewArray(elem *Ty) *Ty {
	return &Ty{Kind: TyArray, Elem: elem}
}

func NewClass(ident *ir.Ident) *Ty {
	return &Ty{Kind: TyClass, Class: ident}
}

func (t *Ty) String() string {
	switch t.Kind {
	case TyErr:
		return "?"
	case TyInfer:
		return "?"
	case TyInt:
		return "int"
	case TyBool:
		return "bool"
	case TyString:
		return "string"
	case TyArray:
		return "[" + t.Elem.String() + "]"
	case TyUnit:
		return "()"
	default:
		panic("unreachable")
	}
}

type BinOp struct {
	Kind BinOpKind
	Ty   *Ty
}

var binOps = [...]struct {
	in  ast.BinOpKind
	xTy TyKind
	yTy TyKind
	out BinOp
}{
	{ast.Add, TyInt, TyInt, BinOp{Kind: Add, Ty: BasicTys[TyInt]}},
	{ast.Sub, TyInt, TyInt, BinOp{Kind: Sub, Ty: BasicTys[TyInt]}},
	{ast.Mul, TyInt, TyInt, BinOp{Kind: Mul, Ty: BasicTys[TyInt]}},
	{ast.Div, TyInt, TyInt, BinOp{Kind: Div, Ty: BasicTys[TyInt]}},

	{ast.Gt, TyInt, TyInt, BinOp{Kind: Gt, Ty: BasicTys[TyBool]}},
	{ast.Lt, TyInt, TyInt, BinOp{Kind: Lt, Ty: BasicTys[TyBool]}},
	{ast.Ge, TyInt, TyInt, BinOp{Kind: Ge, Ty: BasicTys[TyBool]}},
	{ast.Le, TyInt, TyInt, BinOp{Kind: Le, Ty: BasicTys[TyBool]}},

	{ast.Eq, TyInt, TyInt, BinOp{Kind: Eq, Ty: BasicTys[TyBool]}},
	{ast.Eq, TyBool, TyBool, BinOp{Kind: Eq, Ty: BasicTys[TyBool]}},
	{ast.Eq, TyString, TyString, BinOp{Kind: Eq, Ty: BasicTys[TyBool]}},

	{ast.Ne, TyInt, TyInt, BinOp{Kind: Ne, Ty: BasicTys[TyBool]}},
	{ast.Ne, TyBool, TyBool, BinOp{Kind: Ne, Ty: BasicTys[TyBool]}},
	{ast.Ne, TyString, TyString, BinOp{Kind: Ne, Ty: BasicTys[TyBool]}},
}

func BindBinOp(astOp ast.BinOpKind, xTy, yTy TyKind) (BinOp, bool) {
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
