package vm02

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type Constant struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
type StringAnnotation struct {
	Address int    `json:"address"`
	Value   string `json:"value"`
	Entry   int    `json:"entry"`
	End     int    `json:"end"`
}
type StringCatalogue struct {
	Sample          Sample             `json:"sample"`
	ProgramBytecode string             `json:"programBytecode"`
	AnchorBytecode  string             `json:"anchorBytecode,omitempty"`
	Annotations     []StringAnnotation `json:"annotations"`
	UnknownLoads    []UnknownString    `json:"unknownLoads,omitempty"`
}

func jsonText(value any) string {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
	return strings.TrimSuffix(buffer.String(), "\n")
}

func Constants(program *Program) []Constant {
	constants := []Constant{}
	seen := map[string]bool{}
	for _, instruction := range program.Instructions {
		for _, atom := range instruction.Operands {
			if atom.Kind != "string" {
				continue
			}
			value := atom.Value.(string)
			if len(jsonText(value)) <= 48 || seen[value] {
				continue
			}
			seen[value] = true
			constants = append(constants, Constant{ID: fmt.Sprintf("C_%03d", len(constants)), Value: value})
		}
	}
	return constants
}

func assemblyMnemonic(instruction Instruction) (string, string) {
	comment := ""
	switch instruction.Semantic {
	case "MOVE":
		if len(instruction.Operands) == 1 && instruction.Operands[0].Kind != "register" {
			return "LOAD_CONST", comment
		}
		return "MOV", comment
	case "MOVE_REGISTER":
		return "MOV", comment
	case "ADD":
		return "CONCAT", comment
	case "GET_PROPERTY":
		return "GET_PROP", comment
	case "SET_PROPERTY":
		return "SET_PROP", comment
	case "CALLBACK_RESULT":
		return "SEND_RESULT", comment
	case "CALLBACK_REGISTERS":
		return "SEND", comment
	case "CHAR_CODES":
		return "STR_CODES", comment
	case "CONCAT_INDEXED":
		return "STR_DEC", comment
	case "STRING_HASH":
		return "HASH", comment
	case "COMPACT_FORMAT":
		return "FORMAT", comment
	case "MAKE_CLOSURE":
		return "FUNCTION", comment
	case "CONSTRUCT":
		return "NEW", comment
	case "CALL_FUNCTION":
		return "APPLY", comment
	case "GUARDED_VALUE":
		return "LOAD_GUARDED", comment
	case "LCG_CLOSURE":
		return "LCG", comment
	case "MAP_HANDLER":
		return "MAP", comment
	case "LOAD_NULL":
		return "NULL", comment
	case "BRANCH_EQ":
		return "JE", comment
	case "BRANCH_LT":
		return "JL", comment
	case "BINARY_MAP_LEFT":
		if instruction.Operation != nil && instruction.Operation.Kind == "binary" && instruction.Operation.Left != nil && instruction.Operation.Right != nil {
			aliases := map[string]string{"^": "XOR", "+": "ADD", "-": "SUB", "*": "MUL", "/": "DIV", "%": "MOD", "||": "OR", "&&": "AND"}
			if alias, ok := aliases[instruction.Operation.Operator]; ok {
				if instruction.Operation.Operator == "&&" || instruction.Operation.Operator == "||" {
					if isDoubleNotArgument(instruction.Operation.Left, 0) && isDoubleNotArgument(instruction.Operation.Right, 1) {
						return "BOOL_" + alias, comment
					}
					if isArgument(instruction.Operation.Left, 0) && isArgument(instruction.Operation.Right, 1) {
						return alias, comment
					}
					return "BINARY_MAP_LEFT", " ; operation=" + jsonText(instruction.Operation)
				}
				if isArgument(instruction.Operation.Left, 0) && isArgument(instruction.Operation.Right, 1) {
					return alias, comment
				}
			}
		}
		return "BINARY_MAP_LEFT", " ; operation=" + jsonText(instruction.Operation)
	}
	return instruction.Semantic, comment
}

func isArgument(operation *Operation, index int) bool {
	return operation != nil && operation.Kind == "argument" && operation.Index != nil && *operation.Index == index
}

func isDoubleNotArgument(operation *Operation, index int) bool {
	return operation != nil && operation.Kind == "not" && isNotArgument(operation.Value, index)
}

func isNotArgument(operation *Operation, index int) bool {
	return operation != nil && operation.Kind == "not" && isArgument(operation.Value, index)
}

func sameFiniteNumericLiteral(left, right any) bool {
	a, aok := numeric(left)
	b, bok := numeric(right)
	return aok && bok && !math.IsNaN(a) && !math.IsInf(a, 0) && !math.IsNaN(b) && !math.IsInf(b, 0) && a == b
}

