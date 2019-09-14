package container

import (
	"fmt"
)

// HostConfiguredContainer represents single container, running on remote host with it's configuration files
type HostConfiguredContainer struct {
	Container Container `json:"container" yaml:"container"`
	// Host host.Host `json:"host" yaml:"host"`
	// ConfigFiles
}

func (m *HostConfiguredContainer) New() (*hostConfiguredContainer, error) {
	if err := m.Container.Validate(); err != nil {
		return nil, fmt.Errorf("failed to valide container configuration: %w", err)
	}

	return &hostConfiguredContainer{
		container: m.Container,
	}, nil
}

type hostConfiguredContainer struct {
	name      string
	container Container
}

// Configure copies configuration on target host
func (m *hostConfiguredContainer) Configure() error {
	return nil
}

// Create creates new container on target host
func (m *hostConfiguredContainer) Create() error {
	return m.container.Create()
}

// Status updates container status
func (m *hostConfiguredContainer) Status() error {
	return m.container.UpdateStatus()
}

// Start starts created container
func (m *hostConfiguredContainer) Start() error {
	return m.container.Start()
}

// Stop stops created container
func (m *hostConfiguredContainer) Stop() error {
	return m.container.Stop()
}

// Delete removes node's data and removes the container
func (m *hostConfiguredContainer) Delete() error {
	return m.container.Delete()
}
