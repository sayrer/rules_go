package generator

import (
	"io/ioutil"
	"os"

	bzl "github.com/bazelbuild/buildifier/core"
)

func (g *Generator) reconcile(generated *bzl.File) (*bzl.File, error) {
	f, err := readExisting(generated.Path)
	if err != nil {
		return nil, err
	}
	g.reconcileLoad(f, generated)
	g.reconcileRules(f, generated)
	return f, nil
}

// reconcileRules inserts go rules in "generated" into "dest" or replace
// the existings.
func (g *Generator) reconcileRules(dest, generated *bzl.File) {
	existings := make(map[string]*bzl.Rule)
	for _, r := range dest.Rules("") {
		if isGoRule(r) {
			existings[r.Name()] = r
		}
	}

	var newRules []bzl.Expr
	for _, r := range generated.Rules("") {
		e := existings[r.Name()]
		if e == nil {
			newRules = append(newRules, r.Call)
			continue
		}
		g.reconcileRule(e, r)
		delete(existings, r.Name())
	}

	for _, r := range existings {
		dest.DelRules(r.Kind(), r.Name())
	}
	dest.Stmt = append(dest.Stmt, newRules...)
}

func (g *Generator) reconcileRule(dest, generated *bzl.Rule) {
	dest.SetKind(generated.Kind())
	for _, attr := range []string{"srcs", "deps", "library"} {
		expr := generated.AttrDefn(attr)
		if expr == nil {
			dest.DelAttr(attr)
			continue
		}
		dest.SetAttr(attr, expr.Y)
	}
}

func (g *Generator) reconcileLoad(dest, generated *bzl.File) {
	head, tail := splitAfterPackage(dest)

	stmts := append([]bzl.Expr(nil), head...)
	if load := g.generateLoad(generated); load != nil {
		stmts = append(stmts, load)
	}
	for _, stmt := range tail {
		if g.isLoadingRulesGo(stmt) {
			continue
		}
		stmts = append(stmts, stmt)
	}

	dest.Stmt = stmts
}

func (g *Generator) isLoadingRulesGo(expr bzl.Expr) bool {
	c, ok := expr.(*bzl.CallExpr)
	if !ok {
		return false
	}
	r := bzl.Rule{Call: c}
	if r.Kind() != "load" {
		return false
	}
	if len(c.List) < 1 {
		return false
	}
	str, ok := c.List[0].(*bzl.StringExpr)
	if !ok {
		return false
	}
	return str.Value == goRulesBzl
}

func (g *Generator) generateLoad(f *bzl.File) bzl.Expr {
	var list []string
	for _, kind := range goRuleKinds {
		if len(f.Rules(kind)) > 0 {
			list = append(list, kind)
		}
	}
	if len(list) == 0 {
		return nil
	}
	return loadExpr(goRulesBzl, list...)
}

func loadExpr(ruleFile string, rules ...string) bzl.Expr {
	var list []bzl.Expr
	for _, r := range append([]string{ruleFile}, rules...) {
		list = append(list, &bzl.StringExpr{Value: r})
	}

	return &bzl.CallExpr{
		X:            &bzl.LiteralExpr{Token: "load"},
		List:         list,
		ForceCompact: true,
	}
}

func splitAfterPackage(f *bzl.File) (head, tail []bzl.Expr) {
	for i, stmt := range f.Stmt {
		if c, ok := stmt.(*bzl.CallExpr); ok {
			r := bzl.Rule{Call: c}
			if r.Kind() == "package" {
				return f.Stmt[:i+1], f.Stmt[i+1:]
			}
		}
	}
	return nil, f.Stmt
}

func readExisting(fname string) (*bzl.File, error) {
	buf, err := ioutil.ReadFile(fname)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if len(buf) == 0 {
		return &bzl.File{Path: fname}, nil
	}
	return bzl.Parse(fname, buf)
}

var (
	goRuleKinds = []string{
		"go_prefix",
		"go_library",
		"go_binary",
		"go_test",
		// TODO(yugui) Support cgo_library
	}
)

func isGoRule(r *bzl.Rule) bool {
	k := r.Kind()
	for _, g := range goRuleKinds {
		if k == g {
			return true
		}
	}
	return false
}
