// Copyright (c) 2021 rj45 (github.com/rj45), MIT Licensed, see LICENSE.
package compiler

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rj45/nanogo/codegen"
	"github.com/rj45/nanogo/codegen/asm"
	"github.com/rj45/nanogo/frontend"
	"github.com/rj45/nanogo/goenv"
	"github.com/rj45/nanogo/html"
	html2 "github.com/rj45/nanogo/html2"
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/parseir"
	"github.com/rj45/nanogo/parser"
	"github.com/rj45/nanogo/regalloc"
	"github.com/rj45/nanogo/regalloc2"
	"github.com/rj45/nanogo/regalloc2/verify"
	"github.com/rj45/nanogo/xform"
	"github.com/rj45/nanogo/xform2"

	_ "github.com/rj45/nanogo/xform2/elaboration"
)

type Arch interface {
	Name() string
	AssemblerFormat() string
	EmulatorCmd() string
	EmulatorArgs() []string
}

func SetArch(a Arch) {
	arch = a
}

var arch Arch

type Mode int

const (
	Asm Mode = 1 << iota
	Assemble
	Run
	IR
)

type dumper interface {
	WritePhase(string, string)
	WriteAsmBuf(string, *bytes.Buffer)
	WriteAsm(string, *asm.Func)
	WriteSources(phase string, fn string, lines []string, startline int)
	Close()
}

type nopDumper struct{}

func (nopDumper) WritePhase(string, string)                                           {}
func (nopDumper) WriteAsmBuf(string, *bytes.Buffer)                                   {}
func (nopDumper) WriteAsm(string, *asm.Func)                                          {}
func (nopDumper) WriteSources(phase string, fn string, lines []string, startline int) {}
func (nopDumper) Close()                                                              {}

type dumper2 interface {
	WritePhase(string, string)
	WriteSources(phase string, fn string, lines []string, startline int)
	Close()
}

type nopDumper2 struct{}

func (nopDumper2) WritePhase(string, string)                                           {}
func (nopDumper2) WriteSources(phase string, fn string, lines []string, startline int) {}
func (nopDumper2) Close()                                                              {}

type nopWriteCloser struct{ w io.Writer }

func (nopWriteCloser) Close() error {
	return nil
}

func (n nopWriteCloser) Write(p []byte) (int, error) {
	return n.w.Write(p)
}

var _ io.WriteCloser = nopWriteCloser{}

var dump = flag.String("dump", "", "Dump a function to ssa.html")
var trace = flag.Bool("trace", false, "debug program with tracing info")

