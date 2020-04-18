package flexkube

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/flexkube/libflexkube/pkg/controlplane"
	"github.com/flexkube/libflexkube/pkg/types"
)

func resourceControlplane() *schema.Resource {
	return &schema.Resource{
		Create:        resourceCreate(controlplaneUnmarshal),
		Read:          resourceRead(controlplaneUnmarshal),
		Delete:        controlplaneDestroy,
		Update:        resourceCreate(controlplaneUnmarshal),
		CustomizeDiff: controlplaneDiff,
		Schema: withCommonFields(map[string]*schema.Schema{
			"ssh":                     sshSchema(false),
			"api_server_address":      optionalString(false),
			"api_server_port":         optionalInt(false),
			"common":                  controlplaneCommonSchema(),
			"kube_apiserver":          kubeAPIServerSchema(),
			"kube_scheduler":          kubeSchedulerSchema(),
			"kube_controller_manager": kubeControllerManagerSchema(),
		}),
	}
}

// controlplaneDiff is a workaround for Terraform issue, where it doesn't mark map fields
// as newComputed, when they are updated using SetNew.
// It's also a workaround, that if generated certificate changes, Terraform provides it as an empty string
// which means, we cannot create new controlplane object to update the configuration fields etc, which
// all results in inconsistent plan. To allow at least adding and removing members, we simply override the
// "state" field diff for the user by simply setting it to NewComputed, as we cannot go down deeper.
//
// See issues:
// - https://github.com/hashicorp/terraform-plugin-sdk/issues/371
// - https://github.com/flexkube/libflexkube/issues/48
func controlplaneDiff(d *schema.ResourceDiff, m interface{}) error {
	rd := reflect.ValueOf(d).Elem()
	rdiff := rd.FieldByName("diff")
	diff := reflect.NewAt(rdiff.Type(), unsafe.Pointer(rdiff.UnsafeAddr())).Elem().Interface().(*terraform.InstanceDiff) // #nosec G103

	if len(diff.Attributes) > 0 {
		keys := []string{"state", "config_yaml", "state_sensitive", "state_yaml"}
		for _, k := range keys {
			if err := d.SetNewComputed(k); err != nil {
				return fmt.Errorf("failed setting new computed for field %q: %w", k, err)
			}
		}
	}

	return resourceDiff(controlplaneUnmarshal)(d, m)
}

func controlplaneUnmarshal(d getter, includeState bool) types.ResourceConfig {
	c := &controlplane.Controlplane{
		APIServerAddress: d.Get("api_server_address").(string),
		APIServerPort:    d.Get("api_server_port").(int),
	}

	if includeState {
		s := getState(d)
		c.State = &s
	}

	if d, ok := d.GetOk("ssh"); ok && len(d.([]interface{})) == 1 {
		c.SSH = sshUnmarshal(d.([]interface{})[0])
	}

	if d, ok := d.GetOk("common"); ok && len(d.([]interface{})) == 1 {
		c.Common = controlplaneCommonUnmarshal(d.([]interface{})[0])
	}

	if d, ok := d.GetOk("kube_apiserver"); ok && len(d.([]interface{})) == 1 {
		c.KubeAPIServer = kubeAPIServerUnmarshal(d.([]interface{})[0])
	}

	if d, ok := d.GetOk("kube_controller_manager"); ok && len(d.([]interface{})) == 1 {
		c.KubeControllerManager = kubeControllerManagerUnmarshal(d.([]interface{})[0])
	}

	if d, ok := d.GetOk("kube_scheduler"); ok && len(d.([]interface{})) == 1 {
		c.KubeScheduler = kubeSchedulerUnmarshal(d.([]interface{})[0])
	}

	return c
}

func controlplaneDestroy(d *schema.ResourceData, m interface{}) error {
	c := controlplaneUnmarshal(d, true)

	// TODO Perhaps ResourceConfig should support generic destroying?
	c.(*controlplane.Controlplane).Destroy = true

	// Validate the configuration.
	r, err := c.New()
	if err != nil {
		return fmt.Errorf("failed creating resource: %w", err)
	}

	// Get current state of the containers.
	if err := r.CheckCurrentState(); err != nil {
		return fmt.Errorf("failed checking current state: %w", err)
	}

	// Deploy changes.
	deployErr := r.Deploy()

	// If deployment succeeded, we are done.
	if deployErr == nil {
		d.SetId("")

		return nil
	}

	return saveState(d, r.Containers().ToExported().PreviousState, controlplaneUnmarshal, deployErr)
}
