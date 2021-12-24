package ir2

import "fmt"

const kindShift = 24
const idMask = (1 << kindShift) - 1

// ObjectKind returns the kind of object the ID is for
func (id ident) ObjectKind() ObjectKind {
	return ObjectKind(id >> kindShift)
}

// IDNum returns the ID number for the ID
func (id ident) IDNum() int {
	return int(id & idMask)
}

// IDString returns the ID string
func (id ident) IDString() string {
	prefix := ""
	switch id.ObjectKind() {
	case BlockObject:
		prefix = "b"
	case InstrObject:
		prefix = "i"
	case ValueObject:
		prefix = "v"
	}
	return fmt.Sprintf("%s%d", prefix, id.IDNum())
}

// idFor returns a new ID for a given object kind
func idFor(kind ObjectKind, idNum int) ident {
	return ident(kind)<<kindShift | (ident(idNum) & idMask)
}

// IsBlock returns if ID points to a Block
func (id ident) IsBlock() bool {
	return id.ObjectKind() == BlockObject
}

// BlockIn returns the Block in the Func or nil
// if this ID is not for a Block
func (id ident) BlockIn(fn *Func) *Block {
	if id.ObjectKind() != BlockObject {
		return nil
	}

	return fn.idBlocks[id&idMask]
}

// IsInstr returns if ID points to a Instr
func (id ident) IsInstr() bool {
	return id.ObjectKind() == InstrObject
}

// InstrIn returns the Instr in the Func or nil
// if this ID is not for a Instr
func (id ident) InstrIn(fn *Func) *Instr {
	if id.ObjectKind() != InstrObject {
		return nil
	}

	return fn.idInstrs[id&idMask]
}

// IsValue returns if ID points to a Value
func (id ident) IsValue() bool {
	return id.ObjectKind() == ValueObject
}

// ValueIn returns the Value in the Func or nil
// if this ID is not for a Value
func (id ident) ValueIn(fn *Func) *Value {
	if id.ObjectKind() != ValueObject {
		return nil
	}

	return fn.idValues[id&idMask]
}
