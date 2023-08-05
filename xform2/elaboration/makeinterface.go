package elaboration

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(makeInterface,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.MakeInterface),
)

func makeInterface(it ir2.Iter) {
	instr := it.Instr()

	// todo: properly implement this
	instr.Op = op.Copy
	it.Changed()
}
