// Package deobfuscate reduces a closed, bounded AST grammar. It cannot invoke
// input functions, host callbacks, property getters, eval or a JavaScript engine.
package deobfuscate

import (
	"fmt"
	"math"
	"unicode/utf16"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

type Failure struct {
	Code   string   `json:"code"`
	Detail string   `json:"detail,omitempty"`
	Node   ast.Node `json:"-"`
}

func (e *Failure) Error() string                     { return e.Code + ": " + e.Detail }
func fail(code string, node ast.Node, detail string) { panic(&Failure{code, detail, node}) }

type environment struct {
	parent *environment
	values map[ast.Id]any
}

func (e *environment) get(id *ast.Identifier) any {
	for p := e; p != nil; p = p.parent {
		if value, ok := p.values[id.ToId()]; ok {
			return value
		}
	}
	fail("UNRESOLVED_STATIC_IDENTIFIER", id, id.Name)
	return nil
}
func (e *environment) set(id *ast.Identifier, value any) any {
	for p := e; p != nil; p = p.parent {
		if _, ok := p.values[id.ToId()]; ok {
			p.values[id.ToId()] = value
			return value
		}
	}
	fail("NONLOCAL_STATIC_WRITE", id, id.Name)
	return nil
}

type closure struct {
	node *ast.FunctionLiteral
	env  *environment
}
type intrinsic struct {
	name     string
	receiver jsString
}
type returned struct{ value any }

type reducer struct {
	fuel, maxValue int
	intrinsics     bool
	constantOnly   bool
}

func newReducer(intrinsics bool, fuel int) *reducer {
	return &reducer{fuel: fuel, maxValue: 1000000, intrinsics: intrinsics}
}
func (r *reducer) tick(n ast.Node) {
	r.fuel--
	if r.fuel < 0 {
		fail("STATIC_BUDGET_EXCEEDED", n, "")
	}
}
func (r *reducer) limit(v any, n ast.Node) any {
	size := 0
	switch x := v.(type) {
	case jsString:
		size = len(x)
	case []any:
		size = len(x)
	}
	if size > r.maxValue {
		fail("STATIC_VALUE_TOO_LARGE", n, fmt.Sprint(size))
	}
	return v
}

type localVisitor struct {
	ast.NoopVisitor
	root    *ast.FunctionLiteral
	ids     map[ast.Id]bool
	nested  bool
	lexical bool
}

func (v *localVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	if n != v.root {
		v.nested = true
		return
	}
	n.VisitChildrenWith(v)
}
func (v *localVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) { v.nested = true }
func (v *localVisitor) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	if n.Token != token.Var {
		v.lexical = true
	}
	n.VisitChildrenWith(v)
}
func (v *localVisitor) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	if id, ok := n.Target.Target.(*ast.Identifier); ok {
		v.ids[id.ToId()] = true
	} else {
		fail("STATIC_DESTRUCTURING_UNSUPPORTED", n, "")
	}
	n.VisitChildrenWith(v)
}
func locals(fn *ast.FunctionLiteral) *localVisitor {
	v := &localVisitor{root: fn, ids: map[ast.Id]bool{}}
	v.V = v
	fn.VisitWith(v)
	return v
}

func (r *reducer) invoke(c *closure, args []any) (result any) {
	r.tick(c.node)
	fn := c.node
	if fn.Async || fn.Generator {
		fail("ASYNC_STATIC_FUNCTION", fn, "")
	}
	if fn.ParameterList.Rest != nil {
		fail("UNSUPPORTED_STATIC_PARAMETER", fn, "rest")
	}
	env := &environment{parent: c.env, values: map[ast.Id]any{}}
	for i, p := range fn.ParameterList.List {
		id, ok := p.Target.Target.(*ast.Identifier)
		if !ok || p.Initializer != nil {
			fail("UNSUPPORTED_STATIC_PARAMETER", fn, "")
		}
		var value any = undefined{}
		if i < len(args) {
			value = args[i]
		}
		env.values[id.ToId()] = value
	}
	local := locals(fn)
	if local.lexical {
		fail("STATIC_LEXICAL_DECLARATION_UNSUPPORTED", fn, "")
	}
	for id := range local.ids {
		if _, ok := env.values[id]; !ok {
			env.values[id] = undefined{}
		}
	}
	defer func() {
		if caught := recover(); caught != nil {
			if ret, ok := caught.(returned); ok {
				result = ret.value
			} else {
				panic(caught)
			}
		}
	}()
	r.statement(fn.Body, env)
	return undefined{}
}

