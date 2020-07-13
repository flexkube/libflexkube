package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func hostConfiguredContainerMarshal(name string, c container.HostConfiguredContainer, sensitive bool) interface{} {
	return map[string]interface{}{
		"container":    containerMarshal(c.Container),
		"name":         name,
		"config_files": configFilesMarshal(c.ConfigFiles, sensitive),
		"host":         hostMarshal(c.Host, sensitive),
	}
}

func hostConfiguredContainerUnmarshal(i interface{}) (string, *container.HostConfiguredContainer) {
	j := i.(map[string]interface{})

	h := &container.HostConfiguredContainer{
		ConfigFiles: configFilesUnmarshal(j["config_files"]),
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if v, ok := j["container"]; ok && len(v.([]interface{})) == 1 {
		h.Container = containerUnmarshal(v.([]interface{})[0])
	}

	if v, ok := j["host"]; ok && len(v.([]interface{})) == 1 {
		h.Host = hostUnmarshal(v.([]interface{})[0])
	}

	return j["name"].(string), h
}

func hostConfiguredContainerSchema(computed, sensitive bool) *schema.Schema {
	return requiredList(computed, sensitive, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":         requiredString(computed),
				"container":    containerSchema(computed),
				"config_files": configFilesSchema(computed),
				"host":         hostSchema(computed),
			},
		}
	})
}
