package types

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

// Resource interface defines flexkube resource like kubelet pool or static controlplane.
type Resource interface {
	StateToYaml() ([]byte, error)
	CheckCurrentState() error
	Deploy() error
}

type ResourceConfig interface {
	New() (Resource, error)
}

func ResourceFromYaml(c []byte, r ResourceConfig) (Resource, error) {
	if err := yaml.Unmarshal(c, &r); err != nil {
		return nil, fmt.Errorf("failed to parse input YAML: %w", err)
	}

	return r.New()
}
