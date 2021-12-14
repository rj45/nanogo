package typ

type Context struct {
	sizes     [17]uint8
	wordBytes uint8
}

func (c *Context) Bytes(typ Type) int {
	// todo: calc sizes of composite types
	return int(c.sizes[typ])
}

func (c *Context) Words(typ Type) int {
	return int(c.sizes[typ] / c.wordBytes)
}

func (c *Context) Elem(typ Type) Type {
	return Unknown
}

func (c *Context) Dir(typ Type) ChanDir {
	return ChanDir(99)
}

func (c *Context) Len(typ Type) int {
	return -1
}

func (c *Context) Key(typ Type) Type {
	return Unknown
}
