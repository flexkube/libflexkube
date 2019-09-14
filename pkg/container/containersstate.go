package container

import (
	"fmt"
)

// ContainersState represents states of multiple containers
type ContainersState map[string]*HostConfiguredContainer

// containerState is a validated version of ContainersState, which can be user to perform operations
type containersState map[string]*hostConfiguredContainer

// New validates ContainersState struct and returns operational containerState.
func (s ContainersState) New() (containersState, error) {
	if s == nil {
		s = ContainersState{}
	}

	state := containersState{}

	for name, container := range s {
		m, err := container.New()
		if err != nil {
			return nil, err
		}
		state[name] = m
	}

	return state, nil
}

// CheckState updates the state of all previously configured containers
func (s containersState) CheckState() error {
	for _, m := range s {
		if err := m.Status(); err != nil {
			return err
		}
	}

	return nil
}

// RemoveContainer removes the container by ID
func (s containersState) RemoveContainer(containerName string) error {
	if _, exists := s[containerName]; !exists {
		return fmt.Errorf("can't remove non-existing container")
	}
	if err := s[containerName].Stop(); err != nil {
		return fmt.Errorf("failed stopping container: %w", err)
	}
	if err := s[containerName].Delete(); err != nil {
		return fmt.Errorf("failed removing container: %w", err)
	}
	delete(s, containerName)
	return nil
}

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
