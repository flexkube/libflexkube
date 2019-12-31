package container

import (
	"reflect"
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

func TestToExported(t *testing.T) {
	c := containersState{
		"foo": &hostConfiguredContainer{
			container: Container{
				Config: types.ContainerConfig{
					Name: "foo",
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
			},
		},
	}

	if r := c.Export(); !reflect.DeepEqual(r, expected) {
		t.Fatalf("expected: %+v, got %+v", expected, r)
	}
}
