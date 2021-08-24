// Package container allows to run and manage multiple containers across multiple hosts,
// by talking directly to the container runtime on local or remote hosts.
package container

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container/types"
)

// ContainersInterface represents capabilities of containers struct.
type ContainersInterface interface {
	// CheckCurrentState iterates over containers defined in the state, checks if they exist, are
	// running etc and writes to containers current state. This allows then to compare current state      // of the containers with desired state, using Containers() method, to check if there are any
	// pending changes to cluster configuration.
	//
	// Calling CheckCurrentState is required before calling Deploy(), to ensure, that Deploy() executes
	// correct actions.
	CheckCurrentState() error

	// Deploy creates configured containers.
	//
	// CheckCurrentState() must be called before calling Deploy(), otherwise error will be returned.
	Deploy() error

	// StateToYaml converts resource's containers state into YAML format and returns it to the user,
	// so it can be persisted, e.g. to the file.
	StateToYaml() ([]byte, error)

	// ToExported converts unexported containers struct into exported one, which can be then
	// serialized and persisted.
	ToExported() *Containers

	// DesiredState returns desired state of configured containers.
	//
	// Desired state differs from
	// exported or user-defined desired state, as it will have container IDs filled from the
	// previous state.
	//
	// All returned containers will also have status set to running, as this is always the desired
	// state of the container.
	//
	// Having those fields modified allows to minimize the difference when comparing previous state
	// and desired state.
	DesiredState() ContainersState
}

// Containers allow to orchestrate and update multiple containers spread
// across multiple hosts and update their configurations.
type Containers struct {
	// PreviousState stores previous state of the containers, which should be obtained and persisted
	// after containers modifications.
	PreviousState ContainersState `json:"previousState,omitempty"`

	// DesiredState is a user-defined desired containers configuration.
	DesiredState ContainersState `json:"desiredState,omitempty"`
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

// New validates Containers configuration and returns container object, which can be
// deployed.
func (c *Containers) New() (ContainersInterface, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("validating containers configuration: %w", err)
	}

	// Validate already checks for errors, so we can skip checking here.
	previousState, _ := c.PreviousState.New()
	desiredState, _ := c.DesiredState.New()

	return &containers{
		previousState: previousState.(containersState),
		desiredState:  desiredState.(containersState),
	}, nil
}

// Validate validates Containers struct and all structs used underneath.
func (c *Containers) Validate() error {
	var errors util.ValidateErrors

	if c == nil {
		errors = append(errors, fmt.Errorf("containers must be defined"))

		return errors.Return()
	}

	if (c.PreviousState == nil && c.DesiredState == nil) || (len(c.PreviousState) == 0 && len(c.DesiredState) == 0) {
		errors = append(errors, fmt.Errorf("either current state or desired state must be defined"))
	}

	if _, err := c.PreviousState.New(); err != nil {
		errors = append(errors, fmt.Errorf("validating previous state failed: %w", err))
	}

	if _, err := c.DesiredState.New(); err != nil {
		errors = append(errors, fmt.Errorf("validating desired state failed: %w", err))
	}

	return errors.Return()
}

// CheckCurrentState iterates over containers defined in the state, checks if they exist, are
// running etc and writes to containers current state. This allows then to compare current state
// of the containers with desired state, using Containers() method, to check if there are any
// pending changes to cluster configuration.
//
// Calling CheckCurrentState is required before calling Deploy(), to ensure, that Deploy() executes
// correct actions.
func (c *Containers) CheckCurrentState() error {
	containers, err := c.New()
	if err != nil {
		return fmt.Errorf("creating containers configuration: %w", err)
	}

	if err := containers.CheckCurrentState(); err != nil {
		return fmt.Errorf("checking current state of the containers: %w", err)
	}

	*c = *containers.ToExported()

	return nil
}

// Deploy creates configured containers.
//
// CheckCurrentState() must be called before calling Deploy(), otherwise error will be returned.
func (c *Containers) Deploy() error {
	containers, err := c.New()
	if err != nil {
		return fmt.Errorf("initializing containers: %w", err)
	}

	// TODO Deploy shouldn't refresh the state. However, due to how we handle exported/unexported
	// structs to enforce validation of objects, we lose current state, as we want it to be computed.
	// On the other hand, maybe it's a good thing to call it once we execute. This way we could compare
	// the plan user agreed to execute with plan calculated right before the execution and fail early if they
	// differ.
	// This is similar to what Terraform is doing and may cause planning to run several times, so it may require
	// some optimization.
	// Alternatively we can have serializable plan and a knob in execute command to control whether we should
	// make additional validation or not.
	if err := containers.CheckCurrentState(); err != nil {
		return fmt.Errorf("checking current state: %w", err)
	}

	if err := containers.Deploy(); err != nil {
		return fmt.Errorf("deploying: %w", err)
	}

	*c = *containers.ToExported()

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

	f := filesToUpdate(*d, r)

	err := d.Configure(f)

	if err != nil && reflect.DeepEqual(f, filesToUpdate(*d, r)) {
		return fmt.Errorf("no files has been updated: %w", err)
	}

	// If current state does not exist, simply replace it with desired state.
	if r == nil {
		c.currentState[n] = d
		r = d
	}

	// Update current state config files map.
	r.configFiles = d.configFiles

	if err != nil {
		return fmt.Errorf("updating configuration: %w", err)
	}

	return nil
}

