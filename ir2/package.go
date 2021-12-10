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

// Func finds a package first by full name, then
// if there is no match, by short name.
func (pkg *Package) Func(name string) *Func {
	for _, fn := range pkg.funcs {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

// Program that the package belongs to
func (pkg *Package) Program() *Program {
	return pkg.prog
}

// AddFunc adds a func to the list
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

func (pkg *Package) genUniqueName(name string) string {
	parts := strings.Split(pkg.Path, "/")
	fullName := fmt.Sprintf("%s__%s", pkg.Name, name)
	for pkg.prog.takenNames[fullName] {
		fullName = parts[len(parts)-1] + "_" + fullName
		parts = parts[:len(parts)-1]
	}
	pkg.prog.claimName(fullName)
	return fullName
}
