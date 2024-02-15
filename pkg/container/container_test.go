package container

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

// New() tests.
func TestNewEmptyConfiguration(t *testing.T) {
	t.Parallel()

	if _, err := (&Container{}).New(); err == nil {
		t.Errorf("Creating container with wrong configuration should fail")
	}
}

func TestNewGoodConfiguration(t *testing.T) {
	t.Parallel()

	testContainer := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "nonexistent",
		},
	}
	if _, err := testContainer.New(); err != nil {
		t.Errorf("Creating container with good configuration should pass, got: %v", err)
	}
}

// Validate() tests.
func TestValidateNoName(t *testing.T) {
	t.Parallel()

	testContainer := &Container{
		Config: types.ContainerConfig{},
	}
	if err := testContainer.Validate(); err == nil {
		t.Errorf("Validating container without name should fail")
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	testContainer := &Container{
		Runtime: RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "nonexistent",
		},
	}
	if err := testContainer.Validate(); err != nil {
		t.Errorf("Validating container with valid configuration should pass, got: %v", err)
	}
}

func TestValidateUnsupportedRuntime(t *testing.T) {
	t.Parallel()

	testContainer := &Container{
		Config: types.ContainerConfig{
			Name:  "foo",
			Image: "nonexistent",
		},
	}
	if err := testContainer.Validate(); err == nil {
		t.Errorf("Validating container with unsupported container runtime should fail")
	}
}

func TestValidateRequireImage(t *testing.T) {
	t.Parallel()

	testContainer := &Container{
		Config: types.ContainerConfig{
			Name: "foo",
		},
	}
	if err := testContainer.Validate(); err == nil {
		t.Errorf("Validating container with no image set should fail")
	}
}

// selectRuntime() tests.
func TestSelectDockerRuntime(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			runtimeConfig: &docker.Config{},
			status:        types.ContainerStatus{},
		},
	}
	if err := testContainer.selectRuntime(); err != nil {
		t.Errorf("Selecting Docker container runtime should succeed, got: %v", err)
	}

	if testContainer.runtime == nil {
		t.Errorf("Selecting container runtime should set container runtime field")
	}
}

// FromStatus() tests.
func TestFromStatusValid(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			status: types.ContainerStatus{
				ID: "nonexistent",
			},
		},
	}
	if _, err := testContainer.FromStatus(); err != nil {
		t.Fatalf("Container instance should be created from valid container, got: %v", err)
	}
}

func TestFromStatusNoID(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}
	if _, err := testContainer.FromStatus(); err == nil {
		t.Fatalf("Container instance should not be created from container with no container ID")
	}
}

// Status() tests.
func TestStatus(t *testing.T) {
	t.Parallel()

	testContainer := &containerInstance{
		base: base{
			runtime: runtime.Fake{
				StatusF: func(string) (types.ContainerStatus, error) {
					return types.ContainerStatus{}, fmt.Errorf("checking status")
				},
			},
		},
	}

	if _, err := testContainer.Status(); err == nil {
		t.Fatalf("Checking container status should propagate failure")
	}
}

// UpdateStatus() tests.
func TestContainerUpdateStatusEmptyStatus(t *testing.T) {
	t.Parallel()

	testContainer := &container{}

	if err := testContainer.UpdateStatus(); err == nil {
		t.Fatalf("Updating status of non-existing container should fail")
	}
}

func TestContainerUpdateStatusFail(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				StatusF: func(string) (types.ContainerStatus, error) {
					return types.ContainerStatus{}, fmt.Errorf("checking status")
				},
			},
			status: types.ContainerStatus{
				ID: "foo",
			},
		},
	}

	if err := testContainer.UpdateStatus(); err == nil {
		t.Fatalf("Updating status with failing runtime should fail")
	}
}

