package ir2

import (
	"go/types"
)

// Iter is a iterator over instructions
type Iter interface {
	// Instr returns the current instruction
	Instr() *Instr

	// InstrIndex returns the index of the current instruction in the Block
	InstrIndex() int

	// Block returns the current block
	Block() *Block

	// BlockIndex returns the index of the Block within the Func
	BlockIndex() int

	// HasNext returns whether Next() will succeed
	HasNext() bool

	// Next increments the position and returns whether that was successful
	Next() bool

	// HasPrev returns whether Prev() will succeed
	HasPrev() bool

	// Prev decrements the position and returns whether that was successful
	Prev() bool

	// Last fast forwards to the end
	Last() bool

	// Insert inserts an instruction at the cursor position and increments the position
	Insert(op Op, typ types.Type, args ...interface{}) *Instr

	// InsertAfter inserts after an instruction at the cursor position
	InsertAfter(op Op, typ types.Type, args ...interface{}) *Instr

	// Remove will remove the instruction at the current position and decrement the position,
	// returning the removed instruction.
	// NOTE: this only removes the instruction from the Block, it does not Unlink() it from
	// any uses.
	Remove() *Instr

	// RemoveInstr removes an instruction from anywhere and will adjust the iterator position
	// appropriately
	RemoveInstr(instr *Instr)

	// Update updates the instruction at the cursor position
	Update(op Op, typ types.Type, args ...interface{}) *Instr

	// HasChanged returns true if `Changed()` was called, or one of the mutation methods
	HasChanged() bool

	// Changed forces `HasChanged()` to return true
	Changed()
}

var _ Iter = &BlockIter{}
var _ Iter = &CrossBlockIter{}

/// in-block iterator

// BlockIter is an iterator that iterates over instructions in a Block
type BlockIter struct {
	blk     *Block
	insIdx  int
	changed bool
}

// InstrIter will return an Iter which iterates over every
// instruction in this block.
func (blk *Block) InstrIter() *BlockIter {
	return &BlockIter{
		blk:    blk,
		insIdx: 0,
	}
}

// Instr returns the current instruction
func (it *BlockIter) Instr() *Instr {
	if uint(it.insIdx) >= uint(len(it.blk.instrs)) {
		return nil
	}
	return it.blk.instrs[it.insIdx]
}

// Block returns the current block
func (it *BlockIter) Block() *Block {
	return it.blk
}

// InstrIndex returns the index of the current instruction in the Block
func (it *BlockIter) InstrIndex() int {
	return it.insIdx
}

// BlockIndex returns the index of the Block within the Func
func (it *BlockIter) BlockIndex() int {
	return it.blk.fn.BlockIndex(it.blk)
}

// HasNext returns whether Next() will succeed
func (it *BlockIter) HasNext() bool {
	return it.insIdx < len(it.blk.instrs)
}

// Next increments the position and returns whether that was successful
func (it *BlockIter) Next() bool {
	if it.insIdx >= len(it.blk.instrs) {
		return false
	}

	it.insIdx++
	return it.insIdx < len(it.blk.instrs)
}

// Prev decrements the position and returns whether that was successful
func (it *BlockIter) Prev() bool {
	if it.insIdx <= 0 {
		return false
	}

	it.insIdx--
	return true
}

// HasPrev returns whether Prev() will succeed
func (it *BlockIter) HasPrev() bool {
	return it.insIdx > 0 // todo: there is a bug here
}

// Last fast forwards to the end of the block
func (it *BlockIter) Last() bool {
	it.insIdx = len(it.blk.instrs) - 1
	return it.insIdx >= 0
}

// HasChanged returns true if `Changed()` was called, or one of the mutation methods
func (it *BlockIter) HasChanged() bool {
	return it.changed
}

// Changed forces `HasChanged()` to return true
func (it *BlockIter) Changed() {
	it.changed = true
}

