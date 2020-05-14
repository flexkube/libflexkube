package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/controlplane"
)

func kubeSchedulerSchema() *schema.Schema {
	return optionalBlock(false, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"common":     controlplaneCommonSchema(),
			"host":       hostSchema(false),
			"kubeconfig": kubeconfigSchema(),
		}
	})
}

func kubeSchedulerUnmarshal(i interface{}) controlplane.KubeScheduler {
	c := controlplane.KubeScheduler{}

	// If optional block is not defined, return empty struct.
	if i == nil {
		return c
	}

	// If optional block is defined, but has no values defined, return empty struct as well.
	j, ok := i.(map[string]interface{})
	if !ok || len(j) == 0 {
		return c
	}

	if v, ok := j["common"]; ok && len(v.([]interface{})) == 1 {
		c.Common = controlplaneCommonUnmarshal(v.([]interface{})[0])
	}

	if v, ok := j["host"]; ok && len(v.([]interface{})) == 1 {
		h := hostUnmarshal(v.([]interface{})[0])
		c.Host = &h
	}

	if v, ok := j["kubeconfig"]; ok && len(v.([]interface{})) == 1 {
		c.Kubeconfig = kubeconfigUnmarshal(v.([]interface{})[0])
	}

	return c
}
