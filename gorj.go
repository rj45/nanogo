package main

import (
	"fmt"
	"log"

	"github.com/rj45/nanogo/parser"
)

func main() {
	log.SetFlags(log.Lshortfile)

	mod := parser.ParseModule("./testfiles/seive/seive.go")

	fmt.Println(mod.LongString())
}
