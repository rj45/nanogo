package typ

import (
	"go/types"
	"math/bits"
)

// SimpleTypeFor attempts to fit the types.Type into a Type. If
// side tables are required for more information, Unknown is
// returned.
func SimpleTypeFor(typ types.Type) Type {
	if bt := basicCodeFor(typ); bt != Unknown {
		return bt
	}
	switch t := typ.(type) {
	case *types.Slice:
		return decoratedCodeFor(typ, Slice)
	case *types.Pointer:
		return decoratedCodeFor(typ, Ptr)
	case *types.Chan:
		ct := decoratedCodeFor(typ, Chan)
		if ct != Unknown {
			return Type(t.Dir())<<extraInfoShift | ct
		}
	case *types.Array:
		at := decoratedCodeFor(typ, Array)
		if at != Unknown {
			if (64 - bits.LeadingZeros64(uint64(t.Len()))) <= (typeBits - extraInfoShift) {
				return Type(t.Len())<<extraInfoShift | at
			}
		}
	case *types.Map:
		bk := basicCodeFor(t.Key())
		ek := decoratedCodeFor(typ, Map)
		if bk != Unknown && ek != Unknown {
			return bk<<extraInfoShift | ek
		}
	}

	return Unknown
}

func basicCodeFor(typ types.Type) Type {
	switch t := typ.(type) {
	case *types.Basic:
		// basic types will not have side tables mapped to them unless
		// they are named/defined
		return Type(t.Kind() << 1)
	case *types.Interface:
		if t.Empty() {
			// empty interface needs no extra info
			return Type(Interface << 1)
		}
	case *types.Signature:
		if t.Recv() == nil &&
			t.Results().Len() == 0 &&
			t.Params().Len() == 0 {
			return Type(Func << 1)
		}
	case *types.Struct:
		if t.NumFields() == 0 {
			return Type(Struct << 1)
		}
	}
	return Unknown
}

func decoratedCodeFor(typ types.Type, kind Kind) Type {
	e := typ.(interface{ Elem() types.Type }).Elem()
	bt := basicCodeFor(e)
	if bt == Unknown {
		return Unknown
	}
	return bt<<4 | Type(kind-firstDecoratorType)<<1 | 1
}
