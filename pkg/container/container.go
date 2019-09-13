package container

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container/runtime"
	"github.com/invidian/etcd-ariadnes-thread/pkg/container/runtime/docker"

	"github.com/pkg/errors"
)

// Container represents public, serializable version of the container object.
//
// It should be used for persisting and restoring container state with combination
// with New(), which make sure that the configuration is actually correct.
type Container struct {
	// Stores runtime configuration of the container.
	Config runtime.Config `json:"config" yaml:"config"`
	// Status of the container
	Status *runtime.Status `json:"status" yaml:"status"`
	// Name of the container runtime to use. If not specified, container runtime
	// will be automatically determined based on configuration options. If no configuration
	// options are specified, "docker" will be chosen.
	RuntimeName string `json:"runtime_name,omitempty" yaml:"runtime_name,omitempty"`
}

// container represents validated version of Container object, which contains all requires
// information for instantiating (by calling Create()).
type container struct {
	// Contains common information between container and containerInstance
	base
	// Optional container status
	status *runtime.Status
}

// container represents created container. It guarantees that container status is initialised.
type containerInstance struct {
	// Contains common information between container and containerInstance
	base
	// Status of the container
	status runtime.Status
}

// base contains basic information about created and not created container.
type base struct {
	// Runtime config which will be used when starting the container.
	config runtime.Config
	// Container runtime which will be used to manage the container
	runtime runtime.Runtime
	// name of container runtime to use
	runtimeName string
}

// New creates new instance of container from Container and validates it's configuration
func New(c *Container) (*container, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrap(err, "container configuration validation failed")
	}
	nc := &container{
		base{
			config: runtime.Config{
				Name:  c.Config.Name,
				Image: c.Config.Image,
			},
			runtimeName: c.RuntimeName,
		},
		c.Status,
	}
	if err := nc.selectRuntime(); err != nil {
		return nil, errors.Wrap(err, "unable to determine container runtime")
	}
	return nc, nil
}

// Validate validates container configuration
func (c *Container) Validate() error {
	if c.Config.Name == "" {
		return fmt.Errorf("name must be set")
	}
	if c.Config.Image == "" {
		return fmt.Errorf("image must be set")
	}

	if !runtime.IsRuntimeSupported(c.RuntimeName) {
		return fmt.Errorf("configured runtime '%s' is not supported", c.RuntimeName)
	}

	return nil
}

// selectRuntime returns container runtime configured for container
//
// It returns error if container runtime configuration is invalid
func (c *container) selectRuntime() error {
	switch runtime.GetRuntimeName(c.runtimeName) {
	case "docker":
		r, err := docker.New(&docker.Docker{})
		if err != nil {
			return errors.Wrap(err, "selecting container runtime failed")
		}
		c.runtime = r
	default:
		return fmt.Errorf("not supported container runtime: %s", c.runtimeName)
	}
	return nil
}

// Create creates container container from it's defintion
func (c *container) Create() (*containerInstance, error) {
	id, err := c.runtime.Create(&c.config)
	if err != nil {
		return nil, errors.Wrap(err, "creating container")
	}

	return &containerInstance{
		base{
			config:      c.config,
			runtime:     c.runtime,
			runtimeName: c.runtimeName,
		},
		runtime.Status{
			ID: id,
		},
	}, nil
}

// FromStatus creates containerInstance from previously restored status
func (c *container) FromStatus() (*containerInstance, error) {
	if c.status == nil || c.status.ID == "" {
		return nil, fmt.Errorf("can't create container instance from invalid status")
	}
	return &containerInstance{
		base:   c.base,
		status: *c.status,
	}, nil
}

// ReadState reads state of the container from container runtime
//
// TODO Should we store the status here or return it instead?
func (container *containerInstance) Status() (*runtime.Status, error) {
	status, err := container.runtime.Status(container.status.ID)
	if err != nil {
		return nil, errors.Wrap(err, "getting container status failed")
	}
	if status == nil {
		return nil, fmt.Errorf("container '%s' does not exist", container.status.ID)
	}

	container.status = *status

	return status, nil
}

// Start starts the container container
func (container *containerInstance) Start() error {
	return container.runtime.Start(container.status.ID)
}

// Stop stops the container container
func (container *containerInstance) Stop() error {
	return container.runtime.Stop(container.status.ID)
}

// Delete removes container container
//
// TODO should we also clean up host directories here?
func (container *containerInstance) Delete() error {
	return container.runtime.Delete(container.status.ID)
}
