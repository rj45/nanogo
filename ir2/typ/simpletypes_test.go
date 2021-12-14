package typ_test

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/rj45/nanogo/ir2/typ"
)

func TestSimpleTypeFor_BasicTypes(t *testing.T) {
	for kind, basicType := range types.Typ {
		t.Run(fmt.Sprintf("%s has type %s", basicType, typ.Kind(kind)), func(t *testing.T) {
			got := typ.SimpleTypeFor(basicType)
			if got.Kind() != typ.Kind(kind) {
				t.Errorf("expected %s to translate to %s but got %s", basicType, typ.Kind(kind), got.Kind())
			}
		})
	}
}

func TestSimpleTypeFor_EmptyInterface(t *testing.T) {
	it := types.NewInterfaceType(nil, nil)
	got := typ.SimpleTypeFor(it)
	if got.Kind() != typ.Interface {
		t.Errorf("expected %s to translate to %s but got %s", it, typ.Interface, got.Kind())
	}
}

func TestSimpleTypeFor_EmptyInterfaceSlice(t *testing.T) {
	it := types.NewSlice(types.NewInterfaceType(nil, nil))
	got := typ.SimpleTypeFor(it)
	if got.Kind() != typ.Slice {
		t.Errorf("expected %s to translate to %s but got %s", it, typ.Slice, got.Kind())
	}
	if got.Elem().Kind() != typ.Interface {
		t.Errorf("expected %s to translate to %s but got %s", it, typ.Interface, got.Elem().Kind())
	}
}

func TestSimpleTypeFor_EmptyFunc(t *testing.T) {
	it := types.NewSignature(nil, nil, nil, false)
	got := typ.SimpleTypeFor(it)
	if got.Kind() != typ.Func {
		t.Errorf("expected %s to translate to %s but got %s", it, typ.Func, got)
	}
}

func TestSimpleTypeFor_EmptyStruct(t *testing.T) {
	it := types.NewStruct(nil, nil)
	got := typ.SimpleTypeFor(it)
	if got.Kind() != typ.Struct {
		t.Errorf("expected %s to translate to %s but got %s", it, typ.Struct, got)
	}
}

func TestSimpleTypeFor_BasicTypeSlices(t *testing.T) {
	for kind, basicType := range types.Typ {
		if kind == int(types.Invalid) {
			continue
		}

		t.Run(fmt.Sprintf("%s has type %s", basicType, typ.Kind(kind)), func(t *testing.T) {
			typesType := types.NewSlice(basicType)
			got := typ.SimpleTypeFor(typesType)
			if got.Kind() != typ.Slice {
				t.Errorf("expected %s to translate to %s but got %s", typesType, typ.Slice, got.Kind())
			}
			if got.Elem().Kind() != typ.Kind(kind) {
				t.Errorf("expected %s to have element %s but got %s", typesType, typ.Kind(kind), got.Elem().Kind())
			}
		})
	}
}

func TestSimpleTypeFor_BasicTypePointers(t *testing.T) {
	for kind, basicType := range types.Typ {
		if kind == int(types.Invalid) {
			continue
		}

		t.Run(fmt.Sprintf("%s has type %s", basicType, typ.Kind(kind)), func(t *testing.T) {
			typesType := types.NewPointer(basicType)
			got := typ.SimpleTypeFor(typesType)
			if got.Kind() != typ.Ptr {
				t.Errorf("expected %s to translate to %s but got %s", typesType, typ.Ptr, got.Kind())
			}
			if got.Elem().Kind() != typ.Kind(kind) {
				t.Errorf("expected %s to have element %s but got %s", typesType, typ.Kind(kind), got.Elem().Kind())
			}
		})
	}
}

func TestSimpleTypeFor_BasicTypeChans(t *testing.T) {
	for kind, basicType := range types.Typ {
		if kind == int(types.Invalid) {
			continue
		}
		for _, dir := range []types.ChanDir{types.SendRecv, types.RecvOnly, types.SendOnly} {
			t.Run(fmt.Sprintf("%s has type %s", basicType, typ.Kind(kind)), func(t *testing.T) {
				typesType := types.NewChan(dir, basicType)
				got := typ.SimpleTypeFor(typesType)
				if got.Kind() != typ.Chan {
					t.Errorf("expected %s to translate to %s but got %s", typesType, typ.Chan, got.Kind())
				}
				if got.Elem().Kind() != typ.Kind(kind) {
					t.Errorf("expected %s to have element %s but got %s", typesType, typ.Kind(kind), got.Elem().Kind())
				}
				if got.Dir() != typ.ChanDir(dir) {
					t.Errorf("expected %s to have direction %d but got %d", typesType, typ.ChanDir(dir), got.Dir())
				}
			})
		}
	}
}

