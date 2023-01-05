package ir2

import "fmt"

// ID is an identifier that's unique within a Func
type ID uint32

type IDKind uint8

const (
	UnknownID IDKind = iota
	BlockID
	InstrID
	ValueID
	PlaceholderID
)

// Placeholder is an invalid ID meant to signal a place that needs to be filled
var Placeholder ID = idFor(PlaceholderID, -1)

const kindShift = 24
const idMask = (1 << kindShift) - 1

// Kind returns the IDKind
func (id ID) Kind() IDKind {
	return IDKind(id >> kindShift)
}

// Index returns the ID number for the ID
func (id ID) Index() int {
	return int(id & idMask)
}

// IDString returns the ID string
func (id ID) IDString() string {
	prefix := ""
	switch id.Kind() {
	case BlockID:
		prefix = "b"
	case InstrID:
		prefix = "i"
	case ValueID:
		prefix = "v"
	}
	return fmt.Sprintf("%s%d", prefix, id.Index())
}

// idFor returns a new ID for a given object kind
func idFor(kind IDKind, idNum int) ID {
	return ID(kind)<<kindShift | (ID(idNum) & idMask)
}

// IsBlock returns if ID points to a Block
func (id ID) IsBlock() bool {
	return id.Kind() == BlockID
}

// BlockIn returns the Block in the Func or nil
// if this ID is not for a Block
func (id ID) BlockIn(fn *Func) *Block {
	if id.Kind() != BlockID {
		return nil
	}

	return fn.idBlocks[id&idMask]
}

// IsInstr returns if ID points to a Instr
func (id ID) IsInstr() bool {
	return id.Kind() == InstrID
}

// InstrIn returns the Instr in the Func or nil
// if this ID is not for a Instr
func (id ID) InstrIn(fn *Func) *Instr {
	if id.Kind() != InstrID {
		return nil
	}

	return fn.idInstrs[id&idMask]
}

// IsValue returns true if ID points to a Value
func (id ID) IsValue() bool {
	return id.Kind() == ValueID
}

// ValueIn returns the Value in the Func or nil
// if this ID is not a value or not in the Func
func (id ID) ValueIn(fn *Func) *Value {
	if id.Kind() != ValueID {
		return nil
	}

	return fn.idValues[id&idMask]
}
