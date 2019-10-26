package flexkube

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"flexkube_etcd_cluster":         resourceEtcdCluster(),
			"flexkube_kubelet_pool":         resourceKubeletPool(),
			"flexkube_controlplane":         resourceControlplane(),
			"flexkube_apiloadbalancer_pool": resourceAPILoadBalancerPool(),
		},
	}
}
