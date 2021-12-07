package ir2

import (
	"fmt"
	"go/types"
	"log"

	"github.com/rj45/nanogo/ir/op"
)

// slab allocation sizes
const valueSlabSize = 16
const instrSlabSize = 16
const blockSlabSize = 4

// Package returns the Func's Package
func (fn *Func) Package() *Package {
	return fn.pkg
}

// ValueForID returns the Value for the ID
func (fn *Func) ValueForID(v ID) *Value {
	return fn.idValues[v]
}

// NewValue creates a new Value of type typ
func (fn *Func) NewValue(typ types.Type) *Value {
	// allocate values in contiguous slabs in memory
	// to increase data locality
	if len(fn.valueslab) == cap(fn.valueslab) {
		fn.valueslab = make([]Value, 0, valueSlabSize)
	}
	fn.valueslab = append(fn.valueslab, Value{})
	val := &fn.valueslab[len(fn.valueslab)-1]

	val.init(ID(len(fn.idValues)), typ)

	fn.idValues = append(fn.idValues, val)

	return val
}

// ValueFor looks up an existing Value
func (fn *Func) ValueFor(v interface{}) *Value {
	switch v := v.(type) {
	// todo: add constants and funcs
	case *Value:
		return v
	}

	panic(fmt.Sprintf("can't get value %#v", v))
}

// Instrs

// InstrForID returns the Instr for the ID
func (fn *Func) InstrForID(i ID) *Instr {
	return fn.idInstrs[i]
}

// NewInstr creates an unbound Instr
func (fn *Func) NewInstr(op op.Op, typ types.Type, args ...interface{}) *Instr {
	// allocate instrs in contiguous slabs in memory
	// to increase data locality
	if len(fn.instrslab) == cap(fn.instrslab) {
		fn.instrslab = make([]Instr, 0, instrSlabSize)
	}
	fn.instrslab = append(fn.instrslab, Instr{})
	instr := &fn.instrslab[len(fn.instrslab)-1]

	instr.init(ID(len(fn.idInstrs)))

	fn.idInstrs = append(fn.idInstrs, instr)

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

// Blocks

// BlockForID returns a Block by ID
func (fn *Func) BlockForID(b ID) *Block {
	return fn.idBlocks[b]
}

// NewBlock adds a new block
func (fn *Func) NewBlock() *Block {
	// allocate blocks in contiguous slabs in memory
	// to increase data locality
	if len(fn.blockslab) == cap(fn.blockslab) {
		fn.blockslab = make([]Block, 0, blockSlabSize)
	}
	fn.blockslab = append(fn.blockslab, Block{})
	blk := &fn.blockslab[len(fn.blockslab)-1]

	blk.init(fn, ID(len(fn.idBlocks)))

	fn.idBlocks = append(fn.idBlocks, blk)

	return blk
}

// InsertBlock inserts the block at the specific
// location in the list
func (fn *Func) InsertBlock(i int, blk *Block) {
	if blk.fn != fn {
		log.Panicf("inserting block %v from %v int another func %v not supported", blk, blk.fn, fn)
	}

	if i < 0 || i >= len(fn.blocks) {
		fn.blocks = append(fn.blocks, blk)
		return
	}

	fn.blocks = append(fn.blocks[:i+1], fn.blocks[i:]...)
	fn.blocks[i] = blk
}

// BlockIndex returns the index of the Block in the list
func (fn *Func) BlockIndex(blk *Block) int {
	for i, b := range fn.blocks {
		if b == blk {
			return i
		}
	}
	return -1
}

// RemoveBlock removes the Block from the list but
// does not remove it from succ/pred lists. See blk.Unlink()
func (fn *Func) RemoveBlock(blk *Block) {
	i := fn.BlockIndex(blk)

	fn.blocks = append(fn.blocks[:i], fn.blocks[i+1:]...)
}
