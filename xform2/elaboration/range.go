package elaboration

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(makeIterator,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.Range),
)

func makeIterator(it ir2.Iter) {
	instr := it.Instr()

	def := instr.Def(0)
	for i := 0; i < def.NumUses(); i++ {

	}
}
