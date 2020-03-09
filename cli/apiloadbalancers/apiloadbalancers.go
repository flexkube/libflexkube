package apiloadbalancers

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
)

// Run runs apiloadbalancers creation tool.
func Run() int {
	return cli.Run(&apiloadbalancer.APILoadBalancers{})
}
