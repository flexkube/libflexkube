package flexkube

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/flexkube/libflexkube/pkg/etcd"
)

func TestEtcdClusterPlanOnly(t *testing.T) {
	t.Parallel()

	config := `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
}

resource "flexkube_pki" "pki" {
  certificate {
    organization = "example"
  }

  etcd {
    peers   = zipmap(local.controller_names, local.controller_ips)
    servers = zipmap(local.controller_names, local.controller_ips)
  }
}

resource "flexkube_etcd_cluster" "etcd" {
  pki_yaml = flexkube_pki.pki.state_yaml

  ssh {
    user     = "core"
    password = "foo"
  }

  dynamic "member" {
    for_each = flexkube_pki.pki.etcd[0].peers

    content {
      name               = member.key
      peer_address       = member.value
      server_address     = member.value

      host {
        ssh {
          address = member.value
        }
      }
    }
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
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

const etcdClusterCreateRuntimeErrorConfig = `
locals {
  controller_ips = ["1.1.1.1"]
  controller_names = ["controller01"]
}

resource "flexkube_pki" "pki" {
  certificate {
    organization = "example"
  }

  etcd {
    peers   = zipmap(local.controller_names, local.controller_ips)
    servers = zipmap(local.controller_names, local.controller_ips)
  }
}

resource "flexkube_etcd_cluster" "etcd" {
  pki_yaml = flexkube_pki.pki.state_yaml

  ssh {
    user     = "core"
    password = "foo"
  }

  dynamic "member" {
    for_each = flexkube_pki.pki.etcd[0].peers

    content {
      name               = member.key
      peer_address       = member.value
      server_address     = member.value

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

func TestEtcdClusterCreateRuntimeError(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      etcdClusterCreateRuntimeErrorConfig,
				ExpectError: regexp.MustCompile(`connection refused`),
			},
		},
	})
}

func TestEtcdClusterValidateFail(t *testing.T) {
	t.Parallel()

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

func TestEtcdClusterUnmarshalIncludeState(t *testing.T) {
	t.Parallel()

	s := map[string]interface{}{
		"state_sensitive": []interface{}{
			map[string]interface{}{
				"foo": []interface{}{},
			},
		},
	}

	r := resourceEtcdCluster()
	d := schema.TestResourceDataRaw(t, r.Schema, s)

	// Mark newly created object as created, so it's state is persisted.
	d.SetId("foo")

	// Create new ResourceData from the state, so it's persisted and there is no diff included.
	dn := r.Data(d.State())

	rc := etcdClusterUnmarshal(dn, true)

	if rc.(*etcd.Cluster).State == nil {
		t.Fatalf("state should be unmarshaled, got: %v", cmp.Diff(nil, rc))
	}
}