func (r *reducer) expression(expr ast.Expr, env *environment) any {
	if expr == nil {
		fail("ABSENT_STATIC_EXPRESSION", nil, "")
	}
	r.tick(expr)
	e := func(n *ast.Expression) any {
		if n == nil {
			fail("ABSENT_STATIC_EXPRESSION", expr, "")
		}
		return r.expression(n.Expr, env)
	}
	switch n := expr.(type) {
	case *ast.NumberLiteral:
		return n.Value
	case *ast.BooleanLiteral:
		return n.Value
	case *ast.NullLiteral:
		return nil
	case *ast.StringLiteral:
		return r.limit(stringValue(n.Value), n)
	case *ast.TemplateLiteral:
		if n.Tag == nil && len(n.Expressions) == 0 && len(n.Elements) == 1 && n.Elements[0].Valid {
			return r.limit(stringValue(n.Elements[0].Parsed), n)
		}
	case *ast.Identifier:
		if !r.constantOnly && r.intrinsics && n.Name == "String" && n.ScopeContext == 0 {
			return intrinsic{name: "String"}
		}
		return env.get(n)
	case *ast.FunctionLiteral:
		if !r.constantOnly {
			return &closure{n, env}
		}
	case *ast.ArrayLiteral:
		if r.constantOnly {
			break
		}
		if len(n.Value) > r.maxValue {
			fail("STATIC_VALUE_TOO_LARGE", n, "array")
		}
		v := make([]any, len(n.Value))
		for i, x := range n.Value {
			v[i] = e(&x)
		}
		return v
	case *ast.SequenceExpression:
		var value any = undefined{}
		for _, x := range n.Sequence {
			value = e(&x)
		}
		return value
	case *ast.AssignExpression:
		if r.constantOnly {
			break
		}
		id, ok := n.Left.Expr.(*ast.Identifier)
		if !ok {
			fail("STATIC_PROPERTY_WRITE_FORBIDDEN", n, "")
		}
		if n.Operator == token.Assign {
			return env.set(id, e(n.Right))
		}
		// go-fast stores compound assignments with the binary operator.
		if n.Operator == token.Plus || n.Operator == token.AddAssign {
			a, b := e(n.Left), e(n.Right)
			return env.set(id, r.limit(binary(token.Plus, a, b), n))
		}
		fail("UNSUPPORTED_STATIC_ASSIGNMENT", n, n.Operator.String())
	case *ast.BinaryExpression:
		a := e(n.Left)
		switch n.Operator {
		case token.LogicalAnd:
			if !truthy(a) {
				return a
			}
			return e(n.Right)
		case token.LogicalOr:
			if truthy(a) {
				return a
			}
			return e(n.Right)
		case token.Coalesce:
			if _, ok := a.(undefined); a != nil && !ok {
				return a
			}
			return e(n.Right)
		}
		b := e(n.Right)
		return r.limit(binary(n.Operator, a, b), n)
	case *ast.ConditionalExpression:
		if truthy(e(n.Test)) {
			return e(n.Consequent)
		}
		return e(n.Alternate)
	case *ast.UnaryExpression:
		return unary(n.Operator, e(n.Operand))
	case *ast.UpdateExpression:
		if r.constantOnly {
			break
		}
		id, ok := n.Operand.Expr.(*ast.Identifier)
		if !ok {
			fail("STATIC_PROPERTY_WRITE_FORBIDDEN", n, "")
		}
		old, ok := e(n.Operand).(float64)
		if !ok {
			fail("NON_NUMERIC_STATIC_UPDATE", n, "")
		}
		step := float64(1)
		if n.Operator == token.Decrement {
			step = -1
		} else if n.Operator != token.Increment {
			fail("UNSUPPORTED_STATIC_UPDATE", n, "")
		}
		next := old + step
		env.set(id, next)
		if n.Postfix {
			return old
		}
		return next
	case *ast.MemberExpression:
		if r.constantOnly {
			break
		}
		object := e(n.Object)
		key := r.property(n, env)
		switch value := object.(type) {
		case []any:
			if key == "length" {
				return float64(len(value))
			}
			i := number(stringValue(key))
			if math.IsNaN(i) || math.Trunc(i) != i || i < 0 || i >= float64(len(value)) {
				fail("STATIC_ARRAY_BOUNDS", n, key)
			}
			return value[int(i)]
		case jsString:
			if key == "length" {
				return float64(len(value))
			}
			if r.intrinsics && (key == "charCodeAt" || key == "codePointAt" || key == "charAt") {
				return intrinsic{key, value}
			}
		case intrinsic:
			if value.name == "String" && (key == "fromCodePoint" || key == "fromCharCode") {
				return intrinsic{name: key}
			}
		}
		fail("STATIC_PROPERTY_FORBIDDEN", n, key)
	case *ast.CallExpression:
		if r.constantOnly {
			break
		}
		fn := e(n.Callee)
		args := make([]any, len(n.ArgumentList))
		for i, x := range n.ArgumentList {
			args[i] = e(&x)
			if !primitive(args[i]) {
				fail("NON_PRIMITIVE_INTRINSIC_ARGUMENT", n, "")
			}
		}
		f, ok := fn.(intrinsic)
		if !ok {
			fail("UNKNOWN_STATIC_CALL", n, fmt.Sprintf("%T", fn))
		}
		return r.intrinsicCall(f, args, n)
	}
	fail("UNSUPPORTED_STATIC_SYNTAX", expr, fmt.Sprintf("%T", expr))
	return nil
}

