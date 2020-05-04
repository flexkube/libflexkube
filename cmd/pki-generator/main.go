package main

import (
	"os"

	"github.com/flexkube/libflexkube/cli/pkigenerator"
)

func main() {
	os.Exit(pkigenerator.Run())
}
