package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/controlplane"
	"github.com/flexkube/libflexkube/pkg/types"
)

func resourceControlplane() *schema.Resource {
	return &schema.Resource{
		Create:        resourceCreate(controlplaneUnmarshal),
		Read:          resourceRead(controlplaneUnmarshal),
		Delete:        resourceDelete(controlplaneUnmarshal, "hosts"),
		Update:        resourceCreate(controlplaneUnmarshal),
		CustomizeDiff: resourceDiff(controlplaneUnmarshal),
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

func controlplaneUnmarshal(d getter, includeState bool) types.ResourceConfig {
	c := &controlplane.Controlplane{
		APIServerAddress: d.Get("api_server_address").(string),
		APIServerPort:    d.Get("api_server_port").(int),
	}

	if includeState {
		c.State = getState(d)
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
