// check validates the publication tree without opening network connections.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type entry struct {
	Path   string `json:"path"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type manifest struct {
	Schema int     `json:"schema"`
	Scope  string  `json:"scope"`
	Files  []entry `json:"files"`
}

var links = regexp.MustCompile(`\]\(([^)]+)\)|(?:src|srcset)="([^"]+)"`)
var drivePath = regexp.MustCompile(`(?i)\b[a-z]:[\\/](?:users|windows|jovm|bot_vm|work)(?:[\\/]|\b)`)
var credentials = regexp.MustCompile(`(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|-----BEGIN [A-Z ]*PRIVATE KEY-----|6L[A-Za-z0-9_-]{35,})`)

func main() {
	write := flag.Bool("write-manifest", false, "refresh publication hashes after reviewing all changed files")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "run from repository root with no positional arguments")
		os.Exit(1)
	}
	if err := check(".", *write); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func files(root string) ([]string, error) {
	var names []string
	err := filepath.WalkDir(root, func(path string, item fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if item.IsDir() {
			if rel == ".git" || rel == ".work" {
				return filepath.SkipDir
			}
			for _, part := range strings.Split(rel, "/") {
				switch strings.ToLower(part) {
				case "private", "files", "captures", "node_modules", "recon", "solve":
					return fmt.Errorf("FORBIDDEN_PUBLICATION_PATH: %s", rel)
				}
			}
			return nil
		}
		if item.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("PUBLICATION_SYMLINK: %s", rel)
		}
		if rel != "PUBLICATION.json" {
			names = append(names, rel)
		}
		return nil
	})
	sort.Strings(names)
	return names, err
}

func inspect(root, name string, data []byte) error {
	if drivePath.Match(data) || credentials.Match(data) {
		return fmt.Errorf("SENSITIVE_CONTENT_PATTERN: %s", name)
	}
	for _, marker := range []string{"private/tasks/", "raw_payload.json", "antcpt.com", "gstatic.com/recaptcha/releases/"} {
		if name != "cmd/check/main.go" && name != "cmd/check/main_test.go" && bytes.Contains(bytes.ToLower(data), []byte(marker)) {
			return fmt.Errorf("WITHHELD_CONTENT_MARKER: %s", name)
		}
	}
	for _, prefix := range []string{"cmd/capture/", "cmd/fetch/", "cmd/payload/", "cmd/sourceflow/", "internal/acquire/", "internal/runtimecontext/", "vms/vm_02/src/"} {
		if strings.HasPrefix(name, prefix) {
			return fmt.Errorf("WITHHELD_OPERATIONAL_TOOL: %s", name)
		}
	}
	if strings.Contains(name, "decompiler") {
		return fmt.Errorf("WITHHELD_DECOMPILER: %s", name)
	}
	if strings.HasSuffix(name, ".go") {
		formatted, err := format.Source(data)
		if err != nil || !bytes.Equal(formatted, data) {
			return fmt.Errorf("GO_FORMAT: %s", name)
		}
	}
	if strings.HasSuffix(name, ".md") {
		for _, match := range links.FindAllSubmatch(data, -1) {
			target := string(match[1])
			if target == "" {
				target = string(match[2])
			}
			target = strings.Trim(target, "<>")
			if strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "#") {
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			if strings.Contains(target, ":") || filepath.IsAbs(target) {
				return fmt.Errorf("NONPORTABLE_DOC_LINK: %s", name)
			}
			resolved := filepath.Clean(filepath.Join(root, filepath.Dir(name), filepath.FromSlash(target)))
			rel, err := filepath.Rel(root, resolved)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return fmt.Errorf("OUTSIDE_DOC_LINK: %s", name)
			}
			if _, err := os.Stat(resolved); err != nil {
				return fmt.Errorf("BROKEN_DOC_LINK: %s -> %s", name, target)
			}
		}
	}
	return nil
}