func Compile(outname, dir string, patterns []string, mode Mode) int {
	log.SetFlags(log.Lshortfile)

	var finalout io.WriteCloser
	var asmout io.WriteCloser

	if outname == "-" {
		finalout = nopWriteCloser{os.Stdout}
	} else {
		f, err := os.Create(outname)
		if err != nil {
			log.Fatal(err)
		}
		finalout = f
	}

	if filepath.Ext(patterns[0]) == ".ngir" {
		f, err := os.Open(patterns[0])
		if err != nil {
			panic(err)
		}
		defer f.Close()

		prog := &ir2.Program{}
		p, err := parseir.NewParser(patterns[0], f, prog, *trace)
		if err != nil {
			panic(err)
		}

		err = p.Parse()
		if err != nil {
			p.PrintErrors()
			return 1
		}

		prog.Emit(finalout, ir2.SSAString{})

		return 0
	}

	if mode&IR != 0 {
		fe, err := frontend.NewFrontEnd(dir, patterns...)
		if err != nil {
			log.Fatal(err)
		}

		fe.Scan()
		for fn := fe.NextUnparsedFunc(); fn != nil; fn = fe.NextUnparsedFunc() {
			var w dumper2
			w = nopDumper2{}
			if *dump != "" && strings.Contains(fn.FullName, *dump) {
				w = html2.NewHTMLWriter("ssa.html", fn)
				filename, lines, start := fe.DumpOrignalSource(fn)
				w.WriteSources("go", filename, lines, start)
				// w.WriteAsmBuf("tools/go/ssa", parser.DumpOriginalSSA(fn))
			}
			defer w.Close()

			fe.ParseFunc(fn)

			w.WritePhase("initial", "initial")

			xform2.Transform(xform2.Elaboration, fn)

			w.WritePhase("elaboration", "elaboration")

			ra := regalloc2.NewRegAlloc(fn)
			err = ra.Allocate()
			regalloc2.WriteGraphvizCFG(ra)
			regalloc2.DumpLivenessChart(ra)
			regalloc2.WriteGraphvizInterferenceGraph(ra)
			regalloc2.WriteGraphvizLivenessGraph(ra)
			if err != nil {
				log.Fatal(err)
			}
			errs := verify.Verify(fn)
			for _, err := range errs {
				log.Printf("verification error: %s\n", err)
			}

			w.WritePhase("regalloc", "regalloc")
		}

		fe.Program().Emit(finalout, ir2.SSAString{})

		return 0
	}

	asmout = finalout

	var binfile string
	var asmcmd *exec.Cmd
	if mode&Assemble != 0 {
		// todo: if specified, allow this to not be a temp file
		asmtemp, err := os.CreateTemp("", "nanogo_*.asm")
		if err != nil {
			log.Fatalln("failed to create temp asm file for customasm:", err)
		}
		defer os.Remove(asmtemp.Name())

		bintemp, err := os.CreateTemp("", "nanogo_*.bin")
		if err != nil {
			log.Fatalln("failed to create temp bin file for customasm:", err)
		}
		bintemp.Close() // customasm will write to it
		binfile = bintemp.Name()
		defer os.Remove(bintemp.Name())

		root := goenv.Get("NANOGOROOT")
		path := filepath.Join(root, "arch", arch.Name(), "customasm")
		cpudef := filepath.Join(path, "cpudef.asm")
		rungo := filepath.Join(path, "rungo.asm")

		asmcmd = exec.Command("customasm", "-q",
			"-f", arch.AssemblerFormat(),
			"-o", bintemp.Name(),
			cpudef, rungo, asmtemp.Name())
		log.Println(asmcmd)
		asmcmd.Stderr = os.Stderr
		asmcmd.Stdout = os.Stdout
		asmout = asmtemp
	}

	var runcmd *exec.Cmd
	if mode&Run != 0 {
		args := arch.EmulatorArgs()
		args = append(args, binfile)
		if *trace {
			args = append(args, "-trace")
		}
		runcmd = exec.Command(arch.EmulatorCmd(), args...)
		runcmd.Stderr = os.Stderr
		runcmd.Stdout = finalout
		runcmd.Stdin = os.Stdin
	}

	parser := parser.NewParser(dir, patterns...)
	parser.Scan()

	pkg := parser.Package()

	gen := codegen.NewGenerator(pkg)
	emit := asm.NewEmitter(asmout)

	for fn := parser.NextUnparsedFunc(); fn != nil; fn = parser.NextUnparsedFunc() {
		var w dumper
		w = nopDumper{}
		if *dump != "" && strings.Contains(fn.Name, *dump) {
			w = html.NewHTMLWriter("ssa.html", fn)
			filename, lines, start := parser.DumpOrignalSource(fn)
			w.WriteSources("go", filename, lines, start)
			w.WriteAsmBuf("tools/go/ssa", parser.DumpOriginalSSA(fn))
		}
		defer w.Close()

		parser.ParseFunc(fn)

		w.WritePhase("initial", "initial")

		xform.AddReturnMoves(fn)
		xform.Transform(xform.Elaboration, fn)
		w.WritePhase("elaboration", "elaboration")

		xform.Transform(xform.Simplification, fn)
		w.WritePhase("simplification", "simplification")

		xform.Transform(xform.Lowering, fn)
		w.WritePhase("lowering", "lowering")

		ra := regalloc.NewRegAlloc(fn)

		used := ra.Allocate(fn)
		w.WritePhase("allocation", "allocation")

		ra.Verify()

		xform.Transform(xform.Legalize, fn)
		w.WritePhase("legalize", "legalize")

		xform.Transform(xform.CleanUp, fn)
		w.WritePhase("cleanup", "cleanup")

		xform.ProEpiLogue(used, fn)
		xform.EliminateEmptyBlocks(fn)
		w.WritePhase("final", "final")

		asm := gen.Func(fn)
		w.WriteAsm("asm", asm)
		emit.Func(asm)
	}

	asmout.Close()

	if asmcmd != nil {
		if err := asmcmd.Run(); err != nil {
			os.Exit(1)
		}
		if mode&Run == 0 {
			// todo: read file and emit to finalout
			f, err := os.Open(binfile)
			if err != nil {
				log.Fatal(err)
			}
			_, err = io.Copy(finalout, f)
			f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if runcmd != nil {
		if err := runcmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// don't treat exit errors as an error, instead return the exit code
				return exitErr.ExitCode()
			}
			return 1
		}
		return runcmd.ProcessState.ExitCode()
	}

	return 0
}