func TestSimpleTypeFor_BasicTypeMaps(t *testing.T) {
	for elemKind, elemType := range types.Typ {
		if elemKind == int(types.Invalid) {
			continue
		}
		for keyKind, keyType := range types.Typ {
			if keyKind == int(types.Invalid) {
				continue
			}
			t.Run(fmt.Sprintf("%s has type %s:%s", elemType, typ.Kind(elemKind), typ.Kind(keyKind)), func(t *testing.T) {
				typesType := types.NewMap(keyType, elemType)
				got := typ.SimpleTypeFor(typesType)
				if got.Kind() != typ.Map {
					t.Errorf("expected %s to translate to %s but got %s", typesType, typ.Map, got.Kind())
				}
				if got.Elem().Kind() != typ.Kind(elemKind) {
					t.Errorf("expected %s to have element %s but got %s", typesType, typ.Kind(elemKind), got.Elem().Kind())
				}
				if got.Key().Kind() != typ.Kind(keyKind) {
					t.Errorf("expected %s to have element %s but got %s", typesType, typ.Kind(keyKind), got.Key().Kind())
				}
			})
		}
	}
}

func TestSimpleTypeFor_BasicTypeSmallArrays(t *testing.T) {
	for kind, basicType := range types.Typ {
		if kind == int(types.Invalid) {
			continue
		}
		for _, size := range []int{1, 63} {
			t.Run(fmt.Sprintf("%s has type %s", basicType, typ.Kind(kind)), func(t *testing.T) {
				typesType := types.NewArray(basicType, int64(size))
				got := typ.SimpleTypeFor(typesType)
				if got.Kind() != typ.Array {
					t.Errorf("expected %s to translate to %s but got %s", typesType, typ.Slice, got.Kind())
				}
				if got.Elem().Kind() != typ.Kind(kind) {
					t.Errorf("expected %s to have element %s but got %s", typesType, typ.Kind(kind), got.Elem().Kind())
				}
				if got.Len() != size {
					t.Errorf("expected %s to have size %d but got %d", typesType, size, got.Len())
				}
			})
		}
	}
}

func TestSimpleTypeFor_TypeString(t *testing.T) {
	for kind, basicType := range types.Typ {
		if kind == int(types.Invalid) {
			continue
		}
		t.Run(fmt.Sprintf("%s prints as %s", basicType, typ.Kind(kind)), func(t *testing.T) {
			it := typ.SimpleTypeFor(basicType)
			if it.String() != typ.Kind(kind).String() {
				t.Errorf("expected %s; got %s", typ.Kind(kind).String(), it.String())
			}
		})

		t.Run(fmt.Sprintf("slice of %s prints as []%s", basicType, typ.Kind(kind)), func(t *testing.T) {
			it := typ.SimpleTypeFor(types.NewSlice(basicType))
			if it.String() != "[]"+typ.Kind(kind).String() {
				t.Errorf("expected []%s; got %s", typ.Kind(kind).String(), it.String())
			}
		})

		t.Run(fmt.Sprintf("ptr of %s prints as *%s", basicType, typ.Kind(kind)), func(t *testing.T) {
			it := typ.SimpleTypeFor(types.NewPointer(basicType))
			if it.String() != "*"+typ.Kind(kind).String() {
				t.Errorf("expected *%s; got %s", typ.Kind(kind).String(), it.String())
			}
		})

		t.Run(fmt.Sprintf("chan of %s prints as chan %s", basicType, typ.Kind(kind)), func(t *testing.T) {
			it := typ.SimpleTypeFor(types.NewChan(types.SendRecv, basicType))
			if it.String() != "chan "+typ.Kind(kind).String() {
				t.Errorf("expected chan %s; got %s", typ.Kind(kind).String(), it.String())
			}
		})

		t.Run(fmt.Sprintf("array of %s prints as [2]%s", basicType, typ.Kind(kind)), func(t *testing.T) {
			it := typ.SimpleTypeFor(types.NewArray(basicType, 2))
			if it.String() != "[2]"+typ.Kind(kind).String() {
				t.Errorf("expected [2]%s; got %s", typ.Kind(kind).String(), it.String())
			}
		})

		t.Run(fmt.Sprintf("map of %s prints as map[i]%s", basicType, typ.Kind(kind)), func(t *testing.T) {
			it := typ.SimpleTypeFor(types.NewMap(types.Typ[types.Int], basicType))
			if it.String() != "map[i]"+typ.Kind(kind).String() {
				t.Errorf("expected map[i]%s; got %s", typ.Kind(kind).String(), it.String())
			}
		})
	}
}