// ensureRunning makes sure that given container is running.
func ensureRunning(c *hostConfiguredContainer) error {
	if c == nil {
		return fmt.Errorf("can't start non-existing container")
	}

	if c.container.Status().Running() {
		return nil
	}

	return c.Start()
}

func (c *containers) ensureExists(n string) error {
	r := c.currentState[n]
	if r != nil && r.container.Status().Exists() {
		return nil
	}

	fmt.Printf("Creating new container '%s'\n", n)

	d := c.desiredState[n]

	err := c.desiredState.CreateAndStart(n)

	// Container creation failed and it does not exist, meaning state is clean.
	if err != nil && !d.container.Status().Exists() {
		return fmt.Errorf("creating container: %w", err)
	}

	// Even if CreateAndStart failed, update current state. This makes the process more robust,
	// for example when creation succeeded (so container ID got assigned), but starting failed (as
	// for example requested port is already in use), so on the next run, engine won't try to create
	// the container again (as that would fail, because container of the same name already exists).

	// If current state does not exist, simply replace it with desired state.
	if r == nil {
		c.currentState[n] = d
		r = d
	}

	// After new container is created, add it to current state, so it can be returned to the user.
	*r.container.Status() = *d.container.Status()

	if err != nil {
		return fmt.Errorf("starting container: %w", err)
	}

	return nil
}

// isUpdatable determines if given container can be updated.
func (c *containers) isUpdatable(n string) error {
	// Container which currently does not exist can't be updated, only created.
	if _, ok := c.currentState[n]; !ok {
		return fmt.Errorf("can't update non-existing container %q", n)
	}

	// Container which is suppose to be removed shouldn't be updated.
	if _, ok := c.desiredState[n]; !ok {
		return fmt.Errorf("can't update container %q, which is scheduler for removal", n)
	}

	return nil
}

// diffHost compares host fields of the container and returns it's diff.
//
// If the container cannot be updated, error is returned.
func (c *containers) diffHost(n string) (string, error) {
	if err := c.isUpdatable(n); err != nil {
		return "", fmt.Errorf("can't diff container: %w", err)
	}

	return cmp.Diff(c.currentState[n].host, c.desiredState[n].host), nil
}

// recreate is a helper, which removes container from current state and creates new one from
// desired state.
func (c *containers) recreate(n string) error {
	if err := c.currentState.RemoveContainer(n); err != nil {
		return fmt.Errorf("removing old container to recreate it: %w", err)
	}

	c.currentState[n] = nil

	err := c.desiredState.CreateAndStart(n)

	c.currentState[n] = c.desiredState[n]

	if err != nil {
		return fmt.Errorf("creating and starting new container %q: %w", n, err)
	}

	return nil
}

// ensureHost makes sure container is running on the right host.
//
// If host configuration changes, existing container will be removed and new one will be created.
//
// TODO This might be an overkill. e.g. changing SSH key for deployment will re-create all containers.
func (c *containers) ensureHost(n string) error {
	diff, err := c.diffHost(n)
	if err != nil {
		return fmt.Errorf("checking host diff: %w", err)
	}

	if diff == "" {
		return nil
	}

	fmt.Printf("Detected host configuration drift '%s'\n", n)
	fmt.Printf("  Diff: %v\n", util.ColorizeDiff(diff))

	return c.recreate(n)
}

// diffContainer compares container fields of the container and returns it's diff.
//
// If the container cannot be updated, error is returned.
func (c *containers) diffContainer(n string) (string, error) {
	if err := c.isUpdatable(n); err != nil {
		return "", fmt.Errorf("can't diff container: %w", err)
	}

	cd := cmp.Diff(c.currentState[n].container.Config(), c.desiredState[n].container.Config())
	rcd := cmp.Diff(c.currentState[n].container.RuntimeConfig(), c.desiredState[n].container.RuntimeConfig())

	return cd + rcd, nil
}

// ensureContainer makes sure container configuration is up to date.
//
// If container configuration changes, existing container will be removed and new one will be created.
func (c *containers) ensureContainer(n string) error {
	diff, err := c.diffContainer(n)
	if err != nil {
		return fmt.Errorf("checking container diff: %w", err)
	}

	if diff == "" {
		return nil
	}

	fmt.Printf("Detected container configuration drift '%s'\n", n)
	fmt.Printf("  Diff: %v\n", util.ColorizeDiff(diff))

	return c.recreate(n)
}

// hasUpdates return bool if there are any pending configuration changes to the container.
func (c *containers) hasUpdates(n string) (bool, error) {
	diffHost, err := c.diffHost(n)
	if err != nil {
		return false, fmt.Errorf("checking host diff: %w", err)
	}

	f := filesToUpdate(*c.desiredState[n], c.currentState[n])

	diffContainer, err := c.diffContainer(n)
	if err != nil {
		return false, fmt.Errorf("checking container diff: %w", err)
	}

	return diffHost != "" || len(f) != 0 || diffContainer != "", nil
}

