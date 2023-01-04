package ir2

import "go/types"

// Global is a global variable or literal stored in memory
type Global struct {
	pkg *Package

	Name       string
	FullName   string
	Type       types.Type
	Referenced bool

	// initial value
	Value Const
}

func (glob *Global) String() string {
	return glob.FullName
}

func (glob *Global) Package() *Package {
	return glob.pkg
}