func TestContainerUpdateStatus(t *testing.T) {
	t.Parallel()

	expectedStatus := types.ContainerStatus{
		ID:     "foo",
		Status: "running",
	}

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				StatusF: func(string) (types.ContainerStatus, error) {
					return expectedStatus, nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := testContainer.UpdateStatus(); err != nil {
		t.Fatalf("Updating status should succeed, got: %v", err)
	}

	if diff := cmp.Diff(testContainer.status, expectedStatus); diff != "" {
		t.Fatalf("Container status should be set to received status: %s", diff)
	}
}

// Start() tests.
func TestContainerStartBadState(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}

	if err := testContainer.Start(); err == nil {
		t.Fatalf("Starting non-existing container should fail")
	}
}

func TestContainerStartRuntimeError(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				StartF: func(string) error {
					return fmt.Errorf("starting container failed")
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := testContainer.Start(); err == nil {
		t.Fatalf("Starting container should fail when runtime error occurs")
	}
}

func TestContainerStart(t *testing.T) {
	t.Parallel()

	expectedStatus := types.ContainerStatus{
		ID:     "foo",
		Status: "running",
	}

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				StartF: func(string) error {
					return nil
				},
				StatusF: func(string) (types.ContainerStatus, error) {
					return expectedStatus, nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := testContainer.Start(); err != nil {
		t.Fatalf("Starting should succeed, got: %v", err)
	}

	if diff := cmp.Diff(testContainer.status, expectedStatus); diff != "" {
		t.Fatalf("Container status should be updated after starting: %s", diff)
	}
}

// Stop() tests.
func TestContainerStopBadState(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}

	if err := testContainer.Stop(); err == nil {
		t.Fatalf("Stopping non-existing container should fail")
	}
}

func TestContainerStopRuntimeError(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				StopF: func(string) error {
					return fmt.Errorf("starting container failed")
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := testContainer.Stop(); err == nil {
		t.Fatalf("Stopping container should fail when runtime error occurs")
	}
}

func TestContainerStop(t *testing.T) {
	t.Parallel()

	expectedStatus := types.ContainerStatus{
		ID:     "foo",
		Status: "stopped",
	}

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				StopF: func(string) error {
					return nil
				},
				StatusF: func(string) (types.ContainerStatus, error) {
					return expectedStatus, nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "running",
			},
		},
	}

	if err := testContainer.Stop(); err != nil {
		t.Fatalf("Stopping should succeed, got: %v", err)
	}

	if diff := cmp.Diff(testContainer.status, expectedStatus); diff != "" {
		t.Fatalf("Container status should be updated after starting: %s", diff)
	}
}

// Delete() tests.
func TestContainerDeleteBadState(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			status: types.ContainerStatus{},
		},
	}

	if err := testContainer.Delete(); err == nil {
		t.Fatalf("Deleting non-existing container should fail")
	}
}

func TestContainerDeleteRuntimeError(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				DeleteF: func(string) error {
					return fmt.Errorf("starting container failed")
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "stopped",
			},
		},
	}

	if err := testContainer.Delete(); err == nil {
		t.Fatalf("Deleting container should fail when runtime error occurs")
	}
}

func TestContainerDelete(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			runtime: runtime.Fake{
				DeleteF: func(string) error {
					return nil
				},
			},
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "running",
			},
		},
	}

	if err := testContainer.Delete(); err != nil {
		t.Fatalf("Deleting should succeed, got: %v", err)
	}

	if testContainer.status.ID != "" {
		t.Fatalf("Delete should remove ID from status")
	}
}

// SetStatus() tests.
func TestContainerSetStatus(t *testing.T) {
	t.Parallel()

	testContainer := &container{
		base: base{
			status: types.ContainerStatus{
				ID:     "foo",
				Status: "running",
			},
		},
	}

	expectedStatus := types.ContainerStatus{
		ID:     "bar",
		Status: "boom",
	}

	testContainer.SetStatus(expectedStatus)

	if diff := cmp.Diff(testContainer.base.status, expectedStatus); diff != "" {
		t.Fatalf("Unexpected diff: %s", diff)
	}
}
