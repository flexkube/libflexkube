package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/etcd"
	"github.com/flexkube/libflexkube/pkg/types"
)

func membersUnmarshal(i interface{}) map[string]etcd.Member {
	j := i.([]interface{})

	members := map[string]etcd.Member{}

	for _, k := range j {
		t := k.(map[string]interface{})

		m := etcd.Member{
			Name:              t["name"].(string),
			Image:             t["image"].(string),
			CACertificate:     types.Certificate(t["ca_certificate"].(string)),
			PeerCertificate:   types.Certificate(t["peer_certificate"].(string)),
			PeerKey:           types.PrivateKey(t["peer_key"].(string)),
			PeerAddress:       t["peer_address"].(string),
			InitialCluster:    t["initial_cluster"].(string),
			PeerCertAllowedCN: t["peer_cert_allowed_cn"].(string),
			ServerCertificate: types.Certificate(t["server_certificate"].(string)),
			ServerKey:         types.PrivateKey(t["server_key"].(string)),
			ServerAddress:     t["server_address"].(string),
		}

		if v, ok := t["host"]; ok && len(v.([]interface{})) == 1 {
			m.Host = hostUnmarshal(v.([]interface{})[0])
		}

		members[t["name"].(string)] = m
	}

	return members
}

func memberSchema() *schema.Schema {
	return requiredList(false, false, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":                 requiredString(false),
				"image":                optionalString(false),
				"host":                 hostSchema(false),
				"ca_certificate":       optionalString(false),
				"peer_certificate":     requiredString(false),
				"peer_key":             requiredSensitiveString(false),
				"peer_address":         optionalString(false),
				"initial_cluster":      optionalString(false),
				"peer_cert_allowed_cn": optionalString(false),
				"server_certificate":   requiredString(false),
				"server_key":           requiredSensitiveString(false),
				"server_address":       requiredString(false),
			},
		}
	})
}
