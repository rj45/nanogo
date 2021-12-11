package ir2

import (
	"go/types"
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

// Update changes the op, type and number of defs and the args
func (in *Instr) Update(op Op, typ types.Type, args ...interface{}) {
	in.update(in.blk.fn, op, typ, args)
}

func (in *Instr) update(fn *Func, op Op, typ types.Type, args []interface{}) {
	in.Op = op

	for i, a := range args {
		arg := fn.ValueFor(typ, a)

		in.ReplaceArg(i, arg)
	}

	for len(in.args) > len(args) {
		// todo: replace with in.RemoveArgAt()?
		in.RemoveArg(in.args[len(in.args)-1])
	}

	if tuple, ok := typ.(*types.Tuple); ok {
		for i := 0; i < tuple.Len(); i++ {
			v := tuple.At(i)

			in.updateDef(fn, i, v.Type())
		}
	} else if typ != nil {
		in.updateDef(fn, 0, typ)
	}
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

// updateDef updates an existing def or adds one if necessary
func (in *Instr) updateDef(fn *Func, i int, typ types.Type) *Value {
	if len(in.defs) < i {
		in.defs[i].Type = typ
		return in.defs[i]
	}
	return in.AddDef(fn.NewValue(typ))
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
	if arg != nil {
		arg.addUse(in)
	}

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
	if in.ArgIndex(arg) == i {
		panic("tried to replace already existing arg")
	}

	if len(in.args) == i {
		in.InsertArg(i, arg)
		return
	}

	if in.args[i] != nil {
		in.args[i].removeUse(in)
	}
	in.args[i] = arg
	in.args[i].addUse(in)
}
