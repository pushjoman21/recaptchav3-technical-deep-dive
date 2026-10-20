package vm01

import (
	"strings"
	"testing"
)

func TestAssemblyKeepsRegistrationEvidenceAndMissingConsumer(t *testing.T) {
	input := []byte(`{"qualified":true,"sample":{"sha256":"sample"},"handlers":[{"ids":[-53,-54],"semantic":"MUL","purpose":"Pop operands; push their product.","dependencies":["stack"],"handler":{"member":"T[2]","selector":17,"evidence":{"start":{"line":123}}}}]}`)
	text, err := Assembly(input, "sample")
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{".unresolved CONSUMING_INTERPRETER_AND_PROGRAM", "H_000:  MUL", "ids -53, -54", ".source T[2](17) @123", ".requires stack"} {
		if !strings.Contains(text, part) {
			t.Fatal("missing", part, text)
		}
	}
	if _, err := Assembly(input, "other"); err == nil {
		t.Fatal("foreign source accepted")
	}
	if _, err := Assembly([]byte(`{"qualified":false,"sample":{"sha256":"sample"}}`), "sample"); err == nil {
		t.Fatal("unqualified catalogue accepted")
	}
}
