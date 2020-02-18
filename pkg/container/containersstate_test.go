package container

import (
	"reflect"
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

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
