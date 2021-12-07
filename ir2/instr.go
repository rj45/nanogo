package ir2

import (
	"log"
)

func (in *Instr) init(id ID) {
	in.ID = id
	in.defs = in.defstorage[:0]
	in.args = in.argstorage[:0]
}

// Func returns the Func the Instr is in
func (in *Instr) Func() *Func {
	return in.blk.fn
}

// Block returns the containing Block
func (in *Instr) Block() *Block {
	return in.blk
}

// Index returns the index in the Block's
// Instr list
func (in *Instr) Index() int {
	if in.blk.instrs[in.index] != in {
		log.Panicf("bad index on %v", in)
	}
	return in.index
}

// MoveBefore moves this instruction before other
func (in *Instr) MoveBefore(other *Instr) {
	in.blk.RemoveInstr(in)
	other.blk.InsertInstr(other.Index(), in)
}

// MoveAfter moves this instruction after other
func (in *Instr) MoveAfter(other *Instr) {
	in.blk.RemoveInstr(in)
	other.blk.InsertInstr(other.Index()+1, in)
}

// Definitions (Defs)

// Defs returns a copy of the list of Values
// defined by this Instr
func (in *Instr) Defs() []*Value {
	return append([]*Value(nil), in.defs...)
}

// NumDefs returns the number of Values defined
func (in *Instr) NumDefs() int {
	return len(in.defs)
}

// Def returns the ith Value defined
func (in *Instr) Def(i int) *Value {
	return in.defs[i]
}

// AddDef adds a Value definition
func (in *Instr) AddDef(val *Value) *Value {
	in.defs = append(in.defs, val)
	val.def = in
	return val
}

// Arguments (Args) / Operands

// Args returns a copy of the arguments
func (in *Instr) Args() []*Value {
	return append([]*Value(nil), in.args...)
}

// NumArgs returns the number of arguments
func (in *Instr) NumArgs() int {
	return len(in.defs)
}

// ArgIndex returns the index of the arg, or
// -1 if not found
func (in *Instr) ArgIndex(arg *Value) int {
	for i, a := range in.args {
		if a == arg {
			return i
		}
	}
	return -1
}

// Arg returns the ith argument
func (in *Instr) Arg(i int) *Value {
	return in.args[i]
}

// InsertArg inserts the Value in the argument
// list at position i, or appending if i is -1
func (in *Instr) InsertArg(i int, arg *Value) {
	arg.addUse(in)

	if i < 0 || i >= len(in.args) {
		in.args = append(in.args, arg)
		return
	}

	in.args = append(in.args[:i+1], in.args[i:]...)
	in.args[i] = arg
}

// RemoveArg removes the value from the arguments list
func (in *Instr) RemoveArg(arg *Value) {
	i := in.ArgIndex(arg)
	arg.removeUse(in)

	in.args = append(in.args[:i], in.args[i+1:]...)
}

// ReplaceArg replaces the ith argument with the
// value specified
func (in *Instr) ReplaceArg(i int, arg *Value) {
	if in.ArgIndex(arg) != -1 {
		panic("tried to replace already existing arg")
	}

	in.args[i].removeUse(in)
	in.args[i] = arg
	in.args[i].addUse(in)
}
