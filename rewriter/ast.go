package rewriter

type Rule struct {
	From *Node
	To   *Node
}

type NodeKind uint8

const (
	Invalid NodeKind = iota
	Call
	Ident
	Nil
)

type Node struct {
	Kind NodeKind
	Name string
	Args []*Node
}
