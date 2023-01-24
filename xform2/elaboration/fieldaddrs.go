package elaboration

import (
	"go/types"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/sizes"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(fieldAddrs,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.FieldAddr),
)

func fieldAddrs(it ir2.Iter) {
	instr := it.Instr()

	field, ok := ir2.IntValue(instr.Arg(0).Const())
	if !ok {
		panic("expected int constant")
	}

	elem := instr.Arg(1).Type.(*types.Pointer).Elem()
	strct := elem.Underlying().(*types.Struct)
	fieldType := strct.Field(field).Type()
	fieldPtr := types.NewPointer(fieldType)

	fields := sizes.Fieldsof(strct)
	offsets := sizes.Offsetsof(fields)
	offset := offsets[field]

	if offset == 0 {
		// would just be adding zero, so this instruction can just be removed
		instr.Def(0).ReplaceUsesWith(instr.Arg(1))
		it.Remove()
		return
	}

	it.Update(op.Add, fieldPtr, instr.Arg(1), offset)
}
