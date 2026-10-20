package deobfuscate

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/t14raptor/go-fast/token"
)

// Strings use JavaScript code units; indexing Go UTF-8 would change decoder keys.
type jsString []uint16
type undefined struct{}

func stringValue(s string) jsString {
	if !utf8.ValidString(s) {
		fail("INVALID_STATIC_STRING_UTF8", nil, "parser value")
	}
	return utf16.Encode([]rune(s))
}

func (s jsString) utf8() (string, error) {
	for i := 0; i < len(s); i++ {
		unit := s[i]
		if unit >= 0xd800 && unit <= 0xdbff {
			if i+1 >= len(s) || s[i+1] < 0xdc00 || s[i+1] > 0xdfff {
				return "", fmt.Errorf("UNPAIRED_STATIC_SURROGATE: unit %d", i)
			}
			i++
		} else if unit >= 0xdc00 && unit <= 0xdfff {
			return "", fmt.Errorf("UNPAIRED_STATIC_SURROGATE: unit %d", i)
		}
	}
	return string(utf16.Decode(s)), nil
}

func primitive(v any) bool {
	switch v.(type) {
	case undefined, nil, bool, float64, jsString:
		return true
	}
	return false
}

func truthy(v any) bool {
	switch x := v.(type) {
	case undefined, nil:
		return false
	case bool:
		return x
	case float64:
		return x != 0 && !math.IsNaN(x)
	case jsString:
		return len(x) != 0
	default:
		return true
	}
}

func number(v any) float64 {
	switch x := v.(type) {
	case undefined:
		return math.NaN()
	case nil:
		return 0
	case bool:
		if x {
			return 1
		}
		return 0
	case float64:
		return x
	case jsString:
		s, err := x.utf8()
		if err != nil {
			return math.NaN()
		}
		s = strings.TrimFunc(s, jsWhitespace)
		if s == "" {
			return 0
		}
		if s == "Infinity" || s == "+Infinity" {
			return math.Inf(1)
		}
		if s == "-Infinity" {
			return math.Inf(-1)
		}
		// ParseFloat accepts forms outside StringNumericLiteral (Inf, hex floats).
		if strings.ContainsAny(s, "pP_") || strings.Contains(s, "Inf") {
			return math.NaN()
		}
		if len(s) > 2 && s[0] == '0' {
			base := 0
			switch s[1] {
			case 'x', 'X':
				base = 16
			case 'b', 'B':
				base = 2
			case 'o', 'O':
				base = 8
			}
			if base != 0 {
				integer, ok := new(big.Int).SetString(s[2:], base)
				if !ok || strings.ContainsAny(s[2:], "+-_") {
					return math.NaN()
				}
				n, _ := new(big.Float).SetInt(integer).Float64()
				return n
			}
		}
		for _, c := range s {
			if !(c >= '0' && c <= '9') && c != '.' && c != 'e' && c != 'E' && c != '+' && c != '-' {
				return math.NaN()
			}
		}
		n, err := strconv.ParseFloat(s, 64)
		if err != nil && !math.IsInf(n, 0) {
			return math.NaN()
		}
		return n
	}
	fail("NON_PRIMITIVE_STATIC_COERCION", nil, fmt.Sprintf("%T", v))
	return 0
}

func jsWhitespace(r rune) bool {
	return r == '\u0009' || r == '\u000b' || r == '\u000c' || r == '\u0020' || r == '\u00a0' || r == '\ufeff' || r == '\u000a' || r == '\u000d' || r == '\u2028' || r == '\u2029' || r == '\u1680' || (r >= '\u2000' && r <= '\u200a') || r == '\u202f' || r == '\u205f' || r == '\u3000'
}

func uint32Value(v any) uint32 {
	n := number(v)
	if n == 0 || math.IsInf(n, 0) || math.IsNaN(n) {
		return 0
	}
	n = math.Mod(math.Trunc(n), 4294967296)
	if n < 0 {
		n += 4294967296
	}
	return uint32(n)
}

func numberText(n float64) string {
	if math.IsNaN(n) {
		return "NaN"
	}
	if math.IsInf(n, 1) {
		return "Infinity"
	}
	if math.IsInf(n, -1) {
		return "-Infinity"
	}
	if n == 0 {
		return "0"
	}
	// ECMAScript uses fixed notation for [1e-6, 1e21).
	format := byte('e')
	if a := math.Abs(n); a >= 1e-6 && a < 1e21 {
		format = 'f'
	}
	s := strconv.FormatFloat(n, format, -1, 64)
	if i := strings.IndexByte(s, 'e'); i >= 0 {
		exp, _ := strconv.Atoi(s[i+1:])
		s = s[:i] + "e"
		if exp >= 0 {
			s += "+"
		}
		s += strconv.Itoa(exp)
	}
	return s
}

func textValue(v any) jsString {
	switch x := v.(type) {
	case jsString:
		return x
	case undefined:
		return stringValue("undefined")
	case nil:
		return stringValue("null")
	case bool:
		return stringValue(strconv.FormatBool(x))
	case float64:
		return stringValue(numberText(x))
	}
	fail("NON_PRIMITIVE_STATIC_COERCION", nil, fmt.Sprintf("%T", v))
	return nil
}

