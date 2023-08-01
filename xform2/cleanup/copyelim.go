package cleanup

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(copyElim,
	xform2.OnlyPass(xform2.CleanUp),
	xform2.OnOp(op.Copy),
)

// copyElim eliminates any copies to the same register.
// Note: this destroys SSA, so make sure it's no longer needed
// when this runs.
func copyElim(it ir2.Iter) {
	instr := it.Instr()

	for i := 0; i < instr.NumDefs(); i++ {
		def := instr.Def(i)
		arg := instr.Arg(i)

		if def.Reg() == arg.Reg() {
			def.ReplaceUsesWith(arg)
			instr.RemoveArg(arg)
			instr.RemoveDef(def)
			i--
			it.Changed()
		}
	}

	if instr.NumArgs() == 0 {
		it.Remove()
		if it.Instr() == nil {
			panic("broke iter")
		}
	}
}