func (c *containers) ensureCurrentContainer(n string, r hostConfiguredContainer) (*hostConfiguredContainer, error) {
	// Gather facts about the container..
	exists := r.container.Status().Exists()
	_, isDesired := c.desiredState[n]
	hasUpdates := false

	if err := c.isUpdatable(n); err == nil {
		// If container is updatable, check if it has any updates pending.
		u, err := c.hasUpdates(n)
		if err != nil {
			return &r, fmt.Errorf("checking if container has pending updates: %w", err)
		}

		hasUpdates = u
	}

	// Container is gone, remove it from current state, so it will be scheduled for recreation.
	if !exists {
		delete(c.currentState, n)

		return nil, nil
	}

	// If container exist, is desired or has no pending updates, make sure it's running.
	if exists && isDesired && !hasUpdates {
		return &r, ensureRunning(&r)
	}

	return &r, nil
}

// ensureNewContainer handles configuring and creating new containers.
func (c *containers) ensureNewContainer(i string) error {
	if _, existingContainer := c.currentState[i]; existingContainer {
		return nil
	}

	if err := c.ensureConfigured(i); err != nil {
		return fmt.Errorf("configuring container %q: %w", i, err)
	}

	if err := c.ensureExists(i); err != nil {
		return fmt.Errorf("creating new container %q: %w", i, err)
	}

	return nil
}

func (c *containers) ensureUpToDate(i string) error {
	// Update containers on hosts.
	// This can move containers between hosts, but NOT the data.
	if err := c.ensureHost(i); err != nil {
		return fmt.Errorf("updating host configuration of container %q: %w", i, err)
	}

	if err := c.ensureConfigured(i); err != nil {
		return fmt.Errorf("updating configuration for container %q: %w", i, err)
	}

	if err := c.ensureContainer(i); err != nil {
		return fmt.Errorf("updating container %q: %w", i, err)
	}

	return nil
}

// updateExistingContainer handles updating existing containers. It either removes them
// if they are not needed anymore or makes sure that their configuration is up to date.
func (c *containers) updateExistingContainers() error {
	for i := range c.currentState {
		if _, exists := c.desiredState[i]; !exists {
			if err := c.currentState.RemoveContainer(i); err != nil {
				return fmt.Errorf("removing old container: %w", err)
			}

			continue
		}

		if err := c.ensureUpToDate(i); err != nil {
			return fmt.Errorf("ensuring, that container %q is up to date: %w", i, err)
		}
	}

	return nil
}

// Deploy checks for containers configuration drifts and tries to reach desired state.
//
// TODO we should break down this function into smaller functions
// TODO add planning, so it is possible to inspect what will be done
// TODO currently we only compare previous configuration with new configuration.
// We should also read runtime parameters and confirm that everything is according
// to the spec.
func (c *containers) Deploy() error {
	if c.currentState == nil {
		return fmt.Errorf("can't execute without knowing current state of the containers")
	}

	fmt.Println("Checking for stopped and missing containers")

	for n, r := range c.currentState {
		d, err := c.ensureCurrentContainer(n, *r)

		if d != nil {
			c.currentState[n] = d
		}

		if err != nil {
			return fmt.Errorf("handling existing container %q: %w", n, err)
		}
	}

	fmt.Println("Configuring and creating new containers")

	for i := range c.desiredState {
		if err := c.ensureNewContainer(i); err != nil {
			return fmt.Errorf("creating new container %q: %w", i, err)
		}
	}

	fmt.Println("Updating existing containers")

	return c.updateExistingContainers()
}

// FromYaml allows to load containers configuration and state from YAML format.
func FromYaml(c []byte) (ContainersInterface, error) {
	containers := &Containers{}
	if err := yaml.Unmarshal(c, &containers); err != nil {
		return nil, fmt.Errorf("parsing input YAML: %w", err)
	}

	cl, err := containers.New()
	if err != nil {
		return nil, fmt.Errorf("creating containers object: %w", err)
	}

	return cl, nil
}

// StateToYaml dumps current state as previousState in exported format,
// which can be serialized and stored.
func (c *containers) StateToYaml() ([]byte, error) {
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

// DesiredState returns desired state enhanced with current state, to highlight
// important configuration changes from user perspective.
func (c *containers) DesiredState() ContainersState {
	d := c.desiredState.Export()

	for h := range d {
		// If container already exist, append it's ID to desired state to reduce the diff.
		id := ""

		cs, ok := c.previousState[h]
		if ok && cs.container.Status().ID != "" {
			id = cs.container.Status().ID
		}

		// Make sure, that desired state has correct status. Container should always be running
		// and optionally, we also set the ID of already existing container. If there are changes
		// to the container, it will get new ID anyway, but user does not care about this change,
		// so we can hide it this way from the diff.
		d[h].Container.Status = &types.ContainerStatus{
			Status: "running",
			ID:     id,
		}
	}

	return d
}

// Containers implement types.Resource interface.
func (c *containers) Containers() ContainersInterface {
	return c
}
