package container

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

// New()
func TestNewEmptyConfiguration(t *testing.T) {
	if _, err := (&Container{}).New(); err == nil {
		t.Errorf("Creating container with wrong configuration should fail")
	}
}

func TestNewGoodConfiguration(t *testing.T) {
	c := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "nonexistent",
		},
	}
	if _, err := c.New(); err != nil {
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
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
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
		base: base{
			runtimeConfig: &docker.Config{},
			status:        types.ContainerStatus{},
		},
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
		base: base{
			status: types.ContainerStatus{
				ID: "nonexistent",
			},
		},
	}
	if _, err := c.FromStatus(); err != nil {
		t.Fatalf("Container instance should be created from valid container, got: %v", err)
	}
}

func TestFromStatusNoID(t *testing.T) {
	c := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}
	if _, err := c.FromStatus(); err == nil {
		t.Fatalf("Container instance should not be created from container with no container ID")
	}
}

// Status()
func TestStatus(t *testing.T) {
	c := &containerInstance{
		base: base{
			runtime: runtime.Fake{
				StatusF: func(ID string) (types.ContainerStatus, error) {
					return types.ContainerStatus{}, fmt.Errorf("failed checking status")
				},
			},
		},
	}

	if _, err := c.Status(); err == nil {
		t.Fatalf("Checking container status should propagate failure")
	}
}

// UpdateStatus()
func TestContainerUpdateStatusEmptyStatus(t *testing.T) {
	c := &container{}

	if err := c.UpdateStatus(); err == nil {
		t.Fatalf("Updating status of non-existing container should fail")
	}
}

func TestContainerUpdateStatusFail(t *testing.T) {
	c := &container{
		base: base{
			runtime: runtime.Fake{
				StatusF: func(ID string) (types.ContainerStatus, error) {
					return types.ContainerStatus{}, fmt.Errorf("failed checking status")
				},
			},
			status: types.ContainerStatus{
				ID: "foo",
			},
		},
	}

	if err := c.UpdateStatus(); err == nil {
		t.Fatalf("Updating status with failing runtime should fail")
	}
}

func TestContainerUpdateStatus(t *testing.T) {
	ns := types.ContainerStatus{
		ID:     "foo",
		Status: "running",
	}

	c := &container{
		base: base{
			runtime: runtime.Fake{
				StatusF: func(ID string) (types.ContainerStatus, error) {
					return ns, nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := c.UpdateStatus(); err != nil {
		t.Fatalf("Updating status should succeed, got: %v", err)
	}

	if diff := cmp.Diff(ns, c.status); diff != "" {
		t.Fatalf("Container status should be set to received status: %s", diff)
	}
}

// Start()
func TestContainerStartBadState(t *testing.T) {
	c := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}

	if err := c.Start(); err == nil {
		t.Fatalf("Starting non-existing container should fail")
	}
}

func TestContainerStartRuntimeError(t *testing.T) {
	c := &container{
		base: base{
			runtime: runtime.Fake{
				StartF: func(ID string) error {
					return fmt.Errorf("starting container failed")
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := c.Start(); err == nil {
		t.Fatalf("Starting container should fail when runtime error occurs")
	}
}

func TestContainerStart(t *testing.T) {
	ns := types.ContainerStatus{
		ID:     "foo",
		Status: "running",
	}

	c := &container{
		base: base{
			runtime: runtime.Fake{
				StartF: func(ID string) error {
					return nil
				},
				StatusF: func(ID string) (types.ContainerStatus, error) {
					return ns, nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := c.Start(); err != nil {
		t.Fatalf("Starting should succeed, got: %v", err)
	}

	if diff := cmp.Diff(ns, c.status); diff != "" {
		t.Fatalf("Container status should be updated after starting: %s", diff)
	}
}

// Stop()
func TestContainerStopBadState(t *testing.T) {
	c := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}

	if err := c.Stop(); err == nil {
		t.Fatalf("Stopping non-existing container should fail")
	}
}

func TestContainerStopRuntimeError(t *testing.T) {
	c := &container{
		base: base{
			runtime: runtime.Fake{
				StopF: func(ID string) error {
					return fmt.Errorf("starting container failed")
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := c.Stop(); err == nil {
		t.Fatalf("Stopping container should fail when runtime error occurs")
	}
}

func TestContainerStop(t *testing.T) {
	ns := types.ContainerStatus{
		ID:     "foo",
		Status: "stopped",
	}

	c := &container{
		base: base{
			runtime: runtime.Fake{
				StopF: func(ID string) error {
					return nil
				},
				StatusF: func(ID string) (types.ContainerStatus, error) {
					return ns, nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "running",
			},
		},
	}

	if err := c.Stop(); err != nil {
		t.Fatalf("Stopping should succeed, got: %v", err)
	}

	if diff := cmp.Diff(ns, c.status); diff != "" {
		t.Fatalf("Container status should be updated after starting: %s", diff)
	}
}

// Delete()
func TestContainerDeleteBadState(t *testing.T) {
	c := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}

	if err := c.Delete(); err == nil {
		t.Fatalf("Deleting non-existing container should fail")
	}
}

func TestContainerDeleteRuntimeError(t *testing.T) {
	c := &container{
		base: base{
			runtime: runtime.Fake{
				DeleteF: func(ID string) error {
					return fmt.Errorf("starting container failed")
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := c.Delete(); err == nil {
		t.Fatalf("Deleting container should fail when runtime error occurs")
	}
}

func TestContainerDelete(t *testing.T) {
	c := &container{
		base: base{
			runtime: runtime.Fake{
				DeleteF: func(ID string) error {
					return nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "running",
			},
		},
	}

	if err := c.Delete(); err != nil {
		t.Fatalf("Deleting should succeed, got: %v", err)
	}

	if c.status.ID != "" {
		t.Fatalf("Delete should remove ID from status")
	}
}
