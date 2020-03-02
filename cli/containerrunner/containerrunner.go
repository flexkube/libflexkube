package containerrunner

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/container/resource"
)

func Run() int {
	return cli.Run(&resource.Containers{})
}
