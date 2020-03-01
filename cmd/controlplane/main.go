package main

import (
	"os"

	"github.com/flexkube/libflexkube/cli/controlplane"
)

func main() {
	os.Exit(controlplane.Run())
}
