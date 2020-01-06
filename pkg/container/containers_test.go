package container

import (
	"reflect"
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

// New()
func TestContainersNew(t *testing.T) {
	GetContainers(t)
}

func GetContainers(t *testing.T) ContainersInterface {
	cc := &Containers{
		DesiredState: ContainersState{
			"foo": &HostConfiguredContainer{
				Host: host.Host{
					DirectConfig: &direct.Config{},
				},
				Container: Container{
					Runtime: RuntimeConfig{
						Docker: &docker.Config{},
					},
					Config: types.ContainerConfig{
						Name:  "foo",
						Image: "busybox:latest",
					},
				},
			},
		},
	}

	c, err := cc.New()
	if err != nil {
		t.Fatalf("Creating empty containers object should work, got: %v", err)
	}

	return c
}

// CheckCurrentState()
func TestContainersCheckCurrentStateNew(t *testing.T) {
	c := GetContainers(t)

	if err := c.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state for new Containers should work, got: %v", err)
	}
}

// CurrentStateToYaml()
func TestContainersCurrentStateToYAML(t *testing.T) {
	c := GetContainers(t)

	_, err := c.CurrentStateToYaml()
	if err != nil {
		t.Fatalf("Getting current state in YAML format should work, got: %v", err)
	}
}

// ToExported()
func TestContainersToExported(t *testing.T) {
	c := GetContainers(t)

	c.ToExported()
}

// FromYaml()
func TestContainersFromYamlBad(t *testing.T) {
	if _, err := FromYaml([]byte{}); err == nil {
		t.Fatalf("Creating containers from empty YAML should fail")
	}
}

func TestContainersFromYaml(t *testing.T) {
	y := `
desiredState:
 foo:
   host:
     direct: {}
   container:
     runtime:
       docker: {}
     config:
       name: foo
       image: busybox
`

	if _, err := FromYaml([]byte(y)); err != nil {
		t.Fatalf("Creating containers from valid YAML should work, got: %v", err)
	}
}

func TestFilesToUpdateEmpty(t *testing.T) {
	expected := []string{"foo"}

	d := hostConfiguredContainer{
		configFiles: map[string]string{
			"foo": "bar",
		},
	}

	if v := filesToUpdate(d, nil); !reflect.DeepEqual(expected, v) {
		t.Fatalf("Expected %v, got %v", expected, d)
	}
}
