//go:build integration
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

// Create() tests.
func TestDockerCreateNonExistingImage(t *testing.T) {
	t.Parallel()

	containerConfig := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "notexisting",
		},
	}

	c, err := containerConfig.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	if _, err = c.Create(); err == nil {
		t.Fatalf("Creating container with non-existing image should fail")
	}
}

func TestDockerCreate(t *testing.T) {
	t.Parallel()

	containerConfig := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(t),
			Image: defaults.EtcdImage,
		},
	}

	c, err := containerConfig.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	containerID, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	t.Cleanup(func() {
		if err := containerID.Delete(); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

// Status() tests.
func TestDockerStatus(t *testing.T) {
	t.Parallel()

	containerConfig := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(t),
			Image: defaults.EtcdImage,
		},
	}

	c, err := containerConfig.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	containerID, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if _, err := containerID.Status(); err != nil {
		t.Fatalf("Checking container status should succeed, got: %v", err)
	}

	t.Cleanup(func() {
		if err := containerID.Delete(); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

func TestDockerStatusNonExistingContainer(t *testing.T) {
	t.Parallel()

	containerConfig := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(t),
			Image: defaults.EtcdImage,
		},
	}

	c, err := containerConfig.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	containerID, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	testContainerInstance, ok := containerID.(*containerInstance)
	if !ok {
		t.Fatalf("Unexpected type for containerID: %T", containerID)
	}

	originalCIID := testContainerInstance.status.ID

	testContainerInstance.status.ID = ""

	status, err := containerID.Status()
	if err != nil {
		t.Fatalf("Checking container status for non existing container should succeed")
	}

	if status.ID != "" {
		t.Fatalf("Container ID for non existing container should be empty")
	}

	testContainerInstance.status.ID = originalCIID

	t.Cleanup(func() {
		if err := containerID.Delete(); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

// Start() tests.
func TestDockerStart(t *testing.T) {
	t.Parallel()

	containerConfig := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(t),
			Image: defaults.EtcdImage,
		},
	}

	c, err := containerConfig.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	containerID, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if err := containerID.Start(); err != nil {
		t.Fatalf("Starting container should succeed, got: %v", err)
	}

	t.Cleanup(func() {
		if err := containerID.Stop(); err != nil {
			t.Logf("Stopping container should succeed, got: %v", err)

			// Deleting not stopped container will fail, so return early.
			return
		}

		if err := containerID.Delete(); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

// Stop() tests.
func TestDockerStop(t *testing.T) {
	t.Parallel()

	containerConfig := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(t),
			Image: defaults.EtcdImage,
		},
	}

	c, err := containerConfig.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	containerID, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if err := containerID.Start(); err != nil {
		t.Fatalf("Starting container should succeed, got: %v", err)
	}

	if err := containerID.Stop(); err != nil {
		t.Fatalf("Stopping container should succeed, got: %v", err)
	}

	t.Cleanup(func() {
		if err := containerID.Delete(); err != nil {
			t.Logf("Removing container should succeed, got: %v", err)
		}
	})
}

// Delete() tests.
func TestDockerDelete(t *testing.T) {
	t.Parallel()

	containerConfig := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  randomContainerName(t),
			Image: defaults.EtcdImage,
		},
	}

	c, err := containerConfig.New()
	if err != nil {
		t.Fatalf("Initializing container should succeed, got: %v", err)
	}

	containerID, err := c.Create()
	if err != nil {
		t.Fatalf("Creating container should succeed, got: %v", err)
	}

	if err := containerID.Delete(); err != nil {
		t.Fatalf("Removing container should succeed, got: %v", err)
	}
}

func randomContainerName(t *testing.T) string {
	t.Helper()

	token := make([]byte, 32)

	if _, err := rand.Read(token); err != nil {
		t.Fatalf("Generating random container name: %v", err)
	}

	return fmt.Sprintf("foo-%x", sha256.Sum256(token))
}
