package elaboration

import (
	"log"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(builtins,
	xform2.OnlyPass(xform2.Elaboration),
	xform2.OnOp(op.CallBuiltin),
)

func builtins(it ir2.Iter) {
	instr := it.Instr()
	fn := instr.Func()

	builtin, _ := ir2.StringValue(instr.Arg(0).Const())
	switch builtin {
	case "print", "println":
		runtime := fn.Package().Program().Package("runtime")
		printspace := runtime.Func("printspace")

		for a := 1; a < instr.NumArgs(); a++ {
			arg := instr.Arg(a)
			if a != 1 {
				it.Insert(op.Call, nil, fn.ValueFor(printspace.Sig, printspace))
				printspace.Referenced = true
			}

			pfunc := runtime.Func("print" + arg.Type.String())
			if pfunc == nil {
				log.Fatalf("unsupported type %s in %s", arg.Type, instr.Func().FullName)
			}

			pfuncval := fn.ValueFor(pfunc.Sig, pfunc)

			pfunc.Referenced = true
			it.Insert(op.Call, nil, pfuncval, arg)
		}
		if it.Instr().Op != op.CallBuiltin {
			log.Panicf("in wrong loc")
		}
		it.Remove()
	}
}
