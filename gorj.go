package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rj45/nanogo/codegen"
	"github.com/rj45/nanogo/parser"
)

func main() {
	log.SetFlags(log.Lshortfile)

	mod := parser.ParseModule("./testfiles/seive/seive.go")

	fmt.Println(mod.LongString())

	fmt.Print("\n\n--------------------\n\n")

	codegen.GenerateCode(mod, os.Stdout)
}
