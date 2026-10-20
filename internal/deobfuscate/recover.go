package deobfuscate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
	"recaptchav3research/internal/jsast"
)

type Options struct {
	AssumeStandardIntrinsics  bool `json:"assumeStandardIntrinsics"`
	AssumeInitializedDecoders bool `json:"assumeInitializedDecoders"`
	Fuel                      int  `json:"fuel"`
}
type Invariant struct {
	Kind                 string `json:"kind"`
	Cache                string `json:"cache,omitempty"`
	RequiresTruthyResult bool   `json:"requiresTruthyResult"`
}
type StringRecord struct {
	Binding     string         `json:"binding"`
	Factory     string         `json:"factory"`
	Evidence    jsast.Evidence `json:"evidence"`
	Arguments   []any          `json:"arguments"`
	Status      string         `json:"status"`
	Invariant   *Invariant     `json:"invariant,omitempty"`
	Value       *string        `json:"value,omitempty"`
	ValueSHA256 string         `json:"valueSha256,omitempty"`
	Reason      string         `json:"reason,omitempty"`
	Detail      string         `json:"detail,omitempty"`
	id          ast.Id
	units       jsString
}
type Pool struct {
	CodeUnits int            `json:"codeUnits"`
	SHA256    string         `json:"sha256"`
	Evidence  jsast.Evidence `json:"evidence"`
}
type Factory struct {
	Name     string         `json:"name"`
	Evidence jsast.Evidence `json:"evidence"`
	Pools    []Pool         `json:"pools"`
	fn       *ast.FunctionLiteral
}
type Table struct {
	Name     string         `json:"name"`
	Stable   bool           `json:"stable"`
	Members  int            `json:"members"`
	Evidence jsast.Evidence `json:"evidence"`
	id       ast.Id
	array    *ast.ArrayLiteral
}
type Edit struct {
	Evidence    jsast.Evidence `json:"evidence"`
	Rule        string         `json:"rule"`
	Binding     string         `json:"binding"`
	Factory     string         `json:"factory"`
	Replacement string         `json:"replacement"`
}
type Report struct {
	SchemaVersion int            `json:"schemaVersion"`
	Sample        any            `json:"sample"`
	Assumptions   Options        `json:"assumptions"`
	Tables        []Table        `json:"tables"`
	Factories     []Factory      `json:"factories"`
	Strings       []StringRecord `json:"strings"`
	Edits         []Edit         `json:"edits"`
	Recovered     int            `json:"recovered"`
	Limitations   []string       `json:"limitations"`
}

type index struct {
	ast.NoopVisitor
	parent       map[ast.Expr]ast.Expr
	expressions  []ast.Expr
	declarations []*ast.VariableDeclarator
	declCount    map[ast.Id]int
	bindings     map[*ast.Identifier]bool
	uses         map[ast.Id][]*ast.Identifier
	with         map[*ast.VariableDeclarator]bool
	withDepth    int
	withCalls    map[*ast.CallExpression]bool
	owners       map[*ast.VariableDeclarator]*ast.FunctionLiteral
	owner        *ast.FunctionLiteral
	arrowDepth   int
	unsafe       map[*ast.VariableDeclarator]bool
	calls        []*ast.CallExpression
}

