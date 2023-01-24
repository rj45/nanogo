package rewriter

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"io"
	"strings"
)

func Parse(src io.Reader) ([]*Rule, error) {
	s := bufio.NewScanner(src)
	lineno := 0
	var exprs []*Rule
	line := ""
	for s.Scan() {
		lineno++
		line += s.Text()

		parts := strings.Split(line, "=>")
		if len(parts) < 2 || !parensMatched(line) || strings.TrimSpace(parts[1]) == "" {
			continue
		}
		line = ""

		lhs, err := parser.ParseExpr(parts[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse lhs on line %d: %w", lineno, err)
		}
		from, err := translate(lhs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse lhs on line %d: %w", lineno, err)
		}
		rhs, err := parser.ParseExpr(parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse lhs on line %d: %w", lineno, err)
		}
		to, err := translate(rhs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse lhs on line %d: %w", lineno, err)
		}

		exprs = append(exprs, &Rule{From: from, To: to})
	}
	return exprs, s.Err()
}

func parensMatched(line string) bool {
	depth := 0
	for _, c := range line {
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
		}
	}
	return depth == 0
}

func translate(expr ast.Expr) (*Node, error) {
	switch n := expr.(type) {
	case *ast.CallExpr:
		args := make([]*Node, len(n.Args))
		for i, arg := range n.Args {
			t, err := translate(arg)
			if err != nil {
				return nil, err
			}
			args[i] = t
		}
		return &Node{Kind: Call, Name: n.Fun.(*ast.Ident).Name, Args: args}, nil
	case *ast.Ident:
		if n.Name == "nil" {
			return &Node{Kind: Nil}, nil
		}
		return &Node{Kind: Ident, Name: n.Name}, nil
	}
	return nil, fmt.Errorf("unknown ast node: %#v", expr)
}
