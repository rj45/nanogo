package parseir

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"strings"

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
	globals   map[string]map[*ir2.Func]*ir2.Value

	// forward reference links
	blkLinks map[*ir2.Block][]string

	// debugging / diagnostics
	indent int
	trace  bool
}

// NewParser returns a new instance of Parser.
func NewParser(filename string, r io.Reader, prog *ir2.Program, trace bool) (*Parser, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", filename, err)
	}
	fset := token.NewFileSet()
	file := fset.AddFile(filename, -1, len(buf))
	errs := scanner.ErrorList{}
	s := &scanner.Scanner{}
	s.Init(file, buf, errs.Add, 0)

	return &Parser{fset: fset, s: s, errs: errs, prog: prog, trace: trace}, nil
}

// ParseString parses an IR listing and returns the func for that listing, which
// is useful in tests. Note, if the `text` doesn't contain `func main:` then a
// template is used to put the `text` into the main function, useful for
// reducing boilerplate.
func ParseString(text string) (*ir2.Func, error) {
	prog := &ir2.Program{}

	fullText := fmt.Sprintf(`
		package main "test"
		func main:
		%s
	`, text)

	if strings.Contains(text, "func main:") {
		fullText = text
	}

	parser, err := NewParser("test.ngir", strings.NewReader(fullText), prog, false)
	if err != nil {
		return nil, err
	}

	err = parser.Parse()
	if err != nil {
		return nil, err
	}

	return prog.Func("main__main"), nil
}

func (p *Parser) Parse() error {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		p.PrintErrors()
	// 		log.Fatal(r)
	// 	}
	// }()
	p.parse()

	if p.errs.Len() < 1 {
		return nil
	}
	return p.errs
}

func (p *Parser) PrintErrors() {
	scanner.PrintError(os.Stderr, p.errs)
}

func (p *Parser) printTrace(a ...interface{}) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	pos := p.fset.Position(p.buf.pos)
	fmt.Printf("%5d:%3d: ", pos.Line, pos.Column)
	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *Parser, msg string) *Parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *Parser) {
	p.indent--
	p.printTrace(")")
}

// expect checks if the next token is expected, erroring if not
func (p *Parser) expect(expected token.Token, what string) (tok token.Token, lit string) {
	tok, lit = p.scan()
	if tok != expected {
		p.errorf("expected %s parsing %s, got %s:%q", expected, what, tok, lit)
	}

	return
}

// errorf records an error at the current token scan position
func (p *Parser) errorf(format string, args ...interface{}) {
	if p.trace {
		p.printTrace(fmt.Sprintf("error: "+format, args...))
	}

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

	if p.trace && p.buf.pos.IsValid() {
		s := p.buf.tok.String()
		switch {
		case p.buf.tok.IsLiteral():
			p.printTrace(s, p.buf.lit)
		case p.buf.tok.IsOperator(), p.buf.tok.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	// Otherwise read the next token from the scanner.
	pos, tok, lit := p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.pos, p.buf.tok, p.buf.lit = pos, tok, lit

	return tok, lit
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
