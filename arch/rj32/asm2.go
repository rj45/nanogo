package rj32

import (
	"fmt"
	"strings"

	"github.com/rj45/nanogo/ir2"
)

func (cpuArch) Asm(op ir2.Op, defs, args []string) string {
	switch op {
	case Load, Loadb:
		return fmt.Sprintf("%s %s, [%s, %s]", op, defs[0], args[0], args[1])
	case Store, Storeb:
		return fmt.Sprintf("%s [%s, %s], %s", op, args[0], args[1], args[2])
	case Return:
		return "return"
	case Call:
		return "call " + args[0]
	default:
		if op.ClobbersArg() {
			args = args[1:]
		}
		if len(defs) > 0 && len(args) > 0 {
			return op.String() + " " + strings.Join(defs, ", ") + ", " + strings.Join(args, ", ")
		} else if len(defs) > 0 {
			return op.String() + " " + strings.Join(defs, ", ")
		}
		return op.String() + " " + strings.Join(args, ", ")
	}
}
