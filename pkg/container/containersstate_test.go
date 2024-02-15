package container

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// ToExported() tests.
func TestToExported(t *testing.T) {
	t.Parallel()

	testState := containersState{
		"foo": &hostConfiguredContainer{
			container: &container{
				base: base{
					config: types.ContainerConfig{
						Name: "foo",
					},
					runtimeConfig: &docker.Config{},
				},
			},
		},
	}

	expected := ContainersState{
		"foo": &HostConfiguredContainer{
			ConfigFiles: map[string]string{},
			Container: Container{
				Config: types.ContainerConfig{
					Name: "foo",
				},
				Runtime: RuntimeConfig{
					Docker: &docker.Config{},
				},
			},
		},
	}

	if diff := cmp.Diff(expected, testState.Export()); diff != "" {
		t.Fatalf("Unexpected diff %s", diff)
	}
}

// CheckState() tests.
func TestContainersStateCheckStateFailStatus(t *testing.T) {
	t.Parallel()

	testState := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					runtimeConfig: &runtime.FakeConfig{
						Runtime: &runtime.Fake{
							StatusF: func(string) (types.ContainerStatus, error) {
								return types.ContainerStatus{}, fmt.Errorf("fail")
							},
						},
					},
					status: types.ContainerStatus{
						ID: testContainerID,
					},
				},
			},
		},
	}

	if err := testState.CheckState(); err != nil {
		t.Fatalf("Should not fail with failing status")
	}

	if testState["foo"].container.Status().ID != "" {
		t.Errorf("Failing status call should reset container ID, so we assume container is gone")
	}

	if testState["foo"].container.Status().Status == "fail" {
		t.Errorf("Container status should include error message returned by status function")
	}
}

func TestContainersStateCheckStateGone(t *testing.T) {
	t.Parallel()

	testState := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					runtimeConfig: &runtime.FakeConfig{
						Runtime: &runtime.Fake{
							CreateF: func(*types.ContainerConfig) (string, error) {
								return testContainerID, nil
							},
							DeleteF: func(string) error {
								return nil
							},
							StatusF: func(string) (types.ContainerStatus, error) {
								return types.ContainerStatus{}, nil
							},
						},
					},
					status: types.ContainerStatus{
						ID: testContainerID,
					},
				},
			},
		},
	}

	if err := testState.CheckState(); err != nil {
		t.Fatalf("Checking state should succeed, got: %v", err)
	}

	if testState["foo"].container.Status().Status != StatusMissing {
		t.Fatalf("Non existing container should have status %q, got: %q",
			StatusMissing, testState["foo"].container.Status().Status)
	}
}

func failingStopRuntime() *runtime.Fake {
	r := fakeRuntime()
	r.StopF = func(string) error {
		return fmt.Errorf("stopping")
	}

	return r
}

// RemoveContainer() tests.
func TestRemoveContainerDontStopStopped(t *testing.T) {
	t.Parallel()

	failingStopRuntime := failingStopRuntime()
	failingStopRuntime.StatusF = func(string) (types.ContainerStatus, error) {
		return types.ContainerStatus{
			Status: "stopped",
			ID:     "foo",
		}, nil
	}

	testState := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					status: types.ContainerStatus{
						Status: "stopped",
						ID:     "foo",
					},
					runtimeConfig: asRuntime(failingStopRuntime),
				},
			},
		},
	}

	if err := testState.RemoveContainer("foo"); err != nil {
		t.Fatalf("Removing stopped container shouldn't try to stop it again")
	}
}

func TestRemoveContainerDontRemoveMissing(t *testing.T) {
	t.Parallel()

	testState := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					status: types.ContainerStatus{
						Status: "gone",
						ID:     "",
					},
					runtimeConfig: &runtime.FakeConfig{
						Runtime: &runtime.Fake{
							DeleteF: func(string) error {
								return fmt.Errorf("deleting failed")
							},
							StopF: func(string) error {
								return nil
							},
						},
					},
				},
			},
		},
	}

	if err := testState.RemoveContainer("foo"); err != nil {
		t.Fatalf("Removing missing container shouldn't try to remove it again, got: %v", err)
	}
}

func TestRemoveContainerPropagateStopError(t *testing.T) {
	t.Parallel()

	failingStopRuntime := failingStopRuntime()
	failingStopRuntime.StatusF = func(string) (types.ContainerStatus, error) {
		return types.ContainerStatus{
			Status: "running",
			ID:     "foo",
		}, nil
	}

	testState := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					status: types.ContainerStatus{
						Status: "running",
						ID:     "foo",
					},
					runtimeConfig: asRuntime(failingStopRuntime),
				},
			},
		},
	}

	if err := testState.RemoveContainer("foo"); err == nil {
		t.Fatalf("Removing stopped container should propagate stop error")
	}
}

func TestRemoveContainerPropagateDeleteError(t *testing.T) {
	t.Parallel()

	testState := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					status: types.ContainerStatus{
						Status: "running",
						ID:     "foo",
					},
					runtimeConfig: &runtime.FakeConfig{
						Runtime: &runtime.Fake{
							DeleteF: func(string) error {
								return fmt.Errorf("deleting failed")
							},
							StatusF: func(string) (types.ContainerStatus, error) {
								return types.ContainerStatus{
									Status: "running",
									ID:     "foo",
								}, nil
							},
							StopF: func(string) error {
								return nil
							},
						},
					},
				},
			},
		},
	}

	if err := testState.RemoveContainer("foo"); err == nil {
		t.Fatalf("Removing stopped container should propagate delete error")
	}
}

// createAndStart() tests.
func TestCreateAndStartFailOnMissingContainer(t *testing.T) {
	t.Parallel()

	testState := containersState{}

	if err := testState.CreateAndStart("foo"); err == nil {
		t.Fatalf("Creating and starting non existing container should give error")
	}
}
