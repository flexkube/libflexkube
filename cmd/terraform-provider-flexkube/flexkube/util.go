package flexkube

import (
	"crypto/sha256"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/types"
)

func sha256sum(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func loadResource(d *schema.ResourceData, rc types.ResourceConfig) (types.Resource, error) {
	c := []byte(d.Get("state").(string) + d.Get("config").(string))

	r, err := types.ResourceFromYaml(c, rc)
	if err != nil {
		return nil, fmt.Errorf("failed loading resource from YAML: %w", err)
	}

	if err := r.CheckCurrentState(); err != nil {
		return nil, fmt.Errorf("failed checking current state of the resource: %w", err)
	}

	return r, nil
}

func saveResource(d *schema.ResourceData, r types.Resource) error {
	state, err := r.StateToYaml()
	if err != nil {
		return fmt.Errorf("failed serializing resource state to YAML: %w", err)
	}

	if err := d.Set("state", string(state)); err != nil {
		return fmt.Errorf("failed writing resource state to Terraform: %w", err)
	}

	result := sha256sum(state)
	d.SetId(result)

	return nil
}

func createResource(d *schema.ResourceData, rc types.ResourceConfig, name string) error {
	r, err := loadResource(d, rc)
	if err != nil {
		return fmt.Errorf("loading %s configuration failed: %w", name, err)
	}

	if err := r.Deploy(); err != nil {
		return fmt.Errorf("creating %s failed: %w", name, err)
	}

	return saveResource(d, r)
}

func readResource(d *schema.ResourceData, rc types.ResourceConfig, name string) error {
	r, err := loadResource(d, rc)
	if err != nil {
		return fmt.Errorf("loading %s configuration failed: %w", name, err)
	}

	return saveResource(d, r)
}

func deleteResource(d *schema.ResourceData) error {
	d.SetId("")

	return nil
}
