package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// field extracts a trimmed substring from a fixed-width EMAF line (start inclusive, end exclusive).
func field(line string, start, end int) string {
	runes := []rune(line)
	if start >= len(runes) {
		return ""
	}
	if end > len(runes) {
		end = len(runes)
	}
	return strings.TrimSpace(string(runes[start:end]))
}

// RecordIdentifier returns the record type name for a line (positions 9-12).
func RecordIdentifier(line string) string {
	return recordTypeMapper(field(line, 9, 12))
}

// convertAmount inserts a decimal point at position 9 and formats to 2 decimal places.
// Matches the JS: splice(9, 0, '.') -> Number(...).toFixed(2)
func convertAmount(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "0.00"
	}
	var raw string
	if len(s) > 9 {
		raw = s[:9] + "." + s[9:]
	} else {
		raw = s + "."
	}
	n, _ := strconv.ParseFloat(raw, 64)
	return strconv.FormatFloat(n, 'f', 2, 64)
}

// convertInterchangeFee inserts decimal at position 5 and formats to 9 decimal places.
// Matches the JS: splice(5, 0, '.') -> Number(...).toFixed(9)
func convertInterchangeFee(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "0.000000000"
	}
	var raw string
	if len(s) > 5 {
		raw = s[:5] + "." + s[5:]
	} else {
		raw = s + "."
	}
	n, _ := strconv.ParseFloat(raw, 64)
	return fmt.Sprintf("%.9f", n)
}
