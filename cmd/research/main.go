// research provides offline analysis; inputs are parsed as data and never executed.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"recaptchav3research/internal/artifact"
	"recaptchav3research/internal/deobfuscate"
	"recaptchav3research/internal/jsast"
	vm01 "recaptchav3research/vms/vm_01"
	vm02 "recaptchav3research/vms/vm_02"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: research <inventory|deobfuscate|decode|stack> [options]; public synthetic profile only")
	}
	switch args[0] {
	case "inventory":
		return inventory(args[1:])
	case "deobfuscate":
		return deobfuscateSource(args[1:])
	case "decode":
		return decode(args[1:])
	case "stack":
		return stack(args[1:])
	default:
		return fmt.Errorf("UNKNOWN_OR_WITHHELD_COMMAND: %s; see docs/REDACTIONS.md", args[0])
	}
}

func stack(args []string) error {
	flags := flag.NewFlagSet("stack", flag.ContinueOnError)
	input := flags.String("input", "", "authored handler catalogue JSON")
	output := flags.String("out", "", "new output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" || *output == "" || flags.NArg() != 0 {
		return fmt.Errorf("stack requires --input and --out")
	}
	var catalogue vm01.HandlerCatalogue
	data, err := readJSONBytes(*input, &catalogue)
	if err != nil {
		return err
	}
	text, err := vm01.Assembly(data, catalogue.Sample.SHA256)
	if err != nil {
		return err
	}
	directory, err := artifact.CreateDirectory(*output, []string{*input})
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "vm_01_output.asm"), []byte(text), 0600); err != nil {
		return err
	}
	fmt.Println("Authored catalogue written; consuming interpreter and program remain unresolved")
	return nil
}

func deobfuscateSource(args []string) error {
	flags := flag.NewFlagSet("deobfuscate", flag.ContinueOnError)
	input := flags.String("input", "", "JavaScript source; parsed without execution")
	output := flags.String("out", "", "new output directory")
	intrinsics := flags.Bool("assume-standard-intrinsics", false, "assume the supported String intrinsics retain standard behavior")
	initialized := flags.Bool("assume-initialized-decoders", false, "replace invariant calls assuming factories have initialized and dynamic code preserves bindings")
	fuel := flags.Int("fuel", 2000000, "maximum AST reduction steps per factory invocation")
	strict := flags.Bool("strict", false, "fail if no supported factory exists or any observed factory call is unresolved")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" || *output == "" || flags.NArg() != 0 {
		return fmt.Errorf("deobfuscate requires --input and --out")
	}
	source, err := jsast.Read(*input)
	if err != nil {
		return err
	}
	report, err := deobfuscate.Recover(source, deobfuscate.Options{AssumeStandardIntrinsics: *intrinsics, AssumeInitializedDecoders: *initialized, Fuel: *fuel})
	if err != nil {
		return err
	}
	text, err := deobfuscate.Rewrite(source, report)
	if err != nil {
		return err
	}
	directory, err := artifact.CreateDirectory(*output, []string{*input})
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(directory, "report.json"), append(data, '\n'), 0600); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(directory, "deobfuscated.js"), []byte(text), 0600); err != nil {
		return err
	}
	if *strict && len(report.Factories) == 0 {
		return fmt.Errorf("NO_SUPPORTED_STRING_FACTORY: evidence in %s", directory)
	}
	if *strict && len(report.Strings) == 0 {
		return fmt.Errorf("NO_OBSERVED_FACTORY_CALL: evidence in %s", directory)
	}
	if *strict && report.Recovered != len(report.Strings) {
		return fmt.Errorf("UNRESOLVED_STRING_FACTORIES: %d/%d; evidence in %s", len(report.Strings)-report.Recovered, len(report.Strings), directory)
	}
	fmt.Printf("Recovered %d/%d factory strings, replaced %d calls in %s\n", report.Recovered, len(report.Strings), len(report.Edits), directory)
	return nil
}

func readJSON(filename string, value any) error {
	_, err := readJSONBytes(filename, value)
	return err
}

func readJSONBytes(filename string, value any) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 32*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 32*1024*1024 {
		return nil, fmt.Errorf("JSON_INPUT_TOO_LARGE")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("TRAILING_JSON_VALUE")
		}
		return nil, err
	}
	return data, nil
}

func decode(args []string) error {
	flags := flag.NewFlagSet("decode", flag.ContinueOnError)
	configurationFile := flags.String("configuration", "", "authored synthetic configuration JSON; production profiles are withheld")
	input := flags.String("input", "", "bytecode file")
	output := flags.String("out", "", "new output directory")
	stringsFile := flags.String("strings", "", "optional recorded string catalogue; verified against both hashes")
	// [REDACTED: decompiler and executable runtime are not distributed]
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *configurationFile == "" || *input == "" || *output == "" || flags.NArg() != 0 {
		return fmt.Errorf("decode requires --configuration, --input and --out")
	}
	var configuration vm02.Configuration
	if err := readJSON(*configurationFile, &configuration); err != nil {
		return err
	}
	bytes, err := os.ReadFile(*input)
	if err != nil {
		return err
	}
	program, err := vm02.Decode(bytes, configuration)
	if err != nil {
		return err
	}
	var strings *vm02.StringCatalogue
	inputs := []string{*configurationFile, *input}
	if *stringsFile != "" {
		strings = &vm02.StringCatalogue{}
		if err := readJSON(*stringsFile, strings); err != nil {
			return err
		}
		inputs = append(inputs, *stringsFile)
	}
	assembly, err := vm02.Assembly(program, strings)
	if err != nil {
		return err
	}
	directory, err := artifact.CreateDirectory(*output, inputs)
	if err != nil {
		return err
	}
	writeJSON := func(name string, value any) error {
		file, err := os.OpenFile(filepath.Join(directory, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(file)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(value)
		closeErr := file.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	if err := writeJSON("program.json", program); err != nil {
		return err
	}
	if err := writeJSON("constants.json", vm02.Constants(program)); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "vm_02_output.asm"), []byte(assembly), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "bytecode.bin"), bytes, 0600); err != nil {
		return err
	}
	if program.Status != "decoded" {
		return fmt.Errorf("PARTIAL_DECODE: %d unknown regions, %d invalid targets; evidence written to %s", len(program.UnknownRegions), len(program.InvalidTargets), directory)
	}
	fmt.Printf("Decoded %d instructions / %d bytes to %s\n", len(program.Instructions), program.Bytecode.Bytes, directory)
	return nil
}

func inventory(args []string) error {
	flags := flag.NewFlagSet("inventory", flag.ContinueOnError)
	input := flags.String("input", "", "JavaScript source; parsed without execution")
	output := flags.String("out", "", "new JSON report; default stdout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" || flags.NArg() != 0 {
		return fmt.Errorf("inventory requires --input and no positional arguments")
	}
	source, err := jsast.Read(*input)
	if err != nil {
		return err
	}
	functions, err := jsast.Functions(source)
	if err != nil {
		return err
	}
	report := struct {
		SchemaVersion int              `json:"schemaVersion"`
		Sample        any              `json:"sample"`
		Functions     []jsast.Function `json:"functions"`
		Limitations   []string         `json:"limitations"`
	}{1, map[string]any{"filename": *input, "bytes": len(source.Text), "sha256": source.SHA256}, functions, []string{"Structural function inventory only; this command does not infer function purposes or VM identity.", "Public edition: production bindings and runtime capture tools are withheld; see docs/REDACTIONS.md."}}
	writer := os.Stdout
	if *output != "" {
		writer, err = os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		defer writer.Close()
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
