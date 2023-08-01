package cleanup

import (
	"log"

	"github.com/rj45/nanogo/ir/reg"
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

var _ = xform2.Register(sequentializeCopies,
	xform2.OnlyPass(xform2.CleanUp),
	xform2.OnOp(op.Copy),
)

// sequentializeCopies takes a copy with more than one arg and figures out
// how to do the same thing in multiple copy instructions
// Based on Algorithm 13 from Benoit Boisinot's thesis, with fixes and
// extensions by Paul Sokolovsky.
// https://github.com/pfalcon/parcopy/blob/master/parcopy1.py
func sequentializeCopies(it ir2.Iter) {
	instr := it.Instr()

	if instr.NumArgs() < 2 {
		return
	}

	if !instr.Op.IsCopy() {
		log.Panicf("called with non-copy! %s", instr.LongString())
	}

	var ready []reg.Reg
	var todo []reg.Reg
	pred := make(map[reg.Reg]reg.Reg)
	loc := make(map[reg.Reg]reg.Reg)

	srcs := make(map[reg.Reg]*ir2.Value)
	dests := make(map[reg.Reg]*ir2.Value)

	// fmt.Println("seq:", instr.Func().Name, instr.LongString())

	var copied [][2]*ir2.Value

	emit := func(def, arg *ir2.Value) {
		cp := it.Insert(op.Copy, def.Type, arg)
		cpdef := cp.Def(0)
		cpdef.SetReg(def.Reg())
		def.ReplaceUsesWith(cpdef)
		copied = append(copied, [2]*ir2.Value{def, arg})
		it.Changed()
	}

	for i := 0; i < instr.NumDefs(); i++ {
		def := instr.Def(i)
		arg := instr.Arg(i)

		b := def.Reg()
		a := arg.Reg()

		if b == a {
			// wait for copy elimination first
			return
		}

		if arg.IsConst() {
			emit(def, arg)
			continue
		}

		srcs[a] = arg
		dests[b] = def

		loc[a] = a
		pred[b] = a

		for _, todob := range todo {
			if todob == b {
				panic("double destination assignment")
			}
		}

		todo = append(todo, b)
	}

	for i := 0; i < instr.NumDefs(); i++ {
		def := instr.Def(i)
		if instr.Arg(i).IsConst() {
			continue
		}

		b := def.Reg()

		if _, found := loc[b]; !found {
			ready = append(ready, b)
		}
	}

	for len(todo) > 0 {
		for len(ready) > 0 {
			b := ready[len(ready)-1]
			ready = ready[:len(ready)-1]

			a, found := pred[b]
			if !found {
				continue
			}
			c := loc[a]

			// fmt.Println("copy", b, "<-", c)
			emit(dests[b], srcs[c])

			for i, td := range todo {
				if td == c {
					// remove c from todo
					todo[i] = todo[len(todo)-1]
					todo = todo[:len(todo)-1]
				}
			}

			loc[a] = b

			if a == c {
				ready = append(ready, a)
			}
		}

		if len(todo) == 0 {
			break
		}

		b := todo[len(todo)-1]
		todo = todo[:len(todo)-1]

		if b != loc[pred[b]] {
			// need test program to verify this
			panic("todo: temp needed, or figure out swap chain")
		}
	}

	for _, pair := range copied {
		def, arg := pair[0], pair[1]
		instr.RemoveArg(arg)
		instr.RemoveDef(def)
	}
	if instr.NumDefs() == 0 {
		it.Remove()
	}
}
