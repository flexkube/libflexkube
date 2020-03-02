package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func hostMarshal(h host.Host, sensitive bool) interface{} {
	i := map[string]interface{}{}

	if h.DirectConfig != nil {
		i["direct"] = directMarshal(*h.DirectConfig)
	}

	if h.SSHConfig != nil {
		i["ssh"] = sshMarshal(*h.SSHConfig, sensitive)
	}

	return []interface{}{i}
}

func hostUnmarshal(i interface{}) host.Host {
	h := host.Host{
		DirectConfig: &direct.Config{},
	}

	if i == nil {
		return h
	}

	j, ok := i.(map[string]interface{})
	if !ok || len(j) == 0 {
		return h
	}

	h = host.Host{}

	if d, ok := j["direct"]; ok && len(d.([]interface{})) == 1 {
		h.DirectConfig = directUnmarshal(d.([]interface{})[0])
	}

	if d, ok := j["ssh"]; ok && len(d.([]interface{})) == 1 {
		h.SSHConfig = sshUnmarshal(d.([]interface{})[0])
	}

	return h
}

func hostSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, func(computed bool) map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"direct": directSchema(computed),
			"ssh":    sshSchema(computed),
		}
	})
}
