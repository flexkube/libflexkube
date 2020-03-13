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

// ToExported()
func TestToExported(t *testing.T) {
	c := containersState{
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

	if diff := cmp.Diff(expected, c.Export()); diff != "" {
		t.Fatalf("unexpected diff %s", diff)
	}
}

// CheckState()
func TestContainersStateCheckStateFailStatus(t *testing.T) {
	c := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					runtimeConfig: &runtime.FakeConfig{
						Runtime: &runtime.Fake{
							StatusF: func(id string) (types.ContainerStatus, error) {
								return types.ContainerStatus{}, fmt.Errorf("fail")
							},
						},
					},
					status: types.ContainerStatus{
						ID: foo,
					},
				},
			},
		},
	}

	if err := c.CheckState(); err == nil {
		t.Fatalf("Should fail with failing status")
	}
}

func TestContainersStateCheckStateGone(t *testing.T) {
	c := containersState{
		"foo": &hostConfiguredContainer{
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					runtimeConfig: &runtime.FakeConfig{
						Runtime: &runtime.Fake{
							CreateF: func(config *types.ContainerConfig) (string, error) {
								return foo, nil
							},
							DeleteF: func(id string) error {
								return nil
							},
							StatusF: func(id string) (types.ContainerStatus, error) {
								return types.ContainerStatus{}, nil
							},
						},
					},
					status: types.ContainerStatus{
						ID: foo,
					},
				},
			},
		},
	}

	if err := c.CheckState(); err != nil {
		t.Fatalf("Checking state should succeed, got: %v", err)
	}

	if c["foo"].container.Status().Status != StatusMissing {
		t.Fatalf("Non existing container should have status '%s', got: %s", StatusMissing, c["foo"].container.Status().Status)
	}
}

// RemoveContainer()
func TestRemoveContainerDontStopStopped(t *testing.T) { //nolint:dupl
	c := containersState{
		"foo": &hostConfiguredContainer{
			hooks: &Hooks{},
			host: host.Host{
				DirectConfig: &direct.Config{},
			},
			container: &container{
				base: base{
					status: types.ContainerStatus{
						Status: "stopped",
						ID:     "foo",
					},
					runtimeConfig: &runtime.FakeConfig{
						Runtime: &runtime.Fake{
							DeleteF: func(id string) error {
								return nil
							},
							StatusF: func(id string) (types.ContainerStatus, error) {
								return types.ContainerStatus{
									Status: "stopped",
									ID:     "foo",
								}, nil
							},
							StopF: func(id string) error {
								return fmt.Errorf("stopping failed")
							},
						},
					},
				},
			},
		},
	}

	if err := c.RemoveContainer("foo"); err != nil {
		t.Fatalf("removing stopped container shouldn't try to stop it again")
	}
}

func TestRemoveContainerDontRemoveMissing(t *testing.T) {
	c := containersState{
		"foo": &hostConfiguredContainer{
			hooks: &Hooks{},
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
							DeleteF: func(id string) error {
								return fmt.Errorf("deleting failed")
							},
							StopF: func(id string) error {
								return nil
							},
						},
					},
				},
			},
		},
	}

	if err := c.RemoveContainer("foo"); err != nil {
		t.Fatalf("removing missing container shouldn't try to remove it again, got: %v", err)
	}
}

func TestRemoveContainerPropagateStopError(t *testing.T) { //nolint:dupl
	c := containersState{
		"foo": &hostConfiguredContainer{
			hooks: &Hooks{},
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
							DeleteF: func(id string) error {
								return nil
							},
							StatusF: func(id string) (types.ContainerStatus, error) {
								return types.ContainerStatus{
									Status: "running",
									ID:     "foo",
								}, nil
							},
							StopF: func(id string) error {
								return fmt.Errorf("stopping failed")
							},
						},
					},
				},
			},
		},
	}

	if err := c.RemoveContainer("foo"); err == nil {
		t.Fatalf("removing stopped container should propagate stop error")
	}
}

func TestRemoveContainerPropagateDeleteError(t *testing.T) {
	c := containersState{
		"foo": &hostConfiguredContainer{
			hooks: &Hooks{},
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
							DeleteF: func(id string) error {
								return fmt.Errorf("deleting failed")
							},
							StatusF: func(id string) (types.ContainerStatus, error) {
								return types.ContainerStatus{
									Status: "running",
									ID:     "foo",
								}, nil
							},
							StopF: func(id string) error {
								return nil
							},
						},
					},
				},
			},
		},
	}

	if err := c.RemoveContainer("foo"); err == nil {
		t.Fatalf("removing stopped container should propagate delete error")
	}
}

// createAndStart()
func TestCreateAndStartFailOnMissingContainer(t *testing.T) {
	c := containersState{}

	if err := c.CreateAndStart("foo"); err == nil {
		t.Fatalf("creating and starting non existing container should give error")
	}
}
