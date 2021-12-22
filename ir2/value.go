package ir2

import (
	"go/types"
	"log"

	"github.com/rj45/nanogo/ir/reg"
)

// init initializes the Value
func (val *Value) init(id ID, typ types.Type) {
	val.uses = val.usestorage[:0]
	val.ID = id
	if typ != nil {
		if t, ok := typ.(*types.Basic); ok && t.Kind() == types.Invalid {
			typ = nil
		}
	}
	val.Type = validType(typ)
	val.SetTemp()
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

// stg is the storage for a value
type stg interface {
	Location() Location
}

// temps

type tempStg struct{}

func (tempStg) Location() Location { return InTemp }

func (val *Value) InTemp() bool {
	return val.Location() == InTemp
}

func (val *Value) Temp() ID {
	if val.Location() == InTemp {
		return val.ID
	}
	return Placeholder
}

func (val *Value) SetTemp() {
	val.stg = tempStg{}
}

// regs

type regStg struct{ r reg.Reg }

func (regStg) Location() Location { return InReg }

func (val *Value) InReg() bool {
	return val.Location() == InReg
}

func (val *Value) Reg() reg.Reg {
	if val.Location() == InReg {
		return val.stg.(regStg).r
	}
	return reg.None
}

func (val *Value) SetReg(reg reg.Reg) {
	val.stg = regStg{reg}
}

// param slots

type paramStg uint8

func (paramStg) Location() Location { return InParam }

func (val *Value) InParamSlot() bool {
	return val.Location() == InParam
}

func (val *Value) ParamSlot() int {
	if val.Location() == InParam {
		return int(val.stg.(paramStg))
	}
	return -1
}

func (val *Value) SetParamSlot(slot int) {
	val.stg = paramStg(slot)

	if val.Func().numParamSlots < slot+1 {
		val.Func().numParamSlots = slot + 1
	}
}

// arg slots

type argStg uint8

func (argStg) Location() Location { return InArg }

func (val *Value) InArgSlot() bool {
	return val.Location() == InArg
}

func (val *Value) ArgSlot() int {
	if val.Location() == InArg {
		return int(val.stg.(argStg))
	}
	return -1
}

func (val *Value) SetArgSlot(slot int) {
	val.stg = argStg(slot)

	if val.Func().numArgSlots < slot+1 {
		val.Func().numArgSlots = slot + 1
	}
}

// spill slots

type spillStg uint8

func (spillStg) Location() Location { return InSpill }

func (val *Value) InSpillSlot() bool {
	return val.Location() == InSpill
}

func (val *Value) SpillSlot() int {
	if val.Location() == InSpill {
		return int(val.stg.(spillStg))
	}
	return -1
}

func (val *Value) SetSpillSlot(slot int) {
	val.stg = spillStg(slot)

	if val.Func().numSpillSlots < slot+1 {
		val.Func().numSpillSlots = slot + 1
	}
}

// const

func (val *Value) Const() Const {
	if val.Location() == InConst {
		return val.stg.(Const)
	}
	return notConst{}
}

func (val *Value) SetConst(con Const) {
	val.stg = con
}

func (val *Value) IsConst() bool {
	return val.Location() == InConst
}

// util funcs

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

func validType(typ types.Type) types.Type {
	if typ != nil {
		if t, ok := typ.(*types.Basic); ok && t.Kind() == types.Invalid {
			return nil
		}
	}
	return typ
}
