package flexkube //nolint:dupl

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-tls/tls"
)

func TestEtcdClusterPlanOnly(t *testing.T) {
	config := `
locals {
	controller_ips = ["1.1.1.1"]
	controller_names = ["controller01"]
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

resource "flexkube_etcd_cluster" "etcd" {
  ssh {
    user     = "core"
		password = "foo"
  }

  ca_certificate = module.etcd_pki.etcd_ca_cert

  dynamic "member" {
    for_each = module.etcd_pki.etcd_peer_ips

    content {
      name               = module.etcd_pki.etcd_peer_names[member.key]
      peer_certificate   = module.etcd_pki.etcd_peer_certs[member.key]
      peer_key           = module.etcd_pki.etcd_peer_keys[member.key]
      server_certificate = module.etcd_pki.etcd_server_certs[member.key]
      server_key         = module.etcd_pki.etcd_server_keys[member.key]
      peer_address       = module.etcd_pki.etcd_peer_ips[member.key]
      server_address     = local.controller_ips[member.key]

      host {
        ssh {
          address = local.controller_ips[member.key]
        }
      }
    }
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

func TestEtcdClusterCreateRuntimeError(t *testing.T) {
	config := `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
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

resource "flexkube_etcd_cluster" "etcd" {
  ssh {
    user     = "core"
    password = "foo"
  }

  ca_certificate = module.etcd_pki.etcd_ca_cert

  dynamic "member" {
    for_each = module.etcd_pki.etcd_peer_ips

    content {
      name               = module.etcd_pki.etcd_peer_names[member.key]
      peer_certificate   = module.etcd_pki.etcd_peer_certs[member.key]
      peer_key           = module.etcd_pki.etcd_peer_keys[member.key]
      server_certificate = module.etcd_pki.etcd_server_certs[member.key]
      server_key         = module.etcd_pki.etcd_server_keys[member.key]
      peer_address       = module.etcd_pki.etcd_peer_ips[member.key]
      server_address     = local.controller_ips[member.key]

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

func TestEtcdClusterValidateFail(t *testing.T) {
	config := `
resource "flexkube_etcd_cluster" "etcd" {
	member {
		server_address = ""
		server_key = ""
		server_certificate = ""
		peer_key = ""
		name = "foo"
		peer_certificate = ""
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
				ExpectError: regexp.MustCompile(`failed to validate member`),
			},
		},
	})
}
