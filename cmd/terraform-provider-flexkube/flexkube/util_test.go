package flexkube

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/resource"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// saveState()
func TestSaveStateBadScheme(t *testing.T) {
	r := resourceContainers()
	delete(r.Schema, "state_yaml")

	d := r.Data(&terraform.InstanceState{})

	if err := saveState(d, container.ContainersState{}, containersUnmarshal, nil); err == nil {
		t.Fatalf("save state should fail when called on bad scheme")
	}
}

// resourceDelete()
func TestResourceDeleteRuntimeFail(t *testing.T) {
	r := resourceContainers()

	s := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "foo",
					Image: "busybox:latest",
				},
				Status: types.ContainerStatus{
					ID:     "foo",
					Status: "running",
				},
			},
		},
	}

	d := r.Data(&terraform.InstanceState{})
	if err := d.Set("state_sensitive", containersStateMarshal(s, false)); err != nil {
		t.Fatalf("Failed writing: %v", err)
	}

	if err := r.Delete(d, nil); err == nil {
		t.Fatalf("destroying should fail for unreachable runtime")
	}
}

func TestResourceDeleteEmpty(t *testing.T) {
	r := resourceContainers()
	r.Delete = resourceDelete(containersUnmarshal, "state_sensitive")

	s := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "foo",
					Image: "busybox:latest",
				},
			},
		},
	}

	d := r.Data(&terraform.InstanceState{})
	if err := d.Set("state_sensitive", containersStateMarshal(container.ContainersState{}, false)); err != nil {
		t.Fatalf("Failed writing: %v", err)
	}

	if err := d.Set("container", containersStateMarshal(s, false)); err != nil {
		t.Fatalf("writing containers configuration to state failed: %v", err)
	}

	if err := r.Delete(d, nil); !strings.Contains(err.Error(), "Is the docker daemon running") {
		t.Fatalf("destroying should fail for unreachable runtime")
	}
}

func TestResourceDeleteEmptyState(t *testing.T) {
	r := resourceContainers()

	if err := r.Delete(r.Data(&terraform.InstanceState{}), nil); err == nil {
		t.Fatalf("initializing from empty state should fail")
	}
}

func TestResourceDeleteBadKey(t *testing.T) {
	r := resourceContainers()
	r.Delete = resourceDelete(containersUnmarshal, "foo")

	if err := r.Delete(r.Data(&terraform.InstanceState{}), nil); err == nil {
		t.Fatalf("emptying key not existing in scheme should fail")
	}
}

// newResource()
func TestNewResourceFailRefresh(t *testing.T) {
	cc := &resource.Containers{
		PreviousState: container.ContainersState{
			"foo": &container.HostConfiguredContainer{
				Host: host.Host{
					DirectConfig: &direct.Config{},
				},
				Container: container.Container{
					Runtime: container.RuntimeConfig{
						Docker: &docker.Config{
							Host: "unix:///nonexistent",
						},
					},
					Config: types.ContainerConfig{
						Name:  "foo",
						Image: "busybox:latest",
					},
					Status: types.ContainerStatus{
						ID:     "foo",
						Status: "running",
					},
				},
			},
		},
	}

	if _, err := newResource(cc, true); err == nil {
		t.Fatalf("should check for errors when checking current state")
	}
}

// resourceCreate()
func TestResourceCreate(t *testing.T) {
	r := resourceContainers()

	s := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "foo",
					Image: "busybox:latest",
				},
			},
		},
	}

	d := r.Data(&terraform.InstanceState{})
	if err := d.Set("container", containersStateMarshal(s, false)); err != nil {
		t.Fatalf("writing containers configuration to state failed: %v", err)
	}

	if err := r.Create(d, nil); !strings.Contains(err.Error(), "Is the docker daemon running") {
		t.Fatalf("creating should fail for unreachable runtime, got: %v", err)
	}
}

func TestResourceCreateFailInitialize(t *testing.T) {
	r := resourceContainers()

	s := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "",
					Image: "busybox:latest",
				},
			},
		},
	}

	d := r.Data(&terraform.InstanceState{})
	if err := d.Set("container", containersStateMarshal(s, false)); err != nil {
		t.Fatalf("writing containers configuration to state failed: %v", err)
	}

	if err := r.Create(d, nil); !strings.Contains(err.Error(), "name must be set") {
		t.Fatalf("creating should fail for unreachable runtime, got: %v", err)
	}
}

// resourceRead()
func TestResourceRead(t *testing.T) {
	r := resourceContainers()

	s := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "foo",
					Image: "busybox:latest",
				},
			},
		},
	}

	d := r.Data(&terraform.InstanceState{})
	if err := d.Set("container", containersStateMarshal(s, false)); err != nil {
		t.Fatalf("writing containers configuration to state failed: %v", err)
	}

	if err := r.Read(d, nil); err != nil {
		t.Fatalf("reading with no previous state should succeed, got: %v", err)
	}
}

func TestResourceReadFailInitialize(t *testing.T) {
	r := resourceContainers()

	s := container.ContainersState{
		"foo": &container.HostConfiguredContainer{
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
			Container: container.Container{
				Runtime: container.RuntimeConfig{
					Docker: &docker.Config{
						Host: "unix:///nonexistent",
					},
				},
				Config: types.ContainerConfig{
					Name:  "",
					Image: "busybox:latest",
				},
			},
		},
	}

	d := r.Data(&terraform.InstanceState{})
	if err := d.Set("container", containersStateMarshal(s, false)); err != nil {
		t.Fatalf("writing containers configuration to state failed: %v", err)
	}

	if err := r.Read(d, nil); err == nil {
		t.Fatalf("read should check for initialize errors and fail")
	}
}
