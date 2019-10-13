package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"

	"github.com/invidian/flexkube/pkg/etcd"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider})
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"lokomotive_etcd_cluster": resourceEtcdCluster(),
		},
	}
}

func resourceEtcdCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceEtcdClusterCreate,
		Read:   resourceEtcdClusterRead,
		Delete: resourceEtcdClusterDelete,
		Update: resourceEtcdClusterCreate,
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

func sha256sum(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
