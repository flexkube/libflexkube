package release_test

import (
	"context"
	"fmt"

	"github.com/flexkube/libflexkube/pkg/helm/release"
)

// Creating helm release.
func ExampleConfig_New() {
	config := &release.Config{
		// Put content of your kubeconfig file here.
		Kubeconfig: "",

		// The namespace must be created upfront.
		Namespace: "kube-system",

		// Name of helm release.
		Name: "coredns",

		// Repositories must be added upfront as well.
		Chart: "stable/coredns",

		// Values passed to the release in YAML format.
		Values: `replicas: 1
labels:
  foo: bar
`,
		// Version of the chart to use.
		Version: "1.12.0",
	}

	r, err := config.New()
	if err != nil {
		fmt.Printf("Creating release object failed: %v\n", err)

		return
	}

	if err := r.Install(context.TODO()); err != nil {
		fmt.Printf("Installing release failed: %v\n", err)

		return
	}
}
