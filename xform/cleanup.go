package xform

import (
	"github.com/rj45/nanogo/ir"
	"github.com/rj45/nanogo/ir/op"
)

func dePhi(val *ir.Value) int {
	if val.Op != op.Phi {
		return 0
	}

	for i := 0; i < val.NumArgs(); i++ {
		src := val.Arg(i)
		if src.Op.IsConst() || val.Reg != src.Reg {
			// todo might actually need to be a swap instead
			pred := val.Block().Pred(i)
			pred.InsertCopy(-1, src, val.Reg)
			// TODO: fix copy to pass proper value
		}
	}

	val.Remove()

	return 1
}

var _ = addToPass(CleanUp, dePhi)

func deCopy(val *ir.Value) int {
	if val.Op != op.Copy {
		return 0
	}

	if val.Reg == val.Arg(0).Reg {
		val.ReplaceWith(val.Arg(0))
		return 1
	}

	return 0
}

func EliminateEmptyBlocks(fn *ir.Func) {
	blks := fn.Blocks()
retry:
	for {
		for i, blk := range blks {
			if blk.NumInstrs() == 0 && blk.Op == op.Jump {
				fn.RemoveBlock(i)
				continue retry
			}
		}
		break
	}
}

var _ = addToPass(CleanUp, deCopy)