package simplification

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(mulByConst,
	xform2.OnlyPass(xform2.Simplification),
	xform2.OnOp(op.Mul),
)

func mulByConst(it ir2.Iter) {
	instr := it.Instr()
	if !instr.Arg(1).IsConst() {
		return
	}

	amt, ok := ir2.Int64Value(instr.Arg(1).Const())
	if !ok {
		amt32, ok := ir2.IntValue(instr.Arg(1).Const())
		if !ok {
			return
		}
		amt = int64(amt32)
	}

	if amt == 1 {
		instr.Def(0).ReplaceUsesWith(instr.Arg(0))
		it.Remove()
		return
	}

	if amt == 0 {
		instr.Def(0).ReplaceUsesWith(instr.Arg(1))
		it.Remove()
		return
	}

	i := int64(1)
	n := int64(0)
	for i = 1; i < amt; i <<= 1 {
		n++
	}
	if i != amt {
		// TODO: can use multiple shifts and adds to calculate this
		return
	}

	it.Update(op.ShiftLeft, nil, instr.Arg(0), instr.Func().ValueFor(instr.Arg(0).Type, int(n)))
}
