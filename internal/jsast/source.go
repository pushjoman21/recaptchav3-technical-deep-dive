// Package jsast parses source as data. It has no JavaScript evaluator.
package jsast

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"unicode/utf8"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/parser"
	"github.com/t14raptor/go-fast/resolver"
)

const MaxSourceBytes = 32 * 1024 * 1024

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Evidence struct {
	UTF8  [2]int   `json:"utf8"`
	UTF16 [2]int   `json:"utf16"`
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Source struct {
	Filename string
	Text     string
	SHA256   string
	AST      *ast.Program
	lines    []int
	units    []int
}

func Read(filename string) (*Source, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxSourceBytes {
		return nil, fmt.Errorf("SOURCE_TOO_LARGE: %d", info.Size())
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxSourceBytes+1))
	if err != nil {
		return nil, err
	}
	return Parse(filename, data)
}

func Parse(filename string, data []byte) (source *Source, err error) {
	if len(data) > MaxSourceBytes {
		return nil, fmt.Errorf("SOURCE_TOO_LARGE: %d", len(data))
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("INVALID_SOURCE_UTF8")
	}
	// Parser failures must be diagnostics, never a partially accepted AST.
	defer func() {
		if caught := recover(); caught != nil {
			source = nil
			err = fmt.Errorf("JS_PARSER_FAILURE: %v", caught)
		}
	}()
	program, err := parser.ParseFile(string(data))
	if err != nil {
		return nil, fmt.Errorf("JS_PARSE_FAILED: %w", err)
	}
	resolver.Resolve(program)
	sum := sha256.Sum256(data)
	source = &Source{Filename: filename, Text: string(data), SHA256: hex.EncodeToString(sum[:]), AST: program, lines: []int{0}, units: make([]int, len(data)+1)}
	units := 0
	for offset, value := range source.Text {
		size := utf8.RuneLen(value)
		for i := 0; i < size; i++ {
			source.units[offset+i] = units
		}
		units++
		if value > 0xffff {
			units++
		}
		if value == '\n' {
			source.lines = append(source.lines, offset+size)
		}
		source.units[offset+size] = units
	}
	return source, nil
}

func (s *Source) Evidence(start, end int) (Evidence, error) {
	if start < 0 || end < start || end > len(s.Text) {
		return Evidence{}, fmt.Errorf("INVALID_SOURCE_RANGE: %d..%d", start, end)
	}
	position := func(offset int) Position {
		line := sort.Search(len(s.lines), func(i int) bool { return s.lines[i] > offset }) - 1
		return Position{Line: line + 1, Column: s.units[offset] - s.units[s.lines[line]] + 1}
	}
	return Evidence{UTF8: [2]int{start, end}, UTF16: [2]int{s.units[start], s.units[end]}, Start: position(start), End: position(end)}, nil
}

// LastValue is for data-flow/display inspection, never a source rewrite.
func LastValue(expression ast.Expr) ast.Expr {
	for {
		sequence, ok := expression.(*ast.SequenceExpression)
		if !ok || len(sequence.Sequence) == 0 {
			return expression
		}
		expression = sequence.Sequence[len(sequence.Sequence)-1].Expr
	}
}

func Member(expression ast.Expr) (ast.Expr, string, bool) {
	member, ok := LastValue(expression).(*ast.MemberExpression)
	if !ok || member.Object == nil || member.Property == nil {
		return nil, "", false
	}
	switch property := member.Property.Prop.(type) {
	case *ast.Identifier:
		return LastValue(member.Object.Expr), property.Name, true
	case *ast.ComputedProperty:
		if property.Expr != nil {
			switch key := LastValue(property.Expr.Expr).(type) {
			case *ast.StringLiteral:
				return LastValue(member.Object.Expr), key.Value, true
			case *ast.NumberLiteral:
				return LastValue(member.Object.Expr), fmt.Sprintf("%g", key.Value), true
			}
		}
	}
	return nil, "", false
}

func Name(expression ast.Expr) string {
	if expression == nil {
		return "<absent>"
	}
	switch value := LastValue(expression).(type) {
	case *ast.Identifier:
		return value.Name
	case *ast.ThisExpression:
		return "this"
	case *ast.MemberExpression:
		if object, key, ok := Member(value); ok {
			return Name(object) + "[" + fmt.Sprintf("%q", key) + "]"
		}
		return Name(value.Object.Expr) + "[dynamic]"
	case *ast.CallExpression:
		return Name(value.Callee.Expr) + "(…)"
	default:
		return fmt.Sprintf("<%T>", value)
	}
}

type Function struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Evidence     Evidence             `json:"evidence"`
	SourceSHA256 string               `json:"sourceSha256"`
	Parameters   int                  `json:"parameters"`
	Node         *ast.FunctionLiteral `json:"-"`
}

type functionVisitor struct {
	ast.NoopVisitor
	source    *Source
	functions []Function
	names     map[*ast.FunctionLiteral]string
	err       error
}

func (v *functionVisitor) VisitVariableDeclarator(node *ast.VariableDeclarator) {
	if node.Initializer != nil {
		if fn, ok := LastValue(node.Initializer.Expr).(*ast.FunctionLiteral); ok {
			if name, ok := node.Target.Target.(*ast.Identifier); ok {
				v.names[fn] = name.Name
			}
		}
	}
	node.VisitChildrenWith(v)
}

func (v *functionVisitor) VisitAssignExpression(node *ast.AssignExpression) {
	if node.Right != nil {
		if fn, ok := LastValue(node.Right.Expr).(*ast.FunctionLiteral); ok {
			v.names[fn] = Name(node.Left.Expr)
		}
	}
	node.VisitChildrenWith(v)
}

func (v *functionVisitor) VisitFunctionLiteral(node *ast.FunctionLiteral) {
	start, end := int(node.Function)-1, int(node.Body.RightBrace)
	evidence, err := v.source.Evidence(start, end)
	if err != nil {
		v.err = err
		return
	}
	name := v.names[node]
	if node.Name != nil {
		name = node.Name.Name
	}
	if name == "" {
		name = "<anonymous>"
	}
	sum := sha256.Sum256([]byte(v.source.Text[start:end]))
	v.functions = append(v.functions, Function{ID: fmt.Sprintf("F_%06x", start), Name: name, Evidence: evidence, SourceSHA256: hex.EncodeToString(sum[:]), Parameters: len(node.ParameterList.List), Node: node})
	node.VisitChildrenWith(v)
}

func Functions(source *Source) ([]Function, error) {
	visitor := &functionVisitor{source: source, names: make(map[*ast.FunctionLiteral]string)}
	visitor.V = visitor
	source.AST.VisitWith(visitor)
	sort.Slice(visitor.functions, func(i, j int) bool {
		return visitor.functions[i].Evidence.UTF8[0] < visitor.functions[j].Evidence.UTF8[0]
	})
	return visitor.functions, visitor.err
}
