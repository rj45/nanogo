package legalization

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(addClobberCopies,
	xform2.OnlyPass(xform2.Legalization),
)

// addClobberCopies adds copies for operands that get clobbered
// on two-operand architectures
func addClobberCopies(it ir2.Iter) {
	instr := it.Instr()
	if !instr.Op.ClobbersArg() {
		return
	}

	if instr.NumArgs() < 1 {
		return
	}

	if instr.Arg(0).Def().Instr().Op != nil && instr.Arg(0).Def().Instr().Op.IsCopy() {
		// already added the copy
		return
	}

	cp := it.Insert(op.Copy, instr.Arg(0).Type, instr.Arg(0))
	instr.ReplaceArg(0, cp.Def(0))
}
