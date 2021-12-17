package ir2

import (
	"fmt"
	"go/types"
	"strings"
)

// Funcs returns a copy of the func list
func (pkg *Package) Funcs() []*Func {
	return append([]*Func(nil), pkg.funcs...)
}

// Func finds a func by either Name or FullName
func (pkg *Package) Func(name string) *Func {
	for _, fn := range pkg.funcs {
		if fn.Name == name || fn.FullName == name {
			return fn
		}
	}
	return nil
}

// Globals returns a copy of the global list
func (pkg *Package) Globals() []*Global {
	return append([]*Global(nil), pkg.globals...)
}

// Global finds the Global by Name or FullName
func (pkg *Package) Global(name string) *Global {
	for _, glob := range pkg.globals {
		if glob.Name == name || glob.FullName == name {
			return glob
		}
	}
	return nil
}

// Program that the package belongs to
func (pkg *Package) Program() *Program {
	return pkg.prog
}

// NewFunc adds a func to the list
func (pkg *Package) NewFunc(name string, sig *types.Signature) *Func {
	fn := &Func{
		Name:     name,
		FullName: pkg.genUniqueName(name),
		Sig:      sig,
	}
	fn.pkg = pkg
	pkg.funcs = append(pkg.funcs, fn)
	return fn
}

// NewGlobal adds a global to the list
func (pkg *Package) NewGlobal(name string, typ types.Type) *Global {
	glob := &Global{
		Name:     name,
		FullName: pkg.genUniqueName(name),
		Type:     typ,
	}
	glob.pkg = pkg
	pkg.globals = append(pkg.globals, glob)
	return glob
}

func (pkg *Package) NewStringLiteral(funcname, str string) *Global {
	glob := pkg.prog.strings[str]
	if glob != nil {
		return glob
	}

	// move to building a global as the string literal
	name := pkg.makeUnique(funcname)
	glob = pkg.NewGlobal(name, types.Typ[types.String])
	glob.Value = ConstFor(str)
	pkg.prog.registerStringLiteral(glob)

	return glob
}

func (pkg *Package) makeUnique(name string) string {
	for i := 1; ; i++ {
		uniq := fmt.Sprintf("%s_%d", name, i)
		if pkg.Global(uniq) != nil {
			continue
		}

		if pkg.Func(uniq) != nil {
			continue
		}

		// todo: check types too

		return uniq
	}
}

func (pkg *Package) genUniqueName(name string) string {
	name = strings.ReplaceAll(name, "$", "_")
	parts := strings.Split(pkg.Path, "/")
	fullName := fmt.Sprintf("%s__%s", pkg.Name, name)
	for pkg.prog.takenNames[fullName] {
		fullName = parts[len(parts)-1] + "_" + fullName
		parts = parts[:len(parts)-1]
	}
	pkg.prog.claimName(fullName)
	return fullName
}
