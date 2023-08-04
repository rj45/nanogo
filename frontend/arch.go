package frontend

type Arch interface {
	Name() string
}

var archName string

func SetArch(a Arch) {
	archName = a.Name()
}
