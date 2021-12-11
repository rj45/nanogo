// Copyright (c) 2014 Ben Johnson; MIT Licensed; Similar to LICENSE
// Copyright (c) 2021 rj45 (github.com/rj45); MIT Licensed

package parseir

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/rj45/nanogo/ir/op"
	"github.com/rj45/nanogo/ir2"
)

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}

	// current context
	prog *ir2.Program
	pkg  *ir2.Package
	fn   *ir2.Func
	blk  *ir2.Block

	// context maps
	blkLabels map[string]*ir2.Block
	values    map[string]*ir2.Value

	blkLinks map[*ir2.Block]string
	valLinks map[*ir2.Instr]struct {
		label string
		pos   int
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader, prog *ir2.Program) *Parser {
	return &Parser{s: NewScanner(r), prog: prog}
}

func (p *Parser) Parse() error {
	for {
		if tok, lit := p.scanIgnoreWhitespace(); tok != PACKAGE {
			if tok == EOF {
				return nil
			}
			return fmt.Errorf("found %q, expected package", lit)
		}

		if err := p.parsePackage(); err != nil {
			return err
		}
	}
}

func (p *Parser) parsePackage() error {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != IDENT {
		return fmt.Errorf("found %q, expected package name", lit)
	}

	name := lit

	tok, lit = p.scanIgnoreWhitespace()
	if tok != STR {
		return fmt.Errorf("found %q, expected package path string", lit)
	}

	path := lit

	p.pkg = p.prog.Package(path)
	if p.pkg == nil {
		p.pkg = &ir2.Package{Name: name, Path: path}
		p.prog.AddPackage(p.pkg)
	}

	for {
		if tok, lit := p.scanIgnoreWhitespace(); tok != FUNC {
			if tok == EOF {
				return nil
			}
			if tok == PACKAGE {
				p.unscan()
				return nil
			}
			return fmt.Errorf("found %q, expected func", lit)
		}

		if err := p.parseFunc(); err != nil {
			return err
		}
	}
}

func (p *Parser) parseFunc() error {
	// Read a func label
	name, err := p.parseLabel()
	if err != nil {
		return err
	}

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
		tok, lit := p.scanIgnoreWhitespace()

		switch tok {
		case DOT:
			if err := p.parseBlock(); err != nil {
				return err
			}

		case FUNC:
			p.resolveLinks()
			p.unscan()
			return nil

		case EOF:
			p.resolveLinks()
			p.unscan()
			return nil

		default:
			return fmt.Errorf("found %q, expected func or block label", lit)
		}
	}
}

func (p *Parser) resolveLinks() error {
	for a, label := range p.blkLinks {
		b, ok := p.blkLabels[label]
		if !ok {
			return fmt.Errorf("unable to resolve block link %s from %s", b, a)
		}
		a.AddSucc(b)
		b.AddPred(a)
	}

	for ins, arg := range p.valLinks {
		v, ok := p.values[arg.label]
		if !ok {
			return fmt.Errorf("unable to resolve value link %s from %s in %s", arg.label, ins, p.fn.FullName)
		}
		ins.ReplaceArg(arg.pos, v)
	}
	return nil
}

func (p *Parser) parseBlock() error {
	// Read a block label
	name, err := p.parseLabel()
	if err != nil {
		return err
	}

	p.blk = p.fn.NewBlock()
	p.blkLabels[name] = p.blk
	p.fn.InsertBlock(-1, p.blk)

	for {
		tok, lit := p.scanIgnoreWhitespace()

		switch tok {
		case DOT:
			p.unscan()
			return nil

		case FUNC:
			p.unscan()
			return nil

		case IDENT:
			p.unscan()
			if err := p.parseInstr(); err != nil {
				return err
			}

		case EOF:
			p.unscan()
			return nil

		default:
			return fmt.Errorf("found %q, expected block label or instr", lit)
		}
	}
}

func (p *Parser) parseInstr() error {
	var list []string
	var opcode string
	var last string
	var defs []string

	for {
		tok, lit := p.scanIgnoreSpace()

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
		case EQUALS:
			if last != "" {
				list = append(list, last)
				last = ""
			}

			if len(list) == 0 {
				return fmt.Errorf("found %q, expected list", lit)
			}

			// list is def list
			defs = list
			list = nil

		case IDENT, NUM:
			if last == "" {
				last = lit
			} else if opcode == "" {
				opcode = last
				last = lit
			} else {
				return fmt.Errorf("found %q, expected newline %q %q", lit, last, opcode)
			}

		case COMMA:
			p.unscan()
			list = p.parseList(last)
			last = ""

		case NL:
			if opcode == "" {
				opcode = last
			} else if last != "" {
				list = append(list, last)
			}

			return p.addInstr(defs, opcode, list)

		default:
			return fmt.Errorf("found %q, expected instr", lit)
		}
	}
}

var blockRefRe = regexp.MustCompile(`^b\d+$`)
var valueRefRe = regexp.MustCompile(`^v\d+$`)

func (p *Parser) addInstr(defs []string, opcode string, args []string) error {
	opv, err := op.OpString(opcode)
	if err != nil {
		return fmt.Errorf("unknown instruction %q found: %w", opcode, err)
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

	return nil
}

func (p *Parser) parseList(first string) (list []string) {
	if first != "" {
		list = append(list, first)
	}
	for {
		if tok, _ := p.scanIgnoreSpace(); tok != COMMA {
			p.unscan()
			return
		}

		tok, lit := p.scanIgnoreSpace()

		if tok != IDENT {
			p.unscan()
			return
		}

		list = append(list, lit)
	}
}

func (p *Parser) parseLabel() (label string, err error) {
	// read the func name
	tok, lit := p.scanIgnoreWhitespace()
	if tok != IDENT {
		return "", fmt.Errorf("found %q, expected label", lit)
	}

	label = lit

	tok, lit = p.scanIgnoreWhitespace()
	if tok != COLON {
		return "", fmt.Errorf("found %q, expected label", lit)
	}

	return
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	for tok == WS || tok == NL {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) scanIgnoreSpace() (tok Token, lit string) {
	tok, lit = p.scan()
	for tok == WS {
		tok, lit = p.scan()
	}
	return
}
