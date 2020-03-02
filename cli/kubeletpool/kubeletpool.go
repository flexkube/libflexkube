package kubeletpool

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/kubelet"
)

func Run() int {
	return cli.Run(&kubelet.Pool{})
}
