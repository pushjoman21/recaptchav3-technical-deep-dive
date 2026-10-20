package vm02

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf16"
)

func numeric(value any) (float64, bool) {
	switch value := value.(type) {
	case int:
		return float64(value), true
	case int32:
		return float64(value), true
	case float64:
		return value, true
	default:
		return 0, false
	}
}
func int32Number(value float64) int32 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value == 0 {
		return 0
	}
	value = math.Mod(math.Trunc(value), 4294967296)
	if value < 0 {
		value += 4294967296
	}
	return int32(uint32(value))
}

// numericOperation accepts only the numeric grammar used by the proven loader.
// Boolean coercions and host object properties cannot enter this reducer.
func numericOperation(operation *Operation, left, right float64) (float64, error) {
	if operation == nil {
		return 0, fmt.Errorf("STRING_OPERATOR_MISSING")
	}
	switch operation.Kind {
	case "argument":
		if operation.Index != nil {
			if *operation.Index == 0 {
				return left, nil
			}
			if *operation.Index == 1 {
				return right, nil
			}
		}
	case "binary":
		a, err := numericOperation(operation.Left, left, right)
		if err != nil {
			return 0, err
		}
		b, err := numericOperation(operation.Right, left, right)
		if err != nil {
			return 0, err
		}
		switch operation.Operator {
		case "^":
			return float64(int32Number(a) ^ int32Number(b)), nil
		case "+":
			return a + b, nil
		case "-":
			return a - b, nil
		case "*":
			return a * b, nil
		case "/":
			return a / b, nil
		case "%":
			return math.Mod(a, b), nil
		}
	}
	return 0, fmt.Errorf("UNSUPPORTED_STRING_NUMERIC_OPERATOR: %s %s", operation.Kind, operation.Operator)
}

type UnknownString struct {
	Entry      int    `json:"entry"`
	Address    int    `json:"address"`
	Register   *int   `json:"register,omitempty"`
	Diagnostic string `json:"diagnostic"`
}

