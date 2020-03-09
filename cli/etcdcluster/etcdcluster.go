package etcdcluster

import (
	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/etcd"
)

// Run runs etcd cluster management CLI tool.
func Run() int {
	return cli.Run(&etcd.Cluster{})
}
