package controlplane

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/controlplane"
)

// Run runs static Kubernetes controlplane management tool.
func Run() int {
	return cli.Run(&controlplane.Controlplane{})
}
