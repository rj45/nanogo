package cleanup

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(copies,
	xform2.OnlyPass(xform2.CleanUp),
	xform2.OnOp(op.Copy),
)

func copies(it ir2.Iter) {
	instr := it.Instr()

	_ = instr
}
