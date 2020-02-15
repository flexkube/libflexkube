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
	GetRuntimeAddress() string
	SetRuntimeAddress(string)
	Create() (InstanceInterface, error)
	FromStatus() (InstanceInterface, error)
}

// InstanceInterface represents containerInstance capabilities.
type InstanceInterface interface {
	Status() (types.ContainerStatus, error)
	Read(srcPath []string) ([]*types.File, error)
	Copy(files []*types.File) error
	Stat(paths []string) (map[string]*os.FileMode, error)
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
	// Status of the container
	Status types.ContainerStatus `json:"status"`
	// Runtime stores configuration for various container runtimes
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
	// Contains common information between container and containerInstance
	base
	// Optional container status
	status types.ContainerStatus
}

// container represents created container. It guarantees that container status is initialised.
type containerInstance struct {
	// Contains common information between container and containerInstance
	base
	// Status of the container
	status types.ContainerStatus
}

// base contains basic information about created and not created container.
type base struct {
	// Runtime config which will be used when starting the container.
	config types.ContainerConfig
	// Container runtime which will be used to manage the container
	runtime       runtime.Runtime
	runtimeConfig runtime.Config
}

// New creates new instance of container from Container and validates it's configuration
// It also validates container runtime configuration.
func New(c *Container) (Interface, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("container configuration validation failed: %w", err)
	}

	nc := &container{
		base{
			config:        c.Config,
			runtimeConfig: c.Runtime.Docker,
		},
		c.Status,
	}
	if err := nc.selectRuntime(); err != nil {
		return nil, fmt.Errorf("unable to determine container runtime: %w", err)
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

	if c.Runtime.Docker == nil {
		return fmt.Errorf("docker runtime must be set")
	}

	// TODO check runtime configurations here
	return nil
}

// ToInstance returns containerInstance directly from Container
func (c *Container) ToInstance() (InstanceInterface, error) {
	container, err := New(c)
	if err != nil {
		return nil, err
	}

	return container.FromStatus()
}

// selectRuntime returns container runtime configured for container
//
// It returns error if container runtime configuration is invalid
func (c *container) selectRuntime() error {
	// TODO once we add more runtimes, if there is more than one defined, return an error here
	r, err := c.runtimeConfig.New()
	if err != nil {
		return fmt.Errorf("selecting container runtime failed: %w", err)
	}

	c.runtime = r

	return nil
}

// GetRuntimeAddress returns connection address for configured runtime. This can be used,
// for example when connection needs to be proxied.
func (c *container) GetRuntimeAddress() string {
	return c.runtimeConfig.GetAddress()
}

// SetRuntimeAddress sets connection address for configured runtime. This can be used,
// for example when connection needs to be proxied.
func (c *container) SetRuntimeAddress(a string) {
	c.runtimeConfig.SetAddress(a)
}

// Create creates container container from it's definition
func (c *container) Create() (InstanceInterface, error) {
	id, err := c.runtime.Create(&c.config)
	if err != nil {
		return nil, fmt.Errorf("creating container failed: %w", err)
	}

	return &containerInstance{
		base{
			config:        c.config,
			runtime:       c.runtime,
			runtimeConfig: c.runtimeConfig,
		},
		types.ContainerStatus{
			ID: id,
		},
	}, nil
}

// FromStatus creates containerInstance from previously restored status
func (c *container) FromStatus() (InstanceInterface, error) {
	if c.status.ID == "" {
		return nil, fmt.Errorf("can't create container instance from status without id: %+v", c.status)
	}

	return &containerInstance{
		base:   c.base,
		status: c.status,
	}, nil
}

// ReadState reads state of the container from container runtime and returns it to the user
func (c *containerInstance) Status() (types.ContainerStatus, error) {
	return c.runtime.Status(c.status.ID)
}

// UpdateStatus updates status of exported Container struct. This function is primarily used
// to reduce the boilerplate in helper functions, which allow to perform operations directly on
// Container struct.
func (c *containerInstance) updateStatus(container *Container) error {
	s, err := c.Status()
	if err != nil {
		return err
	}

	container.Status = s

	return nil
}

// Read reads given path from the container and returns reader with TAR format with file content
func (c *containerInstance) Read(srcPath []string) ([]*types.File, error) {
	return c.runtime.Read(c.status.ID, srcPath)
}

// Copy takes output path and TAR reader as arguments and extracts this TAR archive into container
func (c *containerInstance) Copy(files []*types.File) error {
	return c.runtime.Copy(c.status.ID, files)
}

// Stat checks if given path exists on the container and if yes, returns information whether
// it is file, or directory etc.
func (c *containerInstance) Stat(paths []string) (map[string]*os.FileMode, error) {
	return c.runtime.Stat(c.status.ID, paths)
}

// Start starts the container container
func (c *containerInstance) Start() error {
	return c.runtime.Start(c.status.ID)
}

// Stop stops the container container
func (c *containerInstance) Stop() error {
	return c.runtime.Stop(c.status.ID)
}

// Delete removes container container
func (c *containerInstance) Delete() error {
	return c.runtime.Delete(c.status.ID)
}

// UpdateStatus reads container existing status and updates it by communicating with container daemon
// This is a helper function, which simplifies calling containerInstance.Status() from Container.
func (c *Container) UpdateStatus() error {
	ci, err := c.ToInstance()
	if err != nil {
		return err
	}

	return ci.(*containerInstance).updateStatus(c)
}

// Start starts existing Container and updates it's status
func (c *Container) Start() error {
	ci, err := c.ToInstance()
	if err != nil {
		return err
	}

	if err := ci.Start(); err != nil {
		return err
	}

	return ci.(*containerInstance).updateStatus(c)
}

// Stop stops existing Container and updates it's status
func (c *Container) Stop() error {
	ci, err := c.ToInstance()
	if err != nil {
		return err
	}

	if err := ci.Stop(); err != nil {
		return err
	}

	return ci.(*containerInstance).updateStatus(c)
}

// Create creates container and gets it's status
func (c *Container) Create() error {
	nc, err := New(c)
	if err != nil {
		return err
	}

	ci, err := nc.Create()
	if err != nil {
		return err
	}

	return ci.(*containerInstance).updateStatus(c)
}

// Delete removes container and removes it's status
func (c *Container) Delete() error {
	ci, err := c.ToInstance()
	if err != nil {
		return err
	}

	if err := ci.Delete(); err != nil {
		return err
	}

	c.Status.ID = ""

	return nil
}

// Read takes file path as an argument and reads this file from the container
func (c *Container) Read(srcPath []string) ([]*types.File, error) {
	ci, err := c.ToInstance()
	if err != nil {
		return nil, err
	}

	return ci.Read(srcPath)
}

// Copy creates a file in desired path in the container
func (c *Container) Copy(files []*types.File) error {
	ci, err := c.ToInstance()
	if err != nil {
		return err
	}

	return ci.Copy(files)
}

// Stat checks if files exist in the container. It returns map of files mode for each requested file.
func (c *Container) Stat(paths []string) (map[string]*os.FileMode, error) {
	ci, err := c.ToInstance()
	if err != nil {
		return nil, err
	}

	return ci.Stat(paths)
}

// Exists returns true, if the container exists.
func (c *Container) Exists() bool {
	return c.Status.ID != ""
}

// IsRunning returns true, if container exists and it's running.
func (c *Container) IsRunning() bool {
	return c.Exists() && c.Status.Status == "running"
}
