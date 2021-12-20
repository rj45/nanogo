package ir2

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"sort"
	"strings"
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

func (prog *Program) Emit(out io.Writer, dec Decorator) {
	dec.Begin(out, prog)

	pkgs := prog.Packages()

	// ensure deterministic package order
	sort.SliceStable(pkgs, func(i, j int) bool {
		return strings.Compare(pkgs[i].Path, pkgs[j].Path) < 0
	})

	for i, pkg := range pkgs {
		if i != 0 {
			fmt.Fprintln(out)
		}
		pkg.Emit(out, dec)
	}
	dec.End(out, prog)
}

func (pkg *Package) Emit(out io.Writer, dec Decorator) {
	dec.Begin(out, pkg)
	fmt.Fprintf(out, "package %s %q\n", pkg.Name, pkg.Path)

	first := true
	for _, td := range pkg.typedefs {
		// todo: enable this somehow
		// if !td.Referenced {
		// 	continue
		// }

		if first {
			first = false
			fmt.Fprintln(out)
		}

		td.Emit(out, dec)
	}

	first = true
	for _, glob := range pkg.globals {
		if !glob.Referenced {
			continue
		}

		if first {
			first = false
			fmt.Fprintln(out)
		}

		glob.Emit(out, dec)
	}

	for _, fn := range pkg.funcs {
		if !fn.Referenced {
			continue
		}

		fmt.Fprintln(out)
		fn.Emit(out, dec)
	}
	dec.End(out, pkg)
}

func (glob *Global) Emit(out io.Writer, dec Decorator) {
	dec.Begin(out, glob)
	val := glob.Value
	valstr := ""
	if val != nil {
		// todo: wrap this in the decorator?
		if val.Kind() == StringConst {
			valstr = fmt.Sprintf(" = %q", val.String())
		} else {
			valstr = fmt.Sprintf(" = %s", val.String())
		}
	}

	typstr := types.TypeString(glob.Type, qualifier(glob.pkg.Type))
	fmt.Fprintf(out, "var %s:%s%s\n",
		dec.WrapLabel(glob.FullName, glob),
		dec.WrapType(typstr), valstr)

	dec.End(out, glob)
}

func (td *TypeDef) Emit(out io.Writer, dec Decorator) {
	dec.Begin(out, td)

	typ := td.Type.Underlying()
	typstr := types.TypeString(typ, qualifier(td.pkg.Type))
	fmt.Fprintf(out, "type %s:%s\n",
		dec.WrapLabel(td.Name, td),
		dec.WrapType(typstr))

	dec.End(out, td)
}

func (fn *Func) Emit(out io.Writer, dec Decorator) {
	dec.Begin(out, fn)

	// fmt.Fprintf(out, "; %s\n", dec.WrapType(fn.Sig.String()))
	fmt.Fprintf(out, "func %s:\n", dec.WrapLabel(fn.FullName, fn))
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

	// if len(blk.preds) > 0 {
	// 	fmt.Fprintf(out, " ; <=")

	// 	for _, pred := range blk.preds {
	// 		fmt.Fprintf(out, " %s", dec.WrapRef(pred.String(), pred))
	// 	}
	// }

	fmt.Fprintln(out)

	for it := blk.InstrIter(); it.HasNext(); it.Next() {
		it.Instr().Emit(out, dec)
	}

	dec.End(out, blk)
}

func qualifier(rootPkg *types.Package) types.Qualifier {
	return func(pkg *types.Package) string {
		if pkg == rootPkg {
			return ""
		} else {
			// todo: handle package aliases
			return pkg.Name()
		}
	}
}

func (in *Instr) Emit(out io.Writer, dec Decorator) {
	if in == nil {
		fmt.Fprint(out, "  <!nil>\n")
		return
	}

	dec.Begin(out, in)

	defstr := ""
	for i, def := range in.defs {
		if i != 0 {
			defstr += ", "
		}
		defstr += dec.WrapLabel(def.String(), def)
		if def.Type != nil {
			typstr := dec.WrapType(types.TypeString(def.Type, qualifier(in.Func().pkg.Type)))
			defstr += fmt.Sprintf(":%s", typstr)
		}
	}

	argstr := ""
	for i, arg := range in.args {
		if i != 0 {
			argstr += ", "
		}

		if arg == nil {
			argstr += "<!nil>"
			continue
		}

		globref := ""
		if arg.Const != nil && (arg.Const.Kind() == FuncConst || arg.Const.Kind() == GlobalConst) {
			globref = "^" // denote it's a global
		}

		argstr += dec.WrapRef(globref+arg.String(), arg)

		if arg.Const != nil && arg.Const.Kind() == StringConst {
			basic, ok := arg.Type.(*types.Basic)
			if !ok || (basic.Info()&types.IsUntyped) == 0 {
				argstr += ":"
				typstr := types.TypeString(arg.Type, qualifier(in.Func().pkg.Type))
				argstr += dec.WrapType(typstr)
			}
		}
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
			str += "  "
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

	fmt.Fprintln(out, str)

	dec.End(out, in)
}

func (in *Instr) LongString() string {
	buf := &bytes.Buffer{}
	in.Emit(buf, SSAString{})
	return buf.String()
}

func (val *Value) String() string {
	if val.ID == Placeholder {
		return "<" + val.Const.String() + ">"
	}
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
