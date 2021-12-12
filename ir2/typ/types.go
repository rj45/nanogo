// experimental simplified type system
package typ

type TypeKind uint8

//go:generate enumer -type=TypeKind -transform lower -output types_enumer.go

// type Type interface {
// 	Kind() TypeKind
// 	String() string
// 	Words() int
// 	Bytes() int
// }

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

type Type uint16

type Context struct {
	types []info
	arrays []arrayInfo
	elems []Type
	maps []mapInfo
	fields [][]Type
	funcs []funcInfo
	
	sizes [17]uint8
	wordBytes uint8
}

type info struct {
	kind TypeKind
	
	extra uint16
}

type arrayInfo struct{
	len int
	elem Type
}

type mapInfo struct{
	key Type
	elem Type
}

type funcInfo struct{
	receiver Type
	results []Type
	params []Type
}

func (c *Context) TypeKind(typ Type) TypeKind {
	return c.types[typ].kind
}

func (c *Context) Bytes(typ Type) int {
	// todo: calc sizes of composite types
	return int(c.sizes[typ])
}

func (c *Context) Words(typ Type) int {
	return int(c.sizes[typ]/c.wordBytes)
}

func (c *Context) Elem(typ Type) Type {
	switch c.types[typ].kind {
	case Array:
		return c.arrays[c.types[typ].extra].elem
	case Slice, Ptr:
		return c.elems[c.types[typ].extra]
	case Map:
		return c.maps[c.types[typ].extra].elem
	}
	return typ
}

func (c *Context) Key(typ Type) Type {
	if c.types[typ].kind != Map {
		return Invalid
	}
	return c.maps[c.types[typ].extra].key
}


func (c *Context) ArrayLen(typ Type) int {
	if c.types[typ].kind != Array {
		return 0
	}
	return c.arrays[c.types[typ].extra].len
}

func (c *Context) Fields(typ Type) []Type {
	if c.types[typ].kind != Struct {
		return nil
	}
	return c.fields[c.types[typ].extra]
}

func (c *Context) Receiver(typ Type) Type {
	if c.types[typ].kind != Func {
		return Invalid
	}
	return c.funcs[c.types[typ].extra].receiver
}

func (c *Context) Results(typ Type) []Type {
	if c.types[typ].kind != Func {
		return Invalid
	}
	return c.funcs[c.types[typ].extra].results
}

func (c *Context) Args(typ Type) []Type {
	if c.types[typ].kind != Func {
		return Invalid
	}
	return c.funcs[c.types[typ].extra].args
}