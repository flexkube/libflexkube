package flexkube //nolint:dupl

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-tls/tls"
)

func TestControlplanePlanOnly(t *testing.T) {
	config := `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
  first_controller_ip = local.controller_ips[0]
  api_port = 6443
  bootstrap_api_bind = "0.0.0.0"
}

module "root_pki" {
  source = "git::https://github.com/flexkube/terraform-root-pki.git"

  organization = "example"
}

module "etcd_pki" {
  source = "git::https://github.com/flexkube/terraform-etcd-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  peer_ips   = local.controller_ips
  peer_names = local.controller_names

  server_ips   = local.controller_ips
  server_names = local.controller_names

  client_cns = ["kube-apiserver-etcd-client"]

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

resource "flexkube_controlplane" "bootstrap" {
  common {
    kubernetes_ca_certificate  = module.kubernetes_pki.kubernetes_ca_cert
    front_proxy_ca_certificate = module.kubernetes_pki.kubernetes_front_proxy_ca_cert
  }

  kube_apiserver {
    api_server_certificate     = module.kubernetes_pki.kubernetes_api_server_cert
    api_server_key             = module.kubernetes_pki.kubernetes_api_server_key
    front_proxy_certificate    = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_cert
    front_proxy_key            = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_key
    kubelet_client_certificate = module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert
    kubelet_client_key         = module.kubernetes_pki.kubernetes_api_server_kubelet_client_key
    service_account_public_key = module.kubernetes_pki.service_account_public_key
    etcd_ca_certificate        = module.etcd_pki.etcd_ca_cert
    etcd_client_certificate    = module.etcd_pki.client_certs[0]
    etcd_client_key            = module.etcd_pki.client_keys[0]
    service_cidr               = "11.0.0.0/24"
    etcd_servers               = formatlist("https://%s:2379", module.etcd_pki.etcd_peer_ips)
    bind_address               = local.bootstrap_api_bind
    advertise_address          = local.first_controller_ip
    secure_port                = local.api_port
  }

  kube_controller_manager {
    flex_volume_plugin_dir      = "/var/lib/kubelet/volumeplugins"
    kubernetes_ca_key           = module.kubernetes_pki.kubernetes_ca_key
    service_account_private_key = module.kubernetes_pki.service_account_private_key
    root_ca_certificate         = module.root_pki.root_ca_cert

    kubeconfig {
      client_certificate = module.kubernetes_pki.kube_controller_manager_cert
      client_key         = module.kubernetes_pki.kube_controller_manager_key
    }
  }

  kube_scheduler {
    kubeconfig {
      client_certificate = module.kubernetes_pki.kube_scheduler_cert
      client_key         = module.kubernetes_pki.kube_scheduler_key
    }
  }

  api_server_address = local.first_controller_ip
  api_server_port    = local.api_port

  ssh {
    address            = "127.0.0.1"
    port               = 12345
    password           = "bar"
    connection_timeout = "1s"
    retry_interval     = "1s"
    retry_timeout      = "1s"
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
			"tls":      tls.Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestControlplaneCreateRuntimeError(t *testing.T) {
	config := `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
	first_controller_ip = local.controller_ips[0]
	api_port = 6443
	bootstrap_api_bind = "0.0.0.0"
}

module "root_pki" {
  source = "git::https://github.com/flexkube/terraform-root-pki.git"

  organization = "example"
}

module "etcd_pki" {
  source = "git::https://github.com/flexkube/terraform-etcd-pki.git"

  root_ca_cert      = module.root_pki.root_ca_cert
  root_ca_key       = module.root_pki.root_ca_key
  root_ca_algorithm = module.root_pki.root_ca_algorithm

  peer_ips   = local.controller_ips
  peer_names = local.controller_names

  server_ips   = local.controller_ips
  server_names = local.controller_names

