package frontend

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types/typeutil"
)

type TypeMapper struct {
	tmap *typeutil.Map
}

type typeInfo struct {
	count   int
	runtime bool
}

func (tm *TypeMapper) scan(prog *ssa.Program) {
	runtimeTypes := prog.RuntimeTypes()
	if len(runtimeTypes) > 0 {
		for _, typ := range runtimeTypes {
			info := tm.scanType(typ)
			if info != nil {
				info.runtime = true
			}
		}
	}

	pkgs := prog.AllPackages()
	for _, pkg := range pkgs {
		tm.scanScope(pkg.Pkg.Scope())
	}
}

func (tm *TypeMapper) scanScope(scope *types.Scope) {
	for _, name := range scope.Names() {
		tm.scanObject(scope.Lookup(name))
	}

	for i := 0; i < scope.NumChildren(); i++ {
		tm.scanScope(scope.Child(i))
	}
}

func (tm *TypeMapper) scanObject(obj types.Object) {
	switch obj.(type) {
	case *types.Label, *types.PkgName, *types.Builtin, *types.Nil:
		// these don't have a type per se, so ignoring them
	default:
		tm.scanType(obj.Type())
	}
}

func (tm *TypeMapper) scanType(typ types.Type) *typeInfo {
	switch t := typ.(type) {
	case *types.Basic:
		// basic types will not have side tables mapped to them unless
		// they are named/defined. So we can just ignore those
		return nil
	case *types.Slice, *types.Pointer, *types.Chan:
		e := typ.(interface{ Elem() types.Type }).Elem()
		if _, isBasic := e.(*types.Basic); isBasic {
			// a slice/pointer to a basic type also doesn't need a side
			// table
			return nil
		}
	case *types.Map:
		_, basicKey := t.Key().(*types.Basic)
		_, basicValue := t.Elem().(*types.Basic)
		if basicKey && basicValue {
			// a slice/pointer to a basic type also doesn't need a side
			// table
			return nil
		}
	}

	hit := tm.tmap.At(typ)
	if hit != nil {
		info := hit.(*typeInfo)
		info.count++
		return info
	}

	info := &typeInfo{
		count: 1,
	}
	tm.tmap.Set(typ, info)

	if e, ok := typ.(interface{ Elem() types.Type }); ok {
		tm.scanType(e.Elem())
	}

	tm.scanType(typ.Underlying())

	switch t := typ.(type) {
	case *types.Map:
		tm.scanType(t.Key())
	case *types.Tuple:
		tm.scanTuple(t)
	case *types.Signature:
		if t.Recv() != nil {
			tm.scanVar(t.Recv())
		}
		tm.scanTuple(t.Results())
		tm.scanTuple(t.Params())
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			tm.scanVar(t.Field(i))
		}
	case *types.Named:
		tm.scanObject(t.Obj()) // required?
		for i := 0; i < t.NumMethods(); i++ {
			tm.scanObject(t.Method(i))
		}
	case *types.Interface:
		for i := 0; i < t.NumEmbeddeds(); i++ {
			tm.scanType(t.EmbeddedType(i))
		}
		for i := 0; i < t.NumMethods(); i++ {
			tm.scanObject(t.Method(i))
		}
	}
	return info
}

func (tm *TypeMapper) scanTuple(t *types.Tuple) {
	for i := 0; i < t.Len(); i++ {
		tm.scanType(t.At(i).Type())
	}
}

func (tm *TypeMapper) scanVar(v *types.Var) {
	tm.scanType(v.Type())
}