func Assembly(program *Program, catalogue *StringCatalogue) (string, error) {
	annotations := map[int]string{}
	if catalogue != nil {
		if catalogue.Sample.SHA256 != program.Sample.SHA256 || catalogue.ProgramBytecode != program.Bytecode.SHA256 {
			return "", fmt.Errorf("STRING_CATALOGUE_PROGRAM_MISMATCH")
		}
		boundaries := map[int]int{}
		for _, instruction := range program.Instructions {
			boundaries[instruction.Address] = instruction.End
		}
		for _, annotation := range catalogue.Annotations {
			end, exists := boundaries[annotation.Address]
			if !exists || annotation.Entry > annotation.Address || end != annotation.End {
				return "", fmt.Errorf("INVALID_STRING_ANNOTATION: %d", annotation.Address)
			}
			if _, exists := annotations[annotation.Address]; exists {
				return "", fmt.Errorf("DUPLICATE_STRING_ANNOTATION: %d", annotation.Address)
			}
			annotations[annotation.Address] = annotation.Value
		}
	}
	constants := map[string]string{}
	for _, constant := range Constants(program) {
		constants[constant.Value] = constant.ID
	}
	labels := map[int]bool{0: true}
	for _, instruction := range program.Instructions {
		if instruction.Target != nil {
			labels[*instruction.Target] = true
		}
		if instruction.Flow["kind"] == "branch" || instruction.Flow["kind"] == "yield" {
			labels[instruction.End] = true
		}
	}
	label := func(address int) string { return fmt.Sprintf("L_%06x", address) }
	var result strings.Builder
	fmt.Fprintf(&result, "; source sha256   %s\n; bytecode sha256 %s\n\n", program.Sample.SHA256, program.Bytecode.SHA256)
	for _, instruction := range program.Instructions {
		if labels[instruction.Address] {
			fmt.Fprintf(&result, "%s:\n", label(instruction.Address))
		}
		operands := []string{}
		mnemonic, mnemonicComment := assemblyMnemonic(instruction)
		shortJump := false
		if instruction.Semantic == "BRANCH_EQ" && len(instruction.Operands) == 3 && instruction.Target != nil && instruction.Operands[1].Kind == "number" && instruction.Operands[2].Kind == "number" {
			shortJump = sameFiniteNumericLiteral(instruction.Operands[1].Value, instruction.Operands[2].Value)
		}
		if instruction.Destination != nil {
			operands = append(operands, fmt.Sprintf("R%d", *instruction.Destination))
		}
		for index, atom := range instruction.Operands {
			if shortJump && index > 0 {
				continue
			}
			switch {
			case shortJump && index == 0:
				operands = append(operands, label(*instruction.Target))
			case index == 0 && instruction.Target != nil:
				operands = append(operands, label(*instruction.Target))
			case atom.IndexValue != nil:
				operands = append(operands, fmt.Sprintf("R%d", *atom.IndexValue))
			case atom.Kind == "register":
				operands = append(operands, fmt.Sprintf("R%d", *atom.Index))
			case atom.Kind == "string":
				if id, ok := constants[atom.Value.(string)]; ok {
					operands = append(operands, id)
				} else {
					operands = append(operands, jsonText(atom.Value))
				}
			default:
				if special, ok := atom.Value.(string); ok {
					operands = append(operands, special)
				} else {
					operands = append(operands, fmt.Sprint(atom.Value))
				}
			}
		}
		if (instruction.Semantic == "BRANCH_EQ" || instruction.Semantic == "BRANCH_LT") && len(instruction.Operands) == 3 && !shortJump {
			operands = operands[:0]
			for _, index := range []int{1, 2, 0} {
				atom := instruction.Operands[index]
				if index == 0 && instruction.Target != nil {
					operands = append(operands, label(*instruction.Target))
				} else if atom.Index != nil && atom.Kind == "register" {
					operands = append(operands, fmt.Sprintf("R%d", *atom.Index))
				} else if atom.Kind == "string" {
					if id, ok := constants[atom.Value.(string)]; ok {
						operands = append(operands, id)
					} else {
						operands = append(operands, jsonText(atom.Value))
					}
				} else {
					if special, ok := atom.Value.(string); ok {
						operands = append(operands, special)
					} else {
						operands = append(operands, fmt.Sprint(atom.Value))
					}
				}
			}
		}
		if shortJump {
			mnemonic = "JMP"
		}
		line := strings.TrimRight(fmt.Sprintf("0x%06x:  %-16s %s", instruction.Address, mnemonic, strings.Join(operands, ", ")), " ") + mnemonicComment
		if text, ok := annotations[instruction.Address]; ok {
			line += "  ; " + jsonText(text)
		}
		result.WriteString(line + "\n")
	}
	if labels[program.Bytecode.Bytes] {
		fmt.Fprintf(&result, "%s: ; end\n", label(program.Bytecode.Bytes))
	}
	for _, unknown := range program.UnknownRegions {
		fmt.Fprintf(&result, "%s: .unknown %s ; %d bytes\n", label(unknown.Address), unknown.Diagnostic.Code, unknown.End-unknown.Address)
	}
	for _, target := range program.InvalidTargets {
		fmt.Fprintf(&result, "; .invalid_target %s -> %s: %s\n", label(target.Address), label(target.Target), target.Diagnostic)
	}
	if pool := Constants(program); len(pool) > 0 {
		result.WriteString("\n; Complete string constants for this program. JSON escaping preserves control characters.\n")
		for _, constant := range pool {
			fmt.Fprintf(&result, ".const %s = %s\n", constant.ID, jsonText(constant.Value))
		}
	}
	if catalogue != nil {
		result.WriteString(strings.ReplaceAll(catalogue.UnknownText(), " r", " R"))
	}
	return result.String(), nil
}
