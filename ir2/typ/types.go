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
//
// TinyGo uses a system to compactly represent types in a way that
// can be embedded for runtime reflection. We use a very similar system
// here to optimize code reuse.
//
// There are less than 32 basic types (19? plus some compiler specific
// ones) which can fit in 5 bits. These are the BasicTypes in the types
// package.
//
// Extended types come in 8 other flavours. 4 of those flavours just
// decorate another type (chan interface pointer slice). 4 of them
// are more complex (func, array, map, struct).
//
// As much as possible is packed into the Type code itself including
// all the basic types as well as decorated basic types, as long as
// none of them are named types. After that, the Type simply contains
// an ID for further information to be looked up in side tables, and
// it's fundamental Kind.
//
// T = type kind
// K = map key type kind, chan direction, array size
// D = type decoration
// I = type index in side tables
//
// An unnamed basic type, inc: empty struct, empty interface, `func()`
// 0000 0000 00TT TTT0
//
// A named basic type
// IIII IIII IITT TTT0
//
// An unnamed decorated basic type
//   - maps with basic types as both key and value have both specified
//   - channels have their direction specified as K
//   - if an array, and the array length fits in K bits
// KKKK KKTT TTT0 DDD1
//
// A named/extended decorated type
// IIII IIII III1 DDD1
//
package typ

var ctx *Context

type Type uint16

func (t Type) Kind() Kind {
	if t&1 == 0 {
		return Kind(t >> 1)
	}
	return Kind((t>>1)&7) + Chan
}

func (t Type) isExtended() bool {
	if t&1 == 0 {
		return t > Type(NumTypes)
	}
	return t&0b10000 != 0
}

func (t Type) index() int {
	if t&1 == 0 {
		return int(t >> 6)
	}
	if t&0b10000 != 0 {
		return int(t >> 5)
	}
	return 0
}

func (t Type) elem() Type {
	if t&1 == 0 {
		return (t >> 1) & 0b11111
	}
	if t&0b10000 == 0 {
		return t >> 5
	}
	return ctx.Elem(t)
}

func (t Type) String() string {
	return ""
}

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
	Chan
	Interface
	Ptr // pointer
	Slice
	Array
	Func
	Map
	Struct

	NumTypes

	// aliases
	Byte = U8
	Rune = I32
)

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

// todo: finish this
// func (c *Context) TypeFor(typ types.Type) Type {
// 	switch t := typ.(type) {
// 	case *types.Basic:
// 		// basic types will not have side tables mapped to them unless
// 		// they are named/defined
// 		return Type(t.Kind()<<1)
// 	case *types.Slice, *types.Pointer, *types.Chan:
// 		e := typ.(interface{ Elem() types.Type }).Elem()
// 		if _, isBasic := e.(*types.Basic); isBasic {
// 			// a slice/pointer to a basic type also doesn't need a side
// 			// table
// 			return nil
// 		}
// 	case *types.Map:
// 		_, basicKey := t.Key().(*types.Basic)
// 		_, basicValue := t.Elem().(*types.Basic)
// 		if basicKey && basicValue {
// 			// a slice/pointer to a basic type also doesn't need a side
// 			// table
// 			return nil
// 		}
// 	}
// }

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
