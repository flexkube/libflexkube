package flexkube

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-tls/tls"

	"github.com/flexkube/libflexkube/pkg/kubelet"
)

const kubeletPoolConfigPlanOnly = `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
  cgroup_driver = "systemd"
  network_plugin = "cni"
  first_controller_ip = local.controller_ips[0]
  api_port = 6443
  node_load_balancer_address = "127.0.0.1"
  controller_cidrs = ["10.0.1/0/24"]
}

module "root_pki" {
  source = "git::https://github.com/flexkube/terraform-root-pki.git"

  organization = "example"
}

module "kubernetes_pki" {
  source = "git::https://github.com/flexkube/terraform-kubernetes-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  api_server_ips            = local.controller_ips
  api_server_external_ips   = ["127.0.1.1"]
  api_server_external_names = ["kube-apiserver.example.com"]
  organization              = "example"
}

resource "flexkube_kubelet_pool" "controller" {
  bootstrap_config {
    server = local.node_load_balancer_address
    token  = "07401b.f395accd246ae52d"
  }

  cgroup_driver             = local.cgroup_driver
  network_plugin            = local.network_plugin
  kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
  hairpin_mode              = local.network_plugin == "kubenet" ? "promiscuous-bridge" : "hairpin-veth"
  volume_plugin_dir         = "/var/lib/kubelet/volumeplugins"
  cluster_dns_ips = [
    "11.0.0.10"
  ]

  system_reserved = {
    "cpu"    = "100m"
    "memory" = "500Mi"
  }

  kube_reserved = {
    // 100MB for kubelet and 200MB for etcd.
    "memory" = "300Mi"
    "cpu"    = "100m"
  }

  privileged_labels = {
    "node-role.kubernetes.io/master" = ""
  }

  admin_config {
    server             = "${local.first_controller_ip}:${local.api_port}"
    client_certificate = module.kubernetes_pki.kubernetes_admin_cert
    client_key         = module.kubernetes_pki.kubernetes_admin_key
  }

  taints = {
    "node-role.kubernetes.io/master" = "NoSchedule"
  }

  ssh {
    user     = "core"
    port     = 22
    password = "foo"
  }

  dynamic "kubelet" {
    for_each = local.controller_ips

    content {
      name     = local.controller_names[kubelet.key]
      pod_cidr = local.network_plugin == "kubenet" ? local.controller_cidrs[kubelet.key] : ""

      address = local.controller_ips[kubelet.key]

      host {
        ssh {
          address = kubelet.value
        }
      }
    }
  }
}
`

func TestKubeletPool(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
			"tls":      tls.Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:             kubeletPoolConfigPlanOnly,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:      kubeletPoolConfigCreateRuntimeError,
				ExpectError: regexp.MustCompile(`connection refused`),
			},
			{
				Config:      kubeletPoolConfigValidateFail,
				ExpectError: regexp.MustCompile(`networkPlugin must be either`),
			},
		},
	})
}

const kubeletPoolConfigCreateRuntimeError = `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
  cgroup_driver = "systemd"
  network_plugin = "cni"
  first_controller_ip = local.controller_ips[0]
  api_port = 6443
  node_load_balancer_address = "127.0.0.1"
  controller_cidrs = ["10.0.1/0/24"]
}

module "root_pki" {
  source = "git::https://github.com/flexkube/terraform-root-pki.git"

  organization = "example"
}

module "kubernetes_pki" {
  source = "git::https://github.com/flexkube/terraform-kubernetes-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  api_server_ips            = local.controller_ips
  api_server_external_ips   = ["127.0.1.1"]
  api_server_external_names = ["kube-apiserver.example.com"]
  organization              = "example"
}

resource "flexkube_kubelet_pool" "controller" {
  bootstrap_config {
    server = local.node_load_balancer_address
    token  = "07401b.f395accd246ae52d"
  }

  cgroup_driver             = local.cgroup_driver
  network_plugin            = local.network_plugin
  kubernetes_ca_certificate = module.kubernetes_pki.kubernetes_ca_cert
  hairpin_mode              = local.network_plugin == "kubenet" ? "promiscuous-bridge" : "hairpin-veth"
  volume_plugin_dir         = "/var/lib/kubelet/volumeplugins"
  cluster_dns_ips = [
    "11.0.0.10"
  ]

  system_reserved = {
    "cpu"    = "100m"
    "memory" = "500Mi"
  }

  kube_reserved = {
    // 100MB for kubelet and 200MB for etcd.
    "memory" = "300Mi"
    "cpu"    = "100m"
  }

  privileged_labels = {
    "node-role.kubernetes.io/master" = ""
  }

  admin_config {
    server             = "${local.first_controller_ip}:${local.api_port}"
    client_certificate = module.kubernetes_pki.kubernetes_admin_cert
    client_key         = module.kubernetes_pki.kubernetes_admin_key
  }

  taints = {
    "node-role.kubernetes.io/master" = "NoSchedule"
  }

  ssh {
    user     = "core"
    port     = 22
    password = "foo"
  }

  dynamic "kubelet" {
    for_each = local.controller_ips

    content {
      name     = local.controller_names[kubelet.key]
      pod_cidr = local.network_plugin == "kubenet" ? local.controller_cidrs[kubelet.key] : ""

      address = local.controller_ips[kubelet.key]

      host {
        ssh {
          address            = "127.0.0.1"
          port               = 12345
          password           = "bar"
          connection_timeout = "1s"
          retry_interval     = "1s"
          retry_timeout      = "1s"
        }
      }
    }
  }
}

`

const kubeletPoolConfigValidateFail = `
resource "flexkube_kubelet_pool" "controller" {
	kubelet {
		name		= "foo"
		address = ""
	}
}
`

func TestKubeletPoolUnmarshalIncludeState(t *testing.T) {
	t.Parallel()

	s := map[string]interface{}{
		"state_sensitive": []interface{}{
			map[string]interface{}{
				"foo": []interface{}{},
			},
		},
	}

	r := resourceKubeletPool()
	d := schema.TestResourceDataRaw(t, r.Schema, s)

	// Mark newly created object as created, so it's state is persisted.
	d.SetId("foo")

	// Create new ResourceData from the state, so it's persisted and there is no diff included.
	dn := r.Data(d.State())

	rc := kubeletPoolUnmarshal(dn, true)

	if rc.(*kubelet.Pool).State == nil {
		t.Fatalf("state should be unmarshaled, got: %v", cmp.Diff(nil, rc))
	}
}
