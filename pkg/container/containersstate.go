package container

import (
	"fmt"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

const (
	// StatusMissing is a value, which is set to ContainerStatus.Status field,
	// if stored container ID is not found.
	StatusMissing = "gone"
)

// ContainersStateInterface represents 'constainersState' capabilities.
type ContainersStateInterface interface {
	// CheckState updates the state of all previously configured containers
	// and their configuration on the host
	CheckState() error

	// RemoveContainer removes the container by ID.
	RemoveContainer(containerName string) error

	// CreateAndStart is a helper, which creates and spawns given container.
	CreateAndStart(containerName string) error

	// Export converts unexported containersState to exported type, so it can be serialized and stored.
	Export() ContainersState
}

// ContainersState represents states of multiple containers.
type ContainersState map[string]*HostConfiguredContainer

// containerState is a validated version of ContainersState, which can be used to perform operations.
type containersState map[string]*hostConfiguredContainer

// New validates ContainersState struct and returns operational containerState.
func (s ContainersState) New() (ContainersStateInterface, error) {
	if s == nil {
		s = ContainersState{}
	}

	state := containersState{}

	for name, container := range s {
		m, err := container.New()
		if err != nil {
			return nil, err
		}

		state[name] = m.(*hostConfiguredContainer)
	}

	return state, nil
}

// CheckState updates the state of all previously configured containers
// and their configuration on the host.
func (s containersState) CheckState() error {
	for i, hcc := range s {
		if err := hcc.Status(); err != nil {
			return fmt.Errorf("checking container %q status: %w", i, err)
		}

		if hcc.container.Status().ID == "" {
			hcc.container.SetStatus(types.ContainerStatus{
				Status: StatusMissing,
			})
		}

		if err := hcc.ConfigurationStatus(); err != nil {
			return fmt.Errorf("checking container %q configuration status: %w", i, err)
		}
	}

	return nil
}

// RemoveContainer removes the container by ID.
func (s containersState) RemoveContainer(containerName string) error {
	if _, exists := s[containerName]; !exists {
		return fmt.Errorf("can't remove non-existing container")
	}

	status := s[containerName].container.Status()
	if status.Running() || status.Restarting() {
		if err := s[containerName].Stop(); err != nil {
			return fmt.Errorf("failed stopping container: %w", err)
		}
	}

	if s[containerName].container.Status().Exists() {
		if err := s[containerName].Delete(); err != nil {
			return fmt.Errorf("failed removing container: %w", err)
		}
	}

	delete(s, containerName)

	return nil
}

// CreateAndStart is a helper, which creates and spawns given container.
func (s containersState) CreateAndStart(containerName string) error {
	if _, exists := s[containerName]; !exists {
		return fmt.Errorf("can't create non-existing container")
	}

	if err := s[containerName].Create(); err != nil {
		return fmt.Errorf("failed creating new container: %w", err)
	}

	if err := s[containerName].Start(); err != nil {
		return fmt.Errorf("failed starting container: %w", err)
	}

	return nil
}

// Export converts unexported containersState to exported type, so it can be serialized and stored.
func (s containersState) Export() ContainersState {
	cs := ContainersState{}

	for i, m := range s {
		h := &HostConfiguredContainer{
			Container: Container{
				Config: m.container.Config(),
				Runtime: RuntimeConfig{
					Docker: m.container.RuntimeConfig().(*docker.Config),
				},
			},
			Host:        m.host,
			ConfigFiles: m.configFiles,
		}

		if s := m.container.Status(); s.ID != "" && s.Status != "" {
			h.Container.Status = s
		}

		if h.ConfigFiles == nil {
			h.ConfigFiles = map[string]string{}
		}

		cs[i] = h
	}

	return cs
}
