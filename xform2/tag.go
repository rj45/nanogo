package xform2

type Tag uint8

const (
	Invalid Tag = iota
	HasFramePointer
	LoadStoreOffset

	// ...

	NumTags
)

var activeTags []bool
