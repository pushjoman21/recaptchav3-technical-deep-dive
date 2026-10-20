package vm02

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

func authoredConfiguration(semantic string, opcode int) Configuration {
	return Configuration{Sample: Sample{SHA256: SyntheticProfileSHA256}, Entries: []Handler{{Opcode: opcode, Semantic: semantic, Layout: semanticLayouts()[semantic]}}}
}

func TestSignedVarintCursorAndOverflow(t *testing.T) {
	for _, test := range []struct {
		name   string
		bytes  []byte
		value  int
		code   string
		cursor int
	}{
		{"zero", []byte{0}, 0, "", 1}, {"128", []byte{128, 1}, 128, "", 2},
		{"negative", []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 1}, -1, "", 10},
		{"truncated", []byte{128}, 0, "TRUNCATED_INPUT", 1},
		{"overflow", []byte{128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 0}, 0, "VARINT_OVERFLOW", 10},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader := reader{bytes: test.bytes}
			value, err := reader.int32()
			if reader.position != test.cursor {
				t.Fatal(reader.position)
			}
			if test.code == "" {
				if err != nil || value != test.value {
					t.Fatal(value, err)
				}
			} else if diagnostic, ok := err.(*Diagnostic); !ok || diagnostic.Code != test.code {
				t.Fatal(err)
			}
		})
	}
}

func TestBooleanUsesAllGroups(t *testing.T) {
	reader := reader{bytes: []byte{128, 128, 128, 128, 128, 128, 128, 128, 128, 1}}
	value, err := reader.boolean()
	if err != nil || !value || reader.position != 10 {
		t.Fatal(value, err, reader.position)
	}
}

func TestFloatAndUTF8Atoms(t *testing.T) {
	for _, value := range []float64{math.Copysign(0, -1), math.NaN(), math.Inf(1), math.Inf(-1), 1.25} {
		data := []byte{26, 9, 49, 0, 0, 0, 0, 0, 0, 0, 0}
		binary.LittleEndian.PutUint64(data[3:], math.Float64bits(value))
		reader := reader{bytes: data}
		atom, err := reader.atom()
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case math.IsNaN(value):
			if atom.Value != "NaN" {
				t.Fatal(atom.Value)
			}
		case math.IsInf(value, 1):
			if atom.Value != "Infinity" {
				t.Fatal(atom.Value)
			}
		case math.IsInf(value, -1):
			if atom.Value != "-Infinity" {
				t.Fatal(atom.Value)
			}
		case value == 0 && math.Signbit(value):
			if atom.Value != "-0" {
				t.Fatal(atom.Value)
			}
		default:
			if atom.Value != value {
				t.Fatal(atom.Value)
			}
		}
	}
	reader := reader{bytes: []byte{26, 5, 34, 3, 239, 187, 191}}
	atom, err := reader.atom()
	if err != nil || atom.Value != "\ufeff" {
		t.Fatal(atom, err)
	}
	reader.bytes = []byte{26, 3, 34, 1, 255}
	reader.position = 0
	if _, err := reader.atom(); err == nil {
		t.Fatal("invalid UTF8 accepted")
	}
}

func TestRawNumericIndexAndDynamicOpcodeAssignment(t *testing.T) {
	configuration := authoredConfiguration("MAKE_CLOSURE", 77)
	bytes := []byte{2, 77, 16, 3, 26, 5, 45, 0, 0, 0, 0, 26, 2, 24, 7}
	program, err := Decode(bytes, configuration)
	if err != nil || program.Status != "decoded" {
		t.Fatal(program, err)
	}
	instruction := program.Instructions[0]
	if *instruction.Operands[1].IndexValue != 7 || instruction.Operands[1].ReadMode != "raw-int32-index" || *instruction.Target != len(bytes) {
		t.Fatal(instruction)
	}
	configuration.Entries[0].Opcode = 78
	program, err = Decode(bytes, configuration)
	if err != nil || program.Status != "partial" || program.UnknownRegions[0].Diagnostic.Code != "UNKNOWN_OPCODE" {
		t.Fatal(program, err)
	}
}

func TestUnknownAndMalformedRemainExplicit(t *testing.T) {
	configuration := authoredConfiguration("MOVE", 1)
	for _, test := range []struct {
		bytes []byte
		code  string
	}{
		{[]byte{0, 99}, "UNKNOWN_OPCODE"}, {[]byte{1}, "TRUNCATED_INPUT"},
		{[]byte{0, 1}, "INVALID_OPERAND_COUNT"}, {[]byte{1, 1, 16, 0, 26, 0}, "INVALID_ATOM_LENGTH"},
		{[]byte{1, 1, 16, 0, 26, 2, 99, 0}, "UNSUPPORTED_ATOM_TYPE"},
	} {
		program, err := Decode(test.bytes, configuration)
		if err != nil {
			t.Fatal(err)
		}
		if program.Status != "partial" || len(program.Instructions) != 0 || program.UnknownRegions[0].Diagnostic.Code != test.code {
			t.Fatal(program)
		}
	}
}

func TestProfileAndLayoutFailClosed(t *testing.T) {
	configuration := authoredConfiguration("MOVE", 1)
	configuration.Sample.SHA256 = "unknown"
	if _, err := Decode(nil, configuration); err == nil {
		t.Fatal("unknown profile accepted")
	}
	configuration = authoredConfiguration("MOVE", 1)
	configuration.Entries[0].Layout.Min = 0
	if _, err := Decode(nil, configuration); err == nil {
		t.Fatal("unknown layout accepted")
	}
	configuration = authoredConfiguration("BINARY_MAP_LEFT", 2)
	configuration.Entries[0].Operation = &Operation{Kind: "argument", Index: pointer(0)}
	if _, err := Decode(nil, configuration); err == nil {
		t.Fatal("non-binary map accepted")
	}
}

func TestAssemblySeparatesLongStringsAndChecksAnnotationIdentity(t *testing.T) {
	text := strings.Repeat("x", 60)
	bytes := append([]byte{1, 1, 16, 3, 26, 62, 34, 60}, []byte(text)...)
	program, err := Decode(bytes, authoredConfiguration("MOVE", 1))
	if err != nil || program.Status != "decoded" {
		t.Fatal(program, err)
	}
	assembly, err := Assembly(program, nil)
	if err != nil || !strings.Contains(assembly, "R3, C_000") || !strings.Contains(assembly, ".const C_000 = \""+text+"\"") || strings.Contains(assembly, "constants.json") {
		t.Fatal(assembly, err)
	}
	if len(Constants(program)) != 1 || Constants(program)[0].Value != text {
		t.Fatal(Constants(program))
	}
	if _, err := Assembly(program, &StringCatalogue{Sample: program.Sample, ProgramBytecode: "wrong"}); err == nil {
		t.Fatal("mismatched strings accepted")
	}
}
