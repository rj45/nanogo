package regalloc2

import (
	"errors"
	"fmt"

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
			if !arg.NeedsReg() {
				continue
			}
			live[arg.ID] = struct{}{}
		}

		// for each instruction in reverse order
		it := blk.InstrIter()
		it.Last()
		for ; it.HasPrev(); it.Prev() {
			instr := it.Instr()

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
				if !use.NeedsReg() {
					continue
				}

				// add it to the live set
				live[use.ID] = struct{}{}
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
				fmt.Println(work)
			}
		}
	}

	// liveIn set of the entry block should be empty,
	// otherwise a value is used somewhere that never
	// got defined
	if len(ra.info[0].liveIns) > 0 {
		return ErrEntryLiveIns
	}

	return nil
}

// buildBlockRanges figures out the program points for
// the start and end of each block
func (ra *RegAlloc) buildBlockRanges() {
	fn := ra.fn
	count := uint32(0)
	for i := 0; i < fn.NumBlocks(); i++ {
		blk := fn.Block(i)
		info := &ra.info[blk.Index()]

		info.start = count
		count += uint32(blk.NumInstrs() * 2)
		info.end = count
	}
}

// buildLiveRanges builds live ranges (intervals) for
// each variable that needs a register by scanning backwards
// from the bottom and adding and removing variables from
// the live set as they are used and defined respectively.
// The live set tracks which intervals are "open" whose
// start (def) has not yet been seen.
func (ra *RegAlloc) buildLiveRanges() {
	fn := ra.fn
	ranges := []interval{}
	live := map[ir2.ID]*interval{}

	// addLive adds a variable which needs a register
	// into the live set, keeping track of the end, and
	// setting the start to a guess.
	addLive := func(val *ir2.Value, start, end uint32) {
		if !val.NeedsReg() {
			return
		}
		if _, found := live[val.ID]; !found {
			ranges = append(ranges, interval{
				val: val.ID,
				end: end,
			})
			live[val.ID] = &ranges[len(ranges)-1]
		}
		live[val.ID].start = start
	}

	// remLive removes a variable from the live set, setting
	// the start of the variable's live range to `start`
	remLive := func(val *ir2.Value, start uint32) {
		if !val.NeedsReg() {
			return
		}
		live[val.ID].start = start
		delete(live, val.ID)
	}

	// for each block in reverse order
	for i := fn.NumBlocks() - 1; i >= 0; i-- {
		blk := fn.Block(i)
		info := &ra.info[blk.Index()]

		progPoint := info.end

		// ensure the liveOuts of the block are in the live set
		for id := range info.liveOuts {
			addLive(id.ValueIn(fn), info.start, info.end)
		}

		// args to successor blocks are live until just before
		// the end of the block
		for a := 0; a < blk.NumArgs(); a++ {
			arg := blk.Arg(a)

			// todo: maybe should be `info.end-1`?
			addLive(arg, info.start, info.end)
		}

		// for each instruction in the block in reverse order
		for j := blk.NumInstrs() - 1; j >= 0; j-- {
			instr := blk.Instr(j)

			// todo: unsure where these decrements go
			progPoint--

			for d := 0; d < instr.NumDefs(); d++ {
				def := instr.Def(d)

				remLive(def, progPoint)
			}

			progPoint--

			for a := 0; a < instr.NumArgs(); a++ {
				use := instr.Arg(a)

				addLive(use, info.start, progPoint)
			}
		}

		for d := 0; d < blk.NumDefs(); d++ {
			def := blk.Def(d)
			remLive(def, info.start)
		}
	}

	if len(live) > 0 {
		panic("leftover live values")
	}

	ra.liveRanges = ranges
}
