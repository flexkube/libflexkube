package resource

import (
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Containers is a wrapper over container.Containers, which implemens type.Resource.
type Containers struct {
	PreviousState container.ContainersState `json:"previousState,omitempty"`
	DesiredState  container.ContainersState `json:"desiredState,omitempty"`
}

// New creates new container instance, but with generic types.Resource type.
func (c *Containers) New() (types.Resource, error) {
	co := container.Containers{
		PreviousState: c.PreviousState,
		DesiredState:  c.DesiredState,
	}

	return co.New()
}
