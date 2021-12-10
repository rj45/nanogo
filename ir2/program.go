package ir2

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

func (prog *Program) claimName(name string) {
	if prog.takenNames == nil {
		prog.takenNames = make(map[string]bool)
	}
	prog.takenNames[name] = true
}
