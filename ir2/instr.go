package ir2

import (
	"go/token"
	"go/types"
	"log"
)

// Instr is an instruction that may define one or more Values,
// and take as args (operands) one or more Values.
type Instr struct {
	User
	Op

	blk *Block

	Pos   token.Pos
	index int
}

// Op describes an operation (instruction) type
// Note: Implementations of Op should attempt to be uint8 type,
// since this is optimized by Go.
type Op interface {
	String() string
	IsCall() bool
	IsCompare() bool
	IsCopy() bool
	IsCommutative() bool
	IsSink() bool
	ClobbersArg() bool
	IsBranch() bool
}

// Index returns the index in the Block's Instr list
func (in *Instr) Index() int {
	if in.blk.instrs[in.index] != in {
		log.Panicf("bad index on %v", in)
	}
	return in.index
}

// LineNo returns the line number in the original Go source code,
// or 0 if that's not known
func (in *Instr) LineNo() int {
	fset := in.fn.pkg.prog.FileSet
	pos := fset.Position(in.Pos)
	return pos.Line
}

// Update changes the op, type and number of defs and the args
func (in *Instr) Update(op Op, typ types.Type, args ...interface{}) {
	in.update(op, typ, args)
}

func (in *Instr) update(op Op, typ types.Type, args []interface{}) {
	in.Op = op

	if !op.IsSink() {
		if tuple, ok := typ.(*types.Tuple); ok {
			for i := 0; i < tuple.Len(); i++ {
				v := tuple.At(i)

				in.updateDef(i, v.Type())
			}
		} else if typ != nil {
			in.updateDef(0, typ)
		}
	}

	offset := 0
	for i, a := range args {
		if list, ok := a.([]*Value); ok {
			for _, a := range list {
				if i+offset >= len(in.args) || a != in.args[i+offset] {
					in.ReplaceArg(i+offset, a)
				}
				offset++
			}
			offset--
			continue
		}

		arg := in.fn.ValueFor(typ, a)

		if i+offset >= len(in.args) || arg != in.args[i+offset] {
			in.ReplaceArg(i+offset, arg)
		}
	}

	for len(in.args) > (len(args) + offset) {
		// todo: replace with in.RemoveArgAt()?
		in.RemoveArg(in.args[len(in.args)-1])
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
