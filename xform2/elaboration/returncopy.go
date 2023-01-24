package elaboration

import (
	"github.com/rj45/nanogo/ir/reg"
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(returnCopy,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.Return),
	xform2.Once(),
)

func returnCopy(it ir2.Iter) {
	fn := it.Block().Func()

	ret := it.Instr()
	if ret.NumArgs() < 1 {
		return
	}

	results := fn.Sig.Results()
	cp := it.Insert(op.Copy, results, ret.Args())

	for i := 0; i < ret.NumArgs(); i++ {
		if i < len(reg.ArgRegs) {
			cp.Def(i).SetReg(reg.ArgRegs[i])
		} else {
			cp.Def(i).SetArgSlot(i - len(reg.ArgRegs))
		}
		ret.ReplaceArg(i, cp.Def(i))
	}
}
