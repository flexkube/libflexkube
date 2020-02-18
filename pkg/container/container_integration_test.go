// +build integration

package container

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
)

// Create()
func TestDockerCreateNonExistingImage(t *testing.T) {
	cc := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "notexisting",
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	if _, err = c.Create(); err == nil {
		t.Fatalf("Creating container with non-existing image should fail")
	}
}

func TestDockerCreate(t *testing.T) {
	cc := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	if _, err := c.Create(); err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}
}

// Status()
func TestDockerStatus(t *testing.T) {
	cc := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	ci, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if _, err := ci.Status(); err != nil {
		t.Fatalf("Checking container status should succeed, got: %v", err)
	}
}

func TestDockerStatusNonExistingContainer(t *testing.T) {
	cc := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	ci, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	ci.(*containerInstance).status.ID = ""

	status, err := ci.Status()
	if err != nil {
		t.Fatalf("Checking container status for non existing container should succeed")
	}

	if status.ID != "" {
		t.Fatalf("Container ID for non existing container should be empty")
	}
}

// Start()
func TestDockerStart(t *testing.T) {
	cc := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	ci, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if err := ci.Start(); err != nil {
		t.Fatalf("Starting container should succeed, got: %v", err)
	}
}

// Stop()
func TestDockerStop(t *testing.T) {
	cc := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	ci, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if err := ci.Start(); err != nil {
		t.Fatalf("Starting container should succeed, got: %v", err)
	}

	if err := ci.Stop(); err != nil {
		t.Fatalf("Stopping container should succeed, got: %v", err)
	}
}

// Delete()
func TestDockerDelete(t *testing.T) {
	cc := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(),
			Image: defaults.EtcdImage,
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	ci, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if err := ci.Delete(); err != nil {
		t.Fatalf("Removing container should succeed, got: %v", err)
	}
}

func randomContainerName() string {
	token := make([]byte, 32)

	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("foo-%x", sha256.Sum256(token))
}
