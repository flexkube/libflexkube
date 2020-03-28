package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/controlplane"
	"github.com/flexkube/libflexkube/pkg/types"
)

func controlplaneCommonSchema() *schema.Schema {
	return optionalBlock(false, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"kubernetes_ca_certificate":  optionalString(false),
			"front_proxy_ca_certificate": optionalString(false),
			"image":                      optionalString(false),
		}
	})
}

func controlplaneCommonUnmarshal(i interface{}) *controlplane.Common {
	c := &controlplane.Common{}

	// If optional block is not defined, return empty struct.
	if i == nil {
		return c
	}

	// If optional block is defined, but has no values defined, return empty struct as well.
	j, ok := i.(map[string]interface{})
	if !ok || len(j) == 0 {
		return c
	}

	c.KubernetesCACertificate = types.Certificate(j["kubernetes_ca_certificate"].(string))
	c.FrontProxyCACertificate = types.Certificate(j["front_proxy_ca_certificate"].(string))
	c.Image = j["image"].(string)

	return c
}
