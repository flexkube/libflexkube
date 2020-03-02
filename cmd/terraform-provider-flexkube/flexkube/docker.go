package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
)

func dockerMarshal(c docker.Config) interface{} {
	return []interface{}{
		map[string]interface{}{
			"host": c.Host,
		},
	}
}

func dockerUnmarshal(i interface{}) *docker.Config {
	c := docker.DefaultConfig()

	if i == nil {
		return c
	}

	j, ok := i.(map[string]interface{})
	if !ok || len(j) == 0 {
		return c
	}

	if h, ok := j["host"]; ok {
		c.Host = h.(string)
	}

	return c
}

func dockerSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"host": optionalString(computed),
		}
	})
}
