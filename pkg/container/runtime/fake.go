package runtime

import (
	"fmt"
	"os"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

// Fake is a fake runtime client, which can be used for testing.
type Fake struct {
	// CreateF will be Create by method.
	CreateF func(config *types.ContainerConfig) (string, error)

	// DeleteF will be called by Delete method.
	DeleteF func(id string) error

	// StartF will be called by Start method.
	StartF func(id string) error

	// StatusF will be called by Status method.
	StatusF func(id string) (types.ContainerStatus, error)

	// StopF will be called by Stop method.
	StopF func(id string) error

	// CopyF will be called by Copy method.
	CopyF func(id string, files []*types.File) error

	// ReadF will be called by Read method.
	ReadF func(id string, srcPath []string) ([]*types.File, error)

	// StatF will be called by Stat method.
	StatF func(id string, paths []string) (map[string]os.FileMode, error)
}

// Create mocks runtime Create().
func (f Fake) Create(config *types.ContainerConfig) (string, error) {
	return f.CreateF(config)
}

// Delete mocks runtime Delete().
func (f Fake) Delete(id string) error {
	return f.DeleteF(id)
}

// Start mocks runtime Start().
func (f Fake) Start(id string) error {
	return f.StartF(id)
}

// Status mocks runtime Status().
func (f Fake) Status(id string) (types.ContainerStatus, error) {
	return f.StatusF(id)
}

// Stop mocks runtime Stop().
func (f Fake) Stop(id string) error {
	return f.StopF(id)
}

// Copy mocks runtime Copy().
func (f Fake) Copy(id string, files []*types.File) error {
	return f.CopyF(id, files)
}

// Read mocks runtime Read().
func (f Fake) Read(id string, srcPath []string) ([]*types.File, error) {
	return f.ReadF(id, srcPath)
}

// Stat mocks runtime Stat().
func (f Fake) Stat(id string, paths []string) (map[string]os.FileMode, error) {
	return f.StatF(id, paths)
}

// FakeConfig is a Fake runtime configuration struct.
type FakeConfig struct {
	// Runtime holds container runtime to return by New() method.
	Runtime Runtime

	// Address will be used for GetAddress and SetAddress methods.
	Address string
}

// GetAddress implements runtime.Config interface.
func (c FakeConfig) GetAddress() string {
	return c.Address
}

// SetAddress implements runtime.Config interface.
func (c FakeConfig) SetAddress(a string) {
	c.Address = a
}

// New implements runtime.Config interface.
func (c FakeConfig) New() (Runtime, error) {
	if c.Runtime == nil {
		return nil, fmt.Errorf("no runtime defined")
	}

	return c.Runtime, nil
}
