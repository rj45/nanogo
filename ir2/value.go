package ir2

import (
	"go/types"
	"log"

	"github.com/rj45/nanogo/ir/reg"
)

func (val *Value) init(id ID, typ types.Type) {
	val.uses = val.usestorage[:0]
	val.ID = id
	val.Type = typ
}

// Func returns the containing Func
func (val *Value) Func() *Func {
	return val.def.blk.fn
}

// Def returns the Instr defining the Value,
// or nil if it's not defined
func (val *Value) Def() *Instr {
	return val.def
}

// NumUses returns the number of uses
func (val *Value) NumUses() int {
	return len(val.uses)
}

// Use returns the ith Instr using this Value
func (val *Value) Use(i int) *Instr {
	return val.uses[i]
}

// ReplaceUsesWith will go through each use of
// val and replace it with other. Does not modify
// any definitions.
func (val *Value) ReplaceUsesWith(other *Value) {
	tries := 0
	for len(val.uses) > 0 {
		tries++
		use := val.uses[len(val.uses)-1]
		if tries > 1000 {
			log.Panicln("bug in uses ", val, other)
		}
		i := use.ArgIndex(val)
		if i < 0 {
			panic("couldn't find use!")
		}
		use.ReplaceArg(i, other)
	}
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

// Reg returns the register for the Value
// or reg.None if it's not located in a register
func (val *Value) Reg() reg.Reg {
	if val.Loc == VReg {
		return reg.Reg(val.Index)
	}
	return reg.None
}