func check(root string, write bool) error {
	names, err := files(root)
	if err != nil {
		return err
	}
	actual := manifest{Schema: 1, Scope: "Redacted edition; authored public fixtures; excludes this manifest and local .git/.work directories."}
	asm := 0
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		if err := inspect(root, name, data); err != nil {
			return err
		}
		if strings.HasSuffix(name, ".asm") {
			if name != "output/vm_01_output.asm" && name != "output/vm_02_output.asm" {
				return fmt.Errorf("UNEXPECTED_ASM: %s", name)
			}
			if !bytes.Contains(data, []byte("Redacted full research listing")) {
				return fmt.Errorf("ASM_PROVENANCE_MISSING: %s", name)
			}
			if err := validateListing(name, data); err != nil {
				return err
			}
			asm++
		}
		sum := sha256.Sum256(data)
		actual.Files = append(actual.Files, entry{name, len(data), hex.EncodeToString(sum[:])})
	}
	if asm != 2 {
		return fmt.Errorf("ASM_COUNT: %d; expected one per evidenced family", asm)
	}
	encoded, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	path := filepath.Join(root, "PUBLICATION.json")
	if write {
		if err := os.WriteFile(path, encoded, 0600); err != nil {
			return err
		}
	} else {
		expected, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Equal(expected, encoded) {
			return fmt.Errorf("PUBLICATION_MANIFEST_DRIFT: review changes before --write-manifest")
		}
	}
	fmt.Printf("Publication checks passed: %d files, two redacted full ASM listings, local links, Go formatting and sensitive-content patterns\n", len(names))
	return nil
}

func validateListing(name string, data []byte) error {
	for _, value := range data {
		if value >= 128 {
			return fmt.Errorf("ASM_UNREVIEWED_NON_ASCII: encoded pools must be censored in this edition")
		}
	}
	text := string(data)
	if strings.HasSuffix(name, "vm_01_output.asm") {
		records := regexp.MustCompile(`(?m)^H_\d{3}:`).FindAllString(text, -1)
		if len(records) != 77 || !strings.Contains(text, ".unresolved CONSUMING_INTERPRETER_AND_PROGRAM") || strings.Count(text, ".source [REDACTED:") != 77 {
			return fmt.Errorf("STACK_LISTING_STRUCTURE: expected 77 records with censored sources and explicit missing consumer")
		}
		return nil
	}
	expected := []int{279, 5272, 1, 1, 5897, 1}
	programs := strings.Split(text, "\n.program ")[1:]
	if len(programs) != len(expected) {
		return fmt.Errorf("REGISTER_STREAM_COUNT: %d", len(programs))
	}
	instructions := regexp.MustCompile(`(?m)^(\d{5}):\s+(\S+)`)
	targets := regexp.MustCompile(`\bL_(\d{5})\b`)
	constants := regexp.MustCompile(`\bC_\d{3}\b`)
	definitions := regexp.MustCompile(`(?m)^\.const (C_\d{3}) = `)
	for index, program := range programs {
		records := instructions.FindAllStringSubmatch(program, -1)
		if len(records) != expected[index] {
			return fmt.Errorf("REGISTER_INSTRUCTION_COUNT: program_%02d = %d", index, len(records))
		}
		for ordinal, record := range records {
			if record[1] != fmt.Sprintf("%05d", ordinal) {
				return fmt.Errorf("REGISTER_INSTRUCTION_ORDER: program_%02d", index)
			}
		}
		for _, target := range targets.FindAllStringSubmatch(program, -1) {
			var ordinal int
			if _, err := fmt.Sscanf(target[1], "%d", &ordinal); err != nil || ordinal < 0 || ordinal > len(records) {
				return fmt.Errorf("REGISTER_TARGET_RANGE: program_%02d", index)
			}
		}
		known := map[string]bool{}
		for _, definition := range definitions.FindAllStringSubmatch(program, -1) {
			known[definition[1]] = true
		}
		for _, reference := range constants.FindAllString(program, -1) {
			if !known[reference] {
				return fmt.Errorf("MISSING_CONSTANT: program_%02d %s", index, reference)
			}
		}
	}
	if strings.Count(text, ".program_alias ") != 1 || strings.Count(text, ".unknown_string ") != 10 {
		return fmt.Errorf("REGISTER_ALIAS_OR_UNKNOWNS_CHANGED")
	}
	return nil
}
