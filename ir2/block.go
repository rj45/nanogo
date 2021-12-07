package ir2

import (
	"log"
)

func (blk *Block) init(fn *Func, id ID) {
	blk.fn = fn
	blk.ID = id
	blk.instrs = blk.instrstorage[:0]
	blk.preds = blk.predstorage[:0]
	blk.succs = blk.succstorage[:0]
}

func (blk *Block) Func() *Func {
	return blk.fn
}

func (blk *Block) NumPreds() int {
	return len(blk.preds)
}

func (blk *Block) Pred(i int) *Block {
	return blk.preds[i]
}

func (blk *Block) AddPred(pred *Block) {
	blk.preds = append(blk.preds, pred)
}

func (blk *Block) NumSuccs() int {
	return len(blk.succs)
}

func (blk *Block) Succ(i int) *Block {
	return blk.succs[i]
}

func (blk *Block) AddSucc(succ *Block) {
	blk.succs = append(blk.succs, succ)
}

func (blk *Block) NumInstrs() int {
	return len(blk.instrs)
}

func (blk *Block) Instr(i int) *Instr {
	return blk.instrs[i]
}

func (blk *Block) Control() *Instr {
	return blk.instrs[len(blk.instrs)-1]
}

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
