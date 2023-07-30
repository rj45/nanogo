package regalloc2

import (
	"errors"

	"github.com/rj45/nanogo/ir2"
)

var ErrEntryLiveIns = errors.New("entry block has live in values")

// liveInOutScan scans the code from the bottom up to calculate
// the liveIns and liveOuts of each block. This is the first
// step in liveness analysis. LiveIns are values live upon
// entry to the block, liveOuts are live after exit of the
// block.
//
// ErrEntryLiveIns can be returned if the entry block ends up
// with values that are live into the block, indicating a
// malformed program.
//
// The algorithm is essentially to use a work queue to keep
// track of blocks that need to be (re)scanned, and keep
// working on it until a fixpoint is reached (ie. no more
// changes happen). The work queue is initially populated
// with all blocks in reverse order, so the initial scan
// happens from bottom to top. If the liveOut set of a block
// changes and it's not in the work queue, it's added to
// the end of the work queue to be rescanned.
//
// Scanning of a block happens from bottom to top, if a
// value is defined, it's removed from the live set, and
// if a value is used it's added to the live set.
func (ra *RegAlloc) liveInOutScan() error {
	fn := ra.fn

	// set up a work queue
	work := make([]*ir2.Block, 0, fn.NumBlocks())
	inWork := make([]bool, fn.NumBlocks())

	// add all blocks to the queue in reverse
	// order (bottom to top)
	for i := fn.NumBlocks() - 1; i >= 0; i-- {
		blk := fn.Block(i)
		work = append(work, blk)
		inWork[blk.Index()] = true
	}

	// while we have work to do
	for len(work) > 0 {
		// pop a block off the queue
		blk := work[0]
		work = work[1:]
		inWork[blk.Index()] = false

		info := &ra.info[blk.Index()]
		live := map[ir2.ID]struct{}{}
		info.liveIns = live

		// clone liveOuts
		for id := range info.liveOuts {
			live[id] = struct{}{}
		}

		// block args are live just before leaving this block
		for i := 0; i < blk.NumArgs(); i++ {
			arg := blk.Arg(i)
			if arg.NeedsReg() {
				live[arg.ID] = struct{}{}
			}
		}

		// for each instruction in reverse order
		for in := blk.NumInstrs() - 1; in >= 0; in-- {
			instr := blk.Instr(in)

			// for each def
			for i := 0; i < instr.NumDefs(); i++ {
				def := instr.Def(i)

				// remove it from the live set
				delete(live, def.ID)
			}

			// for each use
			for i := 0; i < instr.NumArgs(); i++ {
				use := instr.Arg(i)

				// skip values that don't need registers
				if use.NeedsReg() {
					// add it to the live set
					live[use.ID] = struct{}{}
				}
			}
		}

		// block defs are live just after entering the block
		// so they get removed from the live set
		for i := 0; i < blk.NumDefs(); i++ {
			def := blk.Def(i)
			delete(live, def.ID)
		}

		// copy live ins to the live outs of pred blocks and
		// check if they need to be recomputed because of changes
		for i := 0; i < blk.NumPreds(); i++ {
			pred := blk.Pred(i)
			pinfo := &ra.info[pred.Index()]
			changed := false
			if pinfo.liveOuts == nil {
				pinfo.liveOuts = map[ir2.ID]struct{}{}
			}
			for id := range live {
				if _, found := pinfo.liveOuts[id]; !found {
					changed = true
					pinfo.liveOuts[id] = struct{}{}
				}
			}

			if changed && !inWork[pred.Index()] {
				work = append(work, pred)
				inWork[pred.Index()] = true
			}
		}
	}

	// liveIn set of the entry block should be empty,
	// otherwise a value is used somewhere that never
	// got defined
	if len(ra.info) > 0 && len(ra.info[0].liveIns) > 0 {
		return ErrEntryLiveIns
	}

	return nil
}
