package vm02

import (
	"math"
	"strings"
	"testing"
)

func assemblyProgram(instructions ...Instruction) *Program {
	program := &Program{Sample: Sample{SHA256: "source-hash"}, Instructions: instructions}
	program.Bytecode.Bytes = 64
	program.Bytecode.SHA256 = "bytecode-hash"
	return program
}

func atom(kind string, value any) Atom {
	return Atom{Kind: kind, Value: value}
}

func TestAssemblyUsesCompactSemanticAliasesAndFullConstants(t *testing.T) {
	long := strings.Repeat("q", 64)
	program := assemblyProgram(
		Instruction{Address: 0, End: 4, Opcode: 17, Semantic: "MOVE", Destination: pointer(278), Operands: []Atom{atom("string", long)}},
		Instruction{Address: 4, End: 8, Opcode: 201, Semantic: "MOVE_REGISTER", Destination: pointer(3), Operands: []Atom{{Kind: "register", Index: pointer(278)}}},
	)
	assembly, err := Assembly(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(assembly, "0x000000:  LOAD_CONST       R278, C_000") || !strings.Contains(assembly, "0x000004:  MOV              R3, R278") {
		t.Fatal(assembly)
	}
	if !strings.Contains(assembly, ".const C_000 = \""+long+"\"") || strings.Contains(assembly, "...") {
		t.Fatal("full constant missing or abbreviated", assembly)
	}
	if strings.Contains(assembly, "vm_02") || strings.Contains(assembly, "string sha256") {
		t.Fatal("unexpected assembly header", assembly)
	}
	aliasA, commentA := assemblyMnemonic(Instruction{Opcode: 17, Semantic: "MOVE", Operands: []Atom{atom("number", 1)}})
	aliasB, commentB := assemblyMnemonic(Instruction{Opcode: 201, Semantic: "MOVE", Operands: []Atom{atom("number", 1)}})
	if aliasA != aliasB || commentA != commentB {
		t.Fatal("display alias depends on opcode number", aliasA, aliasB)
	}
}

func TestAssemblyBinaryAliasesRequireCanonicalOperationIR(t *testing.T) {
	arg0 := &Operation{Kind: "argument", Index: pointer(0)}
	arg1 := &Operation{Kind: "argument", Index: pointer(1)}
	doubleNot := func(index int) *Operation {
		return &Operation{Kind: "not", Value: &Operation{Kind: "not", Value: &Operation{Kind: "argument", Index: pointer(index)}}}
	}
	program := assemblyProgram(
		Instruction{Address: 0, Semantic: "BINARY_MAP_LEFT", Operation: &Operation{Kind: "binary", Operator: "||", Left: arg0, Right: arg1}},
		Instruction{Address: 1, Semantic: "BINARY_MAP_LEFT", Operation: &Operation{Kind: "binary", Operator: "&&", Left: doubleNot(0), Right: doubleNot(1)}},
		Instruction{Address: 2, Semantic: "BINARY_MAP_LEFT", Operation: &Operation{Kind: "binary", Operator: "+", Left: &Operation{Kind: "not", Value: arg0}, Right: arg1}},
	)
	assembly, err := Assembly(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"OR", "BOOL_AND", "BINARY_MAP_LEFT"} {
		if !strings.Contains(assembly, expected) {
			t.Fatal("missing", expected, assembly)
		}
	}
	if !strings.Contains(assembly, `; operation={"kind":"binary","operator":"+"`) || !strings.Contains(assembly, `"kind":"not"`) {
		t.Fatal("noncanonical operation IR was not retained", assembly)
	}
}

func TestAssemblyBranchAliasesAndGuardedJump(t *testing.T) {
	target := 40
	program := assemblyProgram(
		Instruction{Address: 0, Semantic: "BRANCH_EQ", Target: &target, Operands: []Atom{atom("number", int32(4)), atom("number", 5), atom("number", 5)}},
		Instruction{Address: 10, Semantic: "BRANCH_EQ", Target: &target, Operands: []Atom{atom("number", int32(4)), atom("number", 5), atom("number", 6)}},
		Instruction{Address: 20, Semantic: "BRANCH_EQ", Target: &target, Operands: []Atom{atom("number", int32(4)), atom("number", "NaN"), atom("number", "NaN")}},
		Instruction{Address: 25, Semantic: "BRANCH_EQ", Target: &target, Operands: []Atom{atom("number", int32(4)), {Kind: "register", Index: pointer(2)}, {Kind: "register", Index: pointer(2)}}},
		Instruction{Address: 30, Semantic: "BRANCH_LT", Target: &target, Operands: []Atom{atom("number", int32(4)), {Kind: "register", Index: pointer(2)}, atom("number", 6)}},
		Instruction{Address: 35, Semantic: "BRANCH_LT", Target: &target, Operands: []Atom{atom("number", int32(4)), atom("number", 5), atom("number", 5)}},
	)
	assembly, err := Assembly(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(assembly, "JMP              L_000028") || !strings.Contains(assembly, "JE               5, 6, L_000028") || strings.Contains(assembly, "JE               L_000028, NaN, NaN") || !strings.Contains(assembly, "JE               R2, R2, L_000028") || !strings.Contains(assembly, "JL               R2, 6, L_000028") {
		t.Fatal(assembly)
	}
	if sameFiniteNumericLiteral(float64(math.NaN()), float64(math.NaN())) {
		t.Fatal("NaN must not collapse to JMP")
	}
	if !strings.Contains(assembly, "0x000023:  JL               5, 5, L_000028") {
		t.Fatal("false less-than branch collapsed into jump", assembly)
	}
}
