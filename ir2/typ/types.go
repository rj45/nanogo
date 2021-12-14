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

import "fmt"

var ctx *Context

type Type uint16

const Unknown Type = 0

const typeBits = 16
const extraInfoShift = 10

func (t Type) Kind() Kind {
	if t&1 == 0 {
		return Kind(t>>1) & 0b11111
	}
	return Kind((t>>1)&7) + firstDecoratorType
}

func (t Type) isExtended() bool {
	if t&1 == 0 {
		return t > Type(NumTypes)
	}
	return t&0b10000 != 0
}

func (t Type) isSimple() bool {
	if t&1 == 0 {
		return t < Type(NumTypes)
	}
	return t&0b10000 == 0
}

func (t Type) index() int {
	if t&1 == 0 {
		return int(t >> 6)
	}
	if t&0b10000 != 0 {
		return int(t >> 5)
	}
	return -1
}

func (t Type) Elem() Type {
	if t&1 == 0 {
		return (t >> 1) & 0b11111
	}
	if t&0b10000 == 0 {
		return t >> 4
	}
	return ctx.Elem(t)
}

func (t Type) Dir() ChanDir {
	if t.Kind() == Chan {
		if t.isSimple() {
			return ChanDir(t >> extraInfoShift)
		}
	}
	return ctx.Dir(t)
}

func (t Type) Len() int {
	if t.Kind() == Array {
		if t.isSimple() {
			return int(t >> extraInfoShift)
		}
	}
	return ctx.Len(t)
}

func (t Type) Key() Type {
	if t.Kind() == Map {
		if t.isSimple() {
			return Type(t >> extraInfoShift)
		}
	}
	return ctx.Key(t)
}

func (t Type) String() string {
	switch t.Kind() {
	case Chan:
		switch t.Dir() {
		case SendRecv:
			return "chan " + t.Elem().String()
		case RecvOnly:
			return "<-chan " + t.Elem().String()
		case SendOnly:
			return "chan<- " + t.Elem().String()
		}
	case Slice:
		return "[]" + t.Elem().String()
	case Ptr:
		return "*" + t.Elem().String()
	case Array:
		return fmt.Sprintf("[%d]%s", t.Len(), t.Elem().String())
	case Map:
		return fmt.Sprintf("map[%s]%s", t.Key().String(), t.Elem().String())
	}
	return t.Kind().String()
}
