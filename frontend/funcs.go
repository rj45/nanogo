package frontend

import (
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
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

	for bn, ssaBlock := range blockList {
		irBlock := irFunc.NewBlock()

		irFunc.InsertBlock(-1, irBlock)

		if bn == 0 {
			for i, param := range ssaFunc.Params {
				instr := irFunc.NewInstr(op.Parameter, param.Type(), i)
				irBlock.InsertInstr(-1, instr)

				instr.Pos = getPos(param)

				fe.val2instr[param] = instr
				fe.val2val[param] = instr.Def(0)
			}
		}

		fe.blockmap[ssaBlock] = irBlock

		fe.translateInstrs(irBlock, ssaBlock)

		// for _, succ := range ssaBlock.Succs {
		// 	if len(ssaBlock.Succs) > 1 && len(succ.Preds) > 1 {
		// 		irBlock := irFunc.NewBlock()
		// 		irFunc.InsertBlock(-1, irBlock)

		// 		irBlock.InsertInstr(-1, irFunc.NewInstr(op.Jump, nil))
		// 	}
		// }
	}

	// todo: if this panic never happens, maybe placeholders not necessary?
	if irFunc.HasPlaceholders() {
		panic("Invalid assumption that placeholders no longer necessary")
	}

	fe.resolvePlaceholders(irFunc)

	for _, block := range blockList {
		irBlock := fe.blockmap[block]

		for _, succ := range block.Succs {
			found := false

			if !found {
				irBlock.AddSucc(fe.blockmap[succ])
			}
		}
		for _, pred := range block.Preds {
			found := false

			if !found {
				irBlock.AddPred(fe.blockmap[pred])
			}
		}
	}
}

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
