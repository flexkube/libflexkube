package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
)

func apiLoadBalancerUnmarshal(i interface{}) []apiloadbalancer.APILoadBalancer {
	j := i.([]interface{})

	albs := []apiloadbalancer.APILoadBalancer{}

	for _, k := range j {
		alb := apiloadbalancer.APILoadBalancer{}

		if k == nil {
			albs = append(albs, alb)

			continue
		}

		t := k.(map[string]interface{})

		alb.Image = t["image"].(string)
		alb.Name = t["name"].(string)
		alb.HostConfigPath = t["host_config_path"].(string)
		alb.BindAddress = t["bind_address"].(string)

		if v, ok := t["servers"]; ok && len(v.([]interface{})) > 0 {
			s := v.([]interface{})

			for _, va := range s {
				alb.Servers = append(alb.Servers, va.(string))
			}
		}

		if v, ok := t["host"]; ok && len(v.([]interface{})) == 1 {
			alb.Host = hostUnmarshal(v.([]interface{})[0])
		}

		albs = append(albs, alb)
	}

	return albs
}

func apiLoadBalancerSchema() *schema.Schema {
	return requiredList(false, false, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"image":            optionalString(false),
				"host":             hostSchema(false),
				"servers":          optionalStringList(false),
				"name":             optionalString(false),
				"host_config_path": optionalString(false),
				"bind_address":     optionalString(false),
			},
		}
	})
}
