package parseir

import (
	"go/token"
	"go/types"
)

func (p *Parser) parseTypeDef() {
	if p.trace {
		defer un(trace(p, "typeDef"))
	}

	tok, lit := p.scan()
	if tok != token.IDENT {
		p.errorf("found %q, expected type name", lit)
	}
	pos := p.buf.pos

	name := lit

	p.expect(token.COLON, "typedef")

	typ := p.parseType()

	typ = types.NewNamed(types.NewTypeName(pos, p.pkg.Type, name, nil), typ, nil)

	p.pkg.NewTypeDef(name, typ)

	tok, _ = p.scan()
	if tok == token.SEMICOLON {
		return
	}
	p.unscan()
}
