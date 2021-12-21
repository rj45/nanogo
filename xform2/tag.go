package xform2

type Tag uint8

const (
	Invalid Tag = iota
	HasFramePointer

	// ...

	NumTags
)

var activeTags []bool

type Arch interface {
	XformTags() []Tag
}

func SetArch(a Arch) {
	activeTags = make([]bool, NumTags)
	for _, tag := range a.XformTags() {
		activeTags[tag] = true
	}
}
