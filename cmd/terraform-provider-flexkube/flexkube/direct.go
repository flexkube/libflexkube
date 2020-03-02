package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func directMarshal(c direct.Config) interface{} {
	return []interface{}{map[string]interface{}{}}
}

func directUnmarshal(i interface{}) *direct.Config {
	return &direct.Config{}
}

func directSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{}
	})
}
