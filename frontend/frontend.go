package frontend

import (
	"bytes"
	"flag"
	"fmt"
	"go/token"
	"go/types"
	"log"
	"os"
	"strings"

	"github.com/rj45/nanogo/ir2"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types/typeutil"
)

type Arch interface {
	Name() string
}

var arch Arch

func SetArch(a Arch) {
	arch = a
}

type FrontEnd struct {
	prog     *ir2.Program
	members  []ssa.Member
	ssaFuncs map[*ir2.Func]*ssa.Function
	parsed   map[*ir2.Func]bool

	instrmap  map[*ir2.Instr]ssa.Instruction
	val2instr map[ssa.Value]*ir2.Instr
	val2val   map[ssa.Value]*ir2.Value
	blockmap  map[*ssa.BasicBlock]*ir2.Block
}

func NewFrontEnd(dir string, patterns ...string) *FrontEnd {
	members, err := parseProgram(dir, patterns...)
	if err != nil {
		log.Fatal(err)
	}

	return &FrontEnd{prog: &ir2.Program{}, members: members}
}

var dumptypes = flag.Bool("dumptypes", false, "Dump all types in a program")

func (fe *FrontEnd) Scan() {

	if *dumptypes {
		pkg := fe.members[0].Package()
		fe.dumpTypes(pkg.Prog)
	}

	fe.ssaFuncs = make(map[*ir2.Func]*ssa.Function)
	for _, member := range fe.members {
		switch member.Token() {
		case token.FUNC:
			name := member.Name()
			fn := member.Package().Func(name)

			main := fn.Pkg.Pkg.Name() == "main"
			referenced := main && (name == "main" || name == "init")

			pkg := fe.getPackage(fn.Pkg.Pkg)

			irFunc := pkg.NewFunc(fn.Name(), fn.Signature)
			irFunc.Referenced = referenced

			fe.ssaFuncs[irFunc] = fn

		case token.VAR:
			pkg := fe.getPackage(member.Package().Pkg)
			pkg.NewGlobal(member.Name(), member.Type())

		case token.TYPE:
		case token.CONST:
		default:
			log.Fatalln("unknown type", member.Token())
		}
	}
	fe.parsed = make(map[*ir2.Func]bool)
}

func (fe *FrontEnd) getPackage(typPkg *types.Package) *ir2.Package {
	pkg := fe.prog.Package(typPkg.Path())
	if pkg == nil {
		pkg = &ir2.Package{
			Name: typPkg.Name(),
			Path: typPkg.Path(),
			Type: typPkg,
		}
		fe.prog.AddPackage(pkg)
	}
	return pkg
}

func (fe *FrontEnd) dumpTypes(prog *ssa.Program) {
	tm := &TypeMapper{
		tmap: &typeutil.Map{},
	}
	tm.scan(prog)

	tm.tmap.Iterate(func(key types.Type, value interface{}) {
		info := value.(*typeInfo)
		fmt.Fprintf(os.Stderr, "%d: %s\n", info.count, key)
	})
}

func (fe *FrontEnd) Program() *ir2.Program {
	return fe.prog
}

func (fe *FrontEnd) NextUnparsedFunc() *ir2.Func {
	for _, pkg := range fe.prog.Packages() {
		for _, fn := range pkg.Funcs() {
			if fn.Referenced && !fe.parsed[fn] {
				return fn
			}
		}
	}

	return nil
}

func (fe *FrontEnd) DumpOrignalSource(fn *ir2.Func) (filename string, lines []string, startline int) {
	ssafn := fe.ssaFuncs[fn]
	fset := ssafn.Prog.Fset

	if ssafn.Syntax() == nil {
		return
	}

	start := ssafn.Syntax().Pos()
	end := ssafn.Syntax().End()

	if start == token.NoPos || end == token.NoPos {
		return
	}

	startp := fset.PositionFor(start, true)
	filename = startp.Filename
	startline = startp.Line - 1

	endp := fset.PositionFor(end, true)
	buf, err := os.ReadFile(startp.Filename)
	if err != nil {
		log.Fatal(err)
	}
	lines = strings.Split(string(buf), "\n")
	lines = lines[startline:endp.Line]

	return
}

func (fe *FrontEnd) DumpOriginalSSA(fn *ir2.Func) *bytes.Buffer {
	buf := &bytes.Buffer{}
	ssa.WriteFunction(buf, fe.ssaFuncs[fn])
	return buf
}

func (fe *FrontEnd) ParseFunc(fn *ir2.Func) {
	fe.parsed[fn] = true

	fe.val2instr = make(map[ssa.Value]*ir2.Instr)
	fe.val2val = make(map[ssa.Value]*ir2.Value)
	fe.instrmap = make(map[*ir2.Instr]ssa.Instruction)
	fe.blockmap = make(map[*ssa.BasicBlock]*ir2.Block)

	fe.translateFunc(fn, fe.ssaFuncs[fn])
}
