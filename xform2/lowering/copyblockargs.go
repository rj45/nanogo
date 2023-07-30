package lowering

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(copyBlockArgs,
	xform2.OnlyPass(xform2.Lowering),
)

// copyBlockArgs inserts a copy at the end of each block which
// jumps to a block with parameters. This is to aide the register
// allocator in lowering out of SSA. This will produce a parallel
// copy which can later be lowered into sequential copies. This
// is important so that there is no artificial constraints imposed
// on which registers can be picked due to the order of the sequential
// copies.
func copyBlockArgs(it ir2.Iter) {
	// if not at the last instruction of the block, skip
	if (it.Block().NumInstrs() - 1) != it.InstrIndex() {
		return
	}

	blk := it.Block()

	if blk.NumArgs() == 0 {
		return
	}

	allCopied := true
	for a := 0; a < blk.NumArgs(); a++ {
		arg := blk.Arg(a)

		definstr := arg.Def()
		if definstr == nil || definstr.ID.InstrIn(blk.Func()) == nil || definstr.ID.InstrIn(blk.Func()).Op != op.Copy {
			allCopied = false
		}
	}

	if allCopied {
		// already done
		return
	}

	instr := it.Insert(op.Copy, nil)

	for a := 0; a < blk.NumArgs(); a++ {
		arg := blk.Arg(a)

		instr.InsertArg(-1, arg)
		def := blk.Func().NewValue(arg.Type)
		instr.AddDef(def)

		// replace the arg with the defined value
		blk.ReplaceArg(a, def)
	}
}
