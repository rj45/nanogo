// experimental simplified type system
package typ

type TypeKind uint8

//go:generate enumer -type=TypeKind -transform lower -output types_enumer.go

type Type interface {
	Kind() TypeKind
	String() string
	Words() int
	Bytes() int
}

const (
	Invalid TypeKind = iota

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

	// composite types
	Ptr // pointer
	Func
	Slice
	Array
	Struct
	Map
	Interface
	Chan

	NumTypes

	// aliases
	Byte = U8
	Rune = I32
)

func (tk TypeKind) Kind() TypeKind { return tk }
