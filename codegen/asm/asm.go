package asm

import (
	"github.com/rj45/nanogo/ir"
)

type Section string

const (
	Code Section = "code"
	Data Section = "data"
	Bss  Section = "bss"
)

type Program struct {
	Pkg   *ir.Package
	Funcs []Func
}

type Func struct {
	Comment string
	Label   string

	Globals []*Global
	Blocks  []*Block
	Func    *ir.Func
}

type Global struct {
	Section Section
	Comment string
	Label   string
	Strings []string
	Value   *ir.Value
}

type Op interface {
	Asm() string
	Fmt() Fmt
	IsMove() bool
	IsCall() bool
}

type Fmt interface {
	Template() string
	Vars(val *ir.Value) []*Var
}

type Instr struct {
	Op     Op
	Args   []*Var
	Indent bool
}

type Block struct {
	Label  string
	Instrs []*Instr

	Block *ir.Block
}

type Var struct {
	String string
	Value  *ir.Value
	Block  *ir.Block
}
