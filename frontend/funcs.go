package frontend

import (
	"fmt"
	"log"
	"strings"

	"github.com/rj45/nanogo/ir/op"
	"github.com/rj45/nanogo/ir2"
	"golang.org/x/tools/go/ssa"
)

func (fe *FrontEnd) translateFunc(irFunc *ir2.Func, ssaFunc *ssa.Function) {
	if ssaFunc.Blocks == nil {
		// extern function
		// handleExternFunc(irFunc, ssaFunc)
		return
	}

	// order blocks by reverse succession
	blockList := reverseSSASuccessorSort(ssaFunc.Blocks[0], nil, make(map[*ssa.BasicBlock]bool))

	// reverse it to get succession ordering
	for i, j := 0, len(blockList)-1; i < j; i, j = i+1, j-1 {
		blockList[i], blockList[j] = blockList[j], blockList[i]
	}

	type critical struct {
		pred *ssa.BasicBlock
		succ *ssa.BasicBlock
		blk  *ir2.Block
	}
	var criticals []critical
	var returns []*ir2.Block

	for bn, ssaBlock := range blockList {
		irBlock := irFunc.NewBlock()

		irFunc.InsertBlock(-1, irBlock)

		if bn == 0 {
			for _, param := range ssaFunc.Params {
				instr := irFunc.NewInstr(op.Parameter, param.Type(), param.Name())
				irBlock.InsertInstr(-1, instr)

				// todo: fixme for multiple defs
				fe.val2instr[param] = instr
			}
		}

		fe.blockmap[ssaBlock] = irBlock

		// walkInstrs(irBlock, block.Instrs, valmap, storemap)
		fe.translateInstrs(irBlock, ssaBlock)

		lastInstr := ssaBlock.Instrs[len(ssaBlock.Instrs)-1]
		if _, ok := lastInstr.(*ssa.Return); ok {
			returns = append(returns, irBlock)
		}

		for _, succ := range ssaBlock.Succs {
			if len(ssaBlock.Succs) > 1 && len(succ.Preds) > 1 {
				irBlock := irFunc.NewBlock()
				irFunc.InsertBlock(-1, irBlock)

				irBlock.InsertInstr(-1, irFunc.NewInstr(op.Jump2, nil))

				criticals = append(criticals, critical{
					pred: ssaBlock,
					succ: succ,
					blk:  irBlock,
				})
			}
		}
	}

	for _, block := range blockList {
		irBlock := fe.blockmap[block]

		for it := irBlock.InstrIter(); it.HasNext(); it.Next() {
			if ssaInstr, ok := fe.instrmap[it.Instr()]; ok {
				fe.translateArgs(it, ssaInstr)
			} else if it.Instr().Op != op.Parameter {
				log.Panicf("missing instruction: %v : %d/%d, %d/%d", it.Instr(),
					it.BlockIndex(), irFunc.NumBlocks(),
					it.InstrIndex(), it.Block().NumInstrs())
			}
		}

		for _, succ := range block.Succs {
			found := false
			for _, crit := range criticals {
				if crit.pred == block && crit.succ == succ {
					irBlock.AddSucc(crit.blk)
					crit.blk.AddPred(irBlock)
					found = true
					break
				}
			}

			if !found {
				irBlock.AddSucc(fe.blockmap[succ])
			}
		}
		for _, pred := range block.Preds {
			found := false
			for _, crit := range criticals {
				if crit.pred == pred && crit.succ == block {
					irBlock.AddPred(crit.blk)
					crit.blk.AddSucc(irBlock)
					found = true
					break
				}
			}

			if !found {
				irBlock.AddPred(fe.blockmap[pred])
			}
		}
	}

	if len(returns) > 1 {
		realRet := irFunc.NewBlock()
		irFunc.InsertBlock(-1, realRet)

		// var phis []*ir.Value

		// for i := 0; i < returns[0].NumControls(); i++ {
		// 	phi := irFunc.NewValue(op.Phi, returns[0].Control(i).Type)
		// 	realRet.InsertInstr(-1, phi)
		// 	realRet.InsertControl(-1, phi)
		// 	phis = append(phis, phi)
		// }

		for _, ret := range returns {
			ret.AddSucc(realRet)
			realRet.AddPred(ret)
			// ret.Op = op.Jump

			// for i := 0; i < ret.NumControls(); i++ {
			// 	phis[i].InsertArg(-1, ret.Control(i))
			// }

			// for ret.NumControls() > 0 {
			// 	ret.RemoveControl(0)
			// }
		}
	}
}

