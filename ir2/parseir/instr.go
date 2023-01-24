// Copyright (c) 2014 Ben Johnson; MIT Licensed; Similar to LICENSE
// Copyright (c) 2021 rj45 (github.com/rj45); MIT Licensed

package parseir

import (
	"go/token"
	"go/types"
	"regexp"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
)

type typedToken struct {
	tok   token.Token
	lit   string
	typ   types.Type
	glob  bool
	block bool
	args  []typedToken
}

func (p *Parser) parseInstr() {
	if p.trace {
		defer un(trace(p, "instr"))
	}

	var list []typedToken
	var opcode string
	var last typedToken
	var defs []typedToken

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
			if last.tok != token.ILLEGAL {
				list = append(list, last)
				last.tok = token.ILLEGAL
			}

			if len(list) == 0 {
				p.errorf("found %q, expected list", lit)
			}

			// list is def list
			defs = list
			list = nil

		case token.SUB, token.INT, token.IDENT, token.XOR, token.PERIOD, token.RANGE, token.IF, token.RETURN:
			p.unscan()

			if last.tok == token.ILLEGAL {
				last = p.parseVar()
			} else if opcode == "" {
				opcode = last.lit
				last = p.parseVar()
			} else {
				return
			}

		case token.COMMA:
			p.unscan()
			list = p.parseList(last)
			last.tok = token.ILLEGAL

		case token.SEMICOLON:
			if opcode == "" {
				opcode = last.lit
			} else if last.tok != token.ILLEGAL {
				list = append(list, last)
			}

			p.addInstr(defs, opcode, list)
			return

		default:
			p.errorf("found %s %q, expected instr", tok, lit)
		}
	}
}

func (p *Parser) parseVar() typedToken {
	if p.trace {
		defer un(trace(p, "var"))
	}

	tok, lit := p.scan()

	switch tok {
	case token.SUB, token.INT:
		p.unscan()
		return p.parseIntVar()

	case token.XOR:
		tok, lit = p.expect(token.IDENT, "global var ref")
		p.scan()
		p.unscan()

		return typedToken{
			tok: tok, lit: lit, typ: nil, glob: true}

	case token.PERIOD:
		_, lit = p.expect(token.IDENT, "block ref")

		blkname := lit

		tok, _ = p.scan()
		p.unscan()

		var args []typedToken
		if tok == token.LPAREN {
			args = p.parseParenList()
		}

		return typedToken{
			tok: token.PERIOD, lit: blkname, typ: nil, glob: true, block: true, args: args}

	case token.IDENT, token.RANGE, token.IF, token.RETURN:
		next, _ := p.scan()
		p.unscan()
		var typ types.Type
		if next == token.COLON {
			typ = p.parseColonType()
		}

		return typedToken{
			tok: tok, lit: lit, typ: typ}
	default:
	}

	p.errorf("expected identifier or literal; got %q %s", lit, tok)
	return typedToken{}
}

func (p *Parser) parseIntVar() typedToken {
	if p.trace {
		defer un(trace(p, "intVar"))
	}

	tok, _ := p.scan()

	negate := false
	if tok == token.SUB {
		negate = true
		p.scan()
	}

	p.unscan()
	tok, lit := p.expect(token.INT, "int variable")
	if negate {
		lit = "-" + lit
	}

	return typedToken{tok: tok, lit: lit, typ: types.Typ[types.UntypedInt]}
}

var valueRefRe = regexp.MustCompile(`^v(\d+)(_\w+)?$`)

