package elaboration

import (
	"go/types"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(index,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.Index),
)

// index converts `Index` instructions, usually indexing a string
func index(it ir2.Iter) {
	instr := it.Instr()
	arg0 := instr.Arg(0)

	typ, ok := arg0.Type.(*types.Basic)
	if !ok {
		return
	}

	// only handle indexing strings for now
	if typ.Kind() != types.String {
		return
	}

	load := it.Insert(op.Load, types.Typ[types.Uintptr], arg0)
	address := load.Def(0)

	add := it.Insert(op.Add, address.Type, address, instr.Arg(1))

	it.Update(op.Load, instr.Def(0).Type, add.Def(0))
}
