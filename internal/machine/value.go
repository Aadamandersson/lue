package machine

import (
	"fmt"

	"github.com/aadamandersson/lue/internal/ir"
	"github.com/aadamandersson/lue/internal/ir/bir"
)

type Value interface {
	String() string
	sealed()
}

type (
	Integer int
	Boolean bool
	Fn      struct {
		Params []*bir.VarDecl
		Body   bir.Expr
	}
	Intrinsic ir.Intrinsic
	Unit      struct{}
)

func (Integer) sealed()   {}
func (Boolean) sealed()   {}
func (*Fn) sealed()       {}
func (Intrinsic) sealed() {}
func (Unit) sealed()      {}

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (b Boolean) String() string {
	return fmt.Sprintf("%t", b)
}

func (f *Fn) String() string {
	return "fn"
}

func (i Intrinsic) String() string {
	return ir.Intrinsic(i).String()
}

func (u Unit) String() string {
	return "()"
}