func makeIndex(source *jsast.Source) *index {
	v := &index{parent: map[ast.Expr]ast.Expr{}, declCount: map[ast.Id]int{}, bindings: map[*ast.Identifier]bool{}, uses: map[ast.Id][]*ast.Identifier{}, with: map[*ast.VariableDeclarator]bool{}, withCalls: map[*ast.CallExpression]bool{}, owners: map[*ast.VariableDeclarator]*ast.FunctionLiteral{}, unsafe: map[*ast.VariableDeclarator]bool{}}
	v.V = v
	source.AST.VisitWith(v)
	return v
}
func (v *index) VisitExpression(n *ast.Expression) {
	if n.Expr == nil {
		return
	}
	if len(v.expressions) > 0 {
		v.parent[n.Expr] = v.expressions[len(v.expressions)-1]
	}
	v.expressions = append(v.expressions, n.Expr)
	n.VisitChildrenWith(v)
	v.expressions = v.expressions[:len(v.expressions)-1]
}
func (v *index) VisitIdentifier(n *ast.Identifier) { v.uses[n.ToId()] = append(v.uses[n.ToId()], n) }
func (v *index) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	if id, ok := n.Target.Target.(*ast.Identifier); ok {
		v.bindings[id] = true
		v.declCount[id.ToId()]++
	}
	v.declarations = append(v.declarations, n)
	v.with[n] = v.withDepth > 0
	v.owners[n] = v.owner
	v.unsafe[n] = v.arrowDepth > 0
	n.VisitChildrenWith(v)
}
func (v *index) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	old := v.owner
	v.owner = n
	n.VisitChildrenWith(v)
	v.owner = old
}
func (v *index) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) {
	v.arrowDepth++
	n.VisitChildrenWith(v)
	v.arrowDepth--
}
func (v *index) VisitWithStatement(n *ast.WithStatement) {
	v.withDepth++
	n.VisitChildrenWith(v)
	v.withDepth--
}
func (v *index) VisitCallExpression(n *ast.CallExpression) {
	v.calls = append(v.calls, n)
	v.withCalls[n] = v.withDepth > 0
	n.VisitChildrenWith(v)
}
func (v *index) written(expr ast.Expr) bool {
	p := v.parent[expr]
	switch n := p.(type) {
	case *ast.AssignExpression:
		return n.Left.Expr == expr
	case *ast.UpdateExpression:
		return n.Operand.Expr == expr
	case *ast.UnaryExpression:
		return n.Operator == token.Delete
	}
	return false
}
func (v *index) bindingWritten(id ast.Id) bool {
	for _, ref := range v.uses[id] {
		if v.written(ref) {
			return true
		}
	}
	return false
}
func (v *index) stable(id ast.Id) bool {
	if v.declCount[id] != 1 || v.bindingWritten(id) {
		return false
	}
	for _, ref := range v.uses[id] {
		if v.bindings[ref] {
			continue
		}
		var value ast.Expr = ref
		for {
			seq, ok := v.parent[value].(*ast.SequenceExpression)
			if !ok || seq.Sequence[len(seq.Sequence)-1].Expr != value {
				break
			}
			value = seq
		}
		member, ok := v.parent[value].(*ast.MemberExpression)
		if !ok || member.Object.Expr != value {
			return false
		}
		// Access is permitted only through an array element, never an alias.
		if _, ok := member.Property.Prop.(*ast.ComputedProperty); !ok {
			return false
		}
		var access ast.Expr = member
		for {
			if v.written(access) {
				return false
			}
			outer, ok := v.parent[access].(*ast.MemberExpression)
			if !ok || outer.Object.Expr != access {
				break
			}
			access = outer
		}
	}
	return true
}

func tableArray(expr ast.Expr) *ast.ArrayLiteral {
	if call, ok := expr.(*ast.CallExpression); ok {
		if len(call.ArgumentList) != 0 {
			return nil
		}
		fn, ok := call.Callee.Expr.(*ast.FunctionLiteral)
		if !ok || fn.Async || fn.Generator || len(fn.ParameterList.List) != 0 || fn.ParameterList.Rest != nil || len(fn.Body.List) != 1 {
			return nil
		}
		ret, ok := fn.Body.List[0].Stmt.(*ast.ReturnStatement)
		if !ok || ret.Argument == nil {
			return nil
		}
		expr = ret.Argument.Expr
	}
	arr, ok := expr.(*ast.ArrayLiteral)
	if !ok || len(arr.Value) == 0 {
		return nil
	}
	for _, element := range arr.Value {
		if _, ok := element.Expr.(*ast.FunctionLiteral); !ok {
			return nil
		}
	}
	return arr
}

type candidateVisitor struct {
	ast.NoopVisitor
	root            *ast.FunctionLiteral
	long            []*ast.StringLiteral
	loops, closures int
	memberDepth     int
}

func (v *candidateVisitor) VisitStringLiteral(n *ast.StringLiteral) {
	if len(stringValue(n.Value)) >= 32 && v.memberDepth > 0 {
		v.long = append(v.long, n)
	}
}
func (v *candidateVisitor) VisitMemberExpression(n *ast.MemberExpression) {
	v.memberDepth++
	n.VisitChildrenWith(v)
	v.memberDepth--
}
func (v *candidateVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	if n != v.root {
		v.closures++
	}
	n.VisitChildrenWith(v)
}
func (v *candidateVisitor) VisitForStatement(n *ast.ForStatement) { v.loops++; n.VisitChildrenWith(v) }

type invariantVisitor struct {
	ast.NoopVisitor
	root   *ast.FunctionLiteral
	local  map[ast.Id]bool
	writes []*ast.AssignExpression
	bad    bool
}

