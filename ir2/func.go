package ir2

import (
	"fmt"
	"go/types"
	"log"
	"sort"
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
func (fn *Func) ValueFor(typ types.Type, v interface{}) *Value {
	switch v := v.(type) {
	case *Value:
		if v != nil {
			return v
		}
	case *Instr:
		if v != nil && len(v.defs) == 1 {
			return v.defs[0]
		}
	}

	con := ConstFor(v)
	if con.Kind() != UnknownConst {
		if conval, ok := fn.consts[con]; ok {
			return conval
		}
		conval := fn.NewValue(typ)
		conval.Const = con

		if fn.consts == nil {
			fn.consts = make(map[Const]*Value)
		}

		fn.consts[con] = conval

		return conval
	}

	panic(fmt.Sprintf("can't get value for %T %#v", v, v))
}

// Placeholders

// PlaceholderFor creates a special placeholder value that can be later
// resolved with a different value. This is useful for marking and
// resolving forward references.
func (fn *Func) PlaceholderFor(label string) *Value {
	ph, found := fn.placeholders[label]
	if found {
		return ph
	}

	ph = &Value{
		ID:    Placeholder,
		Const: ConstFor(label),
	}

	if fn.placeholders == nil {
		fn.placeholders = make(map[string]*Value)
	}

	fn.placeholders[label] = ph

	return ph
}

// HasPlaceholders returns whether there are unresolved placeholders or not
func (fn *Func) HasPlaceholders() bool {
	return len(fn.placeholders) > 0
}

// ResolvePlaceholder removes the placeholder from the list, replacing its
// uses with the specified value
func (fn *Func) ResolvePlaceholder(label string, value *Value) {
	if len(fn.placeholders[label].uses) < 1 {
		panic("resolving unused placeholder " + label)
	}

	fn.placeholders[label].ReplaceUsesWith(value)

	delete(fn.placeholders, label)
	if len(fn.placeholders) == 0 {
		fn.placeholders = nil
	}
}

// PlaceholderLabels returns a sorted list of placeholder labels
func (fn *Func) PlaceholderLabels() []string {
	labels := make([]string, len(fn.placeholders))
	i := 0
	for lab := range fn.placeholders {
		labels[i] = lab
		i++
	}

	// sort to make this deterministic since maps have random order
	sort.Strings(labels)

	return labels
}

// Instrs

// InstrForID returns the Instr for the ID
func (fn *Func) InstrForID(i ID) *Instr {
	return fn.idInstrs[i]
}

// NewInstr creates an unbound Instr
func (fn *Func) NewInstr(op Op, typ types.Type, args ...interface{}) *Instr {
	// allocate instrs in contiguous slabs in memory
	// to increase data locality
	if len(fn.instrslab) == cap(fn.instrslab) {
		fn.instrslab = make([]Instr, 0, instrSlabSize)
	}
	fn.instrslab = append(fn.instrslab, Instr{})
	instr := &fn.instrslab[len(fn.instrslab)-1]

	instr.init(ID(len(fn.idInstrs)))

	fn.idInstrs = append(fn.idInstrs, instr)

	instr.update(fn, op, typ, args)

	return instr
}

// Blocks

// NumBlocks returns the number of Blocks
func (fn *Func) NumBlocks() int {
	return len(fn.blocks)
}

// Block returns the ith Block
func (fn *Func) Block(i int) *Block {
	return fn.blocks[i]
}

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
