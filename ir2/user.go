package ir2

import (
	"go/types"
	"log"
)

// User uses and defines Values. Blocks and
// Instrs are Users.
type User struct {
	ID
	fn *Func

	defs []*Value
	args []*Value

	defstorage [2]*Value
	argstorage [3]*Value
}

func (use *User) init(fn *Func, id ID) {
	use.fn = fn
	use.ID = id
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

// emptyInstr is an empty Instr returned by Instr().
// To check for this, check .Kind() == UnknownID.
var emptyInstr = &Instr{}

// Instr returns either the Instr or an empty Instr to
// cut down on having to check IsInstr() everywhere.
func (use *User) Instr() *Instr {
	if use == nil {
		return emptyInstr
	}
	if use.IsInstr() {
		return use.InstrIn(use.fn)
	}
	return emptyInstr
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
	if use.Kind() == UnknownID {
		log.Panicf("tried to add def %v to unknown/empty user", val)
	}
	use.defs = append(use.defs, val)
	val.def = use
	return val
}

// RemoveDef removes the value from the defs list
func (use *User) RemoveDef(def *Value) {
	index := -1
	for i, d := range use.defs {
		if d == def {
			index = i
			break
		}
	}

	if index < 0 {
		panic("attempt to remove non-existant def")
	}

	def.def = nil

	use.defs = append(use.defs[:index], use.defs[index+1:]...)
}

// updateDef updates an existing def or adds one if necessary
func (use *User) updateDef(i int, typ types.Type) *Value {
	if use.Kind() == UnknownID {
		log.Panicf("tried to update def %d:%v on unknown/empty user", i, typ)
	}
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
	if use.Kind() == UnknownID {
		log.Panicf("tried to add arg %d:%v on unknown/empty user", i, arg)
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
// value specified. Will call InsertArg instead
// if i == NumArgs().
func (use *User) ReplaceArg(i int, arg *Value) {
	if use.ArgIndex(arg) == i {
		panic("tried to replace already existing arg")
	}
	if use.Kind() == UnknownID {
		log.Panicf("tried to replace arg with %d:%v on unknown/empty user", i, arg)
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
