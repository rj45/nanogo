package ir2

import (
	"log"
)

func (in *Instr) init(id ID) {
	in.ID = id
	in.defs = in.defstorage[:0]
	in.args = in.argstorage[:0]
}

func (in *Instr) Func() *Func {
	return in.blk.fn
}

func (in *Instr) Block() *Block {
	return in.blk
}

func (in *Instr) Index() int {
	if in.blk.instrs[in.index] != in {
		log.Panicf("bad index on %v", in)
	}
	return in.index
}

// Definitions (Defs)

func (in *Instr) Defs() []*Value {
	return append([]*Value(nil), in.defs...)
}

func (in *Instr) NumDefs() int {
	return len(in.defs)
}

func (in *Instr) Def(i int) *Value {
	return in.defs[i]
}

func (in *Instr) AddDef(val *Value) *Value {
	in.defs = append(in.defs, val)
	val.def = in
	return val
}

// Arguments (Args) / Operands

func (in *Instr) Args() []*Value {
	return append([]*Value(nil), in.args...)
}

func (in *Instr) NumArgs() int {
	return len(in.defs)
}

func (in *Instr) ArgIndex(arg *Value) int {
	for i, a := range in.args {
		if a == arg {
			return i
		}
	}
	return -1
}

func (in *Instr) Arg(i int) *Value {
	return in.args[i]
}

func (in *Instr) InsertArg(i int, arg *Value) {
	arg.addUse(in)

	if i < 0 || i >= len(in.args) {
		in.args = append(in.args, arg)
		return
	}

	in.args = append(in.args[:i+1], in.args[i:]...)
	in.args[i] = arg
}

func (in *Instr) RemoveArg(arg *Value) {
	i := in.ArgIndex(arg)
	arg.removeUse(in)

	in.args = append(in.args[:i], in.args[i+1:]...)
}
