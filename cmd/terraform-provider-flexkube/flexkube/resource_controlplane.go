package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/controlplane"
)

func resourceControlplane() *schema.Resource {
	return &schema.Resource{
		Create: resourceControlplaneCreate,
		Read:   resourceControlplaneRead,
		Delete: resourceControlplaneDelete,
		Update: resourceControlplaneCreate,
		Schema: map[string]*schema.Schema{
			"config": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const resourceControlplaneName = "Kubernetes controlplane"

func resourceControlplaneCreate(d *schema.ResourceData, m interface{}) error {
	return createResource(d, &controlplane.Controlplane{}, resourceControlplaneName)
}

func resourceControlplaneRead(d *schema.ResourceData, m interface{}) error {
	return readResource(d, &controlplane.Controlplane{}, resourceControlplaneName)
}

func resourceControlplaneDelete(d *schema.ResourceData, m interface{}) error {
	return deleteResource(d)
}
