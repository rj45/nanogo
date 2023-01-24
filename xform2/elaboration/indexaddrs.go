package elaboration

import (
	"go/types"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/sizes"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(indexAddrs,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.IndexAddr),
)

// indexAddrs converts `IndexAddr` instructions into a `mul` and `add` instruction
// The `mul` is by a constant which can be optimized into shifts and adds later.
func indexAddrs(it ir2.Iter) {
	instr := it.Instr()
	elem := instr.Def(0).Type.(*types.Pointer).Elem()

	size := sizes.Sizeof(elem)

	mul := it.Insert(op.Mul, types.Typ[types.Int], instr.Arg(1), size)

	instr.Op = op.Add
	instr.ReplaceArg(1, mul.Def(0))
}
