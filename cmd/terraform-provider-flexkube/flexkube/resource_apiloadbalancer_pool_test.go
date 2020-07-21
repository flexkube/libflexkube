package flexkube

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
)

func TestAPILoadBalancerPoolPlanOnly(t *testing.T) {
	t.Parallel()

	config := `
resource "flexkube_apiloadbalancer_pool" "bootstrap" {
  name             = "api-loadbalancer-bootstrap"
  host_config_path = "/etc/haproxy/bootstrap.cfg"
  bind_address     = "0.0.0.0:8443"
  servers          = ["192.168.1.2:6443"]

  api_load_balancer {}

  api_load_balancer {
    servers = ["192.168.1.3:6443"]
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

func TestAPILoadBalancerPoolCreateRuntimeError(t *testing.T) {
	t.Parallel()

	config := `
resource "flexkube_apiloadbalancer_pool" "bootstrap" {
  name             = "api-loadbalancer-bootstrap"
  host_config_path = "/etc/haproxy/bootstrap.cfg"
  bind_address     = "0.0.0.0:8443"
  servers          = ["192.168.1.2:6443"]

  ssh {
    port = 12345
  }

  api_load_balancer {
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
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`connection refused`),
			},
		},
	})
}

func TestAPILoadBalancerPoolValidateFail(t *testing.T) {
	t.Parallel()

	config := `
resource "flexkube_apiloadbalancer_pool" "bootstrap" {
  name             = "api-loadbalancer-bootstrap"
  host_config_path = "/etc/haproxy/bootstrap.cfg"
  bind_address     = "0.0.0.0:8443"
  servers          = []

  api_load_balancer {}
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`at least one server must be set`),
			},
		},
	})
}

func TestAPILoadBalancerPoolUnmarshalIncludeState(t *testing.T) {
	t.Parallel()

	s := map[string]interface{}{
		"state_sensitive": []interface{}{
			map[string]interface{}{
				"foo": []interface{}{},
			},
		},
	}

	r := resourceAPILoadBalancerPool()
	d := schema.TestResourceDataRaw(t, r.Schema, s)

	// Mark newly created object as created, so it's state is persisted.
	d.SetId("foo")

	// Create new ResourceData from the state, so it's persisted and there is no diff included.
	dn := r.Data(d.State())

	rc := apiLoadBalancersUnmarshal(dn, true)

	if rc.(*apiloadbalancer.APILoadBalancers).State == nil {
		t.Fatalf("state should be unmarshaled, got: %v", cmp.Diff(nil, rc))
	}
}
