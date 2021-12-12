// experimental simplified type system
//
// Defintions:
// - A `byte` is the smallest addressable unit in the CPU, this does not
//   need to be 8 bits, though it usually is. In rj32 it's 16 bits.
// - A `word` is the largest value that can fit in a register. This can
//   be the the same as the word size, as it is in rj32 (16 bits)
// - A `pointer` can be larger than 1 word, if it is, generally an `int`
//   will also be larger than 1 word. This is common on 8-bit systems.
// - In general, though, you will want `int` and `uint` to be the same
//   size as a word, but at least 16 bits.
package typ

type Kind uint8

//go:generate enumer -type=Kind -transform lower -output types_enumer.go

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

const Unknown Type = 0

type Context struct {
	types  []info
	arrays []arrayInfo
	elems  []Type
	maps   []mapInfo
	fields [][]Type
	funcs  []funcInfo

	sizes     [17]uint8
	wordBytes uint8
}

type info struct {
	kind Kind

	extra uint16
}

type arrayInfo struct {
	len  int
	elem Type
}

type mapInfo struct {
	key  Type
	elem Type
}

type funcInfo struct {
	receiver Type
	results  []Type
	params   []Type
}

// fieldInfo represents a field, method, or func result/param
type fieldInfo struct {
	name string
	typ  Type
}

func (c *Context) TypeKind(typ Type) Kind {
	return c.types[typ].kind
}

func (c *Context) Bytes(typ Type) int {
	// todo: calc sizes of composite types
	return int(c.sizes[typ])
}

func (c *Context) Words(typ Type) int {
	return int(c.sizes[typ] / c.wordBytes)
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
		return Unknown
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
		return Unknown
	}
	return c.funcs[c.types[typ].extra].receiver
}

func (c *Context) Results(typ Type) []Type {
	if c.types[typ].kind != Func {
		return nil
	}
	return c.funcs[c.types[typ].extra].results
}

func (c *Context) Params(typ Type) []Type {
	if c.types[typ].kind != Func {
		return nil
	}
	return c.funcs[c.types[typ].extra].params
}
