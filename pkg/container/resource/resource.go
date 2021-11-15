// Package resource stores a wrapper over container.Containers, which implements
// types.Resource interface. It is stored in separate package to avoid cyclic imports.
package resource

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Containers is a wrapper over container.Containers, which implemens types.ResourceConfig
// and also has JSON field tags the same as other resources.
//
// It allows to orchestrate and update multiple containers spread
// across multiple hosts and update their configurations.
type Containers struct {
	// State holds containers state.
	State container.ContainersState `json:"state,omitempty"`

	// Containers stores user-provider containers to create.
	Containers container.ContainersState `json:"containers,omitempty"`
}

// containers implements both container.ContainersInterface and types.Resource.
type containers struct {
	containers container.ContainersInterface
}

// New creates new containers instance, but returns generic types.Resource type.
//
// This method will validate all the configuration provided.
func (c *Containers) New() (types.Resource, error) {
	containersConfig := container.Containers{
		PreviousState: c.State,
		DesiredState:  c.Containers,
	}

	newContainers, err := containersConfig.New()
	if err != nil {
		return nil, fmt.Errorf("creating containers object: %w", err)
	}

	return &containers{
		containers: newContainers,
	}, nil
}

// Validate validates Containers configuration.
//
// Validate is also part of types.ResourceConfig interface.
func (c *Containers) Validate() error {
	co := container.Containers{
		PreviousState: c.State,
		DesiredState:  c.Containers,
	}

	return co.Validate()
}

// StateToYaml serializes containers state to YAML format.
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

// Deploy creates configured containers.
//
// CheckCurrentState() must be called before calling Deploy(), otherwise error will be returned.
//
// Deploy is part of container.ContainersInterface.
func (c *containers) Deploy() error {
	return c.containers.Deploy()
}

// ToExported converts unexported containers struct into exported one, which can be then
// serialized and persisted.
//
// ToExported is part of container.ContainersInterface.
func (c *containers) ToExported() *container.Containers {
	return c.containers.ToExported()
}

// DesiredState returns desired state of configured containers.
//
// Desired state differs from
// exported or user-defined desired state, as it will have container IDs filled from the
// previous state.
//
// All returned containers will also have status set to running, as this is always the desired
// state of the container.
//
// Having those fields modified allows to minimize the difference when comparing previous state
// and desired state.
//
// DesiredState is part of container.ContainersInterface.
func (c *containers) DesiredState() container.ContainersState {
	return c.containers.DesiredState()
}

// Containers is part of container.ContainersInterface.
func (c *containers) Containers() container.ContainersInterface {
	return c.containers
}
