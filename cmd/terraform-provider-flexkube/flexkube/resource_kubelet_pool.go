package flexkube

import (
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/invidian/libflexkube/pkg/kubelet"
)

func resourceKubeletPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubeletPoolCreate,
		Read:   resourceKubeletPoolRead,
		Delete: resourceKubeletPoolDelete,
		Update: resourceKubeletPoolCreate,
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

func resourceKubeletPoolCreate(d *schema.ResourceData, m interface{}) error {
	c, err := kubelet.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
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
	return resourceKubeletPoolRead(d, m)
}

func resourceKubeletPoolRead(d *schema.ResourceData, m interface{}) error {
	c, err := kubelet.FromYaml([]byte(d.Get("state").(string) + d.Get("config").(string)))
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

func resourceKubeletPoolDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
