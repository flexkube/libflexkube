package containerrunner

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/container/resource"
)

// Run runs generic containers creation CLI tool.
func Run() int {
	return cli.Run(&resource.Containers{})
}
