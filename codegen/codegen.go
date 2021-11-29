package codegen

import (
	"fmt"
	"io"
	"strings"

	"github.com/rj45/nanogo/codegen/asm"
	"github.com/rj45/nanogo/codegen/asm/rj32"
	"github.com/rj45/nanogo/ir"
)

type Generator struct {
	mod *ir.Package
	out io.Writer

	emittedGlobals map[*ir.Value]bool

	src []string

	section string
	indent  string

	arch asm.Arch
	fn   *asm.Func
}

func NewGenerator(mod *ir.Package) *Generator {
	return &Generator{
		mod:            mod,
		emittedGlobals: make(map[*ir.Value]bool),
		arch:           rj32.Rj32{},
	}
}

func (gen *Generator) emit(fmtstr string, args ...interface{}) {
	nextline := ""
	if len(gen.src) > 0 {
		nextline, gen.src = gen.src[len(gen.src)-1], gen.src[:len(gen.src)-1]
	}
	output := fmt.Sprintf(gen.indent+fmtstr, args...)

	if nextline != "" {
		for len(output) < 40 {
			output += " "
		}
		output += "; "
		output += nextline
	}

	fmt.Fprintln(gen.out, output)
}

func (gen *Generator) source(src string) {
	if src == "" {
		return
	}

	lines := strings.Split(src, "\n")
	var revlines []string
	for i := len(lines) - 1; i >= 0; i-- {
		revlines = append(revlines, lines[i])
	}
	revlines = append(revlines, gen.src...)
	gen.src = revlines
}
