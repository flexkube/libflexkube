package util

import (
	"strings"
)

// PickString returns first non-empty string passed.
func PickString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

// PickInt returns first non-zero integer passed.
func PickInt(values ...int) int {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}

	return 0
}

// Indent indents a block of text with an indent string.
func Indent(text, indent string) string {
	if text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			result += indent + j + "\n"
		}

		return result
	}

	result := ""

	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}

	return result[:len(result)-1]
}
