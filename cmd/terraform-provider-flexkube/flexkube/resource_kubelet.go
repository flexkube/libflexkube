package flexkube

import (
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/invidian/flexkube/pkg/etcd"
)

func resourceKubelet() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubeletCreate,
		Read:   resourceKubeletRead,
		Delete: resourceKubeletDelete,
		Update: resourceKubeletCreate,
		Schema: map[string]*schema.Schema{
			"config": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKubeletCreate(d *schema.ResourceData, m interface{}) error {
	c, err := etcd.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
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
	return resourceKubeletRead(d, m)
}

func resourceKubeletRead(d *schema.ResourceData, m interface{}) error {
	c, err := etcd.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
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

func resourceKubeletDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
