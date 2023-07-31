package op

//go:generate go run github.com/dmarkham/enumer -type=Op -transform title-lower

type def uint16

const (
	sink def = 1 << iota
	compare
	constant
	move
	commute
	branch
)

type Op uint8

func (op Op) IsCompare() bool {
	return opDefs[op]&compare != 0
}

func (op Op) IsSink() bool {
	return opDefs[op]&sink != 0
}

func (op Op) IsConst() bool {
	return opDefs[op]&constant != 0
}

func (op Op) IsCopy() bool {
	return opDefs[op]&move != 0
}

func (op Op) IsCommutative() bool {
	return opDefs[op]&commute != 0
}

func (op Op) IsCall() bool {
	return op == Call
}

func (op Op) ClobbersArg() bool {
	return false
}

func (op Op) IsBranch() bool {
	return opDefs[op]&branch != 0
}

func (op Op) Opposite() Op {
	switch op {
	case Equal:
		return NotEqual
	case NotEqual:
		return Equal
	case Less:
		return GreaterEqual
	case LessEqual:
		return Greater
	case Greater:
		return LessEqual
	case GreaterEqual:
		return Less
	}
	return op
}

const (
	Invalid Op = iota
	Builtin
	Call
	CallBuiltin
	ChangeInterface
	ChangeType
	Const
	Convert
	Copy
	Extract
	Field
	FieldAddr
	FreeVar
	Func
	Global
	Index
	IndexAddr
	InlineAsm
	Local
	Lookup
	MakeInterface
	MakeSlice
	Next
	New
	Parameter
	Range
	Reg
	Slice
	SliceToArrayPointer
	Store
	SwapIn
	SwapOut
	TypeAssert
	Add
	Sub
	Mul
	Div
	Rem
	And
	Or
	Xor
	ShiftLeft
	ShiftRight
	AndNot
	Equal
	NotEqual
	Less
	LessEqual
	Greater
	GreaterEqual
	Not
	Negate
	Load
	Invert

	// control flow instrs

	Jump
	If
	Return
	Panic

	IfEqual
	IfNotEqual
	IfLess
	IfLessEqual
	IfGreater
	IfGreaterEqual

	NumOps
)

var opDefs = [...]def{
	Builtin:      constant,
	Const:        constant,
	Copy:         move,
	Func:         constant,
	Global:       constant,
	Reg:          move,
	Store:        sink,
	Add:          commute,
	Mul:          commute,
	And:          commute,
	Or:           commute,
	Xor:          commute,
	Equal:        compare | commute,
	NotEqual:     compare | commute,
	Less:         compare,
	LessEqual:    compare,
	Greater:      compare,
	GreaterEqual: compare,
	NumOps:       0, // make sure array is large enough
}
