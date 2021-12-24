package ir2

import "go/types"

func (use *User) init(fn *Func, id ident) {
	use.fn = fn
	use.ident = id
	use.defs = use.defstorage[:0]
	use.args = use.argstorage[:0]
}

// Func returns the containing function
func (use *User) Func() *Func {
	return use.fn
}

// Block returns either the User Block or parent Block
func (use *User) Block() *Block {
	if use.IsBlock() {
		return use.BlockIn(use.fn)
	}
	return use.Instr().blk
}

// Instr returns either the Instr or nil if User is not
// an Instr
func (use *User) Instr() *Instr {
	return use.InstrIn(use.fn)
}

// Definitions (Defs)

// Defs returns a copy of the list of Values
// defined by this user
func (use *User) Defs() []*Value {
	return append([]*Value(nil), use.defs...)
}

// NumDefs returns the number of Values defined
func (use *User) NumDefs() int {
	return len(use.defs)
}

// Def returns the ith Value defined
func (use *User) Def(i int) *Value {
	return use.defs[i]
}

// AddDef adds a Value definition
func (use *User) AddDef(val *Value) *Value {
	use.defs = append(use.defs, val)
	val.def = use
	return val
}

// updateDef updates an existing def or adds one if necessary
func (use *User) updateDef(i int, typ types.Type) *Value {
	typ = validType(typ)

	if i < len(use.defs) {
		use.defs[i].Type = typ
		return use.defs[i]
	}
	return use.AddDef(use.fn.NewValue(typ))
}

// Arguments (Args) / Operands

// Args returns a copy of the arguments
func (use *User) Args() []*Value {
	return append([]*Value(nil), use.args...)
}

// NumArgs returns the number of arguments
func (use *User) NumArgs() int {
	return len(use.args)
}

// ArgIndex returns the index of the arg, or
// -1 if not found
func (use *User) ArgIndex(arg *Value) int {
	for i, a := range use.args {
		if a == arg {
			return i
		}
	}
	return -1
}

// Arg returns the ith argument
func (use *User) Arg(i int) *Value {
	return use.args[i]
}

// InsertArg inserts the Value in the argument
// list at position i, or appending if i is -1
func (use *User) InsertArg(i int, arg *Value) {
	if arg == nil {
		panic("tried to insert a nil arg, use placeholder instead")
	}

	arg.addUse(use)

	if i < 0 || i >= len(use.args) {
		use.args = append(use.args, arg)
		return
	}

	use.args = append(use.args[:i+1], use.args[i:]...)
	use.args[i] = arg
}

// RemoveArg removes the value from the arguments list
func (use *User) RemoveArg(arg *Value) {
	i := use.ArgIndex(arg)
	arg.removeUse(use)

	use.args = append(use.args[:i], use.args[i+1:]...)
}

// ReplaceArg replaces the ith argument with the
// value specified
func (use *User) ReplaceArg(i int, arg *Value) {
	if use.ArgIndex(arg) == i {
		panic("tried to replace already existing arg")
	}

	if len(use.args) == i {
		use.InsertArg(i, arg)
		return
	}

	if use.args[i] != nil {
		use.args[i].removeUse(use)
	}
	use.args[i] = arg
	use.args[i].addUse(use)
}
