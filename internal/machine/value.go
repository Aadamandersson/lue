package machine

import (
	"fmt"
	"strings"

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
	String  string
	Array   struct {
		Elems []Value
	}
	Fn struct {
		Params []*bir.VarDecl
		Body   bir.Expr
	}
	RetVal struct {
		V Value
	}
	BreakVal struct {
		V Value
	}
	Intrinsic ir.Intrinsic
	Unit      struct{}
)

func (Integer) sealed()   {}
func (Boolean) sealed()   {}
func (String) sealed()    {}
func (*Array) sealed()    {}
func (*Fn) sealed()       {}
func (*RetVal) sealed()   {}
func (*BreakVal) sealed() {}
func (Intrinsic) sealed() {}
func (Unit) sealed()      {}

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (b Boolean) String() string {
	return fmt.Sprintf("%t", b)
}

func (s String) String() string {
	return string(s)
}

func (a *Array) String() string {
	var builder strings.Builder

	builder.WriteByte('[')
	for i := 0; i < len(a.Elems); i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(a.Elems[i].String())
	}
	builder.WriteByte(']')

	return builder.String()
}

func (f *Fn) String() string {
	return "fn"
}

func (r *RetVal) String() string {
	return r.V.String()
}

func (r *BreakVal) String() string {
	return r.V.String()
}

func (i Intrinsic) String() string {
	return ir.Intrinsic(i).String()
}

func (u Unit) String() string {
	return "()"
}
