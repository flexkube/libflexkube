package flexkube

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/invidian/libflexkube/pkg/helm"
)

func resourceHelmRelease() *schema.Resource {
	return &schema.Resource{
		Create: resourceHelmReleaseCreate,
		Read:   resourceHelmReleaseRead,
		Delete: resourceHelmReleaseDelete,
		Update: resourceHelmReleaseCreate,
		Schema: map[string]*schema.Schema{
			"kubeconfig": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"chart": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"values": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceHelmReleaseCreate(d *schema.ResourceData, m interface{}) error {
	r := helm.Release{
		Kubeconfig: d.Get("kubeconfig").(string),
		Namespace:  d.Get("namespace").(string),
		Name:       d.Get("name").(string),
		Chart:      d.Get("chart").(string),
		Values:     d.Get("values").(string),
		Version:    ">0.0.0-0",
	}

	release, err := r.New()
	if err != nil {
		return err
	}

	if err := release.InstallOrUpgrade(); err != nil {
		return err
	}

	result := sha256sum([]byte(d.Get("chart").(string) + d.Get("name").(string) + d.Get("namespace").(string) + d.Get("kubeconfig").(string)))
	d.SetId(result)

	return nil
}

func resourceHelmReleaseRead(d *schema.ResourceData, m interface{}) error {
	r := helm.Release{
		Kubeconfig: d.Get("kubeconfig").(string),
		Namespace:  d.Get("namespace").(string),
		Name:       d.Get("name").(string),
		Chart:      d.Get("chart").(string),
		Values:     d.Get("values").(string),
		Version:    ">0.0.0-0",
	}

	release, err := r.New()
	if err != nil {
		return err
	}

	e, err := release.Exists()
	if err != nil {
		return err
	}

	if e {
		result := sha256sum([]byte(d.Get("chart").(string) + d.Get("name").(string) + d.Get("namespace").(string) + d.Get("kubeconfig").(string)))
		d.SetId(result)
	} else {
		d.SetId("")
	}

	return nil
}

func resourceHelmReleaseDelete(d *schema.ResourceData, m interface{}) error {
	r := helm.Release{
		Kubeconfig: d.Get("kubeconfig").(string),
		Namespace:  d.Get("namespace").(string),
		Name:       d.Get("name").(string),
		Chart:      d.Get("chart").(string),
		Values:     d.Get("values").(string),
		Version:    ">0.0.0-0",
	}

	release, err := r.New()
	if err != nil {
		return err
	}

	if err := release.Uninstall(); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
