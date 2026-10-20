// Package vm02 exposes the redacted research decoder through an authored fixture profile.
// Configuration extraction and semantic curation are separate from wire decoding.
package vm02

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"unicode/utf8"
)

// SyntheticProfileSHA256 identifies authored fixtures, not a production sample.
// [REDACTED: production profile identity and binding extraction]
const SyntheticProfileSHA256 = "400a1f538c879e79b7d1b75e27736879bb262fc8334b767fc8c4d4050ed95e1c"

type Sample struct {
	Filename string `json:"filename,omitempty"`
	Bytes    int    `json:"bytes"`
	SHA256   string `json:"sha256"`
}

type Layout struct {
	Destination      bool  `json:"destination"`
	Min              int   `json:"min"`
	Max              int   `json:"max"`
	AllRegisters     bool  `json:"allRegisters,omitempty"`
	Offset           *int  `json:"offset,omitempty"`
	IndexOperands    []int `json:"indexOperands,omitempty"`
	RegisterOperands []int `json:"registerOperands,omitempty"`
}

type Operation struct {
	Kind     string     `json:"kind"`
	Index    *int       `json:"index,omitempty"`
	Operator string     `json:"operator,omitempty"`
	Left     *Operation `json:"left,omitempty"`
	Right    *Operation `json:"right,omitempty"`
	Value    *Operation `json:"value,omitempty"`
}

type Handler struct {
	Opcode    int             `json:"opcode"`
	Semantic  string          `json:"semantic"`
	Layout    Layout          `json:"layout"`
	Operation *Operation      `json:"operation,omitempty"`
	Evidence  json.RawMessage `json:"evidence"`
	Operator  *struct {
		Evidence json.RawMessage `json:"evidence"`
	} `json:"operator,omitempty"`
	Delegation *struct {
		Evidence json.RawMessage `json:"evidence"`
	} `json:"delegation,omitempty"`
}

type Configuration struct {
	Sample  Sample    `json:"sample"`
	Entries []Handler `json:"entries"`
}

func pointer(value int) *int { return &value }

// Layouts describe proven semantic operations, not numeric opcode assignments.
func semanticLayouts() map[string]Layout {
	return map[string]Layout{
		"CALLBACK_RESULT": {Min: 0, Max: 0}, "MOVE": {Destination: true, Min: 1, Max: 1},
		"ADD": {Destination: true, Min: 2, Max: 2}, "BINARY_MAP_LEFT": {Destination: true, Min: 2, Max: 2},
		"CALL_METHOD": {Destination: true, Min: 2, Max: 255, AllRegisters: true}, "GET_PROPERTY": {Destination: true, Min: 2, Max: 2},
		"SET_PROPERTY": {Min: 3, Max: 3}, "CALLBACK_REGISTERS": {Min: 0, Max: 255, AllRegisters: true},
		"MOVE_REGISTER": {Destination: true, Min: 1, Max: 1, AllRegisters: true}, "LOAD_NULL": {Destination: true, Min: 0, Max: 0},
		"SET_RESULT": {Min: 1, Max: 1}, "SET_GLOBAL": {Min: 2, Max: 2}, "GET_GLOBAL": {Destination: true, Min: 1, Max: 1},
		"CALL_GLOBAL": {Destination: true, Min: 1, Max: 255}, "BRANCH_EQ": {Min: 3, Max: 3, Offset: pointer(0)},
		"STRING_HASH": {Destination: true, Min: 1, Max: 2}, "CHAR_CODES": {Destination: true, Min: 1, Max: 1},
		"REGEXP": {Destination: true, Min: 2, Max: 2}, "NOT": {Destination: true, Min: 1, Max: 1},
		"MAP_HANDLER":    {Destination: true, Min: 3, Max: 255, IndexOperands: []int{1, 2}},
		"COMPACT_FORMAT": {Destination: true, Min: 1, Max: 1}, "CLOCK": {Destination: true, Min: 0, Max: 0},
		"MAKE_CLOSURE": {Destination: true, Min: 2, Max: 2, Offset: pointer(0), IndexOperands: []int{1}},
		"BRANCH_LT":    {Min: 3, Max: 3, Offset: pointer(0)}, "SCHEDULE": {Min: 2, Max: 2, Offset: pointer(0), IndexOperands: []int{1}},
		"YIELD": {Min: 0, Max: 0}, "CONSTRUCT": {Destination: true, Min: 1, Max: 255},
		"CONCAT_INDEXED": {Destination: true, Min: 2, Max: 2, RegisterOperands: []int{1}},
		"LCG_CLOSURE":    {Destination: true, Min: 4, Max: 4}, "CALL_FUNCTION": {Destination: true, Min: 1, Max: 255, AllRegisters: true},
		"ELAPSED": {Destination: true, Min: 0, Max: 0}, "GUARDED_VALUE": {Destination: true, Min: 1, Max: 1},
		"TYPEOF": {Destination: true, Min: 1, Max: 1, AllRegisters: true},
	}
}