func (v *invariantVisitor) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	if n != v.root {
		v.bad = true
		return
	}
	n.VisitChildrenWith(v)
}
func (v *invariantVisitor) VisitArrowFunctionLiteral(n *ast.ArrowFunctionLiteral) { v.bad = true }
func (v *invariantVisitor) VisitAssignExpression(n *ast.AssignExpression) {
	id, ok := n.Left.Expr.(*ast.Identifier)
	if !ok {
		v.bad = true
	} else if !v.local[id.ToId()] {
		v.writes = append(v.writes, n)
	}
	n.VisitChildrenWith(v)
}
func (v *invariantVisitor) VisitUpdateExpression(n *ast.UpdateExpression) {
	id, ok := n.Operand.Expr.(*ast.Identifier)
	if !ok || !v.local[id.ToId()] {
		v.bad = true
	}
	n.VisitChildrenWith(v)
}

func proveInvariant(c *closure) *Invariant {
	fn := c.node
	v := &invariantVisitor{root: fn, local: locals(fn).ids}
	v.V = v
	fn.VisitWith(v)
	if v.bad {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "nested closure, property write or captured update")
	}
	if len(v.writes) == 0 {
		return &Invariant{Kind: "pure-local-state"}
	}
	stmts := []ast.Stmt{}
	for _, s := range fn.Body.List {
		if e, ok := s.Stmt.(*ast.ExpressionStatement); ok {
			if _, ok := e.Expression.Expr.(*ast.StringLiteral); ok {
				continue
			}
		}
		stmts = append(stmts, s.Stmt)
	}
	if len(stmts) < 2 || len(v.writes) != 1 {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "cache writes")
	}
	first, ok := stmts[0].(*ast.IfStatement)
	if !ok || first.Alternate != nil {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "cache guard")
	}
	guard, ok := jsast.LastValue(first.Test.Expr).(*ast.Identifier)
	if !ok {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "guard identifier")
	}
	last, ok := stmts[len(stmts)-1].(*ast.ReturnStatement)
	if !ok || last.Argument == nil || last.Argument.Expr != v.writes[0] || v.writes[0].Operator != token.Assign {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "cache return assignment")
	}
	cache := v.writes[0].Left.Expr.(*ast.Identifier)
	if guard.ToId() != cache.ToId() {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "guard binding")
	}
	branch := first.Consequent.Stmt
	if block, ok := branch.(*ast.BlockStatement); ok {
		if len(block.List) != 1 {
			fail("STATEFUL_DECODER_UNSUPPORTED", fn, "guard body")
		}
		branch = block.List[0].Stmt
	}
	ret, ok := branch.(*ast.ReturnStatement)
	if !ok || ret.Argument == nil {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "guard return")
	}
	ref, ok := ret.Argument.Expr.(*ast.Identifier)
	if !ok || ref.ToId() != cache.ToId() {
		fail("STATEFUL_DECODER_UNSUPPORTED", fn, "guard result binding")
	}
	return &Invariant{Kind: "truthy-memoized-string", Cache: cache.Name, RequiresTruthyResult: true}
}

func hash(text string) string { sum := sha256.Sum256([]byte(text)); return hex.EncodeToString(sum[:]) }
func jsonPrimitive(v any) any {
	if s, ok := v.(jsString); ok {
		text, err := s.utf8()
		if err != nil {
			fail("UNPAIRED_STATIC_SURROGATE", nil, err.Error())
		}
		return text
	}
	if _, ok := v.(undefined); ok {
		return map[string]string{"type": "undefined"}
	}
	if number, ok := v.(float64); ok && (math.IsNaN(number) || math.IsInf(number, 0) || number == 0 && math.Signbit(number)) {
		text := numberText(number)
		if number == 0 {
			text = "-0"
		}
		return map[string]string{"type": "number", "value": text}
	}
	return v
}
func recoverOne(fn *ast.FunctionLiteral, args []any, options Options) (c *closure, invariant *Invariant, value jsString, err *Failure) {
	defer func() {
		if caught := recover(); caught != nil {
			if failure, ok := caught.(*Failure); ok {
				err = failure
			} else {
				panic(caught)
			}
		}
	}()
	r := newReducer(options.AssumeStandardIntrinsics, options.Fuel)
	result := r.invoke(&closure{node: fn, env: &environment{values: map[ast.Id]any{}}}, args)
	c, ok := result.(*closure)
	if !ok {
		fail("FACTORY_DID_NOT_RETURN_CLOSURE", fn, fmt.Sprintf("%T", result))
	}
	invariant = proveInvariant(c)
	result = r.invoke(c, nil)
	value, ok = result.(jsString)
	if !ok {
		fail("FACTORY_DID_NOT_DECODE_STRING", fn, fmt.Sprintf("%T", result))
	}
	if invariant.RequiresTruthyResult && len(value) == 0 {
		fail("EMPTY_CACHE_RESULT_UNPROVEN", fn, "")
	}
	return c, invariant, value, nil
}

