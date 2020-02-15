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

func resourceControlplaneCreate(d *schema.ResourceData, m interface{}) error {
	c, err := controlplane.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
	if err != nil {
		return err
	}

	if err := c.CheckCurrentState(); err != nil {
		return err
	}

	if err := c.Deploy(); err != nil {
		return err
	}

	state, err := c.StateToYaml()
	if err != nil {
		return err
	}

	if err := d.Set("state", string(state)); err != nil {
		return err
	}

	return resourceControlplaneRead(d, m)
}

func resourceControlplaneRead(d *schema.ResourceData, m interface{}) error {
	c, err := controlplane.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
	if err != nil {
		return err
	}

	if err := c.CheckCurrentState(); err != nil {
		return err
	}

	state, err := c.StateToYaml()
	if err != nil {
		return err
	}

	if err := d.Set("state", string(state)); err != nil {
		return err
	}

	result := sha256sum(state)
	d.SetId(result)

	return nil
}

func resourceControlplaneDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")

	return nil
}