func genName(pkg, name string) string {
	sname := strings.Replace(name, "$", "_", -1)
	return fmt.Sprintf("%s__%s", pkg, sname)
}

// func walkFunc(function *ir.Func, fn *ssa.Function) {

// 	for _, block := range blockList {
// 		irBlock := blockmap[block]
// 		for _, succ := range block.Succs {
// 			found := false
// 			for _, crit := range criticals {
// 				if crit.pred == block && crit.succ == succ {
// 					irBlock.AddSucc(crit.blk)
// 					crit.blk.AddPred(irBlock)
// 					found = true
// 					break
// 				}
// 			}

// 			if !found {
// 				irBlock.AddSucc(blockmap[succ])
// 			}
// 		}
// 		for _, pred := range block.Preds {
// 			found := false
// 			for _, crit := range criticals {
// 				if crit.pred == pred && crit.succ == block {
// 					irBlock.AddPred(crit.blk)
// 					crit.blk.AddSucc(irBlock)
// 					found = true
// 					break
// 				}
// 			}

// 			if !found {
// 				irBlock.AddPred(blockmap[pred])
// 			}
// 		}

// 		irBlock.Idom = blockmap[block.Idom()]

// 		for _, dom := range block.Dominees() {
// 			irBlock.Dominees = append(irBlock.Dominees, blockmap[dom])
// 		}

// 		if irBlock.Op != op.Jump {
// 			irBlock.SetControls(getArgs(irBlock, block.Instrs[len(block.Instrs)-1], valmap))
// 		}

// 		var linelist []token.Pos

// 		// do a pass to resolve args
// 		for i, instr := range block.Instrs {
// 			pos := getPos(instr)
// 			linelist = append(linelist, pos)

// 			// skip the last op if the block has a op other than jump
// 			if i == (len(block.Instrs)-1) && irBlock.Op != op.Jump {
// 				continue
// 			}

// 			args := getArgs(irBlock, instr, valmap)
// 			var irVal *ir.Value
// 			if len(args) > 0 {
// 				if val, ok := instr.(ssa.Value); ok {
// 					irVal = valmap[val]
// 				} else if val, ok := instr.(*ssa.Store); ok {
// 					irVal = storemap[val]
// 				} else if _, ok := instr.(*ssa.DebugRef); ok {
// 					continue
// 				} else {
// 					log.Fatalf("can't look up args for %#v", instr)
// 				}
// 				for _, arg := range args {
// 					irVal.InsertArg(-1, arg)
// 				}

// 				// double check everything was wired up correctly
// 				var foundVal *ir.Value
// 				for j := 0; j < irBlock.NumInstrs(); j++ {
// 					val := irBlock.Instr(j)
// 					if val == irVal {
// 						foundVal = val
// 					}
// 				}
// 				if foundVal == nil {
// 					log.Fatalf("val not found! %s", irVal.LongString())
// 				}
// 			}
// 		}

// 		fset := fn.Prog.Fset
// 		var lines []string
// 		filename := ""
// 		src := ""
// 		lastline := 0
// 		for _, pos := range linelist {
// 			if pos != token.NoPos {
// 				position := fset.PositionFor(pos, true)
// 				if filename != position.Filename {
// 					filename = position.Filename
// 					buf, err := os.ReadFile(position.Filename)
// 					if err != nil {
// 						log.Fatal(err)
// 					}
// 					lines = strings.Split(string(buf), "\n")
// 				}

// 				if position.Line != lastline {
// 					lastline = position.Line
// 					src += strings.TrimSpace(lines[position.Line-1]) + "\n"
// 				}
// 			}
// 		}
// 		if filename != "" {
// 			irBlock.Source = src
// 		}
// 	}

// 	if len(returns) > 1 {
// 		realRet := function.NewBlock(ir.Block{
// 			Op:      op.Return,
// 			Comment: "real.return",
// 		})
// 		function.InsertBlock(-1, realRet)

// 		var phis []*ir.Value

// 		for i := 0; i < returns[0].NumControls(); i++ {
// 			phi := function.NewValue(op.Phi, returns[0].Control(i).Type)
// 			realRet.InsertInstr(-1, phi)
// 			realRet.InsertControl(-1, phi)
// 			phis = append(phis, phi)
// 		}