func (p *Parser) addInstr(defs []typedToken, opcode string, args []typedToken) {
	if p.trace {
		defer un(trace(p, "addInstr"))
	}

	opv, err := op.OpString(opcode)
	if err != nil {
		p.errorf("unknown instruction %q found: %s", opcode, err)
	}

	// todo: fix type here
	ins := p.fn.NewInstr(opv, nil)

	for _, def := range defs {
		if def.typ == nil && opcode != "next" {
			p.errorf("def %s is missing a type for instruction %s", def.lit, opcode)
		}
		v := ins.AddDef(p.fn.NewValue(def.typ))
		p.values[def.lit] = v
	}

	for an, arg := range args {
		if arg.tok == token.PERIOD {
			p.blkLinks[p.blk] = append(p.blkLinks[p.blk], arg.lit)

			for _, barg := range arg.args {
				if val, ok := p.values[barg.lit]; ok {
					p.blk.InsertArg(-1, val)

					if p.trace {
						p.printTrace("block value arg:", val)
					}
					continue
				}

				if valueRefRe.MatchString(barg.lit) {
					p.blk.InsertArg(-1, p.fn.PlaceholderFor(barg.lit))
					if p.trace {
						p.printTrace("block arg placeholder:", barg.lit)
					}
					continue
				}

				if barg.typ != nil {
					if p.trace {
						p.printTrace("block typed arg:", barg.typ, barg.lit)
					}
					val := p.fn.ValueFor(barg.typ, barg.lit)
					p.blk.InsertArg(-1, val)
					continue
				}
			}

			continue
		}

		if val, ok := p.values[arg.lit]; ok {
			if p.trace {
				p.printTrace("arg type: value", val)
			}
			ins.InsertArg(an, val)
			continue
		}

		glob := p.prog.Global(arg.lit)
		if glob != nil {
			if p.trace {
				p.printTrace("arg type: global", glob)
			}

			val := p.fn.ValueFor(glob.Type, glob)
			glob.Referenced = true
			ins.InsertArg(an, val)
			continue
		}

		if fn := p.prog.Func(arg.lit); fn != nil {
			if p.trace {
				p.printTrace("arg type: func", fn.FullName)
			}

			fn.Referenced = true
			val := p.fn.ValueFor(fn.Sig, fn)
			ins.InsertArg(an, val)
			continue
		}

		if arg.glob {
			if p.globals == nil {
				p.globals = make(map[string]map[*ir2.Func]*ir2.Value)
			}

			val, found := p.globals[arg.lit][p.fn]
			if !found {
				val = p.fn.NewValue(nil)

				if p.globals[arg.lit] == nil {
					p.globals[arg.lit] = make(map[*ir2.Func]*ir2.Value)
				}

				p.globals[arg.lit][p.fn] = val
			}

			ins.InsertArg(an, val)
			if p.trace {
				p.printTrace("arg type: global placeholder", an)
			}
			continue
		}

		if arg.typ != nil {
			if p.trace {
				p.printTrace("arg type: given type", arg.typ)
			}
			val := p.fn.ValueFor(arg.typ, arg.lit)
			ins.InsertArg(an, val)
			continue
		}

		if valueRefRe.MatchString(arg.lit) {
			ins.InsertArg(an, p.fn.PlaceholderFor(arg.lit))
			if p.trace {
				p.printTrace("arg type: placeholder", an)
			}
			continue
		}

		builtin := types.Universe.Lookup(arg.lit)
		if builtin != nil && builtin.Type() != nil && builtin.Type() != types.Typ[types.Invalid] {
			val := p.fn.ValueFor(builtin.Type(), arg.lit)
			ins.InsertArg(an, val)
			if p.trace {
				p.printTrace("arg type: builtin", builtin.Type().String())
			}
			continue
		}

		if len(defs) == 1 && defs[0].typ != nil {
			typ := defs[0].typ
			if p.trace {
				p.printTrace("arg type: def", defs[0])
			}

			val := p.fn.ValueFor(typ, arg.lit)
			ins.InsertArg(an, val)
			continue
		}

		p.errorf("type is required on arg %s for instruction %s", arg.lit, opcode)
	}

	p.blk.InsertInstr(-1, ins)
}

func (p *Parser) parseParenList() (list []typedToken) {
	if p.trace {
		defer un(trace(p, "parenList"))
	}

	p.expect(token.LPAREN, "param/arg list")

	var args []typedToken

	tok, _ := p.scan()
	if tok != token.RPAREN {
		p.unscan()
		first := p.parseVar()

		tok, _ := p.scan()
		if tok == token.COMMA {
			p.unscan()
			args = p.parseList(first)
		} else {
			args = append(args, first)
			p.unscan()
		}
	}

	p.expect(token.RPAREN, "param/arg list end")

	p.scan()
	p.unscan()

	return args
}

func (p *Parser) parseList(first typedToken) (list []typedToken) {
	if p.trace {
		defer un(trace(p, "list"))
	}

	if first.tok != token.ILLEGAL {
		list = append(list, first)
	}
	for {
		if tok, _ := p.scan(); tok != token.COMMA {
			p.unscan()
			return
		}

		list = append(list, p.parseVar())
	}
}
