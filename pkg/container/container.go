package container

import (
	"fmt"
	"io"
	"os"

	"github.com/invidian/flexkube/pkg/container/runtime"
	"github.com/invidian/flexkube/pkg/container/runtime/docker"
	"github.com/invidian/flexkube/pkg/container/types"
)

// Container represents public, serializable version of the container object.
//
// It should be used for persisting and restoring container state with combination
// with New(), which make sure that the configuration is actually correct.
type Container struct {
	// Stores runtime configuration of the container.
	Config types.ContainerConfig `json:"config" yaml:"config"`
	// Status of the container
	Status *types.ContainerStatus `json:"status" yaml:"status"`
	// Runtime stores configuration for various container runtimes
	Runtime RuntimeConfig `json:"runtime,omitempty" yaml:"runtime,omitempty"`
}

type RuntimeConfig struct {
	Docker *docker.ClientConfig `json:"docker,omitempty" yaml:"docker,omitempty"`
}

// container represents validated version of Container object, which contains all requires
// information for instantiating (by calling Create()).
type container struct {
	// Contains common information between container and containerInstance
	base
	// Optional container status
	status *types.ContainerStatus
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
	runtimeConfig runtime.RuntimeConfig
}

// New creates new instance of container from Container and validates it's configuration
// It also validates container runtime configuration.
func New(c *Container) (*container, error) {
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

func (c *container) GetRuntimeAddress() string {
	return c.runtimeConfig.GetAddress()
}

func (c *container) SetRuntimeAddress(a string) {
	c.runtimeConfig.SetAddress(a)
}

// Create creates container container from it's defintion
func (c *container) Create() (*containerInstance, error) {
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
func (c *container) FromStatus() (*containerInstance, error) {
	if c.status == nil {
		return nil, fmt.Errorf("can't create container instance from empty status")
	}
	if c.status != nil && c.status.ID == "" {
		return nil, fmt.Errorf("can't create container instance from status without id: %+v", c.status)
	}

	return &containerInstance{
		base:   c.base,
		status: *c.status,
	}, nil
}

// ReadState reads state of the container from container runtime and returns it to the user
func (container *containerInstance) Status() (*types.ContainerStatus, error) {
	status, err := container.runtime.Status(container.status.ID)
	if err != nil {
		return nil, fmt.Errorf("getting status for container '%s' failed: %w", container.config.Name, err)
	}
	return status, nil
}

// Read reads given path from the container and returns reader with TAR format with file content
func (c *containerInstance) Read(srcPath string) (io.ReadCloser, error) {
	return c.runtime.Read(c.status.ID, srcPath)
}

// Copy takes output path and TAR reader as arguments and extracts this TAR archive into container
func (c *containerInstance) Copy(dstPath string, content io.Reader) error {
	return c.runtime.Copy(c.status.ID, dstPath, content)
}

// Stat checks if given path exists on the container and if yes, returns information wheather
// it is file, or directory etc.
func (c *containerInstance) Stat(path string) (*os.FileMode, error) {
	s, err := c.runtime.Stat(c.status.ID, path)
	if err != nil {
		return nil, err
	}
	return s, err
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
func (container *containerInstance) Delete() error {
	return container.runtime.Delete(container.status.ID)
}

// UpdateStatus reads container existing status and updates it by communicating with container daemon
// This is a helper function, which simplifies calling containerInstance.Status() from Container.
// TODO look how to remove the boilerplate
func (container *Container) UpdateStatus() error {
	c, err := New(container)
	if err != nil {
		return err
	}
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}
	s, err := ci.Status()
	if err != nil {
		return err
	}
	container.Status = s
	return nil
}

// Start starts existing Container and updates it's status
func (container *Container) Start() error {
	c, err := New(container)
	if err != nil {
		return err
	}
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}
	if err := ci.Start(); err != nil {
		return err
	}
	s, err := ci.Status()
	if err != nil {
		return err
	}
	container.Status = s
	return nil
}

// Stop stops existing Container and updates it's status
func (container *Container) Stop() error {
	c, err := New(container)
	if err != nil {
		return err
	}
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}
	if err := ci.Stop(); err != nil {
		return err
	}
	s, err := ci.Status()
	if err != nil {
		return err
	}
	container.Status = s
	return nil
}

// Create creates container and gets it's status
func (container *Container) Create() error {
	c, err := New(container)
	if err != nil {
		return err
	}
	ci, err := c.Create()
	if err != nil {
		return err
	}
	s, err := ci.Status()
	if err != nil {
		return err
	}
	container.Status = s
	return nil
}

// Delete removes container and removes it's status
func (container *Container) Delete() error {
	c, err := New(container)
	if err != nil {
		return err
	}
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}
	if err := ci.Delete(); err != nil {
		return err
	}
	container.Status = nil
	return nil
}

func (container *Container) Read(srcPath string) (io.ReadCloser, error) {
	c, err := New(container)
	if err != nil {
		return nil, err
	}
	ci, err := c.FromStatus()
	if err != nil {
		return nil, err
	}
	return ci.Read(srcPath)
}

func (container *Container) Copy(dstPath string, content io.Reader) error {
	c, err := New(container)
	if err != nil {
		return err
	}
	ci, err := c.FromStatus()
	if err != nil {
		return err
	}
	return ci.Copy(dstPath, content)
}

func (container *Container) Stat(path string) (*os.FileMode, error) {
	c, err := New(container)
	if err != nil {
		return nil, err
	}
	ci, err := c.FromStatus()
	if err != nil {
		return nil, err
	}
	return ci.Stat(path)
}
