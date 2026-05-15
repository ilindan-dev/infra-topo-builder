// Tests for parsing helper functions in the utils package. These ensure
// correct numeric and string parsing, trimming, nullable handling, and support
// for hexadecimal integer formats where applicable.
package utils

import (
	"reflect"
	"testing"
)

// TestParseInt32 verifies ParseInt32 handles valid numbers, whitespace,
// empty and invalid inputs by returning 0 for missing/invalid values.
func TestParseInt32(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int32
	}{
		{"Valid positive", "42", 42},
		{"Valid negative", "-10", -10},
		{"With spaces", "  15  ", 15},
		{"Empty string", "", 0},
		{"Invalid chars", "abc", 0},
		{"N/A string", "N/A", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseInt32(tt.input); got != tt.want {
				t.Errorf("parseInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParsePtrInt32 ensures ParsePtrInt32 returns pointers for valid numbers
// and nil for empty, "N/A", or unparsable strings.
func TestParsePtrInt32(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *int32
	}{
		{"Valid positive", "42", new(int32(42))},
		{"Valid negative", "-5", new(int32(-5))},
		{"With spaces", " 42 ", new(int32(42))},
		{"Empty string", "", nil},
		{"N/A string", "N/A", nil},
		{"Invalid string", "invalid", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePtrInt32(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePtrInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParsePtrInt64 verifies decimal and hexadecimal parsing and that
// invalid or empty inputs return nil.
func TestParsePtrInt64(t *testing.T) {
	valHex := int64(43981)

	tests := []struct {
		name  string
		input string
		want  *int64
	}{
		{"Valid decimal", "100", new(int64(100))},
		{"Valid hex lowercase", "0xabcd", &valHex},
		{"Valid hex uppercase", "0xABCD", &valHex},
		{"With spaces", " 0xabcd ", &valHex},
		{"Empty string", "", nil},
		{"N/A string", "N/A", nil},
		{"Invalid string", "invalid", nil},
		{"Invalid hex", "0xGHIJ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePtrInt64(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePtrInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParsePtrString checks that ParsePtrString trims whitespace and
// treats empty or "N/A" values as nil, returning pointers for valid text.
func TestParsePtrString(t *testing.T) {
	valStr := "Mellanox"

	tests := []struct {
		name  string
		input string
		want  *string
	}{
		{"Valid string", "Mellanox", &valStr},
		{"Valid with spaces", "  Mellanox  ", &valStr},
		{"Empty string", "", nil},
		{"N/A string", "N/A", nil},
		{"Spaces only", "   ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePtrString(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePtrString() = %v, want %v", got, tt.want)
			}
		})
	}
}
