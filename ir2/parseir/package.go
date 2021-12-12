package parseir

import (
	"go/token"
	"strconv"

	"github.com/rj45/nanogo/ir2"
)

func (p *Parser) parse() {
	for {
		if tok, lit := p.scan(); tok != token.PACKAGE {
			if tok == token.EOF {
				return
			}
			p.errorf("found %q, expected package", lit)
		}

		p.parsePackage()
	}
}

func (p *Parser) parsePackage() {
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
		p.pkg = &ir2.Package{Name: name, Path: path}
		p.prog.AddPackage(p.pkg)
	}

	for {
		if tok, lit := p.scan(); tok != token.FUNC {
			if tok == token.EOF {
				return
			}
			if tok == token.PACKAGE {
				p.unscan()
				return
			}
			p.errorf("found %q, expected func", lit)
		}

		p.parseFunc()
	}
}
