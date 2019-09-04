package runtime

import (
	"testing"
)

// Register()
func TestRegisterOK(t *testing.T) {
	// Make sure we start from clean state
	runtimes = make(map[string]bool)
	n := "foo"
	Register(n)
	if _, exists := runtimes[n]; !exists {
		t.Fatalf("Registering runtime should add it to runtimes")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	// Make sure we start from clean state
	runtimes = make(map[string]bool)
	n := "foo"
	Register(n)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Registering duplicated runtime should panic")
		}
	}()
	Register(n)
}

// IsRuntimeSupported()
func TestRuntimeSupported(t *testing.T) {
	// Make sure we start from clean state
	runtimes = make(map[string]bool)
	n := "foo"
	Register(n)

	if !IsRuntimeSupported(n) {
		t.Fatalf("Registered runtime should be supported")
	}
}

func TestRuntimeNotSupported(t *testing.T) {
	// Make sure we start from clean state
	runtimes = make(map[string]bool)
	if IsRuntimeSupported("foo") {
		t.Fatalf("Registered runtime should be supported")
	}
}

func TestRuntimeDefault(t *testing.T) {
	// Make sure we start from clean state
	runtimes = make(map[string]bool)
	Register(defaultRuntime)
	if !IsRuntimeSupported("") {
		t.Fatalf("Default runtime should be supported")
	}
}

// GetRuntimeName()
func TestGetRuntimeName(t *testing.T) {
	r := "foo"
	if n := GetRuntimeName(r); n != r {
		t.Fatalf("'%s' should be returned, got '%s'", r, n)
	}
}

func TestGetRuntimeNameDefault(t *testing.T) {
	if n := GetRuntimeName(""); n != defaultRuntime {
		t.Fatalf("'%s' should be returned, got '%s'", defaultRuntime, n)
	}
}
