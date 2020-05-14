package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/pki"
)

func pkiKubeAPIServerMarshal(e *pki.KubeAPIServer) interface{} {
	return []interface{}{
		map[string]interface{}{
			"certificate":                    []interface{}{certificateMarshal(&e.Certificate)},
			"external_names":                 stringSliceToInterfaceSlice(e.ExternalNames),
			"server_ips":                     stringSliceToInterfaceSlice(e.ServerIPs),
			"server_certificate":             []interface{}{certificateMarshal(e.ServerCertificate)},
			"kubelet_certificate":            []interface{}{certificateMarshal(e.KubeletCertificate)},
			"front_proxy_client_certificate": []interface{}{certificateMarshal(e.FrontProxyClientCertificate)},
		},
	}
}

func pkiKubeAPIServerUnmarshal(i interface{}) *pki.KubeAPIServer {
	a := &pki.KubeAPIServer{}

	if i == nil {
		return a
	}

	j, ok := i.([]interface{})
	if !ok || len(j) != 1 {
		return a
	}

	k, ok := j[0].(map[string]interface{})

	if !ok || len(j) == 0 {
		return a
	}

	return &pki.KubeAPIServer{
		Certificate:                 *certificateUnmarshal(k["certificate"]),
		ExternalNames:               stringListUnmarshal(k["external_names"]),
		ServerIPs:                   stringListUnmarshal(k["server_ips"]),
		ServerCertificate:           certificateUnmarshal(k["server_certificate"]),
		KubeletCertificate:          certificateUnmarshal(k["kubelet_certificate"]),
		FrontProxyClientCertificate: certificateUnmarshal(k["front_proxy_client_certificate"]),
	}
}

func pkiKubeAPIServerSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"certificate":                    certificateBlockSchema(computed),
			"external_names":                 optionalStringList(computed),
			"server_ips":                     optionalStringList(computed),
			"server_certificate":             certificateBlockSchema(computed),
			"kubelet_certificate":            certificateBlockSchema(computed),
			"front_proxy_client_certificate": certificateBlockSchema(computed),
		}
	})
}
