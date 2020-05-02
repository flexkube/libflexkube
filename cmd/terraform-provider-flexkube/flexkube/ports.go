package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

func portMapMarshal(c []types.PortMap) []interface{} {
	i := []interface{}{}

	for _, v := range c {
		i = append(i, map[string]interface{}{
			"ip":       v.IP,
			"port":     v.Port,
			"protocol": v.Protocol,
		})
	}

	return i
}

func portMapUnmarshal(i interface{}) []types.PortMap {
	j := i.([]interface{})

	// Don't preallocate, as then the diff shows diff between
	// nil and empty slice.
	var p []types.PortMap //nolint:prealloc

	for _, v := range j {
		l := v.(map[string]interface{})

		p = append(p, types.PortMap{
			IP:       l["ip"].(string),
			Port:     l["port"].(int),
			Protocol: l["protocol"].(string),
		})
	}

	return p
}

func portMapSchema(computed bool) *schema.Schema {
	return optionalMap(computed, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ip":       optionalString(computed),
				"port":     optionalInt(computed),
				"protocol": optionalString(computed),
			},
		}
	})
}
