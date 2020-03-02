package container

import (
	"fmt"
	"os"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

// Interface represents container capabilities.
type Interface interface {
	// Methods to instantiate Instance.
	Create() (InstanceInterface, error)
	FromStatus() (InstanceInterface, error)

	// Helpers.
	UpdateStatus() error
	Start() error
	Stop() error
	Delete() error
	Status() *types.ContainerStatus

	// Getters.
	Config() types.ContainerConfig
	RuntimeConfig() runtime.Config
	Runtime() runtime.Runtime

	// Setters.
	SetRuntime(runtime.Runtime)
	SetStatus(types.ContainerStatus)
}

// InstanceInterface represents containerInstance capabilities.
type InstanceInterface interface {
	Status() (types.ContainerStatus, error)
	Read(srcPath []string) ([]*types.File, error)
	Copy(files []*types.File) error
	Stat(paths []string) (map[string]os.FileMode, error)
	Start() error
	Stop() error
	Delete() error
}

// Container represents public, serializable version of the container object.
//
// It should be used for persisting and restoring container state with combination
// with New(), which make sure that the configuration is actually correct.
type Container struct {
	// Stores runtime configuration of the container.
	Config types.ContainerConfig `json:"config"`
	// Status of the container.
	Status types.ContainerStatus `json:"status"`
	// Runtime stores configuration for various container runtimes.
	Runtime RuntimeConfig `json:"runtime,omitempty"`
}

// RuntimeConfig is a collection of various runtime configurations which can be defined
// by user.
type RuntimeConfig struct {
	Docker *docker.Config `json:"docker,omitempty"`
}

// container represents validated version of Container object, which contains all requires
// information for instantiating (by calling Create()).
type container struct {
	// Contains common information between container and containerInstance.
	base
}

// container represents created container. It guarantees that container status is initialised.
type containerInstance struct {
	// Contains common information between container and containerInstance
	base
}

// base contains basic information about created and not created container.
type base struct {
	// Runtime config which will be used when starting the container.
	config types.ContainerConfig
	// Container runtime which will be used to manage the container
	runtime       runtime.Runtime
	runtimeConfig runtime.Config

	status types.ContainerStatus
}

// New creates new instance of container from Container and validates it's configuration
// It also validates container runtime configuration.
func (c *Container) New() (Interface, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("container configuration validation failed: %w", err)
	}

	nc := &container{
		base{
			config:        c.Config,
			runtimeConfig: c.Runtime.Docker,
			status:        c.Status,
		},
	}
	if err := nc.selectRuntime(); err != nil {
		return nil, fmt.Errorf("unable to determine container runtime: %w", err)
	}

	return nc, nil
}

// Validate validates container configuration.
func (c *Container) Validate() error {
	if c.Config.Name == "" {
		return fmt.Errorf("name must be set")
	}

	if c.Config.Image == "" {
		return fmt.Errorf("image must be set")
	}

	if c.Runtime.Docker == nil {
		return fmt.Errorf("docker runtime must be set")
	}

	// TODO check runtime configurations here
	return nil
}

// selectRuntime returns container runtime configured for container.
//
// It returns error if container runtime configuration is invalid.
func (c *container) selectRuntime() error {
	// TODO once we add more runtimes, if there is more than one defined, return an error here.
	r, err := c.runtimeConfig.New()
	if err != nil {
		return fmt.Errorf("selecting container runtime failed: %w", err)
	}

	c.runtime = r

	return nil
}

// Create creates container container from it's definition.
func (c *container) Create() (InstanceInterface, error) {
	id, err := c.runtime.Create(&c.config)
	if err != nil {
		return nil, fmt.Errorf("creating container failed: %w", err)
	}

	return &containerInstance{
		base{
			config:  c.config,
			runtime: c.runtime,
			status: types.ContainerStatus{
				ID: id,
			},
		},
	}, nil
}

// FromStatus creates containerInstance from stored status.
func (c *container) FromStatus() (InstanceInterface, error) {
	if !c.status.Exists() {
		return nil, fmt.Errorf("can't create container instance from status without id: %+v", c.status)
	}

	return &containerInstance{
		base: c.base,
	}, nil
}

// Config returns container config.
func (c *container) Config() types.ContainerConfig {
	return c.config
}

// Runtime returns container's configured runtime.
func (c *container) RuntimeConfig() runtime.Config {
	return c.runtimeConfig
}

func (c *container) UpdateStatus() error {
	ci, err := c.FromStatus()
	if err != nil {
		return fmt.Errorf("failed creating container instance: %w", err)
	}

	s, err := ci.Status()
	if err != nil {
		return fmt.Errorf("failed checking container status: %w", err)
	}

	c.status = s

	return nil
}

func (c *container) Status() *types.ContainerStatus {
	return &c.status
}

func (c *container) SetRuntime(r runtime.Runtime) {
	c.runtime = r
}

func (c *container) SetStatus(s types.ContainerStatus) {
	c.status = s
}

func (c *container) Runtime() runtime.Runtime {
	return c.runtime
}

// Start starts existing Container and updates it's status.
func (c *container) Start() error {
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}

	if err := ci.Start(); err != nil {
		return err
	}

	return c.UpdateStatus()
}

// Stop stops existing Container and updates it's status.
func (c *container) Stop() error {
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}

	if err := ci.Stop(); err != nil {
		return err
	}

	return c.UpdateStatus()
}

// Delete removes container and removes it's status.
func (c *container) Delete() error {
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}

	if err := ci.Delete(); err != nil {
		return err
	}

	c.Status().ID = ""

	return nil
}

// ReadState reads state of the container from container runtime and returns it to the user.
func (c *containerInstance) Status() (types.ContainerStatus, error) {
	return c.runtime.Status(c.status.ID)
}

// Read reads given path from the container and returns reader with TAR format with file content.
func (c *containerInstance) Read(srcPath []string) ([]*types.File, error) {
	return c.runtime.Read(c.status.ID, srcPath)
}

// Copy takes output path and TAR reader as arguments and extracts this TAR archive into container.
func (c *containerInstance) Copy(files []*types.File) error {
	return c.runtime.Copy(c.status.ID, files)
}

// Stat checks if given path exists on the container and if yes, returns information whether
// it is file, or directory etc.
func (c *containerInstance) Stat(paths []string) (map[string]os.FileMode, error) {
	return c.runtime.Stat(c.status.ID, paths)
}

// Start starts the container.
func (c *containerInstance) Start() error {
	return c.runtime.Start(c.status.ID)
}

// Stop stops the container.
func (c *containerInstance) Stop() error {
	return c.runtime.Stop(c.status.ID)
}

// Delete removes the container.
func (c *containerInstance) Delete() error {
	return c.runtime.Delete(c.status.ID)
}
