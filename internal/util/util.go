package util

import (
	"fmt"
	"sort"
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

// JoinSorted takes map of keys and values, sorts them by keys and joins with given separators.
func JoinSorted(values map[string]string, valueSeparator string, keySeparator string) string {
	keys := []string{}

	for k := range values {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	v := []string{}

	for _, k := range keys {
		v = append(v, fmt.Sprintf("%s%s%s", k, valueSeparator, values[k]))
	}

	return strings.Join(v, keySeparator)
}
