package typ

type ChanDir uint8

const (
	SendRecv ChanDir = iota
	SendOnly
	RecvOnly
)

type Kind uint8

//go:generate go run github.com/dmarkham/enumer -type=Kind -transform lower -output types_enumer.go

const (
	Invalid Kind = iota

	// the initial set of the following are in the same order as go/types package BasicKind list

	// basic types
	B // bool
	I // int
	I8
	I16
	I32
	I64
	U // uint
	U8
	U16
	U32
	U64
	Uptr // uintptr
	F32  // float
	F64
	Complex64
	Complex128
	Str       // string
	UnsafePtr // unsafe.Pointer

	// types for untyped (constant) values
	CB       // constant bool
	CI       // constant int
	CR       // constant rune
	CF       // constant float
	CComplex // constant complex
	CStr     // constant string
	CNil     // constant nil

	CSR // flags
	Mem // memory

	Blank // blank variable _

	// Note: the above needs to fit below 32 to fit in 5 bits, as well as Interface
	// Struct and Func below to handle the empty versions of those 3

	// composite types
	Interface
	Struct
	Func
	Chan
	Ptr // pointer
	Slice
	Array
	Map

	NumTypes

	// aliases
	Byte = U8
	Rune = I32
)

const firstDecoratorType = Interface
