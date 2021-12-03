package arch

import (
	"log"
	"strings"

	"github.com/rj45/nanogo/codegen"
	"github.com/rj45/nanogo/compiler"
	"github.com/rj45/nanogo/ir/op"
	"github.com/rj45/nanogo/ir/reg"
	"github.com/rj45/nanogo/sizes"
)

const defaultArch = "rj32"

type Architecture interface {
	codegen.Arch
	reg.Arch
	sizes.Arch
	op.Arch
	compiler.Arch
}

var arch Architecture

var arches map[string]Architecture

func Arch() Architecture {
	return arch
}

func Register(a Architecture) int {
	if arches == nil {
		arches = make(map[string]Architecture)
	}
	name := strings.ToLower(a.Name())
	arches[name] = a
	if name == defaultArch {
		SetArch(name)
	}
	return 0
}

func SetArch(name string) {
	arch = arches[strings.ToLower(name)]
	if arch == nil {
		log.Panicf("unknown arch %s", name)
	}
	reg.SetArch(arch)
	codegen.SetArch(arch)
	sizes.SetArch(arch)
	op.SetArch(arch)
	compiler.SetArch(arch)
}
