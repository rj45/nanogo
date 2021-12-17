package parseir

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/rj45/nanogo/ir2"
)

func (p *Parser) parseGlobal() {
	if p.trace {
		defer un(trace(p, "global"))
	}

	tok, lit := p.scan()
	if tok != token.IDENT {
		p.errorf("found %q, expected global name", lit)
	}

	fullname := lit

	name := strings.Split(fullname, "__")[1]

	p.expect(token.COLON, "global")

	typ := p.parseType()
	glob := p.pkg.NewGlobal(name, typ)

	tok, _ = p.scan()
	if tok == token.SEMICOLON {
		return
	}
	if tok == token.ASSIGN {
		glob.Value = p.parseLiteral()
		return
	}
	p.unscan()
}

func (p *Parser) parseLiteral() ir2.Const {
	tok, lit := p.scan()
	var con ir2.Const
	switch tok {
	case token.STRING:
		str, err := strconv.Unquote(lit)
		if err != nil {
			p.errorf("failed unquoting string: %s", err.Error())
		}
		con = ir2.ConstFor(str)
	case token.INT:
		i, err := strconv.Atoi(lit)
		if err != nil {
			p.errorf("failed converting to int: %s", err.Error())
		}
		con = ir2.ConstFor(i)
	default:
		p.errorf("bad literal %q, expecting string or int", lit)
		return nil
	}

	tok, _ = p.scan()
	if tok != token.SEMICOLON {
		p.unscan()
	}

	return con
}

func (p *Parser) resolveGlobalLinks() {
	for ins, arg := range p.globLinks {
		glob := p.prog.Global(arg.fullname)
		if glob != nil {
			v := ins.Func().ValueFor(glob.Type, glob)
			ins.ReplaceArg(arg.pos, v)
			if p.trace {
				p.printTrace("resolved", arg.fullname, "with global", v)
			}
			glob.Referenced = true
			continue
		}

		fn := p.prog.Func(arg.fullname)
		if fn == nil {
			p.errorf("unable to resolve global link %s from %s in %s", arg.fullname, ins, ins.Func().FullName)
			continue
		}

		fn.Referenced = true

		v := ins.Func().ValueFor(fn.Sig, fn)
		ins.ReplaceArg(arg.pos, v)

		if p.trace {
			p.printTrace("resolved", arg.fullname, "with func", v, "at", arg.pos)
		}
	}
}
