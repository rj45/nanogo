package ir2

func (glob *Global) String() string {
	return glob.FullName
}

func (glob *Global) Package() *Package {
	return glob.pkg
}
