package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rj45/nanogo/codegen"
	"github.com/rj45/nanogo/parser"
	"github.com/rj45/nanogo/regalloc"
	"github.com/rj45/nanogo/xform"
)

func main() {
	log.SetFlags(log.Lshortfile)

	mod := parser.ParseModule("./testfiles/seive/seive.go")

	fmt.Println(mod.LongString())

	for _, fn := range mod.Funcs {
		xform.Transform(xform.FirstPass, fn)
		regalloc.Allocate(fn)
		xform.Transform(xform.LastPass, fn)
	}

	fmt.Print("\n\n--------------------\n\n")

	codegen.GenerateCode(mod, os.Stdout)
}
