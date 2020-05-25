// Package util provides very generic helpers used across codebase.
package util

import (
	"fmt"
	"sort"
	"strings"

	"github.com/logrusorgru/aurora"
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

// PickStringSlice returns first non-empty strings slice.
func PickStringSlice(values ...[]string) []string {
	for _, v := range values {
		if len(v) > 0 {
			return v
		}
	}

	return []string{}
}

// PickStringMap returns first non-empty map of strings.
func PickStringMap(values ...map[string]string) map[string]string {
	for _, v := range values {
		if len(v) > 0 {
			return v
		}
	}

	return map[string]string{}
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
	if text == "" {
		return ""
	}

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

// KeysStringMap returns keys from given map.
func KeysStringMap(m map[string]string) []string {
	keys := []string{}

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// ColorizeDiff takes diff-formatter output and adds console colors to it.
func ColorizeDiff(diff string) string {
	// Don't even try to process empty strings.
	if diff == "" {
		return diff
	}

	// If string ends with newline, strip it before splitting, then we add it at the end of processing.
	endsWithNewLine := diff[len(diff)-1] == '\n'
	if endsWithNewLine {
		diff = diff[:len(diff)-1]
	}

	lines := strings.Split(diff, "\n")
	l := len(lines)

	output := ""

	for i, line := range strings.Split(diff, "\n") {
		nl := line + "\n"

		// If we process last line and the given diff does not end with newline, don't include it.
		if !endsWithNewLine && i == l-1 {
			nl = line
		}

		if len(line) > 0 && line[0] == '-' {
			nl = aurora.Red(line + "\n").String()
		}

		if len(line) > 0 && line[0] == '+' {
			nl = aurora.Green(line + "\n").String()
		}

		output += nl
	}

	return output
}
