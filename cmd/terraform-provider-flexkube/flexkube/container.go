package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
)

func containerMarshal(c container.Container) interface{} {
	m := map[string]interface{}{
		"config":  containerConfigMarshal(c.Config),
		"runtime": runtimeMarshal(c.Runtime),
	}

	if c.Status.ID != "" || c.Status.Status != "" {
		m["status"] = containerStatusMarshal(c.Status)
	}

	return []interface{}{m}
}

func containerUnmarshal(i interface{}) container.Container {
	j := i.(map[string]interface{})

	r := container.RuntimeConfig{
		Docker: docker.DefaultConfig(),
	}

	if rc, ok := j["runtime"]; ok && len(rc.([]interface{})) == 1 {
		r = runtimeUnmarshal(rc.([]interface{})[0])
	}

	c := container.Container{
		Config:  containerConfigUnmarshal(j["config"].([]interface{})[0]),
		Runtime: r,
	}

	if s, ok := j["status"]; ok && len(s.([]interface{})) == 1 {
		c.Status = containerStatusUnmarshal(s.([]interface{})[0])
	}

	return c
}

func containerSchema(computed bool) *schema.Schema {
	return requiredBlock(computed, func(computed bool) *schema.Resource {
		r := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"config":  containerConfigSchema(computed),
				"runtime": runtimeSchema(computed),
			},
		}

		// Status field should only be visible for computed state block and
		// not be configurable for the user.
		if computed {
			r.Schema["status"] = containerStatusSchema()
		}

		return r
	})
}
