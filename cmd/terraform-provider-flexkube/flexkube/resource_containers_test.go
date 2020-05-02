package flexkube

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestContainersPlanOnly(t *testing.T) {
	config := `
resource "flexkube_containers" "foo" {
  container {
    name = "bar"

    container {
      config {
        name  = "bazhh"
        image = "nginx"
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

func TestContainersCreateRuntimeError(t *testing.T) {
	config := `
resource "flexkube_containers" "foo" {
  container {
    name = "bar"

    container {
			runtime {
				docker {
					host = "unix:///nonexistent"
				}
			}

      config {
        name  = "bazhh"
        image = "nginx"
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
				ExpectError: regexp.MustCompile(`Cannot connect to the Docker daemon`),
			},
		},
	})
}

func TestContainersValidateFail(t *testing.T) {
	config := `
resource "flexkube_containers" "foo" {
  container {
    name = "bar"

    container {
      config {
				name = ""
        image = "nginx"
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
				ExpectError: regexp.MustCompile(`name must be set`),
			},
		},
	})
}
