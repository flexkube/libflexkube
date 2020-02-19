package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

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

const resourceEtcdClusterName = "etcd cluster"

func resourceEtcdClusterCreate(d *schema.ResourceData, m interface{}) error {
	return createResource(d, &etcd.Cluster{}, resourceEtcdClusterName)
}

func resourceEtcdClusterRead(d *schema.ResourceData, m interface{}) error {
	return readResource(d, &etcd.Cluster{}, resourceEtcdClusterName)
}

func resourceEtcdClusterDelete(d *schema.ResourceData, m interface{}) error {
	return deleteResource(d)
}
