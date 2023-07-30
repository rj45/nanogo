// Package verify contains a symbolic verifier to check if registers are allocated correctly
package verify

import (
	"errors"
	"fmt"
	"hash/maphash"

	"github.com/rj45/nanogo/ir/reg"
	"github.com/rj45/nanogo/ir2"
)

var regList []reg.Reg
var regIndex map[reg.Reg]uint8

var ErrNoRegAssigned = errors.New("no register assigned to a variable")
var ErrWrongValueInReg = errors.New("attempt to read wrong value from register")
var ErrMissingCopy = errors.New("missing copy of block parameter")

// Verify executes the function symbolically, tracking which values are
// in which registers. Each block is executed with each permutation of
// input values, which are hashed to ensure the block isn't executed
// twice with the same input parameters, so that it will eventually halt.
//
// This verifier tries not to rely on liveness analysis and tries to
// execute the code from the beginning to the end in order to be very
// different from the way the register allocator works, thus increasing
// the likelihood of catching bugs.
//
// This verifier wasn't really designed to be fast, since it probably
// won't need to run in production. It's mainly for catching bugs in
// tests / development. But if tests end up slow, this could be
// optimized.
func Verify(fn *ir2.Func) []error {
	var errs []error

	if regList == nil {
		regList = append(regList, reg.None)
		regList = append(regList, reg.ArgRegs...)
		regList = append(regList, reg.TempRegs...)
		regList = append(regList, reg.SavedRegs...)

		regIndex = make(map[reg.Reg]uint8, len(regList))
		for idx, reg := range regList {
			regIndex[reg] = uint8(idx)
		}
	}

	if fn.NumBlocks() < 1 {
		return nil
	}

	// calculate the initial live regs for the first block
	firstblk := fn.Block(0)
	firstlive := make([]ir2.ID, len(regList))
	for d := 0; d < firstblk.NumDefs(); d++ {
		arg := firstblk.Def(d)
		firstlive[regIndex[arg.Reg()]] = arg.ID
	}

	// add it to the worklist
	worklist := []struct {
		blk  *ir2.Block
		live []ir2.ID
	}{{firstblk, firstlive}}
	done := map[uint64]struct{}{}

	// mark it as done
	key := make([]byte, len(regList)*2+2)
	done[genKey(firstblk, firstlive, key)] = struct{}{}

	// for each worklist item
	for len(worklist) > 0 {
		// pop a worklist item off the front
		// order doesn't really matter, so the faster way to remove is used
		blk := worklist[0].blk
		live := worklist[0].live
		if len(worklist) > 0 {
			worklist[0] = worklist[len(worklist)-1]
		}
		worklist = worklist[:len(worklist)-1]

		// for each instruction in the block
		for i := 0; i < blk.NumInstrs(); i++ {
			instr := blk.Instr(i)

			// for each arg (use) of the instruction that needs a register
			for a := 0; a < instr.NumArgs(); a++ {
				arg := instr.Arg(a)
				if !arg.NeedsReg() {
					continue
				}

				// check if the arg isn't in a register and report it
				if !arg.InReg() {
					errs = append(errs,
						fmt.Errorf("%w: fn %s blk %s instr %q arg %s", ErrNoRegAssigned, fn.Name, blk, instr, arg))
					continue
				}

				// check the value currently residing in the register, if it doesn't
				// match, then report it
				regidx := regIndex[arg.Reg()]
				if live[regidx] != arg.ID {
					oldval := live[regidx].ValueIn(fn)
					oldstr := "<unk>"
					if oldval != nil {
						oldstr = oldval.IDString()
					}
					errs = append(errs,
						fmt.Errorf("%w: reg %s contains %s but wanted to read %s: fn %s blk %s instr %q arg %s", ErrWrongValueInReg, arg.Reg(), oldstr, arg.IDString(), fn.Name, blk, instr, arg))
				}
			}

			if instr.Op.IsCall() {
				// all arg and temp registers are clobbered in a call
				for _, list := range [][]reg.Reg{reg.ArgRegs, reg.TempRegs} {
					for _, reg := range list {
						regidx := regIndex[reg]
						live[regidx] = 0
					}
				}
			}

			// for each def in the instruction that needs a register
			for d := 0; d < instr.NumDefs(); d++ {
				def := instr.Def(d)
				if !def.NeedsReg() {
					continue
				}

				// if the def isn't in a register, report it
				if !def.InReg() {
					errs = append(errs,
						fmt.Errorf("%w: fn %s blk %s instr %q def %s", ErrNoRegAssigned, fn.Name, blk, instr, def))
					continue
				}

				// update the value in the live set
				regidx := regIndex[def.Reg()]
				live[regidx] = def.ID
			}
		}

		// for each successor block
		nextArgOffset := 0
		for s := 0; s < blk.NumSuccs(); s++ {
			succ := blk.Succ(s)
			argoffset := nextArgOffset
			nextArgOffset += succ.NumDefs()

			// check if the combination of live in and block have
			// already been (or will be) worked on, skip if so
			hashkey := genKey(succ, live, key)
			if _, found := done[hashkey]; found {
				continue
			}

			succlive := make([]ir2.ID, len(regList))
			copy(succlive, live)

			// write all to zero before we write the actual ones so we don't clobber them
			for d := 0; d < succ.NumDefs(); d++ {
				def := succ.Def(d)
				arg := blk.Arg(argoffset + d)

				if def.Reg() != arg.Reg() {
					// todo: when blk parameter copies are implemented uncomment this
					errs = append(errs,
						fmt.Errorf("%w: fn %s from blk %s to blk %s: from arg %s to def %s", ErrMissingCopy, fn.Name, blk, succ, arg, def))
					regidx := regIndex[arg.Reg()]
					succlive[regidx] = 0
				}
			}

			for d := 0; d < succ.NumDefs(); d++ {
				def := succ.Def(d)

				regidx := regIndex[def.Reg()]
				succlive[regidx] = def.ID
			}

			// add it to the worklist and mark it as having been added
			done[hashkey] = struct{}{}
			worklist = append(worklist, struct {
				blk  *ir2.Block
				live []ir2.ID
			}{
				blk:  succ,
				live: succlive,
			})
		}
	}

	return errs
}

var seed = maphash.MakeSeed()

func genKey(blk *ir2.Block, live []ir2.ID, key []byte) uint64 {
	key[0] = byte(blk.ID >> 8)
	key[1] = byte(blk.ID)
	for i := 0; i < len(live); i++ {
		key[(i*2)+2+0] = byte(live[i] >> 8)
		key[(i*2)+2+1] = byte(live[i])
	}
	return maphash.Bytes(seed, key)
}