func (configuration Configuration) validate() (map[int]Handler, error) {
	if configuration.Sample.SHA256 != SyntheticProfileSHA256 {
		return nil, fmt.Errorf("UNSUPPORTED_WIRE_PROFILE: %s; public edition accepts only the authored synthetic profile", configuration.Sample.SHA256)
	}
	if len(configuration.Entries) == 0 {
		return nil, fmt.Errorf("EMPTY_HANDLER_CONFIGURATION")
	}
	result := make(map[int]Handler)
	layouts := semanticLayouts()
	for _, entry := range configuration.Entries {
		if entry.Opcode < 0 {
			return nil, fmt.Errorf("NEGATIVE_CONFIG_OPCODE: %d", entry.Opcode)
		}
		if _, exists := result[entry.Opcode]; exists {
			return nil, fmt.Errorf("DUPLICATE_CONFIG_OPCODE: %d", entry.Opcode)
		}
		expected, ok := layouts[entry.Semantic]
		if !ok {
			return nil, fmt.Errorf("UNKNOWN_CONFIG_SEMANTIC: %s", entry.Semantic)
		}
		actual := entry.Layout
		if actual.Destination != expected.Destination || actual.Min != expected.Min || actual.Max != expected.Max || actual.AllRegisters != expected.AllRegisters || !slices.Equal(actual.IndexOperands, expected.IndexOperands) || !slices.Equal(actual.RegisterOperands, expected.RegisterOperands) || (actual.Offset == nil) != (expected.Offset == nil) || (actual.Offset != nil && *actual.Offset != *expected.Offset) {
			return nil, fmt.Errorf("UNSUPPORTED_CONFIG_LAYOUT: %s", entry.Semantic)
		}
		if entry.Semantic == "BINARY_MAP_LEFT" {
			if entry.Operation == nil || entry.Operation.Kind != "binary" {
				return nil, fmt.Errorf("EXPECTED_BINARY_OPERATOR_IR")
			}
			if err := validateOperation(entry.Operation, 0); err != nil {
				return nil, err
			}
		}
		result[entry.Opcode] = entry
	}
	return result, nil
}

func validateOperation(operation *Operation, depth int) error {
	if operation == nil || depth > 16 {
		return fmt.Errorf("INVALID_OPERATOR_IR")
	}
	switch operation.Kind {
	case "argument":
		if operation.Index != nil && (*operation.Index == 0 || *operation.Index == 1) {
			return nil
		}
	case "not":
		return validateOperation(operation.Value, depth+1)
	case "binary":
		if slices.Contains([]string{"^", "+", "-", "*", "/", "%", "&&", "||"}, operation.Operator) {
			if err := validateOperation(operation.Left, depth+1); err != nil {
				return err
			}
			return validateOperation(operation.Right, depth+1)
		}
	}
	return fmt.Errorf("UNSUPPORTED_OPERATOR_IR: %s %s", operation.Kind, operation.Operator)
}

