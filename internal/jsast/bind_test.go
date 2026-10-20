package jsast

import (
	"testing"

	"github.com/t14raptor/go-fast/ast"
)

type identifiers struct {
	ast.NoopVisitor
	byName map[string][]*ast.Identifier
}

func (v *identifiers) VisitIdentifier(n *ast.Identifier) {
	v.byName[n.Name] = append(v.byName[n.Name], n)
}
func bindingFixture(t *testing.T, text string) map[string][]*ast.Identifier {
	t.Helper()
	s, err := Parse("bindings.js", []byte(text))
	if err != nil {
		t.Fatal(err)
	}
	Bind(s.AST)
	v := &identifiers{byName: map[string][]*ast.Identifier{}}
	v.V = v
	s.AST.VisitWith(v)
	return v.byName
}
func TestBindKeepsGlobalsFreeAndHoistsVar(t *testing.T) {
	ids := bindingFixture(t, `function f(){String.fromCharCode(2);x;{var x=1;}x;}String;`)
	for _, id := range ids["String"] {
		if id.ScopeContext != 0 {
			t.Fatal("global String falsely bound", id)
		}
	}
	x := ids["x"]
	if len(x) != 3 || x[0].ScopeContext == 0 || x[0].ToId() != x[1].ToId() || x[0].ToId() != x[2].ToId() {
		t.Fatal(x)
	}
}
func TestBindSeparatesLexicalCatchAndParameterBindings(t *testing.T) {
	ids := bindingFixture(t, `var x=1;function f(x){x;{let x=2;x;}try{}catch(x){x;}x;}x;`)
	x := ids["x"]
	if len(x) != 9 {
		t.Fatalf("%d", len(x))
	}
	if x[0].ToId() != x[8].ToId() || x[1].ToId() != x[2].ToId() || x[1].ToId() != x[7].ToId() || x[3].ToId() != x[4].ToId() || x[5].ToId() != x[6].ToId() {
		t.Fatal("binding mismatch")
	}
	if x[0].ToId() == x[1].ToId() || x[1].ToId() == x[3].ToId() || x[3].ToId() == x[5].ToId() {
		t.Fatal("shadowing collapsed")
	}
}
