package flexkube

import (
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/flexkube/libflexkube/pkg/etcd"
)

func resourceEtcdCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceEtcdClusterCreate,
		Read:   resourceEtcdClusterRead,
		Delete: resourceEtcdClusterDelete,
		Update: resourceEtcdClusterCreate,
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

func resourceEtcdClusterCreate(d *schema.ResourceData, m interface{}) error {
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
	return resourceEtcdClusterRead(d, m)
}

func resourceEtcdClusterRead(d *schema.ResourceData, m interface{}) error {
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

func resourceEtcdClusterDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