type Diagnostic struct {
	Code    string `json:"code"`
	Address int    `json:"address"`
	Message string `json:"message"`
}

func (diagnostic *Diagnostic) Error() string {
	return fmt.Sprintf("%s at byte %d: %s", diagnostic.Code, diagnostic.Address, diagnostic.Message)
}
func decodeError(code string, address int, detail string) error {
	return &Diagnostic{Code: code, Address: address, Message: detail}
}

type Atom struct {
	Kind           string `json:"kind"`
	Index          *int   `json:"index,omitempty"`
	Value          any    `json:"value,omitempty"`
	OuterTag       byte   `json:"outerTag"`
	TypeTag        byte   `json:"typeTag"`
	PayloadAddress int    `json:"payloadAddress"`
	Address        int    `json:"address"`
	End            int    `json:"end"`
	Raw            string `json:"raw"`
	IndexValue     *int   `json:"indexValue,omitempty"`
	ReadMode       string `json:"readMode,omitempty"`
}

type reader struct {
	bytes    []byte
	position int
}

func (r *reader) byte() (byte, error) {
	if r.position >= len(r.bytes) {
		return 0, decodeError("TRUNCATED_INPUT", r.position, "")
	}
	value := r.bytes[r.position]
	r.position++
	return value, nil
}
func (r *reader) int32() (int, error) {
	start := r.position
	var value uint32
	for i := 0; i < 10; i++ {
		b, err := r.byte()
		if err != nil {
			return 0, err
		}
		if i < 4 {
			value |= uint32(b&127) << uint(i*7)
		} else if i == 4 {
			value |= uint32(b) << 28
		}
		if b&128 == 0 {
			return int(int32(value)), nil
		}
	}
	return 0, decodeError("VARINT_OVERFLOW", start, "continuation exceeds ten bytes")
}
func (r *reader) boolean() (bool, error) {
	start := r.position
	value := false
	for i := 0; i < 10; i++ {
		b, err := r.byte()
		if err != nil {
			return false, err
		}
		value = value || b&127 != 0
		if b&128 == 0 {
			return value, nil
		}
	}
	return false, decodeError("BOOLEAN_VARINT_OVERFLOW", start, "")
}
func (r *reader) take(size int) ([]byte, error) {
	if size < 0 || size > len(r.bytes)-r.position {
		return nil, decodeError("INVALID_LENGTH", r.position, fmt.Sprint(size))
	}
	value := r.bytes[r.position : r.position+size]
	r.position += size
	return value, nil
}
func (r *reader) tag(expected byte) error {
	at := r.position
	value, err := r.byte()
	if err != nil {
		return err
	}
	if value != expected {
		return decodeError("UNSUPPORTED_WIRE_TAG", at, fmt.Sprintf("expected %d, received %d", expected, value))
	}
	return nil
}

