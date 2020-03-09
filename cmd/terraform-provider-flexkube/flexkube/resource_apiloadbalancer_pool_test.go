package flexkube //nolint:dupl

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAPILoadBalancerPoolPlanOnly(t *testing.T) {
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
