package docker

import "testing"

// New()
func TestContainerNew(t *testing.T) {
	_, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker client should work, got: %s", err)
	}
}
