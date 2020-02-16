package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/helm/release"
)

func resourceHelmRelease() *schema.Resource {
	return &schema.Resource{
		Create: resourceHelmReleaseCreate,
		Read:   resourceHelmReleaseRead,
		Delete: resourceHelmReleaseDelete,
		Update: resourceHelmReleaseCreate,
		Schema: map[string]*schema.Schema{
			"kubeconfig": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"chart": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"values": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getRelease(d *schema.ResourceData, m interface{}) (release.Release, error) {
	r := release.Config{
		Kubeconfig: d.Get("kubeconfig").(string),
		Namespace:  d.Get("namespace").(string),
		Name:       d.Get("name").(string),
		Chart:      d.Get("chart").(string),
		Values:     d.Get("values").(string),
		Version:    ">0.0.0-0",
	}

	l := m.(*meta)
	l.helmClientLock.Lock()
	defer l.helmClientLock.Unlock()

	return r.New()
}

func resourceHelmReleaseCreate(d *schema.ResourceData, m interface{}) error {
	release, err := getRelease(d, m)
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
	release, err := getRelease(d, m)
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
	release, err := getRelease(d, m)
	if err != nil {
		return err
	}

	if err := release.Uninstall(); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