func (r *reader) atom() (Atom, error) {
	start := r.position
	atom := Atom{Address: start}
	outer, err := r.byte()
	if err != nil {
		return atom, err
	}
	atom.OuterTag = outer
	if outer != 26 && outer != 34 {
		return atom, decodeError("UNSUPPORTED_WIRE_TAG", start, fmt.Sprintf("atom outer tag %d", outer))
	}
	length, err := r.int32()
	if err != nil {
		return atom, err
	}
	begin := r.position
	atom.PayloadAddress = begin
	if length < 1 || length > len(r.bytes)-begin {
		return atom, decodeError("INVALID_ATOM_LENGTH", start, fmt.Sprint(length))
	}
	tag, err := r.byte()
	if err != nil {
		return atom, err
	}
	atom.TypeTag = tag
	switch tag {
	case 8:
		index, e := r.int32()
		if e != nil {
			return atom, e
		}
		if index < 0 {
			return atom, decodeError("UNSUPPORTED_NEGATIVE_REGISTER", start, "")
		}
		atom.Kind = "register"
		atom.Index = pointer(index)
	case 16:
		value, e := r.boolean()
		if e != nil {
			return atom, e
		}
		atom.Kind = "boolean"
		atom.Value = value
	case 24:
		value, e := r.int32()
		if e != nil {
			return atom, e
		}
		atom.Kind = "number"
		atom.Value = value
	case 34:
		size, e := r.int32()
		if e != nil {
			return atom, e
		}
		bytes, e := r.take(size)
		if e != nil {
			return atom, e
		}
		if !utf8.Valid(bytes) {
			return atom, decodeError("INVALID_UTF8_UNSUPPORTED", start, "original nonfatal decoder needs separate validation")
		}
		atom.Kind = "string"
		atom.Value = string(bytes)
	case 45:
		bytes, e := r.take(4)
		if e != nil {
			return atom, e
		}
		atom.Kind = "number"
		atom.Value = int(int32(binary.LittleEndian.Uint32(bytes)))
	case 49:
		bytes, e := r.take(8)
		if e != nil {
			return atom, e
		}
		value := math.Float64frombits(binary.LittleEndian.Uint64(bytes))
		atom.Kind = "number"
		switch {
		case math.IsNaN(value):
			atom.Value = "NaN"
		case math.IsInf(value, 1):
			atom.Value = "Infinity"
		case math.IsInf(value, -1):
			atom.Value = "-Infinity"
		case value == 0 && math.Signbit(value):
			atom.Value = "-0"
		default:
			atom.Value = value
		}
	default:
		return atom, decodeError("UNSUPPORTED_ATOM_TYPE", start, fmt.Sprint(tag))
	}
	if r.position != begin+length {
		return atom, decodeError("ATOM_LENGTH_MISMATCH", start, fmt.Sprintf("%d declared; %d consumed", length, r.position-begin))
	}
	atom.End = r.position
	atom.Raw = hex.EncodeToString(r.bytes[start:r.position])
	return atom, nil
}

type Instruction struct {
	Address       int                        `json:"address"`
	End           int                        `json:"end"`
	Opcode        int                        `json:"opcode"`
	Semantic      string                     `json:"semantic"`
	Mnemonic      string                     `json:"mnemonic"`
	Count         int                        `json:"count"`
	Destination   *int                       `json:"destination"`
	HeaderEnd     int                        `json:"headerEnd"`
	OperandsStart int                        `json:"operandsStart"`
	Operands      []Atom                     `json:"operands"`
	Raw           string                     `json:"raw"`
	Flow          map[string]any             `json:"flow"`
	Evidence      map[string]json.RawMessage `json:"evidence"`
	Target        *int                       `json:"target,omitempty"`
	Operation     *Operation                 `json:"operation,omitempty"`
	Inline        map[string]any             `json:"inline,omitempty"`
}
type UnknownRegion struct {
	Address    int        `json:"address"`
	End        int        `json:"end"`
	Raw        string     `json:"raw"`
	Diagnostic Diagnostic `json:"diagnostic"`
}
type InvalidTarget struct {
	Address    int    `json:"address"`
	Target     int    `json:"target"`
	Diagnostic string `json:"diagnostic"`
}
type Program struct {
	SchemaVersion int    `json:"schemaVersion"`
	VM            string `json:"vm"`
	Sample        Sample `json:"sample"`
	Profile       string `json:"profile"`
	Bytecode      struct {
		Bytes  int    `json:"bytes"`
		SHA256 string `json:"sha256"`
	} `json:"bytecode"`
	Instructions    []Instruction    `json:"instructions"`
	UnknownRegions  []UnknownRegion  `json:"unknownRegions"`
	InvalidTargets  []InvalidTarget  `json:"invalidTargets"`
	FunctionEntries []map[string]any `json:"functionEntries"`
	Status          string           `json:"status"`
	Coverage        map[string]any   `json:"coverage"`
}

