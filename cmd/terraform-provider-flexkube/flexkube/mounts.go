package flexkube //nolint:dupl

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

func mountsMarshal(c []types.Mount) []interface{} {
	i := []interface{}{}

	for _, v := range c {
		i = append(i, map[string]interface{}{
			"source":      v.Source,
			"target":      v.Target,
			"propagation": v.Propagation,
		})
	}

	return i
}

func mountsUnmarshal(i interface{}) []types.Mount {
	j := i.([]interface{})

	p := []types.Mount{}

	for _, v := range j {
		l := v.(map[string]interface{})

		p = append(p, types.Mount{
			Source:      l["source"].(string),
			Target:      l["target"].(string),
			Propagation: l["propagation"].(string),
		})
	}

	return p
}

func mountsSchema(computed bool) *schema.Schema {
	return optionalMap(computed, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"source":      optionalString(computed),
				"target":      optionalString(computed),
				"propagation": optionalString(computed),
			},
		}
	})
}
