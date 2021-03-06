// Package ir2 contains an Intermediate Representation (IR) in
// Static Single Assignment (SSA) form.
//
//   - Each Program contains a list of Packages.
//   - Packages are a list of Globals, TypeDefs and Funcs.
//   - Each Func is a list of Blocks.
//   - Blocks have a list of Instrs.
//   - Instrs Def (define) Values, and have Values as Args.
//   - Values can be constants, types, temps, registers, memory locations, etc.
//
// Note: Unlike other SSA representations, this representation
// separates the concept of instructions from the concept of
// values. This allows an instruction to define multiple values.
// This is handy to avoid needing tuples and unpacking tuples to
// handle instructions (like function calls) that return multiple
// values.
//
package ir2

import (
	"go/token"
	"go/types"
)

type ObjectKind uint8

const (
	UnknownObject ObjectKind = iota
	BlockObject
	InstrObject
	ValueObject
	ValuePlaceholder
)

// ident is an identifier that's unique within a Func
type ident uint32

// Placeholder is an invalid ID meant to signal a place that needs to be filled
var Placeholder ident = idFor(ValuePlaceholder, -1)

// Program is a collection of packages,
// which comprise a whole program.
type Program struct {
	packages []*Package

	FileSet *token.FileSet

	takenNames map[string]bool
	strings    map[string]*Global
}

// Package is a collection of Funcs and Globals
// which comprise a part of a program.
type Package struct {
	prog *Program

	Type *types.Package

	Name string
	Path string

	funcs    []*Func
	globals  []*Global
	typedefs []*TypeDef
}

// Global is a global variable or literal stored in memory
type Global struct {
	pkg *Package

	Name       string
	FullName   string
	Type       types.Type
	Referenced bool

	// initial value
	Value Const
}

// TypeDef is a type definition
type TypeDef struct {
	pkg *Package

	Name       string
	Referenced bool

	Type types.Type
}

// Func is a collection of Blocks, which comprise
// a function or method in a Program.
type Func struct {
	Name     string
	FullName string
	Sig      *types.Signature

	Referenced bool
	NumCalls   int

	numArgSlots   int
	numParamSlots int
	numSpillSlots int

	pkg *Package

	blocks []*Block

	consts map[Const]*Value

	// placeholders that need filling
	placeholders map[string]*Value

	// ID to node mappings
	idBlocks []*Block
	idValues []*Value
	idInstrs []*Instr

	// allocate in slabs so related
	// stuff is close together in memory
	blockslab []Block
	valueslab []Value
	instrslab []Instr
}

// User uses and defines Values. Blocks and
// Instrs are Users.
type User struct {
	ident
	fn *Func

	defs []*Value
	args []*Value

	defstorage [2]*Value
	argstorage [3]*Value
}

// Block is a collection of Instrs which is a basic block
// in a control flow graph. The last Instr of a block must
// be a control flow Instr. A block may begin with one or more
// Phi Instrs, and all Phis should be at the start of the Block.
// Blocks can have Preds and Succs for the blocks that
// come before or after in the control flow graph respectively.
type Block struct {
	User

	instrs []*Instr

	preds []*Block
	succs []*Block

	instrstorage [5]*Instr
	predstorage  [2]*Block
	succstorage  [2]*Block
}

// Op describes an operation (instruction) type
type Op interface {
	String() string
	IsCall() bool
	IsCompare() bool
	IsCopy() bool
	IsCommutative() bool
	IsSink() bool
}

// Instr is an instruction that may define one or more Values,
// and take as args (operands) one or more Values.
type Instr struct {
	User
	Op

	blk *Block

	Pos   token.Pos
	index int
}

// Location is the location of a Value
type Location uint8

const (
	InTemp Location = iota
	InConst
	InReg

	// Param slots are an area of the stack for func parameters.
	// Specifically, they are in the caller's arg slot area.
	InParamSlot

	// Arg slots are an area of the stack reserved for call arguments.
	InArgSlot

	// Spill slots are an area of the stack reserved for register spills.
	InSpillSlot
)

// Value is a single value that may be stored in a
// single place. This may be a constant or variable,
// stored in a temp, register or on the stack.
type Value struct {
	stg

	ident

	// Type is the type of the Value
	Type types.Type

	def  *User
	uses []*User

	usestorage [2]*User
}

// ConstKind is a kind of constant
type ConstKind uint8

const (
	// no const, or not a const
	NotConst ConstKind = iota

	// nil, which is different than no const at all
	NilConst

	// non-numeric values
	BoolConst
	StringConst

	// numeric values
	IntConst

	// funcs and globals
	FuncConst
	GlobalConst
)

// Const is a constant value of some sort
type Const interface {
	Location() Location
	Kind() ConstKind
	String() string
	private()
}
