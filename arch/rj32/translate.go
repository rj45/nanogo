package rj32

import (
	"go/types"

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
	switch instr.Op {
	// case op.Copy:
	// 	if instr.NumArgs() == 1 {
	// 		it.Update(Move, nil, instr.Args())
	// 	}
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
	case op.If:
		compare := instr.Arg(0).Def().Instr()
		typ := compare.Arg(0).Type.Underlying()
		branchOp := branchesSigned[compare.Op.(op.Op)]
		if basic, ok := typ.(*types.Basic); ok {
			// is signed integer?
			if basic.Info()&(types.IsInteger|types.IsUnsigned) == types.IsInteger {
				branchOp = branchesUnsigned[compare.Op.(op.Op)]
			}
		}
		instr.Update(branchOp, nil, compare.Args())
		if compare.Def(0).NumUses() == 0 {
			compare.Block().RemoveInstr(compare)
		}
	}
}
