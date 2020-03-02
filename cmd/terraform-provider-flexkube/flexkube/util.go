package flexkube

import (
	"crypto/sha256"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/types"
)

const (
	// If schema is Type: schema.TypeList and Elem: &schema.Resource,
	// MaxItems should be set to this value, to treat the property as a
	// standalone, singleton block.
	blockMaxItems = 1
)

type getter interface {
	Get(key string) interface{}
	GetChange(key string) (interface{}, interface{})
	GetOk(key string) (interface{}, bool)
}

func requiredString(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: !computed,
		Computed: computed,
	}
}

func optionalString(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: computed,
	}
}

func sensitiveString(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeString,
		Optional:  true,
		Computed:  computed,
		Sensitive: true,
	}
}

func optionalStringList(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: computed,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func optionalBool(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Computed: computed,
	}
}

func optionalInt(computed bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Computed: computed,
	}
}

func optionalMap(computed bool, elem func(bool) *schema.Resource) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: computed,
		Elem:     elem(computed),
	}
}

func optionalMapPrimitive(computed bool, elem func(bool) *schema.Schema) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: computed,
		Elem:     elem(computed),
	}
}

func requiredBlock(computed bool, elem func(bool) *schema.Resource) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: !computed,
		Computed: computed,
		MaxItems: blockMaxItems,
		Elem:     elem(computed),
	}
}

func optionalBlock(computed bool, elem func(bool) map[string]*schema.Schema) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: computed,
		MaxItems: blockMaxItems,
		Elem: &schema.Resource{
			Schema: elem(computed),
		},
	}
}

func requiredList(computed bool, sensitive bool, elem func(bool) *schema.Resource) *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeList,
		Computed:  computed,
		Sensitive: sensitive,
		Required:  !computed,
		Elem:      elem(computed),
	}
}

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