// 		for _, ret := range returns {
// 			ret.AddSucc(realRet)
// 			realRet.AddPred(ret)
// 			ret.Op = op.Jump

// 			for i := 0; i < ret.NumControls(); i++ {
// 				phis[i].InsertArg(-1, ret.Control(i))
// 			}

// 			for ret.NumControls() > 0 {
// 				ret.RemoveControl(0)
// 			}
// 		}
// 	}
// }

// func handleExternFunc(function *ir.Func, fn *ssa.Function) {
// 	filename := fn.Prog.Fset.File(fn.Pos()).Name()
// 	folder, err := filepath.EvalSymlinks(filepath.Dir(filename))
// 	if err != nil {
// 		log.Fatalf("could not follow symlinks for folder %s", folder)
// 	}
// 	asm := ""
// 	filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
// 		ext := filepath.Ext(d.Name())
// 		if ext == ".asm" || ext == ".s" || ext == ".S" {
// 			noext := strings.TrimSuffix(d.Name(), ext)
// 			parts := strings.Split(noext, "_")

// 			if len(parts) > 1 && parts[len(parts)-1] != arch.Name() {
// 				// skip files with an underscore and the last part of the name
// 				// does not match the arch.Name()
// 				return nil
// 			}

// 			buf, err := os.ReadFile(path)
// 			if err != nil {
// 				log.Fatalln(err)
// 			}

// 			// TODO: find build tags and ensure they match

// 			lines := bytes.Split(buf, []byte("\n"))
// 			startLine := -1
// 			label := []byte(fmt.Sprintf("%s:", fn.Name()))
// 			for i, line := range lines {
// 				if bytes.HasPrefix(bytes.TrimSpace(line), label) {
// 					startLine = i + 1
// 					break
// 				}
// 			}

// 			if startLine == -1 {
// 				return nil
// 			}

// 			endLine := -1
// 			for i := startLine; i < len(lines); i++ {
// 				trimmed := bytes.TrimSpace(lines[i])
// 				lines[i] = trimmed
// 				// if doesn't start with a dot, but does end in a colon
// 				if !bytes.HasPrefix(trimmed, []byte(".")) && bytes.HasSuffix(trimmed, []byte(":")) {
// 					endLine = i + 1
// 					break
// 				}
// 			}

// 			if endLine == -1 {
// 				endLine = len(lines)
// 			}

// 			if asm != "" {
// 				log.Fatalf("found duplicate of extern func %s in %s", fn.Name(), path)
// 			}

// 			asm = string(bytes.Join(lines[startLine:endLine], []byte("\n")))
// 		}
// 		return nil
// 	})
// 	if asm == "" {
// 		log.Fatalf("could not find assembly for extern func %s path %s", fn.Name(), folder)
// 	}
// 	genInlineAsmFunc(function, asm)
// }

// func genInlineAsmFunc(fn *ir.Func, asm string) {
// 	entry := fn.NewBlock(ir.Block{
// 		Comment: "entry",
// 		Op:      op.Jump,
// 	})
// 	body := fn.NewBlock(ir.Block{
// 		Comment: "inline.asm",
// 		Op:      op.Jump,
// 	})
// 	exit := fn.NewBlock(ir.Block{
// 		Comment: "exit",
// 		Op:      op.Return,
// 	})

// 	entry.AddSucc(body)
// 	body.AddSucc(exit)

// 	exit.AddPred(body)
// 	body.AddPred(entry)

// 	fn.InsertBlock(-1, entry)
// 	fn.InsertBlock(-1, body)
// 	fn.InsertBlock(-1, exit)
// 	blk := fn.Blocks()[1]

// 	val := fn.NewValue(op.InlineAsm, fn.Type.Results())
// 	val.Value = constant.MakeString(asm)
// 	blk.InsertInstr(-1, val)
// }

func reverseSSASuccessorSort(block *ssa.BasicBlock, list []*ssa.BasicBlock, visited map[*ssa.BasicBlock]bool) []*ssa.BasicBlock {
	visited[block] = true

	for i := len(block.Succs) - 1; i >= 0; i-- {
		succ := block.Succs[i]
		if !visited[succ] {
			list = reverseSSASuccessorSort(succ, list, visited)
		}
	}

	return append(list, block)
}
