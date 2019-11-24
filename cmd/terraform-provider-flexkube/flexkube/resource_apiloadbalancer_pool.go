package flexkube

import (
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/flexkube/libflexkube/pkg/apiloadbalancer"
)

func resourceAPILoadBalancerPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPILoadBalancerPoolCreate,
		Read:   resourceAPILoadBalancerPoolRead,
		Delete: resourceAPILoadBalancerPoolDelete,
		Update: resourceAPILoadBalancerPoolCreate,
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

func resourceAPILoadBalancerPoolCreate(d *schema.ResourceData, m interface{}) error {
	c, err := apiloadbalancer.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
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

	return resourceAPILoadBalancerPoolRead(d, m)
}

func resourceAPILoadBalancerPoolRead(d *schema.ResourceData, m interface{}) error {
	c, err := apiloadbalancer.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
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

func resourceAPILoadBalancerPoolDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")

	return nil
}
