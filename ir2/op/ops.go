package op

//go:generate go run github.com/dmarkham/enumer -type=Op -transform title-lower


type def struct {
	Sink    bool
	Compare bool
	Const   bool
	ClobArg bool
	Copy    bool
	Commute bool
}

type Op uint8

func (op Op) IsCompare() bool {
	return opDefs[op].Compare
}

func (op Op) IsSink() bool {
	return opDefs[op].Sink
}

func (op Op) IsConst() bool {
	return opDefs[op].Const
}

func (op Op) IsCopy() bool {
	return opDefs[op].Copy
}

func (op Op) IsCommutative() bool {
	return opDefs[op].Commute
}

func (op Op) IsCall() bool {
	return op == Call
}

func (op Op) ClobbersArg() bool {
	return opDefs[op].ClobArg && twoOperand
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
	Builtin:      {Const: true},
	Const:        {Const: true},
	Copy:         {Copy: true},
	Func:         {Const: true},
	Global:       {Const: true},
	Reg:          {Copy: true},
	Store:        {Sink: true},
	Add:          {ClobArg: true, Commute: true},
	Sub:          {ClobArg: true},
	Mul:          {Commute: true},
	And:          {ClobArg: true, Commute: true},
	Or:           {ClobArg: true, Commute: true},
	Xor:          {ClobArg: true, Commute: true},
	ShiftLeft:    {ClobArg: true},
	ShiftRight:   {ClobArg: true},
	Equal:        {Compare: true, Commute: true},
	NotEqual:     {Compare: true, Commute: true},
	Less:         {Compare: true},
	LessEqual:    {Compare: true},
	Greater:      {Compare: true},
	GreaterEqual: {Compare: true},
	Negate:       {ClobArg: true},
	Invert:       {ClobArg: true},
	NumOps:       {}, // make sure array is large enough
}

type Arch interface {
	IsTwoOperand() bool
}

func SetArch(a Arch) {
	twoOperand = a.IsTwoOperand()
}

var twoOperand bool
