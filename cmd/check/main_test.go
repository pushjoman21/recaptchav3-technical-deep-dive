package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectionRejectsPrivateDataAndOutsideLinks(t *testing.T) {
	for _, content := range []string{"raw_payload.json", "private/tasks/example", "[outside](../../hidden.md)", "[local](missing.md)"} {
		if err := inspect(t.TempDir(), "docs/example.md", []byte(content)); err == nil {
			t.Fatal("accepted", content)
		}
	}
	credential := "ghp_" + strings.Repeat("a", 32)
	if err := inspect(t.TempDir(), "example.md", []byte(credential)); err == nil {
		t.Fatal("credential shape accepted")
	}
	workstation := string([]byte{'C', ':', '\\'}) + "work"
	if err := inspect(t.TempDir(), "example.md", []byte(workstation)); err == nil {
		t.Fatal("workstation path accepted")
	}
}

func TestForbiddenDirectoriesAreRejectedRatherThanIgnored(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "captures"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := files(root); err == nil {
		t.Fatal("capture directory ignored")
	}
}

func TestDocumentLinksCanResolveWithinThePublication(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := inspect(root, "docs/example.md", []byte("[read](../README.md)")); err != nil {
		t.Fatal(err)
	}
}

func TestFullResearchListingsRetainRecordsAndRejectTruncation(t *testing.T) {
	for _, name := range []string{"vm_01_output.asm", "vm_02_output.asm"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "output", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateListing(name, data); err != nil {
			t.Fatal(err)
		}
		if err := validateListing(name, data[:len(data)/2]); err == nil {
			t.Fatal("truncated listing accepted", name)
		}
		if err := validateListing(name, append(append([]byte{}, data...), 0xc2, 0x85)); err == nil {
			t.Fatal("unreviewed Unicode separator accepted", name)
		}
	}
}

func TestPublicationRejectsDecompilerImplementation(t *testing.T) {
	if err := inspect(t.TempDir(), "vms/vm_02/decompiler.go", []byte("package vm02\n")); err == nil {
		t.Fatal("withheld decompiler accepted")
	}
}
