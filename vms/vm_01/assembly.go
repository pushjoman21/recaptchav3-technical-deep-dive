// Package vm01 reports recovered stack-handler evidence. No consuming bytecode
// interpreter has been established for these registrations.
package vm01

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type HandlerCatalogue struct {
	Sample struct {
		SHA256 string `json:"sha256"`
	} `json:"sample"`
	Qualified bool `json:"qualified"`
	Handlers  []struct {
		IDs          []int64  `json:"ids"`
		Semantic     string   `json:"semantic"`
		Purpose      string   `json:"purpose"`
		Dependencies []string `json:"dependencies"`
		Handler      struct {
			Member   string `json:"member"`
			Selector int64  `json:"selector"`
			Evidence struct {
				Start struct {
					Line int `json:"line"`
				} `json:"start"`
			} `json:"evidence"`
		} `json:"handler"`
	} `json:"handlers"`
}

// Assembly renders a source-qualified catalogue. It does not turn registration
// IDs into opcodes, instruction addresses or a reconstructed program.
func Assembly(data []byte, sampleHash string) (string, error) {
	var catalogue HandlerCatalogue
	if err := json.Unmarshal(data, &catalogue); err != nil {
		return "", err
	}
	if !catalogue.Qualified || catalogue.Sample.SHA256 != sampleHash || len(catalogue.Handlers) == 0 {
		return "", fmt.Errorf("UNQUALIFIED_STACK_HANDLER_CATALOGUE")
	}
	var out strings.Builder
	fmt.Fprintf(&out, "; vm_01 stack handlers; IDs are registration metadata.\n; source %s\n.unresolved CONSUMING_INTERPRETER_AND_PROGRAM\n\n", sampleHash)
	for index, handler := range catalogue.Handlers {
		if len(handler.IDs) == 0 || handler.Semantic == "" || strings.ContainsAny(handler.Semantic, "\r\n\t ") || handler.Handler.Member == "" || handler.Handler.Evidence.Start.Line < 1 {
			return "", fmt.Errorf("INVALID_STACK_HANDLER_RECORD: %d", index)
		}
		ids := make([]string, len(handler.IDs))
		for i, id := range handler.IDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		if strings.ContainsAny(handler.Purpose, "\r\n") || strings.ContainsAny(handler.Handler.Member, "\r\n") {
			return "", fmt.Errorf("INVALID_MULTILINE_HANDLER_METADATA: %d", index)
		}
		for _, name := range handler.Dependencies {
			if strings.ContainsAny(name, "\r\n") {
				return "", fmt.Errorf("INVALID_HANDLER_DEPENDENCY: %d", index)
			}
		}
		fmt.Fprintf(&out, "H_%03d:  %-26s ; ids %s\n  ; %s\n  .source %s(%d) @%d\n", index, handler.Semantic, strings.Join(ids, ", "), handler.Purpose, handler.Handler.Member, handler.Handler.Selector, handler.Handler.Evidence.Start.Line)
		if len(handler.Dependencies) > 0 {
			fmt.Fprintf(&out, "  .requires %s\n", strings.Join(handler.Dependencies, ", "))
		}
		out.WriteByte('\n')
	}
	return out.String(), nil
}
