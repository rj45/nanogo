package parseir

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"

	"github.com/rj45/nanogo/ir2"
)

type Parser struct {
	fset *token.FileSet
	s    *scanner.Scanner
	errs scanner.ErrorList

	buf struct {
		pos token.Pos   // position of last read token
		tok token.Token // last read token
		lit string      // last read literal
		n   int         // buffer size (max=1)
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
func NewParser(filename string, r io.Reader, prog *ir2.Program) (*Parser, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", filename, err)
	}
	fset := token.NewFileSet()
	file := fset.AddFile(filename, -1, len(buf))
	errs := scanner.ErrorList{}
	s := &scanner.Scanner{}
	s.Init(file, buf, errs.Add, 0)

	return &Parser{fset: fset, s: s, errs: errs, prog: prog}, nil
}

func (p *Parser) Parse() error {
	defer func() {
		if r := recover(); r != nil {
			p.PrintErrors()
			panic(r)
		}
	}()
	p.parse()

	if p.errs.Len() < 1 {
		return nil
	}
	return p.errs
}

func (p *Parser) PrintErrors() {
	scanner.PrintError(os.Stderr, p.errs)
}

// errorf records an error at the current token scan position
func (p *Parser) errorf(format string, args ...interface{}) {
	pos := p.fset.Position(p.buf.pos)
	p.errs.Add(pos, fmt.Sprintf(format, args...))
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok token.Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	pos, tok, lit := p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.pos, p.buf.tok, p.buf.lit = pos, tok, lit

	return tok, lit
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
