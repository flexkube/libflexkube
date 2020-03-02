package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
)

func runtimeMarshal(rc container.RuntimeConfig) interface{} {
	return []interface{}{
		map[string]interface{}{
			"docker": dockerMarshal(*rc.Docker),
		},
	}
}

func runtimeUnmarshal(i interface{}) container.RuntimeConfig {
	rc := container.RuntimeConfig{
		Docker: docker.DefaultConfig(),
	}

	if i == nil {
		return rc
	}

	j, ok := i.(map[string]interface{})
	if !ok || len(j) == 0 {
		return rc
	}

	if d, ok := j["docker"]; ok && len(d.([]interface{})) == 1 {
		rc.Docker = dockerUnmarshal(d.([]interface{})[0])
	}

	return rc
}

func runtimeSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"docker": dockerSchema(computed),
		}
	})
}
