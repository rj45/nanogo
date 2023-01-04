package ir2

import (
	"fmt"
	"go/constant"
)

// Const is a constant value of some sort
type Const interface {
	Location() Location
	Kind() ConstKind
	String() string
	private()
}

// ConstKind is a kind of constant
type ConstKind uint8

const (
	// no const, or not a const
	NotConst ConstKind = iota

	// nil, which is different than no const at all
	NilConst

	// non-numeric values
	BoolConst
	StringConst

	// numeric values
	IntConst

	// funcs and globals
	FuncConst
	GlobalConst
)

type (
	notConst    struct{}
	nilConst    struct{}
	boolConst   bool
	stringConst string
	intConst    int64
	funcConst   struct{ fn *Func }
	globalConst struct{ glob *Global }
)

func (notConst) Location() Location { return InConst }
func (notConst) Kind() ConstKind    { return NotConst }
func (notConst) String() string     { return "<unk>" }
func (notConst) private()           {}

func (nilConst) Location() Location { return InConst }
func (nilConst) Kind() ConstKind    { return NilConst }
func (nilConst) String() string     { return "nil" }
func (nilConst) private()           {}

func (c boolConst) Location() Location { return InConst }
func (c boolConst) Kind() ConstKind    { return BoolConst }
func (c boolConst) String() string     { return fmt.Sprintf("%v", bool(c)) }
func (c boolConst) private()           {}

func (c stringConst) Location() Location { return InConst }
func (c stringConst) Kind() ConstKind    { return StringConst }
func (c stringConst) String() string     { return string(c) }
func (c stringConst) private()           {}

func (c intConst) Location() Location { return InConst }
func (c intConst) Kind() ConstKind    { return IntConst }
func (c intConst) String() string     { return fmt.Sprintf("%d", int64(c)) }
func (c intConst) private()           {}

func (c funcConst) Location() Location { return InConst }
func (c funcConst) Kind() ConstKind    { return FuncConst }
func (c funcConst) String() string     { return c.fn.FullName }
func (c funcConst) private()           {}

func (c globalConst) Location() Location { return InConst }
func (c globalConst) Kind() ConstKind    { return GlobalConst }
func (c globalConst) String() string     { return c.glob.FullName }
func (c globalConst) private()           {}

// Return a Const for a value
func ConstFor(v interface{}) Const {
	switch v := v.(type) {
	case bool:
		return boolConst(v)
	case string:
		return stringConst(v)
	case int:
		return intConst(v)
	case int32:
		return intConst(v)
	case int64:
		return intConst(v)
	case *Func:
		return funcConst{v}
	case *Global:
		return globalConst{v}
	case constant.Value:
		// convert constants
		switch v.Kind() {
		case constant.Bool:
			return boolConst(constant.BoolVal(v))
		case constant.String:
			return stringConst(constant.StringVal(v))
		case constant.Int:
			if i, ok := constant.Int64Val(v); ok {
				return intConst(i)
			}
			return notConst{}

		default:
			return notConst{}
		}
	case Const:
		return v
	}
	if v == nil {
		return nilConst{}
	}
	return notConst{}
}

// BoolValue returns a bool for a BoolConst
func BoolValue(c Const) (bool, bool) {
	if c.Kind() != BoolConst {
		return false, false
	}
	return bool(c.(boolConst)), true
}

// StringValue return a string for a StringConst
func StringValue(c Const) (string, bool) {
	if c.Kind() != StringConst {
		return "", false
	}
	return string(c.(stringConst)), true
}

// IntValue returns an int for an IntConst
func IntValue(c Const) (int, bool) {
	if c.Kind() != IntConst {
		return 0, false
	}
	return int(c.(intConst)), true
}

// Int64Value returns an int64 for an IntConst
func Int64Value(c Const) (int64, bool) {
	if c.Kind() != IntConst {
		return 0, false
	}
	return int64(c.(intConst)), true
}

// FuncValue returns a *Func for a FuncConst
func FuncValue(c Const) (*Func, bool) {
	if c.Kind() != FuncConst {
		return nil, false
	}
	return c.(funcConst).fn, true
}

// GlobalValue returns a *Func for a GlobalConst
func GlobalValue(c Const) (*Global, bool) {
	if c.Kind() != GlobalConst {
		return nil, false
	}
	return c.(globalConst).glob, true
}
