package parseir

import (
	"go/token"
	"strings"

	"github.com/rj45/nanogo/ir2"
)

func (p *Parser) parseFunc() {
	if p.trace {
		defer un(trace(p, "func"))
	}

	// Read a func label
	name := p.parseLabel()

	parts := strings.Split(name, "__")
	name = parts[len(parts)-1]

	p.fn = p.pkg.NewFunc(name, nil)
	p.fn.Referenced = true // todo: fixme with correct value
	p.blk = nil

	p.blkLabels = make(map[string]*ir2.Block)
	p.values = make(map[string]*ir2.Value)
	p.blkLinks = make(map[*ir2.Block][]string)

	for {
		tok, lit := p.scan()

		switch tok {
		case token.PERIOD:
			p.parseBlock()

		case token.FUNC:
			p.resolveLinks()
			p.unscan()
			return

		case token.EOF:
			p.resolveLinks()
			p.unscan()
			return

		case token.PACKAGE:
			p.resolveLinks()
			p.unscan()
			return

		default:
			p.errorf("found %q, expected func or block label", lit)
		}
	}
}

func (p *Parser) resolveLinks() {
	for a, labels := range p.blkLinks {
		for _, label := range labels {
			b, ok := p.blkLabels[label]
			if !ok {
				p.errorf("unable to resolve block link %s from %s", b, a)
			}
			a.AddSucc(b)
			b.AddPred(a)
		}
	}

	for _, label := range p.fn.PlaceholderLabels() {
		v, ok := p.values[label]
		if !ok {
			p.errorf("unable to resolve placeholder value %s in %s", label, p.fn.FullName)
			continue
		}
		p.fn.ResolvePlaceholder(label, v)
	}
}

func (p *Parser) parseBlock() {
	if p.trace {
		defer un(trace(p, "block"))
	}

	// Read a block label
	name := p.parseLabel()

	p.blk = p.fn.NewBlock()
	p.blkLabels[name] = p.blk
	p.fn.InsertBlock(-1, p.blk)

	for {
		tok, lit := p.scan()

		switch tok {
		case token.PERIOD:
			p.unscan()
			return

		case token.FUNC:
			p.unscan()
			return

		case token.IDENT:
			p.unscan()
			p.parseInstr()

		case token.PACKAGE:
			p.unscan()
			return

		case token.EOF:
			p.unscan()
			return

		default:
			p.errorf("found %q, expected block label or instr", lit)
		}
	}
}

func (p *Parser) parseLabel() string {
	if p.trace {
		defer un(trace(p, "label"))
	}

	tok, lit := p.scan()
	if tok != token.IDENT {
		p.errorf("found %q, expected label", lit)
	}

	label := lit

	tok, lit = p.scan()
	if tok != token.COLON {
		p.errorf("found %q, expected label", lit)
	}

	return label
}
