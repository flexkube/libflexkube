package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
	"github.com/flexkube/libflexkube/pkg/types"
)

func resourceAPILoadBalancerPool() *schema.Resource {
	return &schema.Resource{
		Create:        resourceCreate(apiLoadBalancersUnmarshal),
		Read:          resourceRead(apiLoadBalancersUnmarshal),
		Delete:        resourceDelete(apiLoadBalancersUnmarshal, "api_load_balancer"),
		Update:        resourceCreate(apiLoadBalancersUnmarshal),
		CustomizeDiff: resourceDiff(apiLoadBalancersUnmarshal),
		Schema: withCommonFields(map[string]*schema.Schema{
			"image":             optionalString(false),
			"ssh":               sshSchema(false),
			"servers":           optionalStringList(false),
			"api_load_balancer": apiLoadBalancerSchema(),
			"name":              optionalString(false),
			"host_config_path":  optionalString(false),
			"bind_address":      optionalString(false),
		}),
	}
}

func apiLoadBalancersUnmarshal(d getter, includeState bool) types.ResourceConfig {
	servers := []string{}

	if i, ok := d.GetOk("servers"); ok {
		s := i.([]interface{})

		for _, v := range s {
			servers = append(servers, v.(string))
		}
	}

	cc := &apiloadbalancer.APILoadBalancers{
		Image:            d.Get("image").(string),
		Servers:          servers,
		APILoadBalancers: apiLoadBalancerUnmarshal(d.Get("api_load_balancer")),
		Name:             d.Get("name").(string),
		HostConfigPath:   d.Get("host_config_path").(string),
		BindAddress:      d.Get("bind_address").(string),
	}

	if includeState {
		cc.State = getState(d)
	}

	if d, ok := d.GetOk("ssh"); ok && len(d.([]interface{})) == 1 {
		cc.SSH = sshUnmarshal(d.([]interface{})[0])
	}

	return cc
}
