package parseir

import (
	"go/token"
	"go/types"
	"strconv"
)

type opaqueType struct {
	types.Type
	name string
}

func (t *opaqueType) String() string { return t.name }

var tRangeIter = &opaqueType{nil, "iter"}

func (p *Parser) parseColonType() types.Type {
	if p.trace {
		defer un(trace(p, "colonType"))
	}

	tok, _ := p.scan()
	if tok != token.COLON {
		p.unscan()
		return nil
	}

	p.scan()
	p.unscan()

	return p.parseType()
}

func (p *Parser) parseType() types.Type {
	if p.trace {
		defer un(trace(p, "type"))
	}

	typ := p.tryParseType()

	if typ == nil {
		// no type found
		p.errorf("expected type; wasn't found")
	}

	return typ
}

func (p *Parser) tryParseType() types.Type {
	if p.trace {
		defer un(trace(p, "tryType"))
	}

	tok, _ := p.scan()
	p.unscan()

	switch tok {
	case token.IDENT:
		typ := p.parseTypeName()
		return typ
	case token.LBRACK:
		return p.parseArrayType()
	case token.MUL:
		return p.parsePointerType()
	case token.FUNC:
		return p.parseFuncType()
	case token.INTERFACE:
		return p.parseInterfaceType()
	case token.STRUCT:
		return p.parseStructType()

	// TODO: parse these types
	// case token.MAP:
	// 	return p.parseMapType()
	// case token.CHAN, token.ARROW:
	// 	return p.parseChanType()
	// case token.LPAREN:
	// 	lparen := p.pos
	// 	p.next()
	// 	typ := p.parseType()
	// 	rparen := p.expect(token.RPAREN)
	// 	return &ast.ParenExpr{Lparen: lparen, X: typ, Rparen: rparen}
	default:
	}

	// no type found
	return nil
}

func (p *Parser) parseTypeName() types.Type {
	if p.trace {
		defer un(trace(p, "typeName"))
	}

	tok, lit := p.scan()
	if tok != token.IDENT {
		p.errorf("expected type or package name; found %q", lit)
	}

	name := lit

	pkg := p.pkg

	tok, _ = p.scan()

	if tok == token.PERIOD {
		// todo: handle aliases
		pkg = p.prog.Package(name)

		tok, lit := p.scan()
		if tok != token.IDENT {
			p.errorf("expected type name; found %q", lit)
		}
		name = lit
	} else {
		p.unscan()
	}

	typ := types.Universe.Lookup(name)
	if typ != nil {
		return typ.Type()
	}

	if name == "iter" {
		return tRangeIter
	}

	if pkg.Type == nil {
		panic("missing type on pkg " + pkg.Name)
	}

	td := pkg.TypeDef(name)
	if td != nil {
		return td.Type
	}

	p.errorf("unable to resolve type %s.%s", pkg.Name, name)

	return nil
}

func (p *Parser) lookupTypeFor(pkgname string, name string) types.Type {
	pkg := p.pkg

	if pkgname == "" {
		typ := types.Universe.Lookup(name)
		if typ != nil {
			return typ.Type()
		}
	}

	if pkgname != "" {
		// todo: handle aliases
		pkg = p.prog.Package(pkgname)
		if pkg == nil {
			panic("implement forward refs for packages")
		}
	}

	// todo: not sure this is correct.... may not add types to scopes
	scope := pkg.Type.Scope().Lookup(name)
	if scope != nil {
		return scope.Type()
	} else {
		panic("implement forward refs for extern packages")
	}
}

func (p *Parser) parseInterfaceType() types.Type {
	if p.trace {
		defer un(trace(p, "interfaceType"))
	}

	p.expect(token.INTERFACE, "interface type")
	p.expect(token.LBRACE, "interface type")
	p.expect(token.RBRACE, "interface type")

	return types.NewInterfaceType(nil, nil)
}

func (p *Parser) parsePointerType() types.Type {
	if p.trace {
		defer un(trace(p, "pointerType"))
	}

	p.expect(token.MUL, "pointer type")

	p.scan()
	p.unscan()

	typ := p.parseType()

	return types.NewPointer(typ)
}

func (p *Parser) parseArrayType() types.Type {
	if p.trace {
		defer un(trace(p, "arrayType"))
	}

	p.expect(token.LBRACK, "array or slice type")

	tok, _ := p.scan()

	if tok == token.RBRACK {
		return types.NewSlice(p.parseType())
	}

	p.unscan()
	_, lit := p.expect(token.INT, "array length")
	len, err := strconv.ParseUint(lit, 10, 64)
	if err != nil {
		p.errorf("bad array length: %s", err.Error())
	}

	p.expect(token.RBRACK, "array type")

	return types.NewArray(p.parseType(), int64(len))
}

