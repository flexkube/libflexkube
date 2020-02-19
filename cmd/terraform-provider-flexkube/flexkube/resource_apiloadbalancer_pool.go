package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

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

const resourceAPILoadBalancerName = "API load balancer"

func resourceAPILoadBalancerPoolCreate(d *schema.ResourceData, m interface{}) error {
	return createResource(d, &apiloadbalancer.APILoadBalancers{}, resourceAPILoadBalancerName)
}

func resourceAPILoadBalancerPoolRead(d *schema.ResourceData, m interface{}) error {
	return readResource(d, &apiloadbalancer.APILoadBalancers{}, resourceAPILoadBalancerName)
}

func resourceAPILoadBalancerPoolDelete(d *schema.ResourceData, m interface{}) error {
	return deleteResource(d)
}
