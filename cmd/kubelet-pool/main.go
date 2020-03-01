package main

import (
	"os"

	"github.com/flexkube/libflexkube/cli/kubeletpool"
)

func main() {
	os.Exit(kubeletpool.Run())
}