  client_cns = ["kube-apiserver-etcd-client"]

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

resource "flexkube_controlplane" "bootstrap" {
  common {
    kubernetes_ca_certificate  = module.kubernetes_pki.kubernetes_ca_cert
    front_proxy_ca_certificate = module.kubernetes_pki.kubernetes_front_proxy_ca_cert
  }

  kube_apiserver {
    api_server_certificate     = module.kubernetes_pki.kubernetes_api_server_cert
    api_server_key             = module.kubernetes_pki.kubernetes_api_server_key
    front_proxy_certificate    = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_cert
    front_proxy_key            = module.kubernetes_pki.kubernetes_api_server_front_proxy_client_key
    kubelet_client_certificate = module.kubernetes_pki.kubernetes_api_server_kubelet_client_cert
    kubelet_client_key         = module.kubernetes_pki.kubernetes_api_server_kubelet_client_key
    service_account_public_key = module.kubernetes_pki.service_account_public_key
    etcd_ca_certificate        = module.etcd_pki.etcd_ca_cert
    etcd_client_certificate    = module.etcd_pki.client_certs[0]
    etcd_client_key            = module.etcd_pki.client_keys[0]
    service_cidr               = "11.0.0.0/24"
    etcd_servers               = formatlist("https://%s:2379", module.etcd_pki.etcd_peer_ips)
    bind_address               = local.bootstrap_api_bind
    advertise_address          = local.first_controller_ip
    secure_port                = local.api_port

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

		common {
			image = "bar"
		}
  }

  kube_controller_manager {
    flex_volume_plugin_dir      = "/var/lib/kubelet/volumeplugins"
    kubernetes_ca_key           = module.kubernetes_pki.kubernetes_ca_key
    service_account_private_key = module.kubernetes_pki.service_account_private_key
    root_ca_certificate         = module.root_pki.root_ca_cert

    kubeconfig {
      client_certificate = module.kubernetes_pki.kube_controller_manager_cert
      client_key         = module.kubernetes_pki.kube_controller_manager_key
    }

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

		common {
      image = "baz"
    }

  }

  kube_scheduler {
    kubeconfig {
      client_certificate = module.kubernetes_pki.kube_scheduler_cert
      client_key         = module.kubernetes_pki.kube_scheduler_key
    }

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

		common {
      image = "doh"
    }
  }

  api_server_address = local.first_controller_ip
  api_server_port    = local.api_port

	ssh {
    address            = "127.0.0.1"
    port               = 12345
    password           = "bar"
    connection_timeout = "1s"
    retry_interval     = "1s"
		retry_timeout      = "1s"
	}
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
			"tls":      tls.Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`connection refused`),
			},
		},
	})
}

func TestControlplaneValidateFail(t *testing.T) {
	config := `
resource "flexkube_controlplane" "bootstrap" {
  common {
    kubernetes_ca_certificate  = ""
    front_proxy_ca_certificate = ""
  }

  kube_apiserver {
    api_server_certificate     = ""
    api_server_key             = ""
    front_proxy_certificate    = ""
    front_proxy_key            = ""
    kubelet_client_certificate = ""
    kubelet_client_key         = ""
    service_account_public_key = ""
    etcd_ca_certificate        = ""
    etcd_client_certificate    = ""
    etcd_client_key            = ""
    service_cidr               = "11.0.0.0/24"
    etcd_servers               = []
    bind_address               = ""
    advertise_address          = ""
    secure_port                = 6443
  }

  kube_controller_manager {
    flex_volume_plugin_dir      = ""
    kubernetes_ca_key           = ""
    service_account_private_key = ""
    root_ca_certificate         = ""

    kubeconfig {
      client_certificate = ""
      client_key         = ""
    }
  }

  kube_scheduler {
    kubeconfig {
      client_certificate = ""
      client_key         = ""
    }
  }

  api_server_address = ""
	api_server_port    = 0

  ssh {
    user        = "core"
    address     = ""
    port        = 22
    private_key = ""
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`failed to decode PEM format`),
			},
		},
	})
}

func TestControlplaneDecodeEmptyBlocks(t *testing.T) {
	config := `
resource "flexkube_controlplane" "bootstrap" {
  common {}

  kube_apiserver {}

  kube_controller_manager {}

  kube_scheduler {}

  ssh {}
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`is required, but no definition was found`),
			},
		},
	})
}

func TestControlplaneDecodeEmptyConfig(t *testing.T) {
	config := `
resource "flexkube_controlplane" "bootstrap" {}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`required field is not set`),
			},
		},
	})
}

func TestControlplaneDecodeEmptyKubeconfig(t *testing.T) {
	config := `
resource "flexkube_controlplane" "bootstrap" {
	kube_controller_manager {
		kubeconfig {}
		common {}
	}
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`is required, but no definition was found`),
			},
		},
	})
}
