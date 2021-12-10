package ir2

import "fmt"

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

func (prog *Program) StringLiteral(str string, name string) *Literal {
	if lit, ok := prog.strings[str]; ok {
		return lit
	}
	if prog.strings == nil {
		prog.strings = make(map[string]*Literal)
	}
	prog.claimName(name)
	lit := &Literal{Name: name, Value: ConstFor(str)}
	prog.strings[str] = lit
	return lit
}

func (prog *Program) MakeUnique(name string) string {
	for i := 1; ; i++ {
		uniq := fmt.Sprintf("%s_%d", name, i)
		if prog.takenNames[uniq] {
			continue
		}
		return uniq
	}
}

func (prog *Program) claimName(name string) {
	if prog.takenNames == nil {
		prog.takenNames = make(map[string]bool)
	}
	prog.takenNames[name] = true
}