// Insert inserts an instruction at the cursor position and increments the position
func (it *BlockIter) Insert(op Op, typ types.Type, args ...interface{}) *Instr {
	instr := it.blk.fn.NewInstr(op, typ, args...)

	it.blk.InsertInstr(it.insIdx, instr)
	it.Next()

	it.changed = true

	return instr
}

// InsertAfter inserts after an instruction at the cursor position
func (it *BlockIter) InsertAfter(op Op, typ types.Type, args ...interface{}) *Instr {
	instr := it.blk.fn.NewInstr(op, typ, args...)

	it.blk.InsertInstr(it.insIdx+1, instr)

	it.changed = true

	return instr
}

// Remove will remove the instruction at the current position and decrement the position,
// returning the removed instruction.
// NOTE: this only removes the instruction from the Block, it does not Unlink() it from
// any uses.
func (it *BlockIter) Remove() *Instr {
	instr := it.blk.instrs[it.insIdx]

	// todo: replace with RemoveInstrAt()?
	it.blk.RemoveInstr(instr)
	it.Prev()

	it.changed = true

	return instr
}

// RemoveInstr removes an instruction from the middle of the block somewhere, making
// sure to adjust the iterator position appropriately
func (it *BlockIter) RemoveInstr(instr *Instr) {
	it.changed = true

	index := instr.Index()
	if instr.blk != it.blk {
		// instruction is in a different block, defer to that one
		instr.blk.RemoveInstr(instr)
		return
	}

	// todo: replace with RemoveInstrAt()?
	it.blk.RemoveInstr(instr)
	if it.insIdx >= index {
		it.Prev()
	}
}

// Update updates the instruction at the cursor position
func (it *BlockIter) Update(op Op, typ types.Type, args ...interface{}) *Instr {
	instr := it.blk.instrs[it.insIdx]

	instr.Update(op, typ, args...)

	it.changed = true

	return instr
}

// inter-block iterator (ie, whole function)

type CrossBlockIter struct {
	BlockIter

	fn     *Func
	blkIdx int
}

// InstrIter returns an iterator that will iterate over every
// block and instruction in the func.
func (fn *Func) InstrIter() *CrossBlockIter {
	var blk *Block
	if len(fn.blocks) > 0 {
		blk = fn.blocks[0]
	}

	return &CrossBlockIter{
		BlockIter: BlockIter{blk: blk, insIdx: 0},
		fn:        fn,
		blkIdx:    0,
	}
}

// HasNext returns whether Next() will succeed
func (it *CrossBlockIter) HasNext() bool {
	if it.blk == nil {
		return false
	}

	if (it.blkIdx + 1) < len(it.fn.blocks) {
		return true
	}

	return (it.blkIdx + 1) < len(it.blk.instrs)
}

// Next increments the position and returns whether that was successful
func (it *CrossBlockIter) Next() bool {
	if !it.HasNext() {
		return false
	}

	it.insIdx++

	if it.insIdx >= len(it.blk.instrs) {
		it.blkIdx++
		it.insIdx = 0
		if it.blkIdx >= len(it.fn.blocks) {
			return false
		}
		it.blk = it.fn.blocks[it.blkIdx]
	}

	return true
}

// HasPrev returns whether Prev() will succeed
func (it *CrossBlockIter) HasPrev() bool {
	return it.insIdx >= 0 && it.blkIdx >= 0
}

// Prev decrements the position and returns whether that was successful
func (it *CrossBlockIter) Prev() bool {
	if it.insIdx == 0 {
		if it.blkIdx == 0 {
			return false
		}

		it.blkIdx--
		it.blk = it.fn.blocks[it.blkIdx]
		it.insIdx = len(it.blk.instrs) - 1

		return true
	}

	it.insIdx--
	return true
}

// Last fast forwards to the end of the func
func (it *CrossBlockIter) Last() bool {
	it.blkIdx = len(it.fn.blocks) - 1
	it.insIdx = len(it.blk.instrs) - 1
	return it.blkIdx >= 0 && it.insIdx >= 0
}
