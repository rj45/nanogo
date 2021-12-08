package ir2

import (
	"go/types"

	"github.com/rj45/nanogo/ir/op"
)

/// in-block iterator

// BlockIter is an iterator that iterates over instructions in a Block
type BlockIter struct {
	blk    *Block
	insIdx int
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

// Next increments the position and returns whether that was successful
func (it *BlockIter) Next() bool {
	if it.insIdx >= len(it.blk.instrs) {
		return false
	}

	it.insIdx++
	return true
}

// Prev decrements the position and returns whether that was successful
func (it *BlockIter) Prev() bool {
	if it.insIdx == 0 {
		return false
	}

	it.insIdx--
	return true
}

// Insert inserts an instruction at the cursor position and increments the position
func (it *BlockIter) Insert(op op.Op, typ types.Type, args ...interface{}) *Instr {
	instr := it.blk.fn.NewInstr(op, typ, args...)

	it.blk.InsertInstr(it.insIdx, instr)
	it.Next()

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

	return instr
}

// Update updates the instruction at the cursor position
func (it *BlockIter) Update(op op.Op, typ types.Type, args ...interface{}) *Instr {
	instr := it.blk.instrs[it.insIdx]

	instr.Update(op, typ, args...)

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
	return &CrossBlockIter{
		BlockIter: BlockIter{blk: fn.blocks[0], insIdx: 0},
		fn:        fn,
		blkIdx:    0,
	}
}

// Next increments the position and returns whether that was successful
func (it *CrossBlockIter) Next() bool {
	if it.insIdx >= len(it.blk.instrs) {
		if (it.blkIdx + 1) >= len(it.fn.blocks) {
			return false
		}

		it.blkIdx++
		it.insIdx = 0
		it.blk = it.fn.blocks[it.blkIdx]
		return true
	}

	it.insIdx++
	return true
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