func equalString(a, b jsString) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func strictEqual(a, b any) bool {
	switch x := a.(type) {
	case undefined:
		_, ok := b.(undefined)
		return ok
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case float64:
		y, ok := b.(float64)
		return ok && x == y
	case jsString:
		y, ok := b.(jsString)
		return ok && equalString(x, y)
	}
	fail("NON_PRIMITIVE_STATIC_OPERATION", nil, "equality")
	return false
}

func looseEqual(a, b any) bool {
	if strictEqual(a, b) {
		return true
	}
	_, au := a.(undefined)
	_, bu := b.(undefined)
	if (au && b == nil) || (bu && a == nil) {
		return true
	}
	if au || bu || a == nil || b == nil {
		return false
	}
	if _, ok := a.(bool); ok {
		return looseEqual(number(a), b)
	}
	if _, ok := b.(bool); ok {
		return looseEqual(a, number(b))
	}
	_, an := a.(float64)
	_, bn := b.(float64)
	_, as := a.(jsString)
	_, bs := b.(jsString)
	return ((an && bs) || (bn && as)) && number(a) == number(b)
}

func less(a, b any) bool {
	x, xs := a.(jsString)
	y, ys := b.(jsString)
	if xs && ys {
		for i := 0; i < len(x) && i < len(y); i++ {
			if x[i] != y[i] {
				return x[i] < y[i]
			}
		}
		return len(x) < len(y)
	}
	return number(a) < number(b)
}

func binary(op token.Token, a, b any) any {
	if !primitive(a) || !primitive(b) {
		fail("NON_PRIMITIVE_STATIC_OPERATION", nil, op.String())
	}
	switch op {
	case token.Plus:
		_, as := a.(jsString)
		_, bs := b.(jsString)
		if as || bs {
			x, y := textValue(a), textValue(b)
			return append(append(jsString{}, x...), y...)
		}
		return number(a) + number(b)
	case token.Minus:
		return number(a) - number(b)
	case token.Multiply:
		return number(a) * number(b)
	case token.Slash:
		return number(a) / number(b)
	case token.Remainder:
		return math.Mod(number(a), number(b))
	case token.Exponent:
		x, y := number(a), number(b)
		if y == 0 {
			return float64(1)
		}
		if math.Abs(x) == 1 && math.IsInf(y, 0) {
			return math.NaN()
		}
		return math.Pow(x, y)
	case token.ExclusiveOr:
		return float64(int32(uint32Value(a) ^ uint32Value(b)))
	case token.And:
		return float64(int32(uint32Value(a) & uint32Value(b)))
	case token.Or:
		return float64(int32(uint32Value(a) | uint32Value(b)))
	case token.ShiftLeft:
		return float64(int32(uint32Value(a) << (uint32Value(b) & 31)))
	case token.ShiftRight:
		return float64(int32(uint32Value(a)) >> (uint32Value(b) & 31))
	case token.UnsignedShiftRight:
		return float64(uint32Value(a) >> (uint32Value(b) & 31))
	case token.StrictEqual:
		return strictEqual(a, b)
	case token.StrictNotEqual:
		return !strictEqual(a, b)
	case token.Equal:
		return looseEqual(a, b)
	case token.NotEqual:
		return !looseEqual(a, b)
	case token.Less:
		return less(a, b)
	case token.Greater:
		return less(b, a)
	case token.LessOrEqual:
		if _, ok := a.(jsString); ok {
			if _, ok := b.(jsString); ok {
				return !less(b, a)
			}
		}
		return number(a) <= number(b)
	case token.GreaterOrEqual:
		if _, ok := a.(jsString); ok {
			if _, ok := b.(jsString); ok {
				return !less(a, b)
			}
		}
		return number(a) >= number(b)
	}
	fail("UNSUPPORTED_STATIC_BINARY", nil, op.String())
	return nil
}

func unary(op token.Token, v any) any {
	if !primitive(v) {
		fail("NON_PRIMITIVE_STATIC_OPERATION", nil, op.String())
	}
	switch op {
	case token.Plus:
		return number(v)
	case token.Minus:
		return -number(v)
	case token.Not:
		return !truthy(v)
	case token.BitwiseNot:
		return float64(int32(^uint32Value(v)))
	case token.Void:
		return undefined{}
	case token.Typeof:
		s := "undefined"
		switch v.(type) {
		case nil:
			s = "object"
		case bool:
			s = "boolean"
		case float64:
			s = "number"
		case jsString:
			s = "string"
		}
		return stringValue(s)
	}
	fail("UNSUPPORTED_STATIC_UNARY", nil, op.String())
	return nil
}

// quoteUnits emits valid JavaScript, including lone UTF-16 surrogates.
func quoteUnits(s jsString) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, u := range s {
		if u >= 32 && u < 127 && u != '"' && u != '\\' {
			b.WriteByte(byte(u))
		} else {
			fmt.Fprintf(&b, "\\u%04x", u)
		}
	}
	b.WriteByte('"')
	return b.String()
}
