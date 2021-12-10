package ir2

import (
	"bytes"
	"fmt"
	"io"
)

type Decorator interface {
	Begin(out io.Writer, what interface{})
	End(out io.Writer, what interface{})

	WrapLabel(str string, what interface{}) string
	WrapRef(str string, what interface{}) string
	WrapType(str string) string
	WrapOp(str string, what Op) string
	SSAForm() bool
}

func (fn *Func) Emit(out io.Writer, dec Decorator) {
	dec.Begin(out, fn)
	fmt.Fprintf(out, "; %s\n", dec.WrapType(fn.Sig.String()))
	fmt.Fprintf(out, "%s:\n", dec.WrapLabel(fn.FullName, fn))
	for _, blk := range fn.blocks {
		blk.Emit(out, dec)
	}
	dec.End(out, fn)
}

func (fn *Func) LongString() string {
	buf := &bytes.Buffer{}
	fn.Emit(buf, SSAString{})
	return buf.String()
}

func (blk *Block) String() string {
	return fmt.Sprintf("b%d", blk.ID)
}

func (blk *Block) Emit(out io.Writer, dec Decorator) {
	dec.Begin(out, blk)

	fmt.Fprintf(out, ".%s:", dec.WrapLabel(blk.String(), blk))

	if len(blk.preds) > 0 {
		fmt.Fprintf(out, " ; <=")

		for _, pred := range blk.preds {
			fmt.Fprintf(out, " %s", dec.WrapRef(pred.String(), pred))
		}
	}

	fmt.Fprintln(out)

	for it := blk.InstrIter(); it.HasNext(); it.Next() {
		it.Instr().Emit(out, dec)
	}

	dec.End(out, blk)
}

func (in *Instr) Emit(out io.Writer, dec Decorator) {
	if in == nil {
		fmt.Fprint(out, "  <!nil>\n")
		return
	}

	dec.Begin(out, in)

	defstr := ""
	typstr := ""
	for i, def := range in.defs {
		if i != 0 {
			defstr += ", "
			typstr += ", "
		}
		defstr += dec.WrapLabel(def.String(), def)
		if def.Type != nil {
			typstr += def.Type.String()
		}
	}

	argstr := ""
	for i, arg := range in.args {
		if i != 0 {
			argstr += ", "
		}
		argstr += dec.WrapRef(arg.String(), arg)
	}

	str := ""

	opstr := "<!nilOp>"
	if in.Op != nil {
		opstr = dec.WrapOp(in.String(), in.Op)
	}

	if dec.SSAForm() {
		if len(defstr) > 0 {
			str += fmt.Sprintf("  %s = ", defstr)
		} else {
			str += "            "
		}
		str += fmt.Sprintf("%-6s", opstr)
	} else {
		str += fmt.Sprintf("  %-6s", opstr)
		if len(defstr) > 0 {
			str += fmt.Sprintf(" %s", defstr)
		}
		if len(defstr) > 0 && len(argstr) > 0 {
			str += ","
		}
	}

	if len(argstr) > 0 {
		str += fmt.Sprintf(" %s", argstr)
	}

	if in == in.blk.Control() {
		if len(argstr) > 0 && len(in.blk.succs) > 0 {
			str += ", "
		}
		for i, succ := range in.blk.succs {
			if i != 0 {
				str += ", "
			}
			str += succ.String()
		}
	}

	if dec.SSAForm() && len(typstr) > 0 {
		fmt.Fprintf(out, "%-30s %s", str, dec.WrapType(fmt.Sprintf("<%s>", typstr)))
	} else {
		fmt.Fprint(out, str)
	}

	fmt.Fprintln(out)

	dec.End(out, in)
}

func (in *Instr) LongString() string {
	buf := &bytes.Buffer{}
	in.Emit(buf, SSAString{})
	return buf.String()
}

func (val *Value) String() string {
	if val.Const != nil {
		return val.Const.String()
	}
	return fmt.Sprintf("v%d", val.ID)
}

// SSAString emits a plain string in SSA form
type SSAString struct{}

func (ss SSAString) Begin(out io.Writer, what interface{}) {}
func (ss SSAString) End(out io.Writer, what interface{})   {}

func (ss SSAString) WrapLabel(str string, what interface{}) string {
	return str
}

func (ss SSAString) WrapRef(str string, what interface{}) string {
	return str
}

func (ss SSAString) WrapType(str string) string {
	return str
}

func (ss SSAString) WrapOp(str string, what Op) string {
	return str
}

func (ss SSAString) SSAForm() bool {
	return true
}
