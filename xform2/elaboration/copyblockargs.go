package elaboration

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(copyBlockArgs,
	xform2.OnlyPass(xform2.Elaboration),
)

func copyBlockArgs(it ir2.Iter) {
	// if not at the last instruction of the block, skip
	if (it.Block().NumInstrs() - 1) != it.InstrIndex() {
		return
	}

	blk := it.Block()

	offset := 0
	for s := 0; s < blk.NumSuccs(); s++ {
		succ := blk.Succ(s)
		for d := 0; d < succ.NumDefs(); d++ {
			arg := blk.Arg(offset + d)

			if !arg.NeedsReg() {
				// insert a copy before the final jump
				instr := it.Insert(op.Copy, arg.Type, arg)

				// replace the arg with the defined value
				blk.ReplaceArg(offset+d, instr.Def(0))
			}
		}
		offset += succ.NumDefs()
	}
}
