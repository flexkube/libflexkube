package container

import (
	"fmt"
	"reflect"
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

	if r := c.Export(); !reflect.DeepEqual(r, expected) {
		t.Fatalf("expected: %+v, got %+v", expected, r)
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
