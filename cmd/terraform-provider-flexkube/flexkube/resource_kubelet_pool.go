package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/kubelet"
)

func resourceKubeletPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubeletPoolCreate,
		Read:   resourceKubeletPoolRead,
		Delete: resourceKubeletPoolDelete,
		Update: resourceKubeletPoolCreate,
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

const resourceKubeletPoolName = "kubelet pool"

func resourceKubeletPoolCreate(d *schema.ResourceData, m interface{}) error {
	return createResource(d, &kubelet.Pool{}, resourceKubeletPoolName)
}

func resourceKubeletPoolRead(d *schema.ResourceData, m interface{}) error {
	return readResource(d, &kubelet.Pool{}, resourceKubeletPoolName)
}

func resourceKubeletPoolDelete(d *schema.ResourceData, m interface{}) error {
	return deleteResource(d)
}
