// Copyright (c) 2014 Ben Johnson; MIT Licensed; Similar to LICENSE
// Copyright (c) 2021 rj45 (github.com/rj45); MIT Licensed

package parseir

import (
	"go/token"
	"regexp"

	"github.com/rj45/nanogo/ir/op"
)

func (p *Parser) parseInstr() {
	var list []string
	var opcode string
	var last string
	var defs []string

	nextNegate := false

	for {
		tok, lit := p.scan()

		// a, b = op e, f
		// a, b = op e
		// a, b = op
		// b = op e, f
		// b = op e
		// b = op
		// op
		// op e
		// op e, f

		switch tok {
		case token.ASSIGN:
			if last != "" {
				list = append(list, last)
				last = ""
			}

			if len(list) == 0 {
				p.errorf("found %q, expected list", lit)
			}

			// list is def list
			defs = list
			list = nil

		case token.SUB:
			nextNegate = true

		case token.IDENT, token.INT:
			if nextNegate {
				nextNegate = false
				lit = "-" + lit
			}

			if last == "" {
				last = lit
			} else if opcode == "" {
				opcode = last
				last = lit
			} else {
				p.errorf("found %q, expected newline %q %q", lit, last, opcode)
			}

		case token.COMMA:
			p.unscan()
			list = p.parseList(last)
			last = ""

		case token.SEMICOLON:
			if opcode == "" {
				opcode = last
			} else if last != "" {
				list = append(list, last)
			}

			p.addInstr(defs, opcode, list)
			return

		default:
			p.errorf("found %s %q, expected instr", tok, lit)
		}
	}
}

var blockRefRe = regexp.MustCompile(`^b\d+$`)
var valueRefRe = regexp.MustCompile(`^v\d+$`)

func (p *Parser) addInstr(defs []string, opcode string, args []string) {
	opv, err := op.OpString(opcode)
	if err != nil {
		p.errorf("unknown instruction %q found: %s", opcode, err)
	}

	// todo: fix type here
	ins := p.fn.NewInstr(opv, nil)

	for _, def := range defs {
		// todo: fix type here
		v := ins.AddDef(p.fn.NewValue(nil))
		p.values[def] = v
	}

	for _, arg := range args {
		if blockRefRe.MatchString(arg) {
			p.blkLinks[p.blk] = arg
			continue
		}

		val, ok := p.values[arg]
		if !ok {
			if valueRefRe.MatchString(arg) {
				p.valLinks[ins] = struct {
					label string
					pos   int
				}{
					label: arg,
					pos:   ins.NumArgs(),
				}
				val = nil
			} else {
				// todo: fix type here
				val = p.fn.ValueFor(nil, arg)
			}
		}
		ins.InsertArg(-1, val)
	}

	p.blk.InsertInstr(-1, ins)
}

func (p *Parser) parseList(first string) (list []string) {
	if first != "" {
		list = append(list, first)
	}
	for {
		if tok, _ := p.scan(); tok != token.COMMA {
			p.unscan()
			return
		}

		tok, lit := p.scan()

		if tok != token.IDENT {
			p.unscan()
			return
		}

		list = append(list, lit)
	}
}
