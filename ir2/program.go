package ir2

import "go/token"

// Program is a collection of packages,
// which comprise a whole program.
type Program struct {
	packages []*Package

	FileSet *token.FileSet

	takenNames map[string]bool
	strings    map[string]*Global
}

// Packages returns a copy of the package list
func (prog *Program) Packages() []*Package {
	return append([]*Package(nil), prog.packages...)
}

// AddPackage adds a package to the list
func (prog *Program) AddPackage(pkg *Package) {
	pkg.prog = prog
	prog.packages = append(prog.packages, pkg)
}

// Package finds a package first by full name, then
// if there is no match, by short name.
func (prog *Program) Package(name string) *Package {
	for _, pkg := range prog.packages {
		if pkg.Path == name {
			return pkg
		}
	}
	for _, pkg := range prog.packages {
		if pkg.Name == name {
			return pkg
		}
	}
	return nil
}

// Global searches each package for a global
func (prog *Program) Global(name string) *Global {
	for _, pkg := range prog.packages {
		glob := pkg.Global(name)
		if glob != nil {
			return glob
		}
	}
	return nil
}

// Func searches each package for a func
func (prog *Program) Func(name string) *Func {
	for _, pkg := range prog.packages {
		fn := pkg.Func(name)
		if fn != nil {
			return fn
		}
	}
	return nil
}

func (prog *Program) StringLiteral(str string, fullname string) *Global {
	if glob, ok := prog.strings[str]; ok {
		return glob
	}
	return nil
}

func (prog *Program) registerStringLiteral(glob *Global) {
	if prog.strings == nil {
		prog.strings = make(map[string]*Global)
	}
	prog.strings[glob.Value.String()] = glob
}

func (prog *Program) claimName(name string) {
	if prog.takenNames == nil {
		prog.takenNames = make(map[string]bool)
	}
	prog.takenNames[name] = true
}
