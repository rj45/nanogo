// Copyright (c) 2021 rj45 (github.com/rj45), MIT Licensed, see LICENSE.

package regalloc

import (
	"flag"
	"log"

	"github.com/rj45/nanogo/ir"
	"github.com/rj45/nanogo/ir/op"
	"github.com/rj45/nanogo/ir/reg"
)

var debugColour = flag.Bool("debugcolour", false, "emit register allocation colouring logs")

func (ra *RegAlloc) colour() {
	// ra.wrongGuesses = make(map[*ir.Value]bool)

	ra.Func.Blocks()[0].VisitSuccessors(ra.allocateBlock)

	// ra.guessedRegs = ra.wrongGuesses

	// // reallocate blocks with guessed regs
	// for len(ra.guessedRegs) > 0 {
	// 	visited := map[*ir.Block]bool{}

	// 	ra.wrongGuesses = make(map[*ir.Value]bool)

	// 	for val := range ra.guessedRegs {
	// 		blk := val.Block()
	// 		if visited[blk] {
	// 			continue
	// 		}
	// 		visited[blk] = true

	// 		for i := 0; i < blk.NumInstrs(); i++ {
	// 			val := blk.Instr(i)
	// 			if !ra.guessedRegs[val] {
	// 				val.Reg = reg.None
	// 			}
	// 		}

	// 		ra.allocateBlock(blk)
	// 	}

	// 	ra.guessedRegs = ra.wrongGuesses
	// }
}

func (ra *RegAlloc) allocateBlock(blk *ir.Block) bool {
	info := &ra.blockInfo[blk.ID()]
	var used reg.Reg

	for val := range info.liveIns {
		if val.Reg != reg.None {
			used |= val.Reg
			info.regValues[val.Reg] = val
		}
	}

	phiStart := -1

	for i := 0; i < blk.NumInstrs(); i++ {
		val := blk.Instr(i)

		if val.Op == op.PhiCopy || val.Op == op.Phi {
			if phiStart < 0 {
				phiStart = i
			}
			continue
		} else if phiStart >= 0 {
			// process phis
			used = ra.allocateParallelCopies(info, used, blk, phiStart, i, info.regValues)
			ra.usedRegs |= used
			phiStart = -1
		}

		used = ra.allocateValue(info, val, used, blk, info.regValues)
		ra.usedRegs |= val.Reg
	}

	if phiStart >= 0 {
		// process phis
		used = ra.allocateParallelCopies(info, used, blk, phiStart, blk.NumInstrs(), info.regValues)
		ra.usedRegs |= used
	}

	if *debugColour {
		log.Println("{------------------")
		log.Print(blk.LongString())
		log.Println("}------------------")
	}

	return true
}

func (ra *RegAlloc) allocateValue(info *blockInfo, val *ir.Value, used reg.Reg, blk *ir.Block, regValues map[reg.Reg]*ir.Value) reg.Reg {
	used = ra.processKills(info, val, used, regValues)

	if !val.NeedsReg() {
		return used
	}

	ra.assignRegister(val, info, used, blk)

	used = ra.recordAssignment(used, val, regValues)

	return used
}

// allocateParallelCopies simulates parallel copies by splitting the
// killing and assigning phases
func (ra *RegAlloc) allocateParallelCopies(info *blockInfo, used reg.Reg, blk *ir.Block, start, end int, regValues map[reg.Reg]*ir.Value) reg.Reg {
	for i := start; i < end; i++ {
		val := blk.Instr(i)
		used = ra.processKills(info, val, used, regValues)
	}

	for i := start; i < end; i++ {
		val := blk.Instr(i)
		ra.assignRegister(val, info, used, blk)
		used = ra.recordAssignment(used, val, regValues)
	}

	return used
}

func (*RegAlloc) recordAssignment(used reg.Reg, val *ir.Value, regValues map[reg.Reg]*ir.Value) reg.Reg {
	used |= val.Reg
	regValues[val.Reg] = val
	return used
}

func (ra *RegAlloc) assignRegister(val *ir.Value, info *blockInfo, used reg.Reg, blk *ir.Block) {
	if val.Reg == reg.None {
		val.Reg = ra.chooseReg(info, val, used)
	}

	if val.Reg == reg.None {
		log.Println(blk.LongString())
		log.Panicln("Ran out of registers in", ra.Func, ", spilling not implemented")
	}
}

