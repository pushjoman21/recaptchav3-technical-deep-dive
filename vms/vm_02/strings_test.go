package vm02

import "testing"

func stringFixture(encoded, alphabet string, includeEncoded bool) (*Program, *Program) {
	anchor := &Program{Status: "decoded", Sample: Sample{SHA256: SyntheticProfileSHA256}}
	anchor.Instructions = []Instruction{{Semantic: "MOVE", Destination: pointer(2), Operands: []Atom{{Kind: "string", Value: alphabet}}}}
	if includeEncoded {
		anchor.Instructions = append(anchor.Instructions, Instruction{Semantic: "MOVE", Destination: pointer(1), Operands: []Atom{{Kind: "string", Value: encoded}}})
	}
	program := &Program{Status: "decoded", Sample: anchor.Sample, Instructions: []Instruction{
		{Address: 0, End: 10, Semantic: "CHAR_CODES", Destination: pointer(9), Operands: []Atom{{Kind: "register", Index: pointer(1)}}},
		{Address: 10, End: 20, Semantic: "BINARY_MAP_LEFT", Destination: pointer(9), Operands: []Atom{{Kind: "register", Index: pointer(9)}, {Kind: "number", Value: 3}}, Operation: &Operation{Kind: "binary", Operator: "%", Left: &Operation{Kind: "argument", Index: pointer(0)}, Right: &Operation{Kind: "argument", Index: pointer(1)}}},
		{Address: 20, End: 30, Semantic: "CONCAT_INDEXED", Destination: pointer(9), Operands: []Atom{{Kind: "register", Index: pointer(2)}, {Kind: "register", Index: pointer(9)}}},
	}}
	return anchor, program
}

func TestStringReductionUsesUTF16CodeUnits(t *testing.T) {
	anchor, program := stringFixture("😀", "abc", true)
	catalogue, err := DecodeStrings(anchor, program)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogue.Annotations) != 1 || catalogue.Annotations[0].Value != "ba" || len(catalogue.UnknownLoads) != 0 {
		t.Fatal(catalogue)
	}
}

func TestStringReductionPreservesUndefinedPropertyConcatenation(t *testing.T) {
	anchor, program := stringFixture("A", "x", true)
	catalogue, err := DecodeStrings(anchor, program)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogue.Annotations) != 1 || catalogue.Annotations[0].Value != "undefined" {
		t.Fatal(catalogue)
	}
}

func TestMissingStringContextAndMidSequenceTargetsStayExplicit(t *testing.T) {
	anchor, program := stringFixture("A", "abc", false)
	catalogue, err := DecodeStrings(anchor, program)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogue.UnknownLoads) != 1 || catalogue.UnknownLoads[0].Diagnostic != "ENCODED_REGISTER_UNAVAILABLE" || len(catalogue.Annotations) != 0 {
		t.Fatal(catalogue)
	}
	anchor, program = stringFixture("A", "abc", true)
	program.Instructions[0].Target = pointer(10)
	catalogue, err = DecodeStrings(anchor, program)
	if err != nil || len(catalogue.Annotations) != 0 {
		t.Fatal(catalogue, err)
	}
}

func TestUnpairedDecodedUTF16IsNotReplacedSilently(t *testing.T) {
	anchor, program := stringFixture("B", "😀", true)
	catalogue, err := DecodeStrings(anchor, program)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogue.UnknownLoads) != 1 || catalogue.UnknownLoads[0].Diagnostic != "UNPAIRED_UTF16_IN_DECODED_STRING" {
		t.Fatal(catalogue)
	}
}
