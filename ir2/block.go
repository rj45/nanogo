package ir2

import (
	"log"
)

func (blk *Block) init(fn *Func, id ident) {
	blk.User.init(fn, id)
	blk.instrs = blk.instrstorage[:0]
	blk.preds = blk.predstorage[:0]
	blk.succs = blk.succstorage[:0]
}

// Func returns the containing Func
func (blk *Block) Func() *Func {
	return blk.fn
}

// NumPreds is the number of predecessors
func (blk *Block) NumPreds() int {
	return len(blk.preds)
}

// Pred returns the ith predecessor
func (blk *Block) Pred(i int) *Block {
	return blk.preds[i]
}

// AddPred adds the Block to the predecessor list
func (blk *Block) AddPred(pred *Block) {
	blk.preds = append(blk.preds, pred)
}

// NumSuccs returns the number of successors
func (blk *Block) NumSuccs() int {
	return len(blk.succs)
}

// Succ returns the ith successor
func (blk *Block) Succ(i int) *Block {
	return blk.succs[i]
}

// AddSucc adds the Block to the successor list
func (blk *Block) AddSucc(succ *Block) {
	blk.succs = append(blk.succs, succ)
}

// Unlink removes the Block from the pred/succ
// lists of surrounding Blocks
func (blk *Block) Unlink() {
	if len(blk.preds) == 1 && len(blk.succs) == 1 {
		replPred := blk.preds[0]
		replSucc := blk.succs[0]

		for j, pred := range replSucc.preds {
			if pred == blk {
				replSucc.preds[j] = replPred
			}
		}

		for j, succ := range replPred.succs {
			if succ == blk {
				replPred.succs[j] = replSucc
			}
		}
	} else {
		panic("can't remove block")
	}
}

// NumInstrs returns the number of instructions
func (blk *Block) NumInstrs() int {
	return len(blk.instrs)
}

// Instr returns the ith Instr in the list
func (blk *Block) Instr(i int) *Instr {
	return blk.instrs[i]
}

// Control returns the last instruction, which
// should be a control flow instruction
func (blk *Block) Control() *Instr {
	return blk.instrs[len(blk.instrs)-1]
}

// InsertInstr inserts the instruction at the ith
// position. -1 means append it.
func (blk *Block) InsertInstr(i int, instr *Instr) {
	if instr.blk != nil && instr.blk != blk {
		log.Panicf("remove instr %v from blk %v before inserting into %v", instr, instr.blk, blk)
	}

	instr.blk = blk

	if i < 0 || i >= len(blk.instrs) {
		instr.index = len(blk.instrs)
		blk.instrs = append(blk.instrs, instr)
		return
	}

	instr.index = i

	blk.instrs = append(blk.instrs[:i+1], blk.instrs[i:]...)
	blk.instrs[i] = instr

	for j := i + 1; j < len(blk.instrs); j++ {
		blk.instrs[j].index = j
	}
}

// SwapInstr swaps two instructions
func (blk *Block) SwapInstr(a *Instr, b *Instr) {
	i := a.Index()
	j := b.Index()

	blk.instrs[i], blk.instrs[j] = blk.instrs[j], blk.instrs[i]

	a.index = j
	b.index = i
}

// RemoveInstr removes the Instr from the list
func (blk *Block) RemoveInstr(inst *Instr) {
	i := inst.Index()
	if i < 0 {
		log.Panicf("already removed %v", inst)
	}

	inst.index = -1
	inst.blk = nil

	blk.instrs = append(blk.instrs[:i], blk.instrs[i+1:]...)

	for j := i; j < len(blk.instrs); j++ {
		blk.instrs[j].index = j
	}
}
