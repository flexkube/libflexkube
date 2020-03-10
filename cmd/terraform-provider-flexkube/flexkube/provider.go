// Package flexkube implements Terraform provider for libflexkube.
package flexkube

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider represents Terraform resource provider
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"flexkube_etcd_cluster":         resourceEtcdCluster(),
			"flexkube_kubelet_pool":         resourceKubeletPool(),
			"flexkube_controlplane":         resourceControlplane(),
			"flexkube_apiloadbalancer_pool": resourceAPILoadBalancerPool(),
			"flexkube_helm_release":         resourceHelmRelease(),
			"flexkube_containers":           resourceContainers(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return &meta{
		helmClientLock: sync.Mutex{},
	}, nil
}

// Meta is the meta information structure for the provider
type meta struct {
	// Mutex to create only one helm client as a time, as it is not thread-safe.
	helmClientLock sync.Mutex
}
