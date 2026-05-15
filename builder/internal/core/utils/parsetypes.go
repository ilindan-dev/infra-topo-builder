package utils

import (
	"strconv"
	"strings"
)

// ParseInt32 parses s into an int32 value, trimming surrounding whitespace.
// It returns 0 if s is empty or cannot be parsed. Use when missing or invalid
// numeric input should be treated as zero.
func ParseInt32(s string) int32 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0
	}
	return int32(v)
}

// ParsePtrInt32 parses s into a pointer to int32.
// It returns nil when s is empty, equals "N/A", or cannot be parsed.
// This is useful for representing optional numeric CSV/database fields as nullable values.
func ParsePtrInt32(s string) *int32 {
	s = strings.TrimSpace(s)
	if s == "" || s == "N/A" {
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil
	}
	return new(int32(v))
}

// ParsePtrInt64 parses s into a pointer to int64.
// Supports decimal and hexadecimal strings (hex must use the "0x" prefix).
// Returns nil for empty or "N/A" inputs or on parse failure.
func ParsePtrInt64(s string) *int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "N/A" {
		return nil
	}

	var v int64
	var err error

	if strings.HasPrefix(s, "0x") {
		v, err = strconv.ParseInt(s[2:], 16, 64)
	} else {
		v, err = strconv.ParseInt(s, 10, 64)
	}

	if err != nil {
		return nil
	}
	return &v
}

// ParsePtrString returns a pointer to the trimmed string value.
// It returns nil for empty or "N/A" inputs so optional text fields can be
// represented as NULL in databases or domain objects.
func ParsePtrString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" || s == "N/A" {
		return nil
	}
	return &s
}
