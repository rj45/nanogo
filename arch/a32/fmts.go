package a32

import (
	"github.com/rj45/nanogo/codegen/asm"
	"github.com/rj45/nanogo/ir"
)

type Fmt int

const (
	BadFmt Fmt = iota
	LoadFmt
	StoreFmt
	MoveFmt
	CompareFmt
	BinaryFmt
	UnaryFmt
	CallFmt
	NoFmt
)

var templates = [...]string{
	BadFmt:     "",
	LoadFmt:    "%s, [%s + %s]",
	StoreFmt:   "[%s + %s], %s",
	MoveFmt:    "%s, %s",
	CompareFmt: "%s, %s",
	BinaryFmt:  "%s, %s, %s",
	UnaryFmt:   "%s, %s",
	CallFmt:    "%s",
	NoFmt:      "",
}

func (f Fmt) Template() string {
	return templates[f]
}

func (f Fmt) Vars(val *ir.Value) []*asm.Var {
	switch f {
	case LoadFmt:
		return []*asm.Var{varFor(val), varFor(val.Arg(0)), varFor(val.Arg(1))}
	case StoreFmt:
		return []*asm.Var{varFor(val.Arg(0)), varFor(val.Arg(1)), varFor(val.Arg(2))}
	case MoveFmt:
		return []*asm.Var{varFor(val), varFor(val.Arg(0))}
	case CompareFmt:
		return []*asm.Var{varFor(val.Arg(0)), varFor(val.Arg(1))}
	case BinaryFmt:
		return []*asm.Var{varFor(val), varFor(val.Arg(0)), varFor(val.Arg(1))}
	case UnaryFmt:
		return []*asm.Var{varFor(val), varFor(val.Arg(0))}
	case CallFmt:
		return []*asm.Var{varFor(val.Arg(0))}
	}
	return nil
}

func varFor(val *ir.Value) *asm.Var {
	return &asm.Var{Value: val}
}
