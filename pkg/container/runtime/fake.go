package runtime

import (
	"fmt"
	"os"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

// Fake is a fake runtime client, which can be used for testing.
type Fake struct {
	CreateF func(config *types.ContainerConfig) (string, error)
	DeleteF func(id string) error
	StartF  func(id string) error
	StatusF func(id string) (types.ContainerStatus, error)
	StopF   func(id string) error
	CopyF   func(id string, files []*types.File) error
	ReadF   func(id string, srcPath []string) ([]*types.File, error)
	StatF   func(id string, paths []string) (map[string]*os.FileMode, error)
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
func (f Fake) Stat(id string, paths []string) (map[string]*os.FileMode, error) {
	return f.StatF(id, paths)
}

type FakeConfig struct {
	Runtime Runtime
	Address string
}

func (c FakeConfig) GetAddress() string {
	return c.Address
}

func (c FakeConfig) SetAddress(a string) {
	c.Address = a
}

func (c FakeConfig) New() (Runtime, error) {
	if c.Runtime == nil {
		return nil, fmt.Errorf("no runtime defined")
	}

	return c.Runtime, nil
}
