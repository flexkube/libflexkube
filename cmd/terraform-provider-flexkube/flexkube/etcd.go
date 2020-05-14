package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/pki"
)

func etcdMarshal(e *pki.Etcd) interface{} {
	return []interface{}{
		map[string]interface{}{
			"certificate":         []interface{}{certificateMarshal(&e.Certificate)},
			"ca":                  []interface{}{certificateMarshal(e.CA)},
			"client_cns":          stringSliceToInterfaceSlice(e.ClientCNs),
			"peers":               stringMapMarshal(e.Peers),
			"servers":             stringMapMarshal(e.Servers),
			"peer_certificates":   certificateMapMarshal(e.PeerCertificates),
			"server_certificates": certificateMapMarshal(e.ServerCertificates),
			"client_certificates": certificateMapMarshal(e.ClientCertificates),
		},
	}
}

func etcdUnmarshal(i interface{}) *pki.Etcd {
	if i == nil {
		return nil
	}

	e := &pki.Etcd{}

	j, ok := i.([]interface{})
	if !ok || len(j) != 1 {
		return e
	}

	k, ok := j[0].(map[string]interface{})

	if !ok || len(j) == 0 {
		return e
	}

	return &pki.Etcd{
		Certificate:        *certificateUnmarshal(k["certificate"]),
		CA:                 certificateUnmarshal(k["ca"]),
		ClientCNs:          stringListUnmarshal(k["client_cns"]),
		Peers:              stringMapUnmarshal(k["peers"]),
		Servers:            stringMapUnmarshal(k["servers"]),
		PeerCertificates:   certificateMapUnmarshal(k["peer_certificates"]),
		ServerCertificates: certificateMapUnmarshal(k["server_certificates"]),
		ClientCertificates: certificateMapUnmarshal(k["client_certificates"]),
	}
}

func etcdSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"certificate":         certificateBlockSchema(computed),
			"ca":                  certificateBlockSchema(computed),
			"client_cns":          optionalStringList(computed),
			"peers":               stringMapSchema(false),
			"servers":             stringMapSchema(false),
			"peer_certificates":   certificateMapSchema(computed),
			"server_certificates": certificateMapSchema(computed),
			"client_certificates": certificateMapSchema(computed),
		}
	})
}