func (r *reducer) property(n *ast.MemberExpression, env *environment) string {
	if n.Property == nil {
		fail("ABSENT_STATIC_PROPERTY", n, "")
	}
	switch p := n.Property.Prop.(type) {
	case *ast.Identifier:
		return p.Name
	case *ast.ComputedProperty:
		v := r.expression(p.Expr.Expr, env)
		if !primitive(v) {
			fail("STATIC_PROPERTY_FORBIDDEN", n, "non-primitive")
		}
		s, err := textValue(v).utf8()
		if err != nil {
			fail("STATIC_PROPERTY_FORBIDDEN", n, err.Error())
		}
		return s
	}
	fail("STATIC_PROPERTY_FORBIDDEN", n, "")
	return ""
}

func (r *reducer) intrinsicCall(fn intrinsic, args []any, node ast.Node) any {
	switch fn.name {
	case "codePointAt", "charCodeAt", "charAt":
		if len(args) != 1 {
			fail("STATIC_STRING_BOUNDS", node, "arity")
		}
		x, ok := args[0].(float64)
		if !ok || math.IsNaN(x) || math.Trunc(x) != x || x < 0 || x >= float64(len(fn.receiver)) {
			fail("STATIC_STRING_BOUNDS", node, "")
		}
		i := int(x)
		u := fn.receiver[i]
		if fn.name == "charAt" {
			return jsString{u}
		}
		if fn.name == "codePointAt" && u >= 0xd800 && u <= 0xdbff && i+1 < len(fn.receiver) && fn.receiver[i+1] >= 0xdc00 && fn.receiver[i+1] <= 0xdfff {
			return float64(utf16.DecodeRune(rune(u), rune(fn.receiver[i+1])))
		}
		return float64(u)
	case "fromCodePoint", "fromCharCode":
		if len(args) > r.maxValue/2 {
			fail("STATIC_VALUE_TOO_LARGE", node, "intrinsic arguments")
		}
		out := jsString{}
		for _, arg := range args {
			if fn.name == "fromCharCode" {
				out = append(out, uint16(uint32Value(arg)))
				continue
			}
			x := number(arg)
			if math.IsNaN(x) || math.Trunc(x) != x || x < 0 || x > 0x10ffff {
				fail("STATIC_CODE_POINT_RANGE", node, numberText(x))
			}
			if x <= 0xffff {
				out = append(out, uint16(x))
			} else {
				a, b := utf16.EncodeRune(rune(x))
				out = append(out, uint16(a), uint16(b))
			}
		}
		return r.limit(out, node)
	}
	fail("UNKNOWN_STATIC_INTRINSIC", node, fn.name)
	return nil
}

func (r *reducer) declarations(n *ast.VariableDeclaration, env *environment) {
	if n.Token != token.Var {
		fail("STATIC_LEXICAL_DECLARATION_UNSUPPORTED", n, "")
	}
	for _, d := range n.List {
		id, ok := d.Target.Target.(*ast.Identifier)
		if !ok {
			fail("STATIC_DESTRUCTURING_UNSUPPORTED", n, "")
		}
		if d.Initializer != nil {
			env.set(id, r.expression(d.Initializer.Expr, env))
		}
	}
}

func (r *reducer) statement(stmt ast.Stmt, env *environment) {
	r.tick(stmt)
	e := func(n *ast.Expression) any { return r.expression(n.Expr, env) }
	switch n := stmt.(type) {
	case *ast.BlockStatement:
		for _, s := range n.List {
			r.statement(s.Stmt, env)
		}
	case *ast.ExpressionStatement:
		e(n.Expression)
	case *ast.EmptyStatement:
	case *ast.VariableDeclaration:
		r.declarations(n, env)
	case *ast.ReturnStatement:
		var value any = undefined{}
		if n.Argument != nil {
			value = e(n.Argument)
		}
		panic(returned{value})
	case *ast.IfStatement:
		if truthy(e(n.Test)) {
			r.statement(n.Consequent.Stmt, env)
		} else if n.Alternate != nil {
			r.statement(n.Alternate.Stmt, env)
		}
	case *ast.ForStatement:
		if n.Initializer != nil {
			switch init := n.Initializer.Initializer.(type) {
			case *ast.VariableDeclaration:
				r.declarations(init, env)
			case *ast.Expression:
				e(init)
			default:
				fail("UNSUPPORTED_STATIC_FOR_INITIALIZER", n, "")
			}
		}
		for n.Test == nil || truthy(e(n.Test)) {
			r.tick(n)
			r.statement(n.Body.Stmt, env)
			if n.Update != nil {
				e(n.Update)
			}
		}
	default:
		fail("UNSUPPORTED_STATIC_STATEMENT", stmt, fmt.Sprintf("%T", stmt))
	}
}

func constant(expr ast.Expr) (value any, ok bool) {
	defer func() {
		if caught := recover(); caught != nil {
			if _, known := caught.(*Failure); known {
				value = nil
				ok = false
			} else {
				panic(caught)
			}
		}
	}()
	r := newReducer(false, 2000)
	r.constantOnly = true
	return r.expression(expr, nil), true
}
