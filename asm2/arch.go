package asm2

import "github.com/rj45/nanogo/ir2"

type Arch interface {
	Asm(op ir2.Op, defs []string, args []string) string
}

var arch Arch

func SetArch(a Arch) {
	arch = a
}
