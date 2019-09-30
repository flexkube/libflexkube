package container

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container/types"
)

// New()
func TestNewEmptyConfiguration(t *testing.T) {
	if _, err := New(&Container{}); err == nil {
		t.Errorf("Creating container with wrong configuration should fail")
	}
}

func TestNewGoodConfiguration(t *testing.T) {
	c := &Container{
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "nonexistent",
		},
	}
	if _, err := New(c); err != nil {
		t.Errorf("Creating container with good configuration should pass, got: %v", err)
	}
}

// Validate()
func TestValidateNoName(t *testing.T) {
	c := &Container{
		Config: types.ContainerConfig{},
	}
	if err := c.Validate(); err == nil {
		t.Errorf("Validating container without name should fail")
	}
}

func TestValidate(t *testing.T) {
	c := &Container{
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "nonexistent",
		},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("Validating container with valid configuration should pass, got: %v", err)
	}
}

func TestValidateUnsupportedRuntime(t *testing.T) {
	c := &Container{
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "nonexistent",
		},
	}
	if err := c.Validate(); err == nil {
		t.Errorf("Validating container with unsupported container runtime should fail")
	}
}

func TestValidateRequireImage(t *testing.T) {
	c := &Container{
		Config: types.ContainerConfig{
			Name: "foo",
		},
	}
	if err := c.Validate(); err == nil {
		t.Errorf("Validating container with no image set should fail")
	}
}

// selectRuntime()
func TestSelectDockerRuntime(t *testing.T) {
	c := &container{
		base{},
		&types.ContainerStatus{},
	}
	if err := c.selectRuntime(); err != nil {
		t.Errorf("Selecting Docker container runtime should succeed, got: %v", err)
	}
	if c.runtime == nil {
		t.Errorf("Selecting container runtime should set container runtime field")
	}
}

// FromStatus()
func TestFromStatusValid(t *testing.T) {
	c := &container{
		base{},
		&types.ContainerStatus{
			ID: "nonexistent",
		},
	}
	if _, err := c.FromStatus(); err != nil {
		t.Fatalf("Container instance should be created from valid container, got: %v", err)
	}
}

func TestFromStatusNoID(t *testing.T) {
	c := &container{
		base{},
		&types.ContainerStatus{},
	}
	if _, err := c.FromStatus(); err == nil {
		t.Fatalf("Container instance should not be created from container with no container ID")
	}
}

func TestFromStatusNoStatus(t *testing.T) {
	c := &container{
		base{},
		nil,
	}
	if _, err := c.FromStatus(); err == nil {
		t.Fatalf("Container instance should not be created from container without status")
	}
}