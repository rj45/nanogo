package regalloc2

import (
	"errors"

	"github.com/rj45/nanogo/ir/reg"
	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
)

type RegAlloc struct {
	fn *ir2.Func

	info []blockInfo

	iGraph iGraph
}

type blockInfo struct {
	liveIns  map[ir2.ID]struct{}
	liveOuts map[ir2.ID]struct{}
}

func NewRegAlloc(fn *ir2.Func) *RegAlloc {
	info := make([]blockInfo, fn.NumBlocks())

	return &RegAlloc{
		fn:   fn,
		info: info,
	}
}

var ErrCriticalEdges = errors.New("the CFG has critical edges")

// CheckInput will verify the structure of the
// input code, which is useful in tests and fuzzing.
func (ra *RegAlloc) CheckInput() error {
	fn := ra.fn

	for i := 0; i < fn.NumBlocks(); i++ {
		blk := fn.Block(i)

		if fn.Block(0).NumInstrs() < 1 {
			// skip empty blocks
			continue
		}

		// any block that returns has an implicit successor
		extraSucc := 0
		if blk.Instr(blk.NumInstrs()-1).Op == op.Return {
			extraSucc = 1
		}

		if (blk.NumSuccs() + extraSucc) > 1 {
			for i := 0; i < blk.NumSuccs(); i++ {
				succ := blk.Succ(i)

				// the first block has an implicit pred
				extraPred := 0
				if succ.Index() == 0 {
					extraPred = 1
				}

				if (succ.NumPreds() + extraPred) > 1 {
					return ErrCriticalEdges
				}
			}
		}
	}
	return nil
}

// Allocate will run the allocator and assign a physical
// register or stack slot to each Value that needs one.
func (ra *RegAlloc) Allocate() error {
	if err := ra.liveInOutScan(); err != nil {
		return err
	}

	ra.buildInterferenceGraph()
	ra.iGraph.pickColours()
	ra.assignRegisters()

	return nil
}

var regList []reg.Reg

func (ra *RegAlloc) assignRegisters() bool {
	if regList == nil {
		regList = append(regList, reg.ArgRegs...)
		regList = append(regList, reg.TempRegs...)
		regList = append(regList, reg.SavedRegs...)
	}

	for id := range ra.iGraph.nodes {
		node := &ra.iGraph.nodes[id]

		// colour zero is "noColour", so subtract one
		regIndex := int(node.colour - 1)

		if regIndex >= len(regList) {
			panic("failed to assign registers!")
			// return false
		}

		val := node.val.ValueIn(ra.fn)
		val.SetReg(regList[regIndex])
	}

	return true
}
