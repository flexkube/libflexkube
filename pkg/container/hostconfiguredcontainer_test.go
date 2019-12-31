package container

import (
	"testing"
)

// withHook()
func TestWithHook(t *testing.T) {
	action := false

	if err := withHook(nil, func() error {
		action = true

		return nil
	}, nil); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !action {
		t.Fatalf("withHook should execute action")
	}
}

func TestWithPreHook(t *testing.T) {
	pre := false

	f := Hook(func() error {
		pre = true

		return nil
	})

	if err := withHook(&f, func() error {
		return nil
	}, nil); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !pre {
		t.Fatalf("withHook should call pre-hook")
	}
}

func TestWithPostHook(t *testing.T) {
	post := false

	f := Hook(func() error {
		post = true

		return nil
	})

	if err := withHook(nil, func() error {
		return nil
	}, &f); err != nil {
		t.Fatalf("withHook should not return error, got: %v", err)
	}

	if !post {
		t.Fatalf("withHook should call post-hook")
	}
}
