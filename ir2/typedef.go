package ir2

import "go/types"

// TypeDef is a type definition
type TypeDef struct {
	pkg *Package

	Name       string
	Referenced bool

	Type types.Type
}
