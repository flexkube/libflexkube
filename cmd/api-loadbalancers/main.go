package main

import (
	"os"

	"github.com/flexkube/libflexkube/cli/apiloadbalancers"
)

func main() {
	os.Exit(apiloadbalancers.Run())
}
