// Package ir2 contains an Intermediate Representation (IR) in
// Static Single Assignment (SSA) form.
//
// - Each Program contains a list of Packages.
// - Packages are a list of Globals and Funcs.
// - Each Func is a list of Blocks.
// - Blocks have a list of Instrs.
// - Instrs Def (define) Values, and have Values as Args.
// - Values can be constants, types, temps, registers, memory locations, etc.
//
// The Go compiler does not have a concept of instructions,
// Values are also instructions. But then an operation can only
// return a single value, and tuples are used to get around that.
// In this model, there are no tuples, and operations can return
// multiple values. This simplfies multi-precision math, calls
// that can return multiple results, and PhiCopies / swaps that
// are defined to happen in parallel.
package ir2

import (
	"go/token"
	"go/types"
)

// Program is a collection of packages,
// which comprise a whole program.
type Program struct {
	packages []*Package

	takenNames map[string]bool
	strings    map[string]*Literal
}

// Literal is a piece of constant data stored with the program
type Literal struct {
	Name  string
	Value Const
}

// Package is a collection of Funcs and Globals
// which comprise a part of a program.
type Package struct {
	prog *Program

	Type *types.Package

	Name string
	Path string

	funcs   []*Func
	globals []*Global
}

// Global is a global variable or literal stored in memory
type Global struct {
	pkg *Package

	Name       string
	FullName   string
	Type       types.Type
	Referenced bool
}

// Func is a collection of Blocks, which comprise
// a function or method in a Program.
type Func struct {
	Name     string
	FullName string
	Sig      *types.Signature

	Referenced bool
	NumCalls   int

	pkg *Package

	blocks []*Block

	consts map[Const]*Value

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

// ID is an identifier that's unique within a Func
type ID uint

// Block is a collection of Instrs which is a basic block
// in a control flow graph. The last Instr of a block must
// be a control flow Instr. A block may begin with one or more
// Phi Instrs, and all Phis should be at the start of the Block.
// Blocks can have Preds and Succs for the blocks that
// come before or after in the control flow graph respectively.
type Block struct {
	ID
	fn *Func

	instrs []*Instr

	preds []*Block
	succs []*Block

	instrstorage [5]*Instr
	predstorage  [2]*Block
	succstorage  [2]*Block
}

type Op interface {
	String() string
	IsCompare() bool
	IsCopy() bool
	IsCommutative() bool
}

// Instr is an instruction that may define one or more Values,
// and take as Args (operands) one or more Values.
type Instr struct {
	ID
	Op

	blk *Block

	Pos   token.Pos
	index int

	defs []*Value
	args []*Value

	defstorage [2]*Value
	argstorage [3]*Value
}

// ValueLoc is the location of a Value
type ValueLoc uint8

const (
	Invalid ValueLoc = iota
	VConst
	VFunc
	VTemp
	VReg
	VStack
	VGlob
	VHeap
)

// Value is a single value that may be stored in a
// single place. This may be a constant or variable,
// stored in a temp, register or on the stack.
type Value struct {
	ID

	// Type is the type of the Value
	Type types.Type

	Const Const

	// Loc is the location of the Value
	Loc ValueLoc

	// Index is the index in the location
	Index int

	def  *Instr
	uses []*Instr

	usestorage [2]*Instr
}

type ConstKind uint8

const (
	UnknownConst ConstKind = iota

	NilConst

	// non-numeric values
	BoolConst
	StringConst

	// numeric values
	IntConst

	// funcs, globals and literals
	FuncConst
	GlobalConst
	LiteralConst
)

// Const is a constant value of some sort
type Const interface {
	Kind() ConstKind
	String() string
	private()
}
