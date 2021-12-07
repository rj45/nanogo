package ir2

import (
	"fmt"
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
		if fn.name == name {
			return fn
		}
	}
	return nil
}

// AddFunc adds a func to the list
func (pkg *Package) NewFunc(name string) {
	fn := &Func{
		name:     name,
		FullName: pkg.genUniqueName(name),
	}
	fn.pkg = pkg
	pkg.funcs = append(pkg.funcs, pkg.funcs...)
}

func (pkg *Package) genUniqueName(name string) string {
	parts := strings.Split(pkg.FullName, "/")
	fullName := fmt.Sprintf("%s__%s", parts[len(parts)-1], name)
	parts = parts[:len(parts)-1]
	for pkg.prog.takenNames[fullName] {
		fullName = parts[len(parts)-1] + "_" + fullName
		parts = parts[:len(parts)-1]
	}
	pkg.prog.claimName(fullName)
	return fullName
}
