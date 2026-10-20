package jsast

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

// Bind resolves declarations before references and leaves free globals at zero.
// ScopeContext is analysis metadata; this pass never changes source text.
func Bind(program *ast.Program) {
	b := &binder{}
	b.V = b
	b.push(true)
	b.hoist(program)
	b.lexical(program.Body)
	program.VisitChildrenWith(b)
}

type bindingScope struct {
	parent   *bindingScope
	context  ast.ScopeContext
	function bool
	names    map[string]bool
}
type binder struct {
	ast.NoopVisitor
	current *bindingScope
	next    ast.ScopeContext
}

func (b *binder) push(function bool) {
	b.next++
	b.current = &bindingScope{parent: b.current, context: b.next, function: function, names: map[string]bool{}}
}
func (b *binder) pop() { b.current = b.current.parent }
func (b *binder) variableScope() *bindingScope {
	for s := b.current; s != nil; s = s.parent {
		if s.function {
			return s
		}
	}
	return b.current
}
func declare(scope *bindingScope, id *ast.Identifier) {
	if id != nil {
		scope.names[id.Name] = true
		id.ScopeContext = scope.context
	}
}

func targets(expr ast.Expr, visit func(*ast.Identifier)) {
	if expr == nil {
		return
	}
	switch n := expr.(type) {
	case *ast.Identifier:
		visit(n)
	case *ast.ArrayPattern:
		for _, e := range n.Elements {
			targets(e.Expr, visit)
		}
		if n.Rest != nil {
			targets(n.Rest.Expr, visit)
		}
	case *ast.ObjectPattern:
		for _, p := range n.Properties {
			switch x := p.Prop.(type) {
			case *ast.PropertyShort:
				visit(x.Name)
			case *ast.PropertyKeyed:
				targets(x.Value.Expr, visit)
			}
		}
		targets(n.Rest, visit)
	case *ast.AssignExpression:
		targets(n.Left.Expr, visit)
	case *ast.SpreadElement:
		targets(n.Expression.Expr, visit)
	}
}

type hoistVisitor struct {
	ast.NoopVisitor
	scope *bindingScope
}

func (v *hoistVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral)           {}
func (v *hoistVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {}
func (v *hoistVisitor) VisitClassLiteral(n *ast.ClassLiteral)                 {}
func (v *hoistVisitor) VisitClassDeclaration(n *ast.ClassDeclaration)         {}
func (v *hoistVisitor) VisitClassStaticBlock(n *ast.ClassStaticBlock)         {}
func (v *hoistVisitor) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	declare(v.scope, n.Function.Name)
}
func (v *hoistVisitor) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	if n.Token == token.Var {
		for _, d := range n.List {
			targets(d.Target.Target, func(id *ast.Identifier) { declare(v.scope, id) })
		}
	}
}
func (b *binder) hoist(node ast.VisitableNode) {
	v := &hoistVisitor{scope: b.variableScope()}
	v.V = v
	node.VisitChildrenWith(v)
}
func (b *binder) lexical(statements ast.Statements) {
	for _, s := range statements {
		switch n := s.Stmt.(type) {
		case *ast.VariableDeclaration:
			if n.Token != token.Var {
				for _, d := range n.List {
					targets(d.Target.Target, func(id *ast.Identifier) { declare(b.current, id) })
				}
			}
		case *ast.ClassDeclaration:
			declare(b.current, n.Class.Name)
		case *ast.FunctionDeclaration:
			declare(b.current, n.Function.Name)
		}
	}
}
func (b *binder) VisitIdentifier(n *ast.Identifier) {
	n.ScopeContext = 0
	for s := b.current; s != nil; s = s.parent {
		if s.names[n.Name] {
			n.ScopeContext = s.context
			return
		}
	}
}
func (b *binder) VisitMemberProperty(n *ast.MemberProperty) {
	if p, ok := n.Prop.(*ast.ComputedProperty); ok {
		p.VisitWith(b)
	}
}
func (b *binder) VisitPropertyKeyed(n *ast.PropertyKeyed) {
	if n.Computed {
		n.Key.VisitWith(b)
	}
	n.Value.VisitWith(b)
}
func (b *binder) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	scope := b.current
	if n.Token == token.Var {
		scope = b.variableScope()
	}
	for _, d := range n.List {
		targets(d.Target.Target, func(id *ast.Identifier) { declare(scope, id) })
		b.patternEffects(d.Target.Target)
		if d.Initializer != nil {
			d.Initializer.VisitWith(b)
		}
	}
}

