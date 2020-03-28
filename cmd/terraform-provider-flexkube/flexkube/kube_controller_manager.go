package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/controlplane"
	"github.com/flexkube/libflexkube/pkg/types"
)

func kubeControllerManagerSchema() *schema.Schema {
	return requiredBlock(false, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"common":                      controlplaneCommonSchema(),
				"host":                        hostSchema(false),
				"kubeconfig":                  kubeconfigSchema(),
				"flex_volume_plugin_dir":      requiredString(false),
				"kubernetes_ca_key":           requiredSensitiveString(),
				"service_account_private_key": requiredSensitiveString(),
				"root_ca_certificate":         requiredString(false),
			},
		}
	})
}

func kubeControllerManagerUnmarshal(i interface{}) controlplane.KubeControllerManager {
	c := controlplane.KubeControllerManager{}

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

	c.FlexVolumePluginDir = j["flex_volume_plugin_dir"].(string)
	c.KubernetesCAKey = types.PrivateKey(j["kubernetes_ca_key"].(string))
	c.ServiceAccountPrivateKey = types.PrivateKey(j["service_account_private_key"].(string))
	c.RootCACertificate = types.Certificate(j["root_ca_certificate"].(string))

	return c
}
