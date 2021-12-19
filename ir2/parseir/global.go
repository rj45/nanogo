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
	for _, pkg := range p.prog.Packages() {
		for _, fn := range pkg.Funcs() {
			for _, label := range fn.PlaceholderLabels() {
				glob := p.prog.Global(label)
				if glob != nil {
					v := fn.ValueFor(glob.Type, glob)
					fn.ResolvePlaceholder(label, v)
					if p.trace {
						p.printTrace("resolved", label, "with global", v)
					}
					glob.Referenced = true
					continue
				}

				reffn := p.prog.Func(label)
				if reffn == nil {
					p.errorf("unable to resolve global link %s in %s", label, fn.FullName)
					continue
				}

				reffn.Referenced = true

				v := fn.ValueFor(reffn.Sig, reffn)
				fn.ResolvePlaceholder(label, v)

				if p.trace {
					p.printTrace("resolved", label, "with func", v)
				}
			}
		}
	}
}
