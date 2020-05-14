package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/pki"
)

func kubernetesMarshal(e *pki.Kubernetes) interface{} {
	return []interface{}{
		map[string]interface{}{
			"certificate":                         []interface{}{certificateMarshal(&e.Certificate)},
			"ca":                                  []interface{}{certificateMarshal(e.CA)},
			"front_proxy_ca":                      []interface{}{certificateMarshal(e.FrontProxyCA)},
			"admin_certificate":                   []interface{}{certificateMarshal(e.AdminCertificate)},
			"kube_controller_manager_certificate": []interface{}{certificateMarshal(e.KubeControllerManagerCertificate)},
			"kube_scheduler_certificate":          []interface{}{certificateMarshal(e.KubeSchedulerCertificate)},
			"service_account_certificate":         []interface{}{certificateMarshal(e.ServiceAccountCertificate)},
			"kube_api_server":                     pkiKubeAPIServerMarshal(e.KubeAPIServer),
		},
	}
}

func kubernetesUnmarshal(i interface{}) *pki.Kubernetes {
	// if block is not defined at all, return nil, so PKI for Kubernetes is not triggered.
	if i == nil {
		return nil
	}

	k := &pki.Kubernetes{}

	j, ok := i.([]interface{})

	if !ok || len(j) != 1 {
		return k
	}

	l, ok := j[0].(map[string]interface{})

	if !ok || len(l) == 0 {
		return k
	}

	e := &pki.Kubernetes{
		Certificate:                      *certificateUnmarshal(l["certificate"]),
		CA:                               certificateUnmarshal(l["ca"]),
		FrontProxyCA:                     certificateUnmarshal(l["front_proxy_ca"]),
		AdminCertificate:                 certificateUnmarshal(l["admin_certificate"]),
		KubeControllerManagerCertificate: certificateUnmarshal(l["kube_controller_manager_certificate"]),
		KubeSchedulerCertificate:         certificateUnmarshal(l["kube_scheduler_certificate"]),
		ServiceAccountCertificate:        certificateUnmarshal(l["service_account_certificate"]),
		KubeAPIServer:                    pkiKubeAPIServerUnmarshal(l["kube_api_server"]),
	}

	return e
}

func kubernetesSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"certificate":                         certificateBlockSchema(computed),
			"ca":                                  certificateBlockSchema(computed),
			"front_proxy_ca":                      certificateBlockSchema(computed),
			"admin_certificate":                   certificateBlockSchema(computed),
			"kube_controller_manager_certificate": certificateBlockSchema(computed),
			"kube_scheduler_certificate":          certificateBlockSchema(computed),
			"service_account_certificate":         certificateBlockSchema(computed),
			"kube_api_server":                     pkiKubeAPIServerSchema(computed),
		}
	})
}
