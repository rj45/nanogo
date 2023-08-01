package elaboration

import (
	"go/types"
	"log"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(ifNonCompare,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.If),
)

// ifNonCompare fixes any if instructions without a corresponding
// comparison
func ifNonCompare(it ir2.Iter) {
	instr := it.Instr()
	arg := instr.Arg(0)
	if instr.Arg(0).Def().Instr().IsCompare() {
		// if already a compare, do nothing
		return
	}
	typ := arg.Type.Underlying()
	if basic, ok := typ.(*types.Basic); !ok || basic.Kind() != types.Bool {
		log.Panicf("unexpected type %v", typ)
	}

	compare := it.Insert(op.Equal, arg.Type, arg, true)
	instr.ReplaceArg(0, compare.Def(0))
	it.Changed()
}
