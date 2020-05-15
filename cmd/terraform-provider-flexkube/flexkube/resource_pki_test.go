package flexkube

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestPKI(t *testing.T) {
	config := `
resource "flexkube_pki" "pki" {
	etcd {
		peer_certificates {
			organization = "foo"
		}
	}

	kubernetes {}
}
`

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"flexkube": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config:             config,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
