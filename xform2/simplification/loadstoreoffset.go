package simplification

import (
	"go/types"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(loadOffset,
	xform2.OnlyPass(xform2.Simplification),
	xform2.OnOp(op.Load),
	xform2.Tags(xform2.LoadStoreOffset),
)

func loadOffset(it ir2.Iter) {
	instr := it.Instr()
	if instr.NumArgs() > 1 {
		return
	}

	add := offset(instr)
	if add == nil {
		instr.InsertArg(1, instr.Func().ValueFor(types.Typ[types.UntypedInt], 0))
		return
	}

	// combine the add with the load
	instr.ReplaceArg(0, add.Arg(0))
	instr.InsertArg(-1, add.Arg(1))
	it.RemoveInstr(add)
}

var _ = xform2.Register(storeOffset,
	xform2.OnlyPass(xform2.Simplification),
	xform2.OnOp(op.Store),
	xform2.Tags(xform2.LoadStoreOffset),
)

func storeOffset(it ir2.Iter) {
	instr := it.Instr()

	if instr.NumArgs() > 2 {
		return
	}

	add := offset(instr)
	if add == nil {
		instr.InsertArg(1, instr.Func().ValueFor(types.Typ[types.UntypedInt], 0))
		return
	}

	// combine the add with the store
	instr.ReplaceArg(0, add.Arg(0))
	instr.InsertArg(1, add.Arg(1))
	it.RemoveInstr(add)
}

func offset(instr *ir2.Instr) *ir2.Instr {
	if instr.Arg(0).IsConst() {
		return nil
	}

	add := instr.Arg(0).Def().Instr()
	if add.Op == op.Add && add.Arg(1).IsConst() && add.Def(0).NumUses() == 1 {
		return add
	}
	return nil
}
