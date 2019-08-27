package util

import (
	"testing"
)

// DefaultString
func TestDefaultString(t *testing.T) {
	r := "foo"
	if n := DefaultString(r, "bar"); n != r {
		t.Fatalf("'%s' should be returned, got '%s'", r, n)
	}
}

func TestDefaultStringDefault(t *testing.T) {
	d := "foo"
	if n := DefaultString("", d); n != d {
		t.Fatalf("'%s' should be returned, got '%s'", d, n)
	}
}
