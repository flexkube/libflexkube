package flexkube

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
	// Don't preallocate, as then the diff shows diff between
	// nil and empty slice.
	var m []types.Mount //nolint:prealloc

	if i == nil {
		return m
	}

	j := i.([]interface{})

	for _, v := range j {
		l := v.(map[string]interface{})

		m = append(m, types.Mount{
			Source:      l["source"].(string),
			Target:      l["target"].(string),
			Propagation: l["propagation"].(string),
		})
	}

	return m
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
