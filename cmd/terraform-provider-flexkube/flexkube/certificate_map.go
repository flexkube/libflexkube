package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/pki"
)

func certificateMapMarshal(cm map[string]*pki.Certificate) interface{} {
	r := []interface{}{}

	for _, v := range cm {
		r = append(r, certificateMarshal(v))
	}

	return r
}

func certificateMapUnmarshal(i interface{}) map[string]*pki.Certificate {
	r := map[string]*pki.Certificate{}

	if i == nil {
		return r
	}

	j := i.([]interface{})

	if len(j) == 0 {
		return r
	}

	for _, v := range j {
		if v == nil {
			continue
		}

		vv := v.(map[string]interface{})

		c := certificateUnmarshal(vv)

		r[c.CommonName] = c
	}

	return r
}

func certificateMapSchema(computed bool) *schema.Schema {
	return optionalList(computed, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: certificateSchema(computed),
		}
	})
}
