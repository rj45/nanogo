package parseir

import (
	"go/token"
	"go/types"
	"strconv"

	"github.com/rj45/nanogo/ir2"
)

func (p *Parser) parse() {
	if p.trace {
		defer un(trace(p, "parse"))
	}

	for {
		if tok, lit := p.scan(); tok != token.PACKAGE {
			if tok == token.EOF {
				p.resolveGlobalLinks()
				return
			}
			p.errorf("found %q, expected package", lit)
		}

		p.parsePackage()
	}
}

func (p *Parser) parsePackage() {
	if p.trace {
		defer un(trace(p, "package"))
	}

	tok, lit := p.scan()
	if tok != token.IDENT {
		p.errorf("found %q, expected package name", lit)
	}

	name := lit

	tok, lit = p.scan()
	if tok != token.STRING {
		p.errorf("found %q, expected package path string", lit)
	}

	path, err := strconv.Unquote(lit)
	if err != nil {
		p.errorf("failed unquoting string %s", lit)
	}

	tok, lit = p.scan()
	if tok != token.SEMICOLON {
		p.errorf("found %q, expected newline after package", lit)
	}

	p.pkg = p.prog.Package(path)
	if p.pkg == nil {
		p.pkg = &ir2.Package{
			Name: name,
			Path: path,
			Type: types.NewPackage(path, name),
		}

		p.prog.AddPackage(p.pkg)
	}

	for {
		tok, lit := p.scan()

		switch tok {
		case token.EOF:
			return
		case token.PACKAGE:
			p.unscan()
			return
		case token.FUNC:
			p.parseFunc()
		case token.VAR:
			p.parseGlobal()
		default:
			p.errorf("found %q %q, expected func", lit, tok)
		}
	}
}
