// +build integration

package docker

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
	"github.com/invidian/etcd-ariadnes-thread/pkg/defaults"
)

// Create
func TestContainerCreate(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &container.Config{
		Image: defaults.EtcdImage,
	}

	if _, err = d.Create(c); err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}
}

func TestContainerCreateDelete(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &container.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}

	if err := d.Delete(id); err != nil {
		t.Errorf("Removing container should succeed, got: %s", err)
	}
}

// Start()
func TestContainerStart(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &container.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}

	if err := d.Start(id); err != nil {
		t.Errorf("Starting container should work, got: %s", err)
	}
}

// Stop()
func TestContainerStop(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &container.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}
	if err := d.Start(id); err != nil {
		t.Errorf("Starting container should work, got: %s", err)
	}

	if err := d.Stop(id); err != nil {
		t.Errorf("Stopping container should work, got: %s", err)
	}
}

// Status()
func TestContainerStatus(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}
	c := &container.Config{
		Image: defaults.EtcdImage,
	}
	id, err := d.Create(c)
	if err != nil {
		t.Errorf("Creating container should succeed, got: %s", err)
	}

	if _, err = d.Status(id); err != nil {
		t.Errorf("Getting container status should work, got: %s", err)
	}
}

func TestContainerStatusNonExistent(t *testing.T) {
	d, err := New(&Docker{})
	if err != nil {
		t.Errorf("Creating new docker runtime should succeed, got: %s", err)
	}

	status, err := d.Status("nonexistent")
	if err != nil {
		t.Errorf("Getting non-existent container status shouldn't return error, got: %s", err)
	}
	if status != nil {
		t.Errorf("Getting non-existent container status shouldn't return any status")
	}
}
