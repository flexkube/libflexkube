package resource

import (
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/types"
)

type Containers struct {
	PreviousState container.ContainersState `json:"previousState,omitempty"`
	DesiredState  container.ContainersState `json:"desiredState,omitempty"`
}

func (c *Containers) New() (types.Resource, error) {
	co := container.Containers{
		PreviousState: c.PreviousState,
		DesiredState:  c.DesiredState,
	}

	return co.New()
}
