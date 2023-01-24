package frontend

import (
	"log"

	"github.com/rj45/nanogo/ir2"
	"golang.org/x/tools/go/ssa"
)

func (fe *FrontEnd) translateBlockParams(irBlock *ir2.Block, phi *ssa.Phi) {
	param := irBlock.Func().NewValue(phi.Type())
	irBlock.AddDef(param)

	fe.val2val[phi] = param
}

func (fe *FrontEnd) translateBlockArgs(irBlock *ir2.Block, ssaBlock *ssa.BasicBlock) {
	// for each succ block
	for _, succ := range ssaBlock.Succs {
		pred := 0
		for i, p := range succ.Preds {
			if p == ssaBlock {
				pred = i
			}
		}

		if succ.Preds[pred] != ssaBlock {
			panic("not found")
		}

		// scan through each phi in that succ block
		for _, instr := range succ.Instrs {
			phi, ok := instr.(*ssa.Phi)
			if !ok {
				break
			}

			// pick out the arg for the current pred block
			ssaVal := phi.Edges[pred]
			arg := fe.val2val[ssaVal]

			if con, ok := ssaVal.(*ssa.Const); ok {
				arg = irBlock.Func().ValueFor(phi.Type(), con.Value)
			}

			if arg == nil {
				log.Panicf("Can't find val for %s %T in phi %s in block %s", ssaVal, ssaVal, phi, irBlock)
			}

			found := false

			for _, crit := range fe.critBlocks {
				if crit.from == ssaBlock && crit.to == succ {
					found = true

					// interArg := irBlock.Func().NewValue(phi.Type())

					// crit.blk.AddDef(interArg)
					crit.blk.InsertArg(-1, arg)
					// arg = interArg
				}
			}

			if !found {
				irBlock.InsertArg(-1, arg)
			}
		}
	}
}
