package regalloc2_test

import (
	"testing"

	"github.com/rj45/nanogo/ir2/parseir"
	"github.com/rj45/nanogo/regalloc2"
)

func TestCriticalEdgeFinder_withNoCriticalEdges(t *testing.T) {
	fn, err := parseir.ParseString(`
	.b0:
  		v0:bool = parameter 0
		v1:int = parameter 1
		v2:int = parameter 2
		if v0, .b1, .b2
	.b1:
		jump .b3
	.b2:
		v3:int = add v1, v2
		jump .b3
	.b3:
		return
	`)
	if err != nil {
		t.Error(err)
	}

	ra := regalloc2.NewRegAlloc(fn)

	err = ra.CheckInput()

	// should have no critical edges
	if err != nil {
		t.Error(err)
	}

}

func TestCriticalEdgeFinder_withCriticalEdges(t *testing.T) {
	fn, err := parseir.ParseString(`
	.b0:
  		v0:bool = parameter 0
		v1:int = parameter 1
		v2:int = parameter 2
		if v0, .b3, .b2
	.b2:
		v3:int = add v1, v2
		jump .b3
	.b3:
		return
	`)
	if err != nil {
		t.Error(err)
	}

	ra := regalloc2.NewRegAlloc(fn)

	err = ra.CheckInput()

	// should have critical edges
	if err != regalloc2.ErrCriticalEdges {
		t.Error("expected that the critical edge would be found and reported")
	}
}
