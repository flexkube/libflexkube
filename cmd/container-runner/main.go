package main

import (
	"os"

	"github.com/flexkube/libflexkube/cli/containerrunner"
)

func main() {
	os.Exit(containerrunner.Run())
}
