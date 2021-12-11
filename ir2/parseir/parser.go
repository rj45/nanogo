package parseir

import (
	"fmt"
	"io"
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
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader, prog *ir2.Program) *Parser {
	return &Parser{s: NewScanner(r), prog: prog}
}

func (p *Parser) Parse() error {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != PACKAGE {
		return fmt.Errorf("found %q, expected package", lit)
	}

	tok, lit = p.scanIgnoreWhitespace()
	if tok != IDENT {
		return fmt.Errorf("found %q, expected package name", lit)
	}

	p.pkg = p.prog.Package(lit)
	if p.pkg == nil {
		p.pkg = &ir2.Package{Name: lit, Path: "todo"}
		p.prog.AddPackage(p.pkg)
	}

	for {
		if tok, lit := p.scanIgnoreWhitespace(); tok != FUNC {
			if tok == EOF {
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

	p.fn = p.pkg.NewFunc(name, nil)
	p.blk = nil

	p.blkLabels = make(map[string]*ir2.Block)
	p.values = make(map[string]*ir2.Value)

	fmt.Println("func:", name)

	for {
		tok, lit := p.scanIgnoreWhitespace()

		switch tok {
		case DOT:
			if err := p.parseBlock(); err != nil {
				return err
			}

		case FUNC:
			p.unscan()
			return nil

		case EOF:
			p.unscan()
			return nil

		default:
			return fmt.Errorf("found %q, expected func or block label", lit)
		}
	}
}

func (p *Parser) parseBlock() error {
	// Read a func label
	name, err := p.parseLabel()
	if err != nil {
		return err
	}

	p.fn = p.pkg.NewFunc(name, nil)
	p.blk = nil

	p.blkLabels = make(map[string]*ir2.Block)

	for {
		tok, lit := p.scanIgnoreWhitespace()

		switch tok {
		case DOT:
			// block label
			name, err := p.parseLabel()
			if err != nil {
				return err
			}

			p.blk = p.fn.NewBlock()
			p.blkLabels[name] = p.blk

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

func (p *Parser) addInstr(defs []string, opcode string, args []string) error {
	opv, err := op.OpString(opcode)
	if err != nil {
		return fmt.Errorf("unknown instruction %q found: %w", opcode, err)
	}

	ins := p.fn.NewInstr(opv, nil)

	for _, def := range defs {
		// todo: fix type here
		v := ins.AddDef(p.fn.NewValue(nil))
		p.values[def] = v
	}

	for _, arg := range args {
		val, ok := p.values[arg]
		if !ok {
			if strings.HasPrefix(arg, "v") {
				panic("need to implement deferred resolution")
			}

			val = p.fn.ValueFor(nil, arg)
		}
		ins.InsertArg(-1, val)
	}

	fmt.Println(defs, opcode, args)
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
		return "", fmt.Errorf("found %q, expected func label", lit)
	}

	label = lit

	tok, lit = p.scanIgnoreWhitespace()
	if tok != COLON {
		return "", fmt.Errorf("found %q, expected func label", lit)
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
