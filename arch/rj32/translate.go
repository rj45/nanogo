package rj32

import (
	"go/types"
	"log"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
)

var directTranslate = map[op.Op]Opcode{
	op.Return: Return,
	op.Jump:   Jump,
}

var twoOperandTranslations = map[op.Op]Opcode{
	op.Add:       Add,
	op.Sub:       Sub,
	op.And:       And,
	op.Or:        Or,
	op.Xor:       Xor,
	op.ShiftLeft: Shl,
}

var oneOperandTranslations = map[op.Op]Opcode{
	op.Invert: Not,
	op.Negate: Neg,
}

var branchesSigned = map[op.Op]Opcode{
	op.Equal:        IfEq,
	op.NotEqual:     IfNe,
	op.Less:         IfLt,
	op.LessEqual:    IfLe,
	op.Greater:      IfGt,
	op.GreaterEqual: IfGe,
}

var branchesUnsigned = map[op.Op]Opcode{
	op.Equal:        IfEq,
	op.NotEqual:     IfNe,
	op.Less:         IfUlt,
	op.LessEqual:    IfUle,
	op.Greater:      IfUgt,
	op.GreaterEqual: IfUge,
}

func translate(it ir2.Iter) {
	instr := it.Instr()
	originalOp := instr.Op
	switch instr.Op {
	case op.Copy:
		// copy is done in the finishing stage, after register allocation
	case op.Return, op.Jump:
		it.Update(directTranslate[instr.Op.(op.Op)], nil, instr.Args())
	case op.Add, op.Sub, op.And, op.Or, op.Xor, op.ShiftLeft:
		if instr.NumArgs() == 2 && instr.Arg(0).Reg() == instr.Def(0).Reg() {
			it.Update(twoOperandTranslations[instr.Op.(op.Op)], nil, instr.Args())
		}
	case op.ShiftRight:
		op := Shr
		typ := instr.Def(0).Type.Underlying()
		if basic, ok := typ.(*types.Basic); ok {
			// is signed integer?
			if basic.Info()&(types.IsInteger|types.IsUnsigned) == types.IsInteger {
				op = Asr
			}
		}
		if instr.NumArgs() == 2 && instr.Arg(0).Reg() == instr.Def(0).Reg() {
			it.Update(op, nil, instr.Args())
		}
	case op.Not, op.Negate:
		if instr.NumArgs() == 1 && instr.Arg(0).Reg() == instr.Def(0).Reg() {
			it.Update(oneOperandTranslations[instr.Op.(op.Op)], nil, instr.Args())
		}
	case op.Equal, op.NotEqual, op.Less, op.LessEqual, op.Greater, op.GreaterEqual:
		def := instr.Def(0)
		if def.NumUses() > 1 || def.Use(0).Instr().Op != op.If {
			log.Panicf("Lone comparison not tied to If %s", instr.LongString())
		}
	case op.If:
		compare := instr.Arg(0).Def().Instr()
		if !compare.IsCompare() {
			log.Panicf("expecting if to have compare, but instead had: %s", compare.LongString())
		}

		typ := compare.Arg(0).Type.Underlying()
		branchOp := branchesSigned[compare.Op.(op.Op)]
		if basic, ok := typ.(*types.Basic); ok {
			// is signed integer?
			if basic.Info()&(types.IsInteger|types.IsUnsigned) == types.IsInteger {
				branchOp = branchesUnsigned[compare.Op.(op.Op)]
			}
		}
		if branchOp == 0 {
			log.Panicf("failed to translate compare %s", compare.Op.(op.Op))
		}
		it.Update(branchOp, nil, compare.Args())
		if compare.Def(0).NumUses() == 0 {
			it.RemoveInstr(compare)
		}
	case op.Load:
		it.Update(Load, instr.Def(0).Type, instr.Args())
	case op.Store:
		it.Update(Store, nil, instr.Args())
	case op.Call:
		instr.Op = Call
		it.Changed()
	default:
		// if _, ok := instr.Op.(op.Op); ok {
		// 	log.Panicf("Unknown instruction: %s", instr.LongString())
		// }
	}
	if it.Instr() == nil {
		log.Panicf("translating %s from %s left iter in bad state", originalOp, instr.LongString())
	}
}

func translateCopies(it ir2.Iter) {
	instr := it.Instr()

	it.Update(Move, instr.Def(0).Type, instr.Args())
}
