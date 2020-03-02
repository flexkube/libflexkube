package flexkube

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/types"
)

func resourceContainers() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainersCreate,
		// Update should be exactly the same operation as Create.
		Update:        resourceContainersCreate,
		Read:          resourceContainersRead,
		Delete:        resourceContainersDelete,
		CustomizeDiff: resourceContainersDiff,
		Schema: map[string]*schema.Schema{
			// Configuration specified by the user.
			"container": hostConfiguredContainerSchema(false, false),
			// Existing state of the configuration, used for operations.
			// This filed is entirely marked as sensitive, to prevent leaking secrets when running
			// plan etc.
			"state_sensitive": hostConfiguredContainerSchema(true, true),
			// This state has secrets stripped out and will be presented as a diff to the user.
			"state":       hostConfiguredContainerSchema(true, false),
			"state_yaml":  sensitiveString(true),
			"config_yaml": sensitiveString(true),
		},
	}
}

func containersUnmarshal(d getter, refresh bool) (container.ContainersInterface, string, error) {
	// resourceContainersDiff will update 'state_sensitive' with desired state, so we can show
	// user, that both 'state' and 'state_sensitive' are going to change.
	psi, _ := d.GetChange("state_sensitive")

	cc := &container.Containers{
		PreviousState: containersStateUnmarshal(psi),
		DesiredState:  containersStateUnmarshal(d.Get("container")),
	}

	cd := &container.Containers{
		DesiredState: containersStateUnmarshal(d.Get("container")),
	}

	b, err := yaml.Marshal(cd)
	if err != nil {
		return nil, "", fmt.Errorf("failed serializing generated configuration: %w", err)
	}

	c, err := cc.New()
	if err != nil {
		return nil, "", fmt.Errorf("failed creating containers configuration: %w", err)
	}

	if !refresh {
		return c, string(b), nil
	}

	// Get current state of the containers.
	return c, string(b), c.CheckCurrentState()
}

func resourceContainersCreate(d *schema.ResourceData, m interface{}) error {
	// Create Containers object.
	c, _, err := containersUnmarshal(d, true)
	if err != nil {
		return fmt.Errorf("failed creating containers configuration: %w", err)
	}

	// Deploy changes.
	deployErr := c.Deploy()

	// If there was at least one container created, set the ID to mark, that resource has been at least partially
	// created.
	// If the ID is already set, then also don't update it, as there is no need for that.
	if d.IsNewResource() && len(c.ToExported().PreviousState) != 0 {
		state, err := c.StateToYaml()
		if err != nil {
			return fmt.Errorf("failed serializing resource state to YAML: %w", err)
		}

		d.SetId(sha256sum(state))
	}

	return saveState(d, c, deployErr)
}

func resourceContainersRead(d *schema.ResourceData, m interface{}) error {
	c, _, err := containersUnmarshal(d, true)
	if err != nil {
		return fmt.Errorf("failed creating containers configuration: %w", err)
	}

	// If there is nothing in the current state, mark the resource as destroyed.
	if len(c.ToExported().PreviousState) == 0 {
		d.SetId("")
	}

	return saveState(d, c, nil)
}

func resourceContainersDelete(d *schema.ResourceData, m interface{}) error {
	// Reset user configuration to indicate, that we destroy everything.
	if err := d.Set("container", []interface{}{}); err != nil {
		return fmt.Errorf("failed resetting container field to trigger a destroy: %w", err)
	}

	// Create Containers object.
	c, _, err := containersUnmarshal(d, true)
	if err != nil {
		return fmt.Errorf("failed creating containers configuration: %w", err)
	}

	// Deploy changes.
	deployErr := c.Deploy()

	// If deployment succeeded, we are done.
	if deployErr == nil {
		d.SetId("")

		return nil
	}

	return saveState(d, c, deployErr)
}

func resourceContainersDiff(d *schema.ResourceDiff, v interface{}) error {
	c, cy, err := containersUnmarshal(d, false)
	if err != nil {
		return fmt.Errorf("failed creating containers configuration: %w", err)
	}

	if err := d.SetNew("config_yaml", cy); err != nil {
		return fmt.Errorf("failed writing configuration in YAML format: %w", err)
	}

	s := containersStateMarshal(c.Containers().DesiredState(), false)

	if err := d.SetNew("state_sensitive", s); err != nil {
		return fmt.Errorf("failed writing desired state: %w", err)
	}

	s = containersStateMarshal(c.Containers().DesiredState(), true)

	return d.SetNew("state", s)
}

func saveState(d *schema.ResourceData, c types.Resource, origErr error) error {
	// Save new state to the Terraform.
	s := containersStateMarshal(c.Containers().ToExported().PreviousState, false)

	if err := d.Set("state_sensitive", s); err != nil {
		return fmt.Errorf("operation failed and failed to persist containers state: %w", origErr)
	}

	if err := saveStateYaml(d, c); err != nil {
		return fmt.Errorf("failed saving state in YAML format %w", err)
	}

	s = containersStateMarshal(c.Containers().ToExported().PreviousState, true)

	if err := d.Set("state", s); err != nil {
		return fmt.Errorf("operation failed and failed to persist user containers state: %w", origErr)
	}

	return origErr
}

func saveStateYaml(d *schema.ResourceData, c types.Resource) error {
	cc := &container.Containers{
		PreviousState: c.Containers().ToExported().PreviousState,
	}

	b, err := yaml.Marshal(cc)
	if err != nil {
		return fmt.Errorf("failed serializing state: %w", err)
	}

	if err := d.Set("state_yaml", string(b)); err != nil {
		return fmt.Errorf("failed persisting state in YAML format to Terraform state: %w", err)
	}

	return nil
}
