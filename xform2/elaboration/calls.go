package elaboration

import (
	"go/types"

	"github.com/rj45/nanogo/ir/op"
	"github.com/rj45/nanogo/ir/reg"
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(calls,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.Call),
)

func calls(it ir2.Iter) {
	instr := it.Instr()
	fnType := instr.Arg(0).Type.(*types.Signature)

	if instr.NumArgs() > 1 && instr.Arg(1).Def() != nil && instr.Arg(1).Def().Op == op.Copy {
		return
	}

	if instr.NumDefs() > 0 && instr.Def(0).NumUses() == 1 && instr.Def(0).Use(0).Op == op.Copy {
		return
	}

	// todo:
	// - add parallel copy for clobbered regs?

	if instr.NumArgs() > 1 {
		params := fnType.Params()

		args := make([]interface{}, instr.NumArgs()-1)
		for i := 1; i < instr.NumArgs(); i++ {
			args[i-1] = instr.Arg(i)
		}

		paramCopy := it.Insert(op.Copy, params, args...)
		for i := 0; i < paramCopy.NumDefs(); i++ {
			if i < len(reg.ArgRegs) {
				paramCopy.Def(i).SetReg(reg.ArgRegs[i])
			} else {
				paramCopy.Def(i).SetArgSlot(i - len(reg.ArgRegs))
			}
			instr.ReplaceArg(i+1, paramCopy.Def(i))
		}
	}

	if instr.NumDefs() > 0 {
		results := fnType.Results()

		args := make([]interface{}, instr.NumDefs())
		for i := 0; i < instr.NumDefs(); i++ {
			args[i] = instr.Def(i)
		}

		it.Next()
		resCopy := it.Insert(op.Copy, results, args...)
		for i := 0; i < resCopy.NumArgs(); i++ {
			if i < len(reg.ArgRegs) {
				resCopy.Arg(i).SetReg(reg.ArgRegs[i])
			} else {
				resCopy.Arg(i).SetArgSlot(i - len(reg.ArgRegs))
			}

			// todo: could use a version of this that doesn't
			// clobber the current instruction or something
			instr.Def(i).ReplaceUsesWith(resCopy.Def(i))

			// switch this back to what it was
			resCopy.ReplaceArg(i, instr.Def(i))
		}
	}
}
