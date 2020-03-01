package controlplane

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/controlplane"
)

func Run() int {
	return cli.Run(&controlplane.Controlplane{})
}
