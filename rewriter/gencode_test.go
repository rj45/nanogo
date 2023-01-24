package rewriter_test

import (
	"bytes"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/rj45/nanogo/rewriter"
)

func TestEmpty(t *testing.T) {
	expected := `{
	{
	}
}`

	got := genCodeFor(t, `nil => nil`)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestMatchAndRemoveA(t *testing.T) {
	expected := `{
	{
		if ok := it.a(); ok {
			it.Remove()
		}
	}
}`

	got := genCodeFor(t, `a() => nil`)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestMatchAndRemoveA_withEmptyArgs(t *testing.T) {
	expected := `{
	{
		if _, _, ok := it.a(); ok {
			it.Remove()
		}
	}
}`

	got := genCodeFor(t, `a(_, _) => nil`)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestMatchAndRemoveA_withNestedB(t *testing.T) {
	expected := `{
	{
		if t0, ok := it.a(); ok {
			if ok := t0.b(); ok {
				it.Remove()
			}
		}
	}
}`

	got := genCodeFor(t, `a(b()) => nil`)
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestMatchAndRemoveA_withNestedBC(t *testing.T) {
	expected := `{
	{
		if t0, t2, ok := it.a(); ok {
			if t1, ok := t0.b(); ok {
				if ok := t1.d(); ok {
					if ok := t2.c(); ok {
						it.Remove()
					}
				}
			}
		}
	}
}`

	got := genCodeFor(t, `a(b(d()),c()) => nil`)
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestMatchAndReplace(t *testing.T) {
	expected := `{
	{
		if x, ok := it.a(); ok {
			it.Replace(x)
		}
	}
}`

	got := genCodeFor(t, `a(x) => x`)
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestMatchAndReplaceWithPattern(t *testing.T) {
	expected := `{
	{
		if x, ok := it.a(); ok {
			it.Replace(b.b(x))
		}
	}
}`

	got := genCodeFor(t, `a(x) => b(x)`)
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func genCodeFor(t *testing.T, text string) string {
	rules, err := rewriter.Parse(strings.NewReader(text))
	if err != nil {
		t.Error(err)
	}

	blk := rewriter.GenRuleCodeBlocks(rules)
	buf := &bytes.Buffer{}
	printer.Fprint(buf, token.NewFileSet(), blk)

	return buf.String()
}
