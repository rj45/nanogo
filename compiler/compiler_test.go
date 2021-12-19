package compiler_test

import (
	"bytes"
	"testing"

	"github.com/rj45/nanogo/arch"
	"github.com/rj45/nanogo/compiler"
	"github.com/rj45/nanogo/frontend"
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/parseir"

	// load the supported architectures so they register with the arch package
	_ "github.com/rj45/nanogo/arch/a32"
	_ "github.com/rj45/nanogo/arch/rj32"
)

var testCases = []struct {
	desc     string
	filename string
}{
	{
		desc:     "simple test",
		filename: "./simple/",
	},
	{
		desc:     "hello world",
		filename: "./hello/",
	},
	{
		desc:     "fibonacci",
		filename: "./fib/",
	},
	{
		desc:     "multiply and divide",
		filename: "./muldiv/",
	},
	{
		desc:     "incremental seive of eratosthenes",
		filename: "./seive/",
	},
	{
		desc:     "n queens problem",
		filename: "./nqueens/",
	},
	{
		desc:     "multiple return values",
		filename: "./multireturn/",
	},
	{
		desc:     "iterating and storing strings",
		filename: "./print/",
	},
	{
		desc:     "external assembly",
		filename: "./externasm/",
	},
}

func TestCompilerForRj32(t *testing.T) {
	for _, tC := range testCases {
		t.Run("runs "+tC.desc+" on rj32", func(t *testing.T) {
			arch.SetArch("rj32")
			result := compiler.Compile("-", "../testdata/", []string{tC.filename}, compiler.Assemble|compiler.Run)
			if result != 0 {
				t.Errorf("test %s failed with code %d", tC.filename, result)
			}
		})
	}
}

func TestCompilerForA32(t *testing.T) {
	for _, tC := range testCases {
		t.Run("runs "+tC.desc+" on a32", func(t *testing.T) {
			arch.SetArch("a32")
			result := compiler.Compile("-", "../testdata/", []string{tC.filename}, compiler.Assemble|compiler.Run)
			if result != 0 {
				t.Errorf("test %s failed with code %d", tC.filename, result)
			}
		})
	}
}

func TestCompileToFromIR(t *testing.T) {
	for _, tC := range testCases {
		t.Run("compiles "+tC.desc+" IR identically", func(t *testing.T) {
			arch.SetArch("rj32")

			if tC.filename == "./nqueens/" {
				// todo: in the parsed IR, placeholders must be used for func refs
				// but in the go parser, all globals/funcs are known ahead of time so
				// placeholders aren't needed. Need to figure out a way to not renumber
				// subsequent values because of this.
				t.Skip()
			}

			bufA := &bytes.Buffer{}
			bufB := &bytes.Buffer{}

			// compile the go to IR
			{
				fe := frontend.NewFrontEnd("../testdata/", tC.filename)
				fe.Scan()
				for fn := fe.NextUnparsedFunc(); fn != nil; fn = fe.NextUnparsedFunc() {
					fe.ParseFunc(fn)
				}

				fe.Program().Emit(bufA, ir2.SSAString{})
			}

			// compile the IR
			{
				prog := &ir2.Program{}
				p, err := parseir.NewParser(tC.filename, bytes.NewReader(bufA.Bytes()), prog, false)
				if err != nil {
					t.Fatal(err)
				}

				err = p.Parse()
				if err != nil {
					p.PrintErrors()
					t.Fatal(err)
				}

				prog.Emit(bufB, ir2.SSAString{})
			}

			// check if they are equal
			if !bytes.Equal(bufA.Bytes(), bufB.Bytes()) {
				t.Error("expected equal outputs, was not!")
			}
		})
	}
}
