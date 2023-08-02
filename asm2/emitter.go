package asm2

import (
	"fmt"
	"io"
	"strings"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/sizes"
)

type Section string

const (
	Code Section = "code"
	Data Section = "data"
	Bss  Section = "bss"
)

type Formatter interface {
	Section(s Section) string
	GlobalLabel(global *ir2.Global) string
	PCRelAddress(offsetWords int) string
	Word(val string) string
	String(val string) string
	Reserve(bytes int) string
	Comment(comment string) string
	BlockLabel(id string) string
}

type Emitter struct {
	out   io.Writer
	fmter Formatter

	emittedGlobals map[*ir2.Global]bool
	emittedFuncs   map[*ir2.Func]bool

	section Section
	indent  string
}

func NewEmitter(out io.Writer, fmter Formatter) *Emitter {
	return &Emitter{
		out:            out,
		fmter:          fmter,
		emittedGlobals: make(map[*ir2.Global]bool),
		emittedFuncs:   make(map[*ir2.Func]bool),
	}
}

func Emit(out io.Writer, fmter Formatter, prog *ir2.Program) {
	emitter := NewEmitter(out, fmter)
	emitter.Program(prog)
}

func (emit *Emitter) Program(prog *ir2.Program) {
	mainpkg := prog.Package("main")
	emit.assemble(mainpkg.Func("init"))
	emit.assemble(mainpkg.Func("main"))
}

func (emit *Emitter) assemble(fn *ir2.Func) {
	var funcs []*ir2.Func
	var globals []*ir2.Global

	seenFunc := map[*ir2.Func]bool{fn: true}

	todo := []*ir2.Func{fn}
	for len(todo) > 0 {
		fn := todo[0]
		todo = todo[1:]

		funcs, globals = emit.scan(fn, funcs, globals)

		for _, f := range funcs {
			if !seenFunc[f] && !emit.emittedFuncs[f] {
				seenFunc[f] = true
				todo = append(todo, f)
			}
		}
		funcs = funcs[:]

		for _, glob := range globals {
			if !emit.emittedGlobals[glob] {
				emit.emittedGlobals[glob] = true
				emit.global(glob)
			}
		}
		globals = globals[:]

		emit.emittedFuncs[fn] = true
		emit.fn(fn)
	}
}

func (emit *Emitter) fn(fn *ir2.Func) {
	emit.ensureSection(Code)
	params := fn.Sig.Params()
	pstrs := make([]string, params.Len())

	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		pstrs[i] = fmt.Sprintf("%s %s", param.Name(), param.Type())
	}

	res := fn.Sig.Results()
	resstr := res.String()
	if res.Len() == 0 {
		resstr = ""
	} else if res.Len() == 1 {
		resstr = res.At(0).String()
	}

	emit.comment("func %s(%s)%s", fn.FullName, strings.Join(pstrs, ", "), resstr)

	for b := 0; b < fn.NumBlocks(); b++ {
		blk := fn.Block(b)

		emit.line(emit.fmter.BlockLabel(blk.IDString()) + ":")
		emit.indent = "    "

		for i := 0; i < blk.NumInstrs(); i++ {
			instr := blk.Instr(i)

			defs := make([]string, 0, instr.NumDefs())
			for d := 0; d < instr.NumDefs(); d++ {
				def := instr.Def(d)
				str := ""
				switch {
				case def.InReg():
					str = def.Reg().String()
				default:
					str = def.String()
				}
				defs = append(defs, str)
			}

			args := make([]string, 0, instr.NumArgs())
			for a := 0; a < instr.NumArgs(); a++ {
				arg := instr.Arg(a)
				str := ""
				switch {
				case arg.InReg():
					str = arg.Reg().String()
				case arg.IsConst() && arg.Const().Kind() == ir2.BoolConst:
					// todo: should probably be an xform
					if b, _ := ir2.BoolValue(arg.Const()); b {
						str = "1"
					} else {
						str = "0"
					}
				default:
					str = arg.String()
				}
				args = append(args, str)
			}

			if i == blk.NumInstrs()-1 {
				if blk.NumSuccs() == 1 { // jump
					if b < fn.NumBlocks()-1 && blk.Succ(0) == fn.Block(b+1) {
						// block falls through
						continue
					}
				}
				if blk.NumSuccs() > 0 {
					args = append(args, emit.fmter.BlockLabel(blk.Succ(0).IDString()))
				}
			}

			emit.line("%s", arch.Asm(instr.Op, defs, args))
		}
		emit.indent = ""
	}

	emit.line("")
}

func (emit *Emitter) global(glob *ir2.Global) {
	if glob.Value != nil {
		emit.ensureSection(Data)
	} else {
		emit.ensureSection(Bss)
	}
	emit.line("%s:", emit.fmter.GlobalLabel(glob))
	if glob.Value == nil {
		bytes := sizes.Sizeof(glob.Type)
		emit.line("%s", emit.fmter.Reserve(int(bytes)))
	} else if str, ok := ir2.StringValue(glob.Value); ok {
		emit.line("%s", emit.fmter.Word(emit.fmter.PCRelAddress(int(sizes.WordSize()*2))))

		emit.line("%s", emit.fmter.Word(fmt.Sprintf("%d", len(str))))
		emit.line("%s", emit.fmter.String(str))
	} else if val, ok := ir2.IntValue(glob.Value); ok {
		// todo: implement more types
		emit.line("%s", emit.fmter.Word(fmt.Sprintf("%d", val)))
	} else {
		panic("todo: implement more types")
	}
	emit.line("")
}

func (emit *Emitter) scan(fn *ir2.Func, funcs []*ir2.Func, globals []*ir2.Global) ([]*ir2.Func, []*ir2.Global) {
	for b := 0; b < fn.NumBlocks(); b++ {
		blk := fn.Block(b)
		for i := 0; i < blk.NumInstrs(); i++ {
			instr := blk.Instr(i)

			for a := 0; a < instr.NumArgs(); a++ {
				arg := instr.Arg(a)

				if arg.IsConst() {
					constant := arg.Const()
					if fnc, ok := ir2.FuncValue(constant); ok {
						funcs = append(funcs, fnc)
					} else if glob, ok := ir2.GlobalValue(constant); ok {
						globals = append(globals, glob)
					}
				}
			}
		}
	}

	return funcs, globals
}

func (emit *Emitter) ensureSection(section Section) {
	if emit.section != section {
		emit.line(emit.fmter.Section(section))
		emit.section = section
	}
}

func (emit *Emitter) line(fmtstr string, args ...interface{}) {
	output := fmt.Sprintf(emit.indent+fmtstr, args...)
	fmt.Fprintln(emit.out, output)
}

func (emit *Emitter) comment(fmtstr string, args ...interface{}) {
	output := emit.fmter.Comment(fmt.Sprintf(fmtstr, args...))
	fmt.Fprintln(emit.out, emit.indent+output)
}
