package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/controlplane"
	"github.com/flexkube/libflexkube/pkg/types"
)

func kubeAPIServerSchema() *schema.Schema {
	return requiredBlock(false, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"common":                     controlplaneCommonSchema(),
				"host":                       hostSchema(false),
				"api_server_certificate":     requiredString(false),
				"api_server_key":             requiredSensitiveString(),
				"front_proxy_certificate":    requiredString(false),
				"front_proxy_key":            requiredSensitiveString(),
				"kubelet_client_certificate": requiredString(false),
				"kubelet_client_key":         requiredSensitiveString(),
				"service_account_public_key": requiredString(false),
				"etcd_ca_certificate":        requiredString(false),
				"etcd_client_certificate":    requiredString(false),
				"etcd_client_key":            requiredSensitiveString(),
				"service_cidr":               requiredString(false),
				"etcd_servers":               requiredStringList(false),
				"bind_address":               requiredString(false),
				"advertise_address":          requiredString(false),
				"secure_port":                optionalInt(false),
			},
		}
	})
}

func kubeAPIServerUnmarshal(i interface{}) controlplane.KubeAPIServer {
	c := controlplane.KubeAPIServer{}

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
		c.Host = hostUnmarshal(v.([]interface{})[0])
	}

	etcdServers := []string{}
	e := j["etcd_servers"].([]interface{})

	for _, v := range e {
		etcdServers = append(etcdServers, v.(string))
	}

	c.APIServerCertificate = types.Certificate(j["api_server_certificate"].(string))
	c.APIServerKey = types.PrivateKey(j["api_server_key"].(string))
	c.ServiceAccountPublicKey = j["service_account_public_key"].(string)
	c.BindAddress = j["bind_address"].(string)
	c.AdvertiseAddress = j["advertise_address"].(string)
	c.EtcdServers = etcdServers
	c.ServiceCIDR = j["service_cidr"].(string)
	c.SecurePort = j["secure_port"].(int)
	c.FrontProxyCertificate = types.Certificate(j["front_proxy_certificate"].(string))
	c.FrontProxyKey = types.PrivateKey(j["front_proxy_key"].(string))
	c.KubeletClientCertificate = types.Certificate(j["kubelet_client_certificate"].(string))
	c.KubeletClientKey = types.PrivateKey(j["kubelet_client_key"].(string))
	c.EtcdCACertificate = types.Certificate(j["etcd_ca_certificate"].(string))
	c.EtcdClientCertificate = types.Certificate(j["etcd_client_certificate"].(string))
	c.EtcdClientKey = types.PrivateKey(j["etcd_client_key"].(string))

	return c
}
