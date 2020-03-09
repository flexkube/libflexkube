package types

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/container"
)

// Resource interface defines flexkube resource like kubelet pool or static controlplane.
type Resource interface {
	StateToYaml() ([]byte, error)
	CheckCurrentState() error
	Deploy() error
	Containers() container.ContainersInterface
}

// ResourceConfig interface defines flexkube resource configuration functionality.
type ResourceConfig interface {
	New() (Resource, error)
}

// ResourceFromYaml allows to create any resource instance from YAML configuration.
func ResourceFromYaml(c []byte, r ResourceConfig) (Resource, error) {
	if err := yaml.Unmarshal(c, &r); err != nil {
		return nil, fmt.Errorf("failed to parse input YAML: %w", err)
	}

	return r.New()
}
