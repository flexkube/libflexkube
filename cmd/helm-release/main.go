package main

import (
	"fmt"
	"io/ioutil"

	"github.com/invidian/libflexkube/pkg/helm/release"
)

func main() {
	config, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}

	release, err := release.FromYaml(config)
	if err != nil {
		panic(err)
	}

	if err := release.ValidateChart(); err != nil {
		panic(err)
	}

	e, err := release.Exists()
	if err != nil {
		panic(err)
	}

	if e {
		fmt.Println("Release does exist, updating")
	} else {
		fmt.Println("Release does NOT exist, installing")
	}

	if err := release.InstallOrUpgrade(); err != nil {
		panic(err)
	}
}
