package container

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Containers allow to orchestrate and update multiple containers spread
// across multiple hosts and update their configurations.
type Containers struct {
	// PreviousState stores previous state of the containers, which should be obtained and persisted
	// after containers modifications.
	PreviousState ContainersState `json:"previousState:omitempty" yaml:"previousState,omitempty"`
	// DesiredState is a user-defined desired containers configuration.
	DesiredState ContainersState `json:"desiredState,omitempty" yaml:"desiredState,omitempty"`
}

// containers is a validated version of the Containers, which allows user to perform operations on them
// like planning, getting status etc.
type containers struct {
	// previousState is a previous state of the containers, given by user
	previousState containersState
	// currentState stores current state of the containers. It is fed by calling Refresh() function.
	currentState containersState
	// resiredState is a user-defined desired containers configuration after validation.
	desiredState containersState
}

// New validates Containers configuration and returns operational containers
func (c *Containers) New() (*containers, error) {
	if c.PreviousState == nil && c.DesiredState == nil {
		return nil, fmt.Errorf("either current state or desired state must be defined")
	}

	previousState, err := c.PreviousState.New()
	if err != nil {
		return nil, err
	}

	desiredState, err := c.DesiredState.New()
	if err != nil {
		return nil, err
	}

	if len(previousState) == 0 && len(desiredState) == 0 {
		return nil, fmt.Errorf("either current state or desired state should have containers defined")
	}

	return &containers{
		previousState: previousState,
		desiredState:  desiredState,
	}, nil
}

// CheckCurrentState copies previous state to current state, to mark, that it has been called at least once
// and then updates state of all containers
func (c *containers) CheckCurrentState() error {
	if c.currentState == nil {
		// We just assing the pointer, but it's fine, since we don't need previous
		// state anyway.
		c.currentState = c.previousState
	}
	return c.currentState.CheckState()
}

// Execute checks for containers configurion drifts and tries to reach desired state
//
// TODO we should break down this function into smaller functions
func (c *containers) Execute() error {
	if c.currentState == nil {
		return fmt.Errorf("can't execute without knowing current state of the containers")
	}

	fmt.Println("Checking for stopped containers")
	// Start stopped containers
	for i, _ := range c.currentState {
		if c.currentState[i].container.Status != nil && c.currentState[i].container.Status.Status != "running" {
			fmt.Printf("Starting stopped container '%s'\n", i)
			if err := c.currentState[i].Start(); err != nil {
				return fmt.Errorf("failed starting container: %w", err)
			}
		}
	}

	fmt.Println("Checking for missing containers to re-create")
	// Schedule missing containers for recreation
	for i, _ := range c.currentState {
		// Container is gone, we need to re-create it
		// TODO also remove from the containers
		if c.currentState[i].container.Status == nil {
			delete(c.currentState, i)
		}
	}

	fmt.Println("Checking for new containers to create")
	// Iterate over desired state to find which containers should be created
	for i, _ := range c.desiredState {
		// TODO Move this logic to function
		// Simple logic can stay inline, complex logic should go to function?
		if _, exists := c.currentState[i]; !exists {
			fmt.Printf("Creating new container '%s'\n", i)
			if err := c.desiredState.CreateAndStart(i); err != nil {
				return fmt.Errorf("failed creating new container: %w", err)
			}
			// After new container is created, add it to current state, so it can be returned to the user
			c.currentState[i] = c.desiredState[i]
		}
	}

	fmt.Println("Checking for configuration updates on containers")
	// Update containers to desired image version
	for i, _ := range c.currentState {
		// Don't update nodes sheduled for removal
		if _, exists := c.desiredState[i]; !exists {
			fmt.Printf("Skipping configuration check for container '%s', as it will be removed\n", i)
			continue
		}

		// If image differs, re-create the container
		// TODO compare image SHA from state rather than what we store
		if !reflect.DeepEqual(c.currentState[i].container.Config, c.desiredState[i].container.Config) {
			fmt.Printf("Updating container configuration '%s'\n", i)
			fmt.Printf("  From: %+v\n", c.currentState[i].container.Config)
			fmt.Printf("  To: %+v\n", c.desiredState[i].container.Config)

			if err := c.currentState.RemoveContainer(i); err != nil {
				return fmt.Errorf("failed removing old container: %w", err)
			}
			if err := c.desiredState.CreateAndStart(i); err != nil {
				return fmt.Errorf("failed creating new container: %w", err)
			}
			// After new container is created, add it to current state, so it can be returned to the user
			c.currentState[i] = c.desiredState[i]
		}
	}

	fmt.Println("Checking for old containers to remove")
	// Remove old containers
	for i, _ := range c.currentState {
		if _, exists := c.desiredState[i]; !exists {
			fmt.Printf("Removing old container '%s'\n", i)
			if err := c.currentState.RemoveContainer(i); err != nil {
				return fmt.Errorf("failed removing old container: %w", err)
			}
		}
	}

	return nil
}

// FromYaml allows to restore containers state from YAML.
func FromYaml(c []byte) (*containers, error) {
	containers := &Containers{}
	if err := yaml.Unmarshal(c, containers); err != nil {
		return nil, fmt.Errorf("failed to parse input yaml: %w", err)
	}
	cl, err := containers.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create containers object: %w", err)
	}

	return cl, nil
}

// ToYaml allows to dump containers state to YAML, so it can be restored later.
func (c *containers) ToYaml() ([]byte, error) {
	containers := &Containers{
		PreviousState: ContainersState{},
	}
	for i, m := range c.currentState {
		containers.PreviousState[i] = &HostConfiguredContainer{
			Container: m.container,
		}
	}
	return yaml.Marshal(containers)
}
