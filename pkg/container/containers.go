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

// New validates Containers configuration and returns "executable" containers object
func (c *Containers) New() (*containers, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate containers configuration: %w", err)
	}

	// Validate already checks for errors, so we can skip checking here
	previousState, _ := c.PreviousState.New()
	desiredState, _ := c.DesiredState.New()

	return &containers{
		previousState: previousState,
		desiredState:  desiredState,
	}, nil
}

func (c *Containers) Validate() error {
	if c.PreviousState == nil && c.DesiredState == nil {
		return fmt.Errorf("either current state or desired state must be defined")
	}

	previousState, err := c.PreviousState.New()
	if err != nil {
		return err
	}

	desiredState, err := c.DesiredState.New()
	if err != nil {
		return err
	}

	if len(previousState) == 0 && len(desiredState) == 0 {
		return fmt.Errorf("either current state or desired state should have containers defined")
	}

	return nil
}

// CheckCurrentState copies previous state to current state, to mark, that it has been called at least once
// and then updates state of all containers
func (c *containers) CheckCurrentState() error {
	if c.currentState == nil {
		// We just assing the pointer, but it's fine, since we don't need previous
		// state anyway.
		// TODO we could keep previous state to inform user, that some external changes happened since last run
		c.currentState = c.previousState
	}
	return c.currentState.CheckState()
}

// Execute checks for containers configurion drifts and tries to reach desired state
//
// TODO we should break down this function into smaller functions
// TODO add planning, so it is possible to inspect what will be done
// TODO currently we only compare previous configuration with new configuration.
// We should also read runtime parameters and confirm that everything is according
// to the spec.
func (c *containers) Execute() error {
	if c.currentState == nil {
		return fmt.Errorf("can't execute without knowing current state of the containers")
	}

	fmt.Println("Checking for configuration files updates")
	for i, _ := range c.desiredState {
		// Loop over desired config files, check if they exist
		for p, content := range c.desiredState[i].configFiles {
			var currentContent string
			if _, exists := c.currentState[i]; exists {
				currentContent, exists := c.currentState[i].configFiles[p]
				// If file does not exist or it content differs, deploy it
				if exists && content == currentContent {
					continue
				}
			}
			// TODO convert all prints to logging, so we can add more verbose information too
			fmt.Printf("Detected configuration drift for file '%s'\n", p)
			fmt.Printf("  current: \n%+v\n", currentContent)
			fmt.Printf("  desired: \n%+v\n", content)
			if err := c.desiredState[i].Configure(p); err != nil {
				return err
			}
		}

		if _, exists := c.currentState[i]; !exists {
			c.currentState[i] = c.desiredState[i]
		}
		c.currentState[i].configFiles = c.desiredState[i].configFiles
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

	fmt.Println("Checking for host configuration updates")
	// Update containers on hosts
	// This can move containers between hosts, but NOT the data
	for i, _ := range c.currentState {
		// Don't update nodes sheduled for removal
		if _, exists := c.desiredState[i]; !exists {
			fmt.Printf("Skipping runtime configuration check for container '%s', as it will be removed\n", i)
			continue
		}

		if !reflect.DeepEqual(c.currentState[i].host, c.desiredState[i].host) {
			fmt.Printf("Detected host configuration drift '%s'\n", i)
			fmt.Printf("  From: %+v\n%+v\n%+v\n", c.currentState[i].host, c.currentState[i].host.DirectConfig, c.currentState[i].host.SSHConfig)
			fmt.Printf("  To:   %+v\n%+v\n%+v\n", c.desiredState[i].host, c.desiredState[i].host.DirectConfig, c.desiredState[i].host.SSHConfig)

			if err := c.currentState.RemoveContainer(i); err != nil {
				return fmt.Errorf("failed removing old container: %w", err)
			}
			if err := c.desiredState.CreateAndStart(i); err != nil {
				return fmt.Errorf("failed creating new container: %w", err)
			}
			// After new container is created, add it to current state, so it can be returned to the user
			c.currentState[i].host = c.desiredState[i].host
		}
	}

	fmt.Println("Checking for configuration updates on containers")
	// Update containers configurations
	for i, _ := range c.currentState {
		// Don't update containers sheduled for removal
		if _, exists := c.desiredState[i]; !exists {
			fmt.Printf("Skipping configuration check for container '%s', as it will be removed\n", i)
			continue
		}

		// If container configuration differs, re-create the container
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
	if err := yaml.Unmarshal(c, &containers); err != nil {
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
			Container:   m.container,
			Host:        m.host,
			ConfigFiles: m.configFiles,
		}
	}
	return yaml.Marshal(containers)
}