// Default values and computed destructuring keys are expressions, not bindings.
func (b *binder) patternEffects(expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.ArrayPattern:
		for _, e := range n.Elements {
			b.patternEffects(e.Expr)
		}
		if n.Rest != nil {
			b.patternEffects(n.Rest.Expr)
		}
	case *ast.ObjectPattern:
		for _, p := range n.Properties {
			switch x := p.Prop.(type) {
			case *ast.PropertyShort:
				if x.Initializer != nil {
					x.Initializer.VisitWith(b)
				}
			case *ast.PropertyKeyed:
				if x.Computed {
					x.Key.VisitWith(b)
				}
				b.patternEffects(x.Value.Expr)
			}
		}
		b.patternEffects(n.Rest)
	case *ast.AssignExpression:
		n.Right.VisitWith(b)
		b.patternEffects(n.Left.Expr)
	}
}
func (b *binder) parameters(p *ast.ParameterList) {
	for _, d := range p.List {
		targets(d.Target.Target, func(id *ast.Identifier) { declare(b.current, id) })
	}
	targets(p.Rest, func(id *ast.Identifier) { declare(b.current, id) })
	for _, d := range p.List {
		b.patternEffects(d.Target.Target)
		if d.Initializer != nil {
			d.Initializer.VisitWith(b)
		}
	}
}
func (b *binder) function(n *ast.FunctionLiteral, declaration bool) {
	b.push(true)
	n.ScopeContext = b.current.context
	if !declaration {
		declare(b.current, n.Name)
	}
	b.parameters(&n.ParameterList)
	b.hoist(n.Body)
	n.Body.ScopeContext = b.current.context
	b.lexical(n.Body.List)
	n.Body.VisitChildrenWith(b)
	b.pop()
}
func (b *binder) VisitFunctionLiteral(n *ast.FunctionLiteral) { b.function(n, false) }
func (b *binder) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	b.VisitIdentifier(n.Function.Name)
	b.function(n.Function, true)
}
func (b *binder) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	b.push(true)
	n.ScopeContext = b.current.context
	b.parameters(&n.ParameterList)
	switch body := n.Body.Body.(type) {
	case *ast.BlockStatement:
		b.hoist(body)
		body.ScopeContext = b.current.context
		b.lexical(body.List)
		body.VisitChildrenWith(b)
	case *ast.Expression:
		body.VisitWith(b)
	}
	b.pop()
}
func (b *binder) VisitBlockStatement(n *ast.BlockStatement) {
	b.push(false)
	n.ScopeContext = b.current.context
	b.lexical(n.List)
	n.VisitChildrenWith(b)
	b.pop()
}
func (b *binder) VisitCatchStatement(n *ast.CatchStatement) {
	b.push(false)
	if n.Parameter != nil {
		targets(n.Parameter.Target, func(id *ast.Identifier) { declare(b.current, id) })
		b.patternEffects(n.Parameter.Target)
	}
	n.Body.VisitWith(b)
	b.pop()
}
func (b *binder) VisitForStatement(n *ast.ForStatement) {
	b.push(false)
	n.VisitChildrenWith(b)
	b.pop()
}
func (b *binder) VisitForInStatement(n *ast.ForInStatement) {
	b.push(false)
	n.VisitChildrenWith(b)
	b.pop()
}
func (b *binder) VisitForOfStatement(n *ast.ForOfStatement) {
	b.push(false)
	n.VisitChildrenWith(b)
	b.pop()
}
func (b *binder) VisitSwitchStatement(n *ast.SwitchStatement) {
	n.Discriminant.VisitWith(b)
	b.push(false)
	for _, c := range n.Body {
		b.lexical(c.Consequent)
	}
	for i := range n.Body {
		n.Body[i].VisitWith(b)
	}
	b.pop()
}
func (b *binder) class(n *ast.ClassLiteral) {
	b.push(false)
	declare(b.current, n.Name)
	if n.SuperClass != nil {
		n.SuperClass.VisitWith(b)
	}
	for i := range n.Body {
		n.Body[i].VisitWith(b)
	}
	b.pop()
}
func (b *binder) VisitClassDeclaration(n *ast.ClassDeclaration) { b.class(n.Class) }
func (b *binder) VisitClassLiteral(n *ast.ClassLiteral)         { b.class(n) }
func (b *binder) VisitMethodDefinition(n *ast.MethodDefinition) {
	if n.Computed {
		n.Key.VisitWith(b)
	}
	b.function(n.Body, false)
}
func (b *binder) VisitFieldDefinition(n *ast.FieldDefinition) {
	if n.Computed {
		n.Key.VisitWith(b)
	}
	if n.Initializer != nil {
		n.Initializer.VisitWith(b)
	}
}
func (b *binder) VisitClassStaticBlock(n *ast.ClassStaticBlock) {
	b.push(true)
	b.hoist(n.Block)
	n.Block.VisitWith(b)
	b.pop()
}
