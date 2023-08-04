package frontend

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"golang.org/x/tools/go/ssa"
)

func handleExternAssembly(function *ir2.Func, fn *ssa.Function) {
	filename := fn.Prog.Fset.File(fn.Pos()).Name()
	folder, err := filepath.EvalSymlinks(filepath.Dir(filename))
	if err != nil {
		log.Fatalf("could not follow symlinks for folder %s", folder)
	}
	asm := ""
	filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		ext := filepath.Ext(d.Name())
		if ext == ".asm" || ext == ".s" || ext == ".S" {
			noext := strings.TrimSuffix(d.Name(), ext)
			parts := strings.Split(noext, "_")

			if len(parts) > 1 && parts[len(parts)-1] != archName {
				// skip files with an underscore and the last part of the name
				// does not match the archName
				return nil
			}

			buf, err := os.ReadFile(path)
			if err != nil {
				log.Fatalln(err)
			}

			// TODO: find build tags and ensure they match

			lines := bytes.Split(buf, []byte("\n"))
			startLine := -1
			label := []byte(fmt.Sprintf("%s:", fn.Name()))
			for i, line := range lines {
				if bytes.HasPrefix(bytes.TrimSpace(line), label) {
					startLine = i + 1
					break
				}
			}

			if startLine == -1 {
				return nil
			}

			endLine := -1
			for i := startLine; i < len(lines); i++ {
				trimmed := bytes.TrimSpace(lines[i])
				lines[i] = trimmed
				// if doesn't start with a dot, but does end in a colon
				if !bytes.HasPrefix(trimmed, []byte(".")) && bytes.HasSuffix(trimmed, []byte(":")) {
					endLine = i + 1
					break
				}
			}

			if endLine == -1 {
				endLine = len(lines)
			}

			if asm != "" {
				log.Fatalf("found duplicate of extern func %s in %s", fn.Name(), path)
			}

			asm = string(bytes.Join(lines[startLine:endLine], []byte("\n")))
		}
		return nil
	})
	if asm == "" {
		log.Fatalf("could not find assembly for extern func %s path %s", fn.Name(), folder)
	}
	genInlineAsmFunc(function, asm)
}

func genInlineAsmFunc(fn *ir2.Func, asm string) {
	entry := fn.NewBlock()
	body := fn.NewBlock()
	exit := fn.NewBlock()

	fn.InsertBlock(-1, entry)
	fn.InsertBlock(-1, body)
	fn.InsertBlock(-1, exit)

	entry.InsertInstr(-1, fn.NewInstr(op.Jump, nil))
	entry.AddSucc(body)

	body.AddPred(entry)
	instr := fn.NewInstr(op.InlineAsm, nil, asm)
	body.InsertInstr(-1, instr)
	body.InsertInstr(-1, fn.NewInstr(op.Jump, nil))
	body.AddSucc(exit)

	exit.AddPred(body)
	exit.InsertInstr(-1, fn.NewInstr(op.Return, nil))

	// todo: add defs and args to instr
}
