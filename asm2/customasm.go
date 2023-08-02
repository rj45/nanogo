package asm2

import (
	"fmt"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/sizes"
)

type CustomASM struct{}

func (CustomASM) Section(s Section) string {
	switch s {
	case Code:
		return "#bank code"
	case Data:
		return "#bank data"
	case Bss:
		return "#bank bss"
	}
	panic("unknown section")
}

func (CustomASM) GlobalLabel(global *ir2.Global) string {
	return global.FullName
}

func (CustomASM) PCRelAddress(offsetWords int) string {
	return fmt.Sprintf("$ + %d", offsetWords)
}

func (CustomASM) Word(value string) string {
	wordsize := int(sizes.WordSize()) * sizes.MinAddressableBits()
	return fmt.Sprintf("#d%d le(%s)", wordsize, value)
}

func (CustomASM) String(val string) string {
	bytesize := sizes.MinAddressableBits()
	switch bytesize {
	case 8:
		return fmt.Sprintf("#d8 %q", val)
	case 16:
		return fmt.Sprintf("#d16 utf16le(%q)", val)
	case 32:
		return fmt.Sprintf("#d32 utf32le(%q)", val)
	}
	panic("unsupported byte size")
}

func (CustomASM) Reserve(bytes int) string {
	return fmt.Sprintf("#res %d", bytes)
}

func (CustomASM) Comment(comment string) string {
	return fmt.Sprintf("; %s", comment)
}

func (CustomASM) BlockLabel(id string) string {
	return fmt.Sprintf(".%s", id)
}
