// Package resource stores a wrapper over container.Containers, which implements
// types.Resource interface. It is stored in separate package to avoid cyclic imports.
package resource

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Containers is a wrapper over container.Containers, which implemens type.ResourceConfig.
type Containers struct {
	State      container.ContainersState `json:"state,omitempty"`
	Containers container.ContainersState `json:"containers,omitempty"`
}

// containers implements both container.ContainersInterface and type.Resource.
type containers struct {
	containers container.ContainersInterface
}

// New creates new container instance, but with generic types.Resource type.
func (c *Containers) New() (types.Resource, error) {
	co := container.Containers{
		PreviousState: c.State,
		DesiredState:  c.Containers,
	}

	ci, err := co.New()
	if err != nil {
		return nil, fmt.Errorf("failed creating containers object: %w", err)
	}

	return &containers{
		containers: ci,
	}, nil
}

// Validate validates Containers configuration.
func (c *Containers) Validate() error {
	co := container.Containers{
		PreviousState: c.State,
		DesiredState:  c.Containers,
	}

	return co.Validate()
}

// StateToYaml serializes containers to YAML format.
func (c *containers) StateToYaml() ([]byte, error) {
	co := &Containers{
		State: c.containers.ToExported().PreviousState,
	}

	return yaml.Marshal(co)
}

// CheckCurrentState is part of container.ContainersInterface.
func (c *containers) CheckCurrentState() error {
	return c.containers.CheckCurrentState()
}

// Deploy is part of container.ContainersInterface.
func (c *containers) Deploy() error {
	return c.containers.Deploy()
}

// ToExported is part of container.ContainersInterface.
func (c *containers) ToExported() *container.Containers {
	return c.containers.ToExported()
}

// DesiredState is part of container.ContainersInterface.
func (c *containers) DesiredState() container.ContainersState {
	return c.containers.DesiredState()
}

// Containers is part of container.ContainersInterface.
func (c *containers) Containers() container.ContainersInterface {
	return c.containers.Containers()
}