// DecodeStrings reduces contiguous, single-destination character-code loaders.
// It does not execute instructions or rewrite their evaluation/cursor effects.
func DecodeStrings(anchor, program *Program) (*StringCatalogue, error) {
	if anchor.Sample.SHA256 != program.Sample.SHA256 {
		return nil, fmt.Errorf("STRING_BANK_SAMPLE_MISMATCH")
	}
	if anchor.Status != "decoded" || program.Status != "decoded" {
		return nil, fmt.Errorf("STRING_PROGRAM_PARTIAL")
	}
	bank := map[int]any{}
	constantValue := func(atom Atom) (any, bool) {
		if atom.Kind == "number" {
			if value, ok := atom.Value.(string); ok {
				switch value {
				case "NaN":
					return math.NaN(), true
				case "Infinity":
					return math.Inf(1), true
				case "-Infinity":
					return math.Inf(-1), true
				case "-0":
					return math.Copysign(0, -1), true
				default:
					return nil, false
				}
			}
		}
		return atom.Value, true
	}
	for _, instruction := range anchor.Instructions {
		if instruction.Semantic == "MOVE" && instruction.Destination != nil && len(instruction.Operands) == 1 && instruction.Operands[0].Kind != "register" {
			if value, known := constantValue(instruction.Operands[0]); known {
				bank[*instruction.Destination] = value
			}
		}
	}
	valueOf := func(atom Atom) (any, bool) {
		if atom.Kind == "register" {
			if atom.Index == nil {
				return nil, false
			}
			value, ok := bank[*atom.Index]
			return value, ok
		}
		return constantValue(atom)
	}
	targets := map[int]bool{}
	for _, instruction := range program.Instructions {
		if instruction.Target != nil {
			targets[*instruction.Target] = true
		}
	}
	for _, entry := range program.FunctionEntries {
		if address, ok := entry["address"].(int); ok {
			targets[address] = true
		}
	}
	result := &StringCatalogue{Sample: program.Sample, ProgramBytecode: program.Bytecode.SHA256, AnchorBytecode: anchor.Bytecode.SHA256, Annotations: []StringAnnotation{}, UnknownLoads: []UnknownString{}}
	for start, first := range program.Instructions {
		if first.Semantic != "CHAR_CODES" || first.Destination == nil || len(first.Operands) != 1 {
			continue
		}
		cursor := start + 1
		type step struct {
			operation *Operation
			right     float64
		}
		operations := []step{}
		for cursor < len(program.Instructions) {
			instruction := program.Instructions[cursor]
			if targets[instruction.Address] || instruction.Semantic != "BINARY_MAP_LEFT" || instruction.Destination == nil || *instruction.Destination != *first.Destination || len(instruction.Operands) != 2 || instruction.Operands[0].Index == nil || *instruction.Operands[0].Index != *first.Destination {
				break
			}
			right, known := valueOf(instruction.Operands[1])
			number, numberOK := numeric(right)
			if !known || !numberOK || instruction.Operation == nil {
				break
			}
			operations = append(operations, step{operation: instruction.Operation, right: number})
			cursor++
		}
		if len(operations) == 0 || cursor >= len(program.Instructions) {
			continue
		}
		last := program.Instructions[cursor]
		if targets[last.Address] || last.Semantic != "CONCAT_INDEXED" || len(last.Operands) != 2 || last.Operands[1].Index == nil || *last.Operands[1].Index != *first.Destination {
			continue
		}
		alphabetValue, known := valueOf(last.Operands[0])
		alphabet, alphabetOK := alphabetValue.(string)
		if !known || !alphabetOK {
			result.UnknownLoads = append(result.UnknownLoads, UnknownString{Entry: first.Address, Address: last.Address, Register: last.Operands[0].Index, Diagnostic: "ALPHABET_REGISTER_UNAVAILABLE"})
			continue
		}
		encodedValue, known := valueOf(first.Operands[0])
		encoded, encodedOK := encodedValue.(string)
		if !known || !encodedOK {
			result.UnknownLoads = append(result.UnknownLoads, UnknownString{Entry: first.Address, Address: last.Address, Register: first.Operands[0].Index, Diagnostic: "ENCODED_REGISTER_UNAVAILABLE"})
			continue
		}
		alphabetUnits := utf16.Encode([]rune(alphabet))
		units := []uint16{}
		unsupported := ""
		for _, unit := range utf16.Encode([]rune(encoded)) {
			index := float64(unit)
			for _, operation := range operations {
				value, err := numericOperation(operation.operation, index, operation.right)
				if err != nil {
					unsupported = err.Error()
					break
				}
				index = value
			}
			if unsupported != "" {
				break
			}
			if math.IsNaN(index) || math.IsInf(index, 0) || index < 0 || index != math.Trunc(index) || index >= float64(len(alphabetUnits)) {
				units = append(units, utf16.Encode([]rune("undefined"))...)
			} else {
				units = append(units, alphabetUnits[int(index)])
			}
		}
		if unsupported != "" {
			result.UnknownLoads = append(result.UnknownLoads, UnknownString{Entry: first.Address, Address: last.Address, Diagnostic: unsupported})
			continue
		}
		valid := true
		for index := 0; index < len(units); index++ {
			unit := units[index]
			if unit >= 0xd800 && unit <= 0xdbff {
				if index+1 >= len(units) || units[index+1] < 0xdc00 || units[index+1] > 0xdfff {
					valid = false
					break
				}
				index++
			} else if unit >= 0xdc00 && unit <= 0xdfff {
				valid = false
				break
			}
		}
		if !valid {
			result.UnknownLoads = append(result.UnknownLoads, UnknownString{Entry: first.Address, Address: last.Address, Diagnostic: "UNPAIRED_UTF16_IN_DECODED_STRING"})
			continue
		}
		result.Annotations = append(result.Annotations, StringAnnotation{Address: last.Address, Value: string(utf16.Decode(units)), Entry: first.Address, End: last.End})
	}
	return result, nil
}

func SameStringAnnotations(expected, actual *StringCatalogue) bool {
	if expected.Sample.SHA256 != actual.Sample.SHA256 || expected.ProgramBytecode != actual.ProgramBytecode || len(expected.Annotations) != len(actual.Annotations) {
		return false
	}
	if expected.AnchorBytecode != "" && expected.AnchorBytecode != actual.AnchorBytecode {
		return false
	}
	for index, annotation := range expected.Annotations {
		if annotation != actual.Annotations[index] {
			return false
		}
	}
	return true
}

func (catalogue *StringCatalogue) UnknownText() string {
	var result strings.Builder
	for _, unknown := range catalogue.UnknownLoads {
		register := ""
		if unknown.Register != nil {
			register = fmt.Sprintf(" r%d", *unknown.Register)
		}
		fmt.Fprintf(&result, ".unknown_string L_%06x%s ; %s; sequence starts L_%06x\n", unknown.Address, register, unknown.Diagnostic, unknown.Entry)
	}
	return result.String()
}
