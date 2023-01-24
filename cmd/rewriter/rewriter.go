package main

import (
	"flag"
	"log"
	"os"

	"github.com/rj45/nanogo/rewriter"
)

func main() {
	rulefile := flag.String("i", "translate.rules", "The rule file to generate code from")
	name := flag.String("func", "translate", "The name of the function to generate")
	pkg := flag.String("pkg", "xform", "The name of the package where the code belongs")
	matcher := flag.String("matcher", "matcher", "The name of the matcher struct")
	builder := flag.String("builder", "builder", "The name of the builder struct")
	outfile := flag.String("o", "translate_gen.go", "The name of the Go file to be generated")

	flag.Parse()

	inf, err := os.Open(*rulefile)
	if err != nil {
		log.Fatal(err)
	}
	defer inf.Close()

	rules, err := rewriter.Parse(inf)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	rewriter.GenCode(rules, *name, *pkg, *matcher, *builder, out)
}
