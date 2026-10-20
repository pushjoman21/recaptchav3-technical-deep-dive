package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const factoryFixture = `var T=[function(a,c,k,x){return function(i,o){if(x)return x;for(o="",i=0;i<c;i++)o+=String.fromCharCode("abcdefghijklmnopqrstuvwxyz0123456789".charCodeAt(a+i)^k);return x=o;};}];var f=T[0](1,2,0);f();`

func TestDeobfuscateCommandProducesVerifiedDataWithoutOverwriting(t *testing.T) {
	input := filepath.Join(t.TempDir(), "source.js")
	if err := os.WriteFile(input, []byte(factoryFixture), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "deobfuscated")
	args := []string{"--input", input, "--out", out, "--assume-standard-intrinsics", "--assume-initialized-decoders", "--strict"}
	if err := deobfuscateSource(args); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "deobfuscated.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), `("bc");`) {
		t.Fatal(string(data))
	}
	report, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Recovered int   `json:"recovered"`
		Edits     []any `json:"edits"`
	}
	if err := json.Unmarshal(report, &decoded); err != nil || decoded.Recovered != 1 || len(decoded.Edits) != 1 {
		t.Fatal(decoded, err)
	}
	if err := deobfuscateSource(args); err == nil {
		t.Fatal("existing output overwritten")
	}
	original, err := os.ReadFile(input)
	if err != nil || string(original) != factoryFixture {
		t.Fatal("source changed", err)
	}
}

func TestStrictDeobfuscatorRejectsUnsupportedSourceWithDiagnostic(t *testing.T) {
	input := filepath.Join(t.TempDir(), "source.js")
	if err := os.WriteFile(input, []byte("var ordinary=1;"), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "diagnostic")
	err := deobfuscateSource([]string{"--input", input, "--out", out, "--strict"})
	if err == nil || !strings.Contains(err.Error(), "NO_SUPPORTED_STRING_FACTORY") {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "report.json")); err != nil {
		t.Fatal("diagnostic report missing", err)
	}
}
