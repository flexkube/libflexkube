package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

func containerStatusMarshal(s types.ContainerStatus) interface{} {
	return []interface{}{
		map[string]interface{}{
			"id":     s.ID,
			"status": s.Status,
		},
	}
}

func containerStatusUnmarshal(i interface{}) types.ContainerStatus {
	j := i.(map[string]interface{})

	return types.ContainerStatus{
		ID:     j["id"].(string),
		Status: j["status"].(string),
	}
}

func containerStatusSchema() *schema.Schema {
	return requiredBlock(true, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id":     requiredString(true),
				"status": requiredString(true),
			},
		}
	})
}