func Decode(bytes []byte, configuration Configuration) (*Program, error) {
	if len(bytes) > 32*1024*1024 {
		return nil, fmt.Errorf("BYTECODE_TOO_LARGE")
	}
	handlers, err := configuration.validate()
	if err != nil {
		return nil, err
	}
	program := &Program{SchemaVersion: 2, VM: "register-stream", Sample: configuration.Sample, Profile: "register-stream-atoms-v1", Instructions: []Instruction{}, UnknownRegions: []UnknownRegion{}, InvalidTargets: []InvalidTarget{}}
	program.Bytecode.Bytes = len(bytes)
	sum := sha256.Sum256(bytes)
	program.Bytecode.SHA256 = hex.EncodeToString(sum[:])
	reader := reader{bytes: bytes}
	for reader.position < len(bytes) {
		start := reader.position
		instruction, err := decodeInstruction(&reader, handlers)
		if err != nil {
			diagnostic, ok := err.(*Diagnostic)
			if !ok {
				return nil, err
			}
			program.UnknownRegions = append(program.UnknownRegions, UnknownRegion{Address: start, End: len(bytes), Raw: hex.EncodeToString(bytes[start:]), Diagnostic: *diagnostic})
			break
		}
		program.Instructions = append(program.Instructions, instruction)
	}
	boundaries := map[int]bool{len(bytes): true}
	for _, instruction := range program.Instructions {
		boundaries[instruction.Address] = true
	}
	program.FunctionEntries = []map[string]any{{"address": 0, "reason": "supplied stream start; original function identity unknown"}}
	decodedBytes := 0
	for _, instruction := range program.Instructions {
		decodedBytes = instruction.End
		if instruction.Target != nil {
			if !boundaries[*instruction.Target] {
				program.InvalidTargets = append(program.InvalidTargets, InvalidTarget{Address: instruction.Address, Target: *instruction.Target, Diagnostic: "TARGET_NOT_INSTRUCTION_BOUNDARY"})
			}
			if instruction.Semantic == "MAKE_CLOSURE" || instruction.Semantic == "SCHEDULE" {
				program.FunctionEntries = append(program.FunctionEntries, map[string]any{"address": *instruction.Target, "reason": instruction.Semantic, "createdAt": instruction.Address})
			}
		}
	}
	program.Status = "decoded"
	if len(program.UnknownRegions) > 0 || len(program.InvalidTargets) > 0 {
		program.Status = "partial"
	}
	program.Coverage = map[string]any{"decodedInstructions": len(program.Instructions), "decodedBytes": decodedBytes, "supportedHandlerSlots": len(handlers), "configurationTrust": "Caller-supplied semantic configuration; native Go extraction parity is still required."}
	return program, nil
}

