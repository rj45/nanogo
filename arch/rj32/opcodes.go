package rj32

import (
	"fmt"
	"strings"

	"github.com/rj45/nanogo/codegen/asm"
	"github.com/rj45/nanogo/ir/op"
)

type Opcode int

//go:generate go run github.com/dmarkham/enumer -type=Opcode -transform snake

const (
	// Natively implemented instructions
	Nop Opcode = iota
	Rets
	Error
	Halt
	Rcsr
	Wcsr
	Move
	Loadc
	Jump
	Imm
	Call
	Imm2
	Load
	Store
	Loadb
	Storeb
	Add
	Sub
	Addc
	Subc
	Xor
	And
	Or
	Shl
	Shr
	Asr
	IfEq
	IfNe
	IfLt
	IfGe
	IfUlt
	IfUge

	// Psuedoinstructions
	Not
	Neg
	Swap
	IfGt
	IfLe
	IfUgt
	IfUle
	Return

	NumOps
)

func (op Opcode) Asm() string {
	return strings.ReplaceAll(op.String(), "_", ".")
}

func (op Opcode) Fmt() asm.Fmt {
	return opDefs[op].fmt
}

func (op Opcode) IsMove() bool {
	return op == Move || op == Swap
}

func (op Opcode) IsCall() bool {
	return op == Call
}

func (op Opcode) IsCommutative() bool {
	return opDefs[op].flags&commutative != 0
}

func (op Opcode) IsCompare() bool {
	return opDefs[op].flags&compare != 0
}

func (op Opcode) IsBranch() bool {
	return opDefs[op].flags&compare != 0
}

func (op Opcode) IsCopy() bool {
	return op == Move
}

func (op Opcode) IsSink() bool {
	return opDefs[op].flags&sink != 0
}

func (op Opcode) ClobbersArg() bool {
	return opDefs[op].flags&clobbers != 0
}

type flags uint16

const (
	commutative flags = 1 << iota
	compare
	sink
	clobbers
)

type def struct {
	fmt   Fmt
	op    op.Op
	flags flags
}

var opDefs = [...]def{
	Nop:    {fmt: NoFmt, flags: sink},
	Rets:   {fmt: NoFmt, flags: sink},
	Error:  {fmt: NoFmt, flags: sink},
	Halt:   {fmt: NoFmt, flags: sink},
	Rcsr:   {fmt: BinaryFmt},
	Wcsr:   {fmt: BinaryFmt},
	Move:   {fmt: MoveFmt, op: op.Copy},
	Loadc:  {fmt: BinaryFmt},
	Jump:   {fmt: CallFmt, flags: sink},
	Imm:    {fmt: UnaryFmt},
	Call:   {fmt: CallFmt, op: op.Call},
	Imm2:   {fmt: BinaryFmt},
	Load:   {fmt: LoadFmt, op: op.Load},
	Store:  {fmt: StoreFmt, op: op.Store, flags: sink},
	Loadb:  {fmt: LoadFmt},
	Storeb: {fmt: StoreFmt, flags: sink},
	Add:    {fmt: BinaryFmt, op: op.Add, flags: clobbers},
	Sub:    {fmt: BinaryFmt, op: op.Sub, flags: clobbers},
	Addc:   {fmt: BinaryFmt, flags: clobbers},
	Subc:   {fmt: BinaryFmt, flags: clobbers},
	Xor:    {fmt: BinaryFmt, op: op.Xor, flags: clobbers},
	And:    {fmt: BinaryFmt, op: op.And, flags: clobbers},
	Or:     {fmt: BinaryFmt, op: op.Or, flags: clobbers},
	Shl:    {fmt: BinaryFmt, op: op.ShiftLeft, flags: clobbers},
	Shr:    {fmt: BinaryFmt, op: op.ShiftRight, flags: clobbers},
	Asr:    {fmt: BinaryFmt, flags: clobbers},
	IfEq:   {fmt: CompareFmt, flags: commutative | compare | sink},
	IfNe:   {fmt: CompareFmt, flags: commutative | compare | sink},
	IfLt:   {fmt: CompareFmt, flags: compare | sink},
	IfGe:   {fmt: CompareFmt, flags: compare | sink},
	IfUlt:  {fmt: CompareFmt, flags: compare | sink},
	IfUge:  {fmt: CompareFmt, flags: compare | sink},
	Not:    {fmt: UnaryFmt, op: op.Invert, flags: clobbers},
	Neg:    {fmt: UnaryFmt, op: op.Negate, flags: clobbers},
	Swap:   {fmt: CompareFmt, op: op.SwapIn},
	IfGt:   {fmt: CompareFmt, flags: compare | sink},
	IfLe:   {fmt: CompareFmt, flags: compare | sink},
	IfUgt:  {fmt: CompareFmt, flags: compare | sink},
	IfUle:  {fmt: CompareFmt, flags: compare | sink},
	Return: {fmt: NoFmt, flags: sink},
}

var translations []Opcode

func init() {
	translations = make([]Opcode, op.NumOps)
	for i := Nop; i < NumOps; i++ {
		if opDefs[i].fmt == BadFmt {
			panic(fmt.Sprintf("missing opDef for %s", i))
		}
		translations[opDefs[i].op] = i
	}
}

func (cpuArch) IsTwoOperand() bool {
	return true
}
