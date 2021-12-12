package parseir

import (
	"go/token"
	"strings"

	"github.com/rj45/nanogo/ir2"
)

func (p *Parser) parseFunc() {
	// Read a func label
	name := p.parseLabel()

	parts := strings.Split(name, "__")
	name = parts[len(parts)-1]

	p.fn = p.pkg.NewFunc(name, nil)
	p.fn.Referenced = true // todo: fixme with correct value
	p.blk = nil

	p.blkLabels = make(map[string]*ir2.Block)
	p.values = make(map[string]*ir2.Value)
	p.blkLinks = make(map[*ir2.Block]string)
	p.valLinks = make(map[*ir2.Instr]struct {
		label string
		pos   int
	})

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

		default:
			p.errorf("found %q, expected func or block label", lit)
		}
	}
}

func (p *Parser) resolveLinks() {
	for a, label := range p.blkLinks {
		b, ok := p.blkLabels[label]
		if !ok {
			p.errorf("unable to resolve block link %s from %s", b, a)
		}
		a.AddSucc(b)
		b.AddPred(a)
	}

	for ins, arg := range p.valLinks {
		v, ok := p.values[arg.label]
		if !ok {
			p.errorf("unable to resolve value link %s from %s in %s", arg.label, ins, p.fn.FullName)
		}
		ins.ReplaceArg(arg.pos, v)
	}
}

func (p *Parser) parseBlock() {
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

		case token.EOF:
			p.unscan()
			return

		default:
			p.errorf("found %q, expected block label or instr", lit)
		}
	}
}

func (p *Parser) parseLabel() string {
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
