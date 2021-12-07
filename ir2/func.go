package ir2

import (
	"fmt"
	"go/types"

	"github.com/rj45/nanogo/ir/op"
)

const valueSlabSize = 16
const instrSlabSize = 16
const blockSlabSize = 4

func (fn *Func) Name() string {
	return fn.name
}

func (fn *Func) Package() *Package {
	return fn.pkg
}

func (fn *Func) Block(b ID) *Block {
	return fn.blocks[b]
}

func (fn *Func) Value(v ID) *Value {
	return fn.values[v]
}

func (fn *Func) Instr(i ID) *Instr {
	return fn.instrs[i]
}

func (fn *Func) NewValue(typ types.Type) *Value {
	// allocate values in contiguous slabs in memory
	// to increase data locality
	if len(fn.valueslab) == cap(fn.valueslab) {
		fn.valueslab = make([]Value, 0, valueSlabSize)
	}
	fn.valueslab = append(fn.valueslab, Value{})
	val := &fn.valueslab[len(fn.valueslab)-1]

	val.init(ID(len(fn.values)), typ)

	fn.values = append(fn.values, val)

	return val
}

func (fn *Func) NewInstr(op op.Op, typ types.Type, args ...interface{}) *Instr {
	// allocate instrs in contiguous slabs in memory
	// to increase data locality
	if len(fn.instrslab) == cap(fn.instrslab) {
		fn.instrslab = make([]Instr, 0, instrSlabSize)
	}
	fn.instrslab = append(fn.instrslab, Instr{})
	instr := &fn.instrslab[len(fn.instrslab)-1]

	instr.init(ID(len(fn.instrs)))

	fn.instrs = append(fn.instrs, instr)

	for _, a := range args {
		arg := fn.ValueFor(a)

		instr.InsertArg(-1, arg)
	}

	if tuple, ok := typ.(*types.Tuple); ok {
		for i := 0; i < tuple.Len(); i++ {
			v := tuple.At(i)
			val := fn.NewValue(v.Type())
			instr.AddDef(val)
		}
	} else {
		val := fn.NewValue(typ)
		instr.AddDef(val)
	}

	return instr
}

func (fn *Func) NewBlock() *Block {
	// allocate blocks in contiguous slabs in memory
	// to increase data locality
	if len(fn.blockslab) == cap(fn.blockslab) {
		fn.blockslab = make([]Block, 0, blockSlabSize)
	}
	fn.blockslab = append(fn.blockslab, Block{})
	blk := &fn.blockslab[len(fn.blockslab)-1]

	blk.init(fn, ID(len(fn.blocks)))

	fn.blocks = append(fn.blocks, blk)

	return blk
}

func (fn *Func) ValueFor(v interface{}) *Value {
	switch v := v.(type) {
	case *Value:
		return v
	}

	panic(fmt.Sprintf("can't get value %#v", v))
}
