package a32

import (
	"strings"

	"github.com/rj45/nanogo/ir2"
)

func (cpuArch) Asm(op ir2.Op, defs []string, args []string) string {
	return op.String() + " " + strings.Join(defs, ", ") + " " + strings.Join(args, ", ")
}
