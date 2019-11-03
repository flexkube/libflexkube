package main

import (
	"fmt"
	"io/ioutil"

	"sigs.k8s.io/yaml"

	"github.com/invidian/libflexkube/pkg/helm"
)

func main() {
	config, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}

	r := helm.Release{}

	if err := yaml.Unmarshal(config, &r); err != nil {
		panic(err)
	}

	if err := r.Validate(); err != nil {
		panic(err)
	}

	release, _ := r.New()

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
