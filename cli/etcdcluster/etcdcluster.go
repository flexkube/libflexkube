package etcdcluster

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/etcd"
)

func Run() int {
	return cli.Run(&etcd.Cluster{})
}
