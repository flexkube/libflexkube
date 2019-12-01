// +build integration

package container

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
)

// Create()
func TestDockerCreateNonExistingImage(t *testing.T) {
	node := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "notexisting",
		},
	}
	n, err := New(node)
	if err != nil {
		t.Fatalf("Initializing node should succeed, got: %v", err)
	}
	if _, err = n.Create(); err == nil {
		t.Fatalf("Creating node with non-existing image should fail")
	}
}

func TestDockerCreate(t *testing.T) {
	node := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}
	n, err := New(node)
	if err != nil {
		t.Fatalf("Initializing node should succeed, got: %v", err)
	}
	if _, err := n.Create(); err != nil {
		t.Fatalf("Creating node should succeed, got: %v", err)
	}
}

// Status()
func TestDockerStatus(t *testing.T) {
	node := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}
	n, err := New(node)
	if err != nil {
		t.Fatalf("Initializing node should succeed, got: %v", err)
	}
	c, err := n.Create()
	if err != nil {
		t.Fatalf("Creating node should succeed, got: %v", err)
	}
	if _, err := c.Status(); err != nil {
		t.Fatalf("Checking node status should succeed, got: %v", err)
	}
}

func TestDockerStatusNonExistingContainer(t *testing.T) {
	node := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}
	n, err := New(node)
	if err != nil {
		t.Fatalf("Initializing node should succeed, got: %v", err)
	}
	c, err := n.Create()
	if err != nil {
		t.Fatalf("Creating node should succeed, got: %v", err)
	}
	c.status.ID = "nonexistent"
	status, err := c.Status()
	if err != nil {
		t.Fatalf("Checking node status for non existing container should succeed")
	}
	if status != nil {
		t.Fatalf("Node status for non existing container should be nil")
	}
}

// Start()
func TestDockerStart(t *testing.T) {
	node := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}
	n, err := New(node)
	if err != nil {
		t.Fatalf("Initializing node should succeed, got: %v", err)
	}
	c, err := n.Create()
	if err != nil {
		t.Fatalf("Creating node should succeed, got: %v", err)
	}
	if err := c.Start(); err != nil {
		t.Fatalf("Starting container should succeed, got: %v", err)
	}
}

// Stop()
func TestDockerStop(t *testing.T) {
	node := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}
	n, err := New(node)
	if err != nil {
		t.Fatalf("Initializing node should succeed, got: %v", err)
	}
	c, err := n.Create()
	if err != nil {
		t.Fatalf("Creating node should succeed, got: %v", err)
	}
	if err := c.Start(); err != nil {
		t.Fatalf("Starting container should succeed, got: %v", err)
	}
	if err := c.Stop(); err != nil {
		t.Fatalf("Stopping container should succeed, got: %v", err)
	}
}

// Delete()
func TestDockerDelete(t *testing.T) {
	node := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}
	n, err := New(node)
	if err != nil {
		t.Fatalf("Initializing node should succeed, got: %v", err)
	}
	c, err := n.Create()
	if err != nil {
		t.Fatalf("Creating node should succeed, got: %v", err)
	}
	if err := c.Delete(); err != nil {
		t.Fatalf("Removing container should succeed, got: %v", err)
	}
}

func randomContainerName() string {
	token := make([]byte, 32)
	rand.Read(token)
	return fmt.Sprintf("foo-%x", md5.Sum(token))
}
