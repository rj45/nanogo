package regalloc2

import "github.com/rj45/nanogo/ir2"

type interval struct {
	val   ir2.ID
	start uint32
	end   uint32
}

type intervalTree struct {
	mid       uint32
	less      *intervalTree
	gtr       *intervalTree
	intervals []*interval
}
