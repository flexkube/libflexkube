package kubeletpool

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/kubelet"
)

// Run runs CLI tool for managing kubelet containers.
func Run() int {
	return cli.Run(&kubelet.Pool{})
}
