package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

func kubeconfigSchema() *schema.Schema {
	return requiredBlock(false, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"server":             optionalString(false),
				"ca_certificate":     optionalString(false),
				"client_certificate": requiredString(false),
				"client_key":         requiredSensitiveString(),
			},
		}
	})
}

func kubeconfigUnmarshal(i interface{}) client.Config {
	c := client.Config{}

	// If optional block is not defined, return empty struct.
	if i == nil {
		return c
	}

	// If optional block is defined, but has no values defined, return empty struct as well.
	j, ok := i.(map[string]interface{})
	if !ok || len(j) == 0 {
		return c
	}

	c.Server = j["server"].(string)
	c.CACertificate = types.Certificate(j["ca_certificate"].(string))
	c.ClientCertificate = types.Certificate(j["client_certificate"].(string))
	c.ClientKey = types.PrivateKey(j["client_key"].(string))

	return c
}
