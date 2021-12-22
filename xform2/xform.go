package xform2

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/rj45/nanogo/ir2"
)

type Pass int

const (
	Elaboration Pass = iota
	Simplification
	Lowering
	Legalize
	CleanUp

	NumPasses
)

type desc struct {
	name     string
	passes   []Pass
	tags     []Tag
	op       ir2.Op
	once     bool
	disabled bool
	fn       func(ir2.Iter)
}

type Option func(d *desc)

func OnlyPass(p Pass) Option {
	return func(d *desc) {
		d.passes = []Pass{p}
	}
}

func Passes(p ...Pass) Option {
	return func(d *desc) {
		d.passes = p
	}
}

func Tags(t ...Tag) Option {
	return func(d *desc) {
		d.tags = t
	}
}

func OnOp(op ir2.Op) Option {
	return func(d *desc) {
		d.op = op
	}
}

func Once() Option {
	return func(d *desc) {
		d.once = true
	}
}

var xformers []desc

// Register an xform function
func Register(fn func(ir2.Iter), options ...Option) int {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	xformers = append(xformers, desc{
		name: name,
		fn:   fn,
	})
	d := &xformers[len(xformers)-1]

	for _, option := range options {
		option(d)
	}

	return 0
}

func Transform(pass Pass, fn *ir2.Func) {
	active, opXforms, otherXforms := activeXforms(pass, fn)
	tries := 0

	for {
		it := fn.InstrIter()
		var iter ir2.Iter = it

		for ; it.HasNext(); it.Next() {
			// run the xforms specific to the current op
			op := it.Instr().Op
			for _, xform := range opXforms[op] {
				perform(xform, iter)
			}

			// run the xforms that always run
			for _, xform := range otherXforms {
				perform(xform, iter)
			}
		}

		if !it.HasChanged() {
			break
		}

		tries++
		if tries > 1000 {
			panic(fmt.Sprintf("transforms do not terminate: pass: %d active: %v", pass, active))
		}
	}
}

func perform(xform *desc, it ir2.Iter) {
	if xform.disabled {
		return
	}
	xform.fn(it)
	if xform.once {
		xform.disabled = true
	}
}

// activeXforms determines the active xform functions for the current pass and tags
func activeXforms(pass Pass, fn *ir2.Func) ([]string, map[ir2.Op][]*desc, []*desc) {
	var active []string
	opXforms := make(map[ir2.Op][]*desc)
	var otherXforms []*desc

next:
	for i, xf := range xformers {
		inPass := false
		for _, p := range xf.passes {
			if p == pass {
				inPass = true
				break
			}
		}
		if !inPass {
			continue
		}

		for _, tag := range xf.tags {
			if !activeTags[tag] {
				continue next
			}
		}

		if xf.op != nil {
			opXforms[xf.op] = append(opXforms[xf.op], &xformers[i])
		} else {
			otherXforms = append(otherXforms, &xformers[i])
		}
		active = append(active, xf.name)
	}

	return active, opXforms, otherXforms
}