func (p *Parser) parseFuncType() types.Type {
	if p.trace {
		defer un(trace(p, "funcType"))
	}

	p.expect(token.FUNC, "func type")

	params := p.parseParams()
	results := p.parseResults()
	var tparams *types.Tuple
	var tresults *types.Tuple
	if len(params) > 0 {
		tparams = types.NewTuple(params...)
	}
	if len(results) > 0 {
		tresults = types.NewTuple(results...)
	}

	return types.NewSignature(nil, tparams, tresults, false)
}

func (p *Parser) parseResults() []*types.Var {
	if p.trace {
		defer un(trace(p, "results"))
	}

	tok, _ := p.scan()
	pos := p.buf.pos
	p.unscan()
	if tok == token.LPAREN {
		return p.parseParams()
	}

	if tok == token.EOF || tok == token.SEMICOLON || tok == token.COMMA || tok == token.COLON {
		return nil
	}

	typ := p.tryParseType()
	if typ != nil {
		return []*types.Var{types.NewVar(pos, nil, "", typ)}
	}

	return nil
}

func (p *Parser) parseParams() []*types.Var {
	if p.trace {
		defer un(trace(p, "params"))
	}

	p.expect(token.LPAREN, "start func parameters")
	tok, _ := p.scan()
	p.unscan()
	var params []*types.Var
	if tok != token.RPAREN {
		params = p.parseParamList()
	}
	p.expect(token.RPAREN, "end func parameters")
	p.scan()
	p.unscan()
	return params
}

func (p *Parser) parseParamList() []*types.Var {
	if p.trace {
		defer un(trace(p, "paramList"))
	}

	var params []*types.Var

	tok, _ := p.scan()
	if tok == token.EOF || tok == token.RPAREN {
		return nil
	}
	p.unscan()

	for {
		params = append(params, p.parseParamDecl())

		tok, _ := p.scan()
		p.unscan()
		if tok == token.EOF || tok == token.RPAREN {
			return params
		}

		p.expect(token.COMMA, "param list comma")
	}
}

func (p *Parser) parseParamDecl() *types.Var {
	if p.trace {
		defer un(trace(p, "paramDecl"))
	}

	var name string
	var typ types.Type
	pos := p.buf.pos

	tok, lit := p.scan()
	switch tok {
	case token.IDENT:
		// name
		name = lit
		tok, lit = p.scan()
		p.unscan()

		switch tok {
		case token.IDENT, token.MUL, token.ARROW, token.FUNC, token.CHAN, token.MAP, token.STRUCT, token.INTERFACE, token.LPAREN, token.LBRACK:
			typ = p.parseType()
		case token.ELLIPSIS:
			element := p.parseType()
			typ = types.NewSlice(element)
		case token.PERIOD:
			p.scan()
			typ = p.lookupTypeFor(name, lit)
			name = ""
		case token.RPAREN, token.COMMA:
			typ = p.lookupTypeFor("", name)
			name = ""
		default:
			p.errorf("expected type, got %s %q", tok, lit)
		}

	case token.MUL, token.ARROW, token.FUNC, token.LBRACK, token.CHAN, token.MAP, token.STRUCT, token.INTERFACE, token.LPAREN:
		p.unscan()
		typ = p.parseType()
	case token.ELLIPSIS:
		element := p.parseType()
		typ = types.NewSlice(element)
	default:
		p.errorf("expected param/result, got %s %q", tok, lit)
	}

	return types.NewVar(pos, p.pkg.Type, name, typ)
}

func (p *Parser) parseStructType() types.Type {
	if p.trace {
		defer un(trace(p, "structType"))
	}

	p.expect(token.STRUCT, "struct type")
	p.expect(token.LBRACE, "struct type")

	var list []*types.Var

	for {
		tok, _ := p.scan()
		p.unscan()
		if tok != token.IDENT && tok != token.MUL {
			break
		}

		list = append(list, p.parseFieldDecl())
	}

	p.expect(token.RBRACE, "struct type")

	return types.NewStruct(list, nil)
}

func (p *Parser) parseFieldDecl() *types.Var {
	if p.trace {
		defer un(trace(p, "fieldDecl"))
	}

	tok, lit := p.scan()
	var typ types.Type
	pos := p.buf.pos
	var name string

	if tok == token.IDENT {
		name = lit
		tok, lit = p.scan()
		p.unscan()
		if tok == token.PERIOD {
			typ = p.lookupTypeFor(name, lit)
		} else if tok == token.STRING || tok == token.SEMICOLON || tok == token.RBRACE {
			typ = p.lookupTypeFor(p.pkg.Name, name)
		} else if tok == token.LBRACK {
			typ = p.parseArrayType()
		} else {
			// T P
			typ = p.parseType()
		}
	} else {
		// embedded type
		typ = p.parseType()
	}

	tok, _ = p.scan()
	if tok != token.SEMICOLON {
		p.unscan()
	}

	return types.NewVar(pos, p.pkg.Type, name, typ)
}
