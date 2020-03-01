package apiloadbalancers

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
)

func Run() int {
	return cli.Run(&apiloadbalancer.APILoadBalancers{})
}
