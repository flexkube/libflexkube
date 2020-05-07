// Package pkigenerator contains implementation of CLI tool for
// managing Kubernetes PKI.
package pkigenerator

import (
	"fmt"
	"io/ioutil"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/cli"
	"github.com/flexkube/libflexkube/pkg/pki"
)

func getPKI() (*pki.PKI, error) {
	pki := &pki.PKI{}

	b, err := cli.ReadYamlFile("pki.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading pki.yaml file failed: %w", err)
	}

	if err := yaml.Unmarshal(b, pki); err != nil {
		return nil, fmt.Errorf("parsing pki.yaml file failed: %w", err)
	}

	return pki, nil
}

func savePKI(pki *pki.PKI) error {
	o, err := yaml.Marshal(pki)
	if err != nil {
		return fmt.Errorf("failed serializing PKI: %w", err)
	}

	if err := ioutil.WriteFile("pki.yaml", o, 0600); err != nil {
		return fmt.Errorf("failed writing pki.yaml file: %w", err)
	}

	return nil
}

// Run runs Kubernetes PKI manager.
func Run() int {
	pki, err := getPKI()
	if err != nil {
		fmt.Printf("Failed preparing PKI: %v", err)

		return cli.ExitError
	}

	gErr := pki.Generate()

	if err := savePKI(pki); err != nil {
		fmt.Printf("Failed saving PKI: %v\n", err)

		return cli.ExitError
	}

	if gErr != nil {
		fmt.Printf("Generating PKI failed: %v\n", gErr)

		return cli.ExitError
	}

	return cli.ExitOK
}
