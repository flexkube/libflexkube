package container

import (
	"fmt"
	"os"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

// Interface represents container capabilities, which may or may not exist.
//
//nolint:interfacebloat // Perhaps at some point we can refactor this.
type Interface interface {
	// Create creates the container.
	Create() (InstanceInterface, error)

	// From status restores container instance from given status.
	FromStatus() (InstanceInterface, error)

	// UpdateStatus updates container status.
	UpdateStatus() error

	// Start starts the container.
	Start() error

	// Stop stops the container.
	Stop() error

	// Delete removes the container.
	Delete() error

	// Status returns container status.
	Status() *types.ContainerStatus

	// Config allows reading container configuration.
	Config() types.ContainerConfig

	// RuntimeConfig allows reading container runtime configuration.
	RuntimeConfig() runtime.Config

	// Runtime allows getting container runtime.
	Runtime() runtime.Runtime

	// SetRuntime allows overriding used container runtime.
	SetRuntime(newRuntime runtime.Runtime)

	// SetStatus allows overriding container status.
	SetStatus(newStatus types.ContainerStatus)
}

// InstanceInterface represents operations, which can be executed on existing
// container.
type InstanceInterface interface {
	// Status returns container status read from the configured container runtime.
	Status() (types.ContainerStatus, error)

	// Read reads content of the given file paths in the container.
	Read(srcPath []string) ([]*types.File, error)

	// Copy copies file into the container.
	Copy(files []*types.File) error

	// Stat checks if given files exist on the container and returns map of
	// file modes. If key is missing, it means file does not exist in the container.
	Stat(paths []string) (map[string]os.FileMode, error)

	// Start starts the container.
	Start() error

	// Stop stops the container.
	Stop() error

	// Delete deletes the container.
	Delete() error
}

// Container allows managing single container on directly reachable, configured container
// runtime, for example Docker using 'unix:///run/docker.sock' address.
type Container struct {
	// Config defines the properties of the container to manage.
	Config types.ContainerConfig `json:"config"`
	// Status stores container status. Setting status allows to manage existing containers
	// and e.g. removing them.
	Status *types.ContainerStatus `json:"status,omitempty"`
	// Runtime stores configuration for various container runtimes.
	Runtime RuntimeConfig `json:"runtime,omitempty"`
}

// RuntimeConfig is a collection of various runtime configurations which can be defined
// by user.
type RuntimeConfig struct {
	// Docker stores Docker runtime configuration.
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

// New creates new instance of container from Container and validates it's configuration.
// It also validates container runtime configuration.
func (c *Container) New() (Interface, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("container configuration validation failed: %w", err)
	}

	newContainer := &container{
		base{
			config:        c.Config,
			runtimeConfig: c.Runtime.Docker,
		},
	}

	if c.Status != nil {
		newContainer.base.status = *c.Status
	}

	if err := newContainer.selectRuntime(); err != nil {
		return nil, fmt.Errorf("determining container runtime: %w", err)
	}

	return newContainer, nil
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
		return fmt.Errorf("selecting container runtime: %w", err)
	}

	c.runtime = r

	return nil
}

// Create creates container from it's definition.
func (c *container) Create() (InstanceInterface, error) {
	containerID, err := c.runtime.Create(&c.config)
	if err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}

	return &containerInstance{
		base{
			config:  c.config,
			runtime: c.runtime,
			status: types.ContainerStatus{
				ID: containerID,
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
		return fmt.Errorf("creating container instance: %w", err)
	}

	s, err := ci.Status()
	if err != nil {
		return fmt.Errorf("checking container status: %w", err)
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
		return fmt.Errorf("getting containers instance from status: %w", err)
	}

	if err := ci.Start(); err != nil {
		return fmt.Errorf("starting container: %w", err)
	}

	return c.UpdateStatus()
}

// Stop stops existing Container and updates it's status.
func (c *container) Stop() error {
	ci, err := c.FromStatus()
	if err != nil {
		return fmt.Errorf("getting containers instance from status: %w", err)
	}

	if err := ci.Stop(); err != nil {
		return fmt.Errorf("stopping container: %w", err)
	}

	return c.UpdateStatus()
}

// Delete removes container and removes it's status.
func (c *container) Delete() error {
	ci, err := c.FromStatus()
	if err != nil {
		return fmt.Errorf("getting containers instance from status: %w", err)
	}

	if err := ci.Delete(); err != nil {
		return fmt.Errorf("deleting container: %w", err)
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
