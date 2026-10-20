package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicDecodeRequiresSyntheticIdentityAndWritesCompleteArtifacts(t *testing.T) {
	configuration := "../../examples/register/configuration.json"
	input := "../../examples/register/program.bin"
	out := filepath.Join(t.TempDir(), "decoded")
	args := []string{"--configuration", configuration, "--input", input, "--out", out}
	if err := decode(args); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"vm_02_output.asm", "program.json", "constants.json", "bytecode.bin"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatal(name, err)
		}
	}
	if err := decode(args); err == nil {
		t.Fatal("existing output overwritten")
	}
	data, err := os.ReadFile(configuration)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	config["sample"].(map[string]any)["sha256"] = "unreleased-production-profile"
	data, err = json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(t.TempDir(), "foreign.json")
	if err := os.WriteFile(foreign, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := decode([]string{"--configuration", foreign, "--input", input, "--out", filepath.Join(t.TempDir(), "rejected")}); err == nil || !strings.Contains(err.Error(), "UNSUPPORTED_WIRE_PROFILE") {
		t.Fatal(err)
	}
}
