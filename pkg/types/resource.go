// Package types provides reusable structs and interfaces used across libflexkube, which
// can also be used by external projects.
package types

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/container"
)

// Resource interface defines common functionality between Flexkube resources like kubelet pool
// or static controlplane, which allows to manage group of containers.
type Resource interface {
	// StateToYaml converts resource's containers state into YAML format and returns it to the user,
	// so it can be persisted, e.g. to the file.
	StateToYaml() ([]byte, error)

	// CheckCurrentState iterates over containers defined in the state, checks if they exist, are
	// running etc and writes to containers current state. This allows then to compare current state
	// of the containers with desired state, using Containers() method, to check if there are any
	// pending changes to cluster configuration.
	//
	// Calling CheckCurrentState is required before calling Deploy(), to ensure, that Deploy() executes
	// correct actions.
	CheckCurrentState() error

	// Deploy creates configured containers.
	//
	// CheckCurrentState() must be called before calling Deploy(), otherwise error will be returned.
	Deploy() error

	// Containers gives access to the ContainersInterface from the resource, which allows accessing
	// methods like DesiredState() and ToExported(), which can be used to calculate pending changes
	// to the resource configuration.
	Containers() container.ContainersInterface
}

// ResourceConfig interface defines common functionality between all Flexkube resource configurations.
type ResourceConfig interface {
	// New creates new Resource object from given configuration and ensures, that the configuration
	// is valid.
	New() (Resource, error)

	// Validate validates the configuration.
	Validate() error
}

// ResourceFromYaml allows to create any resource instance from YAML configuration.
func ResourceFromYaml(c []byte, r ResourceConfig) (Resource, error) {
	if err := yaml.Unmarshal(c, &r); err != nil {
		return nil, fmt.Errorf("failed to parse input YAML: %w", err)
	}

	return r.New()
}