func (*RegAlloc) processKills(info *blockInfo, val *ir.Value, used reg.Reg, regValues map[reg.Reg]*ir.Value) reg.Reg {
	for _, kill := range info.kills[val] {
		used &^= kill.Reg
		delete(regValues, kill.Reg)
	}
	return used
}

func (ra *RegAlloc) chooseReg(info *blockInfo, val *ir.Value, used reg.Reg) reg.Reg {
	liveThroughCalls := ra.liveThroughCalls[val]

	// a phi must have the same register assigned to itself and all args
	if val.Op == op.Phi {
		for i := 0; i < val.NumArgs(); i++ {
			arg := val.Arg(i)
			if arg.Reg != reg.None {
				if *debugColour {
					log.Println("Assigning", val.IDString(), "to register", arg.Reg, "from PhiCopy", arg.IDString())
				}
				return arg.Reg
			}
			liveThroughCalls = liveThroughCalls || ra.liveThroughCalls[arg]
		}
	}

	if val.Op == op.PhiCopy {
		// should have one use, which is the phi
		phi := val.ArgUse(0)
		if phi.Op != op.Phi {
			log.Panicf("expecting %s to be a phi!", phi.String())
		}

		// if the phi already has a reg, go with that
		if phi.Reg != reg.None {
			if *debugColour {
				log.Println("Assigning", val.IDString(), "to register", phi.Reg, "from Phi", phi.IDString())
			}
			return phi.Reg
		}

		// otherwise scan the phi's args and run with the first reg assigned
		for i := 0; i < phi.NumArgs(); i++ {
			arg := phi.Arg(i)
			if arg.Reg != reg.None {
				if *debugColour {
					log.Println("Assigning", val.IDString(), "to register", arg.Reg, "from PhiCopy", arg.IDString(), "from Phi", phi.IDString())
				}
				return arg.Reg
			}
			liveThroughCalls = liveThroughCalls || ra.liveThroughCalls[arg]
		}

		liveThroughCalls = liveThroughCalls || ra.liveThroughCalls[phi]
	}

	// check if this is a copy
	if val.Op.IsCopy() && val.NumArgs() == 1 {
		arg := val.Arg(0)

		// check if the copy's arg is killed
		if len(info.kills[val]) > 0 && info.kills[val][0] == arg {
			// if so, if the arg has a register already and using it is safe
			if arg.Reg != reg.None && (!liveThroughCalls || arg.Reg.IsSavedReg()) {
				if *debugColour {
					log.Println("Assigning", val.IDString(), "to register", arg.Reg, "from copied", arg.IDString())
				}

				return arg.Reg
			}
		}
	}

	// check all uses of this value
	// if they are all copies, try to pick the same register
	if val.NumBlockUses() == 0 {
		allCopies := true
		regs := map[reg.Reg]int{}
		for i := 0; i < val.NumArgUses(); i++ {
			use := val.ArgUse(i)

			if !use.Op.IsCopy() {
				allCopies = false
				break
			}

			callSafe := !liveThroughCalls || use.Reg.IsSavedReg()
			if use.Reg != reg.None && (use.Reg&used) == 0 && callSafe {
				regs[use.Reg]++
			} else if use.Op == op.PhiCopy {
				// if it's a phi copy check if the phi has a register
				phi := use.ArgUse(0)
				callSafe = !liveThroughCalls || phi.Reg.IsSavedReg()
				if phi.Reg != reg.None && (phi.Reg&used) == 0 && callSafe {
					regs[phi.Reg]++
				}
			}
		}

		if allCopies {
			top := -1
			choice := reg.None
			for reg, count := range regs {
				if count > top {
					top = count
					choice = reg
				}
			}
			if choice != reg.None {
				if *debugColour {
					log.Println("Assigning", val.IDString(), "to register", choice, "from choices", regs)
				}
				return choice
			}
		}
	}

	sets := [][]reg.Reg{reg.TempRegs, reg.ArgRegs, reg.RevSavedRegs}

	// if the value is live through a call site, then restrict the registers
	// allowed to just saved ones for now
	if liveThroughCalls {
		sets = [][]reg.Reg{reg.SavedRegs}
	}

	for _, set := range sets {
		for _, reg := range set {
			if (used & reg) == 0 {
				if *debugColour {
					log.Println("Assigning", val.IDString(), "to register", reg, "as it was free")
				}
				return reg
			}
		}
	}

	return reg.None
}