func decodeInstruction(reader *reader, handlers map[int]Handler) (Instruction, error) {
	instruction := Instruction{Address: reader.position, Operands: []Atom{}}
	count, err := reader.byte()
	if err != nil {
		return instruction, err
	}
	instruction.Count = int(count)
	opcode, err := reader.int32()
	if err != nil {
		return instruction, err
	}
	instruction.Opcode = opcode
	instruction.HeaderEnd = reader.position
	entry, ok := handlers[opcode]
	if !ok {
		return instruction, decodeError("UNKNOWN_OPCODE", instruction.Address, fmt.Sprint(opcode))
	}
	layout := entry.Layout
	if int(count) < layout.Min || int(count) > layout.Max {
		return instruction, decodeError("INVALID_OPERAND_COUNT", instruction.Address, entry.Semantic)
	}
	if layout.Destination {
		if err := reader.tag(16); err != nil {
			return instruction, err
		}
		destination, err := reader.int32()
		if err != nil {
			return instruction, err
		}
		if destination < 0 {
			return instruction, decodeError("NEGATIVE_DESTINATION", instruction.Address, "")
		}
		instruction.Destination = pointer(destination)
	}
	instruction.OperandsStart = reader.position
	for index := 0; index < int(count); index++ {
		operand, err := reader.atom()
		if err != nil {
			return instruction, err
		}
		if (layout.AllRegisters || slices.Contains(layout.RegisterOperands, index)) && operand.Kind != "register" {
			return instruction, decodeError("EXPECTED_REGISTER_ATOM", operand.Address, fmt.Sprintf("%s operand %d", entry.Semantic, index))
		}
		if slices.Contains(layout.IndexOperands, index) {
			if operand.TypeTag != 8 && operand.TypeTag != 24 {
				return instruction, decodeError("EXPECTED_RAW_INT32_INDEX", operand.Address, entry.Semantic)
			}
			if operand.Index != nil {
				operand.IndexValue = pointer(*operand.Index)
			} else {
				operand.IndexValue = pointer(operand.Value.(int))
			}
			operand.ReadMode = "raw-int32-index"
		}
		if layout.Offset != nil && *layout.Offset == index && operand.TypeTag != 45 {
			return instruction, decodeError("EXPECTED_FIXED32_OFFSET", operand.Address, "")
		}
		instruction.Operands = append(instruction.Operands, operand)
	}
	instruction.End = reader.position
	instruction.Semantic = entry.Semantic
	instruction.Mnemonic = entry.Semantic
	instruction.Raw = hex.EncodeToString(reader.bytes[instruction.Address:reader.position])
	instruction.Flow = map[string]any{"kind": "fallthrough", "next": reader.position}
	implementation := entry.Evidence
	if entry.Delegation != nil {
		implementation = entry.Delegation.Evidence
	}
	instruction.Evidence = map[string]json.RawMessage{"handler": entry.Evidence, "implementation": implementation}
	if entry.Semantic == "BINARY_MAP_LEFT" {
		instruction.Operation = entry.Operation
		if entry.Operator != nil {
			instruction.Evidence["operator"] = entry.Operator.Evidence
		}
		names := map[string]string{"^": "XOR", "+": "ADD", "-": "SUB", "*": "MUL", "/": "DIV", "%": "REM", "&&": "AND", "||": "OR"}
		instruction.Mnemonic = names[entry.Operation.Operator] + "_MAP_LEFT"
		if entry.Operation.Left.Kind == "not" {
			instruction.Mnemonic = "BOOL_" + instruction.Mnemonic
		}
	}
	if layout.Offset != nil {
		target := instruction.End + instruction.Operands[*layout.Offset].Value.(int)
		if target < 0 || target > len(reader.bytes) {
			return instruction, decodeError("TARGET_OUT_OF_BOUNDS", instruction.Address, fmt.Sprint(target))
		}
		instruction.Target = pointer(target)
		switch entry.Semantic {
		case "BRANCH_EQ", "BRANCH_LT":
			instruction.Flow = map[string]any{"kind": "branch", "taken": target, "fallthrough": instruction.End}
		case "MAKE_CLOSURE":
			instruction.Flow["closureEntry"] = target
		default:
			instruction.Flow["scheduledEntry"] = target
		}
	}
	if entry.Semantic == "YIELD" {
		instruction.Flow = map[string]any{"kind": "yield", "resume": instruction.End, "exit": len(reader.bytes)}
	}
	if entry.Semantic == "MAP_HANDLER" {
		var nested any
		if instruction.Operands[0].Kind == "number" {
			nested = instruction.Operands[0].Value
		}
		instruction.Inline = map[string]any{"opcode": nested, "destinationRegister": *instruction.Operands[2].IndexValue, "elementRegister": *instruction.Operands[1].IndexValue, "operands": instruction.Operands[3:], "encodingNote": "Third atom payload is reused as the nested destination header; remaining atoms form the nested operand stream."}
	}
	return instruction, nil
}
