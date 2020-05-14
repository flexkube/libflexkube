package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/etcd"
	"github.com/flexkube/libflexkube/pkg/types"
)

func resourceEtcdCluster() *schema.Resource {
	return &schema.Resource{
		Create:        resourceCreate(etcdClusterUnmarshal),
		Read:          resourceRead(etcdClusterUnmarshal),
		Delete:        resourceDelete(etcdClusterUnmarshal, "member"),
		Update:        resourceCreate(etcdClusterUnmarshal),
		CustomizeDiff: resourceDiff(etcdClusterUnmarshal),
		Schema: withCommonFields(map[string]*schema.Schema{
			"image":          optionalString(false),
			"ssh":            sshSchema(false),
			"ca_certificate": optionalString(false),
			"member":         memberSchema(),
			"pki_yaml":       sensitiveString(false),
		}),
	}
}

func etcdClusterUnmarshal(d getter, includeState bool) types.ResourceConfig {
	c := &etcd.Cluster{
		Image:         d.Get("image").(string),
		CACertificate: types.Certificate(d.Get("ca_certificate").(string)),
		Members:       membersUnmarshal(d.Get("member")),
		PKI:           unmarshalPKI(d),
	}

	if s := getState(d); includeState && s != nil {
		c.State = *s
	}

	if d, ok := d.GetOk("ssh"); ok && len(d.([]interface{})) == 1 {
		c.SSH = sshUnmarshal(d.([]interface{})[0])
	}

	return c
}
