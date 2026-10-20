package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOutputProtectionAndNoOverwrite(t *testing.T) {
	root := t.TempDir()
	inputs := filepath.Join(root, "inputs")
	if err := os.Mkdir(inputs, 0700); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(inputs, "sample.js")
	if err := os.WriteFile(input, []byte("authored fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDirectory(filepath.Join(inputs, "output"), []string{input}); err == nil {
		t.Fatal("output created inside input tree")
	}
	if _, err := os.Stat(filepath.Join(inputs, "output")); !os.IsNotExist(err) {
		t.Fatal("failed validation mutated input tree", err)
	}
	output := filepath.Join(root, "reports", "new")
	if _, err := CreateDirectory(output, []string{input}); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDirectory(output, []string{input}); err == nil {
		t.Fatal("existing output overwritten")
	}
}
