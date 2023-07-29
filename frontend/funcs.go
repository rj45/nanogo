package frontend

import (
	"github.com/rj45/nanogo/ir/reg"
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
				blkdef := irBlock.Func().NewValue(param.Type())
				irBlock.AddDef(blkdef)

				if i < len(reg.ArgRegs) {
					blkdef.SetReg(reg.ArgRegs[i])
				} else {
					blkdef.SetParamSlot(i - len(reg.ArgRegs))
				}

				instr := irFunc.NewInstr(op.Copy, param.Type(), blkdef)
				irBlock.InsertInstr(-1, instr)

				instr.Pos = getPos(param)
				val := instr.Def(0)

				fe.val2instr[param] = instr
				fe.val2val[param] = val
			}
		}

		fe.blockmap[ssaBlock] = irBlock

		prevCritBlockNum := len(fe.critBlocks) - 1

		// any block that returns has an implicit successor
		extraSucc := 0
		if _, ok := ssaBlock.Instrs[len(ssaBlock.Instrs)-1].(*ssa.Return); ok {
			extraSucc = 1
		}

		// reserve blocks for breaking critical edges
		if (len(ssaBlock.Succs) + extraSucc) > 1 {
			for _, succ := range ssaBlock.Succs {
				// the entry block has an implicit Pred
				extraPred := 0
				if succ.Index == 0 {
					extraPred = 1
				}

				if (len(succ.Preds) + extraPred) > 1 {
					irBlock := irFunc.NewBlock()
					irFunc.InsertBlock(-1, irBlock)

					fe.critBlocks = append(fe.critBlocks, critBlock{
						blk:  irBlock,
						from: ssaBlock,
						to:   succ,
					})
				}
			}
		}

		fe.translateInstrs(irBlock, ssaBlock)

		for i := prevCritBlockNum; i < len(fe.critBlocks) && i >= 0; i++ {
			if fe.critBlocks[i].blk.NumInstrs() > 0 {
				continue
			}
			fe.critBlocks[i].blk.InsertInstr(-1, irFunc.NewInstr(op.Jump, nil))
		}
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

			for _, crit := range fe.critBlocks {
				if crit.from == block && crit.to == succ {
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

			for _, crit := range fe.critBlocks {
				if crit.from == pred && crit.to == block {
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
