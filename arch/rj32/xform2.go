package rj32

import "github.com/rj45/nanogo/xform2"

func (cpuArch) XformTags2() []xform2.Tag {
	return nil
}

func (cpuArch) RegisterXforms() {
	xform2.Register(translate, xform2.OnlyPass(xform2.Lowering))
}
