package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/kubelet"
	"github.com/flexkube/libflexkube/pkg/types"
)

func kubeletsUnmarshal(i interface{}) []kubelet.Kubelet {
	j := i.([]interface{})

	kubelets := []kubelet.Kubelet{}

	for _, v := range j {
		if v == nil {
			continue
		}

		t := v.(map[string]interface{})

		k := kubelet.Kubelet{
			Image:                   t["image"].(string),
			KubernetesCACertificate: types.Certificate(t["kubernetes_ca_certificate"].(string)),
			CgroupDriver:            t["cgroup_driver"].(string),
			NetworkPlugin:           t["network_plugin"].(string),
			HairpinMode:             t["hairpin_mode"].(string),
			VolumePluginDir:         t["volume_plugin_dir"].(string),
			Name:                    t["name"].(string),
			Address:                 t["address"].(string),
			ClusterDNSIPs:           stringListUnmarshal(t["cluster_dns_ips"]),
			Taints:                  stringMapUnmarshal(t["taints"]),
			Labels:                  stringMapUnmarshal(t["labels"]),
			PrivilegedLabels:        stringMapUnmarshal(t["privileged_labels"]),
			SystemReserved:          stringMapUnmarshal(t["system_reserved"]),
			KubeReserved:            stringMapUnmarshal(t["kube_reserved"]),
			ExtraMounts:             mountsUnmarshal(t["mount"]),
		}

		if v, ok := t["wait_for_node_ready"]; ok {
			k.WaitForNodeReady = v.(bool)
		}

		if v, ok := t["host"]; ok && len(v.([]interface{})) == 1 {
			k.Host = hostUnmarshal(v.([]interface{})[0])
		}

		if v, ok := t["bootstrap_config"]; ok && len(v.([]interface{})) == 1 {
			k.BootstrapConfig = clientUnmarshal(v.([]interface{})[0])
		}

		if v, ok := t["admin_config"]; ok && len(v.([]interface{})) == 1 {
			k.AdminConfig = clientUnmarshal(v.([]interface{})[0])
		}

		kubelets = append(kubelets, k)
	}

	return kubelets
}

func kubeletSchema() *schema.Schema {
	return requiredList(false, false, func(computed bool) *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":                      requiredString(false),
				"image":                     optionalString(false),
				"host":                      hostSchema(false),
				"address":                   requiredString(false),
				"bootstrap_config":          clientSchema(false),
				"kubernetes_ca_certificate": optionalString(false),
				"cluster_dns_ips":           optionalStringList(false),
				"taints": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
					return &schema.Schema{
						Type: schema.TypeString,
					}
				}),
				"labels": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
					return &schema.Schema{
						Type: schema.TypeString,
					}
				}),
				"privileged_labels": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
					return &schema.Schema{
						Type: schema.TypeString,
					}
				}),
				"admin_config":   clientSchema(false),
				"cgroup_driver":  optionalString(false),
				"network_plugin": optionalString(false),
				"system_reserved": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
					return &schema.Schema{
						Type: schema.TypeString,
					}
				}),
				"kube_reserved": optionalMapPrimitive(false, func(computed bool) *schema.Schema {
					return &schema.Schema{
						Type: schema.TypeString,
					}
				}),
				"hairpin_mode":        optionalString(false),
				"volume_plugin_dir":   optionalString(false),
				"pod_cidr":            optionalString(false),
				"extra_mount":         mountsSchema(false),
				"wait_for_node_ready": optionalBool(false),
			},
		}
	})
}
