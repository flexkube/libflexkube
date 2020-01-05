package container

import (
	"fmt"
	"reflect"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
)

// Containers allow to orchestrate and update multiple containers spread
// across multiple hosts and update their configurations.
type Containers struct {
	// PreviousState stores previous state of the containers, which should be obtained and persisted
	// after containers modifications.
	PreviousState ContainersState `json:"previousState" yaml:"previousState"`
	// DesiredState is a user-defined desired containers configuration.
	DesiredState ContainersState `json:"desiredState" yaml:"desiredState"`
}

// containers is a validated version of the Containers, which allows user to perform operations on them
// like planning, getting status etc.
type containers struct {
	// previousState is a previous state of the containers, given by user.
	previousState containersState
	// currentState stores current state of the containers. It is fed by calling Refresh() function.
	currentState containersState
	// resiredState is a user-defined desired containers configuration after validation.
	desiredState containersState
}

// New validates Containers configuration and returns "executable" containers object.
func (c *Containers) New() (*containers, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate containers configuration: %w", err)
	}

	// Validate already checks for errors, so we can skip checking here.
	previousState, _ := c.PreviousState.New()
	desiredState, _ := c.DesiredState.New()

	return &containers{
		previousState: previousState,
		desiredState:  desiredState,
	}, nil
}

// Validate validates Containers struct and all structs used underneath.
func (c *Containers) Validate() error {
	if c == nil || (c.PreviousState == nil && c.DesiredState == nil) {
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
// and then updates state of all containers.
func (c *containers) CheckCurrentState() error {
	if c.currentState == nil {
		// We just assign the pointer, but it's fine, since we don't need previous
		// state anyway.
		// TODO we could keep previous state to inform user, that some external changes happened since last run
		c.currentState = c.previousState
	}

	return c.currentState.CheckState()
}

// filesToUpdate returns list of files, which needs to be updated, based on the current state of the container.
// If the file is missing or it's content is not the same as desired content, it will be added to the list.
func filesToUpdate(d hostConfiguredContainer, c *hostConfiguredContainer) []string {
	// If current state does not exist, just return all files.
	if c == nil {
		return util.KeysStringMap(d.configFiles)
	}

	files := []string{}

	// Loop over desired config files and check if they exist.
	for p, content := range d.configFiles {
		if currentContent, exists := c.configFiles[p]; !exists || content != currentContent {
			// TODO convert all prints to logging, so we can add more verbose information too
			fmt.Printf("Detected configuration drift for file '%s'\n", p)
			fmt.Printf("  current: \n%+v\n", currentContent)
			fmt.Printf("  desired: \n%+v\n", content)

			files = append(files, p)
		}
	}

	return files
}

// ensureConfigured makes sure that all desired configuration files are correct.
func (c *containers) ensureConfigured(n string) error {
	d := c.desiredState[n]

	// Container won't be needed anyway, so skip everything.
	if d == nil {
		return nil
	}

	r := c.currentState[n]

	if err := d.Configure(filesToUpdate(*d, r)); err != nil {
		return fmt.Errorf("failed creating config files: %w", err)
	}

	// If current state does not exist, simply replace it with desired state.
	if r == nil {
		c.currentState[n] = d
		r = d
	}

	// Update current state config files map.
	r.configFiles = d.configFiles

	return nil
}

// ensureRunning makes sure that given container is running.
func ensureRunning(c *hostConfiguredContainer) error {
	if c == nil {
		return fmt.Errorf("can't start non-existing container")
	}

	if c.container.IsRunning() {
		return nil
	}

	return c.Start()
}

func (c *containers) ensureExists(n string) error {
	r := c.currentState[n]
	if r != nil && r.container.Exists() {
		return nil
	}

	fmt.Printf("Creating new container '%s'\n", n)

	if err := c.desiredState.CreateAndStart(n); err != nil {
		return fmt.Errorf("failed creating new container: %w", err)
	}

	d := c.desiredState[n]

	// If current state does not exist, simply replace it with desired state.
	if r == nil {
		c.currentState[n] = d
		r = d
	}

	// After new container is created, add it to current state, so it can be returned to the user.
	r.container.Status = d.container.Status

	return nil
}

// ensureHost makes sure container is running on the right host.
//
// If host configuration changes, existing container will be removed and new one will be created.
//
// TODO This might be an overkill. e.g. changing SSH key for deployment will re-create all containers.
func (c *containers) ensureHost(n string) error {
	r := c.currentState[n]
	if r == nil {
		return fmt.Errorf("can't update non-existing container")
	}

	dr := c.desiredState[n]

	// Don't update containers scheduled for removal and containers with unchanged configuration
	if dr == nil || reflect.DeepEqual(r.host, dr.host) {
		return nil
	}

	fmt.Printf("Detected host configuration drift '%s'\n", n)
	fmt.Printf("  From: %+v\n%+v\n%+v\n", r.host, r.host.DirectConfig, r.host.SSHConfig)
	fmt.Printf("  To:   %+v\n%+v\n%+v\n", dr.host, dr.host.DirectConfig, dr.host.SSHConfig)

	if err := c.currentState.RemoveContainer(n); err != nil {
		return fmt.Errorf("failed removing old container: %w", err)
	}

	if err := c.desiredState.CreateAndStart(n); err != nil {
		return fmt.Errorf("failed creating new container: %w", err)
	}

	// After new container is created, add it to current state, so it can be returned to the user.
	r.host = dr.host

	return nil
}

// Execute checks for containers configuration drifts and tries to reach desired state.
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

	fmt.Println("Checking for stopped and missing containers")

	for n, r := range c.currentState {
		if err := ensureRunning(r); err != nil {
			return fmt.Errorf("failed to start stopped container: %w", err)
		}

		// Container is gone, we need to re-create it.
		if !r.container.Exists() {
			delete(c.currentState, n)
		}
	}

	fmt.Println("Configuring and creating new containers")

	for i := range c.desiredState {
		if err := c.ensureConfigured(i); err != nil {
			return fmt.Errorf("failed configuring container %s: %w", i, err)
		}

		if err := c.ensureExists(i); err != nil {
			return fmt.Errorf("failed creating new container %s: %w", i, err)
		}
	}

	fmt.Println("Updating existing containers")

	for i := range c.currentState {
		// Update containers on hosts.
		// This can move containers between hosts, but NOT the data.
		if err := c.ensureHost(i); err != nil {
			return fmt.Errorf("failed updating host configuration of container %s: %w", i, err)
		}

		if err := c.ensureConfigured(i); err != nil {
			return fmt.Errorf("failed updating configuration for container %s: %w", i, err)
		}

		if _, exists := c.desiredState[i]; exists {
			continue
		}

		if err := c.currentState.RemoveContainer(i); err != nil {
			return fmt.Errorf("failed removing old container: %w", err)
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

// CurrentStateToYaml dumps current state as previousState in exported format,
// which can be serialized and stored.
func (c *containers) CurrentStateToYaml() ([]byte, error) {
	containers := &Containers{
		PreviousState: c.previousState.Export(),
	}

	return yaml.Marshal(containers)
}

// ToExported converts containers struct to exported Containers.
func (c *containers) ToExported() *Containers {
	return &Containers{
		PreviousState: c.previousState.Export(),
		DesiredState:  c.desiredState.Export(),
	}
}
