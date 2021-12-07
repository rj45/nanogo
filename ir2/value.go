package ir2

import (
	"go/types"
	"log"

	"github.com/rj45/nanogo/ir/reg"
)

func (val *Value) init(id ID, typ types.Type) {
	val.uses = val.usestorage[:0]
	val.ID = id
	val.typ = typ
}

func (val *Value) Func() *Func {
	return val.def.blk.fn
}

func (val *Value) Def() *Instr {
	return val.def
}

func (val *Value) NumUses() int {
	return len(val.uses)
}

func (val *Value) Use(i int) *Instr {
	return val.uses[i]
}

func (val *Value) addUse(instr *Instr) {
	val.uses = append(val.uses, instr)
}

func (val *Value) removeUse(instr *Instr) {
	index := -1
	for i, use := range val.uses {
		if use == instr {
			index = i
			break
		}
	}
	if index < 0 {
		log.Panicf("%v does not have use %v", val, instr)
	}
	val.uses = append(val.uses[:index], val.uses[index+1:]...)
}

func (val *Value) Loc() ValueLoc {
	return val.loc
}

func (val *Value) Reg() reg.Reg {
	if val.loc == VReg {
		return reg.Reg(val.index)
	}
	return reg.None
}

func (val *Value) Type() types.Type {
	return val.typ
}
