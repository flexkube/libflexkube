package util

import (
	"testing"
)

const expectedValueString = "foo"
const expectedValueInt = 10

func TestPickStringLast(t *testing.T) {
	if v := PickString("", "", expectedValueString); v != expectedValueString {
		t.Fatalf("expected %s, got %s", expectedValueString, v)
	}
}

func TestPickStringNoValue(t *testing.T) {
	if v := PickString(""); v != "" {
		t.Fatalf("expected '%s', got '%s'", "", v)
	}
}

func TestPickStringFirst(t *testing.T) {
	if v := PickString(expectedValueString, "bar"); v != expectedValueString {
		t.Fatalf("expected %s, got %s", expectedValueString, v)
	}
}

func TestPickIntLast(t *testing.T) {
	if v := PickInt(0, 0, expectedValueInt); v != expectedValueInt {
		t.Fatalf("expected %d, got %d", expectedValueInt, v)
	}
}

func TestPickIntNoValue(t *testing.T) {
	if v := PickInt(0); v != 0 {
		t.Fatalf("expected %d, got %d", 0, v)
	}
}

func TestPickIntFirst(t *testing.T) {
	if v := PickInt(expectedValueInt, 5); v != expectedValueInt {
		t.Fatalf("expected %d, got %d", expectedValueInt, v)
	}
}

func TestIndent(t *testing.T) {
	expected := "   foo"
	if a := Indent("foo", "   "); a != expected {
		t.Fatalf("expected '%s', got '%s'", expected, a)
	}
}

func TestIndentWithNewline(t *testing.T) {
	expected := "  foo\n  bar\n"
	if a := Indent("foo\nbar\n", "  "); a != expected {
		t.Fatalf("expected '%s', got '%s'", expected, a)
	}
}

func TestIndentEmpty(t *testing.T) {
	expected := ""
	if a := Indent("", ""); a != expected {
		t.Fatalf("expected '%s', got '%s'", expected, a)
	}
}

func TestIndentEmptyText(t *testing.T) {
	expected := ""
	if a := Indent("", "  "); a != expected {
		t.Fatalf("expected '%s', got '%s'", expected, a)
	}
}

func TestIndentEmptyIndent(t *testing.T) {
	expected := "foo\nbar"
	if a := Indent("foo\nbar", ""); a != expected {
		t.Fatalf("expected '%s', got '%s'", expected, a)
	}
}

func TestJoinSorted(t *testing.T) {
	expected := "baz/doh|foo/bar"

	values := map[string]string{
		"foo": "bar",
		"baz": "doh",
	}

	if a := JoinSorted(values, "/", "|"); a != expected {
		t.Fatalf("expected '%s', got '%s'", expected, a)
	}
}