// Recover derives tables, factories, captured constants and invariance proofs
// from this source. Unsupported decoder grammar is retained with a diagnostic.
func Recover(source *jsast.Source, options Options) (result *Report, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			if failure, ok := caught.(*Failure); ok {
				result = nil
				err = failure
			} else {
				panic(caught)
			}
		}
	}()
	if options.Fuel == 0 {
		options.Fuel = 2000000
	}
	if options.Fuel < 1 || options.Fuel > 10000000 {
		return nil, fmt.Errorf("INVALID_STATIC_BUDGET")
	}
	jsast.Bind(source.AST)
	v := makeIndex(source)
	report := &Report{SchemaVersion: 1, Sample: map[string]any{"filename": source.Filename, "bytes": len(source.Text), "sha256": source.SHA256}, Assumptions: options, Tables: []Table{}, Factories: []Factory{}, Strings: []StringRecord{}, Edits: []Edit{}, Limitations: []string{"Bounded script AST grammar; unsupported factories and mutable/escaped tables are not rewritten.", "Only proven zero-argument string calls are replaced; general constant folding and selector specialization remain separate reference passes.", "Cross-version validation requires additional original releases; synthetic variations are not version evidence."}}
	tables := map[ast.Id]*Table{}
	factories := map[*ast.FunctionLiteral]*Factory{}
	for _, d := range v.declarations {
		id, ok := d.Target.Target.(*ast.Identifier)
		if !ok || d.Initializer == nil || v.with[d] {
			continue
		}
		arr := tableArray(d.Initializer.Expr)
		if arr == nil {
			continue
		}
		ev, err := source.Evidence(int(id.Idx)-1, int(arr.RightBracket))
		if err != nil {
			return nil, err
		}
		table := Table{Name: id.Name, Stable: v.stable(id.ToId()), Members: len(arr.Value), Evidence: ev, id: id.ToId(), array: arr}
		report.Tables = append(report.Tables, table)
	}
	for i := range report.Tables {
		table := &report.Tables[i]
		tables[table.id] = table
		for j, el := range table.array.Value {
			fn := el.Expr.(*ast.FunctionLiteral)
			scan := &candidateVisitor{root: fn}
			scan.V = scan
			fn.VisitWith(scan)
			if len(scan.long) == 0 || scan.loops == 0 || scan.closures == 0 {
				continue
			}
			ev, err := source.Evidence(int(fn.Function)-1, int(fn.Body.RightBrace))
			if err != nil {
				return nil, err
			}
			factory := Factory{Name: fmt.Sprintf("%s[%d]", table.Name, j), Evidence: ev, Pools: []Pool{}, fn: fn}
			for _, literal := range scan.long {
				poolEvidence, err := source.Evidence(int(literal.Idx0())-1, int(literal.Idx1())-1)
				if err != nil {
					return nil, err
				}
				factory.Pools = append(factory.Pools, Pool{len(stringValue(literal.Value)), hash(literal.Value), poolEvidence})
			}
			report.Factories = append(report.Factories, factory)
		}
	}
	for i := range report.Factories {
		factory := &report.Factories[i]
		factories[factory.fn] = factory
	}
	for _, d := range v.declarations {
		id, ok := d.Target.Target.(*ast.Identifier)
		if !ok || d.Initializer == nil {
			continue
		}
		call, ok := jsast.LastValue(d.Initializer.Expr).(*ast.CallExpression)
		if !ok {
			continue
		}
		member, ok := call.Callee.Expr.(*ast.MemberExpression)
		if !ok {
			continue
		}
		receiver, ok := jsast.LastValue(member.Object.Expr).(*ast.Identifier)
		if !ok {
			continue
		}
		table := tables[receiver.ToId()]
		if table == nil {
			continue
		}
		property, ok := member.Property.Prop.(*ast.ComputedProperty)
		if !ok {
			continue
		}
		key, ok := constant(property.Expr.Expr)
		if !ok {
			continue
		}
		n, ok := key.(float64)
		if !ok || n < 0 || n >= float64(table.Members) || n != float64(int(n)) {
			continue
		}
		fn := table.array.Value[int(n)].Expr.(*ast.FunctionLiteral)
		factory := factories[fn]
		if factory == nil {
			continue
		}
		evidence, err := source.Evidence(int(id.Idx)-1, int(call.RightParenthesis))
		if err != nil {
			return nil, err
		}
		record := StringRecord{Binding: id.Name, Factory: factory.Name, Evidence: evidence, Arguments: []any{}, Status: "unresolved", id: id.ToId()}
		args := []any{}
		for _, arg := range call.ArgumentList {
			value, known := constant(arg.Expr)
			if !known {
				record.Reason = "NON_CONSTANT_FACTORY_ARGUMENT"
				break
			}
			args = append(args, value)
			record.Arguments = append(record.Arguments, jsonPrimitive(value))
		}
		if v.with[d] {
			record.Reason = "DYNAMIC_WITH_SCOPE_UNSUPPORTED"
		}
		if v.unsafe[d] {
			record.Reason = "ARROW_FACTORY_SCOPE_UNSUPPORTED"
		}
		if owner := v.owners[d]; owner != nil {
			if owner.ParameterList.Rest != nil {
				record.Reason = "COMPLEX_FACTORY_SCOPE_UNSUPPORTED"
			}
			for _, p := range owner.ParameterList.List {
				if _, ok := p.Target.Target.(*ast.Identifier); !ok || p.Initializer != nil {
					record.Reason = "COMPLEX_FACTORY_SCOPE_UNSUPPORTED"
				}
			}
		}
		if !table.Stable {
			record.Reason = "MUTABLE_OR_ESCAPED_FACTORY_TABLE"
		}
		if id.ScopeContext == 0 || v.declCount[id.ToId()] != 1 || v.bindingWritten(id.ToId()) {
			record.Reason = "MUTABLE_DECODER_BINDING"
		}
		if !options.AssumeStandardIntrinsics {
			record.Reason = "STANDARD_INTRINSICS_NOT_ASSUMED"
		}
		if record.Reason == "" {
			_, invariant, units, failure := recoverOne(fn, args, options)
			if failure != nil {
				record.Reason = failure.Code
				record.Detail = failure.Detail
			} else {
				value, err := units.utf8()
				if err != nil {
					record.Reason = "UNPAIRED_STATIC_SURROGATE"
					record.Detail = err.Error()
				} else {
					record.Status = "recovered"
					record.Value = &value
					record.ValueSHA256 = hash(value)
					record.Invariant = invariant
					record.units = units
					report.Recovered++
				}
			}
		}
		report.Strings = append(report.Strings, record)
	}
	if options.AssumeInitializedDecoders {
		bindings := map[ast.Id]*StringRecord{}
		for i := range report.Strings {
			record := &report.Strings[i]
			if record.Status == "recovered" {
				bindings[record.id] = record
			}
		}
		for _, call := range v.calls {
			if len(call.ArgumentList) != 0 || v.withCalls[call] {
				continue
			}
			id, ok := call.Callee.Expr.(*ast.Identifier)
			if !ok {
				continue
			}
			record := bindings[id.ToId()]
			if record == nil {
				continue
			}
			ev, err := source.Evidence(int(id.Idx)-1, int(call.RightParenthesis))
			if err != nil {
				return nil, err
			}
			report.Edits = append(report.Edits, Edit{ev, "decode-string-call", record.Binding, record.Factory, "(" + quoteUnits(record.units) + ")"})
		}
	}
	sort.Slice(report.Edits, func(i, j int) bool { return report.Edits[i].Evidence.UTF8[0] < report.Edits[j].Evidence.UTF8[0] })
	return report, nil
}

func Rewrite(source *jsast.Source, report *Report) (string, error) {
	text := source.Text
	previous := len(text)
	for i := len(report.Edits) - 1; i >= 0; i-- {
		edit := report.Edits[i]
		start, end := edit.Evidence.UTF8[0], edit.Evidence.UTF8[1]
		if start < 0 || end < start || end > previous {
			return "", fmt.Errorf("OVERLAPPING_OR_INVALID_DEOBFUSCATOR_EDIT: %d..%d", start, end)
		}
		text = text[:start] + edit.Replacement + text[end:]
		previous = start
	}
	if _, err := jsast.Parse("deobfuscated.js", []byte(text)); err != nil {
		return "", fmt.Errorf("INVALID_DEOBFUSCATOR_OUTPUT: %w", err)
	}
	return text, nil
}
