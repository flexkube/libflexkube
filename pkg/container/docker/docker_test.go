package docker

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
)

// New()
func TestContainerNew(t *testing.T) {
	_, err := New()
	if err != nil {
		t.Errorf("Creating new docker client should work, got: %s", err)
	}
}

// Start()
func TestContainerStartValidateUninitialized(t *testing.T) {
	d := &Docker{}
	if err := d.Start("foo"); err == nil {
		t.Errorf("Starting container with uninitilized client should fail")
	}
}

// Create()
func TestContainerCreateValidateUninitialized(t *testing.T) {
	d := &Docker{}
	if _, err := d.Create(&container.Config{}); err == nil {
		t.Errorf("Starting container with uninitilized client should fail")
	}
}

// Status()
func TestContainerStatusValidateUninitialized(t *testing.T) {
	d := &Docker{}
	if _, err := d.Status("foo"); err == nil {
		t.Errorf("Getting status of container with uninitilized client should fail")
	}
}

// Delete()
func TestContainerDeleteValidateUninitialized(t *testing.T) {
	d := &Docker{}
	if err := d.Delete("foo"); err == nil {
		t.Errorf("Removing container with uninitilized client should fail")
	}
}

// validate()
func TestContainerValidateUninitialized(t *testing.T) {
	d := &Docker{}
	if err := d.validate(); err == nil {
		t.Errorf("Uninitialized struct should not be valid")
	}
}

func TestContainerValidate(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Errorf("Creating new docker client should work, got: %s", err)
	}
	if err := d.validate(); err != nil {
		t.Errorf("Initialized struct should be valid, got: %s", err)
	}
}
