package rewriter_test

import (
	"strings"
	"testing"

	"github.com/rj45/nanogo/rewriter"
)

func TestEmptyToEmpty(t *testing.T) {
	rules, err := rewriter.Parse(strings.NewReader("test() => nil"))
	if err != nil {
		t.Error(err)
	}
	if len(rules) != 1 {
		t.Errorf("expected one rule; got %d", len(rules))
	}
	if rules[0].From == nil || rules[0].To == nil {
		t.Errorf("expected non nil rule; got %#v", rules[0])
	}
}

func TestMultiLineLHS(t *testing.T) {
	rules, err := rewriter.Parse(strings.NewReader("test(\n) => nil"))
	if err != nil {
		t.Error(err)
	}
	if len(rules) != 1 {
		t.Errorf("expected one rule; got %d", len(rules))
	}
	if rules[0].From == nil || rules[0].To == nil {
		t.Errorf("expected non nil rule; got %#v", rules[0])
	}
}

func TestMultiLineRHS(t *testing.T) {
	rules, err := rewriter.Parse(strings.NewReader("test(\n) => another(test(\n))"))
	if err != nil {
		t.Error(err)
	}
	if len(rules) != 1 {
		t.Errorf("expected one rule; got %d", len(rules))
	}
	if rules[0].From == nil || rules[0].To == nil {
		t.Errorf("expected non nil rule; got %#v", rules[0])
	}
}

func TestNewlineBeforeRHS(t *testing.T) {
	rules, err := rewriter.Parse(strings.NewReader("test(\n) =>\nanother(test(\n))"))
	if err != nil {
		t.Error(err)
	}
	if len(rules) != 1 {
		t.Errorf("expected one rule; got %d", len(rules))
	}
	if rules[0].From == nil || rules[0].To == nil {
		t.Errorf("expected non nil rule; got %#v", rules[0])
	}
}

func TestTranslateAst(t *testing.T) {
	rule := getFirstRule(t, "test(x, y) => nil")
	if rule.From.Kind != rewriter.Call {
		t.Errorf("expected call; got %#v", rule.From)
	}

	if rule.From.Name != "test" {
		t.Errorf("expected call name to be test; got %#v", rule.From.Name)
	}

	if len(rule.From.Args) != 2 {
		t.Errorf("expected two args; got %#v", rule.From.Args)
	}

	if rule.From.Args[0].Kind != rewriter.Ident || rule.From.Args[0].Name != "x" {
		t.Errorf("expected first arg to be an ident 'x'; got %#v", rule.From.Args[0])
	}

	if rule.To.Kind != rewriter.Nil {
		t.Errorf("expected nil; got %#v", rule.To)
	}
}

func getFirstRule(t *testing.T, text string) *rewriter.Rule {
	rules, err := rewriter.Parse(strings.NewReader(text))
	if err != nil {
		t.Error(err)
	}
	if len(rules) != 1 {
		t.Errorf("expected one rule; got %d", len(rules))
	}
	if rules[0].From == nil || rules[0].To == nil {
		t.Errorf("expected non nil rule; got %#v", rules[0])
	}

	return rules[0]
}
