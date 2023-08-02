package rj32

import (
	"github.com/rj45/nanogo/ir2/op"
	"github.com/rj45/nanogo/xform2"
)

func (cpuArch) XformTags2() []xform2.Tag {
	return []xform2.Tag{xform2.LoadStoreOffset}
}

func (cpuArch) RegisterXforms() {
	xform2.Register(translate, xform2.OnlyPass(xform2.Lowering))
	xform2.Register(translateCopies, xform2.OnlyPass(xform2.Finishing), xform2.OnOp(op.Copy))
}
