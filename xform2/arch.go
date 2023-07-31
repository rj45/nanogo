package xform2

type Arch interface {
	XformTags2() []Tag
	RegisterXforms()
}

func SetArch(a Arch) {
	activeTags = make([]bool, NumTags)
	for _, tag := range a.XformTags2() {
		activeTags[tag] = true
	}
	a.RegisterXforms()
}
