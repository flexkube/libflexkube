package container

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container/runtime"
)

// New()
func TestNewEmptyConfiguration(t *testing.T) {
	if _, err := New(&Container{}); err == nil {
		t.Errorf("Creating container with wrong configuration should fail")
	}
}

func TestNewGoodConfiguration(t *testing.T) {
	c := &Container{
		Config: runtime.Config{
			Name: "foo",
		},
	}
	if _, err := New(c); err != nil {
		t.Errorf("Creating container with good configuration should pass, got: %v", err)
	}
}

func TestNewValidateRuntime(t *testing.T) {
	c := &Container{
		Config: runtime.Config{
			Name: "foo",
		},
		RuntimeName: "doh",
	}
	if _, err := New(c); err == nil {
		t.Errorf("Creating container should validate container runtime configuration")
	}
}

// Validate()
func TestValidateNoName(t *testing.T) {
	c := &Container{
		Config: runtime.Config{},
	}
	if err := c.Validate(); err == nil {
		t.Errorf("Validating container without name should fail")
	}
}

func TestValidate(t *testing.T) {
	c := &Container{
		Config: runtime.Config{
			Name: "foo",
		},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("Validating container with valid configuration should pass, got: %v", err)
	}
}

func TestValidateUnsupportedRuntime(t *testing.T) {
	c := &Container{
		Config: runtime.Config{
			Name: "foo",
		},
		RuntimeName: "foo",
	}
	if err := c.Validate(); err == nil {
		t.Errorf("Validating container with unsupported container runtime should fail")
	}
}

// selectRuntime()
func TestSelectDockerRuntime(t *testing.T) {
	c := &container{
		base{
			runtimeName: "docker",
		},
		&runtime.Status{},
	}
	if err := c.selectRuntime(); err != nil {
		t.Errorf("Selecting Docker container runtime should succeed, got: %v", err)
	}
	if c.runtime == nil {
		t.Errorf("Selecting container runtime should set container runtime field")
	}
}

func TestSelectDefaultRuntime(t *testing.T) {
	c := &container{}
	if err := c.selectRuntime(); err != nil {
		t.Errorf("Selecting container runtime on empty container should succeed, got: %v", err)
	}
	if c.runtime == nil {
		t.Errorf("Selecting container runtime should set container runtime field")
	}
}

func TestSelectBadRuntime(t *testing.T) {
	c := &container{
		base{
			runtimeName: "foo",
		},
		&runtime.Status{},
	}
	if err := c.selectRuntime(); err == nil {
		t.Errorf("Unsupported container runtime name should be rejected")
	}
}

// FromStatus()
func TestFromStatusValid(t *testing.T) {
	c := &container{
		base{
			runtimeName: "docker",
		},
		&runtime.Status{
			ID: "nonexistent",
		},
	}
	if _, err := c.FromStatus(); err != nil {
		t.Fatalf("Container instance should be created from valid container, got: %v", err)
	}
}

func TestFromStatusNoID(t *testing.T) {
	c := &container{
		base{
			runtimeName: "docker",
		},
		&runtime.Status{},
	}
	if _, err := c.FromStatus(); err == nil {
		t.Fatalf("Container instance should not be created from container with no container ID")
	}
}

func TestFromStatusNoStatus(t *testing.T) {
	c := &container{
		base{
			runtimeName: "docker",
		},
		nil,
	}
	if _, err := c.FromStatus(); err == nil {
		t.Fatalf("Container instance should not be created from container without status")
	}
}
