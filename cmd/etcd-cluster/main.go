package main

import (
	"os"

	"github.com/flexkube/libflexkube/cli/etcdcluster"
)

func main() {
	os.Exit(etcdcluster.Run())
}
