package elaboration

import (
	"log"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/sizes"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(converts,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.Convert),
)

func converts(it ir2.Iter) {
	instr := it.Instr()

	destsize := sizes.Sizeof(instr.Def(0).Type)
	srcsize := sizes.Sizeof(instr.Arg(0).Type)

	// todo: do we need to sign extend and/or zero extend?
	if srcsize > destsize && srcsize > sizes.WordSize() {
		log.Fatalf("Unable to convert %#v to %#v in %s",
			instr.Arg(0).Type,
			instr.Def(0).Type,
			instr.Func().FullName)
	}

	instr.Def(0).ReplaceUsesWith(instr.Arg(0))
	it.Remove()
}
